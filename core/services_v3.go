package core

import "groups/core/model"

func (app *Application) findGroupMemberships(clientID string, groupID string, filter *model.MembershipFilter) (model.MembershipCollection, error) {
	return app.storage.FindGroupMemberships(clientID, groupID, filter)
}

func (app *Application) findGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembership(clientID, groupID, userID)
}

func (app *Application) findGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembershipByID(clientID, id)
}

func (app *Application) findUserGroupMemberships(clientID string, userID string) ([]model.GroupMembership, error) {
	return app.storage.FindUserGroupMemberships(clientID, userID)
}
