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
	"groups/driven/storage"
)

func (app *Application) adminAddGroupMemberships(clientID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error {
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		membership, _ := app.storage.FindGroupMembershipWithContext(context, clientID, groupID, current.ID)

		if membership != nil && membership.IsAdmin() {

			group, err := app.storage.FindGroup(context, clientID, groupID, &current.ID)
			if err != nil {
				return err
			}

			netIDs := membershipStatuses.GetAllNetIDs()
			netIDAccounts, err := app.corebb.GetAllCoreAccountsWithNetIDs(netIDs, &current.AppID, &current.OrgID)
			if err != nil {
				return err
			}

			existingMemberships, err := app.storage.FindGroupMembershipsWithContext(context, clientID, model.MembershipFilter{
				GroupIDs: []string{groupID},
				NetIDs:   netIDs,
			})
			if err != nil {
				return err
			}

			var memberships []model.GroupMembership
			mapping := membershipStatuses.GetAllNetIDStatusMapping()
			if len(netIDAccounts) > 0 {
				for _, account := range netIDAccounts {
					if status, ok := mapping[account.GetNetID()]; ok {
						if existingMemberships.GetMembershipBy(func(membership model.GroupMembership) bool {
							return membership.NetID == account.GetNetID()
						}) == nil {
							memberships = append(memberships, account.ToMembership(groupID, status))
						}
					}
				}
				if len(memberships) > 0 {
					return app.storage.CreateMemberships(context, clientID, current, group, memberships)
				}
			}
		}

		return nil
	})

	return err
}

func (app *Application) adminDeleteMembershipsByID(clientID string, current *model.User, groupID string, accountIDs []string) error {

	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		membership, _ := app.storage.FindGroupMembershipWithContext(context, clientID, groupID, current.ID)

		if membership != nil && membership.IsAdmin() {

			err := app.storage.DeleteGroupMembershipsByAccountsIDs(app.logger, context, accountIDs)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
