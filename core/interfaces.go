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

package core

import (
	"groups/core/model"
	"groups/driven/notifications"
	"groups/driven/storage"
	"groups/utils"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	// TODO: Deprecate this method due to missed CurrentMember!
	GetGroupEntity(orgID string, id string) (*model.Group, error)
	GetGroupEntityByTitle(orgID string, title string) (*model.Group, error)
	IsGroupAdmin(orgID string, groupID string, userID string) (bool, error)

	CreateGroup(orgID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError)
	CreateGroupV3(orgID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError)
	UpdateGroup(orgID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupDateUpdated(orgID string, groupID string) error
	DeleteGroup(orgID string, current *model.User, id string) error
	GetAllGroupsUnsecured() ([]model.Group, error)
	GetAllGroups(orgID string) (int64, []model.Group, error)
	GetGroups(orgID string, current *model.User, filter model.GroupsFilter) (int64, []model.Group, error)
	GetGroupFilterStats(orgID string, current *model.User, filter model.StatsFilter, skipMembershipCheck bool) (*model.StatsResult, error)
	GetUserGroups(orgID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	DeleteUser(orgID string, current *model.User) error
	ReportGroupAsAbuse(orgID string, current *model.User, group *model.Group, comment string) error

	GetGroup(orgID string, current *model.User, id string) (*model.Group, error)
	GetGroupsByGroupIDs(groupIDs []string) ([]model.Group, error)

	GetGroupStats(orgID string, id string) (*model.GroupStats, error)

	ApplyMembershipApproval(orgID string, current *model.User, membershipID string, approve bool, rejectReason string) error
	UpdateMembership(orgID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error
	CreateMembershipsStatuses(orgID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error
	UpdateMemberships(orgID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error

	GetGroupMembershipsStatusAndGroupTitle(userID string) ([]model.GetGroupMembershipsResponse, error)
	GetGroupMembershipsByGroupID(groupID string) ([]string, error)

	GetUserData(userID string) (*model.UserDataResponse, error)

	SynchronizeAuthman(orgID string) error
	SynchronizeAuthmanGroup(orgID string, groupID string) error

	GetManagedGroupConfigs(orgID string) ([]model.ManagedGroupConfig, error)
	CreateManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error)
	UpdateManagedGroupConfig(config model.ManagedGroupConfig) error
	DeleteManagedGroupConfig(id string, orgID string) error

	GetSyncConfig(orgID string) (*model.SyncConfig, error)
	UpdateSyncConfig(config model.SyncConfig) error

	// V3
	CheckUserGroupMembershipPermission(orgID string, current *model.User, groupID string) (*model.Group, bool)
	FindGroupsV3(orgID string, filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(orgID string, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembership(orgID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(orgID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(orgID string, userID string) (model.MembershipCollection, error)
	CreateMembership(orgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error
	CreatePendingMembership(orgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error
	DeleteMembership(orgID string, current *model.User, groupID string) error
	DeleteMembershipByID(orgID string, current *model.User, membershipID string) error
	DeletePendingMembership(orgID string, current *model.User, groupID string) error

	// Group Notifications
	SendGroupNotification(orgID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error
	GetResearchProfileUserCount(orgID string, current *model.User, researchProfile map[string]map[string]any) (int64, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error)
}

// Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetGroups(orgID string, current *model.User, filter model.GroupsFilter) (int64, []model.Group, error)
	DeleteGroup(orgID string, current *model.User, id string, inactive bool) error
	AdminDeleteMembershipsByID(orgID string, current *model.User, groupID string, accountIDs []string) error
}

// BBS exposes BBS APIs for the driver adapters
type BBS interface {
	// External Callbacks
	OnUpdatedGroupExternalEntity(groupID string, operation model.ExternalOperation) error
}

// Storage is used by corebb to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener storage.Listener)

	PerformTransaction(transaction func(context storage.TransactionContext) error) error

	LoadSyncConfigs(context storage.TransactionContext) ([]model.SyncConfig, error)
	FindSyncConfigs(context storage.TransactionContext) ([]model.SyncConfig, error)
	FindSyncConfig(context storage.TransactionContext, orgID string) (*model.SyncConfig, error)
	SaveSyncConfig(context storage.TransactionContext, config model.SyncConfig) error

	FindSyncTimes(context storage.TransactionContext, orgID string, key string, legacy bool) (*model.SyncTimes, error)
	SaveSyncTimes(context storage.TransactionContext, times model.SyncTimes) error

	DeleteUser(orgID string, userID string) error

	CreateGroup(context storage.TransactionContext, orgID string, current *model.User, group *model.Group, memberships []model.GroupMembership) (*string, *utils.GroupError)
	UpdateGroup(context storage.TransactionContext, orgID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupWithMembership(context storage.TransactionContext, orgID string, current *model.User, group *model.Group, memberships []model.GroupMembership) *utils.GroupError
	UpdateGroupSyncTimes(context storage.TransactionContext, orgID string, group *model.Group) error
	UpdateGroupStats(context storage.TransactionContext, orgID string, id string, resetUpdateDate bool, resetMembershipUpdateDate bool, resetManagedMembershipUpdateDate bool, resetStats bool) error
	UpdateGroupDateUpdated(orgID string, groupID string) error
	DeleteGroup(ctx storage.TransactionContext, orgID string, id string) error
	FindGroup(context storage.TransactionContext, orgID string, groupID string, userID *string) (*model.Group, error)
	FindGroupByTitle(orgID string, title string) (*model.Group, error)
	FindGroups(orgID string, userID *string, filter model.GroupsFilter, skipMembershipCheck bool) (int64, []model.Group, error)
	FindAllGroupsUnsecured() ([]model.Group, error)
	FindGroupsByGroupIDs(groupIDs []string) ([]model.Group, error)
	FindUserGroups(orgID string, userID string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroupsCount(orgID string, userID string) (*int64, error)
	DeleteUsersByAccountsIDs(log *logs.Logger, context storage.TransactionContext, accountsIDs []string) error

	FindGroupMembershipStatusAndGroupTitle(context storage.TransactionContext, userID string) ([]model.GetGroupMembershipsResponse, error)
	FindGroupMembershipByGroupID(context storage.TransactionContext, groupID string) ([]string, error)
	GetGroupMembershipByUserID(userID string) ([]model.GroupMembership, error)

	ReportGroupAsAbuse(orgID string, userID string, group *model.Group) error

	FindAuthmanGroups(orgID string) ([]model.Group, error)
	FindAuthmanGroupByKey(orgID string, authmanGroupKey string) (*model.Group, error)

	LoadManagedGroupConfigs() ([]model.ManagedGroupConfig, error)
	FindManagedGroupConfig(id string, orgID string) (*model.ManagedGroupConfig, error)
	FindManagedGroupConfigs(orgID string) ([]model.ManagedGroupConfig, error)
	InsertManagedGroupConfig(config model.ManagedGroupConfig) error
	UpdateManagedGroupConfig(config model.ManagedGroupConfig) error
	DeleteManagedGroupConfig(id string, orgID string) error

	// V3
	CalculateGroupFilterStats(orgID string, current *model.User, filter model.StatsFilter, skipMembershipCheck bool) (*model.StatsResult, error)
	FindGroupsV3(context storage.TransactionContext, orgID string, filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(orgID string, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembershipsWithContext(context storage.TransactionContext, orgID string, filter model.MembershipFilter) (model.MembershipCollection, error)

	FindGroupMembership(orgID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipWithContext(context storage.TransactionContext, orgID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(orgID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(orgID string, userID string) (model.MembershipCollection, error)
	FindUserGroupMembershipsWithContext(ctx storage.TransactionContext, orgID string, userID string) (model.MembershipCollection, error)
	BulkUpdateGroupMembershipsByExternalID(orgID string, groupID string, saveOperations []storage.SingleMembershipOperation, updateGroupStats bool) error
	SaveGroupMembershipByExternalID(orgID string, groupID string, externalID string, userID *string, status *string,
		email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string, updateGroupStats bool) (*model.GroupMembership, error)

	CreateMembership(orgID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	CreateMemberships(context storage.TransactionContext, orgID string, current *model.User, group *model.Group, memberships []model.GroupMembership) error
	CreatePendingMembership(orgID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	ApplyMembershipApproval(orgID string, membershipID string, approve bool, rejectReason string) (*model.GroupMembership, error)
	UpdateMembership(orgID string, _ *model.User, membershipID string, membership *model.GroupMembership) error
	UpdateMemberships(orgID string, user *model.User, groupID string, operation model.MembershipMultiUpdate) error
	DeleteMembership(orgID string, groupID string, userID string) error
	DeleteMembershipByID(orgID string, current *model.User, membershipID string) error
	DeleteUnsyncedGroupMemberships(orgID string, groupID string, syncID string) (int64, error)
	DeleteGroupMembershipsByAccountsIDs(log *logs.Logger, context storage.TransactionContext, accountsIDs []string) error

	GetGroupMembershipStats(context storage.TransactionContext, orgID string, groupID string) (*model.GroupStats, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error)

	// Handle external callbacks
	OnUpdatedGroupExternalEntity(context storage.TransactionContext, groupID string, operation model.ExternalOperation) error
}

type storageListenerImpl struct {
	storage.DefaultListenerImpl
	app *Application
}

func (a *storageListenerImpl) OnConfigsChanged() {
	a.app.setupCronTimer()
}

// Notifications exposes Notifications BB APIs for the driver adapters
type Notifications interface {
	SendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string, dateScheduled *time.Time) error
	SendMail(toEmail string, subject string, body string) error
	DeleteNotifications(appID string, orgID string, ids string) error
	AddNotificationRecipients(appID string, orgID string, notificationID string, userIDs []string) error
	RemoveNotificationRecipients(appID string, orgID string, notificationID string, userIDs []string) error
}

// Authman exposes Authman APIs for the driver adapters
type Authman interface {
	RetrieveAuthmanGroupMembers(groupName string) ([]string, error)
	RetrieveAuthmanUsers(externalIDs []string) (map[string]model.AuthmanSubject, error)
	RetrieveAuthmanStemGroups(stemName string) (*model.АuthmanGroupsResponse, error)
	AddAuthmanMemberToGroup(groupName string, uin string) error
	RemoveAuthmanMemberFromGroup(groupName string, uin string) error
}

// Core exposes Core APIs for the driver adapters
type Core interface {
	RetrieveCoreUserAccount(token string) (*model.CoreAccount, error)
	RetrieveCoreServices(serviceIDs []string) ([]model.CoreService, error)
	GetAccounts(searchParams map[string]interface{}, appID *string, orgID *string, limit *int, offset *int) ([]model.CoreAccount, error)
	GetAccountsWithIDs(ids []string, appID *string, orgID *string, limit *int, offset *int) ([]model.CoreAccount, error)
	GetAllCoreAccountsWithNetIDs(netIDs []string, appID *string, orgID *string) ([]model.CoreAccount, error)
	GetAllCoreAccountsWithExternalIDs(externalIDs []string, appID *string, orgID *string) ([]model.CoreAccount, error)
	GetAccountsCount(searchParams map[string]interface{}, appID *string, orgID *string) (int64, error)
	LoadDeletedMemberships() ([]model.DeletedUserData, error)
	RetrieveFerpaAccounts(ids []string) ([]string, error)
}

// Rewards exposes Rewards internal APIs for giving rewards to the users
type Rewards interface {
	CreateUserReward(userID string, rewardType string, description string) error
}

// Calendar exposes Calendar BB APIs for the driver adapters
type Calendar interface {
}

// Social exposes Social BB APIs for the driver adapters
type Social interface {
}

// Polls exposes Polls BB APIs for the driver adapters
type Polls interface {
	DeleteGroupPolls(orgID, groupID string) error
}
