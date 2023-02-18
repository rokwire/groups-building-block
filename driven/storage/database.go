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
	"groups/core/model"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
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

	configs          *collectionWrapper
	syncTimes        *collectionWrapper
	users            *collectionWrapper
	groups           *collectionWrapper
	groupMemberships *collectionWrapper
	events           *collectionWrapper
	posts            *collectionWrapper

	listeners []Listener
}

func (m *database) start(defaultAppID string, defaultOrgID string, defaultAppConfig *model.Config) error {
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

	//apply checks
	db := client.Database(m.mongoDBName)

	//assign the db, db client and the collections
	m.db = db
	m.dbClient = client

	syncTimes := &collectionWrapper{database: m, coll: db.Collection("sync_times")}
	err = m.applySyncTimesChecks(syncTimes)
	if err != nil {
		return err
	}

	users := &collectionWrapper{database: m, coll: db.Collection("users")}
	err = m.applyUsersChecks(users)
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

	posts := &collectionWrapper{database: m, coll: db.Collection("posts")}
	err = m.applyPostsChecks(posts)
	if err != nil {
		return err
	}

	managedGroupConfigs := &collectionWrapper{database: m, coll: db.Collection("managed_group_configs")}
	configs := &collectionWrapper{database: m, coll: db.Collection("configs")}
	err = m.applyConfigsChecks(configs, managedGroupConfigs, defaultAppID, defaultOrgID, defaultAppConfig)
	if err != nil {
		return err
	}

	m.configs = configs
	m.syncTimes = syncTimes
	m.users = users
	m.groups = groups
	m.groupMemberships = groupMemberships
	m.events = events
	m.posts = posts

	//apply multi-tenant
	err = m.applyMultiTenantChecks(defaultAppID, defaultOrgID)
	if err != nil {
		return err
	}

	// apply membership transition
	err = m.ApplyMembershipTransition()
	if err != nil {
		return err
	}

	// apply default group settings
	err = m.ApplyDefaultGroupSettings()
	if err != nil {
		return err
	}

	err = m.ApplyGroupsAttributesTransition()
	if err != nil {
		return err
	}

	go m.configs.Watch(nil)

	m.listeners = []Listener{}

	return nil
}

