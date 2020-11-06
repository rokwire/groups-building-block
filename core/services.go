package core

import "groups/core/model"

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getGroupCategories() ([]string, error) {
	groupCategories, err := app.storage.ReadAllGroupCategories()
	if err != nil {
		return nil, err
	}
	return groupCategories, nil
}

func (app *Application) createGroup(current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	insertedID, err := app.storage.CreateGroup(title, description, category, tags, privacy,
		current.ID, creatorName, creatorEmail, creatorPhotoURL)
	if err != nil {
		return nil, err
	}
	return insertedID, nil
}

func (app *Application) getGroups(category *string) ([]model.Group, error) {

	//TODO - data protection
	groups, err := app.storage.FindGroups(category)
	if err != nil {
		return nil, err
	}
	return groups, nil
}
