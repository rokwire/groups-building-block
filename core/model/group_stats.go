// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

// GroupStats wraps group statistics aggregation result
type GroupStats struct {
	TotalCount      int `json:"total_count" bson:"total_count"` // pending and rejected are excluded
	AdminsCount     int `json:"admins_count" bson:"admins_count"`
	MemberCount     int `json:"member_count" bson:"member_count"`
	PendingCount    int `json:"pending_count" bson:"pending_count"`
	RejectedCount   int `json:"rejected_count" bson:"rejected_count"`
	AttendanceCount int `json:"attendance_count" bson:"attendance_count"`
} //@name GroupStats

// StatsFilter is a filter for event statistics.
type StatsFilter struct {
	BaseFilter GroupsFilter            `json:"base_filter" bson:"base_filter"`
	SubFilters map[string]GroupsFilter `json:"sub_filters" bson:"sub_filters"`
} //@name StatsFilter

// StatsResult represents the result of an event statistics query.
type StatsResult struct {
	Stats map[string]int `json:"stats" bson:"stats"`
} //@name StatsResult
