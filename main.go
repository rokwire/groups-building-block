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
	"groups/driven/calendar"
	"groups/driven/corebb"
	"groups/driven/notifications"
	"groups/driven/polls"
	"groups/driven/rewards"
	"groups/driven/socialbb"
	storage "groups/driven/storage"
	web "groups/driver/web"
	"log"
	"os"
	"strings"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/keys"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/sigauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
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

	serviceID := "gr"
	loggerOpts := logs.LoggerOpts{
		SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version"),
		SensitiveHeaders: []string{"Rokwire-Api-Key", "Rokwire_gs_api_key", "Internal-Api-Key"},
	}
	logger := logs.NewLogger(serviceID, &loggerOpts)

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

	// Auth Service
	groupServiceURL := getEnvKey("GROUP_SERVICE_URL", false)

	authService := auth.Service{
		ServiceID:   serviceID,
		ServiceHost: groupServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := auth.NewRemoteServiceRegLoader(&authService, []string{"rewards", "polls-v2"})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := auth.NewServiceRegManager(&authService, serviceRegLoader, false)
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}

	serviceAccountID := getEnvKey("GR_SERVICE_ACCOUNT_ID", false)
	privKeyRaw := getEnvKey("GR_PRIV_KEY", true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.RS256, privKeyRaw)
	if err != nil {
		log.Fatalf("Error parsing priv key: %v", err)
	}
	signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
	if err != nil {
		log.Fatalf("Error initializing signature auth: %v", err)
	}

	serviceAccountLoader, err := auth.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
	if err != nil {
		log.Fatalf("Error initializing remote service account loader: %v", err)
	}

	serviceAccountManager, err := auth.NewServiceAccountManager(&authService, serviceAccountLoader)
	if err != nil {
		log.Fatalf("Error initializing service account manager: %v", err)
	}

	// Notification adapter
	appID := getEnvKey("GROUPS_APP_ID", true)
	orgID := getEnvKey("GROUPS_ORG_ID", true)
	notificationsReportAbuseEmail := getEnvKey("NOTIFICATIONS_REPORT_ABUSE_EMAIL", true)
	notificationsBaseURL := getEnvKey("NOTIFICATIONS_BASE_URL", true)
	notificationsAdapter, err := notifications.NewNotificationsAdapter(notificationsBaseURL, serviceAccountManager)
	if err != nil {
		log.Fatalf("Error initializing notification adapter: %v", err)
	}

	// Calendar adapter
	calendarBaseURL := getEnvKey("CALENDAR_BASE_URL", true)
	calendarAdapter, err := calendar.NewCalendarAdapter(calendarBaseURL, serviceAccountManager)
	if err != nil {
		log.Fatalf("Error initializing notification adapter: %v", err)
	}

	authmanBaseURL := getEnvKey("AUTHMAN_BASE_URL", true)
	authmanUsername := getEnvKey("AUTHMAN_USERNAME", true)
	authmanPassword := getEnvKey("AUTHMAN_PASSWORD", true)
	authmanAdminUINList := getAuthmanAdminUINList()

	// Authman adapter
	authmanAdapter := authman.NewAuthmanAdapter(authmanBaseURL, authmanUsername, authmanPassword)

	// Core adapter
	coreAdapter := corebb.NewCoreAdapter(logger, coreBBHost, serviceAccountManager)

	// Rewards adapter
	rewardsServiceReg, err := serviceRegManager.GetServiceReg("rewards")
	if err != nil {
		log.Fatalf("error finding rewards service reg: %s", err)
	}
	rewardsAdapter := rewards.NewRewardsAdapter(rewardsServiceReg.Host, intrernalAPIKey)

	supportedOrgIDs := []string{"edu.illinois.rokwire", "edu.illinois.covid"}

	config := &model.ApplicationConfig{
		AuthmanAdminUINList:       authmanAdminUINList,
		ReportAbuseRecipientEmail: notificationsReportAbuseEmail,
		SupportedOrgIDs:           supportedOrgIDs,
		AppID:                     appID,
		OrgID:                     orgID,
	}

	socialBaseURL := getEnvKey("GR_SOCIAL_BASE_URL", true)
	social := socialbb.NewSocialAdapter(logger, socialBaseURL, serviceAccountManager)

	pollsServiceReg, err := serviceRegManager.GetServiceReg("polls-v2")
	if err != nil {
		log.Fatalf("error finding polls v2 service reg: %s", err)
	}
	pollsAdapter, err := polls.NewPollsAdapter(pollsServiceReg.Host, serviceAccountManager)
	if err != nil {
		log.Fatalf("error instancing polls v2 adaptor: %s", err)
	}

	//application
	application := core.NewApplication(Version, Build, storageAdapter, notificationsAdapter, authmanAdapter,
		coreAdapter, rewardsAdapter, calendarAdapter, social, pollsAdapter, serviceID, logger, config)
	application.Start()

	//web adapter
	apiKeys := getAPIKeys()
	host := getEnvKey("GR_HOST", true)
	port := getEnvKey("GR_PORT", true)
	if len(port) == 0 {
		port = "80"
	}
	oidcProvider := getEnvKey("GR_OIDC_PROVIDER", true)
	oidcOrgID := getEnvKey("GR_OIDC_CLIENT_ID", true)
	oidcExtendedOrgIDs := getEnvKey("GR_OIDC_EXTENDED_CLIENT_IDS", false)
	oidcAdminOrgID := getEnvKey("GR_OIDC_ADMIN_CLIENT_ID", true)
	oidcAdminWebOrgID := getEnvKey("GR_OIDC_ADMIN_WEB_CLIENT_ID", true)

	webAdapter := web.NewWebAdapter(application, host, port, supportedOrgIDs, apiKeys, oidcProvider,
		oidcOrgID, oidcExtendedOrgIDs, oidcAdminOrgID, oidcAdminWebOrgID,
		intrernalAPIKey, serviceRegManager, groupServiceURL, logger)
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
			log.Print("No provided environment variable for " + key)
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
