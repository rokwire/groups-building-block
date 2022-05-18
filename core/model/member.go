package model

import "time"

// Member represents group member entity
type Member struct {
	ID            string         `json:"id" bson:"id"`
	UserID        string         `json:"user_id" bson:"user_id"`
	ExternalID    string         `json:"external_id" bson:"external_id"`
	Name          string         `json:"name" bson:"name"`
	Email         string         `json:"email" bson:"email"`
	PhotoURL      string         `json:"photo_url" bson:"photo_url"`
	Status        string         `json:"status" bson:"status"` //pending, member, admin, rejected
	RejectReason  string         `json:"reject_reason" bson:"reject_reason"`
	MemberAnswers []MemberAnswer `json:"member_answers" bson:"member_answers"`

	DateCreated    time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated    *time.Time `json:"date_updated" bson:"date_updated"`
	DateAttendance *time.Time `json:"date_attendance" bson:"date_attendance"`
} //@name Member

// ToMember represents to(destination) member entity
type ToMember struct {
	UserID     string `json:"user_id" bson:"user_id"`
	ExternalID string `json:"external_id" bson:"external_id"`
	Name       string `json:"name" bson:"name"`
	Email      string `json:"email" bson:"email"`
} //@name ToMember

// MemberAnswer represents member answer entity
type MemberAnswer struct {
	Question string `json:"question" bson:"question"`
	Answer   string `json:"answer" bson:"answer"`
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
