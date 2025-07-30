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
)

func (app *Application) analyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error) {
	return app.storage.AnalyticsFindGroups(startDate, endDate)
}

func (app *Application) analyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error) {
	return app.storage.AnalyticsFindMembers(groupID, startDate, endDate)
}
