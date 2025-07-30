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
	"context"
	"fmt"
	"groups/core/model"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client

	configs             *collectionWrapper
	syncTimes           *collectionWrapper
	enums               *collectionWrapper
	groups              *collectionWrapper
	groupMemberships    *collectionWrapper
	events              *collectionWrapper
	managedGroupConfigs *collectionWrapper

	listeners []Listener
}

func (m *database) start() error {
	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	// instance db
	db := client.Database(m.mongoDBName)

	//drop users and posts collections if exists
	m.dropCollectionIfExists(db, "users")
	m.dropCollectionIfExists(db, "posts")

	configs := &collectionWrapper{database: m, coll: db.Collection("configs")}
	err = m.applyConfigsChecks(configs)
	if err != nil {
		return err
	}

	syncTimes := &collectionWrapper{database: m, coll: db.Collection("sync_times")}
	err = m.applySyncTimesChecks(syncTimes)
	if err != nil {
		return err
	}

	enums := &collectionWrapper{database: m, coll: db.Collection("enums")}
	err = m.applyEnumsChecks(enums)
	if err != nil {
		return err
	}

	groups := &collectionWrapper{database: m, coll: db.Collection("groups")}
	err = m.applyGroupsChecks(groups)
	if err != nil {
		return err
	}

	groupMemberships := &collectionWrapper{database: m, coll: db.Collection("group_memberships")}
	err = m.applyGroupMembershipsChecks(groupMemberships)
	if err != nil {
		return err
	}

	events := &collectionWrapper{database: m, coll: db.Collection("events")}
	err = m.applyEventsChecks(events)
	if err != nil {
		return err
	}

	managedGroupConfigs := &collectionWrapper{database: m, coll: db.Collection("managed_group_configs")}
	err = m.applyManagedGroupConfigsChecks(managedGroupConfigs)
	if err != nil {
		return err
	}

	// apply membership transition
	err = m.ApplyMembershipTransition(client, groups, groupMemberships)
	if err != nil {
		return err
	}

	// apply default group settings
	err = m.ApplyDefaultGroupSettings(client, groups)
	if err != nil {
		return err
	}

	err = m.ApplyGroupsAttributesTransition(client, groups)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.configs = configs
	m.syncTimes = syncTimes
	m.enums = enums
	m.groups = groups
	m.groupMemberships = groupMemberships
	m.events = events
	m.managedGroupConfigs = managedGroupConfigs

	go m.configs.Watch(nil)
	go m.managedGroupConfigs.Watch(nil)

	m.listeners = []Listener{}

	return nil
}

func (m *database) dropCollectionIfExists(db *mongo.Database, name string) {
	collections, err := db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		log.Printf("error listing collections: %v", err)
	}
	for _, collName := range collections {
		if collName == name {
			err = db.Collection(name).Drop(context.Background())
			if err != nil {
				log.Printf("error dropping collection %s: %v", name, err)
			} else {
				log.Printf("dropped collection %s", name)
			}
		}
	}
}

func (m *database) applyConfigsChecks(configs *collectionWrapper) error {
	log.Println("apply configs checks.....")

	err := configs.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "type", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("configs checks passed")
	return nil
}

func (m *database) applySyncTimesChecks(syncTimes *collectionWrapper) error {
	log.Println("apply sync times checks.....")

	indexes, err := syncTimes.ListIndexes()
	if err != nil {
		return err
	}
	if len(indexes) > 0 {
		for _, index := range indexes {
			if index["name"] == "client_id_1" {
				log.Printf("drop client_id_1 index")
				err = syncTimes.DropIndex("client_id_1")
				if err != nil {
					return err
				}
				break
			}

		}
	}

	err = syncTimes.AddIndex(bson.D{
		primitive.E{Key: "key", Value: 1},
	}, true)
	if err != nil {
		return err
	}

	log.Println("sync times checks passed")
	return nil
}

func (m *database) applyEnumsChecks(enums *collectionWrapper) error {
	log.Println("apply enums checks.....")

	log.Println("enums checks passed")
	return nil
}

