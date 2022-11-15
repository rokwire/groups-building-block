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
	"fmt"
	"groups/core"
	"groups/core/model"
	"groups/driver/web/rest"
	"groups/utils"
	"log"
	"net/http"

	"github.com/casbin/casbin"
	"github.com/rokwire/core-auth-library-go/authservice"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Adapter entity
type Adapter struct {
	host string
	auth *Auth

	apisHandler         *rest.ApisHandler
	adminApisHandler    *rest.AdminApisHandler
	internalApisHandler *rest.InternalApisHandler
}

// @title Rokwire Groups Building Block API
// @description Rokwire Groups Building Block API Documentation.
// @version 1.6.1
// @host localhost
// @BasePath /gr
// @schemes https

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey AppUserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey IntAPIKeyAuth
// @in header
// @name INTERNAL-API-KEY

// Start starts the web server
func (we *Adapter) Start() {
	router := mux.NewRouter().StrictSlash(true)

	subrouter := router.PathPrefix("/gr").Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	subrouter.HandleFunc("/doc", we.serveDoc)
	subrouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	//handle rest apis
	restSubrouter := router.PathPrefix("/gr/api").Subrouter()
	adminSubrouter := restSubrouter.PathPrefix("/admin").Subrouter()

	// Admin V2 APIs
	adminSubrouter.HandleFunc("/v2/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetGroupsV2)).Methods("GET", "POST")
	adminSubrouter.HandleFunc("/v2/user/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetUserGroupsV2)).Methods("GET", "POST")
	adminSubrouter.HandleFunc("/v2/group/{id}", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetGroupV2)).Methods("GET")

	// Admin V1 APIs
	adminSubrouter.HandleFunc("/authman/synchronize", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.SynchronizeAuthman)).Methods("POST")
	adminSubrouter.HandleFunc("/user/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetUserGroups)).Methods("GET")
	adminSubrouter.HandleFunc("/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetAllGroups)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{id}", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.DeleteGroup)).Methods("DELETE")
	adminSubrouter.HandleFunc("/group/{group-id}/stats", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetGroupStats)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{group-id}/events", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetGroupEvents)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{group-id}/event/{event-id}", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.DeleteGroupEvent)).Methods("DELETE")
	adminSubrouter.HandleFunc("/group/{group-id}/posts", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetGroupPosts)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{group-id}/posts/{postID}", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.DeleteGroupPost)).Methods("DELETE")
	adminSubrouter.HandleFunc("/managed-group-configs", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetManagedGroupConfigs)).Methods("GET")
	adminSubrouter.HandleFunc("/managed-group-configs", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.CreateManagedGroupConfig)).Methods("POST")
	adminSubrouter.HandleFunc("/managed-group-configs", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.UpdateManagedGroupConfig)).Methods("PUT")
	adminSubrouter.HandleFunc("/managed-group-configs/{id}", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.DeleteManagedGroupConfig)).Methods("DELETE")
	adminSubrouter.HandleFunc("/sync-configs", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetSyncConfig)).Methods("GET")
	adminSubrouter.HandleFunc("/sync-configs", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.SaveSyncConfig)).Methods("PUT")

	// Internal key protection
	restSubrouter.HandleFunc("/int/user/{identifier}/groups", we.internalKeyAuthFunc(we.internalApisHandler.IntGetUserGroupMemberships)).Methods("GET")
	restSubrouter.HandleFunc("/int/group/{identifier}", we.internalKeyAuthFunc(we.internalApisHandler.IntGetGroup)).Methods("GET")
	restSubrouter.HandleFunc("/int/group/title/{title}/members", we.internalKeyAuthFunc(we.internalApisHandler.IntGetGroupMembersByGroupTitle)).Methods("GET")
	restSubrouter.HandleFunc("/int/authman/synchronize", we.internalKeyAuthFunc(we.internalApisHandler.SynchronizeAuthman)).Methods("POST")
	restSubrouter.HandleFunc("/int/stats", we.internalKeyAuthFunc(we.internalApisHandler.GroupStats)).Methods("GET")
	restSubrouter.HandleFunc("/int/group/{group-id}/events", we.internalKeyAuthFunc(we.internalApisHandler.CreateGroupEvent)).Methods("POST")
	restSubrouter.HandleFunc("/int/group/{group-id}/events/{event-id}", we.internalKeyAuthFunc(we.internalApisHandler.DeleteGroupEvent)).Methods("DELETE")
	restSubrouter.HandleFunc("/int/group/{group-id}/notification", we.internalKeyAuthFunc(we.internalApisHandler.SendGroupNotification)).Methods("POST")

	// V2 Client APIs
	restSubrouter.HandleFunc("/v2/groups", we.anonymousAuthWrapFunc(we.apisHandler.GetGroupsV2)).Methods("GET", "POST")
	restSubrouter.HandleFunc("/v2/groups/{id}", we.anonymousAuthWrapFunc(we.apisHandler.GetGroupV2)).Methods("GET")
	restSubrouter.HandleFunc("/v2/user/groups", we.idTokenAuthWrapFunc(we.apisHandler.GetUserGroupsV2)).Methods("GET", "POST")

	//V1 Client APIs
	restSubrouter.HandleFunc("/groups", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroup)).Methods("POST")
	restSubrouter.HandleFunc("/groups/{id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroup)).Methods("PUT")
	restSubrouter.HandleFunc("/user/groups", we.idTokenAuthWrapFunc(we.apisHandler.GetUserGroups)).Methods("GET")
	restSubrouter.HandleFunc("/user", we.idTokenAuthWrapFunc(we.apisHandler.DeleteUser)).Methods("DELETE")
	restSubrouter.HandleFunc("/user/login", we.idTokenAuthWrapFunc(we.apisHandler.LoginUser)).Methods("GET")
	restSubrouter.HandleFunc("/user/stats", we.idTokenAuthWrapFunc(we.apisHandler.GetUserStats)).Methods("GET")
	restSubrouter.HandleFunc("/group/{id}/stats", we.anonymousAuthWrapFunc(we.apisHandler.GetGroupStats)).Methods("GET")
	restSubrouter.HandleFunc("/group/{id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroup)).Methods("DELETE")

	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.CreatePendingMember)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.DeletePendingMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFunc(we.apisHandler.GetGroupMembers)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFunc(we.apisHandler.CreateMember)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFunc(we.apisHandler.DeleteMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/authman/synchronize", we.idTokenAuthWrapFunc(we.apisHandler.SynchAuthmanGroup)).Methods("POST")
	restSubrouter.HandleFunc("/memberships/{membership-id}/approval", we.idTokenAuthWrapFunc(we.apisHandler.MembershipApproval)).Methods("PUT")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteMembership)).Methods("DELETE")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateMembership)).Methods("PUT")

	restSubrouter.HandleFunc("/group/{group-id}/events", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroupEvent)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroupEvent)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{group-id}/event/{event-id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroupEvent)).Methods("DELETE")

	//extended client id token protection (eg. allow event managers)
	restSubrouter.HandleFunc("/user/group-memberships", we.idTokenExtendedClientAuthWrapFunc(we.apisHandler.GetUserGroupMemberships)).Methods("GET")

	// Client Post APIs
	restSubrouter.HandleFunc("/group/{groupID}/posts", we.idTokenAuthWrapFunc(we.apisHandler.GetGroupPosts)).Methods("GET")
	restSubrouter.HandleFunc("/group/{groupID}/posts", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroupPost)).Methods("POST")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.apisHandler.GetGroupPost)).Methods("GET")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroupPost)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}/reactions", we.idTokenAuthWrapFunc(we.apisHandler.ReactToGroupPost)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}/report/abuse", we.idTokenAuthWrapFunc(we.apisHandler.ReportAbuseGroupPost)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroupPost)).Methods("DELETE")

	//mixed protection
	restSubrouter.HandleFunc("/groups", we.mixedAuthWrapFunc(we.apisHandler.GetGroups)).Methods("GET")
	restSubrouter.HandleFunc("/groups/{id}", we.mixedAuthWrapFunc(we.apisHandler.GetGroup)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.mixedAuthWrapFunc(we.apisHandler.GetGroupEvents)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/events/v2", we.mixedAuthWrapFunc(we.apisHandler.GetGroupEventsV2)).Methods("GET")

	log.Fatal(http.ListenAndServe(":80", router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/gr/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we *Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

type apiKeyAuthFunc = func(string, http.ResponseWriter, *http.Request)

func (we Adapter) apiKeysAuthWrapFunc(handler apiKeyAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated := we.auth.apiKeyCheck(req)
		if !authenticated {
			log.Printf("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(clientID, w, req)
	}
}

type idTokenAuthFunc = func(string, *model.User, http.ResponseWriter, *http.Request)

func (we Adapter) idTokenAuthWrapFunc(handler idTokenAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, user := we.auth.idTokenCheck(w, req, false)
		if user == nil {
			log.Printf("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(clientID, user, w, req)
	}
}

func (we Adapter) anonymousAuthWrapFunc(handler idTokenAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, user := we.auth.idTokenCheck(w, req, true)
		if user == nil {
			log.Printf("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(clientID, user, w, req)
	}
}

func (we Adapter) idTokenExtendedClientAuthWrapFunc(handler idTokenAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, user := we.auth.customClientTokenCheck(w, req, we.auth.idTokenAuth.extendedClientIDs)
		if user == nil {
			log.Printf("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(clientID, user, w, req)
	}
}

func (we Adapter) internalKeyAuthFunc(handler apiKeyAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated := we.auth.internalAuthCheck(w, req)
		if !authenticated {
			log.Printf("%s %s Unauthorized error - Missing or wrong INTERNAL-API-KEY header", req.Method, req.URL.Path)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(clientID, w, req)
	}
}

func (we Adapter) mixedAuthWrapFunc(handler idTokenAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated, user := we.auth.mixedCheck(req)
		if !authenticated {
			log.Printf("Unauthorized - Mixed Check")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		//user can be nil
		handler(clientID, user, w, req)
	}
}

type adminAuthFunc = func(string, *model.User, http.ResponseWriter, *http.Request)

func (we Adapter) adminIDTokenAuthWrapFunc(handler adminAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, user, forbidden := we.auth.adminCheck(req)
		if user == nil {
			if forbidden {
				log.Printf("Forbidden - Admin")
				w.WriteHeader(http.StatusForbidden)
			} else {
				log.Printf("Unauthorized - Admin")
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		handler(clientID, user, w, req)
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(app *core.Application, host string, supportedClientIDs []string, appKeys []string, oidcProvider string, oidcClientID string,
	oidcExtendedClientIDs string, oidcAdminClientID string, oidcAdminWebClientID string,
	internalAPIKey string, authService *authservice.AuthService, groupServiceURL string) *Adapter {
	authorization := casbin.NewEnforcer("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")

	auth := NewAuth(app, host, supportedClientIDs, appKeys, internalAPIKey, oidcProvider, oidcClientID, oidcExtendedClientIDs, oidcAdminClientID,
		oidcAdminWebClientID, authService, groupServiceURL, authorization)
	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	internalApisHandler := rest.NewInternalApisHandler(app)

	return &Adapter{host: host, auth: auth, apisHandler: apisHandler, adminApisHandler: adminApisHandler, internalApisHandler: internalApisHandler}
}
