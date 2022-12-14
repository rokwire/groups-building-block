package model

import (
	"groups/driven/notifications"
	"time"
)

// MembershipCollection collection wrapper
type MembershipCollection struct {
	Items []GroupMembership
}

// ApplyGroupSettings Applies group settings
func (c *MembershipCollection) ApplyGroupSettings(settings *GroupSettings) {
	for index := range c.Items {
		c.Items[index].ApplyGroupSettings(settings)
	}
}

// GetMembershipBy Finds a membership by
func (c *MembershipCollection) GetMembershipBy(checker func(membership GroupMembership) bool) *GroupMembership {
	if len(c.Items) > 0 {
		for _, membership := range c.Items {
			if checker(membership) {
				return &membership
			}
		}
	}

	return nil
}

// membershipCriteriaPredicate should return shouldSend and isMuted
type membershipCriteriaPredicate = func(membership GroupMembership) (bool, bool)

// GetMembersAsRecipients gets members as list of Recipient recipients. nil status means all users.
func (c *MembershipCollection) GetMembersAsRecipients(predicate membershipCriteriaPredicate) []notifications.Recipient {
	var recipients []notifications.Recipient
	for _, membership := range c.Items {
		if shouldSend, isMuted := predicate(membership); shouldSend {
			recipient := membership.ToNotificationRecipient(isMuted)
			recipients = append(recipients, recipient)
		}
	}

	return recipients
}

// GetMembersByStatus gets members by status field
func (c *MembershipCollection) GetMembersByStatus(status string) []GroupMembership {
	var members []GroupMembership
	if c.Items == nil {
		return nil
	}
	for _, item := range c.Items {
		if item.Status == status {
			members = append(members, item)
		}
	}
	return members
}

// Evaluate membership mute policy. First result is canSend, second result is for muted
type mutePreferencePredicate = func(member GroupMembership) (bool, bool)

// GetMembersAsNotificationRecipients constructs all official members as notification recipients
func (c *MembershipCollection) GetMembersAsNotificationRecipients(predicate mutePreferencePredicate) []notifications.Recipient {

	recipients := []notifications.Recipient{}

	if len(c.Items) > 0 {
		for _, member := range c.Items {
			canSend, mute := predicate(member)
			if canSend {
				recipients = append(recipients, notifications.Recipient{
					UserID: member.UserID,
					Name:   member.Name,
					Mute:   mute,
				})
			}
		}
	}
	return recipients
}

// GroupMembership represents the membership of a user to a given group
type GroupMembership struct {
	ID         string `json:"id" bson:"_id"`
	ClientID   string `json:"client_id" bson:"client_id"`
	GroupID    string `json:"group_id" bson:"group_id"`
	UserID     string `json:"user_id" bson:"user_id"`
	ExternalID string `json:"external_id" bson:"external_id"`
	Name       string `json:"name" bson:"name"`
	NetID      string `json:"net_id" bson:"net_id"`
	Email      string `json:"email" bson:"email"`
	PhotoURL   string `json:"photo_url" bson:"photo_url"`

	// TODO: This is dangerous code-breaking change. There are existing clients that may use it in the old way.
	Status string `json:"status" bson:"status"` //pending, member, rejected
	Admin  bool   `json:"admin" bson:"admin"`

	RejectReason  string         `json:"reject_reason" bson:"reject_reason"`
	MemberAnswers []MemberAnswer `json:"member_answers" bson:"member_answers"`
	SyncID        string         `json:"sync_id" bson:"sync_id"` //ID of sync that last updated this membership

	NotificationsPreferences NotificationsPreferences `json:"notifications_preferences" bson:"notifications_preferences"`

	DateCreated  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated  *time.Time `json:"date_updated" bson:"date_updated"`
	DateAttended *time.Time `json:"date_attended" bson:"date_attended"`
} //@name GroupMembership

// GetDisplayName Constructs a display name based on the current data state
func (m *GroupMembership) GetDisplayName() string {
	if len(m.Name) > 0 {
		return m.Name
	} else if len(m.Email) > 0 {
		return m.Email
	} else if len(m.ExternalID) > 0 {
		return m.ExternalID
	}
	return ""
}

