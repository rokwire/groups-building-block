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

package web

import (
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Auth2 handler
type Auth2 struct {
	bbs tokenauth.Handlers

	logger *logs.Logger
}

// Start starts the auth module
func (auth *Auth2) Start() error {
	auth.logger.Info("Auth -> start")

	return nil
}

// NewAuth2 creates new auth handler
func NewAuth2(serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) (*Auth2, error) {

	bbsStandardHandler, err := newBBsStandardHandler(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "bbs auth", nil, err)
	}
	bbsHandlers := tokenauth.NewHandlers(bbsStandardHandler) //add permissions, user and authenticated

	auth := Auth2{
		bbs:    bbsHandlers,
		logger: logger,
	}
	return &auth, nil
}
