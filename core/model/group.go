package model

import "time"

//Group represents group entity
type Group struct {
	ID                  string   `json:"id"`
	Category            string   `json:"category"` //one of the enums categories list
	Title               string   `json:"title"`
	Privacy             string   `json:"privacy"` //public or private
	Description         *string  `json:"description"`
	ImageURL            *string  `json:"image_url"`
	WebURL              *string  `json:"web_url"`
	MembersCount        int      `json:"members_count"` //to be supported up to date
	Tags                []string `json:"tags"`
	MembershipQuestions []string `json:"membership_questions"`

	Members []Member `json:"members"`

	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
}
