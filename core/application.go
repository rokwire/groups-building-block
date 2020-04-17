package core

import "groups/core/model"

//Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage Storage
}

//Start starts the core part of the application
func (app *Application) Start() {
	//set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.SetStorageListener(&storageListener)
}

//GetUser gets an user
func (app *Application) FindUser(externalID string) (*model.User, error) {
	user, err := app.storage.FindUser(externalID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//NewApplication creates new Application
func NewApplication(version string, build string, storage Storage) *Application {

	application := Application{version: version, build: build, storage: storage}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
