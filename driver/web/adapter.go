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
	apisHandler *rest.ApisHandler
}

//Start starts the web server
func (we *Adapter) Start() {
	router := mux.NewRouter().StrictSlash(true)

	//handle rest apis
	restSubrouter := router.PathPrefix("/groups/api").Subrouter()
	restSubrouter.HandleFunc("/test", we.wrapFunc(we.apisHandler.Test)).Methods("GET")

	log.Fatal(http.ListenAndServe(":80", router))
}

func (we *Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter() *Adapter {
	apisHandler := rest.NewApisHandler()
	return &Adapter{apisHandler: apisHandler}
}
