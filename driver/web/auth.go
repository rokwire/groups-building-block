package web

import (
	"context"
	"fmt"
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
	"strings"

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

func (auth *Auth) idTokenCheck(w http.ResponseWriter, r *http.Request) *model.User {
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

func (auth *IDTokenAuth) check(w http.ResponseWriter, r *http.Request) *model.User {
	//1. Get the token from the request
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		auth.responseBadRequest(w)
		return nil
	}
	splitAuthorization := strings.Fields(authorizationHeader)
	if len(splitAuthorization) != 2 {
		auth.responseBadRequest(w)
		return nil
	}
	// expected - Bearer 1234
	if splitAuthorization[0] != "Bearer" {
		auth.responseBadRequest(w)
		return nil
	}
	rawIDToken := splitAuthorization[1]

	//2. Validate the token
	idToken, err := auth.idTokenVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Printf("error validating token - %s\n", err)

		auth.responseUnauthorized(rawIDToken, w)
		return nil
	}

	//3. Get the user data from the token
	var userData struct {
		UIuceduUIN        *string   `json:"uiucedu_uin"`
		Sub               *string   `json:"sub"`
		Email             *string   `json:"email"`
		UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
	}
	if err := idToken.Claims(&userData); err != nil {
		log.Printf("error getting user data from token - %s\n", err)

		auth.responseUnauthorized(rawIDToken, w)
		return nil
	}
	//we must have UIuceduUIN
	if userData.UIuceduUIN == nil {
		log.Printf("missing uiuceuin data in the token - %s\n", err)

		auth.responseUnauthorized(rawIDToken, w)
		return nil
	}

	//4. Check if we have an user with the provided external id.
	foundedUser, err := auth.app.FindUser(*userData.UIuceduUIN)
	log.Println(foundedUser)

	//TODO

	return &model.User{ID: "123456789"}
}

func (auth *IDTokenAuth) responseBadRequest(w http.ResponseWriter) {
	log.Println(fmt.Sprintf("400 - Bad Request"))

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func (auth *IDTokenAuth) responseUnauthorized(token string, w http.ResponseWriter) {
	log.Println(fmt.Sprintf("401 - Unauthorized for token %s", token))

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
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
