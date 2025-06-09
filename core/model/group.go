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
	IsAbuse             *bool    `json:"is_abuse,omitempty" bson:"is_abuse,omitempty"`

	Settings   *GroupSettings         `json:"settings" bson:"settings"` // TODO: Remove the pointer once the backward support is not needed any more!
	Attributes map[string]interface{} `json:"attributes" bson:"attributes"`

	CurrentMember *GroupMembership `json:"current_member"` // this is indicative and it's not required for update APIs
	Members       []Member         `json:"members,omitempty" bson:"members,omitempty"`
	Stats         GroupStats       `json:"stats" bson:"stats"`

	DateCreated                  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated                  *time.Time `json:"date_updated" bson:"date_updated"`
	DatePostsUpdated             *time.Time `json:"date_posts_updated" bson:"date_posts_updated"`
	DateEventsUpdated            *time.Time `json:"date_events_updated" bson:"date_events_updated"`
	DatePollsUpdated             *time.Time `json:"date_polls_updated" bson:"date_polls_updated"`
	DateMembershipUpdated        *time.Time `json:"date_membership_updated" bson:"date_membership_updated"`
	DateManagedMembershipUpdated *time.Time `json:"date_managed_membership_updated" bson:"date_managed_membership_updated"`

	AuthmanEnabled             bool    `json:"authman_enabled" bson:"authman_enabled"`
	AuthmanGroup               *string `json:"authman_group" bson:"authman_group"`
	OnlyAdminsCanCreatePolls   bool    `json:"only_admins_can_create_polls" bson:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool    `json:"can_join_automatically" bson:"can_join_automatically"`
	BlockNewMembershipRequests bool    `json:"block_new_membership_requests" bson:"block_new_membership_requests"`
	AttendanceGroup            bool    `json:"attendance_group" bson:"attendance_group"`

	ResearchOpen             bool                           `json:"research_open" bson:"research_open"`
	ResearchGroup            bool                           `json:"research_group" bson:"research_group"`
	ResearchConsentStatement string                         `json:"research_consent_statement" bson:"research_consent_statement"`
	ResearchConsentDetails   string                         `json:"research_consent_details" bson:"research_consent_details"`
	ResearchDescription      string                         `json:"research_description" bson:"research_description"`
	ResearchProfile          map[string]map[string][]string `json:"research_profile" bson:"research_profile"`

	SyncStartTime  *time.Time `json:"sync_start_time" bson:"sync_start_time"`
	SyncEndTime    *time.Time `json:"sync_end_time" bson:"sync_end_time"`
	Administrative *bool      `json:"administrative"`
} // @name Group

// GetGroupMembershipsResponse response
type GetGroupMembershipsResponse struct {
	GroupID string `json:"group_id"`
	Title   string `json:"group_title"`
	Status  string `json:"status"`
} // @name GetGroupMembershipsResponse

// ApplyLegacyMembership applies legacy membership to the group for backward compatibility
func (gr *Group) ApplyLegacyMembership(membershipCollection MembershipCollection) {
	var list []Member
	for _, membership := range membershipCollection.Items {
		if membership.GroupID == gr.ID && (gr.CurrentMember != nil && (gr.CurrentMember.IsAdminOrMember() || membership.UserID == gr.CurrentMember.UserID)) {
			list = append(list, membership.ToMember())
		}
	}

	if len(list) > 1 {
		sort.SliceStable(list, func(p, q int) bool {
			if list[p].Status == list[q].Status {
				return list[p].Name < list[q].Name
			}
			return list[p].Status < list[q].Status
		})
	}

	gr.Members = list

}

// CreateMembershipEmptyAnswers creates membership empty answers list for the exact number of questions
func (gr *Group) CreateMembershipEmptyAnswers() []MemberAnswer {

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
func (gr *Group) IsAuthmanSyncEligible() bool {
	return gr.AuthmanEnabled && gr.AuthmanGroup != nil && *gr.AuthmanGroup != ""
}

// GetNewCategory gets new category attribute
func (gr *Group) GetNewCategory() *string {
	if gr.Attributes != nil {
		if val, ok := gr.Attributes["category"]; ok {
			category := val.(string)
			return &category
		}
	}
	return nil
}

// SetNewCategory Sets new category attribute
func (gr *Group) SetNewCategory(category string) {
	if gr.Attributes == nil {
		gr.Attributes = map[string]interface{}{}
	}
	gr.Attributes["category"] = category
}

// GetNewTags Gets new tags attribute
func (gr *Group) GetNewTags() []string {
	if gr.Attributes != nil {
		if val, ok := gr.Attributes["tags"]; ok {

			if interfaceList, ok := val.([]interface{}); ok {
				stringList := []string{}
				for _, v := range interfaceList {
					if stringValue, ok := v.(string); ok {
						stringList = append(stringList, stringValue)
					}
				}
				return stringList
			}
		}
	}
	return nil
}

// SetNewTags Sets new tags attribute
func (gr *Group) SetNewTags(tags []string) {
	if gr.Attributes == nil {
		gr.Attributes = map[string]interface{}{}
	}
	gr.Attributes["tags"] = tags
}
