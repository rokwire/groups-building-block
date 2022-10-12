package model

// MembershipFilter Wraps all possible filters for getting group members call
type MembershipFilter struct {
	ID         *string  `json:"id"`          // membership id
	GroupIDs   []string `json:"group_ids"`   // list of group ids
	UserID     *string  `json:"user_id"`     // core user id
	UserIDs    []string `json:"user_ids"`    // core user ids
	ExternalID *string  `json:"external_id"` // core user external id
	NetID      *string  `json:"net_id"`      // core user net id
	Name       *string  `json:"name"`        // member's name
	Statuses   []string `json:"statuses"`    // lest of membership statuses
	Offset     *int64   `json:"offset"`      // result offset
	Limit      *int64   `json:"limit"`       // result limit
} // @name MembershipFilter

// GroupsFilter Wraps all possible filters for getting a group
type GroupsFilter struct {
	GroupIDs         []string `json:"ids"`                // membership id
	MemberID         *string  `json:"member_id"`          // member id
	MemberUserID     *string  `json:"member_user_id"`     // member user id
	MemberExternalID *string  `json:"member_external_id"` // member user external id
	Category         *string  `json:"category"`           // group category
	Privacy          *string  `json:"privacy"`            // group privacy
	Title            *string  `json:"title"`              // group title
	Order            *string  `json:"order"`              // order by category & name (asc desc)
	Offset           *int64   `json:"offset"`             // result offset
	Limit            *int64   `json:"limit"`              // result limit
} // @name GroupsFilter
