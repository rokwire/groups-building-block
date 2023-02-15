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

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
	rewards       Rewards

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
func (app *Application) FindUser(appID string, orgID string, id *string, external bool) (*model.User, error) {
	if appID == "" {
		return nil, errors.New("appID cannot be empty")
	}
	if orgID == "" {
		return nil, errors.New("orgID cannot be empty")
	}

	if id == nil || *id == "" {
		return nil, errors.New("id cannot be empty")
	}

	user, err := app.storage.FindUser(appID, orgID, *id, external)
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
func (app *Application) CreateUser(id string, appID string, orgID string, externalID *string, email *string, name *string) (*model.User, error) {
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

	user, err := app.storage.CreateUser(id, appID, orgID, externalIDVal, emailVal, nameVal)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (app *Application) setupSyncManagedGroupTimer() {
	log.Println("setupSyncManagedGroupTimer")

	configTypeSync := model.ConfigTypeSync
	configs, err := app.storage.FindConfigs(&configTypeSync, nil, nil)
	if err != nil {
		log.Printf("error loading sync configs: %s", err)
	}

	for _, config := range configs {
		syncConfig, err := model.GetConfigData[model.SyncConfigData](config)
		if err != nil {
			log.Printf("error asserting as sync config for appID %s, orgID %s: %v", config.AppID, config.OrgID, err)
			continue
		}

		managedGroupTasksKey := fmt.Sprintf("%s_%s", config.AppID, config.OrgID)
		task, ok := app.managedGroupTasks[managedGroupTasksKey]

		//cancel if active
		if ok && task.cron != syncConfig.CRON && task.taskID != nil {
			app.scheduler.Remove(*task.taskID)
			delete(app.managedGroupTasks, managedGroupTasksKey)
		}

		if (!ok || task.cron != syncConfig.CRON) && syncConfig.CRON != "" {
			sync := func() {
				log.Printf("syncManagedGroups for appID %s, orgID %s\n"+config.AppID, config.OrgID)
				err := app.synchronizeAuthman(config.AppID, config.OrgID, true)
				if err != nil {
					log.Printf("error syncing Authman groups for appID %s, orgID %s: %s\n", config.AppID, config.OrgID, err.Error())
				}
			}
			taskID, err := app.scheduler.AddFunc(syncConfig.CRON, sync)
			if err != nil {
				log.Printf("error scheduling managed group sync for appID %s, orgID %s: %s\n", config.AppID, config.OrgID, err)
			}
			app.managedGroupTasks[managedGroupTasksKey] = scheduledTask{taskID: &taskID, cron: syncConfig.CRON}
			log.Printf("sync managed group task scheduled for appID=%s, orgID=%s at %s\n", config.AppID, config.OrgID, syncConfig.CRON)
		}
	}
	app.scheduler.Start()
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter,
	rewards *rewards.Adapter) *Application {

	scheduler := cron.New(cron.WithLocation(time.UTC))
	managedGroupTasks := map[string]scheduledTask{}
	application := Application{version: version, build: build, storage: storage, notifications: notifications,
		authman: authman, corebb: core, rewards: rewards, scheduler: scheduler, managedGroupTasks: managedGroupTasks}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
