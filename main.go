package main

import (
	core "groups/core"
	"groups/driven/authman"
	"groups/driven/corebb"
	"groups/driven/notifications"
	"groups/driven/polls"
	"groups/driven/rewards"
	storage "groups/driven/storage"
	web "groups/driver/web"
	"log"
	"os"
	"strings"

	"github.com/rokwire/core-auth-library-go/authservice"
	"github.com/rokwire/logging-library-go/logs"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}
	// core bb host
	coreBBHost := getEnvKey("CORE_BB_HOST", false)

	intrernalAPIKey := getEnvKey("INTERNAL_API_KEY", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("GR_MONGO_AUTH", true)
	mongoDBName := getEnvKey("GR_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("GR_MONGO_TIMEOUT", false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// Notification adapter
	notificationsInternalAPIKey := getEnvKey("NOTIFICATIONS_INTERNAL_API_KEY", true)
	notificationsBaseURL := getEnvKey("NOTIFICATIONS_BASE_URL", true)
	notificationsAdapter := notifications.NewNotificationsAdapter(notificationsInternalAPIKey, notificationsBaseURL)

	authmanBaseURL := getEnvKey("AUTHMAN_BASE_URL", true)
	authmanUsername := getEnvKey("AUTHMAN_USERNAME", true)
	authmanPassword := getEnvKey("AUTHMAN_PASSWORD", true)

	// Authman adapter
	authmanAdapter := authman.NewAuthmanAdapter(authmanBaseURL, authmanUsername, authmanPassword)

	// Core adapter
	coreAdapter := corebb.NewCoreAdapter(coreBBHost)

	// Auth Service
	groupServiceURL := getEnvKey("GROUP_SERVICE_URL", false)
	remoteConfig := authservice.RemoteAuthDataLoaderConfig{
		AuthServicesHost: coreBBHost,
	}

	// Instantiate a remote ServiceRegLoader to load auth service registration record from auth service
	serviceLoader, err := authservice.NewRemoteAuthDataLoader(remoteConfig, []string{"rewards", "polls-v2"}, logs.NewLogger("groupsbb", &logs.LoggerOpts{}))
	if err != nil {
		log.Fatalf("error instancing auth data loader: %s", err)
	}
	// Instantiate AuthService instance
	authService, err := authservice.NewAuthService("groups", groupServiceURL, serviceLoader)
	if err != nil {
		log.Fatalf("error instancing auth service: %s", err)
	}

	// Rewards adapter
	rewardsServiceReg, err := authService.GetServiceReg("rewards")
	if err != nil {
		log.Fatalf("error finding rewards service reg: %s", err)
	}
	rewardsAdapter := rewards.NewRewardsAdapter(rewardsServiceReg.Host, intrernalAPIKey)

	// Rewards adapter
	pollsServiceReg, err := authService.GetServiceReg("polls-v2")
	if err != nil {
		log.Fatalf("error finding polls service reg: %s", err)
	}
	pollsAdapter := polls.NewPollsAdapter(pollsServiceReg.Host, intrernalAPIKey)

	//application
	application := core.NewApplication(Version, Build, storageAdapter, notificationsAdapter, authmanAdapter,
		coreAdapter, rewardsAdapter, pollsAdapter)
	application.Start()

	//web adapter
	apiKeys := getAPIKeys()
	internalAPIKeys := getInternalAPIKeys()
	host := getEnvKey("GR_HOST", true)
	oidcProvider := getEnvKey("GR_OIDC_PROVIDER", true)
	oidcClientID := getEnvKey("GR_OIDC_CLIENT_ID", true)
	oidcExtendedClientIDs := getEnvKey("GR_OIDC_EXTENDED_CLIENT_IDS", false)
	oidcAdminClientID := getEnvKey("GR_OIDC_ADMIN_CLIENT_ID", true)
	oidcAdminWebClientID := getEnvKey("GR_OIDC_ADMIN_WEB_CLIENT_ID", true)

	webAdapter := web.NewWebAdapter(application, host, apiKeys, oidcProvider,
		oidcClientID, oidcExtendedClientIDs, oidcAdminClientID, oidcAdminWebClientID,
		internalAPIKeys, authService, groupServiceURL)
	webAdapter.Start()
}

func getAPIKeys() []string {
	//get from the environment
	rokwireAPIKeys := getEnvKey("ROKWIRE_API_KEYS", true)

	//it is comma separated format
	rokwireAPIKeysList := strings.Split(rokwireAPIKeys, ",")
	if len(rokwireAPIKeysList) <= 0 {
		log.Fatal("For some reasons the apis keys list is empty")
	}

	return rokwireAPIKeysList
}
func getInternalAPIKeys() []string {
	//get from the environment
	internalAPIKeys := getEnvKey("GR_GS_API_KEYS", true)

	//it is comma separated format
	rokwireInternalAPIKeysList := strings.Split(internalAPIKeys, ",")
	if len(rokwireInternalAPIKeysList) <= 0 {
		log.Fatal("Keys list is empty")
	}

	return rokwireInternalAPIKeysList
}

func getEnvKey(key string, required bool) string {
	//get from the environment
	value, exist := os.LookupEnv(key)
	if !exist {
		if required {
			log.Fatal("No provided environment variable for " + key)
		} else {
			log.Printf("No provided environment variable for " + key)
		}
	}
	printEnvVar(key, value)
	return value
}

func printEnvVar(name string, value string) {
	if Version == "dev" {
		log.Printf("%s=%s", name, value)
	}
}
