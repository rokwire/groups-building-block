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
	"groups/utils"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/syncmap"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database

	cachedConfigs *syncmap.Map
	configsLock   *sync.RWMutex
}

// Start starts the storage
func (sa *Adapter) Start(defaultAppID string, defaultOrgID string, defaultAppConfig *model.Config) error {
	err := sa.db.start(defaultAppID, defaultOrgID, defaultAppConfig)
	if err != nil {
		return err
	}

	//register storage listener
	sl := storageListener{adapter: sa}
	sa.RegisterStorageListener(&sl)

	err = sa.cacheConfigs()
	if err != nil {
		return errors.New("error caching sync configs")
	}

	return err
}

// RegisterStorageListener registers a data change listener with the storage adapter
func (sa *Adapter) RegisterStorageListener(storageListener Listener) {
	sa.db.listeners = append(sa.db.listeners, storageListener)
}

// cacheConfigs caches the configs from the DB
func (sa *Adapter) cacheConfigs() error {
	log.Println("cacheConfigs...")

	configs, err := sa.loadConfigs()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionLoad, model.TypeConfig, nil, err)
	}

	sa.setCachedConfigs(configs)

	return nil
}

func (sa *Adapter) setCachedConfigs(configs []model.Config) {
	sa.configsLock.Lock()
	defer sa.configsLock.Unlock()

	sa.cachedConfigs = &syncmap.Map{}

	for _, config := range configs {
		var err error
		switch config.Type {
		case model.ConfigTypeEnv:
			err = parseConfigsData[model.EnvConfigData](&config)
		case model.ConfigTypeSync:
			err = parseConfigsData[model.SyncConfigData](&config)
		case model.ConfigTypeManagedGroup:
			err = parseConfigsData[model.ManagedGroupConfigData](&config)
		case model.ConfigTypeApplication:
			err = parseConfigsData[model.ApplicationConfigData](&config)
		default:
			err = parseConfigsData[map[string]interface{}](&config)
		}
		if err != nil {
			log.Println(err.Error())
		}
		sa.cachedConfigs.Store(fmt.Sprintf("%s_%s_%s", config.Type, config.AppID, config.OrgID), config)
	}
}

func parseConfigsData[T model.ConfigData](config *model.Config) error {
	bsonBytes, err := bson.Marshal(config.Data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfig, nil, err)
	}

	var data T
	err = bson.Unmarshal(bsonBytes, &data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfigData, &logutils.FieldArgs{"type": config.Type}, err)
	}

	config.Data = data
	return nil
}

func (sa *Adapter) getCachedConfig(id string, configType string, appID string, orgID string) (*model.Config, error) {
	sa.configsLock.RLock()
	defer sa.configsLock.RUnlock()

	var item any
	var errArgs logutils.FieldArgs
	if id != "" {
		errArgs = logutils.FieldArgs{"id": id}
		item, _ = sa.cachedConfigs.Load(id)
	} else {
		errArgs = logutils.FieldArgs{"type": configType, "app_id": appID, "org_id": orgID}
		item, _ = sa.cachedConfigs.Load(fmt.Sprintf("%s_%s_%s", configType, appID, orgID))
	}

	if item != nil {
		config, ok := item.(model.Config)
		if !ok {
			return nil, errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &errArgs)
		}
		return &config, nil
	}
	return nil, nil
}

func (sa *Adapter) getCachedConfigs(configType *string, appID *string, orgID *string) ([]model.Config, error) {
	sa.configsLock.RLock()
	defer sa.configsLock.RUnlock()

	var err error
	configList := make([]model.Config, 0)
	sa.cachedConfigs.Range(func(key, item interface{}) bool {
		keyStr, ok := key.(string)
		if !ok || item == nil {
			return false
		}
		if !strings.Contains(keyStr, "_") {
			return true
		}

		config, ok := item.(model.Config)
		if !ok {
			err = errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &logutils.FieldArgs{"key": key})
			return false
		}

		match := true
		if configType != nil && !strings.HasPrefix(keyStr, fmt.Sprintf("%s_", *configType)) {
			match = false
		}
		if appID != nil && !strings.Contains(keyStr, fmt.Sprintf("_%s_", *appID)) {
			match = false
		}
		if orgID != nil && !strings.HasSuffix(keyStr, fmt.Sprintf("_%s", *orgID)) {
			match = false
		}

		if match {
			configList = append(configList, config)
		}

		return true
	})

	return configList, err
}

// loadConfigs loads configs
func (sa *Adapter) loadConfigs() ([]model.Config, error) {
	filter := bson.M{}

	var configs []model.Config
	err := sa.db.configs.FindWithContext(nil, filter, &configs, nil)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

// FindConfig finds the config for the specified type, appID, and orgID
func (sa *Adapter) FindConfig(configType string, appID string, orgID string) (*model.Config, error) {
	return sa.getCachedConfig("", configType, appID, orgID)
}

// FindConfigByID finds the config for the specified ID
func (sa *Adapter) FindConfigByID(id string) (*model.Config, error) {
	return sa.getCachedConfig(id, "", "", "")
}

// FindConfigs finds all configs for the specified type
func (sa *Adapter) FindConfigs(configType *string, appID *string, orgID *string) ([]model.Config, error) {
	return sa.getCachedConfigs(configType, appID, orgID)
}

// InsertConfig inserts a new config
func (sa *Adapter) InsertConfig(config model.Config) error {
	_, err := sa.db.configs.InsertOne(config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}

	return nil
}

// UpdateConfig updates an existing config
func (sa *Adapter) UpdateConfig(config model.Config) error {
	filter := bson.M{"_id": config.ID}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "type", Value: config.Type},
			primitive.E{Key: "app_id", Value: config.AppID},
			primitive.E{Key: "org_id", Value: config.OrgID},
			primitive.E{Key: "system", Value: config.System},
			primitive.E{Key: "data", Value: config.Data},
			primitive.E{Key: "date_updated", Value: config.DateUpdated},
		}},
	}
	_, err := sa.db.configs.UpdateOne(filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeConfig, &logutils.FieldArgs{"id": config.ID}, err)
	}

	return nil
}

