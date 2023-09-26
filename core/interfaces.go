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
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	LoginUser(clientID string, currentGetUserGroups *model.User) error

	// TODO: Deprecate this method due to missed CurrentMember!
	GetGroupEntity(clientID string, id string) (*model.Group, error)
	GetGroupEntityByTitle(clientID string, title string) (*model.Group, error)
	IsGroupAdmin(clientID string, groupID string, userID string) (bool, error)

	CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *utils.GroupError)
	UpdateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupDateUpdated(clientID string, groupID string) error
	DeleteGroup(clientID string, current *model.User, id string) error
	GetAllGroups(clientID string) ([]model.Group, error)
	GetGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	GetUserGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error)
	DeleteUser(clientID string, current *model.User) error

	GetGroup(clientID string, current *model.User, id string) (*model.Group, error)
	GetGroupStats(clientID string, id string) (*model.GroupStats, error)

	ApplyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error
	UpdateMembership(clientID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error

	GetEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error

	GetPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
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
	SendGroupNotification(clientID string, notification model.GroupNotification) error
	GetResearchProfileUserCount(clientID string, current *model.User, researchProfile map[string]map[string][]string) (int64, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error)
	AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error)

	// Calendar BB
	CreateCalendarEventForGroups(clientID string, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error)
	CreateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, error)
	GetGroupCalendarEvents(clientID string, current *model.User, groupID string, filter model.GroupEventFilter) (map[string]interface{}, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

// TODO: Deprecate this method due to missed CurrentMember!
func (s *servicesImpl) GetGroupEntity(clientID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(clientID, id)
}

func (s *servicesImpl) GetGroupEntityByTitle(clientID string, title string) (*model.Group, error) {
	return s.app.getGroupEntityByTitle(clientID, title)
}

func (s *servicesImpl) IsGroupAdmin(clientID string, groupID string, userID string) (bool, error) {
	return s.app.isGroupAdmin(clientID, groupID, userID)
}

func (s *servicesImpl) CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *utils.GroupError) {
	return s.app.createGroup(clientID, current, group)
}

func (s *servicesImpl) UpdateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError {
	return s.app.updateGroup(clientID, current, group)
}

func (s *servicesImpl) UpdateGroupDateUpdated(clientID string, groupID string) error {
	return s.app.updateGroupDateUpdated(clientID, groupID)
}

func (s *servicesImpl) DeleteGroup(clientID string, current *model.User, id string) error {
	return s.app.deleteGroup(clientID, current, id)
}

func (s *servicesImpl) GetGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getGroups(clientID, current, filter)
}

func (s *servicesImpl) GetAllGroups(clientID string) ([]model.Group, error) {
	return s.app.getAllGroups(clientID)
}

func (s *servicesImpl) GetUserGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getUserGroups(clientID, current, filter)
}

func (s *servicesImpl) LoginUser(clientID string, current *model.User) error {
	return s.app.loginUser(clientID, current)
}

func (s *servicesImpl) DeleteUser(clientID string, current *model.User) error {
	return s.app.deleteUser(clientID, current)
}

func (s *servicesImpl) GetGroup(clientID string, current *model.User, id string) (*model.Group, error) {
	return s.app.getGroup(clientID, current, id)
}

func (s *servicesImpl) GetGroupStats(clientID string, id string) (*model.GroupStats, error) {
	return s.app.storage.GetGroupMembershipStats(nil, clientID, id)
}

func (s *servicesImpl) ApplyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(clientID, current, membershipID, approve, rejectReason)
}

func (s *servicesImpl) UpdateMembership(clientID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	return s.app.updateMembership(clientID, current, membershipID, status, dateAttended, notificationsPreferences)
}

func (s *servicesImpl) GetEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	return s.app.getEvents(clientID, current, groupID, filterByToMembers)
}

func (s *servicesImpl) CreateEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	return s.app.createEvent(clientID, current, eventID, group, toMemberList, creator)
}

func (s *servicesImpl) UpdateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error {
	return s.app.updateEvent(clientID, current, eventID, groupID, toMemberList)
}

func (s *servicesImpl) DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error {
	return s.app.deleteEvent(clientID, current, eventID, groupID)
}

func (s *servicesImpl) GetPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {
	return s.app.getPosts(clientID, current, groupID, filterPrivatePostsValue, filterByToMembers, offset, limit, order)
}

func (s *servicesImpl) GetPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return s.app.getPost(clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (s *servicesImpl) GetUserPostCount(clientID string, userID string) (*int64, error) {
	return s.app.getUserPostCount(clientID, userID)
}

func (s *servicesImpl) CreatePost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	return s.app.createPost(clientID, current, post, group)
}

func (s *servicesImpl) UpdatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error) {
	return s.app.updatePost(clientID, current, group, post)
}

func (s *servicesImpl) ReactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error {
	return s.app.reactToPost(clientID, current, groupID, postID, reaction)
}

func (s *servicesImpl) ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	return s.app.reportPostAsAbuse(clientID, current, group, post, comment, sendToDean, sendToGroupAdmins)
}

