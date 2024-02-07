package core

import "groups/core/model"

func (app *Application) findAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error) {
	return app.storage.FindAdminGroupsForEvent(nil, clientID, current, eventID)
}
