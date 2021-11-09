package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Adapter implements the Storage interface
type Adapter struct {
	internalAPIKey string
	baseURL        string
}

// Recipient struct
type Recipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// NewNotificationsAdapter creates a new Notifications BB adapter instance
func NewNotificationsAdapter(internalAPIKey string, baseURL string) *Adapter {
	return &Adapter{internalAPIKey: internalAPIKey, baseURL: baseURL}
}

// SendNotification sends notification to a user
func (na *Adapter) SendNotification(recipients []Recipient, topic *string, title string, text string, data map[string]string) error {
	url := fmt.Sprintf("%s/api/int/message", na.baseURL)

	bodyData := map[string]interface{}{
		"priority":   10,
		"recipients": recipients,
		"topic":      topic,
		"subject":    title,
		"body":       text,
		"data":       data,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Printf("error creating notification request - %s", err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("error creating load user data request - %s", err)
		return err
	}
	req.Header.Set("INTERNAL-API-KEY", na.internalAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error loading user data - %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error with response code - %d", resp.StatusCode)
		return fmt.Errorf("error with response code != 200")
	}

	return nil
}
