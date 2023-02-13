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
	// TypeEnvConfigData env configs type
	TypeEnvConfigData logutils.MessageDataType = "env config data"
	// TypeApplicationConfigData application configs type
	TypeApplicationConfigData logutils.MessageDataType = "application config data"
	// TypeSyncConfigData sync configs type
	TypeSyncConfigData logutils.MessageDataType = "sync config data"
	// TypeManagedGroupConfigData managed group configs type
	TypeManagedGroupConfigData logutils.MessageDataType = "managed group config data"

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
	DateUpdated *time.Time  `json:"date_updated" bson:"date_updated"`
}

// DataAsEnvConfig returns the config Data as an EnvConfigData if the cast succeeds
func (c Config) DataAsEnvConfig() (*EnvConfigData, error) {
	data, ok := c.Data.(EnvConfigData)
	if !ok {
		return nil, errors.ErrorData(logutils.StatusInvalid, TypeEnvConfigData, nil)
	}
	return &data, nil
}

// DataAsApplicationConfig returns the config Data as an ApplicationConfigData if the cast succeeds
func (c Config) DataAsApplicationConfig() (*ApplicationConfigData, error) {
	data, ok := c.Data.(ApplicationConfigData)
	if !ok {
		return nil, errors.ErrorData(logutils.StatusInvalid, TypeEnvConfigData, nil)
	}
	return &data, nil
}

// DataAsSyncConfig returns the config Data as a SyncConfigData if the cast succeeds
func (c Config) DataAsSyncConfig() (*SyncConfigData, error) {
	data, ok := c.Data.(SyncConfigData)
	if !ok {
		return nil, errors.ErrorData(logutils.StatusInvalid, TypeSyncConfigData, nil)
	}
	return &data, nil
}

// DataAsManagedGroupConfig returns the config Data as a ManagedGroupConfigData if the cast succeeds
func (c Config) DataAsManagedGroupConfig() (*ManagedGroupConfigData, error) {
	data, ok := c.Data.(ManagedGroupConfigData)
	if !ok {
		return nil, errors.ErrorData(logutils.StatusInvalid, TypeManagedGroupConfigData, nil)
	}
	return &data, nil
}

// EnvConfigData contains environment configs for this service
type EnvConfigData struct {
	ExampleEnv string `json:"example_env" bson:"example_env"`
}

// ApplicationConfigData defines configs for managing authman sync and sending report emails
type ApplicationConfigData struct {
	//TODO: must not be empty
	AuthmanAdminUINList       []string `json:"authman_admin_uin_list" bson:"authman_admin_uin_list"`
	ReportAbuseRecipientEmail string   `json:"report_abuse_recipient_email" bson:"report_abuse_recipient_email"`
}

// SyncConfigData defines system configs for managed group sync
type SyncConfigData struct {
	CRON          string `json:"cron" bson:"cron"`
	TimeThreshold int    `json:"time_threshold" bson:"time_threshold"` // Threshold from start_time to be considered same run in minutes
	Timeout       int    `json:"timeout" bson:"timeout"`               // Time from start_time to be considered a failed run in minutes
	GroupTimeout  int    `json:"group_timeout" bson:"group_timeout"`   // Time from sync_start_time to be considered a failed run for a single group in minutes
}

// SyncTimes defines the times used to prevent concurrent syncs
type SyncTimes struct {
	AppID     string     `json:"app_id" bson:"app_id"`
	OrgID     string     `json:"org_id" bson:"org_id"`
	StartTime *time.Time `json:"start_time" bson:"start_time"`
	EndTime   *time.Time `json:"end_time" bson:"end_time"`
}

// ManagedGroupConfigData defines a config for a set of managed groups
type ManagedGroupConfigData struct {
	AuthmanStems []string   `json:"authman_stems" bson:"authman_stems"`
	AdminUINs    []string   `json:"admin_uins" bson:"admin_uins"`
	Type         string     `json:"type" bson:"type"`
	DateCreated  time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated  *time.Time `json:"date_updated" bson:"date_updated"`
} //@name ManagedGroupConfigData
