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
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
)

// Adapter implements the Notifications interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *auth.ServiceAccountManager
}

// Recipient struct
type Recipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Mute   bool   `json:"mute"`
}

// NewNotificationsAdapter creates a new Notifications BB adapter instance
func NewNotificationsAdapter(baseURL string, serviceAccountManager *auth.ServiceAccountManager) (*Adapter, error) {
	if serviceAccountManager == nil {
		log.Println("service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	return &Adapter{baseURL: baseURL, serviceAccountManager: serviceAccountManager}, nil
}

// SendNotification sends notification to a user
func (na *Adapter) SendNotification(recipients []Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string, dateScheduled *time.Time) error {
	return na.sendNotification(recipients, topic, title, text, data, appID, orgID, dateScheduled)
}

func (na *Adapter) sendNotification(recipients []Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string, dateScheduled *time.Time) error {
	if len(recipients) > 0 {
		url := fmt.Sprintf("%s/api/bbs/message", na.baseURL)

		async := true
		message := map[string]interface{}{
			"org_id":     orgID,
			"app_id":     appID,
			"priority":   10,
			"recipients": recipients,
			"topic":      topic,
			"subject":    title,
			"body":       text,
			"data":       data,
		}
		if dateScheduled != nil {
			message["time"] = dateScheduled.Unix()
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

		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("SendNotification: unable to read response json: %s", err)
			return fmt.Errorf("SendNotification: unable to parse response json: %s", err)
		}

		if resp.StatusCode != 200 {
			log.Printf("SendNotification: error with response code - %d, Response: %s", resp.StatusCode, responseData)
			return fmt.Errorf("SendNotification:error with response code != 200")
		}
	}
	return nil
}

// SendMail sends email to a user
func (na *Adapter) SendMail(toEmail string, subject string, body string) error {
	return na.sendMail(toEmail, subject, body)
}

func (na *Adapter) sendMail(toEmail string, subject string, body string) error {
	if len(toEmail) > 0 && len(subject) > 0 && len(body) > 0 {
		url := fmt.Sprintf("%s/api/bbs/mail", na.baseURL)

		bodyData := map[string]interface{}{
			"to_mail": toEmail,
			"subject": subject,
			"body":    body,
		}
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			log.Printf("sendMail error creating notification request - %s", err)
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("sendMail error creating load user data request - %s", err)
			return err
		}

		resp, err := na.serviceAccountManager.MakeRequest(req, "all", "all")
		if err != nil {
			log.Printf("sendMail: error sending request - %s", err)
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			responseData, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("sendMail error: unable to read response json: %s", err)
				return fmt.Errorf("sendMail error: unable to parse response json: %s", err)
			}
			if responseData != nil {
				log.Printf("sendMail rror with response code - %d, response: %s", resp.StatusCode, responseData)
			} else {
				log.Printf("sendMail rror with response code - %d", resp.StatusCode)
			}
			return fmt.Errorf("sendMail error with response code != 200")
		}
	}
	return nil
}

// DeleteNotifications deletes notification
func (na *Adapter) DeleteNotifications(appID string, orgID string, ids string) error {

	url := fmt.Sprintf("%s/api/bbs/messages?ids=%s", na.baseURL, ids)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {

		log.Printf("DeleteNotification:error creating load user data request - %s", err)
		return err
	}

	resp, err := na.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("DeleteNotification: error sending request - %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("SendNotification: error with response code - %d", resp.StatusCode)
		return fmt.Errorf("DeleteNotification: error with response code != 200")
	}
	return nil
}

// AddNotificationRecipients Adds a new recipients in the notification
func (na *Adapter) AddNotificationRecipients(appID string, orgID string, notificationID string, userIDs []string) error {

	url := fmt.Sprintf("%s/api/bbs/messages/%s/recipients", na.baseURL, notificationID)

	type recipientType struct {
		Mute   bool   `json:"mute"`
		UserID string `json:"user_id"`
	}
	var recipients []recipientType
	for _, userID := range userIDs {
		recipients = append(recipients, recipientType{
			Mute:   false,
			UserID: userID,
		})
	}

	bodyBytes, err := json.Marshal(recipients)
	if err != nil {
		log.Printf("AddNotificationRecipient::error creating notification request - %s", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {

		log.Printf("AddNotificationRecipient:error creating load user data request - %s", err)
		return err
	}

	resp, err := na.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("AddNotificationRecipient: error sending request - %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("AddNotificationRecipient: error with response code - %d", resp.StatusCode)
		return fmt.Errorf("AddNotificationRecipient: error with response code != 200")
	}
	return nil
}

// RemoveNotificationRecipients Remove recipients from the notification
func (na *Adapter) RemoveNotificationRecipients(appID string, orgID string, notificationID string, userIDs []string) error {

	url := fmt.Sprintf("%s/api/bbs/messages/%s/recipients", na.baseURL, notificationID)

	type recipientType struct {
		Mute   bool   `json:"mute"`
		UserID string `json:"user_id"`
	}
	var recipients []recipientType
	for _, userID := range userIDs {
		recipients = append(recipients, recipientType{
			Mute:   false,
			UserID: userID,
		})
	}

	bodyBytes, err := json.Marshal(recipients)
	if err != nil {
		log.Printf("AddNotificationRecipient::error creating notification request - %s", err)
		return err
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewReader(bodyBytes))
	if err != nil {

		log.Printf("AddNotificationRecipient:error creating load user data request - %s", err)
		return err
	}

	resp, err := na.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		log.Printf("AddNotificationRecipient: error sending request - %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("AddNotificationRecipient: error with response code - %d", resp.StatusCode)
		return fmt.Errorf("AddNotificationRecipient: error with response code != 200")
	}
	return nil
}
