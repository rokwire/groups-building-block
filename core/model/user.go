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

import (
	"time"
)

// User represents user entity
type User struct {
	ID            string     `json:"id" bson:"_id"`
	AppID         string     `json:"app_id"`
	OrgID         string     `json:"org_id"`
	IsAnonymous   bool       `json:"is_anonymous" bson:"is_anonymous"`
	IsCoreUser    bool       `json:"is_core_user" bson:"is_core_user"`
	ExternalID    string     `json:"external_id" bson:"external_id"`
	NetID         string     `json:"net_id" bson:"net_id"`
	Email         string     `json:"email" bson:"email"`
	Name          string     `json:"name" bson:"name"`
	DateCreated   time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated   *time.Time `json:"date_updated" bson:"date_updated"`
	ClientID      string     `bson:"client_id"`
	Permissions   []string   `bson:"permissions"`
	OriginalToken string

	AuthType string
	IsBBUser bool
} // @name User

// ToCreator coverts to Creator
func (u *User) ToCreator() *Creator {
	return &Creator{
		UserID: u.ID,
		Name:   u.Name,
		Email:  u.Email,
	}
}

// HasPermission Checks if the user has desired permission
func (u *User) HasPermission(name string) bool {
	for _, permission := range u.Permissions {
		if permission == name {
			return true
		}
	}

	return false
}

// CoreAccount wraps the account structure from the Core BB
// @name CoreAccount
type CoreAccount struct {
	AuthTypes []struct {
		Active     bool   `json:"active"`
		Code       string `json:"code"`
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
		Params     struct {
			User struct {
				Email          string        `json:"email"`
				FirstName      string        `json:"first_name"`
				Groups         []interface{} `json:"groups"`
				Identifier     string        `json:"identifier"`
				LastName       string        `json:"last_name"`
				MiddleName     string        `json:"middle_name"`
				Roles          []string      `json:"roles"`
				SystemSpecific struct {
					PreferredUsername string `json:"preferred_username"`
				} `json:"system_specific"`
			} `json:"user"`
		} `json:"params"`
	} `json:"auth_types"`
	Groups      []interface{} `json:"groups"`
	ID          string        `json:"id"`
	Permissions []interface{} `json:"permissions"`
	Preferences struct {
		Favorites interface{} `json:"favorites"`
		Interests struct {
		} `json:"interests"`
		PrivacyLevel int      `json:"privacy_level"`
		Roles        []string `json:"roles"`
		Settings     struct {
		} `json:"settings"`
		Tags struct {
		} `json:"tags"`
		Voter struct {
			RegisteredVoter bool        `json:"registered_voter"`
			VotePlace       string      `json:"vote_place"`
			Voted           bool        `json:"voted"`
			VoterByMail     interface{} `json:"voter_by_mail"`
		} `json:"voter"`
	} `json:"preferences"`
	Profile struct {
		Address   string `json:"address"`
		BirthYear int    `json:"birth_year"`
		Country   string `json:"country"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		ID        string `json:"id"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		PhotoURL  string `json:"photo_url"`
		State     string `json:"state"`
		ZipCode   string `json:"zip_code"`
	} `json:"profile"`
	Roles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"roles"`
}

// GetExternalID Gets the external id
func (c *CoreAccount) GetExternalID() string {
	for _, auth := range c.AuthTypes {
		if auth.Active && len(auth.Identifier) > 0 {
			return auth.Identifier
		}
	}
	return ""
}

// GetNetID Gets the external id
func (c *CoreAccount) GetNetID() string {
	for _, auth := range c.AuthTypes {
		if auth.Active && len(auth.Identifier) > 0 {
			return auth.Params.User.SystemSpecific.PreferredUsername
		}
	}
	return ""
}

// GetFullName Builds the fullname
func (c *CoreAccount) GetFullName() string {
	var name string
	if len(c.Profile.FirstName) > 0 {
		name += c.Profile.FirstName
	}
	if len(c.Profile.LastName) > 0 {
		if len(name) > 0 {
			name += " "
		}
		name += c.Profile.LastName
	}
	return name
}

// ToMembership Builds the fullname
func (c *CoreAccount) ToMembership(status string) GroupMembership {
	return GroupMembership{
		UserID:      c.ID,
		ExternalID:  c.GetExternalID(),
		Name:        c.GetFullName(),
		Email:       c.Profile.Email,
		DateCreated: time.Now(),
	}
}

// DeletedUserData represents a user-deleted
type DeletedUserData struct {
	AppID       string              `json:"app_id"`
	Memberships []DeletedMembership `json:"memberships"`
	OrgID       string              `json:"org_id"`
}

// DeletedMembership defines model for DeletedMembership.
type DeletedMembership struct {
	AccountID string                  `json:"account_id"`
	Context   *map[string]interface{} `json:"context,omitempty"`
}
