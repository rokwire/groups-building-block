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

//IsGroupAdminOrMember says if the user is an admin or a member of the group
func (gr Group) IsGroupAdminOrMember(userID string) bool {
	if gr.Members == nil {
		return false
	}
	for _, item := range gr.Members {
		if item.User.ID == userID && (item.Status == "admin" || item.Status == "member") {
			return true
		}
	}
	return false
}

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

//IsGroupRejected says if the user is a group rejected
func (gr Group) UserNameByID(userID string) *string {
	if gr.Members == nil {
		return nil
	}
	for _, item := range gr.Members {
		if item.User.ID == userID {
			name := item.Name
			return &name
		}
	}
	return nil
}
