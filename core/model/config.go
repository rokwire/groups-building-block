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

import "time"

// ApplicationConfig wrapper for in memory storage of configuration
type ApplicationConfig struct {
	AuthmanAdminUINList       []string
	ReportAbuseRecipientEmail string
	SupportedClientIDs        []string
	AppID                     string
	OrgID                     string
}

// SyncConfig defines system configs for managed group sync
type SyncConfig struct {
	Type          string `json:"type" bson:"type"`
	ClientID      string `json:"client_id" bson:"client_id"`
	CRON          string `json:"cron" bson:"cron"`
	TimeThreshold int    `json:"time_threshold" bson:"time_threshold"` // Threshold from start_time to be considered same run in minutes
	Timeout       int    `json:"timeout" bson:"timeout"`               // Time from start_time to be considered a failed run in minutes
	GroupTimeout  int    `json:"group_timeout" bson:"group_timeout"`   // Time from sync_start_time to be considered a failed run for a single group in minutes
}

// SyncTimes defines the times used to prevent concurrent syncs
type SyncTimes struct {
	Key       string     `json:"key" bson:"key"`
	StartTime *time.Time `json:"start_time" bson:"start_time"`
	EndTime   *time.Time `json:"end_time" bson:"end_time"`
}

// ManagedGroupConfig defines a config for a set of managed groups
type ManagedGroupConfig struct {
	ID           string     `json:"id" bson:"_id"`
	ClientID     string     `json:"client_id" bson:"client_id"`
	AuthmanStems []string   `json:"authman_stems" bson:"authman_stems"`
	AdminUINs    []string   `json:"admin_uins" bson:"admin_uins"`
	Type         string     `json:"type" bson:"type"`
	DateCreated  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated  *time.Time `json:"date_updated" bson:"date_updated"`
} //@name ManagedGroupConfig
