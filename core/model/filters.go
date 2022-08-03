package model

// GroupMembersFilter Wraps all possible filters for getting group members call
type GroupMembersFilter struct {
	ID         *string  `json:"id"`          // membership id
	UserID     *string  `json:"user_id"`     // core user id
	ExternalID *string  `json:"external_id"` // core user external id
	NetID      *string  `json:"net_id"`      // core user net id
	Name       *string  `json:"name"`        // member's name
	Statuses   []string `json:"statuses"`    // lest of membership statuses
	Offset     *int64   `json:"offset"`      // result offset
	Limit      *int64   `json:"limit"`       // result limit
} // @name GroupMembersFilter
