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

// АuthmanUserRequest Authman user request wrapper
type АuthmanUserRequest struct {
	WsRestGetSubjectsRequest АuthmanUserData `json:"WsRestGetSubjectsRequest"`
}

// АuthmanUserData Authman user data
type АuthmanUserData struct {
	WsSubjectLookups      []АuthmanSubjectLookup `json:"wsSubjectLookups"`
	SubjectAttributeNames []string               `json:"subjectAttributeNames"`
}

// АuthmanSubjectLookup Authman subject lookup
type АuthmanSubjectLookup struct {
	SubjectID       string `json:"subjectId"`
	SubjectSourceID string `json:"subjectSourceId"`
}

// АuthmanUserResponse Authman user response wrapper
type АuthmanUserResponse struct {
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

// АuthmanGroupsResponse Authman groups response wrapper
type АuthmanGroupsResponse struct {
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

// GetGroupPettyTitleAndAdmins Gets the group pretty name and and group admin UINs
func (a *AuthmanGroupEntry) GetGroupPettyTitleAndAdmins() (string, []string) {
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
