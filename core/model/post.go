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

// Post represents group posts
type Post struct {
	ID                *string             `json:"id" bson:"_id"`
	ClientID          *string             `json:"client_id" bson:"client_id"`
	GroupID           string              `json:"group_id" bson:"group_id"`
	ParentID          *string             `json:"parent_id" bson:"parent_id"`
	TopParentID       *string             `json:"top_parent_id" bson:"top_parent_id"`
	Creator           Creator             `json:"member" bson:"member"`
	Subject           string              `json:"subject" bson:"subject"`
	Body              string              `json:"body" bson:"body"`
	Private           bool                `json:"private" bson:"private"`
	UseAsNotification bool                `json:"use_as_notification" bson:"use_as_notification"`
	IsAbuse           bool                `json:"is_abuse" bson:"is_abuse"`
	Replies           []*Post             `json:"replies,omitempty"`                              // This is constructed by the code (ParentID)
	Reactions         map[string][]string `json:"reactions,omitempty" bson:"reactions,omitempty"` //TODO deprecated delete when deployed to prod
	ReactionStats     map[string]int      `json:"reaction_stats" bson:"reaction_stats"`
	ImageURL          *string             `json:"image_url" bson:"image_url"`

	ToMembersList []ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
	DateCreated   *time.Time `json:"date_created" bson:"date_created"`
	DateUpdated   *time.Time `json:"date_updated" bson:"date_updated"`
}

// UserCanSeePost checks if the user can see the current post or not
func (p *Post) UserCanSeePost(userID string) bool {
	if len(p.ToMembersList) > 0 {
		for _, member := range p.ToMembersList {
			if member.UserID == userID {
				return true
			}
		}

		return p.Creator.UserID == userID
	}

	return true
}

// GetMembersAsNotificationRecipients constructs all to members as notification recipients
func (p Post) GetMembersAsNotificationRecipients(skipUserID *string) []notifications.Recipient {
	recipients := []notifications.Recipient{}
	if len(p.ToMembersList) > 0 {
		for _, member := range p.ToMembersList {
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

// GetMembersAsUserIDs constructs a list of user ids
func (p Post) GetMembersAsUserIDs(skipUserID *string) []string {
	var recipients []string
	if len(p.ToMembersList) > 0 {
		for _, member := range p.ToMembersList {
			if skipUserID == nil || *skipUserID != member.UserID {
				recipients = append(recipients, member.UserID)
			}
		}
	}
	return recipients
}
