package model

// GroupSettings wraps group settings and flags as a separate unit
type GroupSettings struct {
	MemberInfoPreferences MemberInfoPreferences    `json:"member_info_preferences" bson:"member_info_preferences"`
	PostPreferences       PostPreferences          `json:"post_preferences" bson:"post_preferences"`
	ContentItems          []map[string]interface{} `json:"content_items" bson:"content_items"`
} // @name GroupSettings

// DefaultGroupSettings Returns default settings
func DefaultGroupSettings() GroupSettings {
	return GroupSettings{
		MemberInfoPreferences: MemberInfoPreferences{
			AllowMemberInfo:    true,
			CanViewMemberNetID: true,
			CanViewMemberName:  true,
			CanViewMemberEmail: true,
			CanViewMemberPhone: true,
		},
		PostPreferences: PostPreferences{
			AllowSendPost:                true,
			CanSendPostToSpecificMembers: true,
			CanSendPostToAdmins:          true,
			CanSendPostToAll:             true,
			CanSendPostReplies:           true,
			CanSendPostReactions:         true,
		},
		ContentItems: []map[string]interface{}{},
	}
}

// MemberInfoPreferences wrap settings for the visible member information
type MemberInfoPreferences struct {
	AllowMemberInfo    bool `json:"allow_member_info" bson:"allow_member_info"`
	CanViewMemberNetID bool `json:"can_view_member_net_id" bson:"can_view_member_net_id"`
	CanViewMemberName  bool `json:"can_view_member_name" bson:"can_view_member_name"`
	CanViewMemberEmail bool `json:"can_view_member_email" bson:"can_view_member_email"`
	CanViewMemberPhone bool `json:"can_view_member_phone" bson:"can_view_member_phone"`
} // @name MemberInfoPreferences

// PostPreferences wraps post preferences
type PostPreferences struct {
	AllowSendPost                bool `json:"allow_send_post" bson:"allow_send_post"`
	CanSendPostToSpecificMembers bool `json:"can_send_post_to_specific_members" bson:"can_send_post_to_specific_members"`
	CanSendPostToAdmins          bool `json:"can_send_post_to_admins" bson:"can_send_post_to_admins"`
	CanSendPostToAll             bool `json:"can_send_post_to_all" bson:"can_send_post_to_all"`
	CanSendPostReplies           bool `json:"can_send_post_replies" bson:"can_send_post_replies"`
	CanSendPostReactions         bool `json:"can_send_post_reactions" bson:"can_send_post_reactions"`
} // @name PostPreferences
