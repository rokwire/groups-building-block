package model

// MembershipFilter Wraps all possible filters for getting group members call
type MembershipFilter struct {
	ID         *string  `json:"id"`          // membership id
	GroupIDs   []string `json:"group_ids"`   // list of group ids
	UserID     *string  `json:"user_id"`     // core user id
	UserIDs    []string `json:"user_ids"`    // core user ids
	ExternalID *string  `json:"external_id"` // core user external id
	NetID      *string  `json:"net_id"`      // core user net id
	NetIDs     []string `json:"net_ids"`     // core user net ids
	Name       *string  `json:"name"`        // member's name
	Statuses   []string `json:"statuses"`    // lest of membership statuses
	Offset     *int64   `json:"offset"`      // result offset
	Limit      *int64   `json:"limit"`       // result limit
} // @name MembershipFilter

// GroupsFilter Wraps all possible filters for getting a group
type GroupsFilter struct {
	GroupIDs            []string                       `json:"ids"`                // membership id
	MemberID            *string                        `json:"member_id"`          // member id
	MemberUserID        *string                        `json:"member_user_id"`     // member user id
	MemberExternalID    *string                        `json:"member_external_id"` // member user external id
	MemberStatus        []string                       `json:"member_status"`      // member user status
	Title               *string                        `json:"title"`              // group title
	Category            *string                        `json:"category"`           // group category
	Privacy             *string                        `json:"privacy"`            // group privacy
	Tags                []string                       `json:"tags"`               // group tags
	IncludeHidden       *bool                          `json:"include_hidden"`     // Include hidden groups
	Hidden              *bool                          `json:"hidden"`             // Filter by hidden flag. Values: true (show only hidden), false (show only not hidden), missing - don't do any filtering on this field.
	ExcludeMyGroups     *bool                          `json:"exclude_my_groups"`  // Exclude My groups
	AuthmanEnabled      *bool                          `json:"authman_enabled"`
	ResearchOpen        *bool                          `json:"research_open"`
	ResearchGroup       *bool                          `json:"research_group"`
	ResearchAnswers     map[string]map[string][]string `json:"research_answers"`
	Attributes          map[string]interface{}         `json:"attributes"`
	Order               *string                        `json:"order"`                  // order by category & name (asc desc)
	Offset              *int64                         `json:"offset"`                 // result offset
	Limit               *int64                         `json:"limit"`                  // result limit
	LimitID             *string                        `json:"limit_id"`               // limit id
	LimitIDExtraRecords *int64                         `json:"limit_id_extra_records"` // limit id number of extra records, default 0
	DaysInactive        *int64                         `json:"days_inactive"`
	Administrative      *bool                          `json:"administrative"`
} // @name GroupsFilter

// PostsFilter Wraps all possible filters for getting group post call
type PostsFilter struct {
	GroupID       string  `json:"group_id"`
	PostType      *string `json:"type"`
	ScheduledOnly *bool   `json:"scheduled_only"`
	Offset        *int64  `json:"offset"`
	Limit         *int64  `json:"limit"`
	Order         *string `json:"order"`
} // @name PostsFilter
