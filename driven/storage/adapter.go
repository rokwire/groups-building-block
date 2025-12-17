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
	"errors"
	"fmt"
	"groups/core/model"
	"groups/utils"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/syncmap"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database

	cachedSyncConfigs *syncmap.Map
	syncConfigsLock   *sync.RWMutex

	cachedManagedGroupConfigs *syncmap.Map
	managedGroupConfigsLock   *sync.RWMutex
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	if err != nil {
		return err
	}

	//register storage listener
	sl := storageListener{adapter: sa}
	sa.RegisterStorageListener(&sl)

	err = sa.cacheSyncConfigs()
	if err != nil {
		return errors.New("error caching sync configs")
	}

	err = sa.cacheManagedGroupConfigs()
	if err != nil {
		return errors.New("error caching managed group configs")
	}

	return err
}

// RegisterStorageListener registers a data change listener with the storage adapter
func (sa *Adapter) RegisterStorageListener(storageListener Listener) {
	sa.db.listeners = append(sa.db.listeners, storageListener)
}

// cacheSyncConfigs caches the sync configs from the DB
func (sa *Adapter) cacheSyncConfigs() error {
	log.Println("cacheSyncConfigs..")

	configs, err := sa.LoadSyncConfigs(nil)
	if err != nil {
		return err
	}

	sa.setCachedSyncConfigs(&configs)

	return nil
}

func (sa *Adapter) setCachedSyncConfigs(configs *[]model.SyncConfig) {
	sa.syncConfigsLock.Lock()
	defer sa.syncConfigsLock.Unlock()

	sa.cachedSyncConfigs = &syncmap.Map{}
	for _, config := range *configs {
		sa.cachedSyncConfigs.Store(config.ClientID, config)
	}
}

func (sa *Adapter) getCachedSyncConfig(clientID string) (*model.SyncConfig, error) {
	sa.syncConfigsLock.RLock()
	defer sa.syncConfigsLock.RUnlock()

	item, _ := sa.cachedSyncConfigs.Load(clientID)
	if item != nil {
		config, ok := item.(model.SyncConfig)
		if !ok {
			return nil, fmt.Errorf("missing managed group config for clientID: %s", clientID)
		}
		return &config, nil
	}
	return nil, nil
}

func (sa *Adapter) getCachedSyncConfigs() ([]model.SyncConfig, error) {
	sa.syncConfigsLock.RLock()
	defer sa.syncConfigsLock.RUnlock()

	var err error
	configList := make([]model.SyncConfig, 0)
	sa.cachedSyncConfigs.Range(func(key, item interface{}) bool {
		if item == nil {
			return false
		}

		config, ok := item.(model.SyncConfig)
		if !ok {
			err = fmt.Errorf("error casting config with client id: %s", key)
			return false
		}
		configList = append(configList, config)
		return true
	})

	return configList, err
}

