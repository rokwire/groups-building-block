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
func (s *servicesImpl) GetGroupEntity(orgID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(orgID, id)
}

func (s *servicesImpl) GetGroupEntityByTitle(orgID string, title string) (*model.Group, error) {
	return s.app.getGroupEntityByTitle(orgID, title)
}

func (s *servicesImpl) IsGroupAdmin(orgID string, groupID string, userID string) (bool, error) {
	return s.app.isGroupAdmin(orgID, groupID, userID)
}

func (s *servicesImpl) CreateGroup(orgID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError) {
	return s.app.createGroup(orgID, current, group, membersConfig)
}

func (s *servicesImpl) CreateGroupV3(orgID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError) {
	return s.app.createGroupV3(orgID, current, group, membershipStatuses)
}

func (s *servicesImpl) UpdateGroup(orgID string, current *model.User, group *model.Group) *utils.GroupError {
	return s.app.updateGroup(orgID, current, group)
}

func (s *servicesImpl) UpdateGroupDateUpdated(orgID string, groupID string) error {
	return s.app.updateGroupDateUpdated(orgID, groupID)
}

func (s *servicesImpl) DeleteGroup(orgID string, current *model.User, id string) error {
	return s.app.deleteGroup(orgID, current, id, false)
}

func (s *servicesImpl) GetGroups(orgID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getGroups(orgID, current, filter, false)
}

func (s *servicesImpl) GetAllGroupsUnsecured() ([]model.Group, error) {
	return s.app.getAllGroupsUnsecured()
}

func (s *servicesImpl) GetAllGroups(orgID string) ([]model.Group, error) {
	return s.app.getAllGroups(orgID)
}

func (s *servicesImpl) GetUserGroups(orgID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getUserGroups(orgID, current, filter)
}

func (s *servicesImpl) ReportGroupAsAbuse(orgID string, current *model.User, group *model.Group, comment string) error {
	return s.app.reportGroupAsAbuse(orgID, current, group, comment)
}

func (s *servicesImpl) DeleteUser(orgID string, current *model.User) error {
	return s.app.deleteUser(orgID, current)
}

func (s *servicesImpl) GetGroup(orgID string, current *model.User, id string) (*model.Group, error) {
	return s.app.getGroup(orgID, current, id)
}

func (s *servicesImpl) GetGroupFilterStats(orgID string, current *model.User, filter model.StatsFilter) (*model.StatsResult, error) {
	return s.app.getGroupFilterStats(orgID, current, filter)
}

func (s *servicesImpl) GetGroupStats(orgID string, id string) (*model.GroupStats, error) {
	return s.app.storage.GetGroupMembershipStats(nil, orgID, id)
}

func (s *servicesImpl) ApplyMembershipApproval(orgID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(orgID, current, membershipID, approve, rejectReason)
}

func (s *servicesImpl) UpdateMembership(orgID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	return s.app.updateMembership(orgID, current, membershipID, status, dateAttended, notificationsPreferences)
}

