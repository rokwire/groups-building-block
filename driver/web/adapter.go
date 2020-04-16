package web

import (
	"groups/driver/web/rest"
	"groups/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//Adapter entity
type Adapter struct {
	auth *Auth

	apisHandler *rest.ApisHandler
}

//Start starts the web server
func (we *Adapter) Start() {
	//start the auth module
	err := we.auth.Start()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)

	//handle rest apis
	restSubrouter := router.PathPrefix("/groups/api").Subrouter()
	restSubrouter.HandleFunc("/test", we.idTokenAuthWrapFunc(we.apisHandler.Test)).Methods("GET")

	log.Fatal(http.ListenAndServe(":80", router))
}

func (we *Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

func (we Adapter) apiKeysAuthWrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated := we.auth.apiKeyCheck(w, req)
		if !authenticated {
			return
		}

		handler(w, req)
	}
}

func (we Adapter) idTokenAuthWrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated := we.auth.idTokenCheck(w, req)
		if !authenticated {
			return
		}

		handler(w, req)
	}
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(appKeys []string, oidcProvider string, oidcClientID string) *Adapter {
	auth := NewAuth(appKeys, oidcProvider, oidcClientID)
	apisHandler := rest.NewApisHandler()
	return &Adapter{auth: auth, apisHandler: apisHandler}
}
