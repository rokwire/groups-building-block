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
)

// Adapter implements the Core interface
type Adapter struct {
	coreURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager}
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

// GetAccounts retrieves account count for provided params
func (a *Adapter) GetAccounts(searchParams map[string]interface{}, appID *string, orgID *string, limit int, offset int, allAccess bool, approvedKeys []string) ([]map[string]interface{}, error) {
	if a.serviceAccountManager == nil {
		log.Println("GetAccounts: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/accounts", a.coreURL)
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
		log.Printf("GetAccounts: error marshalling body - %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url+queryString, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("GetAccounts: error creating request - %s", err)
		return nil, err
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

	var maping []map[string]interface{}
	err = json.Unmarshal(data, &maping)
	if err != nil {
		log.Printf("GetAccounts: unable to parse json: %s", err)
		return nil, fmt.Errorf("GetAccounts: unable to parse json: %s", err)
	}

	return maping, nil
}
