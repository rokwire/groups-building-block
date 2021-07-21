package model

import "time"

//Post represents group posts
type Post struct {
	ID          *string     `json:"id" bson:"_id"`
	ClientID    *string     `json:"client_id" bson:"client_id"`
	GroupID     string      `json:"group_id" bson:"group_id"`
	ParentID    *string     `json:"parent_id" bson:"parent_id"`
	Member      PostCreator `json:"member" bson:"member"`
	Subject     string      `json:"subject" bson:"subject"`
	Body        string      `json:"body" bson:"body"`
	Private     bool        `json:"private" bson:"private"`
	Replies     []*Post     `json:"replies,omitempty"` // This is constructed by the code (ParentID)
	DateCreated *time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time  `json:"date_updated" bson:"date_updated"`
}

//PostCreator represents group member entity
type PostCreator struct {
	UserID   string `json:"user_id" bson:"user_id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	PhotoURL string `json:"photo_url" bson:"photo_url"`
} //@name PostCreator
