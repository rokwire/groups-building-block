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

package authman

import (
	"encoding/json"
	"fmt"
	"groups/core/model"
	"log"
	"net/http"
	"strings"
)

// Adapter implements the Authman interface
type Adapter struct {
	authmanBaseURL  string
	authmanUsername string
	authmanPassword string
}

// SubjectsourceidUofinetid constant for using in authmanSubjectLookup
const SubjectsourceidUofinetid = "uofinetid"

// NewAuthmanAdapter creates a new adapter for Authman API
func NewAuthmanAdapter(authmanURL string, authmanUsername string, authmanPassword string) *Adapter {
	return &Adapter{authmanBaseURL: authmanURL, authmanUsername: authmanUsername, authmanPassword: authmanPassword}
}

// RetrieveAuthmanGroupMembers retrieves all members for a group
func (a *Adapter) RetrieveAuthmanGroupMembers(groupName string) ([]string, error) {
	if len(groupName) > 0 {
		url := fmt.Sprintf("%s/groups/%s/members", a.authmanBaseURL, groupName)
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

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembersError: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		var authmanData model.AuthmanGroupResponse
		err = json.Unmarshal(data, &authmanData)
		if err != nil {
			log.Printf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		response := []string{}
		for _, subjects := range authmanData.WsGetMembersLiteResult.WsSubjects {
			if subjects.SourceID == SubjectsourceidUofinetid {
				response = append(response, subjects.ID)
			}
		}

		return response, nil
	}
	return nil, nil
}

// AddAuthmanMemberToGroup add a member to an Authman group
func (a *Adapter) AddAuthmanMemberToGroup(groupName string, uin string) error {
	if len(groupName) > 0 {
		url := fmt.Sprintf("%s/groups/%s/members/%s", a.authmanBaseURL, groupName, uin)
		client := &http.Client{}
		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			log.Printf("AddAuthmanMemberToGroup: error creating load user data request - %s", err)
			return err
		}

		req.SetBasicAuth(a.authmanUsername, a.authmanPassword)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("AddAuthmanMemberToGroup: error loading user data - %s", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveAuthmanGroupMembersError: error with response code - %d", resp.StatusCode)
			return fmt.Errorf("AddAuthmanMemberToGroup: error with response code != 200")
		}

		return nil
	}
	return nil
}

// RemoveAuthmanMemberFromGroup remove a member from an Authman group
func (a *Adapter) RemoveAuthmanMemberFromGroup(groupName string, uin string) error {
	if len(groupName) > 0 {
		url := fmt.Sprintf("%s/groups/%s/members/%s", a.authmanBaseURL, groupName, uin)
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Printf("AddAuthmanMemberToGroup: error creating load user data request - %s", err)
			return err
		}

		req.SetBasicAuth(a.authmanUsername, a.authmanPassword)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("AddAuthmanMemberToGroup: error loading user data - %s", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveAuthmanGroupMembersError: error with response code - %d", resp.StatusCode)
			return fmt.Errorf("AddAuthmanMemberToGroup: error with response code != 200")
		}

		return nil
	}
	return nil
}

// RetrieveAuthmanUsers retrieve authman user data based on external IDs
func (a *Adapter) RetrieveAuthmanUsers(externalIDs []string) (map[string]model.AuthmanSubject, error) {
	externalIDCount := len(externalIDs)
	if externalIDCount > 0 {
		subjectLookups := make([]model.АuthmanSubjectLookup, externalIDCount)
		for i, externalID := range externalIDs {
			subjectLookups[i] = model.АuthmanSubjectLookup{
				SubjectID:       externalID,
				SubjectSourceID: SubjectsourceidUofinetid,
			}
		}

		requestBodyStruct := model.АuthmanUserRequest{
			WsRestGetSubjectsRequest: model.АuthmanUserData{
				WsSubjectLookups:      subjectLookups,
				SubjectAttributeNames: []string{"userprincipalname"},
			},
		}
		reqBody, err := json.Marshal(requestBodyStruct)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: marshal request body - %s", err)
			return nil, err
		}

		url := fmt.Sprintf("%s/subjects", a.authmanBaseURL)
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
			errordata, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("RetrieveAuthmanUsers: unable to read error response: %s", errordata)
				return nil, fmt.Errorf("RetrieveAuthmanUsers: unable to  error response: %s", errordata)
			}
			log.Printf("RetrieveAuthmanUsers: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
			return nil, fmt.Errorf("RetrieveAuthmanUsers: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveAuthmanUsers: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanUsers: unable to  read json: %s", err)
		}

		var authmanData model.АuthmanUserResponse
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

// RetrieveAuthmanStemGroups retrieve Authman user data based on external IDs for a given stem
func (a *Adapter) RetrieveAuthmanStemGroups(stemName string) (*model.АuthmanGroupsResponse, error) {

	// Hardcoded until it needs to be configurable
	requestBody := fmt.Sprintf(`{
		  "WsRestFindGroupsRequest":{
			"wsQueryFilter":{
			  "queryFilterType":"FIND_BY_STEM_NAME",
			  "stemName":"%s"
			}
		  }
		}`, stemName)

	url := fmt.Sprintf("%s/groups", a.authmanBaseURL)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(requestBody))
	if err != nil {
		log.Printf("RetrieveAuthmanStemGroups: error creating load user data request - %s", err)
		return nil, err
	}

	req.SetBasicAuth(a.authmanUsername, a.authmanPassword)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("RetrieveAuthmanStemGroups: error loading user data - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errordata, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveAuthmanStemGroups: unable to read error response: %s", errordata)
			return nil, fmt.Errorf("RetrieveAuthmanStemGroups: unable to  error response: %s", errordata)
		}
		log.Printf("RetrieveAuthmanStemGroups: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
		return nil, fmt.Errorf("RetrieveAuthmanStemGroups: error with response code - %d: Response: %s", resp.StatusCode, string(errordata))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("RetrieveAuthmanStemGroups: unable to read json: %s", err)
		return nil, fmt.Errorf("RetrieveAuthmanStemGroups: unable to  read json: %s", err)
	}

	var authmanData model.АuthmanGroupsResponse
	err = json.Unmarshal(data, &authmanData)
	if err != nil {
		log.Printf("RetrieveAuthmanStemGroups: unable to parse json: %s", err)
		return nil, fmt.Errorf("RetrieveAuthmanStemGroups: unable to parse json: %s", err)
	}

	return &authmanData, nil
}