// LoadSyncConfigs loads all sync configs
func (sa *Adapter) LoadSyncConfigs(context TransactionContext) ([]model.SyncConfig, error) {
	filter := bson.M{"type": "sync"}

	var config []model.SyncConfig
	err := sa.db.configs.FindWithContext(context, filter, &config, nil)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// FindSyncConfig finds the sync config for the specified clientID
func (sa *Adapter) FindSyncConfig(context TransactionContext, clientID string) (*model.SyncConfig, error) {
	return sa.getCachedSyncConfig(clientID)
}

// FindSyncConfigs finds all sync configs
func (sa *Adapter) FindSyncConfigs(context TransactionContext) ([]model.SyncConfig, error) {
	return sa.getCachedSyncConfigs()
}

// SaveSyncConfig saves the provided sync config fields
func (sa *Adapter) SaveSyncConfig(context TransactionContext, config model.SyncConfig) error {
	filter := bson.M{"type": "sync", "client_id": config.ClientID}

	config.Type = "sync"

	upsert := true
	opts := options.ReplaceOptions{Upsert: &upsert}
	err := sa.db.configs.ReplaceOneWithContext(context, filter, config, &opts)
	if err != nil {
		return err
	}

	return nil
}

// FindSyncTimes finds the sync times for the specified clientID
func (sa *Adapter) FindSyncTimes(context TransactionContext, clientID string, key string, legacy bool) (*model.SyncTimes, error) {

	// TBD remove client_id

	filter := bson.M{}
	if legacy {
		filter["client_id"] = clientID
	} else {
		filter["key"] = key
	}

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
	filter := bson.M{"key": times.Key}

	upsert := true
	opts := options.ReplaceOptions{Upsert: &upsert}
	err := sa.db.syncTimes.ReplaceOneWithContext(context, filter, times, &opts)
	if err != nil {
		return err
	}

	return nil
}

type getUserPostCountResult struct {
	Count int64 `json:"posts_count" bson:"posts_count"`
}

// GetUserPostCount gets the number of posts for the specified user
func (sa *Adapter) GetUserPostCount(clientID string, userID string) (*int64, error) {
	pipeline := []primitive.M{
		primitive.M{"$match": primitive.M{
			"client_id":      clientID,
			"member.user_id": userID,
		}},
		primitive.M{"$count": "posts_count"},
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

// DeleteUser Deletes a user with all information
func (sa *Adapter) DeleteUser(clientID string, userID string) error {

	return sa.PerformTransaction(func(sessionContext TransactionContext) error {
		posts, err := sa.FindAllUserPosts(sessionContext, clientID, userID)
		if err != nil {
			log.Printf("error on find all posts for user (%s) - %s", userID, err.Error())
			return err
		}
		if len(posts) > 0 {
			for _, post := range posts {
				err = sa.DeletePost(sessionContext, clientID, userID, post.GroupID, post.ID, true)
				if err != nil {
					log.Printf("error on delete all posts for user (%s) - %s", userID, err.Error())
					return err
				}
			}
		}

		memberships, err := sa.FindUserGroupMembershipsWithContext(sessionContext, clientID, userID)
		if err != nil {
			log.Printf("error getting user memberships - %s", err.Error())
			return err
		}
		for _, membership := range memberships.Items {

			err = sa.DeleteMembershipWithContext(sessionContext, clientID, membership.GroupID, membership.UserID, true)
			if err != nil {
				log.Printf("error deleting user membership - %s", err.Error())
				return err
			}
		}

		return nil
	})
}

// CreateGroup creates a group. Returns the id of the created group
func (sa *Adapter) CreateGroup(context TransactionContext, clientID string, current *model.User, group *model.Group, defaultMemberships []model.GroupMembership) (*string, *utils.GroupError) {
	insertedID := uuid.NewString()
	now := time.Now()

	var err error
	wrapperFunc := func(context TransactionContext) error {

		// Check the title is unique. Don't rely on the unique index.
		if err := sa.checkUniqueGroupTitleWithContext(context, clientID, nil, group.Title); err != nil {
			return err
		}

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

		if group.Administrative == nil {
			falseValue := false
			group.Administrative = &falseValue
		}

		// insert the group and the admin member
		group.ID = insertedID
		group.ClientID = clientID
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
		contaignCurrentUser := false
		if len(defaultMemberships) > 0 {
			for _, membership := range defaultMemberships {
				if membership.ExternalID == current.ExternalID || membership.UserID == current.ID {
					contaignCurrentUser = true
				}
				membership.ID = uuid.NewString()
				membership.GroupID = insertedID
				membership.DateCreated = now
				membership.ClientID = clientID
				castedMemberships = append(castedMemberships, membership)
			}
		}
		if current != nil && !contaignCurrentUser {
			castedMemberships = append(castedMemberships, model.GroupMembership{
				ID:          uuid.NewString(),
				GroupID:     insertedID,
				UserID:      current.ID,
				ClientID:    clientID,
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

		err = sa.UpdateGroupStats(context, clientID, group.ID, true, false, false, true)
		if err != nil {
			return err
		}

		sa.UpdateGroupAttributeIndexes(group)

		return nil
	}

	if context != nil {
		err = wrapperFunc(context)
	} else {
		err = sa.PerformTransaction(wrapperFunc)
	}
	if err != nil {
		if strings.Contains(err.Error(), "title_unique") {
			return nil, utils.NewGroupDuplicationError()
		}
		return nil, utils.NewServerError()
	}

	return &insertedID, nil
}

// UpdateGroup updates a group except the members attribute
func (sa *Adapter) UpdateGroup(context TransactionContext, clientID string, current *model.User, group *model.Group) *utils.GroupError {
	return sa.updateGroup(context, clientID, current, group, nil)
}

// UpdateGroupWithMembership updates a group along with the memberships
func (sa *Adapter) UpdateGroupWithMembership(context TransactionContext, clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) *utils.GroupError {
	return sa.updateGroup(context, clientID, current, group, memberships)
}

func (sa *Adapter) updateGroup(context TransactionContext, clientID string, current *model.User, group *model.Group, memberships []model.GroupMembership) *utils.GroupError {

	var err error
	wrapperFunc := func(context TransactionContext) error {

		// Check the title is unique. Don't rely on the unique index.
		if err := sa.checkUniqueGroupTitleWithContext(context, clientID, &group.ID, group.Title); err != nil {
			return err
		}

		setOperation := bson.D{
			primitive.E{Key: "title", Value: group.Title},
			primitive.E{Key: "privacy", Value: group.Privacy},
			primitive.E{Key: "hidden_for_search", Value: group.HiddenForSearch},
			primitive.E{Key: "description", Value: group.Description},
			primitive.E{Key: "image_url", Value: group.ImageURL},
			primitive.E{Key: "web_url", Value: group.WebURL},
			primitive.E{Key: "membership_questions", Value: group.MembershipQuestions},
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

			var userID *string
			if current != nil {
				userID = &current.ID
			}
			persistedGroup, err := sa.FindGroup(context, clientID, group.ID, userID)
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
				primitive.E{Key: "client_id", Value: clientID},
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
						primitive.E{Key: "client_id", Value: clientID},
					}
					err = sa.db.groupMemberships.ReplaceOneWithContext(context, filter, membership, nil)
					if err != nil {
						return err
					}
				}
			}
		}

		// Don't the date_updated field of the group (https://github.com/rokwire/illinois-app/issues/5119#issuecomment-2967421259)
		err = sa.UpdateGroupStats(context, clientID, group.ID, false, len(memberships) > 0, false, true)
		if err != nil {
			return err
		}

		sa.UpdateGroupAttributeIndexes(group)

		return nil
	}

	if context != nil {
		err = wrapperFunc(context)
	} else {
		// transaction
		err = sa.PerformTransaction(wrapperFunc)
	}
	if err != nil {
		if strings.Contains(err.Error(), "title_unique") {
			return utils.NewGroupDuplicationError()
		}
		return utils.NewServerError()
	}

	return nil
}

func (sa *Adapter) checkUniqueGroupTitleWithContext(context TransactionContext, clientID string, id *string, title string) error {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "title", Value: title},
	}
	if id != nil {
		filter = append(filter, primitive.E{Key: "_id", Value: bson.M{"$ne": *id}})
	}

	count, err := sa.db.groups.CountDocumentsWithContext(context, filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("voilation of unique title constraint: title_unique")
	}
	return nil
}

// DeleteGroup deletes a group.
func (sa *Adapter) DeleteGroup(ctx TransactionContext, clientID string, id string) error {

	wrapper := func(context TransactionContext) error {

		// 1. delete mapped group events
		_, err := sa.db.events.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID},
		}, nil)
		if err != nil {
			return err
		}

		// 2. delete mapped group posts
		_, err = sa.db.posts.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID},
		}, nil)
		if err != nil {
			return err
		}

		// 3. delete mapped group memberships
		_, err = sa.db.groupMemberships.DeleteManyWithContext(context, bson.D{
			primitive.E{Key: "group_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID},
		}, nil)
		if err != nil {
			return err
		}

		// 4. delete the group
		_, err = sa.db.groups.DeleteOneWithContext(context, bson.D{
			primitive.E{Key: "_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID},
		}, nil)
		if err != nil {
			return err
		}

		return nil
	}

	if ctx != nil {
		return wrapper(ctx)
	}
	return sa.PerformTransaction(wrapper)
}

