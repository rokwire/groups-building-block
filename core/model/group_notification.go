package model

// GroupNotification wrapper for sending a notifications to members of a group
type GroupNotification struct {
	GroupID        string            `json:"group_id"`
	MemberStatuses []string          `json:"member_statuses"` // default: ["admin", "member"]
	MemberIDs      MemberIDs         `json:"members"`
	Subject        string            `json:"subject"`
	Topic          *string           `json:"topic"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data"`
} // @name GroupNotification

// MemberIDs wrapper for the list of MemberRef
type MemberIDs []MemberRef // @name MemberIDs

// ToUserIDs Constructs list of user ids
func (m MemberIDs) ToUserIDs() []string {
	var userIDs []string
	for _, userRef := range m {
		userIDs = append(userIDs, userRef.UserID)
	}
	return userIDs
}

// MemberRef reference for a concrete user which is member of a group
type MemberRef struct {
	UserID string `json:"user_id"`
} // @name MemberRef