func (s *servicesImpl) DeletePost(clientID string, current *model.User, groupID string, postID string, force bool) error {
	return s.app.deletePost(clientID, current.ID, groupID, postID, force)
}

func (s *servicesImpl) SynchronizeAuthman(clientID string) error {
	return s.app.synchronizeAuthman(clientID, false)
}

func (s *servicesImpl) SynchronizeAuthmanGroup(clientID string, groupID string) error {
	return s.app.synchronizeAuthmanGroup(clientID, groupID)
}

func (s *servicesImpl) GetManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	return s.app.getManagedGroupConfigs(clientID)
}

func (s *servicesImpl) CreateManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error) {
	return s.app.createManagedGroupConfig(config)
}

func (s *servicesImpl) UpdateManagedGroupConfig(config model.ManagedGroupConfig) error {
	return s.app.updateManagedGroupConfig(config)
}

func (s *servicesImpl) DeleteManagedGroupConfig(id string, clientID string) error {
	return s.app.deleteManagedGroupConfig(id, clientID)
}

func (s *servicesImpl) GetSyncConfig(clientID string) (*model.SyncConfig, error) {
	return s.app.getSyncConfig(clientID)
}

func (s *servicesImpl) UpdateSyncConfig(config model.SyncConfig) error {
	return s.app.updateSyncConfig(config)
}

// V3

func (s *servicesImpl) CheckUserGroupMembershipPermission(clientID string, current *model.User, groupID string) (*model.Group, bool) {
	return s.app.checkUserGroupMembershipPermission(clientID, current, groupID)
}

func (s *servicesImpl) FindGroupsV3(clientID string, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.findGroupsV3(clientID, filter)
}

func (s *servicesImpl) FindGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return s.app.findGroupMemberships(clientID, filter)
}

func (s *servicesImpl) FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	return s.app.findGroupMembership(clientID, groupID, userID)
}

func (s *servicesImpl) FindGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error) {
	return s.app.findGroupMembershipByID(clientID, id)
}

func (s *servicesImpl) FindUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error) {
	return s.app.findUserGroupMemberships(clientID, userID)
}

func (s *servicesImpl) CreateMembership(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createMembership(clientID, current, group, membership)
}

func (s *servicesImpl) CreatePendingMembership(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createPendingMembership(clientID, current, group, membership)
}

func (s *servicesImpl) DeletePendingMembership(clientID string, current *model.User, groupID string) error {
	return s.app.deletePendingMembership(clientID, current, groupID)
}

func (s *servicesImpl) DeleteMembershipByID(clientID string, current *model.User, membershipID string) error {
	return s.app.deleteMembershipByID(clientID, current, membershipID)
}

func (s *servicesImpl) DeleteMembership(clientID string, current *model.User, groupID string) error {
	return s.app.deleteMembership(clientID, current, groupID)
}

func (s *servicesImpl) SendGroupNotification(clientID string, notification model.GroupNotification) error {
	return s.app.sendGroupNotification(clientID, notification)
}

func (s *servicesImpl) GetResearchProfileUserCount(clientID string, current *model.User, researchProfile map[string]map[string][]string) (int64, error) {
	return s.app.getResearchProfileUserCount(clientID, current, researchProfile)
}

// Analytics

func (s *servicesImpl) AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error) {
	return s.app.analyticsFindGroups(startDate, endDate)
}

func (s *servicesImpl) AnalyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error) {
	return s.app.analyticsFindPosts(groupID, startDate, endDate)
}

func (s *servicesImpl) AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error) {
	return s.app.analyticsFindMembers(groupID, startDate, endDate)
}

func (s *servicesImpl) CreateCalendarEventForGroups(clientID string, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	return s.app.createCalendarEventForGroups(clientID, current, event, groupIDs)
}

func (s *servicesImpl) CreateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, error) {
	return s.app.createCalendarEventSingleGroup(clientID, current, event, groupID, members)
}

func (s *servicesImpl) GetGroupCalendarEvents(clientID string, current *model.User, groupID string, filter model.GroupEventFilter) (map[string]interface{}, error) {
	return s.app.getGroupCalendarEvents(clientID, current, groupID, filter)
}

// Administration exposes administration APIs for the driver adapters
type Administration interface {
}

type administrationImpl struct {
	app *Application
}

