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
} // @name Group

//IsGroupAdmin says if the user is admin of the group
func (gr Group) IsGroupAdmin(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.User.ID == userID && item.Status == "admin" {
			return true
		}
	}
	return false
}

//IsGroupMember says if the user is a group member
func (gr Group) IsGroupMember(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.User.ID == userID && item.Status == "member" {
			return true
		}
	}
	return false
}

//IsGroupPending says if the user is a group pending
func (gr Group) IsGroupPending(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.User.ID == userID && item.Status == "pending" {
			return true
		}
	}
	return false
}

//IsGroupRejected says if the user is a group rejected
func (gr Group) IsGroupRejected(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.User.ID == userID && item.Status == "rejected" {
			return true
		}
	}
	return false
}
