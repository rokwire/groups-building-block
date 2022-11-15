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
	"context"
	"errors"
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/sync/syncmap"

	"github.com/casbin/casbin"
	"github.com/rokwire/core-auth-library-go/authorization"
	"github.com/rokwire/core-auth-library-go/authservice"
	"github.com/rokwire/core-auth-library-go/authutils"
	"github.com/rokwire/core-auth-library-go/tokenauth"
)

// Auth handler
type Auth struct {
	apiKeysAuth  *APIKeysAuth
	idTokenAuth  *IDTokenAuth
	internalAuth *InternalAuth
	adminAuth    *AdminAuth

	supportedClients []string
}

func (auth *Auth) clientIDCheck(r *http.Request) (bool, string) {
	clientID := r.Header.Get("APP")
	if len(clientID) == 0 {
		clientID = "edu.illinois.rokwire"
	}

	//check if supported
	for _, s := range auth.supportedClients {
		if s == clientID {
			return true, clientID
		}
	}
	return false, ""
}

func (auth *Auth) apiKeyCheck(r *http.Request) (string, bool) {
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		return "", false
	}

	apiKey := auth.getAPIKey(r)
	authenticated := auth.apiKeysAuth.check(apiKey, r)

	return clientID, authenticated
}

func (auth *Auth) idTokenCheck(w http.ResponseWriter, r *http.Request) (string, *model.User) {
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		return "", nil
	}

	idToken := auth.getIDToken(r)
	user := auth.idTokenAuth.check(clientID, idToken, false, nil, r)
	return clientID, user
}

func (auth *Auth) customClientTokenCheck(w http.ResponseWriter, r *http.Request, allowedOIDCClientIDs []string) (string, *model.User) {
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		return "", nil
	}

	idToken := auth.getIDToken(r)
	user := auth.idTokenAuth.check(clientID, idToken, false, allowedOIDCClientIDs, r)
	return clientID, user
}

func (auth *Auth) internalAuthCheck(w http.ResponseWriter, r *http.Request) (string, bool) {
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		log.Printf("%s %s error - missing or bad APP header", r.Method, r.URL.Path)
		return "", false
	}

	internalAuthKey := auth.getInternalAPIKey(r)
	authenticated := auth.internalAuth.check(internalAuthKey, w)

	return clientID, authenticated
}

func (auth *Auth) mixedCheck(r *http.Request) (string, bool, *model.User) {
	//get client ID
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		return "", false, nil
	}

	//first check for id token
	idToken := auth.getIDToken(r)
	if idToken != nil && len(*idToken) > 0 {
		authenticated := false
		user := auth.idTokenAuth.check(clientID, idToken, true, nil, r)
		if user != nil {
			authenticated = true
		}
		return clientID, authenticated, user
	}

	//check api key
	apiKey := auth.getAPIKey(r)
	if apiKey != nil && len(*apiKey) > 0 {
		authenticated := auth.apiKeysAuth.check(apiKey, r)
		return clientID, authenticated, nil
	}
	return clientID, false, nil
}

func (auth *Auth) adminCheck(r *http.Request) (string, *model.User, bool) {
	clientIDOK, clientID := auth.clientIDCheck(r)
	if !clientIDOK {
		return "", nil, false
	}

	user, forbidden := auth.adminAuth.check(clientID, r)
	return clientID, user, forbidden
}

func (auth *Auth) getAPIKey(r *http.Request) *string {
	apiKey := r.Header.Get("ROKWIRE-API-KEY")
	if len(apiKey) == 0 {
		return nil
	}
	return &apiKey
}

// TBD Remove the legacy API key functionality
func (auth *Auth) getInternalAPIKey(r *http.Request) *string {
	internalAPIKey := r.Header.Get("INTERNAL-API-KEY")
	if len(internalAPIKey) > 0 {
		return &internalAPIKey
	}

	legacyAPIKey := r.Header.Get("ROKWIRE_GS_API_KEY")
	if len(legacyAPIKey) > 0 {
		return &legacyAPIKey
	}
	return nil
}