// DeleteConfig deletes a configuration from storage
func (sa *Adapter) DeleteConfig(id string) error {
	delFilter := bson.M{"_id": id}
	_, err := sa.db.configs.DeleteMany(delFilter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, &logutils.FieldArgs{"id": id}, err)
	}

	return nil
}

// FindSyncTimes finds the sync times for the specified appID and orgID
func (sa *Adapter) FindSyncTimes(context TransactionContext, appID string, orgID string) (*model.SyncTimes, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID}

	var configs []model.SyncTimes
	err := sa.db.syncTimes.FindWithContext(context, filter, &configs, nil)
	if err != nil {
		return nil, err
	}
	if len(configs) != 1 {
		return nil, nil
	}

	return &configs[0], nil
}

// SaveSyncTimes saves the provided sync times fields
func (sa *Adapter) SaveSyncTimes(context TransactionContext, times model.SyncTimes) error {
	filter := bson.M{"app_id": times.AppID, "org_id": times.OrgID}

	upsert := true
	opts := options.ReplaceOptions{Upsert: &upsert}
	err := sa.db.syncTimes.ReplaceOne(filter, times, &opts)
	if err != nil {
		return err
	}

	return nil
}

// FindUser finds the user for the provided external id, app id, org id
func (sa *Adapter) FindUser(appID string, orgID string, id string, external bool) (*model.User, error) {
	var filter bson.D
	if external {
		filter = bson.D{primitive.E{Key: "app_id", Value: appID}, primitive.E{Key: "org_id", Value: orgID}, primitive.E{Key: "external_id", Value: id}}
	} else {
		filter = bson.D{primitive.E{Key: "app_id", Value: appID}, primitive.E{Key: "org_id", Value: orgID}, primitive.E{Key: "_id", Value: id}}
	}

	var result []*model.User
	err := sa.db.users.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil {
		//not found
		return nil, nil
	}
	return result[0], nil
}

// FindUsers finds all users for the provided list of (id | external id) and app id, org id
func (sa *Adapter) FindUsers(appID string, orgID string, id []string, external bool) ([]model.User, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID}, primitive.E{Key: "org_id", Value: orgID}}
	if external {
		filter = append(filter, primitive.E{Key: "external_id", Value: primitive.M{"$in": id}})
	} else {
		filter = append(filter, primitive.E{Key: "_id", Value: primitive.M{"$in": id}})
	}

	var result []model.User
	err := sa.db.users.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil {
		//not found
		return nil, nil
	}
	return result, nil
}

// LoginUser Login a user's and refactor legacy record if need
func (sa *Adapter) LoginUser(current *model.User) error {

	now := time.Now()

	//TODO: Do we still need this migration?
	//TODO: If so, handle group_memberships
	//TODO: NEED TO HANDLE ADMINS AND APP USERS SEPARATELY. CURRENTLY THIS IS MIGRATING ADMINS BACK AND FORTH
	// transaction
	err := sa.PerformTransaction(func(context TransactionContext) error {
		if current.IsCoreUser {
			// Repopulate and keep sync of external_id & user_id. Part 1
			filter := bson.D{
				primitive.E{Key: "app_id", Value: current.AppID},
				primitive.E{Key: "org_id", Value: current.OrgID},
				primitive.E{Key: "external_id", Value: current.ExternalID},
			}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "name", Value: current.Name},
					primitive.E{Key: "email", Value: current.Email},
					primitive.E{Key: "user_id", Value: current.ID},
					primitive.E{Key: "external_id", Value: current.ExternalID},
					primitive.E{Key: "net_id", Value: current.NetID},
					primitive.E{Key: "date_updated", Value: now},
				}},
			}
			_, err := sa.db.groupMemberships.UpdateManyWithContext(context, filter, update, nil)
			if err != nil {
				log.Printf("error updating dummy membership records for user(%s | %s) Part 1: %s", current.ID, current.ExternalID, err)
				return err
			}

			// Repopulate and keep sync of external_id & user_id. Part 2
			filter = bson.D{
				primitive.E{Key: "app_id", Value: current.AppID},
				primitive.E{Key: "org_id", Value: current.OrgID},
				primitive.E{Key: "user_id", Value: current.ID},
			}
			update = bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "name", Value: current.Name},
					primitive.E{Key: "email", Value: current.Email},
					primitive.E{Key: "user_id", Value: current.ID},
					primitive.E{Key: "external_id", Value: current.ExternalID},
					primitive.E{Key: "net_id", Value: current.NetID},
					primitive.E{Key: "date_updated", Value: now},
				}},
			}
			_, err = sa.db.groupMemberships.UpdateManyWithContext(context, filter, update, nil)
			if err != nil {
				log.Printf("error updating dummy membership records for user(%s | %s) Part 2: %s", current.ID, current.ExternalID, err)
				return err
			}

			// Repopulate and keep sync of user in the user table. Part 3
			filter = bson.D{
				primitive.E{Key: "app_id", Value: current.AppID},
				primitive.E{Key: "org_id", Value: current.OrgID},
				primitive.E{Key: "_id", Value: current.ID},
			}
			update = bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "is_core_user", Value: true},
					primitive.E{Key: "external_id", Value: current.ExternalID},
					primitive.E{Key: "name", Value: current.Name},
					primitive.E{Key: "email", Value: current.Email},
					primitive.E{Key: "net_id", Value: current.NetID},
					primitive.E{Key: "date_updated", Value: now},
				}},
			}
			_, err = sa.db.users.UpdateOneWithContext(context, filter, update, nil)
			if err != nil {
				log.Printf("error updating user(%s | %s) Part 3: %s", current.ID, current.ExternalID, err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return utils.NewServerError()
	}
	return nil
}

