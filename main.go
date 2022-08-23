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

package main

import (
	core "groups/core"
	"groups/core/model"
	"groups/driven/authman"
	"groups/driven/corebb"
	"groups/driven/notifications"
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

	loggerOpts := logs.LoggerOpts{SuppressRequests: []logs.HttpRequestProperties{logs.NewAwsHealthCheckHttpRequestProperties("/groups/version")}}
	logger := logs.NewLogger("core", &loggerOpts)

	// core bb host
	coreBBHost := getEnvKey("CORE_BB_HOST", false)

	intrernalAPIKey := getEnvKey("INTERNAL_API_KEY", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("GR_MONGO_AUTH", true)
	mongoDBName := getEnvKey("GR_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("GR_MONGO_TIMEOUT", false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// Notification adapter
	notificationsReportAbuseEmail := getEnvKey("NOTIFICATIONS_REPORT_ABUSE_EMAIL", true)
	notificationsInternalAPIKey := getEnvKey("NOTIFICATIONS_INTERNAL_API_KEY", true)
	notificationsBaseURL := getEnvKey("NOTIFICATIONS_BASE_URL", true)
	notificationsAdapter := notifications.NewNotificationsAdapter(notificationsInternalAPIKey, notificationsBaseURL)

	authmanBaseURL := getEnvKey("AUTHMAN_BASE_URL", true)
	authmanUsername := getEnvKey("AUTHMAN_USERNAME", true)
	authmanPassword := getEnvKey("AUTHMAN_PASSWORD", true)
	authmanAdminUINList := getAuthmanAdminUINList()

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
	serviceLoader, err := authservice.NewRemoteAuthDataLoader(remoteConfig, []string{"rewards"}, logs.NewLogger("groupsbb", &logs.LoggerOpts{}))
	if err != nil {
		log.Fatalf("error instancing auth data loader: %s", err)
	}
	// Instantiate AuthService instance
	authService, err := authservice.NewTestAuthService("groups", groupServiceURL, serviceLoader)
	if err != nil {
		log.Fatalf("error instancing auth service: %s", err)
	}

	// Rewards adapter
	rewardsServiceReg, err := authService.GetServiceReg("rewards")
	if err != nil {
		log.Fatalf("error finding rewards service reg: %s", err)
	}
	rewardsAdapter := rewards.NewRewardsAdapter(rewardsServiceReg.Host, intrernalAPIKey)

	supportedClientIDs := []string{"edu.illinois.rokwire", "edu.illinois.covid"}

	config := &model.ApplicationConfig{
		AuthmanAdminUINList:       authmanAdminUINList,
		ReportAbuseRecipientEmail: notificationsReportAbuseEmail,
		SupportedClientIDs:        supportedClientIDs,
	}

	//application
	application := core.NewApplication(Version, Build, storageAdapter, notificationsAdapter, authmanAdapter,
		coreAdapter, rewardsAdapter, config)
	application.Start()

	//web adapter
	apiKeys := getAPIKeys()
	host := getEnvKey("GR_HOST", true)
	oidcProvider := getEnvKey("GR_OIDC_PROVIDER", true)
	oidcClientID := getEnvKey("GR_OIDC_CLIENT_ID", true)
	oidcExtendedClientIDs := getEnvKey("GR_OIDC_EXTENDED_CLIENT_IDS", false)
	oidcAdminClientID := getEnvKey("GR_OIDC_ADMIN_CLIENT_ID", true)
	oidcAdminWebClientID := getEnvKey("GR_OIDC_ADMIN_WEB_CLIENT_ID", true)

	webAdapter := web.NewWebAdapter(application, host, supportedClientIDs, apiKeys, oidcProvider,
		oidcClientID, oidcExtendedClientIDs, oidcAdminClientID, oidcAdminWebClientID,
		intrernalAPIKey, authService, groupServiceURL, logger)
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

func getAuthmanAdminUINList() []string {
	//get from the environment
	authmanAdminUINs := getEnvKey("AUTHMAN_ADMIN_UIN_LIST", true)
	if len(authmanAdminUINs) == 0 {
		return nil
	}

	//it is comma separated format
	authmanAdminUINList := strings.Split(authmanAdminUINs, ",")
	if len(authmanAdminUINList) <= 0 {
		log.Fatal("AUTHMAN_ADMIN_UIN_LIST list is empty")
	}

	return authmanAdminUINList
}

func printEnvVar(name string, value string) {
	if Version == "dev" {
		log.Printf("%s=%s", name, value)
	}
}
