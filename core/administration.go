package core

import "groups/core/model"

func (app *Application) getTODO() error {
	return nil
}

func (app *Application) getGroupsUnprotected(clientID string, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error) {
	groups, err := app.storage.FindGroups(clientID, category, privacy, title, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
