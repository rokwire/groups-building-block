package core

import "groups/core/model"

func (app *Application) findGroupV3(clientID string, filter *model.GroupsFilter) (*model.Group, error) {
	// assume we filter one nd just return the first one. Enough for now
	groups, err := app.findGroupsV3(clientID, filter)
	if len(groups) > 0 {
		return &groups[0], err
	}
	return nil, err
}

func (app *Application) findGroupsV3(clientID string, filter *model.GroupsFilter) ([]model.Group, error) {
	return app.storage.FindGroupsV3(clientID, filter)
}

func (app *Application) findGroupMemberships(clientID string, filter *model.MembershipFilter) (model.MembershipCollection, error) {
	return app.storage.FindGroupMemberships(clientID, filter)
}

func (app *Application) findGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembership(clientID, groupID, userID)
}

func (app *Application) findGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembershipByID(clientID, id)
}

func (app *Application) findUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error) {
	return app.storage.FindUserGroupMemberships(clientID, userID)
}
