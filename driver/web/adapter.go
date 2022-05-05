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

//Adapter entity
type Adapter struct {
	host string
	auth *Auth

	apisHandler         *rest.ApisHandler
	adminApisHandler    *rest.AdminApisHandler
	internalApisHandler *rest.InternalApisHandler
}

// @title Rokwire Groups Building Block API
// @description Rokwire Groups Building Block API Documentation.
// @version 1.5.31
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
// @name ROKWIRE_GS_API_KEY

//Start starts the web server
func (we *Adapter) Start() {
	router := mux.NewRouter().StrictSlash(true)

	subrouter := router.PathPrefix("/gr").Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	subrouter.HandleFunc("/doc", we.serveDoc)
	subrouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	//handle rest apis
	restSubrouter := router.PathPrefix("/gr/api").Subrouter()
	adminSubrouter := restSubrouter.PathPrefix("/admin").Subrouter()

	// Admin APIs
	adminSubrouter.HandleFunc("/user/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetUserGroups)).Methods("GET")
	adminSubrouter.HandleFunc("/user/login", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.LoginUser)).Methods("GET")
	adminSubrouter.HandleFunc("/groups", we.adminIDTokenAuthWrapFunc(we.adminApisHandler.GetAllGroups)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{id}", we.idTokenAuthWrapFunc(we.adminApisHandler.DeleteGroup)).Methods("DELETE")
	adminSubrouter.HandleFunc("/group/{group-id}/events", we.mixedAuthWrapFunc(we.adminApisHandler.GetGroupEvents)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{group-id}/event/{event-id}", we.idTokenAuthWrapFunc(we.adminApisHandler.DeleteGroupEvent)).Methods("DELETE")
	adminSubrouter.HandleFunc("/group/{groupID}/posts", we.idTokenAuthWrapFunc(we.adminApisHandler.GetGroupPosts)).Methods("GET")
	adminSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.adminApisHandler.DeleteGroupPost)).Methods("DELETE")

	//internal key protection
	restSubrouter.HandleFunc("/int/user/{identifier}/groups", we.internalKeyAuthFunc(we.internalApisHandler.IntGetUserGroupMemberships)).Methods("GET")
	restSubrouter.HandleFunc("/int/authman/synchronize", we.internalKeyAuthFunc(we.internalApisHandler.SynchronizeAuthman)).Methods("GET")
	restSubrouter.HandleFunc("/int/stats", we.internalKeyAuthFunc(we.internalApisHandler.GroupStats)).Methods("GET")

	//api key protection
	restSubrouter.HandleFunc("/group-categories", we.apiKeysAuthWrapFunc(we.apisHandler.GetGroupCategories)).Methods("GET")

	//id token protection
	restSubrouter.HandleFunc("/groups", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroup)).Methods("POST")
	restSubrouter.HandleFunc("/groups/{id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroup)).Methods("PUT")
	restSubrouter.HandleFunc("/user/groups", we.idTokenAuthWrapFunc(we.apisHandler.GetUserGroups)).Methods("GET")
	restSubrouter.HandleFunc("/user", we.idTokenAuthWrapFunc(we.apisHandler.DeleteUser)).Methods("DELETE")
	restSubrouter.HandleFunc("/user/login", we.idTokenAuthWrapFunc(we.apisHandler.LoginUser)).Methods("GET")
	restSubrouter.HandleFunc("/user/stats", we.idTokenAuthWrapFunc(we.apisHandler.GetUserStats)).Methods("GET")
	restSubrouter.HandleFunc("/group/{id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroup)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.CreatePendingMember)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.DeletePendingMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFunc(we.apisHandler.DeleteMember)).Methods("DELETE")
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
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroupPost)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{groupID}/posts/{postID}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroupPost)).Methods("DELETE")

	restSubrouter.HandleFunc("/group/{groupID}/polls", we.idTokenAuthWrapFunc(we.apisHandler.GetGroupPolls)).Methods("GET")
	restSubrouter.HandleFunc("/group/{groupID}/polls", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroupPoll)).Methods("POST")
	restSubrouter.HandleFunc("/group/{groupID}/polls/{pollID}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroupPoll)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{groupID}/polls/{pollID}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroupPoll)).Methods("DELETE")

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

		clientID, user := we.auth.idTokenCheck(w, req)
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
			log.Printf("Unauthorized - Internal Key")
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

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(app *core.Application, host string, appKeys []string, oidcProvider string, oidcClientID string, oidcExtendedClientIDs string, oidcAdminClientID string,
	oidcAdminWebClientID string, internalAPIKeys []string, authService *authservice.AuthService, groupServiceURL string) *Adapter {
	authorization := casbin.NewEnforcer("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")

	auth := NewAuth(app, host, appKeys, internalAPIKeys, oidcProvider, oidcClientID, oidcExtendedClientIDs, oidcAdminClientID, oidcAdminWebClientID, authService, groupServiceURL, authorization)
	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	internalApisHandler := rest.NewInternalApisHandler(app)

	return &Adapter{host: host, auth: auth, apisHandler: apisHandler, adminApisHandler: adminApisHandler, internalApisHandler: internalApisHandler}
}
