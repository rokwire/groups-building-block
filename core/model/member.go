package model

import "time"

//Member represents group member entity
type Member struct {
	ID            string
	User          User
	Name          string
	Email         string
	PhotoURL      string
	Status        string //pending, member, admin
	Group         Group
	MemberAnswers []MemberAnswer

	DateCreated time.Time
	DateUpdated *time.Time
}

type MemberAnswer struct {
	Question string
	Answer   string
}
