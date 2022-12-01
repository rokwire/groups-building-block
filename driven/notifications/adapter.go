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

package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
)

// Adapter implements the Notifications interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// Recipient struct
type Recipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Mute   bool   `json:"mute"`
}

// NewNotificationsAdapter creates a new Notifications BB adapter instance
func NewNotificationsAdapter(baseURL string, serviceAccountManager *authservice.ServiceAccountManager) (*Adapter, error) {
	if serviceAccountManager == nil {
		log.Println("service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	return &Adapter{baseURL: baseURL, serviceAccountManager: serviceAccountManager}, nil
}

// SendNotification sends notification to a user
func (na *Adapter) SendNotification(recipients []Recipient, topic *string, title string, text string, data map[string]string, accountCriteria map[string]interface{}, appID string, orgID string) {
	na.sendNotification(recipients, topic, title, text, data, accountCriteria, appID, orgID)
}

func (na *Adapter) sendNotification(recipients []Recipient, topic *string, title string, text string, data map[string]string, accountCriteria map[string]interface{}, appID string, orgID string) error {
	if len(recipients) > 0 || len(accountCriteria) > 0 {
		url := fmt.Sprintf("%s/api/bbs/message", na.baseURL)

		async := true
		message := map[string]interface{}{
			"org_id":                     orgID,
			"app_id":                     appID,
			"priority":                   10,
			"recipients":                 recipients,
			"recipient_account_criteria": accountCriteria,
			"topic":                      topic,
			"subject":                    title,
			"body":                       text,
			"data":                       data,
		}
		bodyData := map[string]interface{}{
			"async":   async,
			"message": message,
		}
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			log.Printf("SendNotification::error creating notification request - %s", err)
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("SendNotification:error creating load user data request - %s", err)
			return err
		}

		resp, err := na.serviceAccountManager.MakeRequest(req, appID, orgID)
		if err != nil {
			log.Printf("SendNotification: error sending request - %s", err)
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("SendNotification: error with response code - %d", resp.StatusCode)
			return fmt.Errorf("SendNotification:error with response code != 200")
		}
	}
	return nil
}

// SendMail sends email to a user
func (na *Adapter) SendMail(toEmail string, subject string, body string) {
	go na.sendMail(toEmail, subject, body)
}

func (na *Adapter) sendMail(toEmail string, subject string, body string) error {
	if len(toEmail) > 0 && len(subject) > 0 && len(body) > 0 {
		url := fmt.Sprintf("%s/api/int/mail", na.baseURL)

		bodyData := map[string]interface{}{
			"to_mail": toEmail,
			"subject": subject,
			"body":    body,
		}
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			log.Printf("error creating notification request - %s", err)
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("error creating load user data request - %s", err)
			return err
		}

		resp, err := na.serviceAccountManager.MakeRequest(req, "all", "all")
		if err != nil {
			log.Printf("sendMail: error sending request - %s", err)
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("error with response code - %d", resp.StatusCode)
			return fmt.Errorf("error with response code != 200")
		}
	}
	return nil
}
