package core

import (
	"groups/core/model"
)

//Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetGroupEntity(id string) (*model.Group, error)

	GetGroupCategories() ([]string, error)

	CreateGroup(current model.User, title string, description *string, category string, tags []string, privacy string,
		creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error)
	GetGroups(category *string) ([]map[string]interface{}, error)
	GetUserGroups(current *model.User) ([]map[string]interface{}, error)
	GetGroup(current *model.User, id string) (map[string]interface{}, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) GetGroupEntity(id string) (*model.Group, error) {
	return s.app.getGroupEntity(id)
}

func (s *servicesImpl) GetGroupCategories() ([]string, error) {
	return s.app.getGroupCategories()
}

func (s *servicesImpl) CreateGroup(current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	return s.app.createGroup(current, title, description, category, tags, privacy, creatorName, creatorEmail, creatorPhotoURL)
}

func (s *servicesImpl) GetGroups(category *string) ([]map[string]interface{}, error) {
	return s.app.getGroups(category)
}

func (s *servicesImpl) GetUserGroups(current *model.User) ([]map[string]interface{}, error) {
	return s.app.getUserGroups(current)
}

func (s *servicesImpl) GetGroup(current *model.User, id string) (map[string]interface{}, error) {
	return s.app.getGroup(current, id)
}

//Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetTODO() error
}

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetTODO() error {
	return s.app.getTODO()
}

//Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetStorageListener(storageListener StorageListener)

	FindUser(externalID string) (*model.User, error)
	CreateUser(externalID string, email string, isMemberOf *[]string) (*model.User, error)
	SaveUser(user *model.User) error

	ReadAllGroupCategories() ([]string, error)

	CreateGroup(title string, description *string, category string, tags []string, privacy string,
		creatorUserID string, creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error)
	FindGroup(id string) (*model.Group, error)
	FindGroups(category *string) ([]model.Group, error)
	FindUserGroups(userID string) ([]model.Group, error)
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
