package core

import (
	"groups/core/model"
)

//Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetGroupCategories() ([]string, error)

	GetGroupEntity(clientID string, id string) (*model.Group, error)
	GetGroupEntityByMembership(clientID string, membershipID string) (*model.Group, error)

	CreateGroup(clientID string, current model.User, title string, description *string, category string, tags []string, privacy string,
		creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error)
	UpdateGroup(clientID string, current *model.User, id string, category string, title string, privacy string, description *string,
		imageURL *string, webURL *string, tags []string, membershipQuestions []string) error
	GetGroups(clientID string, current *model.User, category *string) ([]map[string]interface{}, error)
	GetUserGroups(clientID string, current *model.User) ([]map[string]interface{}, error)
	GetGroup(clientID string, current *model.User, id string) (map[string]interface{}, error)

	CreatePendingMember(clientID string, current model.User, groupID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error
	DeletePendingMember(clientID string, current model.User, groupID string) error
	DeleteMember(clientID string, current model.User, groupID string) error

	ApplyMembershipApproval(clientID string, current model.User, membershipID string, approve bool, rejectReason string) error
	DeleteMembership(clientID string, current model.User, membershipID string) error
	UpdateMembership(clientID string, current model.User, membershipID string, status string) error

	GetEvents(clientID string, groupID string) ([]model.Event, error)
	CreateEvent(clientID string, current model.User, eventID string, groupID string) error
	DeleteEvent(clientID string, current model.User, eventID string, groupID string) error
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

func (s *servicesImpl) GetGroupEntity(clientID string, id string) (*model.Group, error) {
	return s.app.getGroupEntity(clientID, id)
}

func (s *servicesImpl) GetGroupEntityByMembership(clientID string, membershipID string) (*model.Group, error) {
	return s.app.getGroupEntityByMembership(clientID, membershipID)
}

func (s *servicesImpl) CreateGroup(clientID string, current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	return s.app.createGroup(clientID, current, title, description, category, tags, privacy, creatorName, creatorEmail, creatorPhotoURL)
}

func (s *servicesImpl) UpdateGroup(clientID string, current *model.User, id string, category string, title string, privacy string, description *string,
	imageURL *string, webURL *string, tags []string, membershipQuestions []string) error {
	return s.app.updateGroup(clientID, current, id, category, title, privacy, description, imageURL, webURL, tags, membershipQuestions)
}

func (s *servicesImpl) GetGroups(clientID string, current *model.User, category *string) ([]map[string]interface{}, error) {
	return s.app.getGroups(clientID, current, category)
}

func (s *servicesImpl) GetUserGroups(clientID string, current *model.User) ([]map[string]interface{}, error) {
	return s.app.getUserGroups(clientID, current)
}

func (s *servicesImpl) GetGroup(clientID string, current *model.User, id string) (map[string]interface{}, error) {
	return s.app.getGroup(clientID, current, id)
}

func (s *servicesImpl) CreatePendingMember(clientID string, current model.User, groupID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error {
	return s.app.createPendingMember(clientID, current, groupID, name, email, photoURL, memberAnswers)
}

func (s *servicesImpl) DeletePendingMember(clientID string, current model.User, groupID string) error {
	return s.app.deletePendingMember(clientID, current, groupID)
}

func (s *servicesImpl) DeleteMember(clientID string, current model.User, groupID string) error {
	return s.app.deleteMember(clientID, current, groupID)
}

func (s *servicesImpl) ApplyMembershipApproval(clientID string, current model.User, membershipID string, approve bool, rejectReason string) error {
	return s.app.applyMembershipApproval(clientID, current, membershipID, approve, rejectReason)
}

func (s *servicesImpl) DeleteMembership(clientID string, current model.User, membershipID string) error {
	return s.app.deleteMembership(clientID, current, membershipID)
}

func (s *servicesImpl) UpdateMembership(clientID string, current model.User, membershipID string, status string) error {
	return s.app.updateMembership(clientID, current, membershipID, status)
}

func (s *servicesImpl) GetEvents(clientID string, groupID string) ([]model.Event, error) {
	return s.app.getEvents(clientID, groupID)
}

func (s *servicesImpl) CreateEvent(clientID string, current model.User, eventID string, groupID string) error {
	return s.app.createEvent(clientID, current, eventID, groupID)
}

func (s *servicesImpl) DeleteEvent(clientID string, current model.User, eventID string, groupID string) error {
	return s.app.deleteEvent(clientID, current, eventID, groupID)
}

//Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetConfig() (*model.GroupsConfig, error)
	UpdateConfig(config *model.GroupsConfig) error
	GetTODO() error
}

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetTODO() error {
	return s.app.getTODO()
}
func (s *administrationImpl) GetConfig() (*model.GroupsConfig, error) {
	return s.app.getConfig()
}
func (s *administrationImpl) UpdateConfig(config *model.GroupsConfig) error {
	return s.app.updateConfig(config)
}

//Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetStorageListener(storageListener StorageListener)

	FindUser(clientID string, externalID string) (*model.User, error)
	CreateUser(clientID string, externalID string, email string, isMemberOf *[]string) (*model.User, error)
	SaveUser(clientID string, user *model.User) error

	ReadAllGroupCategories() ([]string, error)

	CreateGroup(clientID string, title string, description *string, category string, tags []string, privacy string,
		creatorUserID string, creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error)
	UpdateGroup(clientID string, id string, category string, title string, privacy string, description *string,
		imageURL *string, webURL *string, tags []string, membershipQuestions []string) error
	FindGroup(clientID string, id string) (*model.Group, error)
	FindGroupByMembership(clientID string, membershipID string) (*model.Group, error)
	FindGroups(clientID string, category *string) ([]model.Group, error)
	FindUserGroups(clientID string, userID string) ([]model.Group, error)

	CreatePendingMember(clientID string, groupID string, userID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error
	DeletePendingMember(clientID string, groupID string, userID string) error
	DeleteMember(clientID string, groupID string, userID string) error

	ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) error
	DeleteMembership(clientID string, currentUserID string, membershipID string) error
	UpdateMembership(clientID string, currentUserID string, membershipID string, status string) error

	FindEvents(clientID string, groupID string) ([]model.Event, error)
	CreateEvent(clientID string, eventID string, groupID string) error
	DeleteEvent(clientID string, eventID string, groupID string) error

	SaveConfig(Config *model.GroupsConfig) error
	ReadConfig() (*model.GroupsConfig, error)
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
