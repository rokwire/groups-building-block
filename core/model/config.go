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

import (
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	// TypeConfig configs type
	TypeConfig logutils.MessageDataType = "config"
	// TypeConfigData config data type
	TypeConfigData logutils.MessageDataType = "config data"

	// ConfigTypeEnv is the Config ID for EnvConfigData
	ConfigTypeEnv string = "env"
	// ConfigTypeApplication is the Config ID for ApplicationConfigData
	ConfigTypeApplication string = "application"
	// ConfigTypeSync is the Config ID for SyncConfigData
	ConfigTypeSync string = "sync"
	// ConfigTypeManagedGroup is the Config ID for ManagedGroupConfigData
	ConfigTypeManagedGroup string = "managed_group"
)

// Config contain generic configs
type Config struct {
	ID          string      `json:"id" bson:"_id"`
	Type        string      `json:"type" bson:"type"`
	AppID       string      `json:"app_id" bson:"app_id"`
	OrgID       string      `json:"org_id" bson:"org_id"`
	System      bool        `json:"system" bson:"system"`
	Data        interface{} `json:"data" bson:"data"`
	DateCreated time.Time   `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time  `json:"date_updated,omitempty" bson:"date_updated"`
} //@name Config

// EnvConfigData contains environment configs for this service
type EnvConfigData struct {
	//TODO: move more environment variables here as desired
	NotificationsReportAbuseEmail string `json:"notifications_report_abuse_email" bson:"notifications_report_abuse_email"`
} //@name EnvConfigData

// ApplicationConfigData defines configs for managing authman sync and sending report emails
type ApplicationConfigData struct {
	AuthmanAdminUINList       []string `json:"authman_admin_uin_list" bson:"authman_admin_uin_list"`
	ReportAbuseRecipientEmail string   `json:"report_abuse_recipient_email" bson:"report_abuse_recipient_email"`
} //@name ApplicationConfigData

// SyncConfigData defines system configs for managed group sync
type SyncConfigData struct {
	CRON          string `json:"cron" bson:"cron"`
	TimeThreshold int    `json:"time_threshold" bson:"time_threshold"` // Threshold from start_time to be considered same run in minutes
	Timeout       int    `json:"timeout" bson:"timeout"`               // Time from start_time to be considered a failed run in minutes
	GroupTimeout  int    `json:"group_timeout" bson:"group_timeout"`   // Time from sync_start_time to be considered a failed run for a single group in minutes
} //@name SyncConfigData

// SyncTimes defines the times used to prevent concurrent syncs
type SyncTimes struct {
	AppID     string     `json:"app_id" bson:"app_id"`
	OrgID     string     `json:"org_id" bson:"org_id"`
	StartTime *time.Time `json:"start_time" bson:"start_time"`
	EndTime   *time.Time `json:"end_time" bson:"end_time"`
}

// ManagedGroupConfigData defines a set of managed groups
type ManagedGroupConfigData struct {
	ManagedGroups []ManagedGroupConfig `json:"managed_groups" bson:"managed_groups"`
} //@name ManagedGroupConfigData

// ManagedGroupConfig defines a config for a single managed group
type ManagedGroupConfig struct {
	ID           string   `json:"id" bson:"_id"`
	Type         string   `json:"type" bson:"type"`
	AuthmanStems []string `json:"authman_stems" bson:"authman_stems"`
	AdminUINs    []string `json:"admin_uins" bson:"admin_uins"`
} //@name ManagedGroupConfig

// GetConfigData returns a pointer to the given config's Data as the given type T
func GetConfigData[T ConfigData](c Config) (*T, error) {
	if data, ok := c.Data.(T); ok {
		return &data, nil
	}
	return nil, errors.ErrorData(logutils.StatusInvalid, TypeConfigData, &logutils.FieldArgs{"type": c.Type})
}

// ConfigData represents any set of data that may be stored in a config
type ConfigData interface {
	EnvConfigData | ApplicationConfigData | SyncConfigData | ManagedGroupConfigData | map[string]interface{}
}
