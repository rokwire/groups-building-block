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
	"groups/core/model"
	"groups/driven/storage"
	"log"
	"time"

	"github.com/google/uuid"
)

func (app *Application) synchronizeAuthman(clientID string, checkThreshold bool) error {
	startTime := time.Now()
	syncKey := "authman"
	transaction := func(context storage.TransactionContext) error {
		times, err := app.storage.FindSyncTimes(context, clientID, "authman", true)
		if err != nil {
			return err
		}

		time.Now().Unix()
		if times != nil && times.StartTime != nil {
			config, err := app.storage.FindSyncConfig(context, clientID)
			if err != nil {
				log.Printf("error finding sync configs for clientID %s: %v", clientID, err)
			}
			timeout := defaultConfigSyncTimeout
			if config != nil && config.Timeout > 0 {
				timeout = config.Timeout
			}

			if times.EndTime == nil {
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(timeout))) {
					log.Println("Another Authman sync process is running for clientID " + clientID)
					return fmt.Errorf("another Authman sync process is running" + clientID)
				}
				log.Printf("Authman sync past timeout threshold %d mins for client ID %s\n", timeout, clientID)
			}
			if checkThreshold {
				if config == nil {
					log.Printf("missing sync configs for clientID %s", clientID)
					return fmt.Errorf("missing sync configs for clientID %s: %v", clientID, err)
				}
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(config.TimeThreshold))) {
					log.Println("Authman has already been synced for clientID " + clientID)
					return fmt.Errorf("Authman has already been synced for clientID %s", clientID)
				}
			}
		}

		return app.storage.SaveSyncTimes(context, model.SyncTimes{StartTime: &startTime, EndTime: nil, ClientID: clientID, Key: syncKey})
	}

	err := app.storage.PerformTransaction(transaction)
	if err != nil {
		return err
	}

	log.Printf("Global Authman synchronization started for clientID: %s\n", clientID)

	app.authmanSyncInProgress = true
	finishAuthmanSync := func() {
		endTime := time.Now()
		err := app.storage.SaveSyncTimes(nil, model.SyncTimes{StartTime: &startTime, EndTime: &endTime, ClientID: clientID, Key: syncKey})
		if err != nil {
			log.Printf("Error saving sync configs to end sync: %s\n", err)
			return
		}
		log.Printf("Global Authman synchronization finished for clientID: %s\n", clientID)
	}
	defer finishAuthmanSync()

	configs, err := app.storage.FindManagedGroupConfigs(clientID)
	if err != nil {
		return fmt.Errorf("error finding managed group configs for clientID %s", clientID)
	}

	for _, config := range configs {
		for _, stemName := range config.AuthmanStems {
			stemGroups, err := app.authman.RetrieveAuthmanStemGroups(stemName)
			if err != nil {
				return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
			}

			if stemGroups != nil && len(stemGroups.WsFindGroupsResults.GroupResults) > 0 {
				for _, stemGroup := range stemGroups.WsFindGroupsResults.GroupResults {
					storedStemGroup, err := app.storage.FindAuthmanGroupByKey(clientID, stemGroup.Name)
					if err != nil {
						return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
					}

					title, adminUINs := stemGroup.GetGroupPrettyTitleAndAdmins()

					defaultAdminsMapping := map[string]bool{}
					for _, externalID := range adminUINs {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range app.config.AuthmanAdminUINList {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range config.AdminUINs {
						defaultAdminsMapping[externalID] = true
					}

					constructedAdminUINs := []string{}
					if len(defaultAdminsMapping) > 0 {
						for key := range defaultAdminsMapping {
							constructedAdminUINs = append(constructedAdminUINs, key)
						}
					}

					if storedStemGroup == nil {
						var memberships []model.GroupMembership
						if len(constructedAdminUINs) > 0 {
							memberships = app.buildMembersByExternalIDs(clientID, constructedAdminUINs, "admin")
						}

						emptyText := ""
						_, err := app.storage.CreateGroup(nil, clientID, nil, &model.Group{
							Title:                title,
							Description:          &emptyText,
							Category:             "Academic", // Hardcoded.
							Privacy:              "private",
							HiddenForSearch:      true,
							CanJoinAutomatically: true,
							AuthmanEnabled:       true,
							AuthmanGroup:         &stemGroup.Name,
						}, memberships)
						if err != nil {
							return fmt.Errorf("error on create Authman stem group: '%s' - %s", stemGroup.Name, err)
						}

						log.Printf("Created new `%s` group", title)
					} else {
						missedUINs := []string{}
						groupUpdated := false

						existingAdmins, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
							GroupIDs: []string{storedStemGroup.ID},
							Statuses: []string{"admin"},
						})

						membershipsForUpdate := []model.GroupMembership{}
						if len(existingAdmins.Items) > 0 {
							for _, uin := range adminUINs {
								found := false
								for _, member := range existingAdmins.Items {
									if member.ExternalID == uin {
										if member.Status != "admin" {
											now := time.Now()
											member.Status = "admin"
											member.DateUpdated = &now
											membershipsForUpdate = append(membershipsForUpdate, member)
											groupUpdated = true
											break
										}
										found = true
									}
								}
								if !found {
									missedUINs = append(missedUINs, uin)
								}
							}
						} else if err != nil {
							log.Printf("error rertieving admins for group: %s - %s", stemGroup.Name, err)
						}

						if len(missedUINs) > 0 {
							missedMembers := app.buildMembersByExternalIDs(clientID, missedUINs, "admin")
							if len(missedMembers) > 0 {
								membershipsForUpdate = append(membershipsForUpdate, missedMembers...)
								groupUpdated = true
							}
						}

						if storedStemGroup.Title != title {
							storedStemGroup.Title = title
							groupUpdated = true
						}

						if storedStemGroup.Category == "" {
							storedStemGroup.Category = "Academic" // Hardcoded.
							groupUpdated = true
						}

						if groupUpdated {
							err := app.storage.UpdateGroupWithMembership(nil, clientID, nil, storedStemGroup, membershipsForUpdate)
							if err != nil {
								log.Printf("error app.synchronizeAuthmanGroup() - unable to update group admins of '%s' - %s", storedStemGroup.Title, err)
							}
						}
					}
				}
			}
		}
	}

	authmanGroups, err := app.storage.FindAuthmanGroups(clientID)
	if err != nil {
		return err
	}

	if len(authmanGroups) > 0 {
		for _, authmanGroup := range authmanGroups {
			err := app.synchronizeAuthmanGroup(clientID, authmanGroup.ID)
			if err != nil {
				log.Printf("error app.synchronizeAuthmanGroup() '%s' - %s", authmanGroup.Title, err)
			}
		}
	}

	return nil
}