// Storage is used by corebb to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener storage.Listener)

	PerformTransaction(transaction func(context storage.TransactionContext) error) error

	LoadSyncConfigs(context storage.TransactionContext) ([]model.SyncConfig, error)
	FindSyncConfigs() ([]model.SyncConfig, error)
	FindSyncConfig(clientID string) (*model.SyncConfig, error)
	SaveSyncConfig(context storage.TransactionContext, config model.SyncConfig) error

	FindSyncTimes(context storage.TransactionContext, clientID string) (*model.SyncTimes, error)
	SaveSyncTimes(context storage.TransactionContext, times model.SyncTimes) error

	FindUser(clientID string, id string, external bool) (*model.User, error)
	FindUsers(clientID string, ids []string, external bool) ([]model.User, error)
	GetUserPostCount(clientID string, userID string) (*int64, error)
	LoginUser(clientID string, current *model.User) error
	CreateUser(clientID string, id string, externalID string, email string, name string) (*model.User, error)
	DeleteUser(clientID string, userID string) error

	CreateGroup(clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) (*string, *utils.GroupError)
	UpdateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError
	UpdateGroupWithMembership(clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) *utils.GroupError
	UpdateGroupSyncTimes(context storage.TransactionContext, clientID string, group *model.Group) error
	UpdateGroupStats(context storage.TransactionContext, clientID string, id string, resetUpdateDate bool, resetMembershipUpdateDate bool, resetManagedMembershipUpdateDate bool, resetStats bool) error
	UpdateGroupDateUpdated(clientID string, groupID string) error
	DeleteGroup(clientID string, id string) error
	FindGroup(context storage.TransactionContext, clientID string, groupID string, userID *string) (*model.Group, error)
	FindGroupWithContext(context storage.TransactionContext, clientID string, groupID string, userID *string) (*model.Group, error)
	FindGroupByTitle(clientID string, title string) (*model.Group, error)
	FindGroups(clientID string, userID *string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroups(clientID string, userID string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroupsCount(clientID string, userID string) (*int64, error)

	FindEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(clientID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(clientID string, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, eventID string, groupID string) error

	FindPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
	FindPost(context storage.TransactionContext, clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	FindPostsByParentID(context storage.TransactionContext, clientID string, userID string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]*model.Post, error)
	CreatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error)
	UpdatePost(clientID string, userID string, post *model.Post) (*model.Post, error)
	ReportPostAsAbuse(clientID string, userID string, group *model.Group, post *model.Post) error
	ReactToPost(context storage.TransactionContext, userID string, postID string, reaction string, on bool) error
	DeletePost(ctx storage.TransactionContext, clientID string, userID string, groupID string, postID string, force bool) error

	FindAuthmanGroups(clientID string) ([]model.Group, error)
	FindAuthmanGroupByKey(clientID string, authmanGroupKey string) (*model.Group, error)

	LoadManagedGroupConfigs() ([]model.ManagedGroupConfig, error)
	FindManagedGroupConfig(id string, clientID string) (*model.ManagedGroupConfig, error)
	FindManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error)
	InsertManagedGroupConfig(config model.ManagedGroupConfig) error
	UpdateManagedGroupConfig(config model.ManagedGroupConfig) error
	DeleteManagedGroupConfig(id string, clientID string) error

	// V3
	FindGroupsV3(clientID string, filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembershipsWithContext(context storage.TransactionContext, clientID string, filter model.MembershipFilter) (model.MembershipCollection, error)

	FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error)
	BulkUpdateGroupMembershipsByExternalID(clientID string, groupID string, saveOperations []storage.SingleMembershipOperation, updateGroupStats bool) error
	SaveGroupMembershipByExternalID(clientID string, groupID string, externalID string, userID *string, status *string,
		email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string, updateGroupStats bool) (*model.GroupMembership, error)

	CreateMembership(clientID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	CreatePendingMembership(clientID string, current *model.User, group *model.Group, member *model.GroupMembership) error
	ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) error
	UpdateMembership(clientID string, _ *model.User, membershipID string, membership *model.GroupMembership) error
	DeleteMembership(clientID string, groupID string, userID string) error
	DeleteMembershipByID(clientID string, current *model.User, membershipID string) error
	DeleteUnsyncedGroupMemberships(clientID string, groupID string, syncID string) (int64, error)

	GetGroupMembershipStats(context storage.TransactionContext, clientID string, groupID string) (*model.GroupStats, error)

	// Analytics
	AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error)
	AnalyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error)
	AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error)
}

type storageListenerImpl struct {
	storage.DefaultListenerImpl
	app *Application
}

func (a *storageListenerImpl) OnConfigsChanged() {
	a.app.setupSyncManagedGroupTimer()
}

// Notifications exposes Notifications BB APIs for the driver adapters
type Notifications interface {
	SendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string)
	SendMail(toEmail string, subject string, body string)
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
	GetAccounts(searchParams map[string]interface{}, appID *string, orgID *string, limit int, offset int, allAccess bool, approvedKeys []string) ([]map[string]interface{}, error)
	GetAccountsCount(searchParams map[string]interface{}, appID *string, orgID *string) (int64, error)
}

// Rewards exposes Rewards internal APIs for giving rewards to the users
type Rewards interface {
	CreateUserReward(userID string, rewardType string, description string) error
}

// Calendar exposes Calendar BB APIs for the driver adapters
type Calendar interface {
	CreateCalendarEvent(currentAccountID string, event map[string]interface{}, orgID string, appID string) (map[string]interface{}, error)
	GetGroupCalendarEvents(currentAccountID string, eventIDs []string, appID string, orgID string, filter model.GroupEventFilter) (map[string]interface{}, error)
	AddPeopleToCalendarEvent(people []string, eventID string, orgID string, appID string) error
}
