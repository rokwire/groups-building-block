package web

import (
	"context"
	"fmt"
	"groups/core"
	"log"
	"net/http"

	"gopkg.in/ericchiang/go-oidc.v2"
)

//Auth handler
type Auth struct {
	apiKeysAuth *APIKeysAuth
	idTokenAuth *IDTokenAuth
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

func (auth *Auth) idTokenCheck(w http.ResponseWriter, r *http.Request) bool {
	return auth.idTokenAuth.check(w, r)
}

//NewAuth creates new auth handler
func NewAuth(app *core.Application, appKeys []string, oidcProvider string, oidcClientID string) *Auth {
	apiKeysAuth := newAPIKeysAuth(appKeys)
	idTokenAuth := newIDTokenAuth(app, oidcProvider, oidcClientID)

	auth := Auth{apiKeysAuth: apiKeysAuth, idTokenAuth: idTokenAuth}
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

////////////////////////////////////

//IDTokenAuth entity
type IDTokenAuth struct {
	app *core.Application

	idTokenVerifier *oidc.IDTokenVerifier
}

func (auth *IDTokenAuth) check(w http.ResponseWriter, r *http.Request) bool {
	//TODO
	log.Println("Make ID Token check")

	rawIDToken := "12345"

	log.Println(rawIDToken)

	// Parse and verify ID Token payload.
	idToken, err := auth.idTokenVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		// handle error
		log.Println(err)
	}

	// Extract custom claims
	var claims struct {
		UIuceduUIN        *string   `json:"uiucedu_uin"`
		Sub               *string   `json:"sub"`
		Email             *string   `json:"email"`
		UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
	}
	if err := idToken.Claims(&claims); err != nil {
		// handle error
		log.Println(err)
	}
	log.Println(claims)

	for _, item := range *claims.UIuceduIsMemberOf {
		log.Println(item)
	}

	return true
}

//newIDTokenAuth creates new id token auth
func newIDTokenAuth(app *core.Application, oidcProvider string, oidcClientID string) *IDTokenAuth {
	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}
	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oidcClientID})

	auth := IDTokenAuth{app: app, idTokenVerifier: idTokenVerifier}
	return &auth
}
