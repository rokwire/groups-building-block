package core

import "groups/core/model"

//Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
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