// FindGroup finds group by id and client id
func (sa *Adapter) FindGroup(context TransactionContext, clientID string, groupID string, userID *string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID}}

	var err error
	var membership *model.GroupMembership
	if userID != nil {
		// find group memberships
		membership, err = sa.FindGroupMembershipWithContext(context, clientID, groupID, *userID)
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
func (sa *Adapter) FindGroupByTitle(clientID string, title string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
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
func (sa *Adapter) FindGroups(clientID string, userID *string, groupsFilter model.GroupsFilter, skipMembershipCheck bool) (int64, []model.Group, error) {
	// TODO: Merge the filter logic in a common method (FindGroups, FindGroupsV3, FindUserGroups)
	var count int64
	var list []model.Group
	err := sa.PerformTransaction(func(ctx TransactionContext) error {
		filter, memberships, err := sa.buildMainQuery(ctx, userID, clientID, groupsFilter, skipMembershipCheck)
		if err != nil {
			return err
		}

		type rowNumber struct {
			RowNumber int `json:"_row_number" bson:"_row_number"`
		}

		var aggrSort bson.D
		if groupsFilter.Order != nil && "desc" == *groupsFilter.Order {
			aggrSort = bson.D{
				{Key: "_c_title", Value: -1},
			}
		} else {
			aggrSort = bson.D{
				{Key: "_c_title", Value: 1},
			}

		}

		var limitIDRowNumber int64
		if groupsFilter.LimitID != nil {
			var rowNumbers []rowNumber
			err := sa.db.groups.AggregateWithContext(ctx, bson.A{
				bson.D{{Key: "$match", Value: filter}},
				bson.D{{Key: "$addFields", Value: bson.M{"_c_title": bson.M{"$toUpper": "$title"}}}},
				bson.D{{Key: "$sort", Value: aggrSort}},
				//bson.D{{Key: "$skip", Value: skip}},
				bson.D{
					{Key: "$setWindowFields",
						Value: bson.D{
							{Key: "sortBy", Value: aggrSort},
							{Key: "output", Value: bson.D{{Key: "_row_number", Value: bson.D{{Key: "$documentNumber", Value: bson.D{}}}}}},
						},
					},
				},
				bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: *groupsFilter.LimitID}}}},
			}, &rowNumbers, nil)
			if err != nil {
				return err
			}

			if len(rowNumbers) > 0 && rowNumbers[0].RowNumber != 0 {
				limitIDRowNumber = rowNumbers[0].RowNumber
			}
		}

		offset := int64(0)
		if groupsFilter.Offset != nil {
			offset = *groupsFilter.Offset
		}

		pipeline := bson.A{
			bson.D{{Key: "$match", Value: filter}},
			bson.D{{Key: "$addFields", Value: bson.M{"_c_title": bson.M{"$toUpper": "$title"}}}},
			bson.D{{Key: "$sort", Value: aggrSort}},
			bson.D{{Key: "$skip", Value: offset}},
		}
		if groupsFilter.Limit != nil {
			limitIDRowNumber = int64(math.Max(float64(limitIDRowNumber), float64(*groupsFilter.Limit)))
			if limitIDRowNumber > 0 {
				pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(limitIDRowNumber)}})
			} else {
				pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *groupsFilter.Limit}})
			}
		}

		count, err = sa.db.groups.CountDocumentsWithContext(ctx, filter)
		if err != nil {
			return err
		}

		//
		//err = sa.db.groups.FindWithContext(ctx, filter, &list, findOptions)
		err = sa.db.groups.AggregateWithContext(ctx, pipeline, &list, nil)
		if err != nil {
			return err
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

		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return count, list, nil
}

func (sa *Adapter) buildMainQuery(context TransactionContext, userID *string, clientID string, groupsFilter model.GroupsFilter, skipMembershipCheck bool) (bson.D, model.MembershipCollection, error) {

	groupIDs := []string{}
	groupIDMap := map[string]bool{}
	if len(groupsFilter.GroupIDs) > 0 {
		for _, groupID := range groupsFilter.GroupIDs {
			groupIDs = append(groupIDs, groupID)
			groupIDMap[groupID] = true
		}
	}

	var memberships model.MembershipCollection

	var userIDFilter *string
	if userID != nil && !skipMembershipCheck {
		userIDFilter = userID
	} else if groupsFilter.MemberUserID != nil {
		userIDFilter = groupsFilter.MemberUserID
	}
	// Credits to Ryan Oberlander suggest
	if userID != nil || groupsFilter.MemberID != nil || groupsFilter.MemberExternalID != nil {
		// find group memberships
		var err error
		memberships, err = sa.FindGroupMembershipsWithContext(context, clientID, model.MembershipFilter{
			ID:         groupsFilter.MemberID,
			UserID:     userIDFilter,
			ExternalID: groupsFilter.MemberExternalID,
			Statuses:   groupsFilter.MemberStatus,
		})
		if err != nil {
			return nil, model.MembershipCollection{}, err
		}

		for _, membership := range memberships.Items {
			if len(groupIDMap) == 0 || !groupIDMap[membership.GroupID] {
				groupIDs = append(groupIDs, membership.GroupID)
				groupIDMap[membership.GroupID] = true
			}
		}
	}

	filter := bson.D{}
	if len(groupsFilter.MemberStatus) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": groupIDs}})
	}
	if groupsFilter.GroupIDs != nil {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": groupsFilter.GroupIDs}})
	}
	if userID != nil && !skipMembershipCheck {
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

	if groupsFilter.AuthmanEnabled != nil {
		filter = append(filter, primitive.E{Key: "authman_enabled", Value: groupsFilter.AuthmanEnabled})
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

	if groupsFilter.Administrative != nil {
		filter = append(filter, primitive.E{Key: "administrative", Value: groupsFilter.Administrative})
	}

	if groupsFilter.ResearchOpen != nil {
		if *groupsFilter.ResearchOpen {
			filter = append(filter, primitive.E{Key: "research_open", Value: true})
		} else {
			filter = append(filter, primitive.E{Key: "research_open", Value: primitive.M{"$ne": true}})
		}
	}
	if groupsFilter.ResearchGroup != nil {
		if *groupsFilter.ResearchGroup {
			filter = append(filter, primitive.E{Key: "research_group", Value: true})
		} else {
			filter = append(filter, primitive.E{Key: "research_group", Value: primitive.M{"$ne": true}})
		}
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

	if groupsFilter.DaysInactive != nil {
		pastTime := time.Now().Add(time.Duration(*groupsFilter.DaysInactive) * -24 * time.Hour)
		filter = append(filter, primitive.E{Key: "date_updated", Value: bson.M{"$lt": pastTime}})
	}
	return filter, memberships, nil
}

// FindGroupByID finds one groups by ID and clientID
func (sa *Adapter) FindGroupByID(clientID string, groupID string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "_id", Value: groupID},
	}

	findOptions := options.FindOne()

	var rec model.Group
	err := sa.db.groups.FindOne(filter, &rec, findOptions)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

type findUserGroupsCountResult struct {
	Count int64 `bson:"count"`
}

// FindUserGroupsCount retrieves the count of current groups that the user is member
func (sa *Adapter) FindUserGroupsCount(clientID string, userID string) (*int64, error) {
	pipeline := []primitive.M{
		{"$match": primitive.M{
			"client_id": clientID,
			"user_id":   userID,
		}},
		{"$count": "count"},
	}
	var result []findUserGroupsCountResult
	err := sa.db.groupMemberships.Aggregate(pipeline, &result, &options.AggregateOptions{})
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return &result[0].Count, nil
	}
	return nil, nil
}

