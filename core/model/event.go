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

//Event represents event entity
type Event struct {
	ClientID      string     `json:"client_id" bson:"client_id"`
	EventID       string     `json:"event_id" bson:"event_id"`
	GroupID       string     `json:"group_id" bson:"group_id"`
	DateCreated   time.Time  `json:"date_created" bson:"date_created"`
	Creator       Creator    `json:"creator" bson:"creator"`
	ToMembersList []ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
} // @name Event

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
