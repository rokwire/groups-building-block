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

package corebb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"groups/core/model"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the Core interface
type Adapter struct {
	coreURL               string
	serviceAccountManager *authservice.ServiceAccountManager
	logger                *logs.Logger
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(logger *logs.Logger, coreURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{logger: logger, coreURL: coreURL, serviceAccountManager: serviceAccountManager}
}

// RetrieveCoreUserAccount retrieves Core user account
func (a *Adapter) RetrieveCoreUserAccount(token string) (*model.CoreAccount, error) {
	if len(token) > 0 {
		url := fmt.Sprintf("%s/services/account", token)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error creating load user data request - %s", err)
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error loading user data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveCoreUserAccount: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: error with response code != 200")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: unable to parse json: %s", err)
		}

		var coreAccount model.CoreAccount
		err = json.Unmarshal(data, &coreAccount)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		return &coreAccount, nil
	}
	return nil, nil
}

// RetrieveCoreServices retrieves Core service registrations
func (a *Adapter) RetrieveCoreServices(serviceIDs []string) ([]model.CoreService, error) {
	if len(serviceIDs) > 0 {
		url := fmt.Sprintf("%s/bbs/service-regs?ids=%s", a.coreURL, strings.Join(serviceIDs, ","))
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveCoreServices: error creating load core service regs - %s", err)
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveCoreServices: error loading core service regs data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveCoreServices: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: error with response code != 200")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveCoreServices: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: unable to parse json: %s", err)
		}

		var coreServices []model.CoreService
		err = json.Unmarshal(data, &coreServices)
		if err != nil {
			log.Printf("RetrieveCoreServices: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreServices: unable to parse json: %s", err)
		}

		return coreServices, nil
	}
	return nil, nil
}