// CreateUser creates a new user
func (sa *Adapter) CreateUser(id string, appID string, orgID string, externalID string, email string, name string) (*model.User, error) {
	dateCreated := time.Now()
	user := model.User{ID: id, AppID: appID, OrgID: orgID, ExternalID: externalID, Email: email, Name: name, DateCreated: dateCreated}
	_, err := sa.db.users.InsertOne(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type getUserPostCountResult struct {
	Count int64 `json:"posts_count" bson:"posts_count"`
}

// GetUserPostCount gets the number of posts for the specified user
func (sa *Adapter) GetUserPostCount(appID string, orgID string, userID string) (*int64, error) {
	pipeline := []primitive.M{
		{"$match": primitive.M{
			"app_id":         appID,
			"org_id":         orgID,
			"member.user_id": userID,
		}},
		{"$count": "posts_count"},
	}
	var result []getUserPostCountResult
	err := sa.db.posts.Aggregate(pipeline, &result, &options.AggregateOptions{})
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return &result[0].Count, nil
	}
	return nil, nil
}

// DeleteUserWithContext Deletes a user
func (sa *Adapter) DeleteUserWithContext(context TransactionContext, appID string, orgID string, userID string) error {
	// delete the user
	filter := bson.D{
		primitive.E{Key: "_id", Value: userID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
	}
	_, err := sa.db.users.DeleteOneWithContext(context, filter, nil)
	if err != nil {
		log.Printf("error deleting user - %s", err.Error())
		return err
	}

	return nil
}

// CreateGroup creates a group. Returns the id of the created group
func (sa *Adapter) CreateGroup(current *model.User, group *model.Group, defaultMemberships []model.GroupMembership) (*string, *utils.GroupError) {
	insertedID := uuid.NewString()
	now := time.Now()

	//
	// [#301] Research Groups don't support automatic join feature!!!
	//
	if group.ResearchGroup && group.CanJoinAutomatically {
		group.CanJoinAutomatically = false
	}

	err := sa.PerformTransaction(func(context TransactionContext) error {

		//
		// Handle category and tags backward compatibility and legacy clients [#355]
		//
		if group.Category != "" && group.GetNewCategory() == nil {
			group.SetNewCategory(group.Category)
		}
		if len(group.Tags) > 0 && group.GetNewTags() == nil {
			group.SetNewTags(group.Tags)
		}
		if group.Attributes == nil {
			group.Attributes = map[string]interface{}{}
		}

		// insert the group and the admin member
		group.ID = insertedID
		group.DateCreated = now
		if group.Settings == nil {
			settings := model.DefaultGroupSettings()
			group.Settings = &settings
		}

		_, err := sa.db.groups.InsertOneWithContext(context, &group)
		if err != nil {
			return err
		}

		castedMemberships := []interface{}{}
		if len(defaultMemberships) > 0 {
			for _, membership := range defaultMemberships {
				membership.ID = uuid.NewString()
				membership.GroupID = insertedID
				membership.DateCreated = now
				castedMemberships = append(castedMemberships, membership)
			}
		} else if current != nil {
			castedMemberships = append(castedMemberships, model.GroupMembership{
				ID:          uuid.NewString(),
				GroupID:     insertedID,
				UserID:      current.ID,
				AppID:       current.AppID,
				OrgID:       current.OrgID,
				ExternalID:  current.ExternalID,
				Email:       current.Email,
				NetID:       current.NetID,
				Name:        current.Name,
				Status:      "admin",
				DateCreated: now,
			})
		}

		if len(castedMemberships) > 0 {
			_, err = sa.db.groupMemberships.InsertManyWithContext(context, castedMemberships, nil)
			if err != nil {
				return err
			}
		}

		err = sa.UpdateGroupStats(context, group.AppID, group.OrgID, group.ID, false, false, false, true)
		if err != nil {
			return err
		}

		sa.UpdateGroupAttributeIndexes(group)

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "title_unique") {
			return nil, utils.NewGroupDuplicationError()
		}
		return nil, utils.NewServerError()
	}

	return &insertedID, nil
}

// UpdateGroup updates a group along with the memberships
func (sa *Adapter) UpdateGroup(userID *string, group *model.Group, memberships []model.GroupMembership) *utils.GroupError {

	//
	// [#301] Research Groups don't support automatic join feature!!!
	//

	// transaction
	err := sa.PerformTransaction(func(context TransactionContext) error {
		if group.ResearchGroup && group.CanJoinAutomatically {
			group.CanJoinAutomatically = false
		}

		setOperation := bson.D{
			primitive.E{Key: "title", Value: group.Title},
			primitive.E{Key: "privacy", Value: group.Privacy},
			primitive.E{Key: "hidden_for_search", Value: group.HiddenForSearch},
			primitive.E{Key: "description", Value: group.Description},
			primitive.E{Key: "image_url", Value: group.ImageURL},
			primitive.E{Key: "web_url", Value: group.WebURL},
			primitive.E{Key: "membership_questions", Value: group.MembershipQuestions},
			primitive.E{Key: "date_updated", Value: time.Now()},
			primitive.E{Key: "authman_enabled", Value: group.AuthmanEnabled},
			primitive.E{Key: "authman_group", Value: group.AuthmanGroup},
			primitive.E{Key: "only_admins_can_create_polls", Value: group.OnlyAdminsCanCreatePolls},
			primitive.E{Key: "can_join_automatically", Value: group.CanJoinAutomatically},
			primitive.E{Key: "block_new_membership_requests", Value: group.BlockNewMembershipRequests},
			primitive.E{Key: "attendance_group", Value: group.AttendanceGroup},
			primitive.E{Key: "research_group", Value: group.ResearchGroup},
			primitive.E{Key: "research_open", Value: group.ResearchOpen},
			primitive.E{Key: "research_consent_statement", Value: group.ResearchConsentStatement},
			primitive.E{Key: "research_consent_details", Value: group.ResearchConsentDetails},
			primitive.E{Key: "research_description", Value: group.ResearchDescription},
			primitive.E{Key: "research_profile", Value: group.ResearchProfile},
		}
		if group.Settings != nil {
			setOperation = append(setOperation, primitive.E{Key: "settings", Value: group.Settings})
		}

		//
		// Handle category and tags backward compatibility and legacy clients 355355
		//
		if group.Attributes != nil {
			setOperation = append(setOperation, primitive.E{Key: "attributes", Value: group.Attributes})

			category := group.GetNewCategory()
			if category != nil {
				setOperation = append(setOperation, primitive.E{Key: "category", Value: *category})
			}

			tags := group.GetNewTags()
			if tags != nil {
				setOperation = append(setOperation, primitive.E{Key: "tags", Value: tags})
			}
		} else if group.Category != "" || len(group.Tags) > 0 {
			persistedGroup, err := sa.FindGroupWithContext(context, group.AppID, group.OrgID, group.ID, userID)
			if err != nil {
				return err
			}

			if group.Attributes == nil && persistedGroup.Attributes != nil {
				group.Attributes = persistedGroup.Attributes
			}

			if group.Category != "" {
				group.SetNewCategory(group.Category)
				setOperation = append(setOperation, primitive.E{Key: "category", Value: group.Category})
			}
			if len(group.Tags) > 0 {
				group.SetNewTags(group.Tags)
				setOperation = append(setOperation, primitive.E{Key: "tags", Value: group.Tags})
			}
			if group.Attributes == nil {
				group.Attributes = map[string]interface{}{}
			}
			setOperation = append(setOperation, primitive.E{Key: "attributes", Value: group.Attributes})
		}

		updateOperation := bson.D{
			primitive.E{Key: "$set", Value: setOperation},
		}

		_, err := sa.db.groups.UpdateOneWithContext(
			context,
			bson.D{primitive.E{Key: "_id", Value: group.ID},
				primitive.E{Key: "app_id", Value: group.AppID},
				primitive.E{Key: "org_id", Value: group.OrgID},
			}, updateOperation, nil)
		if err != nil {
			return err
		}

		if len(memberships) > 0 {
			for _, membership := range memberships {
				if membership.ID == "" {
					membership.ID = uuid.NewString()
					membership.DateCreated = time.Now()
					_, err = sa.db.groupMemberships.InsertOneWithContext(context, membership)
					if err != nil {
						return err
					}
				} else {
					filter := bson.D{
						primitive.E{Key: "_id", Value: group.ID},
						primitive.E{Key: "app_id", Value: group.AppID},
						primitive.E{Key: "org_id", Value: group.OrgID},
					}
					err = sa.db.groupMemberships.ReplaceOneWithContext(context, filter, membership, nil)
					if err != nil {
						return err
					}
				}
			}
		}

		err = sa.UpdateGroupStats(context, group.AppID, group.OrgID, group.ID, true, len(memberships) > 0, false, true)
		if err != nil {
			return err
		}

		sa.UpdateGroupAttributeIndexes(group)

		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "title_unique") {
			return utils.NewGroupDuplicationError()
		}
		return utils.NewServerError()
	}

	return nil
}

// DeleteGroup deletes a group.
func (sa *Adapter) DeleteGroup(appID string, orgID string, id string) error {
	err := sa.PerformTransaction(func(context TransactionContext) error {

		// 1. delete mapped group events
		_, err := sa.db.events.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}, nil)
		if err != nil {
			return err
		}

		// 2. delete mapped group posts
		_, err = sa.db.posts.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}, nil)
		if err != nil {
			return err
		}

		// 3. delete mapped group memberships
		_, err = sa.db.groupMemberships.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}, nil)
		if err != nil {
			return err
		}

		// 4. delete the group
		_, err = sa.db.groups.DeleteOneWithContext(context, bson.D{
			primitive.E{Key: "_id", Value: id},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// FindGroupWithContext finds group by id and app id, org id with context
func (sa *Adapter) FindGroupWithContext(context TransactionContext, appID string, orgID string, groupID string, userID *string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: groupID}, primitive.E{Key: "app_id", Value: appID}, primitive.E{Key: "org_id", Value: orgID}}

	var err error
	var membership *model.GroupMembership
	if userID != nil {
		// find group memberships
		membership, err = sa.FindGroupMembershipWithContext(context, appID, orgID, groupID, *userID)
	}

	var rec model.Group
	err = sa.db.groups.FindOneWithContext(context, filter, &rec, nil)
	if err != nil {
		return nil, err
	}

	rec.CurrentMember = membership

	return &rec, nil
}

