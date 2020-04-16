package web

import (
	"fmt"
	"log"
	"net/http"
)

//Auth handler
type Auth struct {
	apiKeysAuth *APIKeysAuth
}

//Start starts the auth module
func (auth *Auth) Start() error {
	//TODO
	//go auth.checkForHangingStates()

	return nil
}

func (auth *Auth) apiKeyCheck(w http.ResponseWriter, r *http.Request) bool {
	return auth.apiKeysAuth.check(w, r)
}

//NewAuth creates new auth handler
func NewAuth(appKeys []string) *Auth {
	auth := Auth{
		apiKeysAuth: newAPIKeysAuth(appKeys)}
	return &auth
}

/////////////////////////////////////

//APIKeysAuth entity
type APIKeysAuth struct {
	appKeys []string
}

func (auth *APIKeysAuth) check(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("ROKWIRE-API-KEY")
	//check if there is api key in the header
	if len(apiKey) == 0 {
		//no key, so return 400
		log.Println(fmt.Sprintf("400 - Bad Request"))

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return false
	}

	//check if the api key is one of the listed
	appKeys := auth.appKeys
	exist := false
	for _, element := range appKeys {
		if element == apiKey {
			exist = true
			break
		}
	}
	if !exist {
		//not exist, so return 401
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", apiKey))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

//NewAPIKeysAuth creates new api keys auth
func newAPIKeysAuth(appKeys []string) *APIKeysAuth {
	auth := APIKeysAuth{appKeys}
	return &auth
}