func (m *database) applyConfigsChecks(configs *collectionWrapper, managedGroupConfigs *collectionWrapper, defaultAppID string, defaultOrgID string, defaultAppConfig *model.Config) error {
	log.Println("apply configs checks.....")

	configs.DropIndex("client_id_1_type_1")

	// need to do the transition after existing index has been removed and before the new one is created
	err := m.applyConfigsTransition(configs, managedGroupConfigs, defaultAppID, defaultOrgID, defaultAppConfig)
	if err != nil {
		return err
	}

	err = configs.AddIndex(bson.D{primitive.E{Key: "type", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("configs checks passed")
	return nil
}

func (m *database) applySyncTimesChecks(syncTimes *collectionWrapper) error {
	log.Println("apply sync times checks.....")

	syncTimes.DropIndex("client_id_1")
	err := syncTimes.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("sync times checks passed")
	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	log.Println("apply users checks.....")

	err := users.AddIndex(bson.D{primitive.E{Key: "external_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	// replace clientID indexes with appID, orgID indexes
	users.DropIndex("client_id_1")
	err = users.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	users.DropIndex("external_id_1_client_id_1")
	err = users.AddIndex(bson.D{primitive.E{Key: "external_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("users checks passed")
	return nil
}

func (m *database) applyGroupsChecks(groups *collectionWrapper) error {
	log.Println("apply groups checks.....")

	// replace clientID index with appID, orgID
	groups.DropIndex("client_id_1")
	err := groups.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "category", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "privacy", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "privacy", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "date_created", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "authman_enabled", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "research_group", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "research_open", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "title", Value: 1}}, false)
	if err != nil {
		return err
	}

	groups.DropIndex("members.id_1")
	groups.DropIndex("members.user_id_1")
	groups.DropIndex("client_id_1_title_1_")
	groups.DropIndex("client_id_1_title_1")
	groups.DropIndex("title_1_client_id_1")

	name := "title_unique"
	unique := true
	groups.DropIndex("title_unique")
	err = groups.AddIndexWithOptions(
		bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "title", Value: 1}},
		&options.IndexOptions{Name: &name, Unique: &unique, Collation: &options.Collation{Locale: "en", Strength: 2}},
	)
	if err != nil {
		return err
	}

	log.Println("groups checks passed")
	return nil
}

func (m *database) applyGroupMembershipsChecks(groupMemberships *collectionWrapper) error {
	log.Println("apply group memberships checks.....")

	// replace clientID indexes with appID, orgID indexes
	groupMemberships.DropIndex("client_id_1_group_id_1_user_id_1")
	err := groupMemberships.AddIndex(bson.D{
		primitive.E{Key: "app_id", Value: 1},
		primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "group_id", Value: 1},
		primitive.E{Key: "user_id", Value: 1},
	}, false)
	if err != nil {
		return err
	}

	groupMemberships.DropIndex("client_id_1_user_id_1")
	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	groupMemberships.DropIndex("client_id_1_group_id_1_external_id_1")
	err = groupMemberships.AddIndex(bson.D{
		primitive.E{Key: "app_id", Value: 1},
		primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "group_id", Value: 1},
		primitive.E{Key: "external_id", Value: 1},
	}, false)
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

	err = groupMemberships.AddIndex(bson.D{primitive.E{Key: "status", Value: 1}, primitive.E{Key: "name", Value: 1}}, false)
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

	// replace clientID indexes with appID, orgID indexes
	events.DropIndex("event_id_1_group_id_1_client_id_1")
	err := events.AddIndex(bson.D{
		primitive.E{Key: "event_id", Value: 1},
		primitive.E{Key: "group_id", Value: 1},
		primitive.E{Key: "app_id", Value: 1},
		primitive.E{Key: "org_id", Value: 1},
	}, true)
	if err != nil {
		return err
	}

	events.DropIndex("client_id_1")
	err = events.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "title", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "event_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "group_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "member.user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "to_members.user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "to_members.external_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "to_members.email", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("events checks passed")
	return nil
}

func (m *database) applyPostsChecks(posts *collectionWrapper) error {
	log.Println("apply posts checks.....")

	// replace clientID indexes with appID, orgID indexes
	posts.DropIndex("client_id_1")
	err := posts.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	posts.DropIndex("private_1_client_id_1__id_1")
	err = posts.AddIndex(bson.D{
		primitive.E{Key: "private", Value: 1},
		primitive.E{Key: "app_id", Value: 1},
		primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "_id", Value: 1},
	}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "private", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "date_created", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "top_parent_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "member.user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "to_members.user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "to_members.external_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = posts.AddIndex(bson.D{primitive.E{Key: "to_members.email", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("posts checks passed")
	return nil
}

func (m *database) applyMultiTenantChecks(defaultAppID string, defaultOrgID string) error {
	log.Println("apply multi-tenant checks.....")

	filter := bson.D{primitive.E{Key: "app_id", Value: bson.M{"$exists": false}}, primitive.E{Key: "org_id", Value: bson.M{"$exists": false}}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "app_id", Value: defaultAppID},
			primitive.E{Key: "org_id", Value: defaultOrgID},
		}},
		primitive.E{Key: "$unset", Value: bson.D{
			primitive.E{Key: "client_id", Value: 1},
		}},
	}

	// transaction
	transaction := func(context TransactionContext) error {
		//apply users collection
		_, err := m.users.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		//apply groups collection
		_, err = m.groups.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		//apply events collection
		_, err = m.events.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		//apply posts collection
		_, err = m.posts.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		//apply memberships collection
		_, err = m.groupMemberships.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		//apply sync times collection
		_, err = m.syncTimes.UpdateManyWithContext(context, filter, update, nil)
		if err != nil {
			return err
		}

		return nil
	}

	err := m.performTransaction(transaction)
	if err != nil {
		return err
	}

	log.Println("multi-tenant checks passed")
	return nil
}

func (m *database) ApplyMembershipTransition() error {
	log.Println("apply memberships transition checks.....")

	var migrationGroup []model.Group
	err := m.groups.Find(bson.D{
		{Key: "members.id", Value: bson.M{"$exists": true}},
	}, &migrationGroup, nil)
	if err != nil {
		return err
	}

	if len(migrationGroup) > 0 {
		transaction := func(context TransactionContext) error {
			_, err = m.groups.UpdateManyWithContext(context, bson.D{}, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "stats", Value: model.GroupStats{}},
				}},
			}, nil)
			if err != nil {
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

					memberships = append(memberships, member.ToGroupMembership(group.AppID, group.OrgID, group.ID))
				}

				_, err = m.groupMemberships.InsertManyWithContext(context, memberships, &options.InsertManyOptions{})
				if err != nil {
					return err
				}

				_, err = m.groups.UpdateOneWithContext(context, bson.D{
					{Key: "app_id", Value: group.AppID},
					{Key: "org_id", Value: group.OrgID},
					{Key: "_id", Value: group.ID},
				}, bson.D{
					{Key: "$set", Value: bson.D{
						{Key: "members", Value: nil},
						{Key: "stats", Value: stats},
					}},
				}, nil)
				if err != nil {
					return err
				}

				log.Printf("Group '%s' has been migrated successfully", group.Title)
			}

			return nil
		}

		err := m.performTransaction(transaction)
		if err != nil {
			return err
		}
	}

	log.Println("memberships transition passed")
	return nil
}

