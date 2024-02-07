package core

import (
	"groups/core/model"
)

func (app *Application) findAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error) {
	return app.storage.FindAdminGroupsForEvent(nil, clientID, current, eventID)
}

func (app *Application) updateGroupMappingsForEvent(clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	return app.storage.UpdateGroupMappingsForEvent(nil, clientID, current, eventID, groupIDs)
}
