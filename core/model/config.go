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

// Config wrapper for in memory storage of configuration
type Config struct {
	AuthmanAdminUINList       []string
	ReportAbuseRecipientEmail string
	SyncManagedGroupsPeriod   int //Period at which to automatically sync managed groups in minutes
	SupportedClientIDs        []string
}

// ManagedGroupConfig defines a configs for a set of managed groups
type ManagedGroupConfig struct {
	ID           string     `json:"id" bson:"_id"`
	ClientID     string     `json:"client_id" bson:"client_id"`
	AuthmanStems []string   `json:"authman_stems" bson:"authman_stems"`
	AdminUINs    []string   `json:"admin_uins" bson:"admin_uins"`
	Type         string     `json:"type" bson:"type"`
	DateCreated  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated  *time.Time `json:"date_updated" bson:"date_updated"`
} //@name ManagedGroupConfig