// FindUserGroups finds the user groups for client id
func (sa *Adapter) FindUserGroups(clientID string, userID string, groupsFilter model.GroupsFilter) ([]model.Group, error) {
	// TODO: Merge the filter logic in a common method (FindGroups, FindGroupsV3, FindUserGroups)

	// find group memberships
	memberships, err := sa.FindUserGroupMemberships(clientID, userID)
	if err != nil {
		return nil, err
	}
	groupIDs := []string{}
	for _, membership := range memberships.Items {
		groupIDs = append(groupIDs, membership.GroupID)
	}

	mongoFilter := bson.M{
		"_id":       bson.M{"$in": groupIDs},
		"client_id": clientID,
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
	if groupsFilter.Administrative != nil {
		mongoFilter["administrative"] = *groupsFilter.Administrative
	}

	if groupsFilter.ResearchOpen != nil {
		if *groupsFilter.ResearchOpen {
			mongoFilter["research_open"] = true
		} else {
			mongoFilter["research_open"] = primitive.M{"$ne": true}
		}
	}
	if groupsFilter.ResearchGroup != nil {
		if *groupsFilter.ResearchGroup {
			mongoFilter["research_group"] = true
		} else {
			mongoFilter["research_group"] = bson.M{"$ne": true}
		}
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
	findOptions.SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 1, // Case and diacritic insensitive
	})

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
func (sa *Adapter) FindEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	filter := bson.D{
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID},
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
func (sa *Adapter) CreateEvent(context TransactionContext, clientID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	event := model.Event{
		ClientID:      clientID,
		EventID:       eventID,
		GroupID:       groupID,
		DateCreated:   time.Now().UTC(),
		ToMembersList: toMemberList,
		Creator:       creator,
	}

	wrapper := func(context TransactionContext) error {
		_, err := sa.db.events.InsertOne(event)
		if err != nil {
			return err
		}

		return sa.UpdateGroupStats(context, clientID, groupID, true, false, false, false)
	}

	var err error
	if context != nil {
		err = wrapper(context)
	} else {
		err = sa.PerformTransaction(wrapper)
	}

	return &event, err
}

// UpdateEvent updates a group event
func (sa *Adapter) UpdateEvent(clientID string, eventID string, groupID string, toMemberList []model.ToMember) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.D{
			primitive.E{Key: "event_id", Value: eventID},
			primitive.E{Key: "group_id", Value: groupID},
			primitive.E{Key: "client_id", Value: clientID},
		}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
				primitive.E{Key: "to_members", Value: toMemberList},
			}},
		}
		_, err := sa.db.events.UpdateOneWithContext(context, filter, change, nil)
		if err == nil {
			return err
		}

		return sa.UpdateGroupStats(context, clientID, groupID, true, false, false, false)
	})
}

// DeleteEvent deletes a group event
func (sa *Adapter) DeleteEvent(clientID string, eventID string, groupID string) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.D{primitive.E{Key: "event_id", Value: eventID},
			primitive.E{Key: "group_id", Value: groupID},
			primitive.E{Key: "client_id", Value: clientID}}
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

		return sa.UpdateGroupStats(context, clientID, groupID, true, false, false, false)
	})
}

