package model

import "time"

// Member represents group member entity
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

// MemberAnswer represents member answer entity
type MemberAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
} //@name MemberAnswer

// IsAdmin says if the user is admin of the group
func (m *Member) IsAdmin() bool {
	return m.Status == "admin"
}

// IsAdminOrMember says if the user is admin or member of the group
func (m *Member) IsAdminOrMember() bool {
	return m.IsMember() || m.IsAdmin()
}

// IsMember says if the member is a group member
func (m *Member) IsMember() bool {
	return m.Status == "member"
}

// IsPendingMember says if the member is a group pending
func (m *Member) IsPendingMember() bool {
	return m.Status == "pending"
}

//IsRejected says if the member is a group rejected
func (m *Member) IsRejected() bool {
	return m.Status == "rejected"
}
