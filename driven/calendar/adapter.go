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

package calendar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
)

// Adapter implements the Notifications interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// NewCalendarAdapter creates a new Calendar BB adapter instance
func NewCalendarAdapter(baseURL string, serviceAccountManager *authservice.ServiceAccountManager) (*Adapter, error) {
	if serviceAccountManager == nil {
		log.Println("service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	return &Adapter{baseURL: baseURL, serviceAccountManager: serviceAccountManager}, nil
}

// CreateCalendarEvent creates calendar event
func (a *Adapter) CreateCalendarEvent(currentAccountID string, event string, orgID string, appID string) ([]map[string]interface{}, error) {
	type calendarRequest struct {
		event            string `json:"event"`
		currentAccountID string `json:"current_account_id"`
	}

	body := calendarRequest{event: event, currentAccountID: currentAccountID}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/admin/event", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("CreateCalendarEvent:error creating event  request - %s", err)
		return nil, err
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("CreateCalendarEvent: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("CreateCalendarEvent: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("CreateCalendarEvent: error with response code != 200")
	}

	dataRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("CreateCalendarEvent: unable to read json: %s", err)
		return nil, fmt.Errorf("CreateCalendarEvent: unable to parse json: %s", err)
	}

	var response []map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("CreateCalendarEvent: unable to parse json: %s", err)
		return nil, fmt.Errorf("CreateCalendarEvent: unable to parse json: %s", err)
	}

	return response, nil
}

// UpdateCalendarEvent updates calendar event
func (a *Adapter) UpdateCalendarEvent(adminIdentifiers []string, eventID string, event string, orgID string, appID string) ([]map[string]interface{}, error) {
	type calendarRequest struct {
		event            string   `json:"event"`
		adminIdentifiers []string `json:"admins_identifiers"`
	}

	body := calendarRequest{event: event, adminIdentifiers: adminIdentifiers}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/admin/event/%s", a.baseURL, eventID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("UpdateCalendarEvent:error updating event  request - %s", err)
		return nil, err
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("UpdateCalendarEvent: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("UpdateCalendarEvent: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("UpdateCalendarEvent: error with response code != 200")
	}

	dataRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("UpdateCalendarEvent: unable to read json: %s", err)
		return nil, fmt.Errorf("UpdateCalendarEvent: unable to parse json: %s", err)
	}

	var response []map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("UpdateCalendarEvent: unable to parse json: %s", err)
		return nil, fmt.Errorf("UpdateCalendarEvent: unable to parse json: %s", err)
	}

	return response, nil
}

// DeleteCalendarEvent deletes calendar event
func (a *Adapter) DeleteCalendarEvent(eventID string, orgID string, appID string) error {
	url := fmt.Sprintf("%s/api/bbs/messages?id=%s", a.baseURL, eventID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("DeleteCalendarEvent:error creating load user data request - %s", err)
		return err
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("DeleteCalendarEvent: error sending request - %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("DeleteCalendarEvent: error with response code - %d", resp.StatusCode)
		return fmt.Errorf("DeleteCalendarEvent: error with response code != 200")
	}
	return nil
}