// FindPosts Retrieves posts for a group
func (sa *Adapter) FindPosts(clientID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error) {

	var resultList = make([]model.Post, 0)
	var postIndexMapping = make(map[string]int)

	wrapper := func(ctx TransactionContext) error {
		var userID *string
		if current != nil {
			userID = &current.ID
		}

		group, errGr := sa.FindGroup(ctx, clientID, filter.GroupID, userID)
		if group == nil {
			if errGr != nil {
				log.Printf("unable to find group with id %s: %s", filter.GroupID, errGr)
			} else {
				log.Printf("group does not exists %s", filter.GroupID)
			}
			return errGr
		}

		mongoFilter := bson.D{
			primitive.E{Key: "client_id", Value: clientID},
			primitive.E{Key: "group_id", Value: filter.GroupID},
		}

		if filter.PostType != nil {
			if *filter.PostType == "message" {
				mongoFilter = append(mongoFilter, bson.E{Key: "$or", Value: []bson.M{
					{"$and": []bson.M{
						{"to_members": bson.M{"$ne": primitive.Null{}}},
						{"parent_id": primitive.Null{}},
					}},
					{"parent_id": bson.M{"$ne": primitive.Null{}}},
				}})
			} else if *filter.PostType == "post" {
				mongoFilter = append(mongoFilter, bson.E{Key: "$or", Value: []bson.M{
					{"$and": []bson.M{
						{"to_members": primitive.Null{}},
						{"parent_id": primitive.Null{}},
					}},
					{"parent_id": bson.M{"$ne": primitive.Null{}}},
				}})
			}
		}

		if filter.ScheduledOnly != nil && *filter.ScheduledOnly {
			mongoFilter = append(mongoFilter, bson.E{Key: "date_scheduled", Value: bson.M{"$gt": time.Now()}})
		} else {
			mongoFilter = append(mongoFilter, bson.E{Key: "$or", Value: []bson.M{
				{"date_scheduled": nil},
				{"date_scheduled": bson.M{"$lt": time.Now()}},
			}})
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
			mongoFilter = append(mongoFilter, primitive.E{Key: "$or", Value: innerFilter})
		}

		if filterPrivatePostsValue != nil {
			mongoFilter = append(mongoFilter, primitive.E{Key: "private", Value: *filterPrivatePostsValue})
		}

		paging := false
		findOptions := options.Find()
		if filter.Order != nil && "desc" == *filter.Order {
			findOptions.SetSort(bson.D{{Key: "date_created", Value: -1}})
		} else {
			findOptions.SetSort(bson.D{{Key: "date_created", Value: 1}})
		}
		if filter.Limit != nil {
			findOptions.SetLimit(*filter.Limit)
			paging = true
		}
		if filter.Offset != nil {
			findOptions.SetSkip(*filter.Offset)
			paging = true
		}

		if paging {
			mongoFilter = append(mongoFilter, primitive.E{Key: "parent_id", Value: nil})
		}

		var list []model.Post
		err := sa.db.posts.FindWithContext(ctx, mongoFilter, &list, findOptions)
		if err != nil {
			return err
		}

		if paging && len(list) > 0 {
			for _, post := range list {
				childPosts, err := sa.FindPostsByTopParentID(ctx, clientID, current, filter.GroupID, post.ID, true, filter.Order)
				if err == nil && childPosts != nil {
					for _, childPost := range childPosts {
						if childPost.UserCanSeePost(current.ID) {
							list = append(list, childPost)
						}
					}
				}
			}
		}

		if list != nil {
			for i := range list {
				postID := list[i].ID
				list[i].Replies = make([]model.Post, 0)
				postIndexMapping[postID] = i
			}
			for _, post := range list {
				if post.ParentID != nil {
					if listIndex, ok := postIndexMapping[*post.ParentID]; ok {
						list[listIndex].Replies = append(list[listIndex].Replies, post)
					}
				}
			}
			for index := range list {
				if list[index].ParentID == nil {
					resultList = append(resultList, list[index])
				}
			}
		}
		return nil
	}

	err := sa.PerformTransaction(wrapper)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

// FindAllUserPosts Retrieves all user posts across all existing groups
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindAllUserPosts(context TransactionContext, clientID string, userID string) ([]model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
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
func (sa *Adapter) FindPost(context TransactionContext, clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {

	var post *model.Post
	wrapper := func(context TransactionContext) error {

		postRecord, err := sa.findPostWithContext(context, clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
		if err != nil {
			return err
		}

		post = postRecord

		if postRecord != nil {
			nestedPosts, err := sa.FindPostsByParentID(context, clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers, true, nil)
			if err != nil {
				return err
			}

			post.Replies = nestedPosts
		}

		return nil
	}

	var err error
	if context != nil {
		err = wrapper(context)
	} else {
		err = sa.PerformTransaction(wrapper)
	}

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (sa *Adapter) findPostWithContext(context TransactionContext, clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "_id", Value: postID},
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
		membership, err := sa.FindGroupMembership(clientID, groupID, *userID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	var post *model.Post
	err := sa.db.posts.FindOne(filter, &post, nil)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// FindTopPostByParentID Finds the top post by parent id
func (sa *Adapter) FindTopPostByParentID(clientID string, current *model.User, groupID string, parentID string, skipMembershipCheck bool) (*model.Post, error) {
	filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: parentID}}

	if !skipMembershipCheck {
		membership, err := sa.FindGroupMembership(clientID, groupID, current.ID)
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
		return sa.FindTopPostByParentID(clientID, current, groupID, *post.ParentID, skipMembershipCheck)
	}

	return post, nil
}

// FindPostsByParentID FindPostByParentID Retrieves a post by groupID and postID
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindPostsByParentID(ctx TransactionContext, clientID string, userID *string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]model.Post, error) {

	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "parent_id", Value: parentID},
	}

	if !skipMembershipCheck && userID != nil {
		membership, err := sa.FindGroupMembershipWithContext(ctx, clientID, groupID, *userID)
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

	var posts []model.Post
	err := sa.db.posts.Find(filter, &posts, findOptions)
	if err != nil {
		return nil, err
	}

	if recursive {
		if len(posts) > 0 {
			for index, post := range posts {
				childPosts, err := sa.FindPostsByParentID(ctx, clientID, userID, groupID, post.ID, true, filterByToMembers, recursive, order)
				if err == nil {
					posts[index].Replies = childPosts
				} else {
					log.Printf("error on finding nested posts: %s", err)
				}
			}
		}
	}

	return posts, nil
}

// FindPostsByTopParentID  Retrieves a post by groupID and top parent id
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindPostsByTopParentID(context TransactionContext, clientID string, current *model.User, groupID string, topParentID string, skipMembershipCheck bool, order *string) ([]model.Post, error) {
	var posts []model.Post
	wrapper := func(ctx TransactionContext) error {
		filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "top_parent_id", Value: topParentID}}

		if !skipMembershipCheck {
			membership, err := sa.FindGroupMembershipWithContext(ctx, clientID, groupID, current.ID)
			if membership == nil || err != nil || !membership.IsAdminOrMember() {
				return fmt.Errorf("the user is not member or admin of the group")
			}
		}

		findOptions := options.Find()
		if order != nil && "desc" == *order {
			findOptions.SetSort(bson.D{{Key: "date_created", Value: -1}})
		} else {
			findOptions.SetSort(bson.D{{Key: "date_created", Value: 1}})
		}

		err := sa.db.posts.FindWithContext(ctx, filter, &posts, findOptions)
		if err != nil {
			return err
		}

		return nil
	}

	var err error
	if context != nil {
		err = wrapper(context)
	} else {
		err = sa.PerformTransaction(wrapper)
	}
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// CreatePost Created a post
func (sa *Adapter) CreatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error) {

	if current != nil && post != nil {
		membership, err := sa.FindGroupMembership(clientID, post.GroupID, current.ID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}

		post.ClientID = clientID

		if post.ID == "" { // Always required
			post.ID = uuid.New().String()
		}

		if post.Replies != nil { // This is constructed only for GET all for group
			post.Replies = nil
		}

		if post.ParentID != nil {
			topPost, _ := sa.FindTopPostByParentID(clientID, current, post.GroupID, *post.ParentID, false)
			if topPost != nil && topPost.ParentID == nil {
				post.TopParentID = &topPost.ID
			}
		}

		now := time.Now()

		// Otherwise it will brake the order
		if post.DateScheduled != nil {
			post.DateCreated = *post.DateScheduled
		} else {
			post.DateCreated = now
		}

		post.Creator = model.Creator{
			UserID: current.ID,
			Email:  current.Email,
			Name:   current.Name,
		}

		err = sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.posts.InsertOneWithContext(context, post)
			if err != nil {
				return err
			}

			err = sa.UpdateGroupStats(context, clientID, post.GroupID, true, false, false, false)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		return post, err
	}
	return nil, nil
}

