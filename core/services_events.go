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
)

func (app *Application) findAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error) {
	return app.storage.FindAdminGroupsForEvent(nil, clientID, current, eventID)
}

func (app *Application) updateGroupMappingsForEvent(clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	return app.storage.UpdateGroupMappingsForEvent(nil, clientID, current, eventID, groupIDs)
}

func (app *Application) findEventUserIDs(eventID string) ([]string, error) {
	return app.storage.FindEventUserIDs(nil, eventID)
}
