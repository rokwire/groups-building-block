// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"groups/core/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AnalyticsFindGroups Retrieves analytics groups
func (sa *Adapter) AnalyticsFindGroups(startDate *time.Time, endDate *time.Time) ([]model.Group, error) {
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

	var list []model.Group
	err := sa.db.groups.Find(filter, &list, opts)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// AnalyticsFindMembers Retrieves analytics groups members
func (sa *Adapter) AnalyticsFindMembers(groupID *string, startDate *time.Time, endDate *time.Time) ([]model.GroupMembership, error) {
	filter := bson.D{}

	if groupID != nil {
		filter = append(filter, bson.E{Key: "group_id", Value: *groupID})
	}
	if startDate != nil {
		filter = append(filter, bson.E{Key: "date_created", Value: bson.M{"$gte": *startDate}})
	}
	if endDate != nil {
		filter = append(filter, bson.E{Key: "date_created", Value: bson.M{"$lte": *endDate}})
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "date_created", Value: 1}},
	}

	var list []model.GroupMembership
	err := sa.db.groupMemberships.Find(filter, &list, opts)
	if err != nil {
		return nil, err
	}

	return list, nil
}
