package core

import "groups/core/model"

//TODO
func (app *Application) applyDataProtection(current *model.User, group model.Group) map[string]interface{} {
	item := make(map[string]interface{})

	item["id"] = group.ID
	item["title"] = group.Title

	return item
}

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

func (app *Application) getGroups(category *string) ([]map[string]interface{}, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(category)
	if err != nil {
		return nil, err
	}

	//apply data protection
	groupsList := make([]map[string]interface{}, len(groups))
	for i, item := range groups {
		groupsList[i] = app.applyDataProtection(nil, item)
	}

	return groupsList, nil
}
