package web

import (
	"fmt"
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

	apisHandler *rest.ApisHandler
}

// @title Rokwire Groups Building Block API
// @description Rokwire Groups Building Block API Documentation.
// @version 1.0.2
// @host localhost
// @BasePath /gr
// @schemes https

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey AppUserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

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
	restSubrouter := router.PathPrefix("/gr/api").Subrouter()

	//api key protection
	restSubrouter.HandleFunc("/group-categories", we.apiKeysAuthWrapFunc(we.apisHandler.GetGroupCategories)).Methods("GET")

	//id token protection
	restSubrouter.HandleFunc("/groups", we.idTokenAuthWrapFunc(we.apisHandler.CreateGroup)).Methods("POST")
	restSubrouter.HandleFunc("/groups/{id}", we.idTokenAuthWrapFunc(we.apisHandler.UpdateGroup)).Methods("PUT")
	restSubrouter.HandleFunc("/user/groups", we.idTokenAuthWrapFunc(we.apisHandler.GetUserGroups)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFuncOld(we.apisHandler.CreatePendingMember)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/pending-members", we.idTokenAuthWrapFuncOld(we.apisHandler.DeletePendingMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/group/{group-id}/members", we.idTokenAuthWrapFuncOld(we.apisHandler.DeleteMember)).Methods("DELETE")
	restSubrouter.HandleFunc("/memberships/{membership-id}/approval", we.idTokenAuthWrapFuncOld(we.apisHandler.MembershipApproval)).Methods("PUT")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFuncOld(we.apisHandler.DeleteMembership)).Methods("DELETE")
	restSubrouter.HandleFunc("/memberships/{membership-id}", we.idTokenAuthWrapFuncOld(we.apisHandler.UpdateMembership)).Methods("PUT")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.idTokenAuthWrapFuncOld(we.apisHandler.CreateGroupEvent)).Methods("POST")
	restSubrouter.HandleFunc("/group/{group-id}/event/{event-id}", we.idTokenAuthWrapFuncOld(we.apisHandler.DeleteGroupEvent)).Methods("DELETE")

	//mixed protection
	restSubrouter.HandleFunc("/groups", we.mixedAuthWrapFuncOld(we.apisHandler.GetGroups)).Methods("GET")
	restSubrouter.HandleFunc("/groups/{id}", we.mixedAuthWrapFuncOld(we.apisHandler.GetGroup)).Methods("GET")
	restSubrouter.HandleFunc("/group/{group-id}/events", we.mixedAuthWrapFuncOld(we.apisHandler.GetGroupEvents)).Methods("GET")

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

//TODO
type idTokenAuthFuncOld = func(*model.User, http.ResponseWriter, *http.Request)

func (we Adapter) idTokenAuthWrapFuncOld(handler idTokenAuthFuncOld) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, user := we.auth.idTokenCheck(w, req)
		//TODO
		log.Println(clientID)
		if user == nil {
			return
		}

		handler(user, w, req)
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

func (we Adapter) mixedAuthWrapFuncOld(handler idTokenAuthFuncOld) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		clientID, authenticated, user := we.auth.mixedCheck(w, req)
		//TODO
		log.Println(clientID)
		if !authenticated {
			return
		}

		//user can be nil
		handler(user, w, req)
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

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(app *core.Application, host string, appKeys []string, oidcProvider string, oidcClientID string) *Adapter {
	auth := NewAuth(app, appKeys, oidcProvider, oidcClientID)
	apisHandler := rest.NewApisHandler(app)
	return &Adapter{host: host, auth: auth, apisHandler: apisHandler}
}
