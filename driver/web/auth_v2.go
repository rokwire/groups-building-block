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
	"net/http"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/tokenauth"
)

// Auth2 handler
type Auth2 struct {
	services tokenauth.Handlers
	admin    tokenauth.Handlers
	bbs      tokenauth.Handlers

	logger *logs.Logger
}

// Start starts the auth module
func (auth *Auth2) Start() error {
	auth.logger.Info("Auth -> start")

	return nil
}

// NewAuth creates new auth handler
func NewAuth2(serviceRegManager *auth.ServiceRegManager, logger *logs.Logger) (*Auth2, error) {
	servicesStandardHandler, err := newServicesStandardHandler(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client auth", nil, err)
	}
	servicesHandlers := tokenauth.NewHandlers(servicesStandardHandler) //add permissions, user and authenticated

	adminStandardHandler, err := newAdminStandardHandler(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin auth", nil, err)
	}
	adminHandlers := tokenauth.NewHandlers(adminStandardHandler) //add permissions, user and authenticated

	bbsStandardHandler, err := newBBsStandardHandler(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "bbs auth", nil, err)
	}
	bbsHandlers := tokenauth.NewHandlers(bbsStandardHandler) //add permissions, user and authenticated

	auth := Auth2{
		services: servicesHandlers,
		admin:    adminHandlers,
		bbs:      bbsHandlers,
		logger:   logger,
	}
	return &auth, nil
}

///////

// START SERVICES auth /////////

// start servicesStandardHandler

type servicesStandardHandler struct {
	tokenAuth *tokenauth.TokenAuth
}

// Check validates the request contains a valid client token
func (auth servicesStandardHandler) Check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := auth.tokenAuth.CheckRequestToken(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction(logutils.ActionValidate, logutils.TypeToken, nil, err)
	}

	if claims.Admin {
		return http.StatusUnauthorized, nil, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
	}
	if claims.System {
		return http.StatusUnauthorized, nil, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
	}

	// TODO: Enable scope authorization
	// err = auth.tokenAuth.AuthorizeRequestScope(claims, req)
	// if err != nil {
	// 	return http.StatusForbidden, nil, errors.WrapErrorAction(logutils.ActionValidate, logutils.TypeScope, nil, err)
	// }

	return http.StatusOK, claims, nil
}

// GetTokenAuth returns the TokenAuth from the handler
func (auth servicesStandardHandler) GetTokenAuth() *tokenauth.TokenAuth {
	return auth.tokenAuth
}

func newServicesStandardHandler(serviceRegManager *auth.ServiceRegManager) (*servicesStandardHandler, error) {

	clientPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_services_permission_policy.csv")
	clientTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, clientPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client token auth", nil, err)
	}

	auth := servicesStandardHandler{tokenAuth: clientTokenAuth}
	return &auth, nil
}

// end servicesStandardHandler

// END SERVICES auth ///////////

// admin auth ///////////

func newAdminStandardHandler(serviceRegManager *auth.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	adminPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_admin_permission_policy.csv")
	adminTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, adminPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Admin {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(adminTokenAuth, check)
	return auth, nil
}

// END admin auth //////////

// BBs auth ///////////

func newBBsStandardHandler(serviceRegManager *auth.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	bbsPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_bbs_permission_policy.csv")
	bbsTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, bbsPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "bbs token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Service {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "service claim", nil)
		}

		if !claims.FirstParty {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "first party claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(bbsTokenAuth, check)
	return auth, nil
}

// END BBs auth //////////

// tps auth ///////////

//start tps handlers