// UpdatePost Updates a post
func (sa *Adapter) UpdatePost(clientID string, userID string, post *model.Post) (*model.Post, error) {
	if post != nil {
		originalPost, _ := sa.FindPost(nil, clientID, &userID, post.GroupID, post.ID, true, true)
		if originalPost == nil {
			return nil, fmt.Errorf("unable to find post with id (%s) ", post.ID)
		}
		if originalPost.Creator.UserID != userID {
			return nil, fmt.Errorf("only creator of the post can update it")
		}

		post.ClientID = clientID

		if post.ID == "" { // Always required
			return nil, fmt.Errorf("Missing id")
		}

		now := time.Now()
		post.DateUpdated = &now

		// Sync date created until the notification is sent
		if post.DateScheduled != nil && originalPost.DateNotified == nil &&
			(originalPost.DateScheduled == nil || *originalPost.DateScheduled != *post.DateScheduled) {
			post.DateCreated = *post.DateScheduled
		}

		filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: post.ID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "subject", Value: post.Subject},
				primitive.E{Key: "body", Value: post.Body},
				primitive.E{Key: "private", Value: post.Private},
				primitive.E{Key: "use_as_notification", Value: post.UseAsNotification},
				primitive.E{Key: "is_abuse", Value: post.IsAbuse},
				primitive.E{Key: "image_url", Value: post.ImageURL},
				primitive.E{Key: "date_updated", Value: post.DateUpdated},
				primitive.E{Key: "date_scheduled", Value: post.DateScheduled},
				primitive.E{Key: "to_members", Value: post.ToMembersList},
			},
			},
		}

		err := sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.posts.UpdateOneWithContext(context, filter, update, nil)
			if err != nil {
				return err
			}

			return sa.UpdateGroupStats(context, clientID, post.GroupID, true, false, false, false)
		})
		if err != nil {
			return nil, err
		}

		return post, err
	}
	return nil, nil
}

// ReportGroupAsAbuse Report group as abuse
func (sa *Adapter) ReportGroupAsAbuse(clientID string, userID string, group *model.Group) error {
	if group != nil {
		filter := bson.D{
			primitive.E{Key: "client_id", Value: clientID},
			primitive.E{Key: "_id", Value: group.ID},
		}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "is_abuse", Value: true},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		_, err := sa.db.groups.UpdateOne(filter, update, nil)

		return err
	}
	return nil
}

// ReportPostAsAbuse Report post as abuse
func (sa *Adapter) ReportPostAsAbuse(clientID string, userID string, group *model.Group, post *model.Post) error {
	if post != nil {
		filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: post.ID}}

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
		return fmt.Errorf("error updating post %s with reaction %s for %s: %v", postID, reaction, userID, err)
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("updated %d posts with reaction %s for %s, but expected 1", res.ModifiedCount, reaction, userID)
	}

	return nil
}

// DeletePost Deletes a post
func (sa *Adapter) DeletePost(ctx TransactionContext, clientID string, userID string, groupID string, postID string, force bool) error {

	deleteWrapper := func(transactionContext TransactionContext) error {
		membership, _ := sa.FindGroupMembershipWithContext(transactionContext, clientID, groupID, userID)
		filterToMembers := true
		if membership != nil && membership.IsAdmin() {
			filterToMembers = false
		}

		originalPost, _ := sa.FindPost(transactionContext, clientID, &userID, groupID, postID, true, filterToMembers)
		if originalPost == nil {
			return fmt.Errorf("unable to find post with id (%s) ", postID)
		}

		if !force {
			if originalPost == nil || membership == nil || (!membership.IsAdmin() && originalPost.Creator.UserID != userID) {
				return fmt.Errorf("only creator of the post or group admin can delete it")
			}
		}

		var recursiveDelete func(post model.Post) error
		recursiveDelete = func(post model.Post) error {
			if post.Replies != nil && len(post.Replies) > 0 {
				for _, innerPost := range post.Replies {
					err := recursiveDelete(innerPost)
					if err != nil {
						return err
					}

					err = sa.DeletePost(transactionContext, clientID, userID, groupID, post.ID, true)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		childPosts, err := sa.FindPostsByParentID(transactionContext, clientID, &userID, groupID, postID, true, false, false, nil)
		if err != nil {
			return err
		}
		for _, childchildPost := range childPosts {
			recursiveDelete(childchildPost)
		}

		filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: postID}}

		_, err = sa.db.posts.DeleteOneWithContext(transactionContext, filter, nil)
		if err != nil {
			return err
		}

		return sa.UpdateGroupStats(transactionContext, clientID, groupID, true, false, false, false)
	}

	if ctx != nil {
		return deleteWrapper(ctx)
	}
	return sa.PerformTransaction(func(transactionContext TransactionContext) error {
		return deleteWrapper(transactionContext)
	})
}

// FindScheduledPosts Finds scheduled posts whithout sent notifications
func (sa *Adapter) FindScheduledPosts(context TransactionContext) ([]model.Post, error) {
	var posts []model.Post
	err := sa.db.posts.FindWithContext(context, bson.D{
		{Key: "date_scheduled", Value: bson.M{"$lt": time.Now()}},
		{Key: "date_notified", Value: nil},
	}, &posts, nil)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// UpdateDateNotifiedForPostIDs Updates the notification time for the desired posts
func (sa *Adapter) UpdateDateNotifiedForPostIDs(context TransactionContext, ids []string, dateNotified time.Time) error {
	_, err := sa.db.posts.UpdateManyWithContext(context,
		bson.D{{Key: "_id", Value: bson.M{"$in": ids}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "date_notified", Value: dateNotified}}}},
		nil)

	return err
}

