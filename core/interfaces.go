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

	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	LoginUser(current *model.User) error

	// TODO: Deprecate this method due to missed CurrentMember!
	GetGroupEntity(appID string, orgID string, id string) (*model.Group, error)
	GetGroupEntityByTitle(appID string, orgID string, title string) (*model.Group, error)
	IsGroupAdmin(appID string, orgID string, groupID string, userID string) (bool, error)

	CreateGroup(current *model.User, group *model.Group) (*string, *utils.GroupError)
	UpdateGroup(userID *string, group *model.Group) *utils.GroupError
	UpdateGroupDateUpdated(appID string, orgID string, groupID string) error
	DeleteGroup(appID string, orgID string, id string) error
	GetAllGroups(appID string, orgID string) ([]model.Group, error)
	GetGroups(userID *string, filter model.GroupsFilter) ([]model.Group, error)
	GetUserGroups(userID string, filter model.GroupsFilter) ([]model.Group, error)
	DeleteUser(current *model.User) error

	GetGroup(current *model.User, id string) (*model.Group, error)
	GetGroupStats(appID string, orgID string, id string) (*model.GroupStats, error)

	ApplyMembershipApproval(appID string, orgID string, membershipID string, approve bool, rejectReason string) error
	UpdateMembership(membership *model.GroupMembership, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error

	GetEvents(current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(appID string, orgID string, eventID string, groupID string) error

	GetPosts(current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
	GetPost(appID string, orgID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	GetUserPostCount(appID string, orgID string, userID string) (*int64, error)
	CreatePost(current *model.User, post *model.Post, group *model.Group) error
	UpdatePost(userID string, group *model.Group, post *model.Post) error
	ReactToPost(current *model.User, groupID string, postID string, reaction string) error
	ReportPostAsAbuse(current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error
	DeletePost(current *model.User, groupID string, postID string, force bool) error

	SynchronizeAuthman(appID string, orgID string) error
	SynchronizeAuthmanGroup(appID string, orgID string, groupID string) error

	// V3
	CheckUserGroupMembershipPermission(current *model.User, groupID string) (*model.Group, bool)
	FindGroupsV3(filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembership(appID string, orgID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(appID string, orgID string, id string) (*model.GroupMembership, error)
	FindUserGroupMemberships(appID string, orgID string, userID string) (model.MembershipCollection, error)
	CreateMembership(current *model.User, group *model.Group, membership *model.GroupMembership) error
	CreatePendingMembership(current *model.User, group *model.Group, membership *model.GroupMembership) error
	DeleteMembership(current *model.User, groupID string) error
	DeleteMembershipByID(current *model.User, membershipID string) error
	DeletePendingMembership(current *model.User, groupID string) error

	// Group Notifications
	SendGroupNotification(appID string, orgID string, notification model.GroupNotification) error
	GetResearchProfileUserCount(current *model.User, researchProfile map[string]map[string][]string) (int64, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

// TODO: Deprecate this method due to missed CurrentMember!
func (s *servicesImpl) GetGroupEntity(appID string, orgID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(appID, orgID, id)
}

func (s *servicesImpl) GetGroupEntityByTitle(appID string, orgID string, title string) (*model.Group, error) {
	return s.app.getGroupEntityByTitle(appID, orgID, title)
}

func (s *servicesImpl) IsGroupAdmin(appID string, orgID string, groupID string, userID string) (bool, error) {
	return s.app.isGroupAdmin(appID, orgID, groupID, userID)
}

func (s *servicesImpl) CreateGroup(current *model.User, group *model.Group) (*string, *utils.GroupError) {
	return s.app.createGroup(current, group)
}

func (s *servicesImpl) UpdateGroup(userID *string, group *model.Group) *utils.GroupError {
	return s.app.updateGroup(userID, group)
}

func (s *servicesImpl) UpdateGroupDateUpdated(appID string, orgID string, groupID string) error {
	return s.app.updateGroupDateUpdated(appID, orgID, groupID)
}

func (s *servicesImpl) DeleteGroup(appID string, orgID, id string) error {
	return s.app.deleteGroup(appID, orgID, id)
}

func (s *servicesImpl) GetGroups(userID *string, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getGroups(userID, filter)
}

func (s *servicesImpl) GetAllGroups(appID string, orgID string) ([]model.Group, error) {
	return s.app.getAllGroups(appID, orgID)
}

func (s *servicesImpl) GetUserGroups(userID string, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getUserGroups(userID, filter)
}

func (s *servicesImpl) LoginUser(current *model.User) error {
	return s.app.storage.LoginUser(current)
}

func (s *servicesImpl) DeleteUser(current *model.User) error {
	return s.app.deleteUser(current)
}

func (s *servicesImpl) GetGroup(current *model.User, id string) (*model.Group, error) {
	return s.app.getGroup(current, id)
}

func (s *servicesImpl) GetGroupStats(appID string, orgID string, id string) (*model.GroupStats, error) {
	return s.app.storage.GetGroupMembershipStats(nil, appID, orgID, id)
}

func (s *servicesImpl) ApplyMembershipApproval(appID string, orgID string, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(appID, orgID, membershipID, approve, rejectReason)
}

func (s *servicesImpl) UpdateMembership(membership *model.GroupMembership, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	return s.app.updateMembership(membership, status, dateAttended, notificationsPreferences)
}

func (s *servicesImpl) GetEvents(current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	return s.app.getEvents(current, groupID, filterByToMembers)
}

func (s *servicesImpl) CreateEvent(eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	return s.app.createEvent(eventID, group, toMemberList, creator)
}

func (s *servicesImpl) UpdateEvent(appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember) error {
	return s.app.updateEvent(appID, orgID, eventID, groupID, toMemberList)
}

func (s *servicesImpl) DeleteEvent(appID string, orgID string, eventID string, groupID string) error {
	return s.app.deleteEvent(appID, orgID, eventID, groupID)
}

func (s *servicesImpl) GetPosts(current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {
	return s.app.storage.FindPosts(current, groupID, filterPrivatePostsValue, filterByToMembers, offset, limit, order)
}

func (s *servicesImpl) GetPost(appID string, orgID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return s.app.storage.FindPost(nil, appID, orgID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (s *servicesImpl) GetUserPostCount(appID string, orgID string, userID string) (*int64, error) {
	return s.app.storage.GetUserPostCount(appID, orgID, userID)
}

func (s *servicesImpl) CreatePost(current *model.User, post *model.Post, group *model.Group) error {
	return s.app.createPost(current, post, group)
}

func (s *servicesImpl) UpdatePost(userID string, group *model.Group, post *model.Post) error {
	return s.app.updatePost(userID, group, post)
}

func (s *servicesImpl) ReactToPost(current *model.User, groupID string, postID string, reaction string) error {
	return s.app.reactToPost(current, groupID, postID, reaction)
}

func (s *servicesImpl) ReportPostAsAbuse(current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	return s.app.reportPostAsAbuse(current, group, post, comment, sendToDean, sendToGroupAdmins)
}

func (s *servicesImpl) DeletePost(current *model.User, groupID string, postID string, force bool) error {
	return s.app.deletePost(nil, current.AppID, current.OrgID, current.ID, groupID, postID, force)
}

func (s *servicesImpl) SynchronizeAuthman(appID string, orgID string) error {
	return s.app.synchronizeAuthman(appID, orgID, false)
}

func (s *servicesImpl) SynchronizeAuthmanGroup(appID string, orgID string, groupID string) error {
	return s.app.synchronizeAuthmanGroup(appID, orgID, groupID)
}

// V3

func (s *servicesImpl) CheckUserGroupMembershipPermission(current *model.User, groupID string) (*model.Group, bool) {
	return s.app.checkUserGroupMembershipPermission(current, groupID)
}

func (s *servicesImpl) FindGroupsV3(filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.findGroupsV3(filter)
}

func (s *servicesImpl) FindGroupMemberships(filter model.MembershipFilter) (model.MembershipCollection, error) {
	return s.app.findGroupMemberships(filter)
}

func (s *servicesImpl) FindGroupMembership(appID string, orgID string, groupID string, userID string) (*model.GroupMembership, error) {
	return s.app.storage.FindGroupMembership(nil, appID, orgID, groupID, userID)
}

func (s *servicesImpl) FindGroupMembershipByID(appID string, orgID string, id string) (*model.GroupMembership, error) {
	return s.app.storage.FindGroupMembershipByID(appID, orgID, id)
}

func (s *servicesImpl) FindUserGroupMemberships(appID string, orgID string, userID string) (model.MembershipCollection, error) {
	return s.app.storage.FindGroupMemberships(nil, model.MembershipFilter{AppID: appID, OrgID: orgID, UserID: &userID})
}

func (s *servicesImpl) CreateMembership(current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createMembership(current, group, membership)
}

func (s *servicesImpl) CreatePendingMembership(current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createPendingMembership(current, group, membership)
}

func (s *servicesImpl) DeletePendingMembership(current *model.User, groupID string) error {
	return s.app.deletePendingMembership(current, groupID)
}

func (s *servicesImpl) DeleteMembershipByID(current *model.User, membershipID string) error {
	return s.app.deleteMembershipByID(current, membershipID)
}

func (s *servicesImpl) DeleteMembership(current *model.User, groupID string) error {
	return s.app.deleteMembership(current, groupID)
}

func (s *servicesImpl) SendGroupNotification(appID string, orgID string, notification model.GroupNotification) error {
	return s.app.sendGroupNotification(appID, orgID, notification)
}

func (s *servicesImpl) GetResearchProfileUserCount(current *model.User, researchProfile map[string]map[string][]string) (int64, error) {
	return s.app.getResearchProfileUserCount(current, researchProfile)
}

// Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetGroups(filter model.GroupsFilter) ([]model.Group, error)

	GetConfig(id string, claims *tokenauth.Claims) (*model.Config, error)
	GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Config, error)
	CreateConfig(config model.Config, claims *tokenauth.Claims) (*model.Config, error)
	UpdateConfig(config model.Config, claims *tokenauth.Claims) error
	DeleteConfig(id string, claims *tokenauth.Claims) error
}

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetGroups(filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getGroupsUnprotected(filter)
}

func (s *administrationImpl) GetConfig(id string, claims *tokenauth.Claims) (*model.Config, error) {
	return s.app.getConfig(id, claims)
}

func (s *administrationImpl) GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Config, error) {
	return s.app.getConfigs(configType, claims)
}

func (s *administrationImpl) CreateConfig(config model.Config, claims *tokenauth.Claims) (*model.Config, error) {
	return s.app.createConfig(config, claims)
}

func (s *administrationImpl) UpdateConfig(config model.Config, claims *tokenauth.Claims) error {
	return s.app.updateConfig(config, claims)
}

func (s *administrationImpl) DeleteConfig(id string, claims *tokenauth.Claims) error {
	return s.app.deleteConfig(id, claims)
}

// Storage is used by corebb to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener storage.Listener)

	PerformTransaction(transaction func(context storage.TransactionContext) error) error

	FindConfig(configType string, appID string, orgID string) (*model.Config, error)
	FindConfigByID(id string) (*model.Config, error)
	FindConfigs(configType *string) ([]model.Config, error)
	InsertConfig(config model.Config) error
	UpdateConfig(config model.Config) error
	DeleteConfig(id string) error

	FindSyncTimes(context storage.TransactionContext, appID string, orgID string) (*model.SyncTimes, error)
	SaveSyncTimes(context storage.TransactionContext, times model.SyncTimes) error

	FindUser(appID string, orgID string, id string, external bool) (*model.User, error)
	FindUsers(appID string, orgID string, ids []string, external bool) ([]model.User, error)
	FindAllUserPosts(context storage.TransactionContext, appID string, orgID string, userID string) ([]model.Post, error)
	GetUserPostCount(appID string, orgID string, userID string) (*int64, error)
	LoginUser(current *model.User) error
	CreateUser(id string, appID string, orgID string, externalID string, email string, name string) (*model.User, error)
	DeleteUser(context storage.TransactionContext, appID string, orgID string, userID string) error

	CreateGroup(current *model.User, group *model.Group, memberships []model.GroupMembership) (*string, *utils.GroupError)
	UpdateGroup(userID *string, group *model.Group, memberships []model.GroupMembership) *utils.GroupError
	UpdateGroupSyncTimes(context storage.TransactionContext, group *model.Group) error
	UpdateGroupStats(context storage.TransactionContext, appID string, orgID string, id string, resetUpdateDate bool, resetMembershipUpdateDate bool, resetManagedMembershipUpdateDate bool, resetStats bool) error
	UpdateGroupDateUpdated(appID string, orgID string, groupID string) error
	DeleteGroup(appID string, orgID string, id string) error
	FindGroup(context storage.TransactionContext, appID string, orgID string, groupID string, userID *string) (*model.Group, error)
	FindGroupByTitle(appID string, orgID string, title string) (*model.Group, error)
	FindGroups(userID *string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroups(userID string, filter model.GroupsFilter) ([]model.Group, error)
	FindUserGroupsCount(appID string, orgID string, userID string) (*int64, error)

	FindEvents(current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error)
	UpdateEvent(context storage.TransactionContext, appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(context storage.TransactionContext, appID string, orgID string, eventID string, groupID string) error

	FindPosts(current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
	FindPost(context storage.TransactionContext, appID string, orgID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	FindPostsByParentID(context storage.TransactionContext, appID string, orgID string, userID string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]*model.Post, error)
	FindTopPostByParentID(current *model.User, groupID string, parentID string, skipMembershipCheck bool) (*model.Post, error)
	CreatePost(context storage.TransactionContext, post *model.Post) error
	UpdatePost(context storage.TransactionContext, userID string, post *model.Post) error
	ReportPostAsAbuse(post *model.Post) error
	ReactToPost(context storage.TransactionContext, userID string, postID string, reaction string, on bool) error
	DeletePost(ctx storage.TransactionContext, appID string, orgID string, userID string, groupID string, postID string, force bool) error

	FindAuthmanGroups(appID string, orgID string) ([]model.Group, error)
	FindAuthmanGroupByKey(appID string, orgID string, authmanGroupKey string) (*model.Group, error)

	// V3
	FindGroupsV3(filter model.GroupsFilter) ([]model.Group, error)
	FindGroupMemberships(context storage.TransactionContext, filter model.MembershipFilter) (model.MembershipCollection, error)
	FindGroupMembership(context storage.TransactionContext, appID string, orgID string, groupID string, userID string) (*model.GroupMembership, error)
	FindGroupMembershipByID(appID string, orgID string, id string) (*model.GroupMembership, error)
	BulkUpdateGroupMembershipsByExternalID(appID string, orgID string, groupID string, saveOperations []storage.SingleMembershipOperation, updateGroupStats bool) error
	SaveGroupMembershipByExternalID(context storage.TransactionContext, appID string, orgID string, groupID string, externalID string, userID *string, status *string,
		email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string) error

	CreateMembership(current *model.User, group *model.Group, member *model.GroupMembership) error
	CreatePendingMembership(current *model.User, group *model.Group, member *model.GroupMembership) error
	ApplyMembershipApproval(context storage.TransactionContext, appID string, orgID string, membershipID string, approve bool, rejectReason string) (*model.GroupMembership, error)
	UpdateMembership(context storage.TransactionContext, membership *model.GroupMembership) error
	DeleteMembership(context storage.TransactionContext, appID string, orgID string, groupID string, userID string) error
	DeleteMembershipByID(appID string, orgID string, membershipID string) error
	DeleteUnsyncedGroupMemberships(appID string, orgID string, groupID string, syncID string) (int64, error)

	GetGroupMembershipStats(context storage.TransactionContext, appID string, orgID string, groupID string) (*model.GroupStats, error)
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
	SendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, accountCriteria map[string]interface{}, appID string, orgID string)
	SendMail(toEmail string, subject string, body string)
}

// Authman exposes Authman APIs for the driver adapters
type Authman interface {
	RetrieveAuthmanGroupMembers(groupName string) ([]string, error)
	RetrieveAuthmanUsers(externalIDs []string) (map[string]model.AuthmanSubject, error)
	RetrieveAuthmanStemGroups(stemName string) (*model.AuthmanGroupsResponse, error)
	AddAuthmanMemberToGroup(groupName string, uin string) error
	RemoveAuthmanMemberFromGroup(groupName string, uin string) error
}

// Core exposes Core APIs for the driver adapters
type Core interface {
	RetrieveCoreUserAccount(token string) (*model.CoreAccount, error)
	RetrieveCoreServices(serviceIDs []string) ([]model.CoreService, error)
	GetAccountsCount(searchParams map[string]interface{}, appID *string, orgID *string) (int64, error)
}

// Rewards exposes Rewards internal APIs for giving rewards to the users
type Rewards interface {
	CreateUserReward(userID string, rewardType string, description string) error
}
