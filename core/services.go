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
	"errors"
	"fmt"
	"groups/driven/rewards"
	"groups/driven/storage"
	"groups/utils"
	"time"

	"github.com/google/uuid"

	"groups/core/model"
	"groups/driven/notifications"
	"log"

	"strings"
)

const (
	defaultConfigSyncTimeout   = 60
	maxEmbeddedMemberGroupSize = 10000
	authmanUserBatchSize       = 5000
)

/*
func (app *Application) applyDataProtection(current *model.User, group model.Group) model.Group {
	//1 apply data protection for "anonymous"
	if current == nil || current.IsAnonymous {
		group.Members = []model.Member{}
	} else {
		member := group.GetMemberByUserID(current.ID)
		if member != nil && (member.IsRejected() || member.IsPendingMember()) {
			group.Members = []model.Member{}
			group.Members = append(group.Members, *member)
		}
	}
	return group
}*/

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getGroupEntity(OrgID string, id string) (*model.Group, error) {
	group, err := app.storage.FindGroup(nil, OrgID, id, nil)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) getGroupEntityByTitle(OrgID string, title string) (*model.Group, error) {
	group, err := app.storage.FindGroupByTitle(OrgID, title)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) isGroupAdmin(OrgID string, groupID string, userID string) (bool, error) {
	membership, err := app.storage.FindGroupMembership(OrgID, groupID, userID)
	if err != nil {
		return false, err
	}
	if membership == nil || membership.Status != "admin" {
		return false, nil
	}

	return true, nil
}

func (app *Application) createGroup(OrgID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError) {

	var groupError *utils.GroupError
	var groupID *string
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		var err error

		// Create intitial members if need
		var members []model.GroupMembership
		if membersConfig != nil && len(membersConfig.NetIDs) > 0 {
			accounts, err := app.corebb.GetAccounts(map[string]interface{}{
				"external_ids.net_id": membersConfig.NetIDs,
			}, &current.AppID, &current.OrgID, nil, nil)
			if err != nil {
				return nil
			}

			for _, account := range accounts {
				externalID := account.GetExternalID()
				fullName := account.GetFullName()
				netID := account.GetNetID()
				if externalID != "" && fullName != "" && netID != "" && netID != current.NetID {
					members = append(members, model.GroupMembership{
						OrgID:      OrgID,
						GroupID:    group.ID,
						UserID:     account.ID,
						ExternalID: externalID,
						NetID:      netID,
						Name:       fullName,
						Email:      account.Profile.Email,
						Status:     membersConfig.Status,
					})
				}
			}
		}

		groupID, groupError = app.storage.CreateGroup(context, OrgID, current, group, members)
		if groupError != nil {
			return err
		}

		if group.ResearchGroup {
			searchParams := app.formatCoreAccountSearchParams(group.ResearchProfile)

			list := []notifications.Recipient{}
			account, err := app.corebb.GetAccounts(searchParams, &current.AppID, &current.OrgID, nil, nil)
			if err != nil {
				return nil
			}
			for _, u := range account {
				id := u.ID
				mute := false
				ne := notifications.Recipient{UserID: id, Mute: mute}
				list = append(list, ne)
			}

			app.notifications.SendNotification(list, nil, "A new research project is available", fmt.Sprintf("%s by %s", group.Title, current.Name),
				map[string]string{
					"type":        "group",
					"operation":   "research_group",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
				current.AppID,
				current.OrgID,
				nil,
			)

		}

		return nil
	})

	handleRewardsAsync := func(OrgID, userID string) {
		count, grErr := app.storage.FindUserGroupsCount(OrgID, current.ID)
		if grErr != nil {
			log.Printf("Error createGroup(): %s", grErr)
		} else {
			if count != nil && *count == 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserCreatedFirstGroup, "")
			}
		}
	}
	go handleRewardsAsync(OrgID, current.ID)

	if groupError != nil {
		return nil, groupError
	}
	if err != nil {
		log.Printf("app.createGroup() error %s", err)
		return nil, utils.NewServerError()
	}

	return groupID, nil
}