// FindGroupByTitle finds group by membership
func (sa *Adapter) FindGroupByTitle(appID string, orgID string, title string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "title", Value: title},
	}
	var result []model.Group
	err := sa.db.groups.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, nil
	}

	return &result[0], nil
}

// FindGroups finds groups
func (sa *Adapter) FindGroups(userID *string, groupsFilter model.GroupsFilter) ([]model.Group, error) {
	// TODO: Merge the filter logic in a common method (FindGroups, FindGroupsV3, FindUserGroups)

	var err error
	groupIDs := []string{}
	var memberships model.MembershipCollection
	if userID != nil {
		// find group memberships
		memberships, err = sa.FindGroupMembershipsWithContext(nil, model.MembershipFilter{AppID: groupsFilter.AppID, OrgID: groupsFilter.OrgID, UserID: userID})
		if err != nil {
			return nil, err
		}

		for _, membership := range memberships.Items {
			groupIDs = append(groupIDs, membership.GroupID)
		}
	}

	filter := bson.D{primitive.E{Key: "app_id", Value: groupsFilter.AppID}, primitive.E{Key: "org_id", Value: groupsFilter.OrgID}}
	if userID != nil {
		innerOrFilter := []bson.M{}

		if groupsFilter.ExcludeMyGroups != nil && *groupsFilter.ExcludeMyGroups {
			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$nin": groupIDs}})
			innerOrFilter = []bson.M{
				{"privacy": bson.M{"$ne": "private"}},
			}
		} else {
			innerOrFilter = []bson.M{
				{"_id": bson.M{"$in": groupIDs}},
				{"privacy": bson.M{"$ne": "private"}},
			}
		}

		if groupsFilter.Title != nil {
			if groupsFilter.IncludeHidden != nil && *groupsFilter.IncludeHidden {
				innerOrFilter = append(innerOrFilter, primitive.M{"$and": []primitive.M{
					{"title": *groupsFilter.Title},
				}})
			} else {
				innerOrFilter = append(innerOrFilter, primitive.M{"$and": []primitive.M{
					{"title": *groupsFilter.Title},
					{"$or": []primitive.M{
						{"hidden_for_search": false},
						{"hidden_for_search": primitive.M{"$exists": false}},
					}},
				}})
			}
		}

		orFilter := primitive.E{Key: "$or", Value: innerOrFilter}

		filter = append(filter, orFilter)
	}

	if groupsFilter.Hidden != nil {
		if *groupsFilter.Hidden {
			filter = append(filter, primitive.E{Key: "hidden_for_search", Value: groupsFilter.Hidden})
		} else {
			filter = append(filter, primitive.E{Key: "hidden_for_search", Value: primitive.M{"$ne": true}})
		}
	}

	if groupsFilter.Category != nil {
		filter = append(filter, primitive.E{Key: "category", Value: groupsFilter.Category})
	}
	if len(groupsFilter.Tags) > 0 {
		filter = append(filter, primitive.E{Key: "tags", Value: bson.M{"$in": groupsFilter.Tags}})
	}
	if groupsFilter.Title != nil {
		filter = append(filter, primitive.E{Key: "title", Value: primitive.Regex{Pattern: *groupsFilter.Title, Options: "i"}})
	}
	if groupsFilter.Privacy != nil {
		filter = append(filter, primitive.E{Key: "privacy", Value: groupsFilter.Privacy})
	}
	if groupsFilter.ResearchOpen != nil {
		if *groupsFilter.ResearchOpen {
			filter = append(filter, primitive.E{Key: "research_open", Value: true})
		} else {
			filter = append(filter, primitive.E{Key: "research_open", Value: primitive.M{"$ne": true}})
		}
	}
	if groupsFilter.ResearchGroup {
		filter = append(filter, primitive.E{Key: "research_group", Value: true})
	} else {
		filter = append(filter, primitive.E{Key: "research_group", Value: primitive.M{"$ne": true}})
	}
	if groupsFilter.ResearchAnswers != nil {
		for outerKey, outerValue := range groupsFilter.ResearchAnswers {
			for innerKey, innerValue := range outerValue {
				filter = append(filter, bson.E{
					Key: "$or", Value: []bson.M{
						{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$elemMatch": bson.M{"$in": innerValue}}},
						{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$exists": false}},
					},
				})
			}
		}
	}

	if groupsFilter.Attributes != nil {
		attributeFilters := []bson.M{}
		for key, value := range groupsFilter.Attributes {
			if reflect.TypeOf(value).Kind() != reflect.Slice {
				attributeFilters = append(attributeFilters, bson.M{fmt.Sprintf("attributes.%s", key): value})
			} else {
				orSubCriterias := []bson.M{}
				var entryList []interface{} = value.([]interface{})
				for _, entry := range entryList {
					orSubCriterias = append(orSubCriterias, bson.M{fmt.Sprintf("attributes.%s", key): entry})
				}
				attributeFilters = append(attributeFilters, bson.M{"$or": orSubCriterias})
			}
		}
		if len(attributeFilters) > 0 {
			filter = append(filter, bson.E{
				Key: "$and", Value: attributeFilters,
			})
		}
	}

	findOptions := options.Find()
	if groupsFilter.Order != nil && "desc" == *groupsFilter.Order {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: -1},
		})
	} else {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: 1},
		})
	}
	if groupsFilter.Limit != nil {
		findOptions.SetLimit(*groupsFilter.Limit)
	}
	if groupsFilter.Offset != nil {
		findOptions.SetSkip(*groupsFilter.Offset)
	}

	var list []model.Group
	err = sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	if userID != nil {
		for index, group := range list {
			group.CurrentMember = memberships.GetMembershipBy(func(membership model.GroupMembership) bool {
				return membership.GroupID == group.ID
			})
			if group.CurrentMember != nil {
				list[index] = group
			}
		}
	}

	return list, nil
}

