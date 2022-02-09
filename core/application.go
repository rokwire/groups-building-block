package core

import (
	"errors"
	"groups/core/model"
	"groups/driven/corebb"
)

// Application represents the corebb application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivrs adapters

	storage       Storage
	notifications Notifications
	authman       Authman
	corebb        Core
}

// Start starts the corebb part of the application
func (app *Application) Start() {
	// set storage listener
	storageListener := storageListenerImpl{app: app}
	app.storage.SetStorageListener(&storageListener)
}

// FindUser finds an user for the provided external id
func (app *Application) FindUser(clientID string, id *string, external bool) (*model.User, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}

	if id == nil || *id == "" {
		return nil, errors.New("id cannot be empty")
	}

	user, err := app.storage.FindUser(clientID, *id, external)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// LoginUser refactors the user using the new id
func (app *Application) LoginUser(clientID string, current *model.User, newID string) error {
	return app.storage.LoginUser(clientID, current)
}

// CreateUser creates an user
func (app *Application) CreateUser(clientID string, id string, externalID *string, email *string, name *string) (*model.User, error) {
	externalIDVal := ""
	if externalID != nil {
		externalIDVal = *externalID
	}

	emailVal := ""
	if email != nil {
		emailVal = *email
	}

	nameVal := ""
	if name != nil {
		nameVal = *name
	}

	user, err := app.storage.CreateUser(clientID, id, externalIDVal, emailVal, nameVal)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, notifications Notifications, authman Authman, core *corebb.Adapter) *Application {

	application := Application{version: version, build: build, storage: storage, notifications: notifications,
		authman: authman, corebb: core}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
