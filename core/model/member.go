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
	"groups/driven/notifications"
	"time"

	"github.com/google/uuid"
)

// Member represents group member entity
type Member struct {
	ID            string         `json:"id" bson:"id"`
	UserID        string         `json:"user_id" bson:"user_id"`
	ExternalID    string         `json:"external_id" bson:"external_id"`
	Name          string         `json:"name" bson:"name"`
	NetID         string         `json:"net_id" bson:"net_id"`
	Email         string         `json:"email" bson:"email"`
	PhotoURL      string         `json:"photo_url" bson:"photo_url"`
	Status        string         `json:"status" bson:"status"` //pending, member, admin, rejected
	RejectReason  string         `json:"reject_reason" bson:"reject_reason"`
	MemberAnswers []MemberAnswer `json:"member_answers" bson:"member_answers"`

	DateCreated  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated  *time.Time `json:"date_updated" bson:"date_updated"`
	DateAttended *time.Time `json:"date_attended" bson:"date_attended"`
} //@name Member

// GetDisplayName Constructs a display name based on the current data state
func (m *Member) GetDisplayName() string {
	if len(m.Name) > 0 {
		return m.Name
	} else if len(m.Email) > 0 {
		return m.Email
	} else if len(m.ExternalID) > 0 {
		return m.ExternalID
	}
	return ""
}

// ApplyFromUserIfEmpty Copy info from the user entity
func (m *Member) ApplyFromUserIfEmpty(user *User) {
	if m.UserID == "" && user.ID != "" {
		m.UserID = user.ID
	}
	if m.ExternalID == "" && user.ExternalID != "" {
		m.ExternalID = user.ExternalID
	}
	if m.Email == "" && user.Email != "" {
		m.Email = user.Email
	}
	if m.Name == "" && user.Name != "" {
		m.Name = user.Name
	}
}

// ToMember construct ToMember based on the data
func (m *Member) ToMember() ToMember {
	return ToMember{
		UserID:     m.UserID,
		ExternalID: m.ExternalID,
		Name:       m.Name,
		Email:      m.Email,
	}
}

// ToNotificationRecipient construct notifications.Recipient based on the data
func (m *Member) ToNotificationRecipient() notifications.Recipient {
	return notifications.Recipient{
		UserID: m.UserID,
		Name:   m.Name,
	}
}

// ToGroupMembership converts the Member model to the GroupMembership model
func (m Member) ToGroupMembership(appID string, orgID string, groupID string) GroupMembership {
	return GroupMembership{
		ID:            uuid.NewString(),
		AppID:         appID,
		OrgID:         orgID,
		GroupID:       groupID,
		UserID:        m.UserID,
		ExternalID:    m.ExternalID,
		Name:          m.Name,
		NetID:         m.NetID,
		Email:         m.Email,
		PhotoURL:      m.PhotoURL,
		Status:        m.Status,
		RejectReason:  m.RejectReason,
		MemberAnswers: m.MemberAnswers,
		DateCreated:   m.DateCreated,
		DateUpdated:   m.DateUpdated,
		DateAttended:  m.DateAttended,
	}
}

// ToMember represents to(destination) member entity
type ToMember struct {
	UserID     string `json:"user_id" bson:"user_id"`
	ExternalID string `json:"external_id" bson:"external_id"`
	Name       string `json:"name" bson:"name"`
	Email      string `json:"email" bson:"email"`
} //@name ToMember

// MemberAnswer represents member answer entity
type MemberAnswer struct {
	Question string `json:"question" bson:"question"`
	Answer   string `json:"answer" bson:"answer"`
} //@name MemberAnswer

// IsAdmin says if the user is admin of the group
func (m *Member) IsAdmin() bool {
	return m.Status == "admin"
}

// IsAdminOrMember says if the user is admin or member of the group
func (m *Member) IsAdminOrMember() bool {
	return m.IsMember() || m.IsAdmin()
}

// IsMember says if the member is a group member
func (m *Member) IsMember() bool {
	return m.Status == "member"
}

// IsPendingMember says if the member is a group pending
func (m *Member) IsPendingMember() bool {
	return m.Status == "pending"
}

// IsRejected says if the member is a group rejected
func (m *Member) IsRejected() bool {
	return m.Status == "rejected"
}

// ToShortMemberRecord converts to ShortMemberRecord
func (m *Member) ToShortMemberRecord() ShortMemberRecord {
	return ShortMemberRecord{
		ID:         m.ID,
		UserID:     m.UserID,
		ExternalID: m.ExternalID,
		Email:      m.Email,
		NetID:      m.NetID,
		Name:       m.Name,
		Status:     m.Status,
	}
}

// ShortMemberRecord represents group short member entity only with important identifiers
type ShortMemberRecord struct {
	ID         string `json:"id" bson:"id"`
	UserID     string `json:"user_id" bson:"user_id"`
	ExternalID string `json:"external_id" bson:"external_id"`
	Name       string `json:"name" bson:"name"`
	NetID      string `json:"net_id" bson:"net_id"`
	Email      string `json:"email" bson:"email"`
	Status     string `json:"status" bson:"status"` //pending, member, admin, rejected
} //@name ShortMemberRecord