type findUserGroupsCountResult struct {
	Count int64 `bson:"count"`
}

// FindUserGroupsCount retrieves the count of current groups that the user is member
func (sa *Adapter) FindUserGroupsCount(appID string, orgID string, userID string) (*int64, error) {
	pipeline := []primitive.M{
		{"$match": primitive.M{
			"app_id":          appID,
			"org_id":          orgID,
			"members.user_id": userID,
		}},
		{"$count": "count"},
	}
	var result []findUserGroupsCountResult
	err := sa.db.groups.Aggregate(pipeline, &result, &options.AggregateOptions{})
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return &result[0].Count, nil
	}
	return nil, nil
}

// FindUserGroups finds the user groups for the filter
func (sa *Adapter) FindUserGroups(userID string, groupsFilter model.GroupsFilter) ([]model.Group, error) {
	// TODO: Merge the filter logic in a common method (FindGroups, FindGroupsV3, FindUserGroups)

	// find group memberships
	memberships, err := sa.FindGroupMembershipsWithContext(nil, model.MembershipFilter{AppID: groupsFilter.AppID, OrgID: groupsFilter.OrgID, UserID: &userID})
	if err != nil {
		return nil, err
	}
	groupIDs := []string{}
	for _, membership := range memberships.Items {
		groupIDs = append(groupIDs, membership.GroupID)
	}

	mongoFilter := bson.M{
		"_id":    bson.M{"$in": groupIDs},
		"app_id": groupsFilter.AppID,
		"org_id": groupsFilter.OrgID,
	}

	if groupsFilter.Category != nil {
		mongoFilter["category"] = *groupsFilter.Category
	}
	if groupsFilter.Title != nil {
		mongoFilter["title"] = primitive.Regex{Pattern: *groupsFilter.Title, Options: "i"}
	}
	if groupsFilter.Privacy != nil {
		mongoFilter["privacy"] = groupsFilter.Privacy
	}
	if groupsFilter.ResearchOpen != nil {
		if *groupsFilter.ResearchOpen {
			mongoFilter["research_open"] = true
		} else {
			mongoFilter["research_open"] = primitive.M{"$ne": true}
		}
	}
	if groupsFilter.ResearchGroup {
		mongoFilter["research_group"] = true
	} else {
		mongoFilter["research_group"] = bson.M{"$ne": true}
	}
	if groupsFilter.ResearchAnswers != nil {
		for outerKey, outerValue := range groupsFilter.ResearchAnswers {
			for innerKey, innerValue := range outerValue {
				mongoFilter["$or"] = []bson.M{
					{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$elemMatch": bson.M{"$in": innerValue}}},
					{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$exists": false}},
				}
			}
		}
	}

	if groupsFilter.Attributes != nil {
		attributeFilters := []bson.M{}
		for key, value := range groupsFilter.Attributes {
			if reflect.TypeOf(value).Kind() != reflect.Slice {
				attributeFilters = append(attributeFilters, bson.M{fmt.Sprintf("attributes.%s", key): value})
			} else {
				orSubCriterias := []bson.M{}
				var entryList []interface{} = value.([]interface{})
				for _, entry := range entryList {
					orSubCriterias = append(orSubCriterias, bson.M{fmt.Sprintf("attributes.%s", key): entry})
				}
				attributeFilters = append(attributeFilters, bson.M{"$or": orSubCriterias})
			}
		}
		if len(attributeFilters) > 0 {
			mongoFilter["$and"] = attributeFilters
		}
	}

	findOptions := options.Find()
	if groupsFilter.Order != nil && "desc" == *groupsFilter.Order {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: -1},
		})
	} else {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: 1},
		})
	}
	if groupsFilter.Limit != nil {
		findOptions.SetLimit(*groupsFilter.Limit)
	}
	if groupsFilter.Offset != nil {
		findOptions.SetSkip(*groupsFilter.Offset)
	}

	var list []model.Group
	err = sa.db.groups.Find(mongoFilter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	for index, group := range list {
		group.CurrentMember = memberships.GetMembershipBy(func(membership model.GroupMembership) bool {
			return membership.GroupID == group.ID
		})
		if group.CurrentMember != nil {
			list[index] = group
		}
	}

	return list, nil
}

