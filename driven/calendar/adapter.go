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
	"io"
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
func (a *Adapter) CreateCalendarEvent(adminIdentifier []model.AccountIdentifiers, currentAccountIdentifier model.AccountIdentifiers, event map[string]interface{}, orgID string, appID string, groupIDs []string) (map[string]interface{}, error) {

	type calendarRequest struct {
		AdminsIdentifiers         []model.AccountIdentifiers `json:"admins_identifiers"`
		Event                     map[string]interface{}     `json:"event"`
		CurrentAccountIdentifiers model.AccountIdentifiers   `json:"current_account_identifiers"`
		AppID                     string                     `json:"app_id"`
		OrgID                     string                     `json:"org_id"`
		GroupIDs                  []string                   `json:"group_ids"`
	}

	body := calendarRequest{AdminsIdentifiers: adminIdentifier, Event: event, CurrentAccountIdentifiers: currentAccountIdentifier, AppID: appID, OrgID: orgID, GroupIDs: groupIDs}
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
	}

	dataRes, err := io.ReadAll(resp.Body)
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

	if resp.StatusCode != 200 {
		log.Printf("UpdateCalendarEvent: error with response code - %d", resp.StatusCode)
		err = fmt.Errorf("%s", dataRes)
	}

	return response, err
}

// UpdateCalendarEvent updates calendar event
func (a *Adapter) UpdateCalendarEvent(currentAccountIdentifier model.AccountIdentifiers, eventID string, event map[string]interface{}, orgID string, appID string) (map[string]interface{}, error) {

	type calendarRequest struct {
		EventID                   string                   `json:"event_id"`
		Event                     map[string]interface{}   `json:"event"`
		CurrentAccountIdentifiers model.AccountIdentifiers `json:"current_account_identifiers"`
		AppID                     string                   `json:"app_id"`
		OrgID                     string                   `json:"org_id"`
	}

	body := calendarRequest{EventID: eventID, Event: event, CurrentAccountIdentifiers: currentAccountIdentifier, AppID: appID, OrgID: orgID}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/bbs/events", a.baseURL)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("UpdateCalendarEvent:error creating event  request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("UpdateCalendarEvent: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	dataRes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("UpdateCalendarEvent: unable to read json: %s", err)
		return nil, fmt.Errorf("UpdateCalendarEvent: unable to parse json: %s", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(dataRes, &response)
	if err != nil {
		log.Printf("UpdateCalendarEvent: unable to parse json: %s", err)
		return nil, fmt.Errorf("UpdateCalendarEvent: unable to parse json: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("UpdateCalendarEvent: error with response code - %d", resp.StatusCode)
		err = fmt.Errorf("%s", dataRes)
	}

	return response, err
}

// GetGroupCalendarEvents gets calendar events for a group
func (a *Adapter) GetGroupCalendarEvents(currentAccountIdentifier model.AccountIdentifiers, eventIDs []string, appID string, orgID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error) {
	type filterType struct {
		IDs       []string `json:"ids"`
		Limit     *int64   `json:"limit,omitempty"`
		Offset    *int64   `json:"offset,omitempty"`
		Published *bool    `json:"published"`

		StartTimeAfter             *int64 `json:"start_time_after,omitempty"`
		StartTimeAfterNullEndTime  *int64 `json:"start_time_after_null_end_time,omitempty"`
		StartTimeBefore            *int64 `json:"start_time_before,omitempty"`
		StartTimeBeforeNullEndTime *int64 `json:"start_time_before_null_end_time,omitempty"`
		EndTimeAfter               *int64 `json:"end_time_after,omitempty"`
		EndTimeBefore              *int64 `json:"end_time_before,omitempty"`
	}
	type calendarRequest struct {
		Filter                    filterType               `json:"filter"`
		CurrentAccountIdentifiers model.AccountIdentifiers `json:"current_account_identifiers"`
		AppID                     string                   `json:"app_id"`
		OrgID                     string                   `json:"org_id"`
	}

	body := calendarRequest{
		AppID:                     appID,
		OrgID:                     orgID,
		CurrentAccountIdentifiers: currentAccountIdentifier,
		Filter: filterType{
			IDs:                        eventIDs,
			Published:                  published,
			Limit:                      filter.Limit,
			Offset:                     filter.Offset,
			StartTimeBefore:            filter.StartTimeBefore,
			StartTimeBeforeNullEndTime: filter.StartTimeBeforeNullEndTime,
			StartTimeAfter:             filter.StartTimeAfter,
			StartTimeAfterNullEndTime:  filter.StartTimeAfterNullEndTime,
			EndTimeBefore:              filter.EndTimeBefore,
			EndTimeAfter:               filter.EndTimeAfter,
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

	dataRes, err := io.ReadAll(resp.Body)
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

	if resp.StatusCode != 200 {
		log.Printf("UpdateCalendarEvent: error with response code - %d", resp.StatusCode)
		err = fmt.Errorf("%s", dataRes)
	}

	return response, err
}

// AddPeopleToCalendarEvent adds people calendar event
func (a *Adapter) AddPeopleToCalendarEvent(people []string, eventID string, orgID string, appID string) error {

	type addPeopleRequest struct {
		People  []string `json:"people"`
		AppID   string   `json:"app_id"`
		OrgID   string   `json:"org_id"`
		EventID string   `json:"event_id"`
	}

	body := addPeopleRequest{People: people, AppID: appID, OrgID: orgID, EventID: eventID}
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

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("AddPeopleToCalendarEvent: unable to read response json: %s", err)
		return fmt.Errorf("AddPeopleToCalendarEvent: unable to parse response json: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("AddPeopleToCalendarEvent: error with response code - %d, Response: %s", resp.StatusCode, responseData)
		return fmt.Errorf("AddPeopleToCalendarEvent:error with response code != 200")
	}
	return nil
}

// RemovePeopleFromCalendarEvent adds people calendar event
func (a *Adapter) RemovePeopleFromCalendarEvent(people []string, eventID string, orgID string, appID string) error {

	type removePeopleRequest struct {
		People  []string `json:"people"`
		AppID   string   `json:"app_id"`
		OrgID   string   `json:"org_id"`
		EventID string   `json:"event_id"`
	}

	body := removePeopleRequest{People: people, AppID: appID, OrgID: orgID, EventID: eventID}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/bbs/events/people/remove", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("RemovePeopleFromCalendarEvent:error creating event  request - %s", err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("RemovePeopleFromCalendarEvent: error sending request - %s", err)
		return err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("RemovePeopleFromCalendarEventt: unable to read response json: %s", err)
		return fmt.Errorf("RemovePeopleFromCalendarEvent: unable to parse response json: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("RemovePeopleFromCalendarEvent: error with response code - %d, Response: %s", resp.StatusCode, responseData)
		return fmt.Errorf("RemovePeopleFromCalendarEvent:error with response code != 200")
	}
	return nil
}
