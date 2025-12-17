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
	"groups/utils"
	"time"
)

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

func (s *servicesImpl) CreateGroup(clientID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError) {
	return s.app.createGroup(clientID, current, group, membersConfig)
}

func (s *servicesImpl) CreateGroupV3(clientID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError) {
	return s.app.createGroupV3(clientID, current, group, membershipStatuses)
}

func (s *servicesImpl) UpdateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError {
	return s.app.updateGroup(clientID, current, group)
}

func (s *servicesImpl) UpdateGroupDateUpdated(clientID string, groupID string) error {
	return s.app.updateGroupDateUpdated(clientID, groupID)
}

func (s *servicesImpl) DeleteGroup(clientID string, current *model.User, id string) error {
	return s.app.deleteGroup(clientID, current, id, false)
}

func (s *servicesImpl) GetGroups(clientID string, current *model.User, filter model.GroupsFilter) (int64, []model.Group, error) {
	return s.app.getGroups(clientID, current, filter, false)
}

func (s *servicesImpl) GetAllGroupsUnsecured() ([]model.Group, error) {
	return s.app.getAllGroupsUnsecured()
}

func (s *servicesImpl) GetAllGroups(clientID string) (int64, []model.Group, error) {
	return s.app.getAllGroups(clientID)
}

func (s *servicesImpl) GetUserGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getUserGroups(clientID, current, filter)
}

func (s *servicesImpl) ReportGroupAsAbuse(clientID string, current *model.User, group *model.Group, comment string) error {
	return s.app.reportGroupAsAbuse(clientID, current, group, comment)
}

func (s *servicesImpl) DeleteUser(clientID string, current *model.User) error {
	return s.app.deleteUser(clientID, current)
}

func (s *servicesImpl) GetGroup(clientID string, current *model.User, id string) (*model.Group, error) {
	return s.app.getGroup(clientID, current, id)
}

func (s *servicesImpl) GetGroupFilterStats(clientID string, current *model.User, filter model.StatsFilter, skipMembershipCheck bool) (*model.StatsResult, error) {
	return s.app.getGroupFilterStats(clientID, current, filter, skipMembershipCheck)
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

func (s *servicesImpl) CreateMembershipsStatuses(clientID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error {
	return s.app.createMembershipsStatuses(clientID, current, groupID, membershipStatuses)
}

func (s *servicesImpl) UpdateMemberships(clientID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error {
	return s.app.updateMemberships(clientID, user, group, operation)
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

func (s *servicesImpl) GetEventUserIDs(eventID string) ([]string, error) {
	return s.app.findEventUserIDs(eventID)
}

func (s *servicesImpl) GetGroupMembershipsStatusAndGroupTitle(userID string) ([]model.GetGroupMembershipsResponse, error) {
	return s.app.findGroupMembershipsStatusAndGroupsTitle(userID)
}
func (s *servicesImpl) GetGroupMembershipsByGroupID(groupID string) ([]string, error) {
	return s.app.findGroupMembershipsByGroupID(groupID)
}

func (s *servicesImpl) GetGroupsEvents(eventIDs []string) ([]model.GetGroupsEvents, error) {
	return s.app.findGroupsEvents(eventIDs)
}
func (s *servicesImpl) GetGroupsByGroupIDs(groupIDs []string) ([]model.Group, error) {
	return s.app.findGroupsByGroupIDs(groupIDs)
}

func (s *servicesImpl) GetUserData(userID string) (*model.UserDataResponse, error) {
	return s.app.getUserData(userID)
}

func (s *servicesImpl) GetPosts(clientID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error) {
	return s.app.getPosts(clientID, current, filter, filterPrivatePostsValue, filterByToMembers)
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
	return s.app.findGroupMemberships(nil, clientID, filter)
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

func (s *servicesImpl) SendGroupNotification(clientID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error {
	return s.app.sendGroupNotification(clientID, notification, predicate)
}

func (s *servicesImpl) GetResearchProfileUserCount(clientID string, current *model.User, researchProfile map[string]map[string]any) (int64, error) {
	return s.app.getResearchProfileUserCount(clientID, current, researchProfile)
}

// Group Events

func (s *servicesImpl) FindAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error) {
	return s.app.findAdminGroupsForEvent(clientID, current, eventID)
}

func (s *servicesImpl) UpdateGroupMappingsForEvent(clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	return s.app.updateGroupMappingsForEvent(clientID, current, eventID, groupIDs)
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

func (s *servicesImpl) CreateCalendarEventForGroups(clientID string, adminIdentifier []model.AccountIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	return s.app.createCalendarEventForGroups(clientID, adminIdentifier, current, event, groupIDs)
}

func (s *servicesImpl) CreateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	return s.app.createCalendarEventSingleGroup(clientID, current, event, groupID, members)
}

func (s *servicesImpl) UpdateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	return s.app.updateCalendarEventSingleGroup(clientID, current, event, groupID, members)
}

func (s *servicesImpl) GetGroupCalendarEvents(clientID string, current *model.User, groupID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error) {
	return s.app.getGroupCalendarEvents(clientID, current, groupID, published, filter)
}