func (auth *Auth) getIDToken(r *http.Request) *string {
	// get the token from the request
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		return nil
	}
	splitAuthorization := strings.Fields(authorizationHeader)
	if len(splitAuthorization) != 2 {
		return nil
	}
	// expected - Bearer 1234
	if splitAuthorization[0] != "Bearer" {
		return nil
	}
	idToken := splitAuthorization[1]
	return &idToken
}

// NewAuth creates new auth handler
func NewAuth(app *core.Application, host string, supportedClientIDs []string, appKeys []string, internalAPIKey string, oidcProvider string, oidcClientID string, oidcExtendedClientIDs string,
	oidcAdminClientID string, oidcAdminWebClientID string, authService *authservice.AuthService, groupServiceURL string, adminAuthorization *casbin.Enforcer) *Auth {
	var tokenAuth *tokenauth.TokenAuth
	if authService != nil {
		permissionAuth := authorization.NewCasbinStringAuthorization("driver/web/permissions_authorization_policy.csv")
		scopeAuth := authorization.NewCasbinScopeAuthorization("driver/web/scope_authorization_policy.csv", authService.GetServiceID())

		// Instantiate TokenAuth instance to perform token validation
		var err error
		tokenAuth, err = tokenauth.NewTokenAuth(true, authService, permissionAuth, scopeAuth)
		if err != nil {
			log.Fatalf("error instancing token auth: %s", err)
		}
	}

	apiKeysAuth := newAPIKeysAuth(appKeys, tokenAuth)
	idTokenAuth := newIDTokenAuth(app, oidcProvider, oidcClientID, oidcExtendedClientIDs, tokenAuth)
	internalAuth := newInternalAuth(internalAPIKey)
	adminAuth := newAdminAuth(app, oidcProvider, oidcAdminClientID, oidcAdminWebClientID, tokenAuth, adminAuthorization)

	auth := Auth{apiKeysAuth: apiKeysAuth, idTokenAuth: idTokenAuth, internalAuth: internalAuth, adminAuth: adminAuth, supportedClients: supportedClientIDs}
	return &auth
}

/////////////////////////////////////

// APIKeysAuth entity
type APIKeysAuth struct {
	appKeys []string

	coreTokenAuth *tokenauth.TokenAuth
}

func (auth *APIKeysAuth) check(apiKey *string, r *http.Request) bool {
	//check if there is api key in the header
	if apiKey == nil || len(*apiKey) == 0 {
		if auth.coreTokenAuth != nil {
			_, err := auth.coreTokenAuth.CheckRequestTokens(r)
			if err == nil {
				return true
			}
			return false
		}
		return false
	}

	//check if the api key is one of the listed
	appKeys := auth.appKeys
	exist := false
	for _, element := range appKeys {
		if element == *apiKey {
			exist = true
			break
		}
	}
	if !exist {
		return false
	}
	return true
}

// NewAPIKeysAuth creates new api keys auth
func newAPIKeysAuth(appKeys []string, coreTokenAuth *tokenauth.TokenAuth) *APIKeysAuth {
	auth := APIKeysAuth{appKeys, coreTokenAuth}
	return &auth
}

////////////////////////////////////

type userData struct {
	UIuceduUIN  *string  `json:"uiucedu_uin"`
	Sub         *string  `json:"sub"`
	Aud         *string  `json:"aud"`
	Email       *string  `json:"email"`
	Name        *string  `json:"name"`
	Uin         *string  `json:"uin"`
	NetID       *string  `json:"net_id"`
	Permissions []string `json:"uiucedu_is_member_of"`
}

////////////////////////////////////

// InternalAuth entity
type InternalAuth struct {
	internalAPIKey string
}

// TBD Remove the legacy API key functionality
func (auth *InternalAuth) check(internalAPIKey *string, w http.ResponseWriter) bool {
	//check if there is internal key in the header
	if internalAPIKey == nil || len(*internalAPIKey) == 0 {
		return false
	}

	if *internalAPIKey == auth.internalAPIKey {
		return true
	}

	return false
}

// newInternalAuth creates new internal auth
func newInternalAuth(internalAPIKey string) *InternalAuth {
	auth := InternalAuth{
		internalAPIKey: internalAPIKey,
	}
	return &auth
}

