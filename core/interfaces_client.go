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
func (s *servicesImpl) GetGroupEntity(OrgID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(OrgID, id)
}

func (s *servicesImpl) GetGroupEntityByTitle(OrgID string, title string) (*model.Group, error) {
	return s.app.getGroupEntityByTitle(OrgID, title)
}

func (s *servicesImpl) IsGroupAdmin(OrgID string, groupID string, userID string) (bool, error) {
	return s.app.isGroupAdmin(OrgID, groupID, userID)
}

func (s *servicesImpl) CreateGroup(OrgID string, current *model.User, group *model.Group, membersConfig *model.DefaultMembershipConfig) (*string, *utils.GroupError) {
	return s.app.createGroup(OrgID, current, group, membersConfig)
}

func (s *servicesImpl) CreateGroupV3(OrgID string, current *model.User, group *model.Group, membershipStatuses model.MembershipStatuses) (*string, *utils.GroupError) {
	return s.app.createGroupV3(OrgID, current, group, membershipStatuses)
}

func (s *servicesImpl) UpdateGroup(OrgID string, current *model.User, group *model.Group) *utils.GroupError {
	return s.app.updateGroup(OrgID, current, group)
}

func (s *servicesImpl) UpdateGroupDateUpdated(OrgID string, groupID string) error {
	return s.app.updateGroupDateUpdated(OrgID, groupID)
}

func (s *servicesImpl) DeleteGroup(OrgID string, current *model.User, id string) error {
	return s.app.deleteGroup(OrgID, current, id, false)
}

func (s *servicesImpl) GetGroups(OrgID string, current *model.User, filter model.GroupsFilter) (int64, []model.Group, error) {
	return s.app.getGroups(OrgID, current, filter, false)
}

func (s *servicesImpl) GetAllGroupsUnsecured() ([]model.Group, error) {
	return s.app.getAllGroupsUnsecured()
}

func (s *servicesImpl) GetAllGroups(OrgID string) (int64, []model.Group, error) {
	return s.app.getAllGroups(OrgID)
}

func (s *servicesImpl) GetUserGroups(OrgID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.getUserGroups(OrgID, current, filter)
}

func (s *servicesImpl) ReportGroupAsAbuse(OrgID string, current *model.User, group *model.Group, comment string) error {
	return s.app.reportGroupAsAbuse(OrgID, current, group, comment)
}

func (s *servicesImpl) DeleteUser(OrgID string, current *model.User) error {
	return s.app.deleteUser(OrgID, current)
}

func (s *servicesImpl) GetGroup(OrgID string, current *model.User, id string) (*model.Group, error) {
	return s.app.getGroup(OrgID, current, id)
}

func (s *servicesImpl) GetGroupFilterStats(OrgID string, current *model.User, filter model.StatsFilter, skipMembershipCheck bool) (*model.StatsResult, error) {
	return s.app.getGroupFilterStats(OrgID, current, filter, skipMembershipCheck)
}

func (s *servicesImpl) GetGroupStats(OrgID string, id string) (*model.GroupStats, error) {
	return s.app.storage.GetGroupMembershipStats(nil, OrgID, id)
}

func (s *servicesImpl) ApplyMembershipApproval(OrgID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(OrgID, current, membershipID, approve, rejectReason)
}

func (s *servicesImpl) UpdateMembership(OrgID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	return s.app.updateMembership(OrgID, current, membershipID, status, dateAttended, notificationsPreferences)
}

func (s *servicesImpl) CreateMembershipsStatuses(OrgID string, current *model.User, groupID string, membershipStatuses model.MembershipStatuses) error {
	return s.app.createMembershipsStatuses(OrgID, current, groupID, membershipStatuses)
}

func (s *servicesImpl) UpdateMemberships(OrgID string, user *model.User, group *model.Group, operation model.MembershipMultiUpdate) error {
	return s.app.updateMemberships(OrgID, user, group, operation)
}

func (s *servicesImpl) GetGroupMembershipsStatusAndGroupTitle(userID string) ([]model.GetGroupMembershipsResponse, error) {
	return s.app.findGroupMembershipsStatusAndGroupsTitle(userID)
}
func (s *servicesImpl) GetGroupMembershipsByGroupID(groupID string) ([]string, error) {
	return s.app.findGroupMembershipsByGroupID(groupID)
}

