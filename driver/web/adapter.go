package web

import (
	"fmt"
	"github.com/casbin/casbin"
	"groups/core"
	"groups/core/model"
	"groups/driver/web/rest"
	"groups/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

//Adapter entity
type Adapter struct {
	host string
	auth *Auth

	authorization *casbin.Enforcer

	apisHandler      *rest.ApisHandler
	adminApisHandler *rest.AdminApisHandler
}

// @title Rokwire Groups Building Block API
// @description Rokwire Groups Building Block API Documentation.
// @version 1.4.5
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
	//start the auth module
	err := we.auth.Start()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)

	subrouter := router.PathPrefix("/gr").Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	subrouter.HandleFunc("/doc", we.serveDoc)
	subrouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	//handle rest apis
	//handle rest apis
	restSubrouter := router.PathPrefix("/gr/api").Subrouter()
	adminSubrouter := restSubrouter.PathPrefix("/admin").Subrouter()

	// Admin APIs
	adminSubrouter.HandleFunc("/groups", we.adminAppIDTokenAuthWrapFunc(we.adminApisHandler.GetGroups)).Methods("GET")

	//internal key protection
	restSubrouter.HandleFunc("/int/user/{identifier}/groups", we.internalKeyAuthFunc(we.apisHandler.GetUserGroupMemberships)).Methods("GET")

	//api key protection
	restSubrouter.HandleFunc("/group-categories", we.apiKeysAuthWrapFunc(we.apisHandler.GetGroupCategories)).Methods("GET")

	//id token protection
	restSubrouter.HandleFunc("/groups", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroup)).Methods("POST")
	restSubrouter.HandleFunc("/groups/{id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroup)).Methods("PUT")
	restSubrouter.HandleFunc("/user/groups", we.idTokenAuthWrapFunc(we.apisHandler.GetUserGroups)).Methods("GET")
	restSubrouter.HandleFunc("/group/{id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroup)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.CreatePendingMember)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFunc(we.apisHandler.DeletePendingMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFunc(we.apisHandler.DeleteMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/memberships/{membership-id}/approval", we.idTokenAuthWrapFunc(we.apisHandler.MembershipApproval)).Methods("PUT")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteMembership)).Methods("DELETE")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateMembership)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroupEvent)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/event/{event-id}", we.idTokenAuthWrapFunc(we.apisHandler.DeleteGroupEvent)).Methods("DELETE")

	//mixed protection
	restSubrouter.HandleFunc("/groups", we.mixedAuthWrapFunc(we.apisHandler.GetGroups)).Methods("GET")
	restSubrouter.HandleFunc("/groups/{id}", we.mixedAuthWrapFunc(we.apisHandler.GetGroup)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.mixedAuthWrapFunc(we.apisHandler.GetGroupEvents)).Methods("GET")

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

		clientID, authenticated := we.auth.apiKeyCheck(w, req)
		if !authenticated {
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
			return
		}

		handler(clientID, user, w, req)
	}
}

type internalKeyAuthFunc = func(string, http.ResponseWriter, *http.Request)

func (we Adapter) internalKeyAuthFunc(handler apiKeyAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated := we.auth.internalAuthCheck(w, req)
		if !authenticated {
			return
		}

		handler(clientID, w, req)
	}
}

func (we Adapter) mixedAuthWrapFunc(handler idTokenAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated, user := we.auth.mixedCheck(w, req)
		if !authenticated {
			return
		}

		//user can be nil
		handler(clientID, user, w, req)
	}
}

type adminAuthFunc = func(string, *model.User, http.ResponseWriter, *http.Request)

func (we Adapter) adminAppIDTokenAuthWrapFunc(handler adminAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		var HasAccess bool = false
		HasAccess, clientID := we.auth.clientIDCheck(w, req)
		if !HasAccess {
			return
		}

		ok, user, shiboUser := we.auth.adminCheck(clientID, w, req)
		if !ok {
			return
		}

		obj := req.URL.Path // the resource that is going to be accessed.
		act := req.Method   // the operation that the user performs on the resource.

		for _, s := range *shiboUser.IsMemberOf {
			HasAccess = we.authorization.Enforce(s, obj, act)
			if HasAccess {
				break
			}
		}

		if !HasAccess {
			log.Printf("Access control error - UIN: %s is trying to apply %s operation for %s\n", shiboUser.Uin, act, obj)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		handler(*clientID, user, w, req)
	}
}

func (auth *Auth) adminCheck(clientID *string, w http.ResponseWriter, r *http.Request) (bool, *model.User, *model.ShibbolethAuth) {
	return auth.adminAuth.check(clientID, w, r)
}

func (auth *AdminAuth) check(clientID *string, w http.ResponseWriter, r *http.Request) (bool, *model.User, *model.ShibbolethAuth) {

	//1. Get the token from the request
	rawIDToken, tokenType, err := auth.getIDToken(r)
	if err != nil {
		auth.responseBadRequest(w)
		return false, nil, nil
	}

	//3. Validate the token
	idToken, err := auth.verify(*rawIDToken, *tokenType)
	if err != nil {
		log.Printf("error validating token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil, nil
	}

	//4. Get the user data from the token
	var userData userData
	if err := idToken.Claims(&userData); err != nil {
		log.Printf("error getting user data from token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil, nil
	}
	//we must have UIuceduUIN
	if userData.UIuceduUIN == nil {
		log.Printf("error - missing uiuceuin data in the token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil, nil
	}

	//4. Get the user for the provided external id.
	user, err := auth.getUser(clientID, userData)
	if err != nil {
		log.Printf("error getting an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return false, nil, nil
	}
	if user == nil {
		log.Printf("for some reasons the user for external id - %s is nil\n", err)

		auth.responseInternalServerError(w)
		return false, nil, nil
	}

	//5. Update the user if needed
	user, err = auth.updateUserIfNeeded(*clientID, *user, userData)
	if err != nil {
		log.Printf("error updating an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return false, nil, nil
	}

	shibboAuth := &model.ShibbolethAuth{Uin: *userData.UIuceduUIN, Email: *userData.Email,
		IsMemberOf: userData.UIuceduIsMemberOf}

	return true, user, shibboAuth
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(app *core.Application, host string, appKeys []string, oidcProvider string, oidcClientID string, oidcAdminClientID string, oidcAdminWebClientID string, internalAPIKeys []string) *Adapter {
	auth := NewAuth(app, appKeys, internalAPIKeys, oidcProvider, oidcClientID, oidcAdminClientID, oidcAdminWebClientID)
	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)

	authorization := casbin.NewEnforcer("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")

	return &Adapter{host: host, auth: auth, apisHandler: apisHandler, authorization: authorization, adminApisHandler: adminApisHandler}
}
