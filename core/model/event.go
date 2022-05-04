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
