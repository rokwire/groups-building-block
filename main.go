package main

import (
	core "groups/core"
	storage "groups/driven/storage"
	web "groups/driver/web"
	"log"
	"os"
	"strings"
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

	//mongoDB adapter
	mongoDBAuth := getEnvKey("GR_MONGO_AUTH", true)
	mongoDBName := getEnvKey("GR_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("GR_MONGO_TIMEOUT", false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	//application
	application := core.NewApplication(Version, Build, storageAdapter)
	application.Start()

	//web adapter
	apiKeys := getAPIKeys()
	internalAPIKeys := getInternalAPIKeys()
	host := getEnvKey("GR_HOST", true)
	oidcProvider := getEnvKey("GR_OIDC_PROVIDER", true)
	oidcClientID := getEnvKey("GR_OIDC_CLIENT_ID", true)
	oidcAdminClientID := getEnvKey("GR_OIDC_ADMIN_CLIENT_ID", true)
	oidcAdminWebClientID := getEnvKey("GR_OIDC_ADMIN_WEB_CLIENT_ID", true)
	webAdapter := web.NewWebAdapter(application, host, apiKeys, oidcProvider, oidcClientID, oidcAdminClientID, oidcAdminWebClientID, internalAPIKeys)
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