func (s *servicesImpl) CreateMembershipsStatuses(orgID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error {
	return s.app.createMembershipsStatuses(orgID, current, groupID, membershipStatuses)
}

func (s *servicesImpl) UpdateMemberships(orgID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error {
	return s.app.updateMemberships(orgID, user, group, operation)
}

func (s *servicesImpl) GetEvents(orgID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	return s.app.getEvents(orgID, current, groupID, filterByToMembers)
}

func (s *servicesImpl) CreateEvent(orgID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	return s.app.createEvent(orgID, current, eventID, group, toMemberList, creator)
}

func (s *servicesImpl) UpdateEvent(orgID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error {
	return s.app.updateEvent(orgID, current, eventID, groupID, toMemberList)
}

func (s *servicesImpl) DeleteEvent(orgID string, current *model.User, eventID string, groupID string) error {
	return s.app.deleteEvent(orgID, current, eventID, groupID)
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

func (s *servicesImpl) GetPosts(orgID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error) {
	return s.app.getPosts(orgID, current, filter, filterPrivatePostsValue, filterByToMembers)
}

func (s *servicesImpl) GetPost(orgID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return s.app.getPost(orgID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (s *servicesImpl) GetUserPostCount(orgID string, userID string) (*int64, error) {
	return s.app.getUserPostCount(orgID, userID)
}

func (s *servicesImpl) CreatePost(orgID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	return s.app.createPost(orgID, current, post, group)
}

func (s *servicesImpl) UpdatePost(orgID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error) {
	return s.app.updatePost(orgID, current, group, post)
}

func (s *servicesImpl) ReactToPost(orgID string, current *model.User, groupID string, postID string, reaction string) error {
	return s.app.reactToPost(orgID, current, groupID, postID, reaction)
}

func (s *servicesImpl) ReportPostAsAbuse(orgID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	return s.app.reportPostAsAbuse(orgID, current, group, post, comment, sendToDean, sendToGroupAdmins)
}

func (s *servicesImpl) DeletePost(orgID string, current *model.User, groupID string, postID string, force bool) error {
	return s.app.deletePost(orgID, current.ID, groupID, postID, force)
}

func (s *servicesImpl) SynchronizeAuthman(orgID string) error {
	return s.app.synchronizeAuthman(orgID, false)
}

func (s *servicesImpl) SynchronizeAuthmanGroup(orgID string, groupID string) error {
	return s.app.synchronizeAuthmanGroup(orgID, groupID)
}

func (s *servicesImpl) GetManagedGroupConfigs(orgID string) ([]model.ManagedGroupConfig, error) {
	return s.app.getManagedGroupConfigs(orgID)
}

func (s *servicesImpl) CreateManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error) {
	return s.app.createManagedGroupConfig(config)
}

func (s *servicesImpl) UpdateManagedGroupConfig(config model.ManagedGroupConfig) error {
	return s.app.updateManagedGroupConfig(config)
}

func (s *servicesImpl) DeleteManagedGroupConfig(id string, orgID string) error {
	return s.app.deleteManagedGroupConfig(id, orgID)
}

func (s *servicesImpl) GetSyncConfig(orgID string) (*model.SyncConfig, error) {
	return s.app.getSyncConfig(orgID)
}

func (s *servicesImpl) UpdateSyncConfig(config model.SyncConfig) error {
	return s.app.updateSyncConfig(config)
}

// V3

func (s *servicesImpl) CheckUserGroupMembershipPermission(orgID string, current *model.User, groupID string) (*model.Group, bool) {
	return s.app.checkUserGroupMembershipPermission(orgID, current, groupID)
}

func (s *servicesImpl) FindGroupsV3(orgID string, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.findGroupsV3(orgID, filter)
}

func (s *servicesImpl) FindGroupMemberships(orgID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return s.app.findGroupMemberships(nil, orgID, filter)
}

func (s *servicesImpl) FindGroupMembership(orgID string, groupID string, userID string) (*model.GroupMembership, error) {
	return s.app.findGroupMembership(orgID, groupID, userID)
}

func (s *servicesImpl) FindGroupMembershipByID(orgID string, id string) (*model.GroupMembership, error) {
	return s.app.findGroupMembershipByID(orgID, id)
}

func (s *servicesImpl) FindUserGroupMemberships(orgID string, userID string) (model.MembershipCollection, error) {
	return s.app.findUserGroupMemberships(orgID, userID)
}

func (s *servicesImpl) CreateMembership(orgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createMembership(orgID, current, group, membership)
}

func (s *servicesImpl) CreatePendingMembership(orgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createPendingMembership(orgID, current, group, membership)
}

func (s *servicesImpl) DeletePendingMembership(orgID string, current *model.User, groupID string) error {
	return s.app.deletePendingMembership(orgID, current, groupID)
}

func (s *servicesImpl) DeleteMembershipByID(orgID string, current *model.User, membershipID string) error {
	return s.app.deleteMembershipByID(orgID, current, membershipID)
}

func (s *servicesImpl) DeleteMembership(orgID string, current *model.User, groupID string) error {
	return s.app.deleteMembership(orgID, current, groupID)
}

func (s *servicesImpl) SendGroupNotification(orgID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error {
	return s.app.sendGroupNotification(orgID, notification, predicate)
}

func (s *servicesImpl) GetResearchProfileUserCount(orgID string, current *model.User, researchProfile map[string]map[string][]string) (int64, error) {
	return s.app.getResearchProfileUserCount(orgID, current, researchProfile)
}

// Group Events

func (s *servicesImpl) FindAdminGroupsForEvent(orgID string, current *model.User, eventID string) ([]string, error) {
	return s.app.findAdminGroupsForEvent(orgID, current, eventID)
}

func (s *servicesImpl) UpdateGroupMappingsForEvent(orgID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	return s.app.updateGroupMappingsForEvent(orgID, current, eventID, groupIDs)
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

func (s *servicesImpl) CreateCalendarEventForGroups(orgID string, adminIdentifier []model.AccountIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	return s.app.createCalendarEventForGroups(orgID, adminIdentifier, current, event, groupIDs)
}

func (s *servicesImpl) CreateCalendarEventSingleGroup(orgID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	return s.app.createCalendarEventSingleGroup(orgID, current, event, groupID, members)
}

func (s *servicesImpl) UpdateCalendarEventSingleGroup(orgID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	return s.app.updateCalendarEventSingleGroup(orgID, current, event, groupID, members)
}

func (s *servicesImpl) GetGroupCalendarEvents(orgID string, current *model.User, groupID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error) {
	return s.app.getGroupCalendarEvents(orgID, current, groupID, published, filter)
}
