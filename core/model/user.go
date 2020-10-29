package model

import "time"

//User represents user entity
type User struct {
	ID          string     `json:"id" bson:"_id"`
	ExternalID  string     `json:"external_id" bson:"external_id"`
	Email       string     `json:"email" bson:"email"`
	IsMemberOf  *[]string  `json:"is_member_of" bson:"is_member_of"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

//IsMemberOfGroup says if the user is member of a group
func (user User) IsMemberOfGroup(group string) bool {
	if user.IsMemberOf == nil || len(*user.IsMemberOf) == 0 {
		return false
	}
	for _, current := range *user.IsMemberOf {
		if current == group {
			return true
		}
	}
	return false
}
