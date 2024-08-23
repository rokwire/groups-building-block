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

// Event represents event entity
type Event struct {
	ClientID      string     `json:"client_id" bson:"client_id"`
	EventID       string     `json:"event_id" bson:"event_id"`
	GroupID       string     `json:"group_id" bson:"group_id"`
	DateCreated   time.Time  `json:"date_created" bson:"date_created"`
	Creator       *Creator   `json:"creator" bson:"creator"`
	ToMembersList []ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
} // @name Event

// AccountIdentifiers represents extended identfier which handles external id in addtion of the account id.
type AccountIdentifiers struct {
	AccountID  *string `json:"account_id"`
	ExternalID *string `json:"external_id"`
} // @name AccountIdentifiers

// GetGroupsEvents response
type GetGroupsEvents struct {
	EventID string `json:"event_id"`
	GroupID string `json:"group_id"`
} // @name GetGroupsEvents

// HasToMembersList Checks if the ToMembersList is not empty
func (e Event) HasToMembersList() bool {
	return len(e.ToMembersList) > 0
}

// HasToMemberUser Checks if user with identifier exists whithin the ToMembers ACL
func (e Event) HasToMemberUser(userID *string, externalID *string) bool {
	for _, toMember := range e.ToMembersList {
		if userID != nil && toMember.UserID == *userID {
			return true
		}
		if externalID != nil && toMember.ExternalID == *externalID {
			return true
		}
	}
	return false
}

// GetMembersAsUserIDs Gets all members as list of user ids
func (e Event) GetMembersAsUserIDs(skipUserID *string) []string {
	var userIDs []string
	if len(e.ToMembersList) > 0 {
		for _, member := range e.ToMembersList {
			if skipUserID == nil || *skipUserID != member.UserID {
				userIDs = append(userIDs, member.UserID)
			}
		}
	}
	return userIDs
}

// GetMembersAsNotificationRecipients constructs all to members as notification recipients
func (e Event) GetMembersAsNotificationRecipients(skipUserID *string) []notifications.Recipient {
	recipients := []notifications.Recipient{}
	if len(e.ToMembersList) > 0 {
		for _, member := range e.ToMembersList {
			if skipUserID == nil || *skipUserID != member.UserID {
				recipients = append(recipients, notifications.Recipient{
					UserID: member.UserID,
					Name:   member.Name,
				})
			}
		}
	}
	return recipients
}

// GroupEventFilter event filter wrapper
type GroupEventFilter struct {
	Limit  *int64 `json:"limit,omitempty"`
	Offset *int64 `json:"offset,omitempty"`

	StartTimeAfter             *int64 `json:"start_time_after,omitempty"`
	StartTimeAfterNullEndTime  *int64 `json:"start_time_after_null_end_time,omitempty"`
	StartTimeBefore            *int64 `json:"start_time_before,omitempty"`
	StartTimeBeforeNullEndTime *int64 `json:"start_time_before_null_end_time,omitempty"`
	EndTimeAfter               *int64 `json:"end_time_after,omitempty"`
	EndTimeBefore              *int64 `json:"end_time_before,omitempty"`
} // @name GroupEventFilter
