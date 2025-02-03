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

import "groups/core/model"

func (app *Application) reloadPostsMigrationConfig() {
	var status = app.storage.FindPostMigrationConfig(nil)
	app.logger.Infof("reloaded posts migration config: old migration status: %v, new migration status: %v", app.postsMigrationConfig.Migrated, status.Migrated)
	app.postsMigrationConfig = status
}

func (app *Application) getPostsMigrationConfig() model.PostsMigrationConfig {
	return app.storage.FindPostMigrationConfig(nil)
}

func (app *Application) savePostsMigrationConfig(config model.PostsMigrationConfig) error {
	err := app.storage.SavePostsMigrationConfig(nil, config)
	if err != nil {
		app.logger.Errorf("failed to save posts migration config: %s", err)
		return err
	}
	app.postsMigrationConfig = config
	return nil
}