func (app *Application) buildMembersByExternalIDs(clientID string, externalIDs []string, memberStatus string) []model.GroupMembership {
	if len(externalIDs) > 0 {
		users, _ := app.corebb.GetAllCoreAccountsWithExternalIDs(externalIDs, nil, nil)
		members := []model.GroupMembership{}
		userExternalIDmapping := map[string]model.CoreAccount{}
		for _, user := range users {
			identifier := user.GetExternalID()
			if identifier != nil {
				userExternalIDmapping[*identifier] = user
			}
		}

		for _, externalID := range externalIDs {
			if value, ok := userExternalIDmapping[externalID]; ok {
				members = append(members, model.GroupMembership{
					ID:          uuid.NewString(),
					ClientID:    clientID,
					UserID:      value.ID,
					ExternalID:  externalID,
					Name:        value.GetFullName(),
					Email:       value.Profile.Email,
					Status:      memberStatus,
					DateCreated: time.Now(),
				})
			} else {
				members = append(members, model.GroupMembership{
					ID:          uuid.NewString(),
					ClientID:    clientID,
					ExternalID:  externalID,
					Status:      memberStatus,
					DateCreated: time.Now(),
				})
			}
		}
		return members
	}
	return nil
}

func (app *Application) synchronizeAuthmanGroup(clientID string, groupID string) error {
	if groupID == "" {
		return errors.New("Missing group ID")
	}
	var group *model.Group
	var err error
	group, err = app.checkGroupSyncTimes(clientID, groupID)
	if err != nil {
		return err
	}

	log.Printf("Authman synchronization for group %s started", *group.AuthmanGroup)

	authmanExternalIDs, authmanErr := app.authman.RetrieveAuthmanGroupMembers(*group.AuthmanGroup)
	if authmanErr != nil {
		return fmt.Errorf("error on requesting Authman for %s: %s", *group.AuthmanGroup, authmanErr)
	}

	app.authmanSyncInProgress = true
	finishAuthmanSync := func() {
		endTime := time.Now()
		group.SyncEndTime = &endTime
		err = app.storage.UpdateGroupSyncTimes(nil, clientID, group)
		if err != nil {
			log.Printf("Error saving group to end sync for Authman %s: %s\n", *group.AuthmanGroup, err)
			return
		}
		log.Printf("Authman synchronization for group %s finished", *group.AuthmanGroup)
	}
	defer finishAuthmanSync()

	err = app.syncAuthmanGroupMemberships(clientID, group, authmanExternalIDs)
	if err != nil {
		return fmt.Errorf("error updating group memberships for Authman %s: %s", *group.AuthmanGroup, err)
	}

	return nil
}

