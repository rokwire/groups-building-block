package model

import "time"

//Post represents group posts
type Post struct {
	ID          *string   `json:"id" bson:"_id"`
	ClientID    *string   `json:"client_id" json:"client_id"`
	GroupID     string    `json:"group_id" bson:"group_id"`
	Member      Member    `json:"member" bson:"member"`
	Subject     string    `json:"subject" bson:"subject"`
	Body        string    `json:"body" bson:"body"`
	Status      string    `json:"status" bson:"status"`
	Private     bool      `json:"private" bson:"private"`
	Replies     *[]Reply  `json:"replies,omitempty" bson:"replies,omitempty"` // This is constructed by the code (ParentID)
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
}

//Reply represents replies on a post
type Reply struct {
	ID          string    `json:"id" bson:"id"`
	ClientID    string    `json:"client_id" json:"client_id"`
	ParentID    string    `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	GroupID     string    `json:"group_id" bson:"group_id"`
	Member      Member    `json:"member" bson:"member"`
	Subject     string    `json:"subject" bson:"subject"`
	Body        string    `json:"body" bson:"body"`
	Status      string    `json:"status" bson:"status"`
	Private     bool      `json:"private" bson:"private"`
	Replies     []Reply   `json:"replies,omitempty" bson:"replies,omitempty"` // This is constructed by the code (ParentID)
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
}
