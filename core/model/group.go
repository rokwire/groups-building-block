package model

import "time"

//Group represents group entity
type Group struct {
	ID                  string
	Category            string //one of the enums categories list
	Title               string
	Privacy             string //public or private
	Description         *string
	ImageURL            *string
	WebURL              *string
	MembersCount        int //to be supported up to date
	Tags                []string
	MembershipQuestions []string

	Members []Member

	DateCreated time.Time
	DateUpdated *time.Time
}
