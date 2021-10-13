package core

import "groups/core/model"

//Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage       Storage
	notifications Notifications
}

//Start starts the core part of the application
func (app *Application) Start() {
	//set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.SetStorageListener(&storageListener)
}

//FindUser finds an user for the provided external id
func (app *Application) FindUser(clientID string, externalID string) (*model.User, error) {
	user, err := app.storage.FindUser(clientID, externalID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//CreateUser creates an user
func (app *Application) CreateUser(clientID string, externalID string, email string, isMemberOf *[]string) (*model.User, error) {
	user, err := app.storage.CreateUser(clientID, externalID, email, isMemberOf)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//UpdateUser updates the user
func (app *Application) UpdateUser(clientID string, user *model.User) error {
	err := app.storage.SaveUser(clientID, user)
	if err != nil {
		return err
	}
	return nil
}

//NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications) *Application {

	application := Application{version: version, build: build, storage: storage, notifications: notifications}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
