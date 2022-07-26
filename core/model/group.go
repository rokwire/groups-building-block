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
)

// Group represents group entity
type Group struct {
	ID                  string   `json:"id" bson:"_id"`
	ClientID            string   `json:"client_id" bson:"client_id"`
	Category            string   `json:"category" bson:"category"` //one of the enums categories list
	Title               string   `json:"title" bson:"title"`
	Privacy             string   `json:"privacy" bson:"privacy"` //public or private
	HiddenForSearch     bool     `json:"hidden_for_search" bson:"hidden_for_search"`
	Description         *string  `json:"description" bson:"description"`
	ImageURL            *string  `json:"image_url" bson:"image_url"`
	WebURL              *string  `json:"web_url" bson:"web_url"`
	Tags                []string `json:"tags" bson:"tags"`
	MembershipQuestions []string `json:"membership_questions" bson:"membership_questions"`

	Members []Member `json:"members" bson:"members"`

	DateCreated                time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated                *time.Time `json:"date_updated" bson:"date_updated"`
	AuthmanEnabled             bool       `json:"authman_enabled" bson:"authman_enabled"`
	AuthmanGroup               *string    `json:"authman_group" bson:"authman_group"`
	OnlyAdminsCanCreatePolls   bool       `json:"only_admins_can_create_polls" bson:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool       `json:"can_join_automatically" bson:"can_join_automatically"`
	BlockNewMembershipRequests bool       `json:"block_new_membership_requests" bson:"block_new_membership_requests"`
	AttendanceGroup            bool       `json:"attendance_group" bson:"attendance_group"`
} // @name Group

// IsGroupAdminOrMember says if the user is an admin or a member of the group
func (gr Group) IsGroupAdminOrMember(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.UserID == userID && item.IsAdminOrMember() {
			return true
		}
	}
	return false
}

// IsGroupAdmin says if the user is admin of the group
func (gr Group) IsGroupAdmin(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.UserID == userID && item.IsAdmin() {
			return true
		}
	}
	return false
}

// IsGroupMember says if the user is a group member
func (gr Group) IsGroupMember(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.UserID == userID && item.IsMember() {
			return true
		}
	}
	return false
}

// IsGroupPending says if the user is a group pending
func (gr Group) IsGroupPending(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.UserID == userID && item.IsPendingMember() {
			return true
		}
	}
	return false
}

// IsGroupRejected says if the user is a group rejected
func (gr Group) IsGroupRejected(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.UserID == userID && item.IsRejected() {
			return true
		}
	}
	return false
}

// UserNameByID Get name of the user
func (gr Group) UserNameByID(userID string) *string {
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.UserID == userID {
			name := item.Name
			return &name
		}
	}
	return nil
}

// GetMemberByID says if the user is a group rejected
func (gr Group) GetMemberByID(userID string) *Member {
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.ID == userID {
			return &item
		}
	}
	return nil
}

// GetMemberByUserID gets member by UserID field
func (gr Group) GetMemberByUserID(userID string) *Member {
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.UserID == userID {
			return &item
		}
	}
	return nil
}

// GetMemberByExternalID gets member by ExternalID field
func (gr Group) GetMemberByExternalID(userID string) *Member {
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.ExternalID == userID {
			return &item
		}
	}
	return nil
}

// GetAllAdminMembers gets all admin members
func (gr Group) GetAllAdminMembers() []Member {
	return gr.GetMembersByStatus("admin")
}

// GetAllAdminsAsRecipients gets all admins as list of Recipient recipients
func (gr Group) GetAllAdminsAsRecipients() []notifications.Recipient {
	admins := gr.GetMembersByStatus("admin")

	var recipients []notifications.Recipient
	if len(admins) > 0 {
		for _, admin := range admins {
			recipients = append(recipients, admin.ToNotificationRecipient())
		}
	}

	return recipients
}

// GetMembersByStatus gets members by status field
func (gr Group) GetMembersByStatus(status string) []Member {
	var members []Member
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.Status == status {
			members = append(members, item)
		}
	}
	return members
}

// GetMembersAsNotificationRecipients constructs all official members as notification recipients
func (gr Group) GetMembersAsNotificationRecipients(skipUserID *string) []notifications.Recipient {

	recipients := []notifications.Recipient{}

	if len(gr.Members) > 0 {
		for _, member := range gr.Members {
			if member.IsAdminOrMember() && (skipUserID == nil || *skipUserID != member.UserID) {
				recipients = append(recipients, notifications.Recipient{
					UserID: member.UserID,
					Name:   member.Name,
				})
			}
		}
	}
	return recipients
}

// CreateMembershipEmptyAnswers creates membership empty answers list for the exact number of questions
func (gr Group) CreateMembershipEmptyAnswers() []MemberAnswer {

	var answers []MemberAnswer
	if len(gr.MembershipQuestions) > 0 {
		for _, question := range gr.MembershipQuestions {
			answers = append(answers, MemberAnswer{
				Question: question,
				Answer:   "",
			})
		}
	}

	return answers
}

// IsAuthmanSyncEligible Checks if the group has all required artefacts for an Authman Synchronization
func (gr Group) IsAuthmanSyncEligible() bool {
	return gr.AuthmanEnabled && gr.AuthmanGroup != nil && *gr.AuthmanGroup != ""
}
