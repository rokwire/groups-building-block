package authman

import (
	"encoding/json"
	"fmt"
	"groups/core/model"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

		req.SetBasicAuth(a.authmanUsername, a.authmanPassword)

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

		var authmanData authmanGroupResponse
		err = json.Unmarshal(data, &authmanData)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		response := []string{}
		for _, subjects := range authmanData.WsGetMembersLiteResult.WsSubjects {
			response = append(response, subjects.ID)
		}

		return response, nil
	}
	return nil, nil
}

type authmanGroupResponse struct {
	WsGetMembersLiteResult struct {
		ResultMetadata struct {
			Success       string `json:"success"`
			ResultCode    string `json:"resultCode"`
			ResultMessage string `json:"resultMessage"`
		} `json:"resultMetadata"`
		WsGroup struct {
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
		WsSubjects []struct {
			SourceID   string `json:"sourceId"`
			Success    string `json:"success"`
			ResultCode string `json:"resultCode"`
			ID         string `json:"id"`
			MemberID   string `json:"memberId"`
		} `json:"wsSubjects"`
	} `json:"WsGetMembersLiteResult"`
}

// RetrieveAuthmanUsers retrieve authman user data based on external IDs
func (a *Adapter) RetrieveAuthmanUsers(externalIDs []string) (map[string]model.AuthmanSubject, error) {
	externalIDCount := len(externalIDs)
	if externalIDCount > 0 {
		subjectLookups := make([]authmanSubjectLookup, externalIDCount)
		for i, externalID := range externalIDs {
			subjectLookups[i] = authmanSubjectLookup{
				SubjectID: externalID,
			}
		}

		requestBodyStruct := authmanUserRequest{
			WsRestGetSubjectsRequest: authmanUserData{
				WsSubjectLookups:      subjectLookups,
				SubjectAttributeNames: []string{"userprincipalname"},
			},
		}
		reqBody, err := json.Marshal(requestBodyStruct)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: marshal request body - %s", err)
			return nil, err
		}

		url := "https://authman.illinois.edu/grouper-ws/servicesRest/v2_5_000/subjects"
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, strings.NewReader(string(reqBody)))
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: error creating load user data request - %s", err)
			return nil, err
		}

		req.SetBasicAuth(a.authmanUsername, a.authmanPassword)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: error loading user data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errordata, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("RetrieveAuthmanUsers: unable to read error response: %s", errordata)
				return nil, fmt.Errorf("RetrieveAuthmanUsers: unable to  error response: %s", errordata)
			}
			log.Printf("RetrieveAuthmanUsers: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
			return nil, fmt.Errorf("RetrieveAuthmanUsers: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanUsers: unable to  read json: %s", err)
		}

		var authmanData authmanUserResponse
		err = json.Unmarshal(data, &authmanData)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanUsers: unable to parse json: %s", err)
		}

		res := map[string]model.AuthmanSubject{}
		for _, subject := range authmanData.WsGetSubjectsResults.WsSubjects {
			res[subject.ID] = subject
		}
		return res, nil
	}
	return nil, nil
}

type authmanUserRequest struct {
	WsRestGetSubjectsRequest authmanUserData `json:"WsRestGetSubjectsRequest"`
}

type authmanUserData struct {
	WsSubjectLookups      []authmanSubjectLookup `json:"wsSubjectLookups"`
	SubjectAttributeNames []string               `json:"subjectAttributeNames"`
}

type authmanSubjectLookup struct {
	SubjectID string `json:"subjectId"`
}

type authmanUserResponse struct {
	WsGetSubjectsResults struct {
		ResultMetadata struct {
			Success       string `json:"success"`
			ResultCode    string `json:"resultCode"`
			ResultMessage string `json:"resultMessage"`
		} `json:"resultMetadata"`
		SubjectAttributeNames []string `json:"subjectAttributeNames"`
		ResponseMetadata      struct {
			ServerVersion string `json:"serverVersion"`
			Millis        string `json:"millis"`
		} `json:"responseMetadata"`
		WsSubjects []model.AuthmanSubject `json:"wsSubjects"`
	} `json:"WsGetSubjectsResults"`
}