func (app *Application) createGroupV3(OrgID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError) {

	var groupError *utils.GroupError
	var groupID *string
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		var err error

		// Create intitial members if need
		var members []model.GroupMembership
		accountIDs := []string{}
		accountIDMapping := map[string]model.MembershipStatus{}
		netIDs := []string{}
		netIDMapping := map[string]model.MembershipStatus{}

		accountIDsMapping := map[string]bool{}

		for _, memberRef := range membershipStatuses {
			if memberRef.UserID != "" {
				accountIDs = append(accountIDs, memberRef.UserID)
				accountIDMapping[memberRef.UserID] = memberRef
			} else if memberRef.NetID != "" {
				netIDs = append(netIDs, memberRef.NetID)
				netIDMapping[memberRef.NetID] = memberRef
			}
		}

		if len(accountIDs) > 0 {
			accounts, err := app.corebb.GetAccountsWithIDs(accountIDs, &current.AppID, &current.OrgID, nil, nil)
			if err != nil {
				return nil
			}

			for _, account := range accounts {
				if _, ok := accountIDsMapping[account.ID]; ok {
					continue
				}

				accountIDsMapping[account.ID] = true

				externalID := account.GetExternalID()
				fullName := account.GetFullName()
				netID := account.GetNetID()
				status := accountIDMapping[account.ID].Status

				members = append(members, model.GroupMembership{
					OrgID:      OrgID,
					GroupID:    group.ID,
					UserID:     account.ID,
					ExternalID: externalID,
					NetID:      netID,
					Name:       fullName,
					Email:      account.Profile.Email,
					Status:     status,
				})
			}
		}

		if len(netIDs) > 0 {
			accounts, err := app.corebb.GetAccounts(map[string]interface{}{
				"external_ids.net_id": netIDs,
			}, &current.AppID, &current.OrgID, nil, nil)
			if err != nil {
				return nil
			}

			for _, account := range accounts {
				if _, ok := accountIDsMapping[account.ID]; ok {
					continue
				}

				accountIDsMapping[account.ID] = true

				externalID := account.GetExternalID()
				fullName := account.GetFullName()
				netID := account.GetNetID()
				status := accountIDMapping[account.ID].Status

				members = append(members, model.GroupMembership{
					OrgID:      OrgID,
					GroupID:    group.ID,
					UserID:     account.ID,
					ExternalID: externalID,
					NetID:      netID,
					Name:       fullName,
					Email:      account.Profile.Email,
					Status:     status,
				})

			}
		}

		groupID, groupError = app.storage.CreateGroup(context, OrgID, current, group, members)
		if groupError != nil {
			return err
		}

		if group.ResearchGroup {
			searchParams := app.formatCoreAccountSearchParams(group.ResearchProfile)

			list := []notifications.Recipient{}
			account, err := app.corebb.GetAccounts(searchParams, &current.AppID, &current.OrgID, nil, nil)
			if err != nil {
				return nil
			}
			for _, u := range account {
				id := u.ID
				mute := false
				ne := notifications.Recipient{UserID: id, Mute: mute}
				list = append(list, ne)
			}

			app.notifications.SendNotification(list, nil, "A new research project is available", fmt.Sprintf("%s by %s", group.Title, current.Name),
				map[string]string{
					"type":        "group",
					"operation":   "research_group",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
				current.AppID,
				current.OrgID,
				nil,
			)

		}

		return nil
	})

	handleRewardsAsync := func(OrgID, userID string) {
		count, grErr := app.storage.FindUserGroupsCount(OrgID, current.ID)
		if grErr != nil {
			log.Printf("Error createGroup(): %s", grErr)
		} else {
			if count != nil && *count == 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserCreatedFirstGroup, "")
			}
		}
	}
	go handleRewardsAsync(OrgID, current.ID)

	if groupError != nil {
		return nil, groupError
	}
	if err != nil {
		log.Printf("app.createGroup() error %s", err)
		return nil, utils.NewServerError()
	}

	return groupID, nil
}