// FindEvents finds the events for a group
func (sa *Adapter) FindEvents(current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	filter := bson.D{
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "app_id", Value: current.AppID},
		primitive.E{Key: "org_id", Value: current.OrgID},
	}
	if filterByToMembers && current != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: []primitive.M{
			{"to_members": primitive.Null{}},
			{"to_members": primitive.M{"$exists": true, "$size": 0}},
			{"to_members.user_id": current.ID},
			{"member.user_id": current.ID},
		}})
	}

	var result []model.Event
	err := sa.db.events.Find(filter, &result, nil)
	return result, err
}

// CreateEvent creates a group event
func (sa *Adapter) CreateEvent(appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	event := model.Event{
		AppID:         appID,
		OrgID:         orgID,
		EventID:       eventID,
		GroupID:       groupID,
		DateCreated:   time.Now().UTC(),
		ToMembersList: toMemberList,
		Creator:       creator,
	}

	err := sa.PerformTransaction(func(context TransactionContext) error {
		_, err := sa.db.events.InsertOne(event)
		if err != nil {
			return err
		}

		return sa.UpdateGroupStats(context, appID, orgID, groupID, true, false, false, false)
	})

	return &event, err
}

// UpdateEventWithContext updates a group event
func (sa *Adapter) UpdateEventWithContext(context TransactionContext, appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember) error {
	filter := bson.D{
		primitive.E{Key: "event_id", Value: eventID},
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
	}
	change := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now()},
			primitive.E{Key: "to_members", Value: toMemberList},
		}},
	}
	res, err := sa.db.events.UpdateOneWithContext(context, filter, change, nil)
	if err == nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.ErrorAction(logutils.ActionUpdate, "event", &logutils.FieldArgs{"event_id": eventID, "modified": res.ModifiedCount, "expected": 1})
	}

	return nil
}

// DeleteEventWithContext deletes a group event
func (sa *Adapter) DeleteEventWithContext(context TransactionContext, appID string, orgID string, eventID string, groupID string) error {
	filter := bson.D{primitive.E{Key: "event_id", Value: eventID},
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID}}
	result, err := sa.db.events.DeleteOneWithContext(context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("result is nil for event with event id " + eventID)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return errors.New("error occured while deleting an event with event id " + eventID)
	}

	return nil
}

// FindPosts Retrieves posts for a group
func (sa *Adapter) FindPosts(current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {

	var userID *string
	if current != nil {
		userID = &current.ID
	}

	group, errGr := sa.FindGroupWithContext(nil, current.AppID, current.OrgID, groupID, userID)
	if group == nil {
		if errGr != nil {
			log.Printf("unable to find group with id %s: %s", groupID, errGr)
		} else {
			log.Printf("group does not exists %s", groupID)
		}
		return nil, errGr
	}

	filter := bson.D{
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "app_id", Value: group.AppID},
		primitive.E{Key: "org_id", Value: group.OrgID},
	}

	if filterByToMembers {
		innerFilter := []primitive.M{
			{"to_members": primitive.Null{}},
			{"to_members": primitive.M{"$exists": true, "$size": 0}},
		}
		if current != nil {
			innerFilter = append(innerFilter, []primitive.M{
				{"to_members.user_id": current.ID},
				{"member.user_id": current.ID},
			}...)
		}
		filter = append(filter, primitive.E{Key: "$or", Value: innerFilter})
	}

	if filterPrivatePostsValue != nil {
		filter = append(filter, primitive.E{Key: "private", Value: *filterPrivatePostsValue})
	}

	paging := false
	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: -1}})
	} else {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: 1}})
	}
	if limit != nil {
		findOptions.SetLimit(*limit)
		paging = true
	}
	if offset != nil {
		findOptions.SetSkip(*offset)
		paging = true
	}

	if paging {
		filter = append(filter, primitive.E{Key: "parent_id", Value: nil})
	}

	var list []*model.Post
	err := sa.db.posts.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	if paging && len(list) > 0 {
		for _, post := range list {
			childPosts, err := sa.FindPostsByTopParentID(current, groupID, post.ID, true, order)
			if err == nil && childPosts != nil {
				for _, childPost := range childPosts {
					if childPost.UserCanSeePost(current.ID) {
						list = append(list, childPost)
					}
				}
			}
		}
	}

	var resultList = make([]*model.Post, 0)
	var postMapping = make(map[string]*model.Post)

	if list != nil {
		for i := range list {
			postID := list[i].ID
			list[i].Replies = make([]*model.Post, 0)
			postMapping[postID] = list[i]
		}
		for _, post := range list {
			if post != nil {
				if post.ParentID != nil {
					if parentPost, ok := postMapping[*post.ParentID]; ok && parentPost != nil {
						parentPost.Replies = append(parentPost.Replies, post)
					}
				} else {
					resultList = append(resultList, post)
				}
			}
		}
	}

	return resultList, nil
}

