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
	"groups/core/model"
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
func (a *Adapter) CreateCalendarEvent(currentAccountID string, event map[string]interface{}, appID string, orgID string) (map[string]interface{}, error) {

	type calendarRequest struct {
		Event            map[string]interface{} `json:"event"`
		CurrentAccountID string                 `json:"current_account_id"`
		AppID            string                 `json:"app_id"`
		OrgID            string                 `json:"org_id"`
	}

	body := calendarRequest{Event: event, CurrentAccountID: currentAccountID, AppID: appID, OrgID: orgID}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/bbs/events", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("CreateCalendarEvent:error creating event  request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

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

	var response map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("CreateCalendarEvent: unable to parse json: %s", err)
		return nil, fmt.Errorf("CreateCalendarEvent: unable to parse json: %s", err)
	}

	return response, nil
}

// GetGroupCalendarEvents gets calendar events for a group
func (a *Adapter) GetGroupCalendarEvents(currentAccountID string, eventIDs []string, appID string, orgID string, filter model.GroupEventFilter) (map[string]interface{}, error) {
	type filterType struct {
		Limit  *int64 `json:"limit,omitempty"`
		Offset *int64 `json:"offset,omitempty"`

		StartTimeAfter *int64 `json:"start_time_after,omitempty"`

		StartTimeBefore *int64 `json:"start_time_before,omitempty"`
	}
	type calendarRequest struct {
		Filter           filterType `json:"filter"`
		CurrentAccountID string     `json:"current_account_id"`
		AppID            string     `json:"app_id"`
		OrgID            string     `json:"org_id"`
	}

	body := calendarRequest{
		AppID:            appID,
		OrgID:            orgID,
		CurrentAccountID: currentAccountID,
		Filter: filterType{
			Limit:           filter.Limit,
			Offset:          filter.Offset,
			StartTimeBefore: filter.StartTimeBefore,
			StartTimeAfter:  filter.StartTimeAfter,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/bbs/events/load", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("GetGroupCalendarEvents:error creating event  request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("GetGroupCalendarEvents: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("GetGroupCalendarEvents: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("GetGroupCalendarEvents: error with response code != 200")
	}

	dataRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetGroupCalendarEvents: unable to read json: %s", err)
		return nil, fmt.Errorf("GetGroupCalendarEvents: unable to parse json: %s", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("GetGroupCalendarEvents: unable to parse json: %s", err)
		return nil, fmt.Errorf("GetGroupCalendarEvents: unable to parse json: %s", err)
	}

	return response, nil
}

// AddPeopleToCalendarEvent adds people calendar event
func (a *Adapter) AddPeopleToCalendarEvent(people []string, appID string, orgID string) error {

	type addPeopleRequest struct {
		People []string `json:"people"`
		AppID  string   `json:"app_id"`
		OrgID  string   `json:"org_id"`
	}

	body := addPeopleRequest{People: people, AppID: appID, OrgID: orgID}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/bbs/events/people/add", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("AddPeopleToCalendarEvent:error creating event  request - %s", err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("AddPeopleToCalendarEvent: error sending request - %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("AddPeopleToCalendarEvent: error with response code - %d", resp.StatusCode)
		return fmt.Errorf("AddPeopleToCalendarEvent: error with response code != 200")
	}

	dataRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("AddPeopleToCalendarEvent: unable to read json: %s", err)
		return fmt.Errorf("AddPeopleToCalendarEvent: unable to parse json: %s", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("AddPeopleToCalendarEvent: unable to parse json: %s", err)
		return fmt.Errorf("AddPeopleToCalendarEvent: unable to parse json: %s", err)
	}

	return nil
}
