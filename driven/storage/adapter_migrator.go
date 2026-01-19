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
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MigrateGroups migrates groups and related records to use org_id instead of client_id
func (sa *Adapter) MigrateGroups(ctx TransactionContext, defaultOrgID string) error {

	wrapper := (func(context TransactionContext) error {
		filter := bson.D{
			{Key: "$or",
				Value: bson.A{
					bson.D{{Key: "org_id", Value: primitive.Null{}}},
					bson.D{{Key: "org_id", Value: bson.D{{Key: "$exists", Value: false}}}},
					bson.D{{Key: "org_id", Value: ""}},
				},
			},
		}
		update := bson.D{
			{Key: "$set", Value: bson.M{"org_id": defaultOrgID}},
			{Key: "$unset", Value: bson.M{"client_id": ""}},
		}
		log.Printf("Starting migration for org_id %s", defaultOrgID)
		defer log.Printf("Finished migration for org_id %s", defaultOrgID)

		result, err := sa.db.configs.UpdateManyWithContext(ctx, filter, update, nil)
		if err != nil {
			return err
		}
		log.Printf("configs: updated %d records to org_id %s", result.ModifiedCount, defaultOrgID)

		result, err = sa.db.enums.UpdateManyWithContext(ctx, filter, update, nil)
		if err != nil {
			return err
		}
		log.Printf("enums: updated %d records to org_id %s", result.ModifiedCount, defaultOrgID)

		result, err = sa.db.groups.UpdateManyWithContext(ctx, filter, update, nil)
		if err != nil {
			return err
		}
		log.Printf("groups: updated %d records to org_id %s", result.ModifiedCount, defaultOrgID)

		result, err = sa.db.syncTimes.UpdateManyWithContext(ctx, filter, update, nil)
		if err != nil {
			return err
		}
		log.Printf("syncTimes: updated %d records to org_id %s", result.ModifiedCount, defaultOrgID)

		result, err = sa.db.groupMemberships.UpdateManyWithContext(ctx, filter, update, nil)
		if err != nil {
			return err
		}
		log.Printf("groupMemberships: updated %d records to org_id %s", result.ModifiedCount, defaultOrgID)
		return nil
	})

	if ctx != nil {
		return wrapper(ctx)
	}
	return sa.PerformTransaction(wrapper)

}
