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
	"groups/core/model"
	"groups/driven/calendar"
	"groups/driven/corebb"
	"groups/driven/rewards"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type scheduledTask struct {
	taskID *cron.EntryID
	cron   string
}

// Application represents the corebb application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	config *model.ApplicationConfig

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
	rewards       Rewards
	calendar      Calendar

	authmanSyncInProgress bool

	//synchronize managed groups timer
	scheduler         *cron.Cron
	managedGroupTasks map[string]scheduledTask
}

// Start starts the corebb part of the application
func (app *Application) Start() {
	// set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.RegisterStorageListener(&storageListener)

	app.setupSyncManagedGroupTimer()
}

// FindUser finds an user for the provided external id
func (app *Application) FindUser(clientID string, id *string, external bool) (*model.User, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}

	if id == nil || *id == "" {
		return nil, errors.New("id cannot be empty")
	}

	user, err := app.storage.FindUser(clientID, *id, external)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// LoginUser refactors the user using the new id
func (app *Application) LoginUser(clientID string, current *model.User, newID string) error {
	return app.storage.LoginUser(clientID, current)
}

// CreateUser creates an user
func (app *Application) CreateUser(clientID string, id string, externalID *string, email *string, name *string) (*model.User, error) {
	externalIDVal := ""
	if externalID != nil {
		externalIDVal = *externalID
	}

	emailVal := ""
	if email != nil {
		emailVal = *email
	}

	nameVal := ""
	if name != nil {
		nameVal = *name
	}

	user, err := app.storage.CreateUser(clientID, id, externalIDVal, emailVal, nameVal)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (app *Application) setupSyncManagedGroupTimer() {
	log.Println("setupSyncManagedGroupTimer")

	configs, err := app.storage.FindSyncConfigs()
	if err != nil {
		log.Printf("error loading sync configs: %s", err)
	}

	for _, config := range configs {
		task, ok := app.managedGroupTasks[config.ClientID]

		//cancel if active
		if ok && task.cron != config.CRON && task.taskID != nil {
			app.scheduler.Remove(*task.taskID)
			delete(app.managedGroupTasks, config.ClientID)
		}

		if (!ok || task.cron != config.CRON) && config.CRON != "" {
			sync := func() {
				log.Println("syncManagedGroups for clientID " + config.ClientID)
				err := app.synchronizeAuthman(config.ClientID, true)
				if err != nil {
					log.Printf("error syncing Authman groups for clientID %s: %s\n", config.ClientID, err.Error())
				}
			}
			taskID, err := app.scheduler.AddFunc(config.CRON, sync)
			if err != nil {
				log.Printf("error scheduling managed group sync for clientID %s: %s\n", config.ClientID, err)
			}
			app.managedGroupTasks[config.ClientID] = scheduledTask{taskID: &taskID, cron: config.CRON}
			log.Printf("sync managed group task scheduled for clientID=%s at %s\n", config.ClientID, config.CRON)
		}
	}
	app.scheduler.Start()
}

func (app *Application) createCalendarEventForGroups(clientID string, adminIdentifiers []model.AdminsIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	memberships, err := app.findGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: groupIDs,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if memberships.GetMembershipByAccountID(current.ID) != nil {
		mergedAdminIdentifiers, newGroupIDs := app.buildAdminIDs(current, memberships, adminIdentifiers)

		createdEvent, err := app.calendar.CreateCalendarEvent(mergedAdminIdentifiers, current.ID, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			for _, groupID := range newGroupIDs {
				mapping, err := app.storage.CreateEvent(clientID, eventID, groupID, nil, &model.Creator{
					UserID: current.ID,
					Name:   current.Name,
					Email:  current.Email,
				})
				if err != nil {
					log.Printf("Error create goup mapping: %s", err)
				}
				if mapping != nil {
					mappedGroupIDs = append(mappedGroupIDs, mapping.GroupID)
				}
			}
			return createdEvent, mappedGroupIDs, nil
		}
	}

	return nil, nil, nil
}

func (app *Application) buildAdminIDs(current *model.User, collection model.MembershipCollection, admins []model.AdminsIdentifiers) ([]model.AdminsIdentifiers, []string) {
	var newGroupIDs []string
	var newGroupIDsMapping = map[string]bool{}
	var newAdminIDsBulMapping = map[string]bool{}
	var newAdminIDsValMapping = map[string]model.AdminsIdentifiers{}

	// Reconstruct admins list
	for _, admin := range admins {
		userID := admin.AccountID
		newAdminIDsBulMapping[*userID] = true
		newAdminIDsValMapping[*userID] = admin
	}

	// Current user as default admin
	newAdminIDsBulMapping[current.ID] = true
	currentUserID := current.ID
	currentExternalID := current.ExternalID
	newAdminIDsValMapping[current.ID] = model.AdminsIdentifiers{
		AccountID:  &currentUserID,
		ExternalID: &currentExternalID,
	}

	// Construct new admin mappings and resolve duplications
	for _, membership := range collection.Items {
		if membership.ExternalID != "" || membership.UserID != "" {
			newGroupIDsMapping[membership.GroupID] = true
			userID := membership.UserID
			var externalID *string
			if membership.ExternalID != "" {
				externalID = &membership.ExternalID
			}
			if !newAdminIDsBulMapping[userID] {
				newAdminIDsBulMapping[userID] = true
				newAdminIDsValMapping[userID] = model.AdminsIdentifiers{
					AccountID:  &userID,
					ExternalID: externalID,
				}
			}
		}
	}

	// Construct group IDs
	for groupID := range newGroupIDsMapping {
		newGroupIDs = append(newGroupIDs, groupID)
	}

	// Construct admin IDs
	for adminID := range newAdminIDsValMapping {
		admins = append(admins, newAdminIDsValMapping[adminID])
	}

	return admins, newGroupIDs
}

func (app *Application) createCalendarEventForGroupsMembers(clientID string, orgID string, appID string, eventID string, groupIDs []string, members []model.ToMember) error {
	for _, groupID := range groupIDs {

		var userIDs []string
		memberships, err := app.findGroupMemberships(clientID, model.MembershipFilter{
			GroupIDs: []string{groupID},
			Statuses: []string{"admin", "member"},
		})
		if err != nil {
			return err
		}

		for _, membership := range memberships.Items {
			if len(members) > 0 {
				for _, toMember := range members {
					if toMember.UserID == membership.UserID {
						userIDs = append(userIDs, membership.UserID)
						break
					}
				}
			} else {
				userIDs = append(userIDs, membership.UserID)
			}
		}
		err = app.calendar.AddPeopleToCalendarEvent(userIDs, eventID, orgID, appID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) createCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	memberships, err := app.findGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{groupID},
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if memberships.GetMembershipByAccountID(current.ID) != nil {
		adminIdentifiers, newGroupIDs := app.buildAdminIDs(current, memberships, nil)

		createdEvent, err := app.calendar.CreateCalendarEvent(adminIdentifiers, current.ID, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			for _, constructedGroupID := range newGroupIDs {
				mapping, err := app.storage.CreateEvent(clientID, eventID, constructedGroupID, members, &model.Creator{
					UserID: current.ID,
					Name:   current.Name,
					Email:  current.Email,
				})
				if err != nil {
					log.Printf("Error create goup mapping: %s", err)
				}
				if mapping != nil {
					mappedGroupIDs = append(mappedGroupIDs, mapping.GroupID)
				}
			}
			return createdEvent, members, nil
		}
	}

	return nil, nil, nil
}

func (app *Application) updateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	memberships, err := app.findGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{groupID},
		UserID:   &current.ID,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if len(memberships.Items) > 0 {
		var groupIDs []string
		for _, membership := range memberships.Items {
			groupIDs = append(groupIDs, membership.GroupID)
		}

		eventID := event["id"].(string)
		createdEvent, err := app.calendar.UpdateCalendarEvent(current.ID, eventID, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			err := app.storage.UpdateEvent(clientID, eventID, groupID, members)
			if err != nil {
				return nil, nil, err
			}

			for _, groupID := range groupIDs {
				mapping, err := app.storage.CreateEvent(clientID, eventID, groupID, members, &model.Creator{
					UserID: current.ID,
					Name:   current.Name,
					Email:  current.Email,
				})
				if err != nil {
					log.Printf("app.updateCalendarEventSingleGroup() Error create goup mapping: %s", err)
				}
				if mapping != nil {
					mappedGroupIDs = append(mappedGroupIDs, mapping.GroupID)
				}
			}
			return createdEvent, members, nil
		}
	}

	return nil, nil, nil
}

func (app *Application) getGroupCalendarEvents(clientID string, current *model.User, groupID string, filter model.GroupEventFilter) (map[string]interface{}, error) {
	mappings, err := app.storage.FindEvents(clientID, current, groupID, true)
	if err != nil {
		return nil, err
	}

	if len(mappings) > 0 {
		var eventIDs []string
		for _, eventMapping := range mappings {
			eventIDs = append(eventIDs, eventMapping.EventID)
		}
		return app.calendar.GetGroupCalendarEvents(current.ID, eventIDs, current.AppID, current.OrgID, filter)
	}

	return nil, err
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter,
	rewards *rewards.Adapter, calendar *calendar.Adapter, config *model.ApplicationConfig) *Application {

	scheduler := cron.New(cron.WithLocation(time.UTC))
	managedGroupTasks := map[string]scheduledTask{}
	application := Application{version: version, build: build, storage: storage, notifications: notifications,
		authman: authman, corebb: core, rewards: rewards, calendar: calendar, config: config, scheduler: scheduler, managedGroupTasks: managedGroupTasks}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
