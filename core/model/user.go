package model

import "time"

//User represents user entity
type User struct {
	ID          string     `json:"id" bson:"_id"`
	IsCoreUser  bool       `json:"is_core_user" bson:"is_core_user"`
	ExternalID  string     `json:"external_id" bson:"external_id"`
	Email       string     `json:"email" bson:"email"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
	ClientID    string     `bson:"client_id"`
} // @name User
