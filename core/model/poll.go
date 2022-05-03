package model

import "time"

// Poll represents event entity
type Poll struct {
	ClientID      string     `json:"client_id" bson:"client_id"`
	PollID        string     `json:"poll_id" bson:"poll_id"`
	GroupID       string     `json:"group_id" bson:"group_id"`
	DateCreated   time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated   time.Time  `json:"date_updated" bson:"date_updated"`
	Creator       Creator    `json:"creator" bson:"creator"`
	ToMembersList []ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
} // @name Poll
