package model

// GroupStats wraps group statistics aggregation result
type GroupStats struct {
	TotalCount      int `json:"total_count" bson:"total_count"`
	AdminsCount     int `json:"admins_count" bson:"admins_count"`
	MemberCount     int `json:"member_count" bson:"member_count"`
	PendingCount    int `json:"pending_count" bson:"pending_count"`
	RejectedCount   int `json:"rejected_count" bson:"rejected_count"`
	AttendanceCount int `json:"attendance_count" bson:"attendance_count"`
} //@name GroupStats
