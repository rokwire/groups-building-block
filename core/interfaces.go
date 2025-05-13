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

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	// TODO: Deprecate this method due to missed CurrentMember!
	GetGroupEntity(clientID string, id string) (*model.Group, error)
	GetGroupEntityByTitle(clientID string, title string) (*model.Group, error)
	IsGroupAdmin(clientID string, groupID string, userID string) (bool, error)

	CreateGroup(clientID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError)
	CreateGroupV3(clientID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError)
	UpdateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupDateUpdated(clientID string, groupID string) error
	DeleteGroup(clientID string, current *model.User, id string) error
	GetAllGroupsUnsecured() ([]model.Group, error)
	GetAllGroups(clientID string) ([]model.Group, error)
	GetGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	GetUserGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	DeleteUser(clientID string, current *model.User) error
	ReportGroupAsAbuse(clientID string, current *model.User, group *model.Group, comment string) error

	GetGroup(clientID string, current *model.User, id string) (*model.Group, error)
	GetGroupsByGroupIDs(groupIDs []string) ([]model.Group, error)

	GetGroupStats(clientID string, id string) (*model.GroupStats, error)

	ApplyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error
	UpdateMembership(clientID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error
	CreateMembershipsStatuses(clientID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error
	UpdateMemberships(clientID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error

	GetEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error
	GetEventUserIDs(eventID string) ([]string, error)
	GetGroupMembershipsStatusAndGroupTitle(userID string) ([]model.GetGroupMembershipsResponse, error)
	GetGroupMembershipsByGroupID(groupID string) ([]string, error)

	GetGroupsEvents(eventIDs []string) ([]model.GetGroupsEvents, error)
	GetUserData(userID string) (*model.UserDataResponse, error)

	GetPosts(clientID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error)
	GetPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	GetUserPostCount(clientID string, userID string) (*int64, error)
	CreatePost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error)
	UpdatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error)
	ReactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error
	ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error
	DeletePost(clientID string, current *model.User, groupID string, postID string, force bool) error

	SynchronizeAuthman(clientID string) error
	SynchronizeAuthmanGroup(clientID string, groupID string) error

	GetManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error)
	CreateManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error)
	UpdateManagedGroupConfig(config model.ManagedGroupConfig) error
	DeleteManagedGroupConfig(id string, clientID string) error

	GetSyncConfig(clientID string) (*model.SyncConfig, error)
	UpdateSyncConfig(config model.SyncConfig) error

	// V3
	CheckUserGroupMembershipPermission(clientID string, current *model.User, groupID string) (*model.Group, bool)
	FindGroupsV3(clientID string, filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error)
	CreateMembership(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error
	CreatePendingMembership(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error
	DeleteMembership(clientID string, current *model.User, groupID string) error
	DeleteMembershipByID(clientID string, current *model.User, membershipID string) error
	DeletePendingMembership(clientID string, current *model.User, groupID string) error

	// Group Notifications
	SendGroupNotification(clientID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error
	GetResearchProfileUserCount(clientID string, current *model.User, researchProfile map[string]map[string][]string) (int64, error)

	// Group Events
	FindAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error)
	UpdateGroupMappingsForEvent(clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error)
	AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error)

	// Calendar BB
	CreateCalendarEventForGroups(clientID string, adminIdentifier []model.AccountIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error)
	CreateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error)
	UpdateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error)
	GetGroupCalendarEvents(clientID string, current *model.User, groupID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error)
}

// Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	AdminDeleteMembershipsByID(clientID string, current *model.User, groupID string, accountIDs []string) error
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
	FindSyncConfig(context storage.TransactionContext, clientID string) (*model.SyncConfig, error)
	SaveSyncConfig(context storage.TransactionContext, config model.SyncConfig) error

	FindSyncTimes(context storage.TransactionContext, clientID string, key string, legacy bool) (*model.SyncTimes, error)
	SaveSyncTimes(context storage.TransactionContext, times model.SyncTimes) error

	GetUserPostCount(clientID string, userID string) (*int64, error)
	DeleteUser(clientID string, userID string) error

	CreateGroup(context storage.TransactionContext, clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) (*string, *utils.GroupError)
	UpdateGroup(context storage.TransactionContext, clientID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupWithMembership(context storage.TransactionContext, clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) *utils.GroupError
	UpdateGroupSyncTimes(context storage.TransactionContext, clientID string, group *model.Group) error
	UpdateGroupStats(context storage.TransactionContext, clientID string, id string, resetUpdateDate bool, resetMembershipUpdateDate bool, resetManagedMembershipUpdateDate bool, resetStats bool) error
	UpdateGroupDateUpdated(clientID string, groupID string) error
	DeleteGroup(ctx storage.TransactionContext, clientID string, id string) error
	FindGroup(context storage.TransactionContext, clientID string, groupID string, userID *string) (*model.Group, error)
	FindGroupByTitle(clientID string, title string) (*model.Group, error)
	FindGroups(clientID string, userID *string, filter model.GroupsFilter, skipMembershipCheck bool) ([]model.Group, error)
	FindAllGroupsUnsecured() ([]model.Group, error)
	FindGroupsByGroupIDs(groupIDs []string) ([]model.Group, error)
	FindUserGroups(clientID string, userID string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroupsCount(clientID string, userID string) (*int64, error)
	DeleteUsersByAccountsIDs(log *logs.Logger, context storage.TransactionContext, accountsIDs []string) error

	FindEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(context storage.TransactionContext, clientID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(clientID string, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, eventID string, groupID string) error
	PullMembersFromEventsByUserIDs(log *logs.Logger, context storage.TransactionContext, accountsIDs []string) error

	FindEventUserIDs(context storage.TransactionContext, eventID string) ([]string, error)
	GetEventByUserID(userID string) ([]model.Event, error)
	FindGroupMembershipStatusAndGroupTitle(context storage.TransactionContext, userID string) ([]model.GetGroupMembershipsResponse, error)
	FindGroupMembershipByGroupID(context storage.TransactionContext, groupID string) ([]string, error)
	GetGroupMembershipByUserID(userID string) ([]model.GroupMembership, error)

	FindGroupsEvents(context storage.TransactionContext, eventIDs []string) ([]model.GetGroupsEvents, error)

	ReportGroupAsAbuse(clientID string, userID string, group *model.Group) error

	FindAuthmanGroups(clientID string) ([]model.Group, error)
	FindAuthmanGroupByKey(clientID string, authmanGroupKey string) (*model.Group, error)

	LoadManagedGroupConfigs() ([]model.ManagedGroupConfig, error)
	FindManagedGroupConfig(id string, clientID string) (*model.ManagedGroupConfig, error)
	FindManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error)
	InsertManagedGroupConfig(config model.ManagedGroupConfig) error
	UpdateManagedGroupConfig(config model.ManagedGroupConfig) error
	DeleteManagedGroupConfig(id string, clientID string) error

	// V3
	FindGroupsV3(context storage.TransactionContext, clientID string, filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembershipsWithContext(context storage.TransactionContext, clientID string, filter model.MembershipFilter) (model.MembershipCollection, error)

	FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipWithContext(context storage.TransactionContext, clientID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error)
	FindUserGroupMembershipsWithContext(ctx storage.TransactionContext, clientID string, userID string) (model.MembershipCollection, error)
	BulkUpdateGroupMembershipsByExternalID(clientID string, groupID string, saveOperations []storage.SingleMembershipOperation, updateGroupStats bool) error
	SaveGroupMembershipByExternalID(clientID string, groupID string, externalID string, userID *string, status *string,
		email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string, updateGroupStats bool) (*model.GroupMembership, error)

	CreateMembership(clientID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	CreateMemberships(context storage.TransactionContext, clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) error
	CreatePendingMembership(clientID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) (*model.GroupMembership, error)
	UpdateMembership(clientID string, _ *model.User, membershipID string, membership *model.GroupMembership) error
	UpdateMemberships(clientID string, user *model.User, groupID string, operation model.MembershipMultiUpdate) error
	DeleteMembership(clientID string, groupID string, userID string) error
	DeleteMembershipByID(clientID string, current *model.User, membershipID string) error
	DeleteUnsyncedGroupMemberships(clientID string, groupID string, syncID string) (int64, error)
	DeleteGroupMembershipsByAccountsIDs(log *logs.Logger, context storage.TransactionContext, accountsIDs []string) error

	GetGroupMembershipStats(context storage.TransactionContext, clientID string, groupID string) (*model.GroupStats, error)

	// Group Events
	FindAdminGroupsForEvent(context storage.TransactionContext, clientID string, current *model.User, eventID string) ([]string, error)
	UpdateGroupMappingsForEvent(context storage.TransactionContext, clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error)
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
	CreateCalendarEvent(adminIdentifier []model.AccountIdentifiers, currentAccountIdentifier model.AccountIdentifiers, event map[string]interface{}, orgID string, appID string, groupIDs []string) (map[string]interface{}, error)
	UpdateCalendarEvent(currentAccountIdentifier model.AccountIdentifiers, eventID string, event map[string]interface{}, orgID string, appID string) (map[string]interface{}, error)
	GetGroupCalendarEvents(currentAccountIdentifier model.AccountIdentifiers, eventIDs []string, appID string, orgID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error)
	AddPeopleToCalendarEvent(people []string, eventID string, orgID string, appID string) error
	RemovePeopleFromCalendarEvent(people []string, eventID string, orgID string, appID string) error
}

// Social exposes Social BB APIs for the driver adapters
type Social interface {
	GetPosts(clientID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error)
	GetPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	GetUserPostCount(clientID string, userID string) (*int64, error)
	CreatePost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error)
	UpdatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error)
	ReactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error
	ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error
	DeletePost(clientID string, userID string, groupID string, postID string, force bool) error
}
