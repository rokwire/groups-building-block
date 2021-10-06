package web

import (
	"context"
	"errors"
	"fmt"
	"groups/core"
	"groups/core/model"
	"groups/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/syncmap"
	"gopkg.in/ericchiang/go-oidc.v2"

	"github.com/casbin/casbin"
	"github.com/rokmetro/auth-library/authorization"
	"github.com/rokmetro/auth-library/authservice"
	"github.com/rokmetro/auth-library/tokenauth"
)

//Auth handler
type Auth struct {
	apiKeysAuth  *APIKeysAuth
	idTokenAuth  *IDTokenAuth
	internalAuth *InternalAuth
	adminAuth    *AdminAuth

	supportedClients []string
}

//Start starts the auth module
func (auth *Auth) Start() error {
	auth.idTokenAuth.start()

	return nil
}

func (auth *Auth) clientIDCheck(w http.ResponseWriter, r *http.Request) (bool, string) {
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

	log.Println(fmt.Sprintf("400 - Bad Request"))
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
	return false, ""
}

func (auth *Auth) apiKeyCheck(w http.ResponseWriter, r *http.Request) (string, bool) {
	clientIDOK, clientID := auth.clientIDCheck(w, r)
	if !clientIDOK {
		return "", false
	}

	apiKey := auth.getAPIKey(r)
	authenticated := auth.apiKeysAuth.check(apiKey, r, w)

	return clientID, authenticated
}

func (auth *Auth) idTokenCheck(w http.ResponseWriter, r *http.Request) (string, *model.User) {
	clientIDOK, clientID := auth.clientIDCheck(w, r)
	if !clientIDOK {
		return "", nil
	}

	idToken := auth.getIDToken(r)
	user := auth.idTokenAuth.check(clientID, idToken, r, w)
	return clientID, user
}

func (auth *Auth) internalAuthCheck(w http.ResponseWriter, r *http.Request) (string, bool) {
	clientIDOK, clientID := auth.clientIDCheck(w, r)
	if !clientIDOK {
		return "", false
	}

	internalAuthKey := auth.getInternalAPIKey(r)
	authenticated := auth.internalAuth.check(internalAuthKey, w)

	return clientID, authenticated
}

func (auth *Auth) mixedCheck(w http.ResponseWriter, r *http.Request) (string, bool, *model.User) {
	//get client ID
	clientIDOK, clientID := auth.clientIDCheck(w, r)
	if !clientIDOK {
		return "", false, nil
	}

	//first check for id token
	idToken := auth.getIDToken(r)
	if idToken != nil && len(*idToken) > 0 {
		authenticated := false
		user := auth.idTokenAuth.check(clientID, idToken, r, w)
		if user != nil {
			authenticated = true
		}
		return clientID, authenticated, user
	}

	//check api key
	apiKey := auth.getAPIKey(r)
	if apiKey != nil && len(*apiKey) > 0 {
		authenticated := auth.apiKeysAuth.check(apiKey, r, w)
		return clientID, authenticated, nil
	}

	//neither id token nor api key - so bad request
	log.Println("400 - Bad Request")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
	return clientID, false, nil
}

func (auth *Auth) adminCheck(w http.ResponseWriter, r *http.Request) (string, *model.User) {
	clientIDOK, clientID := auth.clientIDCheck(w, r)
	if !clientIDOK {
		return "", nil
	}

	user := auth.adminAuth.check(clientID, w, r)
	return clientID, user
}

func (auth *Auth) getAPIKey(r *http.Request) *string {
	apiKey := r.Header.Get("ROKWIRE-API-KEY")
	if len(apiKey) == 0 {
		return nil
	}
	return &apiKey
}