func (m *database) applyGroupsChecks(groups *collectionWrapper) error {
	log.Println("apply groups checks.....")

	indexes, _ := groups.ListIndexes()
	indexMapping := map[string]interface{}{}

	for _, index := range indexes {
		name := index["name"].(string)
		indexMapping[name] = index
	}

	if indexMapping["client_id_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "org_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["category_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "category", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["privacy_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "privacy", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["privacy_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "privacy", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_created_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "date_created", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["authman_enabled_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "authman_enabled", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["members.id_1"] != nil {
		err := groups.DropIndex("members.id_1")
		if err != nil {
			return err
		}
	}

	if indexMapping["members.user_id_1"] != nil {
		err := groups.DropIndex("members.user_id_1")
		if err != nil {
			return err
		}
	}

	if indexMapping["research_group_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "research_group", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["research_open_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "research_open", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["title_1"] == nil {
		err := groups.AddIndex(
			bson.D{
				primitive.E{Key: "title", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["client_id_1_title_1_"] != nil {
		err := groups.DropIndex("client_id_1_title_1_")
		if err != nil {
			return err
		}
	}

	if indexMapping["client_id_1_title_1"] != nil {
		err := groups.DropIndex("client_id_1_title_1")
		if err != nil {
			return err
		}
	}

	if indexMapping["title_1_client_id_1"] != nil {
		// Drop the old one
		err := groups.DropIndex("title_1_client_id_1")
		if err != nil {
			return err
		}
	}

	name := "title_unique"
	unique := true
	if indexMapping["title_unique"] == nil {
		err := groups.AddIndexWithOptions(
			bson.D{
				primitive.E{Key: "org_id", Value: 1},
				primitive.E{Key: "title", Value: 1},
			},
			&options.IndexOptions{
				Name:   &name,
				Unique: &unique,
				Collation: &options.Collation{
					Locale:   "en",
					Strength: 2,
				},
			})
		if err != nil {
			return err
		}
	}
	log.Println("groups checks passed")
	return nil
}

func (m *database) applyGroupMembershipsChecks(groupMemberships *collectionWrapper) error {
	log.Println("apply group memberships checks.....")

	err := groupMemberships.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "group_id", Value: 1}, primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "group_id", Value: 1}, primitive.E{Key: "external_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "group_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "name", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "net_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "email", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "status", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{
		primitive.E{Key: "status", Value: 1},
		primitive.E{Key: "name", Value: 1},
	}, false)
	if err != nil {
		return err
	}

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "date_created", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("group memberships checks passed")
	return nil
}

func (m *database) applyEventsChecks(events *collectionWrapper) error {
	log.Println("apply events checks.....")

	indexes, _ := events.ListIndexes()
	indexMapping := map[string]interface{}{}
	for _, index := range indexes {
		name := index["name"].(string)
		indexMapping[name] = index
	}

	if indexMapping["event_id_1_group_id_1_client_id_1"] == nil {
		err := events.AddIndex(bson.D{
			primitive.E{Key: "event_id", Value: 1},
			primitive.E{Key: "group_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1}},
			true)
		if err != nil {
			return err
		}
	}

	if indexMapping["title_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "title", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["event_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "event_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["group_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "group_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["client_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "org_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["member.user_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "member.user_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["to_members.user_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "to_members.user_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["to_members.external_id_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "to_members.external_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["to_members.email_1"] == nil {
		err := events.AddIndex(
			bson.D{
				primitive.E{Key: "to_members.email", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	log.Println("events checks passed")
	return nil
}

func (m *database) applyManagedGroupConfigsChecks(managedGroupConfigs *collectionWrapper) error {
	log.Println("apply managed group configs checks.....")

	//TODO: Set up indexes

	log.Println("managed group configs checks passed")
	return nil
}

func (m *database) ApplyMembershipTransition(client *mongo.Client, groups *collectionWrapper, groupMemberships *collectionWrapper) error {
	log.Println("apply memberships transition checks.....")

	var migrationGroup []model.Group
	err := groups.Find(bson.D{
		{Key: "members.id", Value: bson.M{"$exists": true}},
	}, &migrationGroup, nil)
	if err != nil {
		return err
	}

	if len(migrationGroup) > 0 {
		err = client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				return err
			}

			_, err = groups.UpdateManyWithContext(sessionContext, bson.D{}, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "stats", Value: model.GroupStats{}},
				}},
			}, nil)
			if err != nil {
				abortTransaction(sessionContext)
				return err
			}

			for _, group := range migrationGroup {
				log.Printf("Start migrating '%s' group", group.Title)
				memberships := []interface{}{}
				stats := model.GroupStats{}
				for _, member := range group.Members {
					if member.Status == "pending" {
						stats.PendingCount++
					} else if member.Status == "rejected" {
						stats.RejectedCount++
					} else if member.Status == "member" {
						stats.TotalCount++
						stats.MemberCount++
					} else if member.Status == "admin" {
						stats.TotalCount++
						stats.AdminsCount++
					}

					if member.DateAttended != nil {
						stats.AttendanceCount++
					}

					memberships = append(memberships, member.ToGroupMembership(group.OrgID, group.ID))
				}

				_, err = groupMemberships.InsertManyWithContext(sessionContext, memberships, &options.InsertManyOptions{})
				if err != nil {
					abortTransaction(sessionContext)
					return err
				}

				_, err = groups.UpdateOneWithContext(sessionContext, bson.D{
					{Key: "org_id", Value: group.OrgID},
					{Key: "_id", Value: group.ID},
				}, bson.D{
					{Key: "$set", Value: bson.D{
						{Key: "members", Value: nil},
						{Key: "stats", Value: stats},
					}},
				}, nil)
				if err != nil {
					abortTransaction(sessionContext)
					return err
				}

				log.Printf("Grouop '%s' has been migrated successfull", group.Title)
			}

			//commit the transaction
			err = sessionContext.CommitTransaction(sessionContext)
			if err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	log.Println("memberships transition passed")
	return nil
}

func (m *database) ApplyDefaultGroupSettings(client *mongo.Client, groups *collectionWrapper) error {
	log.Println("apply group settings migration.....")

	err := client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}
		var migrationGroup []model.Group

		filter := bson.D{
			{Key: "settings", Value: bson.M{"$exists": false}},
		}

		err = groups.FindWithContext(sessionContext, filter, &migrationGroup, nil)
		if err != nil {
			return err
		}

		if len(migrationGroup) > 0 {
			_, err = groups.UpdateManyWithContext(sessionContext, filter, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "settings", Value: model.DefaultGroupSettings()},
				}},
			}, nil)
			if err != nil {
				abortTransaction(sessionContext)
				return err
			}
		}

		//commit the transaction
		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	log.Println("group settings migration passed")
	return nil
}

func (m *database) ApplyGroupsAttributesTransition(client *mongo.Client, groups *collectionWrapper) error {
	log.Println("apply group attributes migration.....")

	err := client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		filter := bson.D{
			{Key: "$or", Value: []bson.M{
				{"attributes": bson.M{"$exists": false}},
				{"attributes.category": "$category"},
				{"attributes.tags": "$tags"},
			}},
		}
		_, err = groups.UpdateManyWithContext(sessionContext, filter,
			[]bson.M{
				{"$set": bson.M{
					"attributes.category": "$category",
					"attributes.tags":     "$tags",
				}},
			}, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}

		//commit the transaction
		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	log.Println("group attributes migration passed")
	return nil
}

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	switch coll {
	case "configs":
		log.Println("configs collection changed")

		for _, listener := range m.listeners {
			//Don't use goroutine to ensure cache is updated first
			listener.OnConfigsChanged()
		}
	case "managed_group_configs":
		log.Println("managed_group_configs collection changed")

		for _, listener := range m.listeners {
			go listener.OnManagedGroupConfigsChanged()
		}
	}
}
