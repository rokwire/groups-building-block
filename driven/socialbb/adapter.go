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

package socialbb

import (
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Adapter implements the SocialBB interface
type Adapter struct {
	socialURL             string
	serviceAccountManager *auth.ServiceAccountManager
	logger                *logs.Logger
}

// NewSocialAdapter creates a new adapter for Core API
func NewSocialAdapter(logger *logs.Logger, socialURL string, serviceAccountManager *auth.ServiceAccountManager) *Adapter {
	return &Adapter{logger: logger, socialURL: socialURL, serviceAccountManager: serviceAccountManager}
}
