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

package model

import "strings"

// AuthmanSubject contains user name and user email
type AuthmanSubject struct {
	SourceID        string   `json:"sourceId"`
	Success         string   `json:"success"`
	AttributeValues []string `json:"attributeValues"`
	Name            string   `json:"name"`
	ResultCode      string   `json:"resultCode"`
	ID              string   `json:"id"`
} // @name AuthmanSubject

// AuthmanGroupResponse Authman group response wrapper
type AuthmanGroupResponse struct {
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

// AuthmanUserRequest Authman user request wrapper
type AuthmanUserRequest struct {
	WsRestGetSubjectsRequest AuthmanUserData `json:"WsRestGetSubjectsRequest"`
}

// AuthmanUserData Authman user data
type AuthmanUserData struct {
	WsSubjectLookups      []AuthmanSubjectLookup `json:"wsSubjectLookups"`
	SubjectAttributeNames []string               `json:"subjectAttributeNames"`
}

// AuthmanSubjectLookup Authman subject lookup
type AuthmanSubjectLookup struct {
	SubjectID       string `json:"subjectId"`
	SubjectSourceID string `json:"subjectSourceId"`
}

// AuthmanUserResponse Authman user response wrapper
type AuthmanUserResponse struct {
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
		WsSubjects []AuthmanSubject `json:"wsSubjects"`
	} `json:"WsGetSubjectsResults"`
}

// AuthmanGroupsResponse Authman groups response wrapper
type AuthmanGroupsResponse struct {
	WsFindGroupsResults struct {
		GroupResults   []AuthmanGroupEntry `json:"groupResults"`
		ResultMetadata struct {
			Success       string `json:"success"`
			ResultCode    string `json:"resultCode"`
			ResultMessage string `json:"resultMessage"`
		} `json:"resultMetadata"`
		ResponseMetadata struct {
			ServerVersion string `json:"serverVersion"`
			Millis        string `json:"millis"`
		} `json:"responseMetadata"`
	} `json:"WsFindGroupsResults"`
}

// AuthmanGroupEntry wrapper for single group entry
type AuthmanGroupEntry struct {
	Extension        string `json:"extension"`
	DisplayName      string `json:"displayName"`
	UUID             string `json:"uuid"`
	Description      string `json:"description"`
	Enabled          string `json:"enabled"`
	DisplayExtension string `json:"displayExtension"`
	Name             string `json:"name"`
	TypeOfGroup      string `json:"typeOfGroup"`
	IDIndex          string `json:"idIndex"`
}

// HasDescription checks if the group entry has a description
func (a *AuthmanGroupEntry) HasDescription() bool {
	return a.Description != ""
}

// GetGroupPrettyTitleAndAdmins Gets the group pretty name and and group admin UINs
func (a *AuthmanGroupEntry) GetGroupPrettyTitleAndAdmins() (string, []string) {
	if strings.Contains(a.Description, "|") {
		var first string
		var adminUINs []string
		segments := strings.Split(a.Description, "|")
		if len(segments) > 1 {
			for index, segment := range segments {
				if index == 0 {
					first = strings.ReplaceAll(segment, "\"", "")
					first = strings.TrimSpace(first)
				} else {
					adminUINs = append(adminUINs, strings.ReplaceAll(segment, " ", ""))
				}
			}
			return first, adminUINs
		}
	}

	if a.Description != "" {
		return a.Description, nil
	}
	return a.DisplayExtension, nil
}
