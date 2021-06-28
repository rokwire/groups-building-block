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

	ClientID string `bson:"client_id"`
} // @name User

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

//ShibbolethAuth represents shibboleth auth entity
type ShibbolethAuth struct {
	Uin        string    `json:"uiucedu_uin" bson:"uiucedu_uin"`
	Email      string    `json:"email" bson:"email"`
	IsMemberOf *[]string `json:"uiucedu_is_member_of" bson:"uiucedu_is_member_of"`
}