///////////////////////////////////

// IDTokenAuth entity
type IDTokenAuth struct {
	app *core.Application

	idTokenVerifier   *oidc.IDTokenVerifier
	appClientIDs      []string
	extendedClientIDs []string

	coreTokenAuth *tokenauth.TokenAuth

	cachedUsers            *syncmap.Map //cache users while active - 5 minutes timeout
	cachedUsersLock        *sync.RWMutex
	cachedUsersLockMapping map[string]*sync.Mutex
}

func (auth *IDTokenAuth) check(clientID string, token *string, allowAnonymousCoreToken bool, allowedClientIDs []string, r *http.Request) *model.User {
	var data *userData
	var isCoreUser = false
	var isAnonymous = false
	var coreErr error
	if auth.coreTokenAuth != nil {
		var claims *tokenauth.Claims
		claims, coreErr = auth.coreTokenAuth.CheckRequestTokens(r)
		if coreErr == nil && claims != nil && (allowAnonymousCoreToken || !claims.Anonymous) {
			err := auth.coreTokenAuth.AuthorizeRequestScope(claims, r)
			if err != nil {
				return nil
			}

			var netID *string
			if claims.ExternalIDs != nil {
				if value, ok := claims.ExternalIDs["net_id"]; ok {
					netID = &value
				}
			}

			log.Printf("Authentication successful for user: %v", claims)
			permissions := strings.Split(claims.Permissions, ",")
			data = &userData{Sub: &claims.Subject, Email: &claims.Email, Name: &claims.Name,
				Permissions: permissions, UIuceduUIN: &claims.UID, NetID: netID}
			isCoreUser = true
			isAnonymous = claims.Anonymous
		}
	}

	if data == nil {
		//Return error from core validation if OIDC is not configured
		if auth.idTokenVerifier == nil {
			log.Printf("error validating token - %s\n", coreErr)
			return nil
		}

		//1. Check if there is a token
		if token == nil || len(*token) == 0 {
			return nil
		}
		rawIDToken := *token

		//2. Validate the token
		idToken, err := auth.idTokenVerifier.Verify(context.Background(), rawIDToken)
		if err != nil {
			log.Printf("error validating token - %s\n", err)
			return nil
		}

		//3. Get the user data from the token
		if err := idToken.Claims(&data); err != nil {
			log.Printf("error getting user data from token - %s\n", err)
			return nil
		}

		if allowedClientIDs == nil {
			allowedClientIDs = auth.appClientIDs
		}

		validAud := false
		if data.Aud != nil {
			validAud = authutils.ContainsString(allowedClientIDs, *data.Aud)
		}
		if !validAud {
			log.Printf("invalid audience in token: expected %v, got %s\n", allowedClientIDs, *data.Aud)
			return nil
		}

		//we must have UIuceduUIN
		if data.UIuceduUIN == nil {
			log.Printf("missing uiuceuin data in the token - %s\n", err)
			return nil
		}
	}

	if data == nil {
		log.Println("nil user data")
		return nil
	}

	// 4. Use corebb user id or legacy user id
	// NOTE: In general we assume the corebb user is already refactored e.g the login API has been invoked at least once.
	// The difference would be only the user ID.
	var userID string
	if isCoreUser {
		userID = *data.Sub
	} else {
		persistedUser, err := auth.app.FindUser(clientID, data.UIuceduUIN, true)
		if err != nil {
			log.Printf("error retriving user (UIuceduUIN: %s): %s", *data.UIuceduUIN, err)
		}
		if persistedUser != nil {
			isCoreUser = persistedUser.IsCoreUser
			userID = persistedUser.ID
		} else {
			legacyUser, err := auth.app.CreateUser(clientID, uuid.NewString(), data.UIuceduUIN, data.Email, data.Name)
			if err != nil {
				log.Printf("error creating legacy user (UIuceduUIN: %s): %s", *data.UIuceduUIN, err)
			}
			if legacyUser != nil {
				userID = legacyUser.ID
			}
		}
	}

	//5. Get the user for the provided external id.
	var name, externalID, email, netID string
	if data.Name != nil {
		name = *data.Name
	}
	if data.UIuceduUIN != nil {
		externalID = *data.UIuceduUIN
	}
	if data.Email != nil {
		email = *data.Email
	}
	if data.NetID != nil {
		netID = *data.NetID
	}
	return &model.User{
		ID: userID, ClientID: clientID, ExternalID: externalID, NetID: netID,
		Email: email, Name: name, IsCoreUser: isCoreUser, IsAnonymous: isAnonymous,
		Permissions: data.Permissions,
	}
}

