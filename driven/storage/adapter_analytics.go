package storage

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"groups/core/model"
	"time"
)

// AnalyticsFindPosts Retrieves analytics posts
func (sa *Adapter) AnalyticsFindPosts(startDate *time.Time, endDate *time.Time) ([]model.Post, error) {
	filter := bson.D{}

	if startDate != nil {
		filter = append(filter, bson.E{Key: "date_created", Value: bson.M{"$gte": *startDate}})
	}
	if endDate != nil {
		filter = append(filter, bson.E{Key: "date_created", Value: bson.M{"$lte": *endDate}})
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "date_created", Value: 1}},
	}

	var list []model.Post
	err := sa.db.posts.Find(filter, &list, opts)
	if err != nil {
		return nil, err
	}

	return list, nil
}
