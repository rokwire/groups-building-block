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
	"groups/driven/corebb"
	"groups/driven/rewards"
)

// Application represents the corebb application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	config *model.Config

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
	rewards       Rewards
}

// Start starts the corebb part of the application
func (app *Application) Start() {
	// set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.SetStorageListener(&storageListener)
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

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter,
	rewards *rewards.Adapter, config *model.Config) *Application {

	application := Application{version: version, build: build, storage: storage, notifications: notifications,
		authman: authman, corebb: core, rewards: rewards, config: config}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
