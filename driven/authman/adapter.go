package authman

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Adapter implements the Storage interface
type Adapter struct {
	authmanURL      string
	authmanUsername string
	authmanPassword string
}

// NewAuthmanAdapter creates a new adapter for Authman API
func NewAuthmanAdapter(authmanURL string, authmanUsername string, authmanPassword string) *Adapter {
	return &Adapter{authmanURL: authmanURL, authmanUsername: authmanUsername, authmanPassword: authmanPassword}
}

// RetrieveAuthmanGroupMembers retrieves all members for a group
func (a *Adapter) RetrieveAuthmanGroupMembers(groupName string) ([]string, error) {
	if len(groupName) > 0 {
		url := fmt.Sprintf(a.authmanURL, groupName)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembers: error creating load user data request - %s", err)
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembers: error loading user data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveAuthmanGroupMembersError: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: error with response code != 200")
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembersError: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		var authmanData authmanResponse
		err = json.Unmarshal(data, &authmanData)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		response := []string{}
		for _, subjects := range authmanData.wsGetMembersLiteResult.wsSubjects {
			response = append(response, subjects.ID)
		}

		return response, nil
	}
	return nil, nil
}

type authmanResponse struct {
	wsGetMembersLiteResult struct {
		ResultMetadata struct {
			Success       string `json:"success"`
			ResultCode    string `json:"resultCode"`
			ResultMessage string `json:"resultMessage"`
		} `json:"resultMetadata"`
		wsGroup struct {
			Extension        string `json:"extension"`
			DisplayName      string `json:"displayName"`
			Description      string `json:"description"`
			UUID             string `json:"uuid"`
			Enabled          string `json:"enabled"`
			DisplayExtension string `json:"displayExtension"`
			Name             string `json:"name"`
			TypeOfGroup      string `json:"typeOfGroup"`
			IDIndex          string `json:"idIndex"`
		} `json:"wsGroup"`
		ResponseMetadata struct {
			ServerVersion string `json:"serverVersion"`
			Millis        string `json:"millis"`
		} `json:"responseMetadata"`
		wsSubjects []struct {
			SourceID   string `json:"sourceId"`
			Success    string `json:"success"`
			ResultCode string `json:"resultCode"`
			ID         string `json:"id"`
			MemberID   string `json:"memberId"`
		} `json:"wsSubjects"`
	} `json:"WsGetMembersLiteResult"`
}
