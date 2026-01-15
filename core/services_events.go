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
	"sync"
)

func (app *Application) findGroupMembershipsStatusAndGroupsTitle(userID string) ([]model.GetGroupMembershipsResponse, error) {
	return app.storage.FindGroupMembershipStatusAndGroupTitle(nil, userID)
}

func (app *Application) findGroupMembershipsByGroupID(groupID string) ([]string, error) {
	return app.storage.FindGroupMembershipByGroupID(nil, groupID)
}

func (app *Application) getUserData(userID string) (*model.UserDataResponse, error) {
	var wg sync.WaitGroup
	var groupMemberships []model.GroupMembership
	var groups []model.Group
	var eventsErr, membershipsErr, groupsErr, postsErr error

	// Fetch group memberships asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		groupMemberships, membershipsErr = app.storage.GetGroupMembershipByUserID(userID)
	}()

	// Wait for group memberships to be fetched, then fetch groups
	wg.Add(1)
	go func() {
		defer wg.Done()
		var groupIDs []string
		if groupMemberships != nil {
			for _, membership := range groupMemberships {
				groupIDs = append(groupIDs, membership.GroupID)
			}
		}
		groups, groupsErr = app.storage.FindGroupsByGroupIDs(groupIDs)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors from any of the goroutines
	if eventsErr != nil {
		return nil, eventsErr
	}
	if membershipsErr != nil {
		return nil, membershipsErr
	}
	if groupsErr != nil {
		return nil, groupsErr
	}
	if postsErr != nil {
		return nil, postsErr
	}

	// Prepare the response
	userData := &model.UserDataResponse{
		GroupMembershipsResponse: groupMemberships,
		GroupResponse:            groups,
	}

	return userData, nil
}

func (app *Application) findGroupsByGroupIDs(groupIDs []string) ([]model.Group, error) {
	return app.storage.FindGroupsByGroupIDs(groupIDs)
}