// ApplyFromUserIfEmpty Copy info from the user entity
func (m *GroupMembership) ApplyFromUserIfEmpty(user *User) {
	if m.UserID == "" && user.ID != "" {
		m.UserID = user.ID
	}
	if m.ExternalID == "" && user.ExternalID != "" {
		m.ExternalID = user.ExternalID
	}
	if m.Email == "" && user.Email != "" {
		m.Email = user.Email
	}
	if m.Name == "" && user.Name != "" {
		m.Name = user.Name
	}
}

// IsAdmin says if the user is admin of the group
func (m *GroupMembership) IsAdmin() bool {
	return m.Status == "admin"
}

// IsAdminOrMember says if the user is admin or member of the group
func (m *GroupMembership) IsAdminOrMember() bool {
	return m.IsMember() || m.IsAdmin()
}

// IsMember says if the member is a group member
func (m *GroupMembership) IsMember() bool {
	return m.Status == "member"
}

// IsPendingMember says if the member is a group pending
func (m *GroupMembership) IsPendingMember() bool {
	return m.Status == "pending"
}

// IsRejected says if the member is a group rejected
func (m *GroupMembership) IsRejected() bool {
	return m.Status == "rejected"
}

// ApplyGroupSettings Applies group settings and removes forbidden fields
func (m *GroupMembership) ApplyGroupSettings(settings *GroupSettings) {
	if settings == nil {
		val := DefaultGroupSettings()
		settings = &val
	}
	if !settings.MemberInfoPreferences.CanViewMemberName {
		m.Name = ""
	}
	if !settings.MemberInfoPreferences.CanViewMemberNetID {
		m.NetID = ""
	}
	if !settings.MemberInfoPreferences.CanViewMemberEmail {
		m.Email = ""
	}
	if !settings.MemberInfoPreferences.CanViewMemberPhone {
		// Just a placeholder.
	}
	m.ExternalID = ""

}

// ToShortMemberRecord converts to ShortMemberRecord
func (m *GroupMembership) ToShortMemberRecord() ShortMemberRecord {
	return ShortMemberRecord{
		ID:         m.ID,
		UserID:     m.UserID,
		ExternalID: m.ExternalID,
		Email:      m.Email,
		NetID:      m.NetID,
		Name:       m.Name,
		Status:     m.Status,
	}
}

// ToNotificationRecipient construct notifications.Recipient based on the data
func (m *GroupMembership) ToNotificationRecipient(mute bool) notifications.Recipient {
	return notifications.Recipient{
		UserID: m.UserID,
		Name:   m.Name,
		Mute:   mute,
	}
}

// ToMember converts the GroupMembership model to the Member model
func (m *GroupMembership) ToMember() Member {
	status := m.Status
	if m.Admin {
		status = "admin"
	}
	return Member{
		ID:            m.ID,
		UserID:        m.UserID,
		ExternalID:    m.ExternalID,
		Name:          m.Name,
		NetID:         m.NetID,
		Email:         m.Email,
		PhotoURL:      m.PhotoURL,
		Status:        status,
		RejectReason:  m.RejectReason,
		MemberAnswers: m.MemberAnswers,
		DateCreated:   m.DateCreated,
		DateUpdated:   m.DateUpdated,
		DateAttended:  m.DateAttended,
	}
}

// NotificationsPreferences overrides default notification preferences on group level
type NotificationsPreferences struct {
	OverridePreferences bool `json:"override_preferences" bson:"override_preferences"`
	AllMute             bool `json:"all_mute" bson:"all_mute"`
	InvitationsMuted    bool `json:"invitations_mute" bson:"invitations_mute"`
	PostsMuted          bool `json:"posts_mute" bson:"posts_mute"`
	EventsMuted         bool `json:"events_mute" bson:"events_mute"`
	PollsMuted          bool `json:"polls_mute" bson:"polls_mute"`
} // @name NotificationsPreferences