// OnUpdatedGroupExternalEntity updates the group with the date of the last update by the linked external BBs
func (sa *Adapter) OnUpdatedGroupExternalEntity(context TransactionContext, groupID string, operation model.ExternalOperation) error {
	innerUpdate := bson.D{}
	now := time.Now()

	switch operation {
	case model.ExternalOperationEventUpdate:
		innerUpdate = append(innerUpdate, primitive.E{Key: "date_events_updated", Value: now})
	case model.ExternalOperationPollUpdate:
		innerUpdate = append(innerUpdate, primitive.E{Key: "date_polls_updated", Value: now})
	case model.ExternalOperationPostUpdate:
		innerUpdate = append(innerUpdate, primitive.E{Key: "date_posts_updated", Value: now})
	}
	innerUpdate = append(innerUpdate, primitive.E{Key: "date_updated", Value: now})

	// update the group
	filter := bson.D{
		primitive.E{Key: "_id", Value: groupID},
	}
	update := bson.D{
		primitive.E{Key: "$set", Value: innerUpdate},
	}

	_, err := sa.db.groups.UpdateOneWithContext(context, filter, update, nil)
	return err
}

// UpdateGroupStats set the updated date to the current date time (now)
func (sa *Adapter) UpdateGroupStats(context TransactionContext, clientID string, id string, resetUpdateDate, resetMembershipUpdateDate, resetManagedMembershipUpdateDate, resetStats bool) error {

	updateStats := func(ctx TransactionContext) error {
		innerUpdate := bson.D{}

		if resetStats {
			stats, err := sa.GetGroupMembershipStats(ctx, clientID, id)
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
			primitive.E{Key: "client_id", Value: clientID},
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
func (sa *Adapter) UpdateGroupDateUpdated(clientID string, groupID string) error {
	filter := bson.D{
		primitive.E{Key: "_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID},
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
func (sa *Adapter) FindAuthmanGroups(clientID string) ([]model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "authman_enabled", Value: true},
	}

	findOptions := options.Find()

	var list []model.Group
	err := sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindAuthmanGroupByKey Finds an Authman group by group long name
func (sa *Adapter) FindAuthmanGroupByKey(clientID string, authmanGroupKey string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "authman_group", Value: authmanGroupKey},
	}

	findOptions := options.Find()

	var list []model.Group
	err := sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

// cacheManagedGroupConfigs caches the managed group configs from the DB
func (sa *Adapter) cacheManagedGroupConfigs() error {
	log.Println("cacheManagedGroupConfigs..")

	configs, err := sa.LoadManagedGroupConfigs()
	if err != nil {
		return err
	}

	sa.setCachedManagedGroupConfigs(&configs)

	return nil
}

func (sa *Adapter) setCachedManagedGroupConfigs(configs *[]model.ManagedGroupConfig) {
	sa.managedGroupConfigsLock.Lock()
	defer sa.managedGroupConfigsLock.Unlock()

	sa.cachedManagedGroupConfigs = &syncmap.Map{}
	for _, config := range *configs {
		sa.cachedManagedGroupConfigs.Store(config.ID, config)
	}
}

func (sa *Adapter) getCachedManagedGroupConfig(id string) (*model.ManagedGroupConfig, error) {
	sa.managedGroupConfigsLock.RLock()
	defer sa.managedGroupConfigsLock.RUnlock()

	item, _ := sa.cachedManagedGroupConfigs.Load(id)
	if item != nil {
		config, ok := item.(model.ManagedGroupConfig)
		if !ok {
			return nil, fmt.Errorf("missing managed group config with id: %s", id)
		}
		return &config, nil
	}
	return nil, nil
}

func (sa *Adapter) getCachedManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	sa.managedGroupConfigsLock.RLock()
	defer sa.managedGroupConfigsLock.RUnlock()

	var err error
	configList := make([]model.ManagedGroupConfig, 0)
	sa.cachedManagedGroupConfigs.Range(func(key, item interface{}) bool {
		if item == nil {
			return false
		}

		config, ok := item.(model.ManagedGroupConfig)
		if !ok {
			err = fmt.Errorf("error casting config with id: %s", key)
			return false
		}
		if config.ClientID == clientID {
			configList = append(configList, config)
		}
		return true
	})

	return configList, err
}

// LoadManagedGroupConfigs loads all admin group config
func (sa *Adapter) LoadManagedGroupConfigs() ([]model.ManagedGroupConfig, error) {
	filter := bson.M{}

	findOptions := options.Find()

	var list []model.ManagedGroupConfig
	err := sa.db.managedGroupConfigs.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindManagedGroupConfig finds a managed group config by ID
func (sa *Adapter) FindManagedGroupConfig(id string, clientID string) (*model.ManagedGroupConfig, error) {
	config, err := sa.getCachedManagedGroupConfig(id)
	if err != nil {
		return nil, err
	}
	if config.ClientID != clientID {
		return nil, fmt.Errorf("invalid clientID %s for config ID %s", id, clientID)
	}
	return config, nil
}

// FindManagedGroupConfigs finds all managed group configs for a specified clientID
func (sa *Adapter) FindManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	return sa.getCachedManagedGroupConfigs(clientID)
}

// InsertManagedGroupConfig inserts a new managed group config
func (sa *Adapter) InsertManagedGroupConfig(config model.ManagedGroupConfig) error {
	_, err := sa.db.managedGroupConfigs.InsertOne(config)
	if err != nil {
		return err
	}

	return nil
}

// UpdateManagedGroupConfig updates an existing managed group config
func (sa *Adapter) UpdateManagedGroupConfig(config model.ManagedGroupConfig) error {
	filter := bson.M{"_id": config.ID, "client_id": config.ClientID}
	update := bson.M{"$set": bson.M{
		"authman_stems": config.AuthmanStems,
		"admin_uins":    config.AdminUINs,
		"type":          config.Type,
		"date_updated":  time.Now().UTC(),
	}}

	res, err := sa.db.managedGroupConfigs.UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("managed config could not be found for id: %s", config.ID)
	}

	return nil
}

// DeleteManagedGroupConfig deletes an existing managed group config
func (sa *Adapter) DeleteManagedGroupConfig(id string, clientID string) error {
	filter := bson.M{"_id": id, "client_id": clientID}

	res, err := sa.db.managedGroupConfigs.DeleteOne(filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount != 1 {
		return fmt.Errorf("managed config could not be found for id: %s", id)
	}
	return nil
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(context TransactionContext) error) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			sa.abortTransaction(sessionContext)
			return err
		}

		err = transaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return err
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return err
		}
		return nil
	})

	return err
}

