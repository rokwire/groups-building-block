// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/storage"
	"log"
	"strings"
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

func (app *Application) findGroupsV3(clientID string, filter model.GroupsFilter) ([]model.Group, error) {
	return app.storage.FindGroupsV3(nil, clientID, filter)
}

func (app *Application) findGroupMemberships(context storage.TransactionContext, clientID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	collection, err := app.storage.FindGroupMembershipsWithContext(context, clientID, filter)

	if len(filter.GroupIDs) > 0 {
		groups, err := app.findGroupsV3(clientID, model.GroupsFilter{
			GroupIDs: filter.GroupIDs,
		})
		if err != nil {
			return model.MembershipCollection{}, fmt.Errorf("app.findGroupMemberships() error: %s", err)
		}

		groupIDMapping := map[string]model.Group{}
		for _, group := range groups {
			groupIDMapping[group.ID] = group
		}

		for index, member := range collection.Items {
			if group, ok := groupIDMapping[member.GroupID]; ok {
				collection.Items[index].ApplyGroupSettings(group.Settings)
			}
		}
		var userIds []string
		for _, s := range collection.Items {
			if s.UserID != "" {
				userIds = append(userIds, s.UserID)
			}
		}

		ferpa, err := app.corebb.RetrieveFerpaAccounts(userIds)
		if err != nil {
			return model.MembershipCollection{}, fmt.Errorf("RetrieveFerpaAccounts error: %s", err)
		}

		for i := range collection.Items {
			if contains(ferpa, collection.Items[i].UserID) {
				collection.Items[i] = model.GroupMembership{
					UserID: collection.Items[i].UserID, // Keep UserID but nullify others
				}
			}
		}
	}

	return collection, err
}

// Helper function to check if a slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
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
				return true, member.NotificationsPreferences.OverridePreferences &&
					(member.NotificationsPreferences.InvitationsMuted || member.NotificationsPreferences.AllMute)
			})

			if len(recipients) > 0 {
				topic := "group.invitations"
				groupStr := "Group"
				if group.ResearchGroup {
					groupStr = "Research Project"
				}

				message := fmt.Sprintf("New membership request for '%s' %s has been submitted", group.Title, strings.ToLower(groupStr))
				if group.CanJoinAutomatically {
					message = fmt.Sprintf("%s joined '%s' %s", member.GetDisplayName(), group.Title, strings.ToLower(groupStr))
				}

				app.notifications.SendNotification(
					recipients,
					&topic,
					fmt.Sprintf("%s - %s", groupStr, group.Title),
					message,
					map[string]string{
						"type":        "group",
						"operation":   "pending_member",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
					current.AppID,
					current.OrgID,
					nil,
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
		coreAccounts, err := app.corebb.GetAccountsWithIDs([]string{membership.UserID}, nil, nil, nil, nil)
		if err == nil && len(coreAccounts) > 0 {
			membership.ApplyFromCoreAccountIfEmpty(coreAccounts[0])
		} else {
			log.Printf("error app.createMembership() - unable to find core user by id: %s", err)
		}
	} else if membership.ExternalID != "" {
		coreAccounts, err := app.corebb.GetAllCoreAccountsWithExternalIDs([]string{membership.ExternalID}, nil, nil)
		if err == nil && len(coreAccounts) > 0 {
			membership.ApplyFromCoreAccountIfEmpty(coreAccounts[0])
		} else {
			log.Printf("error app.createMembership() - unable to find core user by external id: %s", err)
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
			return member.UserID != current.ID, member.NotificationsPreferences.OverridePreferences &&
				(member.NotificationsPreferences.InvitationsMuted || membership.NotificationsPreferences.AllMute)
		})

		groupStr := "Group"
		if group.ResearchGroup {
			groupStr = "Research Project"
		}

		var message string
		if membership.Status == "membership" || membership.Status == "admin" {
			message = fmt.Sprintf("New membership joined '%s' %s", group.Title, strings.ToLower(groupStr))
		} else {
			message = fmt.Sprintf("New membership request for '%s' %s has been submitted", group.Title, strings.ToLower(groupStr))
		}

		if len(recipients) > 0 {
			topic := "group.invitations"
			app.notifications.SendNotification(
				recipients,
				&topic,
				fmt.Sprintf("%s - %s", groupStr, group.Title),
				message,
				map[string]string{
					"type":        "group",
					"operation":   "pending_member",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
				current.AppID,
				current.OrgID,
				nil,
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
					log.Printf("err app.deleteMembershipByID() - error storing member: %s", err)
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
