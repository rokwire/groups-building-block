package model

import "time"

//Member represents group member entity
type Member struct {
	ID            string         `json:"id"`
	User          User           `json:"user"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	PhotoURL      string         `json:"photo_url"`
	Status        string         `json:"status"` //pending, member, admin, rejected
	RejectReason  string         `json:"reject_reason"`
	Group         Group          `json:"group"`
	MemberAnswers []MemberAnswer `json:"member_answers"`

	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
} //@name Member

//MemberAnswer represents member answer entity
type MemberAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
} //@name MemberAnswer
