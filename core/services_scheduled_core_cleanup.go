/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/storage"
)

func (app Application) processCoreAccountsCleanup() {
	app.logger.Infof("processCoreAccountsCleanup:BEGIN")
	defer app.logger.Infof("processCoreAccountsCleanup:END")

	//load deleted accounts
	deletedMemberships, err := app.corebb.LoadDeletedMemberships()
	if err != nil {
		app.logger.Errorf("error on loading deleted accounts - %s", err)
		return
	}
	fmt.Print(deletedMemberships)
	//process by app org
	for _, appOrgSection := range deletedMemberships {
		app.logger.Infof("delete - [app-id:%s org-id:%s]", appOrgSection.AppID, appOrgSection.OrgID)

		accountsIDs := app.getAccountsIDs(appOrgSection.Memberships)
		if len(accountsIDs) == 0 {
			app.logger.Info("no accounts for deletion")
			continue
		}

		app.logger.Infof("accounts for deletion - %s", accountsIDs)

		//delete the data
		app.deleteAppOrgUsersData(accountsIDs)
	}
}

func (app Application) deleteAppOrgUsersData(accountsIDs []string) {

	//in transaction
	errTr := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		// delete the group memberships
		err := app.storage.DeleteGroupMembershipsByAccountsIDs(nil, nil, accountsIDs)
		if err != nil {
			app.logger.Errorf("error deleting the group memberships by account ID - %s", err)
			return err
		}

		// delete users
		err = app.storage.DeleteUsersByAccountsIDs(nil, nil, accountsIDs)
		if err != nil {
			app.logger.Errorf("error deleting users by account ID - %s", err)
			return err
		}

		return nil
	})

	if errTr != nil {
		app.logger.Errorf("error deleting - %s", errTr)
		return
	}

	return
}

func (app Application) getAccountsIDs(memberships []model.DeletedMembership) []string {
	res := make([]string, len(memberships))
	for i, item := range memberships {
		res[i] = item.AccountID
	}
	return res
}
