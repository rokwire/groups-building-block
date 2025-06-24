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
	"groups/driven/socialbb"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
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

	Services Services
	Admin    Administration
	BBS      BBS

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
	rewards       Rewards
	calendar      Calendar
	social        Social

	authmanSyncInProgress bool

	//synchronize managed groups timer
	scheduler *cron.Cron
	logger    *logs.Logger
}

// Start starts the corebb part of the application
func (app *Application) Start() {
	// set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.RegisterStorageListener(&storageListener)

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

	app.startCoreCleanupTask()

	app.scheduler.Start()
}

func (app *Application) startCoreCleanupTask() {
	// TBD: Implement CRUD APIs for config and load them from DB
	_, err := app.scheduler.AddFunc("0 0 * * *", func() {
		log.Println("run scheduled core account cleanup tick")
		app.processCoreAccountsCleanup()
	})
	if err != nil {
		log.Printf("error on running core account cleanup task: %s", err)
	}
	log.Printf("successful running of core account cleanup scheduling task")
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter,
	rewards *rewards.Adapter, calendar *calendar.Adapter, social *socialbb.Adapter, serviceID string, logger *logs.Logger, config *model.ApplicationConfig) *Application {

	scheduler := cron.New(cron.WithLocation(time.UTC))
	application := Application{version: version,
		build:         build,
		storage:       storage,
		notifications: notifications,
		authman:       authman,
		corebb:        core,
		rewards:       rewards,
		calendar:      calendar,
		social:        social,
		config:        config,
		scheduler:     scheduler,
		logger:        logger,
	}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Admin = &administrationImpl{app: &application}
	application.BBS = &bbsImpl{app: &application}

	return &application
}