func (m *database) ApplyDefaultGroupSettings() error {
	log.Println("apply group settings migration.....")

	transaction := func(context TransactionContext) error {
		var migrationGroup []model.Group
		filter := bson.D{{Key: "settings", Value: bson.M{"$exists": false}}}
		err := m.groups.FindWithContext(context, filter, &migrationGroup, nil)
		if err != nil {
			return err
		}

		if len(migrationGroup) > 0 {
			_, err = m.groups.UpdateManyWithContext(context, filter, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "settings", Value: model.DefaultGroupSettings()},
				}},
			}, nil)
			return err
		}

		return nil
	}

	err := m.performTransaction(transaction)
	if err != nil {
		return err
	}

	log.Println("group settings migration passed")
	return nil
}

func (m *database) ApplyGroupsAttributesTransition() error {
	log.Println("apply group attributes migration.....")

	filter := bson.D{{Key: "attributes", Value: bson.M{"$exists": false}}}
	_, err := m.groups.UpdateMany(filter, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "attributes.category", Value: "$category"},
			{Key: "attributes.tags", Value: "$tags"},
		}},
	}, nil)
	if err != nil {
		return err
	}

	log.Println("group attributes migration passed")
	return nil
}

func (m *database) applyConfigsTransition(configs *collectionWrapper, managedGroupConfigs *collectionWrapper, defaultAppID string, defaultOrgID string, defaultAppConfig *model.Config) error {
	log.Println("apply configs migration.....")

	transaction := func(context TransactionContext) error {
		now := time.Now()

		//1. insert default app config if provided and not already existing
		if defaultAppConfig != nil {
			var appConfigs []model.Config
			appConfigFilter := bson.M{"type": defaultAppConfig.Type, "app_id": defaultAppID, "org_id": defaultOrgID}
			err := configs.FindWithContext(context, appConfigFilter, &appConfigs, nil)
			if err != nil {
				return err
			}

			if len(appConfigs) == 0 {
				defaultAppConfig.ID = uuid.NewString()
				defaultAppConfig.AppID = defaultAppID
				defaultAppConfig.OrgID = defaultOrgID
				defaultAppConfig.DateCreated = now
				_, err = configs.InsertOneWithContext(context, defaultAppConfig)
				if err != nil {
					return err
				}
			}
		}

		//2. wrap existing sync configs in new config model
		var syncConfigs []model.SyncConfigData
		syncConfigFilter := bson.M{"type": model.ConfigTypeSync, "client_id": "edu.illinois.rokwire"}
		err := configs.FindWithContext(context, syncConfigFilter, &syncConfigs, nil)
		if err != nil {
			return err
		}

		if len(syncConfigs) > 0 {
			_, err = configs.DeleteManyWithContext(context, syncConfigFilter, nil)
			if err != nil {
				return err
			}
		}

		for _, syncConfig := range syncConfigs {
			newSyncConfig := model.Config{
				ID:          uuid.NewString(),
				Type:        model.ConfigTypeSync,
				AppID:       defaultAppID,
				OrgID:       defaultOrgID,
				System:      false,
				Data:        syncConfig,
				DateCreated: now,
			}
			_, err = configs.InsertOneWithContext(context, newSyncConfig)
			if err != nil {
				return err
			}
		}

		//3. wrap mg configs in new config model
		var mgConfigs []model.ManagedGroupConfigData
		mgFilter := bson.M{"client_id": "edu.illinois.rokwire"}
		err = managedGroupConfigs.FindWithContext(context, mgFilter, &mgConfigs, nil)
		if err != nil {
			return err
		}

		if len(mgConfigs) > 0 {
			_, err = managedGroupConfigs.DeleteManyWithContext(context, mgFilter, nil)
			if err != nil {
				return err
			}
		}

		for _, mgConfig := range mgConfigs {
			newManagedGroupConfig := model.Config{
				ID:          uuid.NewString(),
				Type:        model.ConfigTypeManagedGroup,
				AppID:       defaultAppID,
				OrgID:       defaultOrgID,
				System:      false,
				Data:        mgConfig,
				DateCreated: now,
			}
			_, err = configs.InsertOneWithContext(context, newManagedGroupConfig)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := m.performTransaction(transaction)
	if err != nil {
		return err
	}

	log.Println("configs migration passed")
	return nil
}

func (m *database) performTransaction(transaction func(TransactionContext) error) error {
	session, err := m.dbClient.StartSession()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionStart, "mongo session", nil, err)
	}
	context := context.Background()
	defer session.EndSession(context)

	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		return nil, transaction(sessionContext)
	}
	_, err = session.WithTransaction(context, callback)
	if err != nil {
		return errors.WrapErrorAction("performing", "transaction", nil, err)
	}

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
	}
}
