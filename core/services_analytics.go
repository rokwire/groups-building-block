package core

import (
	"groups/core/model"
	"time"
)

func (app *Application) analyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error) {
	return app.storage.AnalyticsFindGroups(startDate, endDate)
}

func (app *Application) analyticsFindPosts(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.Post, error) {
	return app.storage.AnalyticsFindPosts(groupID, startDate, endDate)
}

func (app *Application) analyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error) {
	return app.storage.AnalyticsFindMembers(groupID, startDate, endDate)
}