// FindAllUserPosts Retrieves all user posts across all existing groups
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindAllUserPosts(context TransactionContext, appID string, orgID string, userID string) ([]model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "member.user_id", Value: userID},
	}

	var posts []model.Post
	err := sa.db.posts.FindWithContext(context, filter, &posts, nil)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// FindPost Retrieves a post by groupID and postID
func (sa *Adapter) FindPost(context TransactionContext, appID string, orgID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "_id", Value: postID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
	}

	if filterByToMembers {
		innerFilter := []primitive.M{
			{"to_members": primitive.Null{}},
			{"to_members": primitive.M{"$exists": true, "$size": 0}},
		}
		if userID != nil {
			innerFilter = append(innerFilter, []primitive.M{
				{"to_members.user_id": *userID},
				{"member.user_id": *userID},
			}...)
		}
		filter = append(filter, primitive.E{Key: "$or", Value: innerFilter})
	}

	if !skipMembershipCheck && userID != nil {
		membership, err := sa.FindGroupMembershipWithContext(context, appID, orgID, groupID, *userID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	var post *model.Post
	err := sa.db.posts.FindOneWithContext(context, filter, &post, nil)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// FindTopPostByParentID Finds the top post by parent id
func (sa *Adapter) FindTopPostByParentID(current *model.User, groupID string, parentID string, skipMembershipCheck bool) (*model.Post, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: current.AppID}, primitive.E{Key: "org_id", Value: current.OrgID}, primitive.E{Key: "_id", Value: parentID}}

	if !skipMembershipCheck {
		membership, err := sa.FindGroupMembershipWithContext(nil, current.AppID, current.OrgID, groupID, current.ID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	var post *model.Post
	err := sa.db.posts.FindOne(filter, &post, nil)
	if err != nil {
		return nil, err
	}

	if post.ParentID != nil {
		return sa.FindTopPostByParentID(current, groupID, *post.ParentID, skipMembershipCheck)
	}

	return post, nil
}

// FindPostsByParentID FindPostByParentID Retrieves a post by groupID and postID
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindPostsByParentID(ctx TransactionContext, appID string, orgID string, userID string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]*model.Post, error) {

	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "parent_id", Value: parentID},
	}

	if !skipMembershipCheck {
		membership, err := sa.FindGroupMembershipWithContext(ctx, appID, orgID, groupID, userID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: -1}})
	} else {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: 1}})
	}

	var posts []*model.Post
	err := sa.db.posts.Find(filter, &posts, findOptions)
	if err != nil {
		return nil, err
	}

	if recursive {
		for _, post := range posts {
			childPosts, err := sa.FindPostsByParentID(ctx, appID, orgID, userID, groupID, post.ID, true, filterByToMembers, recursive, order)
			if err == nil && childPosts != nil {
				for _, childPost := range childPosts {
					posts = append(posts, childPost)
				}
			}
		}
	}

	return posts, nil
}

// FindPostsByTopParentID  Retrieves a post by groupID and top parent id
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindPostsByTopParentID(current *model.User, groupID string, topParentID string, skipMembershipCheck bool, order *string) ([]*model.Post, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: current.AppID}, primitive.E{Key: "org_id", Value: current.OrgID}, primitive.E{Key: "top_parent_id", Value: topParentID}}

	if !skipMembershipCheck {
		membership, err := sa.FindGroupMembershipWithContext(nil, current.AppID, current.OrgID, groupID, current.ID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: -1}})
	} else {
		findOptions.SetSort(bson.D{{Key: "date_created", Value: 1}})
	}

	var posts []*model.Post
	err := sa.db.posts.Find(filter, &posts, findOptions)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// CreatePost Creates a post
func (sa *Adapter) CreatePost(context TransactionContext, post *model.Post) error {
	_, err := sa.db.posts.InsertOneWithContext(context, post)
	if err != nil {
		return fmt.Errorf("error inserting post: %v", err)
	}

	return nil
}

// UpdatePost Updates a post
func (sa *Adapter) UpdatePost(context TransactionContext, userID string, post *model.Post) error {
	if post == nil {
		return errors.New("post is missing")
	}

	filter := bson.D{primitive.E{Key: "app_id", Value: post.AppID}, primitive.E{Key: "org_id", Value: post.OrgID}, primitive.E{Key: "_id", Value: post.ID}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "subject", Value: post.Subject},
		primitive.E{Key: "body", Value: post.Body},
		primitive.E{Key: "private", Value: post.Private},
		primitive.E{Key: "use_as_notification", Value: post.UseAsNotification},
		primitive.E{Key: "is_abuse", Value: post.IsAbuse},
		primitive.E{Key: "image_url", Value: post.ImageURL},
		primitive.E{Key: "date_updated", Value: post.DateUpdated},
		primitive.E{Key: "to_members", Value: post.ToMembersList},
	}}}

	res, err := sa.db.posts.UpdateOneWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.ErrorAction(logutils.ActionUpdate, "post", &logutils.FieldArgs{"modified": res.ModifiedCount, "expected": 1})
	}

	return nil
}

// ReportPostAsAbuse Report post as abuse
func (sa *Adapter) ReportPostAsAbuse(post *model.Post) error {
	if post != nil {
		filter := bson.D{primitive.E{Key: "_id", Value: post.ID}, primitive.E{Key: "app_id", Value: post.AppID}, primitive.E{Key: "org_id", Value: post.OrgID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "is_abuse", Value: true},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		_, err := sa.db.posts.UpdateOne(filter, update, nil)

		return err
	}
	return nil
}

// ReactToPost React to a post
func (sa *Adapter) ReactToPost(context TransactionContext, userID string, postID string, reaction string, on bool) error {
	filter := bson.D{primitive.E{Key: "_id", Value: postID}}

	updateOperation := "$pull"
	if on {
		updateOperation = "$push"
	}
	update := bson.D{
		primitive.E{Key: updateOperation, Value: bson.D{
			primitive.E{Key: "reactions." + reaction, Value: userID},
		}},
	}

	res, err := sa.db.posts.UpdateOneWithContext(context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "post", &logutils.FieldArgs{"id": postID, "user_id": userID, "reaction": reaction}, err)
	}
	if res.ModifiedCount != 1 {
		return errors.ErrorAction(logutils.ActionUpdate, "post", &logutils.FieldArgs{"modified": res.ModifiedCount, "expected": 1})
	}

	return nil
}

