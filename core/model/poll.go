package model

import "time"

// Poll represents poll entity
type Poll struct {
	PollID      string    `json:"poll_id_id" bson:"poll_id"`
	GroupID     string    `json:"group_id" bson:"group_id"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
} // @name Poll