func (app *Application) checkGroupSyncTimes(clientID string, groupID string) (*model.Group, error) {
	var group *model.Group
	var err error
	startTime := time.Now()
	transaction := func(context storage.TransactionContext) error {
		group, err = app.storage.FindGroupWithContext(context, clientID, groupID, nil)
		if err != nil {
			return fmt.Errorf("error finding group for ID %s: %s", groupID, err)
		}
		if group == nil {
			return fmt.Errorf("missing group for ID %s", groupID)
		}
		if !group.IsAuthmanSyncEligible() {
			return fmt.Errorf("Authman synchronization failed for group '%s' due to bad settings", group.Title)
		}

		if group.SyncStartTime != nil {
			config, err := app.storage.FindSyncConfig(context, clientID)
			if err != nil {
				log.Printf("error finding sync configs for clientID %s: %v", clientID, err)
			}
			timeout := defaultConfigSyncTimeout
			if config != nil && config.GroupTimeout > 0 {
				timeout = config.GroupTimeout
			}
			if group.SyncEndTime == nil {
				if !startTime.After(group.SyncStartTime.Add(time.Minute * time.Duration(timeout))) {
					log.Println("Another Authman sync process is running for group ID " + group.ID)
					return fmt.Errorf("another Authman sync process is running for group ID %s", group.ID)
				}
				log.Printf("Authman sync timed out after %d mins for group ID %s\n", timeout, group.ID)
			}
		}

		group.SyncStartTime = &startTime
		group.SyncEndTime = nil
		err = app.storage.UpdateGroupSyncTimes(context, clientID, group)
		if err != nil {
			return fmt.Errorf("error switching to group memberships for Authman %s: %s", *group.AuthmanGroup, err)
		}
		return nil
	}

	err = app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (app *Application) syncAuthmanGroupMemberships(clientID string, authmanGroup *model.Group, authmanExternalIDs []string) error {
	syncID := uuid.NewString()
	log.Printf("Sync ID %s for Authman %s...\n", syncID, *authmanGroup.AuthmanGroup)

	// Get list of all member external IDs (Authman members + admins)
	allExternalIDs := append([]string{}, authmanExternalIDs...)

	// Load existing admins
	adminExternalIDsMap := map[string]bool{}
	adminMembers, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{authmanGroup.ID},
		Statuses: []string{"admin"},
	})
	if err != nil {
		return fmt.Errorf("error finding admin memberships in authman %s: %s", *authmanGroup.AuthmanGroup, err)
	}

	for _, adminMember := range adminMembers.Items {
		if len(adminMember.ExternalID) > 0 {
			allExternalIDs = append(allExternalIDs, adminMember.ExternalID)
			adminExternalIDsMap[adminMember.ExternalID] = true
		}
	}

	step := 0
	updateExternalIDs := []string{}
	updateOperations := []storage.SingleMembershipOperation{}
	batchUpdate := func(externalIDs []string, operations []storage.SingleMembershipOperation) {

		localUsersMapping := map[string]model.CoreAccount{}
		localUsers, err := app.corebb.GetAllCoreAccountsWithExternalIDs(externalIDs, nil, nil)
		if err != nil {
			log.Printf("Error on bulk loading %d core accounts in Authman %s: %s\n", len(externalIDs), *authmanGroup.AuthmanGroup, err)
		} else {
			log.Printf("Bulk load %d external IDs -> Loaded %d accounts in Authman %s: %s\n", len(externalIDs), len(localUsers), *authmanGroup.AuthmanGroup, err)
		}

		if len(localUsers) > 0 {
			for _, user := range localUsers {
				identifier := user.GetExternalID()
				if identifier != nil {
					localUsersMapping[*identifier] = user
				}
			}
		}
		for index := range operations {
			if localUser, ok := localUsersMapping[operations[index].ExternalID]; ok {
				operations[index].UserID = &localUser.ID
			}
		}

		err = app.storage.BulkUpdateGroupMembershipsByExternalID(clientID, authmanGroup.ID, updateOperations, false)
		if err != nil {
			log.Printf("Error on bulk saving step: %d, items: %d memberships, core accounts: %d in Authman %s: %s\n", step, len(updateOperations), len(localUsers), *authmanGroup.AuthmanGroup, err)
		} else {
			log.Printf("Successful bulk saving step: %d, items: %d memberships, core accounts: %d in Authman '%s'", step, len(updateOperations), len(localUsers), *authmanGroup.AuthmanGroup)
		}
		step++
	}

	log.Printf("Processing %d current members for Authman %s...\n", len(authmanExternalIDs), *authmanGroup.AuthmanGroup)
	for _, externalID := range authmanExternalIDs {

		status := "member"
		if _, ok := adminExternalIDsMap[externalID]; ok {
			status = "admin"
		}
		var userID *string
		var name *string
		var email *string
		updateExternalIDs = append(updateExternalIDs, externalID)
		updateOperations = append(updateOperations, storage.SingleMembershipOperation{
			ClientID:   clientID,
			GroupID:    authmanGroup.ID,
			ExternalID: externalID,
			UserID:     userID,
			Status:     &status,
			Email:      email,
			Name:       name,
			SyncID:     &syncID,
			Answers:    authmanGroup.CreateMembershipEmptyAnswers(),
		})

		if len(updateOperations) >= 1000 {
			batchUpdate(updateExternalIDs, updateOperations)
			updateExternalIDs = []string{}
			updateOperations = []storage.SingleMembershipOperation{}
		}

	}
	if len(updateOperations) > 0 {
		batchUpdate(updateExternalIDs, updateOperations)
	}

	// Delete removed non-admin members
	log.Printf("Deleting removed members for Authman %s...\n", *authmanGroup.AuthmanGroup)
	deleteCount, err := app.storage.DeleteUnsyncedGroupMemberships(clientID, authmanGroup.ID, syncID)
	if err != nil {
		log.Printf("Error deleting removed memberships in Authman %s\n", *authmanGroup.AuthmanGroup)
	} else {
		log.Printf("%d memberships removed from Authman %s\n", deleteCount, *authmanGroup.AuthmanGroup)
	}

	err = app.storage.UpdateGroupStats(nil, clientID, authmanGroup.ID, false, false, true, true)
	if err != nil {
		log.Printf("Error updating group stats for '%s' - %s", *authmanGroup.AuthmanGroup, err)
	}

	return nil
}
