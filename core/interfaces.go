package core

import (
	"groups/core/model"
	"groups/driven/notifications"
	"time"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	LoginUser(clientID string, currentGetUserGroups *model.User) error

	GetGroupCategories() ([]string, error)
	GetUserGroupMembershipsByID(id string) ([]*model.Group, error)
	GetUserGroupMembershipsByExternalID(externalID string) ([]*model.Group, *model.User, error)

	GetGroupEntity(clientID string, id string) (*model.Group, error)
	GetGroupEntityByMembership(clientID string, membershipID string) (*model.Group, error)

	CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *GroupError)
	UpdateGroup(clientID string, current *model.User, group *model.Group) *GroupError
	DeleteGroup(clientID string, current *model.User, id string) error
	GetAllGroups(clientID string) ([]model.Group, error)
	GetGroups(clientID string, current *model.User, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]map[string]interface{}, error)
	GetUserGroups(clientID string, currentGetUserGroups *model.User) ([]map[string]interface{}, error)
	DeleteUser(clientID string, current *model.User) error
	GetGroup(clientID string, current *model.User, id string) (map[string]interface{}, error)

	CreatePendingMember(clientID string, current *model.User, group *model.Group, member *model.Member) error
	DeletePendingMember(clientID string, current *model.User, groupID string) error
	CreateMember(clientID string, current *model.User, group *model.Group, member *model.Member) error
	DeleteMember(clientID string, current *model.User, groupID string) error

	ApplyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error
	DeleteMembership(clientID string, current *model.User, membershipID string) error
	UpdateMembership(clientID string, current *model.User, membershipID string, status string, dateAttended *time.Time) error

	GetEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember) (*model.Event, error)
	UpdateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error

	GetPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
	GetPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	GetUserPostCount(clientID string, userID string) (*int64, error)
	CreatePost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error)
	UpdatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error)
	ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post) error
	DeletePost(clientID string, current *model.User, groupID string, postID string, force bool) error

	SynchronizeAuthman(clientID string) error
	SynchronizeAuthmanGroup(clientID string, group *model.Group) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) GetGroupCategories() ([]string, error) {
	return s.app.getGroupCategories()
}

func (s *servicesImpl) GetUserGroupMembershipsByID(id string) ([]*model.Group, error) {
	memberships, _, err := s.app.getUserGroupMemberships(id, false)
	return memberships, err
}

func (s *servicesImpl) GetUserGroupMembershipsByExternalID(externalID string) ([]*model.Group, *model.User, error) {
	return s.app.getUserGroupMemberships(externalID, true)
}

func (s *servicesImpl) GetGroupEntity(clientID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(clientID, id)
}

func (s *servicesImpl) GetGroupEntityByMembership(clientID string, membershipID string) (*model.Group, error) {
	return s.app.getGroupEntityByMembership(clientID, membershipID)
}

func (s *servicesImpl) CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *GroupError) {
	return s.app.createGroup(clientID, current, group)
}

func (s *servicesImpl) UpdateGroup(clientID string, current *model.User, group *model.Group) *GroupError {
	return s.app.updateGroup(clientID, current, group)
}

func (s *servicesImpl) DeleteGroup(clientID string, current *model.User, id string) error {
	return s.app.deleteGroup(clientID, current, id)
}

func (s *servicesImpl) GetGroups(clientID string, current *model.User, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]map[string]interface{}, error) {
	return s.app.getGroups(clientID, current, category, privacy, title, offset, limit, order)
}

func (s *servicesImpl) GetAllGroups(clientID string) ([]model.Group, error) {
	return s.app.getAllGroups(clientID)
}

func (s *servicesImpl) GetUserGroups(clientID string, current *model.User) ([]map[string]interface{}, error) {
	return s.app.getUserGroups(clientID, current)
}

func (s *servicesImpl) LoginUser(clientID string, current *model.User) error {
	return s.app.loginUser(clientID, current)
}

func (s *servicesImpl) DeleteUser(clientID string, current *model.User) error {
	return s.app.deleteUser(clientID, current)
}

func (s *servicesImpl) GetGroup(clientID string, current *model.User, id string) (map[string]interface{}, error) {
	return s.app.getGroup(clientID, current, id)
}

func (s *servicesImpl) CreatePendingMember(clientID string, current *model.User, group *model.Group, member *model.Member) error {
	return s.app.createPendingMember(clientID, current, group, member)
}

func (s *servicesImpl) DeletePendingMember(clientID string, current *model.User, groupID string) error {
	return s.app.deletePendingMember(clientID, current, groupID)
}

func (s *servicesImpl) CreateMember(clientID string, current *model.User, group *model.Group, member *model.Member) error {
	return s.app.createMember(clientID, current, group, member)
}

func (s *servicesImpl) DeleteMember(clientID string, current *model.User, groupID string) error {
	return s.app.deleteMember(clientID, current, groupID)
}

func (s *servicesImpl) ApplyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(clientID, current, membershipID, approve, rejectReason)
}

func (s *servicesImpl) DeleteMembership(clientID string, current *model.User, membershipID string) error {
	return s.app.deleteMembership(clientID, current, membershipID)
}

