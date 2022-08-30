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
	"sort"
	"time"
)

// GroupV2 removes members list and adds just the current user as a member
type GroupV2 struct {
	ID                         string     `json:"id" bson:"_id"`
	ClientID                   string     `json:"client_id" bson:"client_id"`
	Category                   string     `json:"category" bson:"category"` //one of the enums categories list
	Title                      string     `json:"title" bson:"title"`
	Privacy                    string     `json:"privacy" bson:"privacy"` //public or private
	HiddenForSearch            bool       `json:"hidden_for_search" bson:"hidden_for_search"`
	Description                *string    `json:"description" bson:"description"`
	ImageURL                   *string    `json:"image_url" bson:"image_url"`
	WebURL                     *string    `json:"web_url" bson:"web_url"`
	Tags                       []string   `json:"tags" bson:"tags"`
	MembershipQuestions        []string   `json:"membership_questions" bson:"membership_questions"`
	AuthmanEnabled             bool       `json:"authman_enabled" bson:"authman_enabled"`
	AuthmanGroup               *string    `json:"authman_group" bson:"authman_group"`
	OnlyAdminsCanCreatePolls   bool       `json:"only_admins_can_create_polls" bson:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool       `json:"can_join_automatically" bson:"can_join_automatically"`
	BlockNewMembershipRequests bool       `json:"block_new_membership_requests" bson:"block_new_membership_requests"`
	AttendanceGroup            bool       `json:"attendance_group" bson:"attendance_group"`
	CurrentMember              *Member    `json:"current_member"` // this is indicative and it's not required for update APIs
	DateCreated                time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated                *time.Time `json:"date_updated" bson:"date_updated"`
} // @name GroupV2

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

	CurrentMember *GroupMembership `json:"current_member"` // this is indicative and it's not required for update APIs
	Members       []Member         `json:"members" bson:"members"`

	DateCreated                time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated                *time.Time `json:"date_updated" bson:"date_updated"`
	AuthmanEnabled             bool       `json:"authman_enabled" bson:"authman_enabled"`
	AuthmanGroup               *string    `json:"authman_group" bson:"authman_group"`
	OnlyAdminsCanCreatePolls   bool       `json:"only_admins_can_create_polls" bson:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool       `json:"can_join_automatically" bson:"can_join_automatically"`
	BlockNewMembershipRequests bool       `json:"block_new_membership_requests" bson:"block_new_membership_requests"`
	AttendanceGroup            bool       `json:"attendance_group" bson:"attendance_group"`

	SyncStartTime *time.Time `json:"sync_start_time" bson:"sync_start_time"`
	SyncEndTime   *time.Time `json:"sync_end_time" bson:"sync_end_time"`
} // @name Group

func (gr *Group) ApplyLegacyMembership(membershipCollection MembershipCollection) {
	var list []Member
	for _, membership := range membershipCollection.Items {
		if membership.GroupID == gr.ID {
			list = append(list, membership.ToMember())
		}
	}

	sort.SliceStable(list, func(p, q int) bool {
		if list[p].Status == list[q].Status {
			return list[p].Name < list[q].Name
		}
		return list[p].Status < list[q].Status
	})

	gr.Members = list
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

// IsAuthmanSyncEligible Checks if the group has all required artefacts for an Authman Synchronization
func (gr Group) IsAuthmanSyncEligible() bool {
	return gr.AuthmanEnabled && gr.AuthmanGroup != nil && *gr.AuthmanGroup != ""
}

// ToGroupV2 Converts the legacy data model to a GroupV2 model
func (gr Group) ToGroupV2(currentUserID *string) GroupV2 {

	var currentMember *Member
	if currentUserID != nil {
		currentMember = gr.GetMemberByUserID(*currentUserID)
	}

	return GroupV2{
		ID:                         gr.ID,
		ClientID:                   gr.ClientID,
		Category:                   gr.Category,
		Title:                      gr.Title,
		Privacy:                    gr.Privacy,
		HiddenForSearch:            gr.HiddenForSearch,
		Description:                gr.Description,
		ImageURL:                   gr.ImageURL,
		WebURL:                     gr.WebURL,
		Tags:                       gr.Tags,
		MembershipQuestions:        gr.MembershipQuestions,
		AuthmanEnabled:             gr.AuthmanEnabled,
		AuthmanGroup:               gr.AuthmanGroup,
		OnlyAdminsCanCreatePolls:   gr.OnlyAdminsCanCreatePolls,
		CanJoinAutomatically:       gr.CanJoinAutomatically,
		BlockNewMembershipRequests: gr.BlockNewMembershipRequests,
		AttendanceGroup:            gr.AttendanceGroup,
		DateCreated:                gr.DateCreated,
		DateUpdated:                gr.DateUpdated,
		CurrentMember:              currentMember,
	}
}
