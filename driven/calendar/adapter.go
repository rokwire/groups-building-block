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
	"errors"
	"log"

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
func (a *Adapter) CreateCalendarEvent(adminIdentifiers []string, event string, orgID string, appID string) (string, error) {

	return "", nil
}

// UpdateCalendarEvent updates calendar event
func (a *Adapter) UpdateCalendarEvent(adminIdentifiers []string, event string, orgID string, appID string) (string, error) {

	return "", nil
}

// DeleteCalendarEvent deletes calendar event
func (a *Adapter) DeleteCalendarEvent(eventID string, orgID string, appID string) error {

	return nil
}