func (s *servicesImpl) GetGroupsByGroupIDs(groupIDs []string) ([]model.Group, error) {
	return s.app.findGroupsByGroupIDs(groupIDs)
}

func (s *servicesImpl) GetUserData(userID string) (*model.UserDataResponse, error) {
	return s.app.getUserData(userID)
}

func (s *servicesImpl) SynchronizeAuthman(OrgID string) error {
	return s.app.synchronizeAuthman(OrgID, false)
}

func (s *servicesImpl) SynchronizeAuthmanGroup(OrgID string, groupID string) error {
	return s.app.synchronizeAuthmanGroup(OrgID, groupID)
}

func (s *servicesImpl) GetManagedGroupConfigs(OrgID string) ([]model.ManagedGroupConfig, error) {
	return s.app.getManagedGroupConfigs(OrgID)
}

func (s *servicesImpl) CreateManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error) {
	return s.app.createManagedGroupConfig(config)
}

func (s *servicesImpl) UpdateManagedGroupConfig(config model.ManagedGroupConfig) error {
	return s.app.updateManagedGroupConfig(config)
}

func (s *servicesImpl) DeleteManagedGroupConfig(id string, OrgID string) error {
	return s.app.deleteManagedGroupConfig(id, OrgID)
}

func (s *servicesImpl) GetSyncConfig(OrgID string) (*model.SyncConfig, error) {
	return s.app.getSyncConfig(OrgID)
}

func (s *servicesImpl) UpdateSyncConfig(config model.SyncConfig) error {
	return s.app.updateSyncConfig(config)
}

// V3

func (s *servicesImpl) CheckUserGroupMembershipPermission(OrgID string, current *model.User, groupID string) (*model.Group, bool) {
	return s.app.checkUserGroupMembershipPermission(OrgID, current, groupID)
}

func (s *servicesImpl) FindGroupsV3(OrgID string, filter model.GroupsFilter) ([]model.Group, error) {
	return s.app.findGroupsV3(OrgID, filter)
}

func (s *servicesImpl) FindGroupMemberships(OrgID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return s.app.findGroupMemberships(nil, OrgID, filter)
}

func (s *servicesImpl) FindGroupMembership(OrgID string, groupID string, userID string) (*model.GroupMembership, error) {
	return s.app.findGroupMembership(OrgID, groupID, userID)
}

func (s *servicesImpl) FindGroupMembershipByID(OrgID string, id string) (*model.GroupMembership, error) {
	return s.app.findGroupMembershipByID(OrgID, id)
}

func (s *servicesImpl) FindUserGroupMemberships(OrgID string, userID string) (model.MembershipCollection, error) {
	return s.app.findUserGroupMemberships(OrgID, userID)
}

func (s *servicesImpl) CreateMembership(OrgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createMembership(OrgID, current, group, membership)
}

func (s *servicesImpl) CreatePendingMembership(OrgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	return s.app.createPendingMembership(OrgID, current, group, membership)
}

func (s *servicesImpl) DeletePendingMembership(OrgID string, current *model.User, groupID string) error {
	return s.app.deletePendingMembership(OrgID, current, groupID)
}

func (s *servicesImpl) DeleteMembershipByID(OrgID string, current *model.User, membershipID string) error {
	return s.app.deleteMembershipByID(OrgID, current, membershipID)
}

func (s *servicesImpl) DeleteMembership(OrgID string, current *model.User, groupID string) error {
	return s.app.deleteMembership(OrgID, current, groupID)
}

func (s *servicesImpl) SendGroupNotification(OrgID string, notification model.GroupNotification, predicate model.MutePreferencePredicate) error {
	return s.app.sendGroupNotification(OrgID, notification, predicate)
}

func (s *servicesImpl) GetResearchProfileUserCount(OrgID string, current *model.User, researchProfile map[string]map[string]any) (int64, error) {
	return s.app.getResearchProfileUserCount(OrgID, current, researchProfile)
}

// Analytics

func (s *servicesImpl) AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error) {
	return s.app.analyticsFindGroups(startDate, endDate)
}

func (s *servicesImpl) AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error) {
	return s.app.analyticsFindMembers(groupID, startDate, endDate)
}