func (sa *Adapter) abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error aborting a transaction - %s\n", err)
	}
}

// DeleteGroupMembershipsByAccountsIDs deletes the groups memberships by accountsIDs
func (sa *Adapter) DeleteGroupMembershipsByAccountsIDs(log *logs.Logger, context TransactionContext, accountsIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.groupMemberships.DeleteManyWithContext(context, filter, nil)
	return err
}

// DeleteUsersByAccountsIDs deletes users by accountsIDs
func (sa *Adapter) DeleteUsersByAccountsIDs(log *logs.Logger, context TransactionContext, accountsIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.users.DeleteManyWithContext(context, filter, nil)
	return err
}

// DeletePostsByAccountsIDs deletes posts users by accountsIDs
func (sa *Adapter) DeletePostsByAccountsIDs(log *logs.Logger, context TransactionContext, accountsIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "member.user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.posts.DeleteManyWithContext(context, filter, nil)
	return err
}

// PullMembersFromEventsByUserIDs deletes event members by accountsIDs
func (sa *Adapter) PullMembersFromEventsByUserIDs(log *logs.Logger, context TransactionContext, accountIDs []string) error {

	// Create filter to match documents containing the specified user IDs in to_members
	filter := bson.D{
		{Key: "to_members.user_id", Value: bson.D{{Key: "$in", Value: accountIDs}}},
	}

	// Create update to pull the whole object where user_id matches one of the specified accountIDs
	update := bson.D{
		{Key: "$pull", Value: bson.D{
			{Key: "to_members", Value: bson.M{"user_id": bson.D{{Key: "$in", Value: accountIDs}}}},
		}},
	}

	// Perform the update operation
	_, err := sa.db.events.UpdateManyWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}

	return nil
}

// PullMembersFromPostsByUserIDs deletes posts members by accountsIDs
func (sa *Adapter) PullMembersFromPostsByUserIDs(log *logs.Logger, context TransactionContext, accountIDs []string) error {

	// Create filter to match documents containing the specified user IDs in to_members
	filter := bson.D{
		{Key: "to_members.user_id", Value: bson.D{{Key: "$in", Value: accountIDs}}},
	}

	// Create update to pull the whole object where user_id matches one of the specified accountIDs
	update := bson.D{
		{Key: "$pull", Value: bson.D{
			{Key: "to_members", Value: bson.M{"user_id": bson.D{{Key: "$in", Value: accountIDs}}}},
		}},
	}

	// Perform the update operation
	_, err := sa.db.posts.UpdateManyWithContext(context, filter, update, nil)
	if err != nil {
		return err
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

	cachedSyncConfigs := &syncmap.Map{}
	syncConfigsLock := &sync.RWMutex{}

	cachedManagedGroupConfigs := &syncmap.Map{}
	managedGroupConfigsLock := &sync.RWMutex{}
	return &Adapter{db: db, cachedSyncConfigs: cachedSyncConfigs, syncConfigsLock: syncConfigsLock,
		cachedManagedGroupConfigs: cachedManagedGroupConfigs, managedGroupConfigsLock: managedGroupConfigsLock}
}

func abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error on aborting a transaction - %s", err)
	}
}

type storageListener struct {
	adapter *Adapter
	DefaultListenerImpl
}

func (sl *storageListener) OnConfigsChanged() {
	sl.adapter.cacheSyncConfigs()
}

func (sl *storageListener) OnManagedGroupConfigsChanged() {
	sl.adapter.cacheManagedGroupConfigs()
}

// Listener  listens for change data storage events
type Listener interface {
	OnConfigsChanged()
	OnManagedGroupConfigsChanged()
}

// DefaultListenerImpl default listener implementation
type DefaultListenerImpl struct{}

// OnConfigsChanged notifies configs have been updated
func (d *DefaultListenerImpl) OnConfigsChanged() {}

// OnManagedGroupConfigsChanged notifies managed group configs have been updated
func (d *DefaultListenerImpl) OnManagedGroupConfigsChanged() {}

// TransactionContext wraps mongo.SessionContext for use by external packages
type TransactionContext interface {
	mongo.SessionContext
}
