package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/notifications"
	"log"
)

func (app *Application) checkUserGroupMembershipPermission(clientID string, current *model.User, groupID string) (*model.Group, bool) {
	group, err := app.getGroup(clientID, current, groupID)
	if err != nil {
		log.Printf("app.checkUserGroupMembershipPermission() error - unable to find group %s - %s", groupID, err)
		return group, false
	}
	if group != nil {
		if group.Privacy == "private" {
			if current == nil || current.IsAnonymous {
				log.Println("app.checkUserGroupMembershipPermission() error - Anonymous user cannot see the events for a private group")
				return group, false
			}
			if (group.CurrentMember != nil && !group.CurrentMember.IsAdminOrMember() && !group.HiddenForSearch) ||
				(group.CurrentMember == nil && group.HiddenForSearch) {
				log.Printf("app.checkUserGroupMembershipPermission() error - %s cannot see  %s private group as he/she is not a member or admin", current.Email, group.Title)
				return group, false
			}
		}
		return group, true
	}
	return nil, false
}

func (app *Application) findGroupV3(clientID string, filter *model.GroupsFilter) (*model.Group, error) {
	// assume we filter one nd just return the first one. Enough for now
	groups, err := app.findGroupsV3(clientID, filter)
	if len(groups) > 0 {
		return &groups[0], err
	}
	return nil, err
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
			recipients := []notifications.Recipient{}
			for _, admin := range adminMemberships.Items {
				recipients = append(recipients, notifications.Recipient{
					UserID: admin.UserID,
					Name:   admin.Name,
				})
			}
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

	if (membership.UserID == "" && membership.ExternalID != "") ||
		(membership.UserID != "" && membership.ExternalID == "") {
		if membership.ExternalID == "" {
			user, err := app.storage.FindUser(clientID, membership.UserID, false)
			if err == nil && user != nil {
				membership.ApplyFromUserIfEmpty(user)
			} else {
				log.Printf("error app.createMembership() - unable to find user: %s", err)
			}
		}
		if membership.UserID == "" {
			user, err := app.storage.FindUser(clientID, membership.ExternalID, true)
			if err == nil && user != nil {
				membership.ApplyFromUserIfEmpty(user)
			} else {
				log.Printf("error app.createMembership() - unable to find user: %s", err)
			}
		}
	}

	err := app.storage.CreateMembershipUnchecked(clientID, current, group, membership)
	if err != nil {
		return err
	}

	memberships, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		Statuses: []string{"admin"},
	})
	if err == nil && len(memberships.Items) > 0 {
		recipients := []notifications.Recipient{}
		for _, adminMember := range memberships.Items {
			if adminMember.UserID != current.ID {
				recipients = append(recipients, notifications.Recipient{
					UserID: adminMember.UserID,
					Name:   adminMember.Name,
				})
			}
		}

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