func (auth *Auth) getInternalAPIKey(r *http.Request) *string {
	apiKey := r.Header.Get("ROKWIRE_GS_API_KEY")
	if len(apiKey) == 0 {
		return nil
	}
	return &apiKey
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

//NewAuth creates new auth handler
func NewAuth(app *core.Application, host string, appKeys []string, internalAPIKeys []string, oidcProvider string, oidcClientID string,
	oidcAdminClientID string, oidcAdminWebClientID string, coreBBHost string, adminAuthorization *casbin.Enforcer) *Auth {
	var tokenAuth *tokenauth.TokenAuth
	if coreBBHost != "" {
		serviceID := "groups"
		// Instantiate a remote ServiceRegLoader to load auth service registration record from auth service
		serviceLoader := authservice.NewRemoteServiceRegLoader(coreBBHost, nil)

		// Instantiate AuthService instance
		authService, err := authservice.NewAuthService(serviceID, host, serviceLoader)
		if err == nil {
			permissionAuth := authorization.NewCasbinAuthorization("driver/web/permissions_authorization_policy.csv")
			scopeAuth := authorization.NewCasbinScopeAuthorization("driver/web/scope_authorization_policy.csv", serviceID)

			// Instantiate TokenAuth instance to perform token validation
			tokenAuth, _ = tokenauth.NewTokenAuth(true, authService, permissionAuth, scopeAuth)
		}
	}

	apiKeysAuth := newAPIKeysAuth(appKeys, tokenAuth)
	idTokenAuth := newIDTokenAuth(app, oidcProvider, oidcClientID, tokenAuth)
	internalAuth := newInternalAuth(internalAPIKeys)
	adminAuth := newAdminAuth(app, oidcProvider, oidcAdminClientID, oidcAdminWebClientID, tokenAuth, adminAuthorization)

	supportedClients := []string{"edu.illinois.rokwire", "edu.illinois.covid"}

	auth := Auth{apiKeysAuth: apiKeysAuth, idTokenAuth: idTokenAuth, internalAuth: internalAuth, adminAuth: adminAuth, supportedClients: supportedClients}
	return &auth
}

/////////////////////////////////////

//APIKeysAuth entity
type APIKeysAuth struct {
	appKeys []string

	coreTokenAuth *tokenauth.TokenAuth
}

func (auth *APIKeysAuth) check(apiKey *string, r *http.Request, w http.ResponseWriter) bool {
	//check if there is api key in the header
	if apiKey == nil || len(*apiKey) == 0 {
		if auth.coreTokenAuth != nil {
			_, err := auth.coreTokenAuth.CheckRequestTokens(r)
			if err == nil {
				return true
			}

			log.Printf("401 - Invalid API key and token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return false
		}

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
		if element == *apiKey {
			exist = true
			break
		}
	}
	if !exist {
		//not exist, so return 401
		log.Printf("401 - Unauthorized for key %s\n", *apiKey)

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

//NewAPIKeysAuth creates new api keys auth
func newAPIKeysAuth(appKeys []string, coreTokenAuth *tokenauth.TokenAuth) *APIKeysAuth {
	auth := APIKeysAuth{appKeys, coreTokenAuth}
	return &auth
}

////////////////////////////////////

type userData struct {
	UIuceduUIN        *string   `json:"uiucedu_uin"`
	Sub               *string   `json:"sub"`
	Email             *string   `json:"email"`
	UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
}

type cacheUser struct {
	user      *model.User
	lastUsage time.Time
}

////////////////////////////////////

//InternalAuth entity
type InternalAuth struct {
	appKeys []string
}

func (auth *InternalAuth) check(internalKey *string, w http.ResponseWriter) bool {
	//check if there is internal key in the header
	if internalKey == nil || len(*internalKey) == 0 {
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
		if element == *internalKey {
			exist = true
			break
		}
	}
	if !exist {
		//not exist, so return 401
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", *internalKey))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

//newInternalAuth creates new internal auth
func newInternalAuth(internalAPIKeys []string) *InternalAuth {
	auth := InternalAuth{appKeys: internalAPIKeys}
	return &auth
}

///////////////////////////////////

//IDTokenAuth entity
type IDTokenAuth struct {
	app *core.Application

	idTokenVerifier *oidc.IDTokenVerifier
	coreTokenAuth   *tokenauth.TokenAuth

	cachedUsers     *syncmap.Map //cache users while active - 5 minutes timeout
	cachedUsersLock *sync.RWMutex
}

func (auth *IDTokenAuth) start() {
	go auth.cleanCacheUser()
}

//cleanChacheUser cleans all users from the cache with no activity > 5 minutes
func (auth *IDTokenAuth) cleanCacheUser() {
	log.Println("IDTokenAuth -> cleanCacheUser -> start")

	toRemove := []string{}

	//find all users to remove - more than 5 minutes period from their last usage
	now := time.Now().Unix()
	auth.cachedUsers.Range(func(key, value interface{}) bool {
		cacheUser, ok := value.(*cacheUser)
		if !ok {
			return false //break the iteration
		}
		identifier, ok := key.(string)
		if !ok {
			return false //break the iteration
		}

		difference := now - cacheUser.lastUsage.Unix()
		//5 minutes
		if difference > 300 {
			toRemove = append(toRemove, identifier)
		}

		// this will continue iterating
		return true
	})

	//remove the selected ones
	count := len(toRemove)
	if count > 0 {
		log.Printf("IDTokenAuth -> cleanCacheUser -> %d items to remove\n", count)

		for _, key := range toRemove {
			auth.deleteCacheUser(key)
		}
	} else {
		log.Println("IDTokenAuth -> cleanCacheUser -> nothing to remove")
	}

	nextLoad := time.Minute * 5
	log.Printf("IDTokenAuth -> cleanCacheUser() -> next exec after %s\n", nextLoad)
	timer := time.NewTimer(nextLoad)
	<-timer.C
	log.Println("IDTokenAuth -> cleanCacheUser() -> timer expired")

	auth.cleanCacheUser()
}

func (auth *IDTokenAuth) check(clientID string, token *string, r *http.Request, w http.ResponseWriter) *model.User {
	var data *userData
	isCoreUser := false

	if auth.coreTokenAuth != nil {
		claims, err := auth.coreTokenAuth.CheckRequestTokens(r)
		if err == nil && claims != nil && !claims.Anonymous {
			err = auth.coreTokenAuth.AuthorizeRequestScope(claims, r)
			if err != nil {
				log.Printf("Scope error: %v\n", err)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return nil
			}

			log.Printf("Authentication successful for user: %v", claims)
			permissions := strings.Split(claims.Permissions, ",")
			data = &userData{Sub: &claims.Subject, Email: &claims.Email, UIuceduIsMemberOf: &permissions}
			if claims.UID != "" && claims.AuthType == "illinois_oidc" {
				data.UIuceduUIN = &claims.UID
			}

			isCoreUser = true
		}
	}

	if data == nil {
		//1. Check if there is a token
		if token == nil || len(*token) == 0 {
			auth.responseBadRequest(w)
			return nil
		}
		rawIDToken := *token

		//2. Validate the token
		idToken, err := auth.idTokenVerifier.Verify(context.Background(), rawIDToken)
		if err != nil {
			log.Printf("error validating token - %s\n", err)

			auth.responseUnauthorized(rawIDToken, w)
			return nil
		}

		//3. Get the user data from the token
		if err := idToken.Claims(&data); err != nil {
			log.Printf("error getting user data from token - %s\n", err)

			auth.responseUnauthorized(rawIDToken, w)
			return nil
		}
		//we must have UIuceduUIN
		if data.UIuceduUIN == nil {
			log.Printf("missing uiuceuin data in the token - %s\n", err)

			auth.responseUnauthorized(rawIDToken, w)
			return nil
		}
	}

	if data == nil {
		log.Println("nil user data")
		auth.responseInternalServerError(w)
		return nil
	}

	//4. Get the user for the provided external id.
	user, err := auth.getUser(clientID, *data, isCoreUser)
	if err != nil {
		log.Printf("error getting an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return nil
	}
	if user == nil {
		log.Printf("for some reasons the user for external id - %s is nil\n", err)

		auth.responseInternalServerError(w)
		return nil
	}

	//5. Update the user if needed
	user, err = auth.updateUserIfNeeded(clientID, *user, *data)
	if err != nil {
		log.Printf("error updating an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return nil
	}

	//6. Return the user
	return user
}

func (auth *IDTokenAuth) updateUserIfNeeded(clientID string, current model.User, userData userData) (*model.User, error) {
	currentList := current.IsMemberOf
	newList := userData.UIuceduIsMemberOf

	isEqual := utils.EqualPointers(currentList, newList)
	if !isEqual {
		log.Println("updateUserIfNeeded -> need to update user")

		//1. remove it from the cache
		auth.deleteCacheUser(current.ExternalID + "_" + clientID)

		//2. update it
		current.IsMemberOf = userData.UIuceduIsMemberOf
		err := auth.app.UpdateUser(clientID, &current)
		if err != nil {
			return nil, err
		}
	}

	return &current, nil
}

//the identifier is externalID_clientID
func (auth *IDTokenAuth) getCachedUser(identifier string) *cacheUser {
	auth.cachedUsersLock.RLock()
	defer auth.cachedUsersLock.RUnlock()

	var cachedUser *cacheUser //to return

	item, _ := auth.cachedUsers.Load(identifier)
	if item != nil {
		cachedUser = item.(*cacheUser)
	}

	//keep the last get time
	if cachedUser != nil {
		cachedUser.lastUsage = time.Now()
		auth.cachedUsers.Store(identifier, cachedUser)
	}

	return cachedUser
}

//the identifier is externalID_clientID
func (auth *IDTokenAuth) cacheUser(identifier string, user *model.User) {
	auth.cachedUsersLock.RLock()

	cacheUser := &cacheUser{user: user, lastUsage: time.Now()}
	auth.cachedUsers.Store(identifier, cacheUser)

	auth.cachedUsersLock.RUnlock()
}

//the identifier is externalID_clientID
func (auth *IDTokenAuth) deleteCacheUser(identifier string) {
	auth.cachedUsersLock.RLock()

	auth.cachedUsers.Delete(identifier)

	auth.cachedUsersLock.RUnlock()
}

func (auth *IDTokenAuth) getUser(clientID string, userData userData, isCoreUser bool) (*model.User, error) {
	if userData.Sub == nil {
		return nil, errors.New("user sub cannot be nil")
	}

	var err error

	//1. First check if cached
	cachedUser := auth.getCachedUser(*userData.Sub + "_" + clientID)
	if cachedUser != nil {
		return cachedUser.user, nil
	}

	var user *model.User

	//2. Check if we have a such user by Core BB Account ID in the application
	if isCoreUser {
		user, err := auth.app.FindUser(clientID, userData.Sub, false)
		if err != nil {
			log.Printf("error finding user for id %s: %s\n", *userData.Sub, err.Error())
			return nil, err
		}
		if user != nil {
			//cache it
			auth.cacheUser(*userData.Sub+"_"+clientID, user)
			return user, nil
		}
	}

	//3. Check if we have a such user by external ID in the application
	user, err = auth.app.FindUser(clientID, userData.UIuceduUIN, true)
	if err != nil {
		log.Printf("error finding user for external id %v: %s\n", userData.UIuceduUIN, err.Error())
		return nil, err
	}
	if user != nil {
		if isCoreUser {
			// Refactor user to use Core BB Account ID
			refactoredUser, err := auth.app.RefactorUser(clientID, user, *userData.Sub)
			if err != nil {
				log.Printf("error refactoring user for id %s, external id %v: %s\n", *userData.Sub, userData.UIuceduUIN, err.Error())
			}
			if refactoredUser != nil {
				//cache it
				auth.cacheUser(*userData.Sub+"_"+clientID, user)
				return refactoredUser, nil
			}
		}

		//cache it
		auth.cacheUser(*userData.Sub+"_"+clientID, user)
		return user, nil
	}

	//4. This is the first call for the user, so we need to create it
	user, err = auth.app.CreateUser(clientID, *userData.Sub, userData.UIuceduUIN, userData.Email, userData.UIuceduIsMemberOf)
	if err != nil {
		log.Printf("error creating an user - %s\n", err.Error())
		return nil, err
	}
	//cache it
	auth.cacheUser(*userData.Sub+"_"+clientID, user)
	return user, nil
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

//newIDTokenAuth creates new id token auth
func newIDTokenAuth(app *core.Application, oidcProvider string, oidcClientID string, coreTokenAuth *tokenauth.TokenAuth) *IDTokenAuth {
	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}
	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oidcClientID})

	cacheUsers := &syncmap.Map{}
	lock := &sync.RWMutex{}

	auth := IDTokenAuth{app: app, idTokenVerifier: idTokenVerifier, coreTokenAuth: coreTokenAuth,
		cachedUsers: cacheUsers, cachedUsersLock: lock}
	return &auth
}

////////////////////////////////////

//AdminAuth entity
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

func (auth *AdminAuth) check(clientID string, w http.ResponseWriter, r *http.Request) *model.User {
	var data *userData
	isCoreUser := false

	if auth.coreTokenAuth != nil {
		claims, err := auth.coreTokenAuth.CheckRequestTokens(r)
		if err == nil && claims != nil && !claims.Anonymous {
			err = auth.coreTokenAuth.AuthorizeRequestPermissions(claims, r)
			if err != nil {
				log.Printf("Permission error: %v\n", err)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return nil
			}

			permissions := strings.Split(claims.Permissions, ",")
			data = &userData{Sub: &claims.Subject, Email: &claims.Email, UIuceduIsMemberOf: &permissions}
			if claims.UID != "" && claims.AuthType == "illinois_oidc" {
				data.UIuceduUIN = &claims.UID
			}

			isCoreUser = true
		}
	}

	if data == nil {
		//1. Get the token from the request
		rawIDToken, tokenType, err := auth.getIDToken(r)
		if err != nil {
			auth.responseBadRequest(w)
			return nil
		}

		//3. Validate the token
		idToken, err := auth.verify(*rawIDToken, *tokenType)
		if err != nil {
			log.Printf("error validating token - %s\n", err)

			auth.responseUnauthorized(*rawIDToken, w)
			return nil
		}

		//4. Get the user data from the token
		if err := idToken.Claims(&data); err != nil {
			log.Printf("error getting user data from token - %s\n", err)

			auth.responseUnauthorized(*rawIDToken, w)
			return nil
		}

		//we must have UIuceduUIN
		if data.UIuceduUIN == nil {
			log.Printf("error - missing uiuceuin data in the token - %s\n", err)

			auth.responseUnauthorized(*rawIDToken, w)
			return nil
		}

		obj := r.URL.Path // the resource that is going to be accessed.
		act := r.Method   // the operation that the user performs on the resource.

		hasAccess := false
		for _, s := range *data.UIuceduIsMemberOf {
			hasAccess := auth.authorization.Enforce(s, obj, act)
			if hasAccess {
				break
			}
		}

		if !hasAccess {
			log.Printf("Access control error - UIN: %s is trying to apply %s operation for %s\n", *data.UIuceduUIN, act, obj)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return nil
		}
	}

	if data == nil {
		log.Println("nil user data")
		auth.responseInternalServerError(w)
		return nil
	}

	//4. Get the user for the provided external id.
	user, err := auth.getUser(clientID, *data, isCoreUser)
	if err != nil {
		log.Printf("error getting an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return nil
	}
	if user == nil {
		log.Printf("for some reasons the user for external id - %s is nil\n", err)

		auth.responseInternalServerError(w)
		return nil
	}

	//5. Update the user if needed
	user, err = auth.updateUserIfNeeded(clientID, *user, *data)
	if err != nil {
		log.Printf("error updating an user for external id - %s\n", err)

		auth.responseInternalServerError(w)
		return nil
	}

	return user
}

//gets the token from the request - as cookie or as Authorization header.
//returns the id token and its type - mobile or web. If the token is taken by the cookie it is web otherwise it is mobile
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

func (auth *AdminAuth) getUser(clientID string, userData userData, isCoreUser bool) (*model.User, error) {
	if userData.Sub == nil {
		return nil, errors.New("user sub cannot be nil")
	}

	var err error
	var user *model.User

	//2. Check if we have a such user by Core BB Account ID in the application
	if isCoreUser {
		user, err = auth.app.FindUser(clientID, userData.Sub, false)
		if err != nil {
			log.Printf("error finding user for id %s: %s\n", *userData.Sub, err.Error())
		}
		if user != nil {
			return user, nil
		}
	}

	//3. Check if we have a such user by external ID in the application
	user, err = auth.app.FindUser(clientID, userData.UIuceduUIN, true)
	if err != nil {
		log.Printf("error finding user for external id %s: %s\n", *userData.UIuceduUIN, err.Error())
		return nil, err
	}
	if user != nil {
		if isCoreUser {
			// Refactor user to use Core BB Account ID
			refactoredUser, err := auth.app.RefactorUser(clientID, user, *userData.Sub)
			if err != nil {
				log.Printf("error refactoring user for id %s, external id %s: %s\n", *userData.Sub, *userData.UIuceduUIN, err.Error())
			}
			if refactoredUser != nil {
				return refactoredUser, nil
			}
		}
		return user, nil
	}

	//4. This is the first call for the user, so we need to create it
	user, err = auth.app.CreateUser(clientID, *userData.Sub, userData.UIuceduUIN, userData.Email, userData.UIuceduIsMemberOf)
	if err != nil {
		log.Printf("error creating an user - %s\n", err.Error())
		return nil, err
	}
	return user, nil
}

func (auth *AdminAuth) updateUserIfNeeded(clientID string, current model.User, userData userData) (*model.User, error) {
	currentList := current.IsMemberOf
	newList := userData.UIuceduIsMemberOf

	isEqual := utils.EqualPointers(currentList, newList)
	if !isEqual {
		log.Println("updateUserIfNeeded -> need to update user")

		//2. update it
		current.IsMemberOf = userData.UIuceduIsMemberOf
		err := auth.app.UpdateUser(clientID, &current)
		if err != nil {
			return nil, err
		}
	}

	return &current, nil
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
