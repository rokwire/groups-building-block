package core

import (
	"groups/core/model"
	"time"
)

func (app *Application) analyticsFindPosts(startDate *time.Time, endDate *time.Time) ([]model.Post, error) {
	return app.storage.AnalyticsFindPosts(startDate, endDate)
}
