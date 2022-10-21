package model

// GroupNotification wrapper for sending a notifications to members of a group
type GroupNotification struct {
	GroupID        string            `json:"group_id"`
	Sender         *Sender           `json:"sender"`
	Members        UserRefs          `json:"members"`
	MemberStatuses []string          `json:"member_statuses"` // default: ["admin", "member"]
	Subject        string            `json:"subject"`
	Topic          *string           `json:"topic"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data"`
} // @name GroupNotification

// UserRefs wrapper for the list of UserRef
type UserRefs []UserRef // @name UserRefs

// ToUserIDs Constructs list of user ids
func (m UserRefs) ToUserIDs() []string {
	var userIDs []string
	for _, userRef := range m {
		userIDs = append(userIDs, userRef.UserID)
	}
	return userIDs
}

// UserRef reference for a concrete user which is member of a group
type UserRef struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
} // @name UserRef

// Sender Wraps sender type and user ref
type Sender struct {
	Type string   `json:"type" bson:"type"` // user or system
	User *UserRef `json:"user,omitempty" bson:"user,omitempty"`
} // @name Sender
