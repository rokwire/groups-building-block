package core

import (
	"fmt"
	"groups/core/model"
	"log"
)

func (app *Application) checkUserGroupMembershipPermission(clientID string, current *model.User, groupID string) (*model.Group, bool) {
	if current == nil || current.IsAnonymous {
		log.Println("app.checkUserGroupMembershipPermission() error - Anonymous user cannot see the events for a private group")
		return nil, false
	}

	group, err := app.getGroup(clientID, current, groupID)
	if err != nil {
		log.Printf("app.checkUserGroupMembershipPermission() error - unable to find group %s - %s", groupID, err)
		return group, false
	}
	if group != nil {
		if group.CurrentMember != nil && group.CurrentMember.IsAdminOrMember() {
			return group, true
		}
	}
	return nil, false
}

func (app *Application) findGroupsV3(clientID string, filter *model.GroupsFilter) ([]model.Group, error) {
	return app.storage.FindGroupsV3(clientID, filter)
}

func (app *Application) findGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return app.storage.FindGroupMemberships(clientID, filter)
}

func (app *Application) findGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembershipByID(clientID, id)
}

func (app *Application) findUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error) {
	return app.storage.FindUserGroupMemberships(clientID, userID)
}

func (app *Application) createPendingMembership(clientID string, current *model.User, group *model.Group, member *model.GroupMembership) error {

	if group.CanJoinAutomatically {
		member.Status = "member"
	} else {
		member.Status = "pending"
	}

	err := app.storage.CreatePendingMembership(clientID, current, group, member)
	if err != nil {
		return err
	}

	adminMemberships, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		Statuses: []string{"admin"},
	})
	if err == nil && len(adminMemberships.Items) > 0 {
		if len(adminMemberships.Items) > 0 {
			recipients := adminMemberships.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
				return true, member.NotificationsPreferences.OverridePreferences && member.NotificationsPreferences.InvitationsMuted
			})

			if len(recipients) > 0 {
				topic := "group.invitations"

				message := fmt.Sprintf("New membership request for '%s' group has been submitted", group.Title)
				if group.CanJoinAutomatically {
					message = fmt.Sprintf("%s joined '%s' group", member.GetDisplayName(), group.Title)
				}

				app.notifications.SendNotification(
					recipients,
					&topic,
					fmt.Sprintf("Group - %s", group.Title),
					message,
					map[string]string{
						"type":        "group",
						"operation":   "pending_member",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
				)
			}
		}
	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}

	if group.CanJoinAutomatically && group.AuthmanEnabled {
		err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, member.ExternalID)
		if err != nil {
			log.Printf("err app.createPendingMembership() - error storing member in Authman: %s", err)
		}
	}

	return nil
}

func (app *Application) createMembership(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {

	if membership.UserID != "" {
		user, err := app.storage.FindUser(clientID, membership.UserID, false)
		if err == nil && user != nil {
			membership.ApplyFromUserIfEmpty(user)
		} else {
			log.Printf("error app.createMembership() - unable to find user: %s", err)
		}
	} else if membership.ExternalID != "" {
		user, err := app.storage.FindUser(clientID, membership.ExternalID, true)
		if err == nil && user != nil {
			membership.ApplyFromUserIfEmpty(user)
		} else {
			log.Printf("error app.createMembership() - unable to find user: %s", err)
		}
	}

	err := app.storage.CreateMembership(clientID, current, group, membership)
	if err != nil {
		return err
	}

	memberships, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		Statuses: []string{"admin"},
	})
	if err == nil && len(memberships.Items) > 0 {
		recipients := memberships.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
			return member.UserID != current.ID, member.NotificationsPreferences.OverridePreferences && member.NotificationsPreferences.InvitationsMuted
		})

		var message string
		if membership.Status == "membership" || membership.Status == "admin" {
			message = fmt.Sprintf("New membership joined '%s' group", group.Title)
		} else {
			message = fmt.Sprintf("New membership request for '%s' group has been submitted", group.Title)
		}

		if len(recipients) > 0 {
			topic := "group.invitations"
			app.notifications.SendNotification(
				recipients,
				&topic,
				fmt.Sprintf("Group - %s", group.Title),
				message,
				map[string]string{
					"type":        "group",
					"operation":   "pending_member",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
			)

		}

		if group.AuthmanEnabled && group.AuthmanGroup != nil {
			err = app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, membership.ExternalID)
			if err != nil {
				return err
			}
		}

	} else if err != nil {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createMembership() - error storing membership in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) deletePendingMembership(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeleteMembership(clientID, groupID, current.ID)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroup(nil, clientID, groupID, nil)
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createPendingMembership() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) deleteMembershipByID(clientID string, current *model.User, membershipID string) error {

	membership, _ := app.storage.FindGroupMembershipByID(clientID, membershipID)

	if membership != nil {
		err := app.storage.DeleteMembershipByID(clientID, current, membership.ID)
		if err != nil {
			return err
		}

		if membership != nil {
			group, _ := app.storage.FindGroup(nil, clientID, membership.GroupID, nil)
			if group.CanJoinAutomatically && group.AuthmanEnabled && membership.ExternalID != "" {
				err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, membership.ExternalID)
				if err != nil {
					log.Printf("err app.createPendingMembership() - error storing member in Authman: %s", err)
				}
			}
		}
	}
	return nil
}

func (app *Application) deleteMembership(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeleteMembership(clientID, groupID, current.ID)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroup(nil, clientID, groupID, nil)
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createPendingMembership() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}