func (app *Application) updateGroup(OrgID string, current *model.User, group *model.Group) *utils.GroupError {

	err := app.storage.UpdateGroup(nil, OrgID, current, group)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateGroupDateUpdated(OrgID string, groupID string) error {
	err := app.storage.UpdateGroupDateUpdated(OrgID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) deleteGroup(OrgID string, current *model.User, id string, inactive bool) error {
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		group, err := app.storage.FindGroup(context, OrgID, id, nil)
		if err != nil {
			log.Printf("error finding group: %s", err)
			return err
		}

		var admins model.MembershipCollection
		if inactive {
			admins, err = app.storage.FindGroupMembershipsWithContext(context, OrgID, model.MembershipFilter{
				GroupIDs: []string{id},
				Statuses: []string{"admin"},
			})
			if err != nil {
				log.Printf("error finding group admins: %s", err)
				return err
			}
		}

		err = app.polls.DeleteGroupPolls(app.config.OrgID, id)
		if err != nil {
			log.Printf("error deleting group polls: %s", err)
			//return err // Do not block group deletion if polls cannot be deleted
		}

		err = app.storage.DeleteGroup(nil, OrgID, id)
		if err != nil {
			return err
		}

		if inactive {
			if len(admins.Items) > 0 {
				app.notifications.SendNotification(
					admins.GetMembersAsRecipients(func(membership model.GroupMembership) (bool, bool) {
						return membership.IsAdmin(), false
					}),
					nil,
					fmt.Sprintf("Your Group, \"%s\", has been removed due to inactivity.", group.Title), "", nil, current.AppID, current.OrgID, nil)
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("error deleting group: %s", err)
		return errors.New("error deleting group: " + err.Error())
	}

	return nil
}

func (app *Application) getAllGroupsUnsecured() ([]model.Group, error) {
	return app.storage.FindAllGroupsUnsecured()
}

func (app *Application) getGroups(OrgID string, current *model.User, filter model.GroupsFilter, skipMembershipCheck bool) (int64, []model.Group, error) {
	var userID *string
	if current != nil {
		userID = &current.ID
	}
	// find the groups objects
	return app.storage.FindGroups(OrgID, userID, filter, skipMembershipCheck)
}

func (app *Application) getAllGroups(OrgID string) (int64, []model.Group, error) {
	return app.storage.FindGroups(OrgID, nil, model.GroupsFilter{}, false)
}

func (app *Application) getUserGroups(OrgID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	// find the user groups
	groups, err := app.storage.FindUserGroups(OrgID, current.ID, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) deleteUser(OrgID string, current *model.User) error {
	return app.storage.DeleteUser(OrgID, current.ID)
}

func (app *Application) getGroup(OrgID string, current *model.User, id string) (*model.Group, error) {
	// find the group
	var userID *string
	if current != nil {
		userID = &current.ID
	}

	group, err := app.storage.FindGroup(nil, OrgID, id, userID)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (app *Application) getGroupFilterStats(OrgID string, current *model.User, filter model.StatsFilter, skipMembershipCheck bool) (*model.StatsResult, error) {
	return app.storage.CalculateGroupFilterStats(OrgID, current, filter, skipMembershipCheck)
}

func (app *Application) applyMembershipApproval(OrgID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	membership, err := app.storage.ApplyMembershipApproval(OrgID, membershipID, approve, rejectReason)
	if err != nil {
		return fmt.Errorf("error applying membership approval: %s", err)
	}
	if err == nil && membership != nil {
		group, _ := app.storage.FindGroup(nil, OrgID, membership.GroupID, nil)
		topic := "group.invitations"
		groupStr := "Group"
		if group.ResearchGroup {
			groupStr = "Research Project"
		}
		if approve {
			app.notifications.SendNotification(
				[]notifications.Recipient{
					membership.ToNotificationRecipient(membership.NotificationsPreferences.OverridePreferences &&
						(membership.NotificationsPreferences.InvitationsMuted || membership.NotificationsPreferences.AllMute)),
				},
				&topic,
				fmt.Sprintf("%s - %s", groupStr, group.Title),
				fmt.Sprintf("Your membership in '%s' %s has been approved", group.Title, strings.ToLower(groupStr)),
				map[string]string{
					"type":        "group",
					"operation":   "membership_approve",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
				current.AppID,
				current.OrgID,
				nil,
			)
		} else {
			app.notifications.SendNotification(
				[]notifications.Recipient{
					membership.ToNotificationRecipient(membership.NotificationsPreferences.OverridePreferences &&
						(membership.NotificationsPreferences.InvitationsMuted || membership.NotificationsPreferences.AllMute)),
				},
				&topic,
				fmt.Sprintf("%s - %s", groupStr, group.Title),
				fmt.Sprintf("Your membership in '%s' %s has been denied with a reason: %s", group.Title, strings.ToLower(groupStr), rejectReason),
				map[string]string{
					"type":        "group",
					"operation":   "membership_reject",
					"entity_type": "group",
					"entity_id":   group.ID,
					"entity_name": group.Title,
				},
				current.AppID,
				current.OrgID,
				nil,
			)
		}

		if approve && group.CanJoinAutomatically && group.AuthmanEnabled && membership.ExternalID != "" {
			err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, membership.ExternalID)
			if err != nil {
				log.Printf("err app.applyMembershipApproval() - error storing member in Authman: %s", err)
			}
		}
	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}

	return nil
}

func (app *Application) updateMembership(OrgID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	membership, _ := app.storage.FindGroupMembershipByID(OrgID, membershipID)
	if membership != nil {
		if status != nil && membership.Status != *status {
			membership.Status = *status
		}
		if dateAttended != nil && membership.DateAttended == nil {
			membership.DateAttended = dateAttended
		}
		if notificationsPreferences != nil {
			membership.NotificationsPreferences = *notificationsPreferences
		}

		err := app.storage.UpdateMembership(OrgID, current, membershipID, membership)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) createMembershipsStatuses(OrgID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error {
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		membership, _ := app.storage.FindGroupMembershipWithContext(context, OrgID, groupID, current.ID)

		if membership != nil && membership.IsAdmin() {

			group, err := app.storage.FindGroup(context, OrgID, groupID, &current.ID)
			if err != nil {
				return err
			}

			netIDs := membershipStatuses.GetAllNetIDs()
			var netIDAccounts []model.CoreAccount
			if len(netIDs) > 0 {
				netIDAccounts, err = app.corebb.GetAllCoreAccountsWithNetIDs(netIDs, &current.AppID, &current.OrgID)
				if err != nil {
					return err
				}
			}

			userIDs := membershipStatuses.GetAllUserIDs()
			var userIDAccounts []model.CoreAccount
			if len(userIDs) > 0 {
				userIDAccounts, err = app.corebb.GetAccountsWithIDs(userIDs, &current.AppID, &current.OrgID, nil, nil)
				if err != nil {
					return err
				}
			}

			var memberships []model.GroupMembership
			existingIDs := map[string]bool{}
			for _, membership := range membershipStatuses {
				found := false
				for _, account := range userIDAccounts {
					if membership.UserID == account.ID {
						if _, ok := existingIDs[account.ID]; !ok {
							existingIDs[account.ID] = true
							memberships = append(memberships, model.GroupMembership{
								OrgID:      OrgID,
								GroupID:    group.ID,
								UserID:     account.ID,
								ExternalID: account.GetExternalID(),
								NetID:      account.GetNetID(),
								Name:       account.GetFullName(),
								Email:      account.Profile.Email,
								Status:     membership.Status,
							})
							break
						}
					}
				}

				if !found {
					for _, account := range netIDAccounts {
						if membership.NetID == account.GetNetID() {
							if _, ok := existingIDs[account.ID]; !ok {
								existingIDs[account.ID] = true
								memberships = append(memberships, model.GroupMembership{
									OrgID:      OrgID,
									GroupID:    group.ID,
									UserID:     account.ID,
									ExternalID: account.GetExternalID(),
									NetID:      account.GetNetID(),
									Name:       account.GetFullName(),
									Email:      account.Profile.Email,
									Status:     membership.Status,
								})
								break
							}
						}
					}
				}
			}

			if len(memberships) > 0 {
				err := app.storage.CreateMemberships(context, OrgID, current, group, memberships)
				if err != nil {
					return err
				}

				return app.storage.UpdateGroupStats(context, OrgID, groupID, false, true, false, true)
			}

			log.Printf("Unable to find user with the specified details for group %s", groupID)
			return errors.New("Unable to find user with the specified details")
		}

		return nil
	})

	return err
}

func (app *Application) updateMemberships(OrgID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error {
	if group != nil && group.CurrentMember != nil && group.CurrentMember.IsAdmin() {
		err := app.storage.UpdateMemberships(OrgID, user, group.ID, operation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) reportGroupAsAbuse(OrgID string, current *model.User, group *model.Group, comment string) error {

	err := app.storage.ReportGroupAsAbuse(OrgID, current.ID, group)
	if err != nil {
		log.Printf("error while reporting an abuse group: %s", err)
		return fmt.Errorf("error while reporting an abuse group: %s", err)
	}

	subject := fmt.Sprintf("Report violation of Student Code to Dean of Students for group: %s", group.Title)

	body := fmt.Sprintf(`
<div>Group title: %s\n</div>
<div>Reported by: %s %s\n</div>
<div>Reported comment: %s\n</div>
	`, group.Title, current.ExternalID, current.Name, comment)
	body = strings.ReplaceAll(body, `\n`, "\n")
	return app.notifications.SendMail(app.config.ReportAbuseRecipientEmail, subject, body)
}

func (app *Application) sendGroupNotification(OrgID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error {
	memberStatuses := notification.MemberStatuses
	if len(memberStatuses) == 0 {
		memberStatuses = []string{"admin", "member"}
	}

	members, err := app.findGroupMemberships(nil, OrgID, model.MembershipFilter{
		GroupIDs: []string{notification.GroupID},
		UserIDs:  notification.Members.ToUserIDs(),
		Statuses: memberStatuses,
	})

	if err != nil {
		return err
	}

	recipients := members.GetMembersAsNotificationRecipients(predicate)
	app.sendNotification(recipients, notification.Topic, notification.Subject, notification.Body, notification.Data, app.config.AppID, app.config.OrgID)

	return nil
}

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string) {
	app.notifications.SendNotification(recipients, topic, title, text, data, appID, orgID, nil)
}

func (app *Application) getManagedGroupConfigs(OrgID string) ([]model.ManagedGroupConfig, error) {
	return app.storage.FindManagedGroupConfigs(OrgID)
}

func (app *Application) createManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error) {
	config.ID = uuid.NewString()
	config.DateCreated = time.Now()
	config.DateUpdated = nil
	err := app.storage.InsertManagedGroupConfig(config)
	return &config, err
}

func (app *Application) updateManagedGroupConfig(config model.ManagedGroupConfig) error {
	return app.storage.UpdateManagedGroupConfig(config)
}

func (app *Application) deleteManagedGroupConfig(id string, OrgID string) error {
	return app.storage.DeleteManagedGroupConfig(id, OrgID)
}

func (app *Application) getSyncConfig(OrgID string) (*model.SyncConfig, error) {
	return app.storage.FindSyncConfig(nil, OrgID)
}

func (app *Application) updateSyncConfig(config model.SyncConfig) error {
	return app.storage.SaveSyncConfig(nil, config)
}

func (app *Application) findGroupMembership(OrgID string, groupID string, userID string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembership(OrgID, groupID, userID)
}

func (app *Application) getResearchProfileUserCount(OrgID string, current *model.User, researchProfile map[string]map[string]any) (int64, error) {
	searchParams := app.formatCoreAccountSearchParams(researchProfile)
	return app.corebb.GetAccountsCount(searchParams, &current.AppID, &current.OrgID)
}

func (app *Application) formatCoreAccountSearchParams(researchProfile map[string]map[string]any) map[string]interface{} {
	searchParams := map[string]interface{}{}
	for k1, v1 := range researchProfile {
		for k2, v2 := range v1 {
			searchParams["profile.unstructured_properties.research_questionnaire_answers."+k1+"."+k2] = map[string]interface{}{"operation": "any", "value": v2}
		}
	}
	// If empty profile is provided, find all users that have filled out the profile
	//TODO: Handle filled out profile search better
	if len(searchParams) == 0 {
		searchParams["profile.unstructured_properties.research_questionnaire_answers.demographics"] = "$exists"
	}

	return searchParams
}

func (app *Application) onUpdatedGroupExternalEntity(groupID string, operation model.ExternalOperation) error {
	return app.storage.OnUpdatedGroupExternalEntity(nil, groupID, operation)
}
