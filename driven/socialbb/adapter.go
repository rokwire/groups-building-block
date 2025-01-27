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
	"groups/core/model"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the SocialBB interface
type Adapter struct {
	socialURL             string
	serviceAccountManager *authservice.ServiceAccountManager
	logger                logs.Logger
}

// NewSocialAdapter creates a new adapter for Core API
func NewSocialAdapter(socialURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{socialURL: socialURL, serviceAccountManager: serviceAccountManager}
}

// InvokePostsOperation invokes the posts operation
func (a *Adapter) InvokePostsOperation(operation string, post model.Post) (*model.Post, error) {
	return nil, nil
}