func (auth *IDTokenAuth) responseBadRequest(w http.ResponseWriter) {
	log.Println("400 - Bad Request")

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func (auth *IDTokenAuth) responseUnauthorized(token string, w http.ResponseWriter) {
	log.Printf("401 - Unauthorized for token %s", token)

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func (auth *IDTokenAuth) responseInternalServerError(w http.ResponseWriter) {
	log.Printf("500 - Internal Server Error")

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

// newIDTokenAuth creates new id token auth
func newIDTokenAuth(app *core.Application, oidcProvider string, appClientIDs string, extendedClientIDs string, coreTokenAuth *tokenauth.TokenAuth) *IDTokenAuth {
	var idTokenVerifier *oidc.IDTokenVerifier
	if len(oidcProvider) != 0 {
		provider, err := oidc.NewProvider(context.Background(), oidcProvider)
		if err != nil {
			log.Fatalln(err)
		}
		idTokenVerifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: true})
	}

	cacheUsers := &syncmap.Map{}
	lock := &sync.RWMutex{}

	appClientIDList := strings.Split(appClientIDs, ",")
	extendedClientIDList := strings.Split(extendedClientIDs, ",")

	auth := IDTokenAuth{app: app, idTokenVerifier: idTokenVerifier, coreTokenAuth: coreTokenAuth,
		appClientIDs: appClientIDList, extendedClientIDs: extendedClientIDList,
		cachedUsers: cacheUsers, cachedUsersLock: lock}
	return &auth
}

////////////////////////////////////

// AdminAuth entity
type AdminAuth struct {
	app *core.Application

	appVerifier    *oidc.IDTokenVerifier
	appClientID    string
	webAppVerifier *oidc.IDTokenVerifier
	webAppClientID string

	authorization *casbin.Enforcer

	coreTokenAuth *tokenauth.TokenAuth

	cachedUsers     *syncmap.Map //cache users while active - 5 minutes timeout
	cachedUsersLock *sync.RWMutex
}

func (auth *AdminAuth) start() {

}

func (auth *AdminAuth) check(clientID string, r *http.Request) (*model.User, bool) {
	var data *userData
	var isCoreUser = false

	if auth.coreTokenAuth != nil {
		claims, err := auth.coreTokenAuth.CheckRequestTokens(r)
		if err == nil && claims != nil && !claims.Anonymous {
			err = auth.coreTokenAuth.AuthorizeRequestPermissions(claims, r)
			if err != nil {
				return nil, true
			}

			permissions := strings.Split(claims.Permissions, ",")
			data = &userData{Sub: &claims.Subject, Email: &claims.Email, Name: &claims.Name,
				Permissions: permissions, UIuceduUIN: &claims.UID}
			isCoreUser = true
		}
	}

	if data == nil {
		//1. Get the token from the request
		rawIDToken, tokenType, err := auth.getIDToken(r)
		if err != nil {
			return nil, false
		}

		//3. Validate the token
		idToken, err := auth.verify(*rawIDToken, *tokenType)
		if err != nil {
			log.Printf("error validating token - %s\n", err)
			return nil, false
		}

		//4. Get the user data from the token
		if err := idToken.Claims(&data); err != nil {
			log.Printf("error getting user data from token - %s\n", err)
			return nil, false
		}

		//we must have UIuceduUIN
		if data.UIuceduUIN == nil {
			log.Printf("error - missing uiuceuin data in the token - %s\n", err)
			return nil, false
		}

		obj := r.URL.Path // the resource that is going to be accessed.
		act := r.Method   // the operation that the user performs on the resource.

		hasAccess := false
		for _, s := range data.Permissions {
			hasAccess = auth.authorization.Enforce(s, obj, act)
			if hasAccess {
				break
			}
		}

		if !hasAccess {
			log.Printf("Access control error - UIN: %s is trying to apply %s operation for %s\n", *data.UIuceduUIN, act, obj)
			return nil, true
		}
	}

	if data == nil {
		log.Println("nil user data")
		return nil, false
	}

	var name, externalID, email, userID string
	if data.Sub != nil {
		userID = *data.Sub
	}
	if data.Name != nil {
		name = *data.Name
	}
	if data.UIuceduUIN != nil {
		externalID = *data.UIuceduUIN
	}
	if data.Email != nil {
		email = *data.Email
	}
	if data.Email != nil {
		email = *data.Email
	}
	return &model.User{
		ID:          userID,
		ClientID:    clientID,
		ExternalID:  externalID,
		Email:       email,
		Name:        name,
		IsCoreUser:  isCoreUser,
		Permissions: data.Permissions,
	}, false
}

// gets the token from the request - as cookie or as Authorization header.
// returns the id token and its type - mobile or web. If the token is taken by the cookie it is web otherwise it is mobile
func (auth *AdminAuth) getIDToken(r *http.Request) (*string, *string, error) {
	var tokenType string

	//1. Check if there is a cookie
	cookie, err := r.Cookie("rwa-at-data")
	if err == nil && cookie != nil && len(cookie.Value) > 0 {
		//there is a cookie
		tokenType = "web"
		return &cookie.Value, &tokenType, nil
	}

	//2. Check if there is a token in the Authorization header
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		return nil, nil, errors.New("error getting Authorization header")
	}
	splitAuthorization := strings.Fields(authorizationHeader)
	if len(splitAuthorization) != 2 {
		return nil, nil, errors.New("error processing the Authorization header")
	}
	// expected - Bearer 1234
	if splitAuthorization[0] != "Bearer" {
		return nil, nil, errors.New("error processing the Authorization header")
	}
	rawIDToken := splitAuthorization[1]
	tokenType = "mobile"
	return &rawIDToken, &tokenType, nil
}