// DeletePost Deletes a post
func (sa *Adapter) DeletePost(ctx TransactionContext, appID string, orgID string, userID string, groupID string, postID string, force bool) error {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID}, primitive.E{Key: "org_id", Value: orgID}, primitive.E{Key: "_id", Value: postID}}

	res, err := sa.db.posts.DeleteOneWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount != 1 {
		return fmt.Errorf("unexpected deleted post count: %d", res.DeletedCount)
	}

	return nil
}

// UpdateGroupStats set the updated date to the current date time (now)
func (sa *Adapter) UpdateGroupStats(context TransactionContext, appID string, orgID string, id string, resetUpdateDate, resetMembershipUpdateDate, resetManagedMembershipUpdateDate, resetStats bool) error {

	updateStats := func(ctx TransactionContext) error {
		innerUpdate := bson.D{}

		if resetStats {
			stats, err := sa.GetGroupMembershipStats(ctx, appID, orgID, id)
			if err != nil {
				return err
			}
			if stats != nil {
				innerUpdate = append(innerUpdate, primitive.E{Key: "stats", Value: stats})
			}
		}

		if resetUpdateDate {
			innerUpdate = append(innerUpdate, primitive.E{Key: "date_updated", Value: time.Now()})
		}
		if resetMembershipUpdateDate {
			innerUpdate = append(innerUpdate, primitive.E{Key: "date_membership_updated", Value: time.Now()})
		}
		if resetManagedMembershipUpdateDate {
			innerUpdate = append(innerUpdate, primitive.E{Key: "date_managed_membership_updated", Value: time.Now()})
		}

		// update the group
		filter := bson.D{
			primitive.E{Key: "_id", Value: id},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}
		update := bson.D{
			primitive.E{Key: "$set", Value: innerUpdate},
		}

		_, err := sa.db.groups.UpdateOneWithContext(ctx, filter, update, nil)
		return err
	}

	if context != nil {
		return updateStats(context)
	}
	return sa.PerformTransaction(func(context TransactionContext) error {
		return updateStats(context)
	})
}

// UpdateGroupAttributeIndexes Analyses and updates the indexes if need. This method is async  without transaction.
func (sa *Adapter) UpdateGroupAttributeIndexes(group *model.Group) {
	if group != nil {
		updateIndexes := func() {

			indexes, err := sa.db.groups.ListIndexesWithContext(context.Background())
			if err != nil {
				log.Printf("sa.UpdateGroupAttributeIndexes error on retrieving indexes: %s", err)
				return
			}
			for key := range group.Attributes {
				fieldName := fmt.Sprintf("attributes.%s", key)

				found := false
				for _, index := range indexes {
					indexName := index["name"].(string)

					if strings.Contains(indexName, fieldName) {
						found = true
						break
					}
				}

				if !found {
					err := sa.db.groups.AddIndexWithContext(
						context.Background(),
						bson.D{
							primitive.E{Key: fieldName, Value: 1},
						}, false)
					if err != nil {
						log.Printf("sa.UpdateGroupAttributeIndexes error on retrieving indexes: %s", err)
						return
					}
				}
			}
		}

		go updateIndexes()
	}
}

// UpdateGroupDateUpdated Updates group's date updated
func (sa *Adapter) UpdateGroupDateUpdated(appID string, orgID string, groupID string) error {
	filter := bson.D{
		primitive.E{Key: "_id", Value: groupID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
	}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now()},
		}},
	}

	_, err := sa.db.groups.UpdateOne(filter, update, nil)
	return err
}

// FindAuthmanGroups finds all groups that are associated with Authman
func (sa *Adapter) FindAuthmanGroups(appID string, orgID string) ([]model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "authman_enabled", Value: true},
	}

	var list []model.Group
	err := sa.db.groups.Find(filter, &list, nil)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindAuthmanGroupByKey Finds an Authman group by group long name
func (sa *Adapter) FindAuthmanGroupByKey(appID string, orgID string, authmanGroupKey string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "authman_group", Value: authmanGroupKey},
	}

	var list []model.Group
	err := sa.db.groups.Find(filter, &list, nil)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(context TransactionContext) error) error {
	// transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		err := transaction(sessionContext)
		if err != nil {
			if wrappedErr, ok := err.(*errors.Error); ok && wrappedErr.Internal() != nil {
				return nil, wrappedErr.Internal()
			}
			return nil, err
		}

		return nil, nil
	}

	session, err := sa.db.dbClient.StartSession()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionStart, "mongo session", nil, err)
	}
	context := context.Background()
	defer session.EndSession(context)

	_, err = session.WithTransaction(context, callback)
	if err != nil {
		return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
	}
	return nil
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS}

	cachedConfigs := &syncmap.Map{}
	configsLock := &sync.RWMutex{}
	return &Adapter{db: db, cachedConfigs: cachedConfigs, configsLock: configsLock}
}

type storageListener struct {
	adapter *Adapter
	DefaultListenerImpl
}

func (sl *storageListener) OnConfigsChanged() {
	sl.adapter.cacheConfigs()
}

// Listener  listens for change data storage events
type Listener interface {
	OnConfigsChanged()
}

// DefaultListenerImpl default listener implementation
type DefaultListenerImpl struct{}

// OnConfigsChanged notifies configs have been updated
func (d *DefaultListenerImpl) OnConfigsChanged() {}

// TransactionContext wraps mongo.SessionContext for use by external packages
type TransactionContext interface {
	mongo.SessionContext
}