// GetAccountsCount retrieves account count for provided params
func (a *Adapter) GetAccountsCount(searchParams map[string]interface{}, appID *string, orgID *string) (int64, error) {
	if a.serviceAccountManager == nil {
		log.Println("GetAccountsCount: service account manager is nil")
		return 0, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/accounts/count", a.coreURL)
	queryString := ""
	if appID != nil {
		queryString += "?app_id=" + *appID
	}
	if orgID != nil {
		if queryString == "" {
			queryString += "?"
		} else {
			queryString += "&"
		}
		queryString += "org_id=" + *orgID
	}
	bodyBytes, err := json.Marshal(searchParams)
	if err != nil {
		log.Printf("GetAccountsCount: error marshalling body - %s", err)
		return 0, err
	}

	req, err := http.NewRequest("POST", url+queryString, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("GetAccountsCount: error creating request - %s", err)
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")

	appIDVal := "all"
	if appID != nil {
		appIDVal = *appID
	}
	orgIDVal := "all"
	if orgID != nil {
		appIDVal = *orgID
	}
	resp, err := a.serviceAccountManager.MakeRequest(req, appIDVal, orgIDVal)
	if err != nil {
		log.Printf("GetAccountsCount: error sending request - %s", err)
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("GetAccountsCount: error with response code - %d", resp.StatusCode)
		return 0, fmt.Errorf("GetAccountsCount: error with response code != 200")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetAccountsCount: unable to read json: %s", err)
		return 0, fmt.Errorf("GetAccountsCount: unable to parse json: %s", err)
	}

	var count int64
	err = json.Unmarshal(data, &count)
	if err != nil {
		log.Printf("GetAccountsCount: unable to parse json: %s", err)
		return 0, fmt.Errorf("GetAccountsCount: unable to parse json: %s", err)
	}

	return count, nil
}

// GetAllCoreAccountsWithExternalIDs Gets all Core accounts with external IDs
func (a *Adapter) GetAllCoreAccountsWithExternalIDs(externalIDs []string, appID *string, orgID *string) ([]model.CoreAccount, error) {
	var list []model.CoreAccount
	var limit int = 100
	var offset int = 0

	for {
		buffer, err := a.GetAccounts(map[string]interface{}{
			"external_ids.uin": externalIDs,
		}, appID, orgID, &limit, &offset)
		if err != nil {
			return nil, err
		}

		if len(buffer) == 0 {
			break
		} else {
			list = append(list, buffer...)
			offset += limit
		}
	}

	return list, nil
}

// GetAllCoreAccountsWithNetIDs Gets all Core accounts with net IDs
func (a *Adapter) GetAllCoreAccountsWithNetIDs(netIDs []string, appID *string, orgID *string) ([]model.CoreAccount, error) {
	var list []model.CoreAccount
	var limit int = 100
	var offset int = 0

	for {
		buffer, err := a.GetAccounts(map[string]interface{}{
			"external_ids.net_id": netIDs,
		}, appID, orgID, &limit, &offset)
		if err != nil {
			return nil, err
		}

		if len(buffer) == 0 {
			break
		} else {
			list = append(list, buffer...)
			offset += limit
		}
	}

	return list, nil
}

// GetAccountsWithIDs Gets all core accaunts with IDs
func (a *Adapter) GetAccountsWithIDs(ids []string, appID *string, orgID *string, limit *int, offset *int) ([]model.CoreAccount, error) {
	return a.GetAccounts(map[string]interface{}{
		"id": ids,
	}, appID, orgID, limit, offset)
}

// GetAccounts retrieves account count for provided params
func (a *Adapter) GetAccounts(searchParams map[string]interface{}, appID *string, orgID *string, limit *int, offset *int) ([]model.CoreAccount, error) {
	if a.serviceAccountManager == nil {
		log.Println("GetAccounts: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/accounts", a.coreURL)

	bodyBytes, err := json.Marshal(searchParams)
	if err != nil {
		log.Printf("GetAccounts: error marshalling body - %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("GetAccounts: error creating request - %s", err)
		return nil, err
	}

	params := req.URL.Query()
	if appID != nil {
		params.Add("app_id", *appID)
	}
	if orgID != nil {
		params.Add("org_id", *orgID)
	}
	if limit != nil {
		params.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if offset != nil {
		params.Add("offset", fmt.Sprintf("%d", *offset))
	}
	req.URL.RawQuery = params.Encode()

	req.Header.Add("Content-Type", "application/json")

	appIDVal := "all"
	if appID != nil {
		appIDVal = *appID
	}
	orgIDVal := "all"
	if orgID != nil {
		appIDVal = *orgID
	}
	resp, err := a.serviceAccountManager.MakeRequest(req, appIDVal, orgIDVal)
	if err != nil {
		log.Printf("GetAccounts: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("GetAccounts: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("GetAccounts: error with response code != 200")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetAccounts: unable to read json: %s", err)
		return nil, fmt.Errorf("GetAccounts: unable to parse json: %s", err)
	}

	var maping []model.CoreAccount
	err = json.Unmarshal(data, &maping)
	if err != nil {
		log.Printf("GetAccounts: unable to parse json: %s", err)
		return nil, fmt.Errorf("GetAccounts: unable to parse json: %s", err)
	}

	return maping, nil
}

// LoadDeletedMemberships loads deleted memberships
func (a *Adapter) LoadDeletedMemberships() ([]model.DeletedUserData, error) {

	if a.serviceAccountManager == nil {
		log.Println("LoadDeletedMemberships: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/deleted-memberships?service_id=%s", a.coreURL, a.serviceAccountManager.AuthService.ServiceID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		a.logger.Errorf("delete membership: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		log.Printf("LoadDeletedMemberships: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("LoadDeletedMemberships: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("LoadDeletedMemberships: error with response code != 200")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to read json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	var deletedMemberships []model.DeletedUserData
	err = json.Unmarshal(data, &deletedMemberships)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to parse json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	return deletedMemberships, nil
}

// RetrieveFerpaAccounts retrieves ferpa accounts
func (a *Adapter) RetrieveFerpaAccounts(ids []string) ([]string, error) {
	var allFerpaAccounts []string
	var batch []string
	var batchSize int

	// https://github.com/rokwire/groups-building-block/issues/542
	// This workaround on blind is to avoid the 4000 character limit on the URL
	// Otherwise it will fail due to query string being too long
	for _, id := range ids {
		if batchSize+len(id)+1 > 4000 {
			ferpaAccounts, err := a.retrieveFerpaAccounts(batch)
			if err != nil {
				return nil, err
			}
			allFerpaAccounts = append(allFerpaAccounts, ferpaAccounts...)
			batch = []string{}
			batchSize = 0
		}
		batch = append(batch, id)
		batchSize += len(id) + 1
	}

	if len(batch) > 0 {
		ferpaAccounts, err := a.retrieveFerpaAccounts(batch)
		if err != nil {
			return nil, err
		}
		allFerpaAccounts = append(allFerpaAccounts, ferpaAccounts...)
	}

	return allFerpaAccounts, nil
}

// RetrieveFerpaAccounts retrieves ferpa accounts
func (a *Adapter) retrieveFerpaAccounts(ids []string) ([]string, error) {
	if len(ids) > 0 {
		url := fmt.Sprintf("%s/bbs/accounts/ferpa?ids=%s", a.coreURL, strings.Join(ids, ","))

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			a.logger.Errorf("delete membership: error creating request - %s", err)
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
		if err != nil {
			log.Printf("RetrieveFerpaAccounts: error sending request - %s", err)
			return nil, err
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveFerpaAccounts: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveFerpaAccounts: error with response code != 200")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveFerpaAccounts: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveFerpaAccounts: unable to parse json: %s", err)
		}

		var ferpaAccountIDs []string
		err = json.Unmarshal(data, &ferpaAccountIDs)
		if err != nil {
			log.Printf("RetrieveFerpaAccounts: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveFerpaAccounts: unable to parse json: %s", err)
		}

		return ferpaAccountIDs, nil
	}
	return nil, nil
}
