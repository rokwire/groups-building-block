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

package polls

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
)

// Adapter implements the Notifications interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *auth.ServiceAccountManager
}

// NewPollsAdapter creates a new Polls BB adapter instance
func NewPollsAdapter(baseURL string, serviceAccountManager *auth.ServiceAccountManager) (*Adapter, error) {
	if serviceAccountManager == nil {
		log.Println("service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	return &Adapter{baseURL: baseURL, serviceAccountManager: serviceAccountManager}, nil
}

// DeleteGroupPolls deletes all polls for a group
func (a *Adapter) DeleteGroupPolls(orgID, groupID string) error {

	url := fmt.Sprintf("%s/api/bbs/group/%s/polls", a.baseURL, groupID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("DeleteGroupPolls:error creating event  request - %s", err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", orgID)
	if err != nil {
		log.Printf("DeleteGroupPolls: error sending request - %s", err)
		return err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("DeleteGroupPolls: unable to read response json: %s", err)
		return fmt.Errorf("DeleteGroupPolls: unable to parse response json: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("DeleteGroupPolls: error with response code - %d, Response: %s", resp.StatusCode, responseData)
		return fmt.Errorf("DeleteGroupPolls:error with response code != 200")
	}
	return nil
}