func (auth *AdminAuth) verify(rawIDToken string, tokenType string) (*oidc.IDToken, error) {
	switch tokenType {
	case "mobile":
		log.Println("AdminAuth -> mobile app client token")
		return auth.appVerifier.Verify(context.Background(), rawIDToken)
	case "web":
		log.Println("AdminAuth -> web app client token")
		return auth.webAppVerifier.Verify(context.Background(), rawIDToken)
	default:
		return nil, errors.New("AdminAuth -> there is an issue with the audience")
	}
}

func (auth *AdminAuth) responseBadRequest(w http.ResponseWriter) {
	log.Println("AdminAuth -> 400 - Bad Request")

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func (auth *AdminAuth) responseUnauthorized(token string, w http.ResponseWriter) {
	log.Printf("AdminAuth -> 401 - Unauthorized for token %s", token)

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func (auth *AdminAuth) responseForbbiden(info string, w http.ResponseWriter) {
	log.Printf("AdminAuth -> 403 - Forbidden - %s", info)

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Forbidden"))
}

func (auth *AdminAuth) responseInternalServerError(w http.ResponseWriter) {
	log.Println("AdminAuth -> 500 - Internal Server Error")

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func newAdminAuth(app *core.Application, oidcProvider string, appClientID string, webAppClientID string, coreTokenAuth *tokenauth.TokenAuth, authorization *casbin.Enforcer) *AdminAuth {
	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}

	appVerifier := provider.Verifier(&oidc.Config{ClientID: appClientID})
	webAppVerifier := provider.Verifier(&oidc.Config{ClientID: webAppClientID})

	cacheUsers := &syncmap.Map{}
	lock := &sync.RWMutex{}

	auth := AdminAuth{app: app, appVerifier: appVerifier, appClientID: appClientID,
		webAppVerifier: webAppVerifier, webAppClientID: webAppClientID, coreTokenAuth: coreTokenAuth,
		cachedUsers: cacheUsers, cachedUsersLock: lock, authorization: authorization}
	return &auth
}