func (s *servicesImpl) UpdateMembership(clientID string, current *model.User, membershipID string, status string, dateAttended *time.Time) error {
	return s.app.updateMembership(clientID, current, membershipID, status, dateAttended)
}

func (s *servicesImpl) GetEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	return s.app.getEvents(clientID, current, groupID, filterByToMembers)
}

func (s *servicesImpl) CreateEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember) (*model.Event, error) {
	return s.app.createEvent(clientID, current, eventID, group, toMemberList)
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

func (s *servicesImpl) UpdatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error) {
	return s.app.updatePost(clientID, current, post)
}

func (s *servicesImpl) ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post) error {
	return s.app.reportPostAsAbuse(clientID, current, group, post)
}

func (s *servicesImpl) DeletePost(clientID string, current *model.User, groupID string, postID string, force bool) error {
	return s.app.deletePost(clientID, current.ID, groupID, postID, force)
}

func (s *servicesImpl) SynchronizeAuthman(clientID string) error {
	return s.app.synchronizeAuthman(clientID)
}

func (s *servicesImpl) SynchronizeAuthmanGroup(clientID string, group *model.Group) error {
	return s.app.synchronizeAuthmanGroup(clientID, group)
}

// Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetGroups(clientID string, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error)
}

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetGroups(clientID string, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error) {
	return s.app.getGroupsUnprotected(clientID, category, privacy, title, offset, limit, order)
}

// Storage is used by corebb to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetStorageListener(storageListener StorageListener)

	FindUser(clientID string, id string, external bool) (*model.User, error)
	FindUsers(clientID string, ids []string, external bool) ([]model.User, error)
	GetUserPostCount(clientID string, userID string) (*int64, error)
	LoginUser(clientID string, current *model.User) error
	CreateUser(clientID string, id string, externalID string, email string, name string) (*model.User, error)
	DeleteUser(clientID string, userID string) error

	ReadAllGroupCategories() ([]string, error)
	FindUserGroupsMemberships(id string, external bool) ([]*model.Group, *model.User, error)

	CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *GroupError)
	UpdateGroup(clientID string, current *model.User, group *model.Group) *GroupError
	DeleteGroup(clientID string, id string) error
	FindGroup(clientID string, id string) (*model.Group, error)
	FindGroupByMembership(clientID string, membershipID string) (*model.Group, error)
	FindGroups(clientID string, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error)
	FindUserGroups(clientID string, userID string) ([]model.Group, error)
	FindUserGroupsCount(clientID string, userID string) (*int64, error)

	UpdateGroupMembers(clientID string, groupID string, members []model.Member) error
	CreatePendingMember(clientID string, current *model.User, group *model.Group, member *model.Member) error
	DeletePendingMember(clientID string, groupID string, userID string) error
	CreateMemberUnchecked(clientID string, current *model.User, group *model.Group, member *model.Member) error
	DeleteMember(clientID string, groupID string, userID string, force bool) error

	ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) error
	DeleteMembership(clientID string, current *model.User, membershipID string) error
	UpdateMembership(clientID string, current *model.User, membershipID string, status string, dateAttended *time.Time) error

	FindEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error)
	CreateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) (*model.Event, error)
	UpdateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error
	DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error

	FindPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error)
	FindPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error)
	FindPostsByParentID(clientID string, userID string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]*model.Post, error)
	CreatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error)
	UpdatePost(clientID string, userID string, post *model.Post) (*model.Post, error)
	ReportPostAsAbuse(clientID string, userID string, group *model.Group, post *model.Post) error
	DeletePost(clientID string, userID string, groupID string, postID string, force bool) error

	FindAuthmanGroups(clientID string) ([]model.Group, error)
	FindAuthmanGroupByKey(clientID string, authmanGroupKey string) (*model.Group, error)
}

//StorageListener listenes for change data storage events
type StorageListener interface {
	OnConfigsChanged()
}

type storageListenerImpl struct {
	app *Application
}

func (a *storageListenerImpl) OnConfigsChanged() {
	//do nothing for now
}

// Notifications exposes Notifications BB APIs for the driver adapters
type Notifications interface {
	SendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string) error
	SendMail(toEmail string, subject string, body string) error
}

type notificationsImpl struct {
	app *Application
}

func (n *notificationsImpl) SendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string) error {
	return n.app.sendNotification(recipients, topic, title, text, data)
}

// Authman exposes Authman APIs for the driver adapters
type Authman interface {
	RetrieveAuthmanGroupMembers(groupName string) ([]string, error)
	RetrieveAuthmanUsers(externalIDs []string) (map[string]model.AuthmanSubject, error)
	RetrieveAuthmanGiesGroups() (*model.АuthmanGroupsResponse, error)
	AddAuthmanMemberToGroup(groupName string, uin string) error
	RemoveAuthmanMemberFromGroup(groupName string, uin string) error
}

// Core exposes Core APIs for the driver adapters
type Core interface {
	RetrieveCoreUserAccount(token string) (*model.CoreAccount, error)
	RetrieveCoreServices(serviceIDs []string) ([]model.CoreService, error)
}

// Rewards exposes Rewards internal APIs for giving rewards to the users
type Rewards interface {
	CreateUserReward(userID string, rewardType string, description string) error
}
