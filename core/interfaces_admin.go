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

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetGroups(OrgID string, current *model.User, filter model.GroupsFilter) (int64, []model.Group, error) {
	skipMembershipCheck := false
	if current != nil {
		skipMembershipCheck = current.IsGroupsBBAdministrator()
	}
	return s.app.getGroups(OrgID, current, filter, skipMembershipCheck)
}

func (s *administrationImpl) DeleteGroup(OrgID string, current *model.User, id string, inactive bool) error {
	return s.app.deleteGroup(OrgID, current, id, inactive)
}

func (s *administrationImpl) AdminDeleteMembershipsByID(OrgID string, current *model.User, groupID string, accountIDs []string) error {
	return s.app.adminDeleteMembershipsByID(OrgID, current, groupID, accountIDs)
}
