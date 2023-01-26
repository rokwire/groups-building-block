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
	"time"

	"github.com/google/uuid"
)

func (app *Application) getTODO() error {
	return nil
}

func (app *Application) getGroupsUnprotected(clientID string, filter model.GroupsFilter) ([]model.Group, error) {
	groups, err := app.storage.FindGroups(clientID, nil, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	return app.storage.FindManagedGroupConfigs(clientID)
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

func (app *Application) deleteManagedGroupConfig(id string, clientID string) error {
	return app.storage.DeleteManagedGroupConfig(id, clientID)
}

func (app *Application) getConfig(configType string, appID string, orgID string) (*model.Config, error) {
	return app.storage.FindConfig(configType, appID, orgID)
}
