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
	"groups/core/model"
	"groups/driven/calendar"
	"groups/driven/corebb"
	"groups/driven/rewards"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rokwire/logging-library-go/v2/logs"
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

	Services Services //expose to the drivers adapters

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
	rewards       Rewards
	calendar      Calendar

	authmanSyncInProgress bool

	// delete data logic
	deleteDataLogic deleteDataLogic

	//synchronize managed groups timer
	scheduler *cron.Cron
}

// Start starts the corebb part of the application
func (app *Application) Start() {
	// set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.RegisterStorageListener(&storageListener)
	app.deleteDataLogic.start()

	app.setupCronTimer()
}

func (app *Application) setupCronTimer() {
	log.Println("setupCronTimer")

	configs, err := app.storage.FindSyncConfigs(nil)
	if err != nil {
		log.Printf("error loading sync configs: %s", err)
	}

	// TBD refactor this!
	for _, config := range configs {

		if config.CRON != "" {
			sync := func() {
				log.Println("syncManagedGroups for clientID " + config.ClientID)
				err := app.synchronizeAuthman(config.ClientID, true)
				if err != nil {
					log.Printf("error syncing Authman groups for clientID %s: %s\n", config.ClientID, err.Error())
				}
			}
			_, err := app.scheduler.AddFunc(config.CRON, sync)
			if err != nil {
				log.Printf("error scheduling managed group sync for clientID %s: %s\n", config.ClientID, err)
			}
			log.Printf("sync managed group task scheduled for clientID=%s at %s\n", config.ClientID, config.CRON)
		}
	}

	_, err = app.scheduler.AddFunc("* * * * *", func() {
		log.Println("run scheduled post tick")
		err := app.processScheduledPosts()
		if err != nil {
			log.Printf("error processing scheduled prosts: %s", err)
		}
	})
	if err != nil {
		log.Printf("error on running post scheduling task: %s", err)
	}
	log.Printf("successful running of post scheduling task")

	app.scheduler.Start()
}

func (app *Application) createCalendarEventForGroups(clientID string, adminIdentifiers []model.AccountIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	memberships, err := app.findGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: groupIDs,
		UserID:   &current.ID,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if memberships.GetMembershipByAccountID(current.ID) != nil {
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
		createdEvent, err := app.calendar.CreateCalendarEvent(adminIdentifiers, currentAccount, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			for _, membership := range memberships.Items {
				mapping, err := app.storage.CreateEvent(clientID, eventID, membership.GroupID, nil, &model.Creator{
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
		UserID:   &current.ID,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if memberships.GetMembershipByAccountID(current.ID) != nil {
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.ExternalID}
		createdEvent, err := app.calendar.CreateCalendarEvent([]model.AccountIdentifiers{}, currentAccount, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			mapping, err := app.storage.CreateEvent(clientID, eventID, groupID, members, &model.Creator{
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
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
		createdEvent, err := app.calendar.UpdateCalendarEvent(currentAccount, eventID, event, current.OrgID, current.AppID)
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

func (app *Application) getGroupCalendarEvents(clientID string, current *model.User, groupID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error) {
	mappings, err := app.storage.FindEvents(clientID, current, groupID, true)
	if err != nil {
		return nil, err
	}

	if len(mappings) > 0 {
		var eventIDs []string
		for _, eventMapping := range mappings {
			eventIDs = append(eventIDs, eventMapping.EventID)
		}
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
		return app.calendar.GetGroupCalendarEvents(currentAccount, eventIDs, current.AppID, current.OrgID, published, filter)
	}

	return nil, err
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter,
	rewards *rewards.Adapter, calendar *calendar.Adapter, serviceID string, logger *logs.Logger, config *model.ApplicationConfig) *Application {
	deleteDataLogic := deleteDataLogic{logger: *logger, coreAdapter: core, serviceID: serviceID, storage: storage}

	scheduler := cron.New(cron.WithLocation(time.UTC))
	application := Application{version: version,
		build:           build,
		storage:         storage,
		notifications:   notifications,
		authman:         authman,
		corebb:          core,
		rewards:         rewards,
		calendar:        calendar,
		config:          config,
		deleteDataLogic: deleteDataLogic,
		scheduler:       scheduler}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}

	return &application
}
