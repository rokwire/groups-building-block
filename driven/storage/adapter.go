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
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/syncmap"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type enumItem struct {
	ID     string   `bson:"_id"`
	Values []string `bson:"values"`
}

type group struct {
	ID                  string   `bson:"_id"`
	Category            string   `bson:"category"` //one of the enums categories list
	Title               string   `bson:"title"`
	Privacy             string   `bson:"privacy"` //public or private
	HiddenForSearch     bool     `bson:"hidden_for_search"`
	Description         *string  `bson:"description"`
	ImageURL            *string  `bson:"image_url"`
	WebURL              *string  `bson:"web_url"`
	Tags                []string `bson:"tags"`
	MembershipQuestions []string `bson:"membership_questions"`

	Members []member `bson:"members"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`

	ClientID                   string  `bson:"client_id"`
	AuthmanEnabled             bool    `bson:"authman_enabled"`
	AuthmanGroup               *string `bson:"authman_group"`
	OnlyAdminsCanCreatePolls   bool    `bson:"only_admins_can_create_polls"`
	BlockNewMembershipRequests bool    `bson:"block_new_membership_requests"`
	CanJoinAutomatically       bool    `json:"can_join_automatically" bson:"can_join_automatically"`
	AttendanceGroup            bool    `bson:"attendance_group"`
}

type member struct {
	ID            string         `bson:"id"`
	UserID        string         `bson:"user_id"`
	ExternalID    string         `bson:"external_id"`
	NetID         string         `bson:"net_id"`
	Name          string         `bson:"name"`
	Email         string         `bson:"email"`
	PhotoURL      string         `bson:"photo_url"`
	Status        string         `bson:"status"` //pending, member, admin, reject
	RejectReason  string         `bson:"reject_reason"`
	MemberAnswers []memberAnswer `bson:"member_answers"`

	DateCreated  time.Time  `bson:"date_created"`
	DateUpdated  *time.Time `bson:"date_updated"`
	DateAttended *time.Time `bson:"date_attended"`
}

type memberAnswer struct {
	Question string `bson:"question"`
	Answer   string `bson:"answer"`
}

//Adapter implements the Storage interface
type Adapter struct {
	db *database

	cachedManagedGroupConfigs *syncmap.Map
	managedGroupConfigsLock   *sync.RWMutex
}

//Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()

	//register storage listener
	sl := storageListener{adapter: sa}
	sa.RegisterStorageListener(&sl)

	err = sa.cacheManagedGroupConfigs()
	if err != nil {
		return errors.New("error caching managed group configs")
	}

	return err
}

//RegisterStorageListener registers a data change listener with the storage adapter
func (sa *Adapter) RegisterStorageListener(storageListener Listener) {
	sa.db.listeners = append(sa.db.listeners, storageListener)
}

// FindUser finds the user for the provided external id and client id
func (sa *Adapter) FindUser(clientID string, id string, external bool) (*model.User, error) {
	var filter bson.D
	if external {
		filter = bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "external_id", Value: id}}
	} else {
		filter = bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: id}}
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

// FindUsers finds all users for the provided list of (id | external id) and client id
func (sa *Adapter) FindUsers(clientID string, id []string, external bool) ([]model.User, error) {
	var filter bson.D
	if external {
		filter = bson.D{
			primitive.E{Key: "client_id", Value: clientID},
			primitive.E{Key: "external_id", Value: primitive.M{"$in": id}},
		}
	} else {
		filter = bson.D{
			primitive.E{Key: "client_id", Value: clientID},
			primitive.E{Key: "_id", Value: primitive.M{"$in": id}},
		}
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
func (sa *Adapter) LoginUser(clientID string, current *model.User) error {

	legacyUser, _ := sa.FindUser(clientID, current.ExternalID, true)

	now := time.Now()

	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		if legacyUser != nil && legacyUser.ID != current.ID {
			if legacyUser != nil {
				// delete the old user
				filter := bson.D{primitive.E{Key: "_id", Value: legacyUser.ID}, primitive.E{Key: "client_id", Value: clientID}}
				_, err = sa.db.users.DeleteOneWithContext(sessionContext, filter, nil)
				if err != nil {
					log.Printf("error deleting user - %s", err)
					abortTransaction(sessionContext)
					return err
				}
			}

			// insert the new user
			coreUser := model.User{ID: current.ID, ClientID: clientID, Email: current.Email, Name: current.Name,
				ExternalID: current.ExternalID, DateCreated: now, DateUpdated: &now, IsCoreUser: true}
			_, err = sa.db.users.InsertOneWithContext(sessionContext, &coreUser)
			if err != nil {
				log.Printf("error inserting user - %s", err)
				abortTransaction(sessionContext)
				return err
			}

			if legacyUser != nil {
				// Guess it may use $or but didnt managed to succeed with two positional arguments
				// 1.1 update all user's groups
				filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "members.user_id", Value: legacyUser.ID}}
				update := bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "members.$.user_id", Value: current.ID},
						primitive.E{Key: "members.$.date_updated", Value: now},
					}},
				}
				_, err = sa.db.groups.UpdateManyWithContext(sessionContext, filter, update, nil)
				if err != nil {
					log.Printf("error updating groups - %s", err)
					abortTransaction(sessionContext)
					return err
				}

				// 1.2 update all user's groups again for the email
				filter = bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "members.email", Value: legacyUser.Email}}
				update = bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "members.$.email", Value: legacyUser.Email},
						primitive.E{Key: "members.$.user_id", Value: current.ID},
						primitive.E{Key: "members.$.date_updated", Value: now},
					}},
				}
				_, err = sa.db.groups.UpdateManyWithContext(sessionContext, filter, update, nil)
				if err != nil {
					log.Printf("error updating groups - %s", err)
					abortTransaction(sessionContext)
					return err
				}

				// 1.3. update all user's posts
				filter = bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "member.user_id", Value: current.ID}}
				update = bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "member.user_id", Value: current.ID},
					}},
				}
				_, err = sa.db.posts.UpdateManyWithContext(sessionContext, filter, update, nil)
				if err != nil {
					log.Printf("error updating posts - %s", err)
					abortTransaction(sessionContext)
					return err
				}

				// 1.4. update all user's posts again but for the email
				filter = bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "member.email", Value: legacyUser.Email}}
				update = bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "member.email", Value: legacyUser.Email},
						primitive.E{Key: "member.user_id", Value: current.ID},
						primitive.E{Key: "member.date_updated", Value: now},
					}},
				}
				_, err = sa.db.posts.UpdateManyWithContext(sessionContext, filter, update, nil)
				if err != nil {
					log.Printf("error updating posts - %s", err)
					abortTransaction(sessionContext)
					return err
				}
			}
		} else {
			coreUser, _ := sa.FindUser(clientID, current.ID, false)
			if coreUser == nil {
				coreUser := model.User{ID: current.ID, ClientID: clientID, Email: current.Email, Name: current.Name,
					ExternalID: current.ExternalID, DateCreated: now, DateUpdated: &now, IsCoreUser: true}

				_, err := sa.db.users.InsertOneWithContext(sessionContext, &coreUser)
				if err != nil {
					abortTransaction(sessionContext)
					log.Printf("error inserting user - %s", err)
					return fmt.Errorf("error inserting user - %s", err)
				}
			}
		}

		if current.IsCoreUser {
			// Repopulate and keep sync of external_id & user_id. Part 1
			filter := bson.D{
				primitive.E{Key: "client_id", Value: clientID},
				primitive.E{Key: "members.external_id", Value: current.ExternalID},
			}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "members.$.name", Value: current.Name},
					primitive.E{Key: "members.$.email", Value: current.Email},
					primitive.E{Key: "members.$.user_id", Value: current.ID},
					primitive.E{Key: "members.$.external_id", Value: current.ExternalID},
					primitive.E{Key: "members.$.net_id", Value: current.NetID},
					primitive.E{Key: "members.$.date_updated", Value: now},
				}},
			}
			_, err := sa.db.groups.UpdateManyWithContext(sessionContext, filter, update, nil)
			if err != nil {
				log.Printf("error updating dummy member records for user(%s | %s) Part 1: %s", current.ID, current.ExternalID, err)
				return err
			}

			// Repopulate and keep sync of external_id & user_id. Part 2
			filter = bson.D{
				primitive.E{Key: "client_id", Value: clientID},
				primitive.E{Key: "members.user_id", Value: current.ID},
			}
			update = bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "members.$.name", Value: current.Name},
					primitive.E{Key: "members.$.email", Value: current.Email},
					primitive.E{Key: "members.$.user_id", Value: current.ID},
					primitive.E{Key: "members.$.external_id", Value: current.ExternalID},
					primitive.E{Key: "members.$.net_id", Value: current.NetID},
					primitive.E{Key: "members.$.date_updated", Value: now},
				}},
			}
			_, err = sa.db.groups.UpdateManyWithContext(sessionContext, filter, update, nil)
			if err != nil {
				log.Printf("error updating dummy member records for user(%s | %s) Part 2: %s", current.ID, current.ExternalID, err)
				return err
			}

			// Repopulate and keep sync of user in the user table. Part 3
			filter = bson.D{
				primitive.E{Key: "client_id", Value: clientID},
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
			_, err = sa.db.users.UpdateOneWithContext(sessionContext, filter, update, nil)
			if err != nil {
				log.Printf("error updating user(%s | %s) Part 3: %s", current.ID, current.ExternalID, err)
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
		return utils.NewServerError()
	}

	return nil
}

// CreateUser creates a new user
func (sa *Adapter) CreateUser(clientID string, id string, externalID string, email string, name string) (*model.User, error) {
	dateCreated := time.Now()
	user := model.User{ID: id, ClientID: clientID, ExternalID: externalID, Email: email, Name: name, DateCreated: dateCreated}
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

	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		filter := bson.D{primitive.E{Key: "members.user_id", Value: userID},
			primitive.E{Key: "client_id", Value: clientID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
			}},
			primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "members", Value: bson.M{"user_id": userID}}}},
		}
		_, err = sa.db.groups.UpdateManyWithContext(sessionContext, filter, change, nil)
		if err != nil {
			log.Printf("error unlinking user membership(user_id: %s) - %s", userID, err)
			abortTransaction(sessionContext)
			return err
		}

		posts, err := sa.FindAllUserPosts(clientID, userID)
		if err != nil {
			log.Printf("error on find all posts for user (%s) - %s", userID, err.Error())
			abortTransaction(sessionContext)
			return err
		}
		if len(posts) > 0 {
			for _, post := range posts {
				sa.deletePost(sessionContext, clientID, userID, post.GroupID, *post.ID, true)
			}
		}

		// delete the user
		filter = bson.D{primitive.E{Key: "_id", Value: userID}, primitive.E{Key: "client_id", Value: clientID}}
		_, err = sa.db.users.DeleteOneWithContext(sessionContext, filter, nil)
		if err != nil {
			log.Printf("error deleting user - %s", err.Error())
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

	return err
}

//FindUserGroupsMemberships stores user group membership
func (sa *Adapter) FindUserGroupsMemberships(id string, external bool) ([]*model.Group, *model.User, error) {
	userID := ""
	var err error
	var user *model.User
	if external {
		filter := bson.D{primitive.E{Key: "external_id", Value: id}}
		var result []*model.User
		err := sa.db.users.Find(filter, &result, nil)
		if err != nil {
			return nil, nil, err
		}
		if result == nil || len(result) == 0 {
			//not found
			return nil, nil, nil
		}
		user = result[0]
		userID = user.ID
	} else {
		userID = id
	}

	filterID := bson.D{primitive.E{Key: "members.user_id", Value: userID}}
	var resultList []*group
	err = sa.db.groups.Find(filterID, &resultList, nil)
	if err != nil {
		return nil, nil, err
	}

	modelGroups := make([]*model.Group, len(resultList))
	for i, current := range resultList {

		members := current.Members
		newMembers := make([]model.Member, len(members))
		for i, c := range members {
			newMembers[i] = model.Member{
				ID: c.ID, Status: c.Status, ExternalID: c.ExternalID, UserID: c.UserID,
			}
		}
		modelGroups[i] = &model.Group{ID: current.ID, Title: current.Title, Privacy: current.Privacy, Members: newMembers}
	}

	return modelGroups, user, nil

}

//ReadAllGroupCategories reads all group categories
func (sa *Adapter) ReadAllGroupCategories() ([]string, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: "categories"}}
	var result []enumItem
	err := sa.db.enums.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//not found
		return nil, nil
	}
	categoryItem := result[0]

	return categoryItem.Values, nil
}

//CreateGroup creates a group. Returns the id of the created group
func (sa *Adapter) CreateGroup(clientID string, current *model.User, group *model.Group) (*string, *utils.GroupError) {
	insertedID := uuid.NewString()

	existingGroups, err := sa.FindGroups(clientID, nil, nil, &group.Title, nil, nil, nil)
	if err == nil && len(existingGroups) > 0 {
		for _, persistedGrop := range existingGroups {
			if persistedGrop.ID != group.ID && strings.ToLower(persistedGrop.Title) == strings.ToLower(group.Title) {
				return nil, utils.NewGroupDuplicationError()
			}
		}
	}

	// transaction
	err = sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// insert the group and the admin member
		group.ID = insertedID
		group.ClientID = clientID
		group.DateCreated = time.Now()
		if current != nil && len(group.Members) == 0 {
			group.Members = []model.Member{{
				ID: uuid.NewString(), UserID: current.ID,
				Name: current.Name, Email: current.Email,
				ExternalID: current.ExternalID,
				PhotoURL:   "", Status: "admin", DateCreated: time.Now(),
			}}
		}

		_, err = sa.db.groups.InsertOneWithContext(sessionContext, &group)
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
		return nil, utils.NewServerError()
	}

	return &insertedID, nil
}

// UpdateGroupWithoutMembers updates a group except the members attribute
func (sa *Adapter) UpdateGroupWithoutMembers(clientID string, current *model.User, group *model.Group) *utils.GroupError {

	return sa.updateGroup(clientID, current, group, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "category", Value: group.Category},
			primitive.E{Key: "title", Value: group.Title},
			primitive.E{Key: "privacy", Value: group.Privacy},
			primitive.E{Key: "hidden_for_search", Value: group.HiddenForSearch},
			primitive.E{Key: "description", Value: group.Description},
			primitive.E{Key: "image_url", Value: group.ImageURL},
			primitive.E{Key: "web_url", Value: group.WebURL},
			primitive.E{Key: "tags", Value: group.Tags},
			primitive.E{Key: "membership_questions", Value: group.MembershipQuestions},
			primitive.E{Key: "date_updated", Value: time.Now()},
			primitive.E{Key: "authman_enabled", Value: group.AuthmanEnabled},
			primitive.E{Key: "authman_group", Value: group.AuthmanGroup},
			primitive.E{Key: "only_admins_can_create_polls", Value: group.OnlyAdminsCanCreatePolls},
			primitive.E{Key: "can_join_automatically", Value: group.CanJoinAutomatically},
			primitive.E{Key: "block_new_membership_requests", Value: group.BlockNewMembershipRequests},
			primitive.E{Key: "attendance_group", Value: group.AttendanceGroup},
		}},
	})
}

// UpdateGroupWithMembers updates a group along with the members
func (sa *Adapter) UpdateGroupWithMembers(clientID string, current *model.User, group *model.Group) *utils.GroupError {
	return sa.updateGroup(clientID, current, group, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "category", Value: group.Category},
			primitive.E{Key: "title", Value: group.Title},
			primitive.E{Key: "privacy", Value: group.Privacy},
			primitive.E{Key: "hidden_for_search", Value: group.HiddenForSearch},
			primitive.E{Key: "description", Value: group.Description},
			primitive.E{Key: "image_url", Value: group.ImageURL},
			primitive.E{Key: "web_url", Value: group.WebURL},
			primitive.E{Key: "tags", Value: group.Tags},
			primitive.E{Key: "membership_questions", Value: group.MembershipQuestions},
			primitive.E{Key: "date_updated", Value: time.Now()},
			primitive.E{Key: "authman_enabled", Value: group.AuthmanEnabled},
			primitive.E{Key: "authman_group", Value: group.AuthmanGroup},
			primitive.E{Key: "only_admins_can_create_polls", Value: group.OnlyAdminsCanCreatePolls},
			primitive.E{Key: "can_join_automatically", Value: group.CanJoinAutomatically},
			primitive.E{Key: "block_new_membership_requests", Value: group.BlockNewMembershipRequests},
			primitive.E{Key: "attendance_group", Value: group.AttendanceGroup},
			primitive.E{Key: "members", Value: group.Members},
		}},
	})
}

func (sa *Adapter) updateGroup(clientID string, current *model.User, group *model.Group, updateOperation bson.D) *utils.GroupError {
	existingGroups, err := sa.FindGroups(clientID, nil, nil, &group.Title, nil, nil, nil)
	if err == nil && len(existingGroups) > 0 {
		for _, persistedGrop := range existingGroups {
			if persistedGrop.ID != group.ID && strings.ToLower(persistedGrop.Title) == strings.ToLower(group.Title) {
				return utils.NewGroupDuplicationError()
			}
		}
	}

	// transaction
	err = sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// update the group
		filter := bson.D{primitive.E{Key: "_id", Value: group.ID},
			primitive.E{Key: "client_id", Value: clientID}}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, filter, updateOperation, nil)
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
		return utils.NewServerError()
	}
	return nil
}

//DeleteGroup deletes a group.
func (sa *Adapter) DeleteGroup(clientID string, id string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//1. delete mapped group events
		eventFilter := bson.D{primitive.E{Key: "group_id", Value: id}, primitive.E{Key: "client_id", Value: clientID}}
		_, err = sa.db.events.DeleteOne(eventFilter, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}

		//2. delete the group
		filter := bson.D{primitive.E{Key: "_id", Value: id}, primitive.E{Key: "client_id", Value: clientID}}
		_, err = sa.db.groups.DeleteOne(filter, nil)
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
	return nil
}

//FindGroup finds group by id and client id
func (sa *Adapter) FindGroup(clientID string, id string) (*model.Group, error) {
	return sa.findGroupWithContext(clientID, id)
}

func (sa *Adapter) findGroupWithContext(clientID string, id string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: id},
		primitive.E{Key: "client_id", Value: clientID}}
	var rec group
	err := sa.db.groups.FindOne(filter, &rec, nil)
	if err != nil {
		return nil, err
	}

	group := constructGroup(rec)
	return &group, nil
}

//FindGroupByMembership finds group by membership
func (sa *Adapter) FindGroupByMembership(clientID string, membershipID string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "members.id", Value: membershipID}, primitive.E{Key: "client_id", Value: clientID}}
	var result []*group
	err := sa.db.groups.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, nil
	}
	group := result[0]
	resultEntity := constructGroup(*group)
	return &resultEntity, nil
}

//FindGroupByTitle finds group by membership
func (sa *Adapter) FindGroupByTitle(clientID string, title string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "title", Value: title},
	}
	var result []*group
	err := sa.db.groups.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, nil
	}
	group := result[0]
	resultEntity := constructGroup(*group)
	return &resultEntity, nil
}

//FindGroups finds groups
func (sa *Adapter) FindGroups(clientID string, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error) {
	filter := bson.D{primitive.E{Key: "client_id", Value: clientID}}
	if category != nil {
		filter = append(filter, primitive.E{Key: "category", Value: category})
	}
	if title != nil {
		filter = append(filter, primitive.E{Key: "title", Value: primitive.Regex{Pattern: *title, Options: "i"}})
	}
	if privacy != nil {
		filter = append(filter, primitive.E{Key: "privacy", Value: privacy})
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{"date_created", -1}})
	} else {
		findOptions.SetSort(bson.D{{"date_created", 1}})
	}
	if limit != nil {
		findOptions.SetLimit(*limit)
	}
	if offset != nil {
		findOptions.SetSkip(*offset)
	}

	var list []group
	err := sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	result := make([]model.Group, len(list))
	if list != nil {
		for i, current := range list {
			item := constructGroup(current)
			result[i] = item
		}
	}

	return result, nil
}

// FindOneGroupBtID finds one groups by ID and clientID
func (sa *Adapter) FindOneGroupBtID(clientID string, groupID string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "_id", Value: groupID},
	}

	findOptions := options.FindOne()

	var rec group
	err := sa.db.groups.FindOne(filter, &rec, findOptions)
	if err != nil {
		return nil, err
	}

	group := constructGroup(rec)

	return &group, nil
}

type findUserGroupsCountResult struct {
	Count int64 `bson:"count"`
}

// FindUserGroupsCount retrieves the count of current groups that the user is member
func (sa *Adapter) FindUserGroupsCount(clientID string, userID string) (*int64, error) {
	pipeline := []primitive.M{
		primitive.M{"$match": primitive.M{
			"client_id":       clientID,
			"members.user_id": userID,
		}},
		primitive.M{"$count": "count"},
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

//FindUserGroups finds the user groups for client id
func (sa *Adapter) FindUserGroups(clientID string, userID string) ([]model.Group, error) {
	filter := bson.D{primitive.E{Key: "members.user_id", Value: userID},
		primitive.E{Key: "client_id", Value: clientID}}

	var list []group
	err := sa.db.groups.Find(filter, &list, nil)
	if err != nil {
		return nil, err
	}

	result := make([]model.Group, len(list))
	if list != nil {
		for i, current := range list {
			item := constructGroup(current)
			result[i] = item
		}
	}
	return result, nil
}

// UpdateGroupMembers Updates members for specific group
func (sa *Adapter) UpdateGroupMembers(clientID string, groupID string, members []model.Member) error {
	group, err := sa.FindGroup(clientID, groupID)
	if err != nil {
		log.Printf("error on find group %s: %s", groupID, err)
		return err
	}

	now := time.Now().UTC()
	group.Members = members
	group.DateUpdated = &now

	saveFilter := bson.D{primitive.E{Key: "_id", Value: groupID}, primitive.E{Key: "client_id", Value: clientID}}
	err = sa.db.groups.ReplaceOne(saveFilter, group, nil)
	if err != nil {
		log.Printf("error on updating members for group %s: %s", groupID, err)
		return err
	}

	return nil
}

//CreatePendingMember creates a pending member for a specific group
func (sa *Adapter) CreatePendingMember(clientID string, user *model.User, group *model.Group, member *model.Member) error {
	if member != nil && group != nil {
		// transaction
		err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				return err
			}

			//1. first check if there is a group for the prvoided group id
			groupFilter := bson.D{primitive.E{Key: "_id", Value: group.ID}, primitive.E{Key: "client_id", Value: clientID}}
			var result []*model.Group
			err = sa.db.groups.FindWithContext(sessionContext, groupFilter, &result, nil)
			if err != nil {
				abortTransaction(sessionContext)
				return err
			}
			if result == nil || len(result) == 0 {
				//there is no a group for the provided id
				abortTransaction(sessionContext)
				return errors.New("there is no a group for the provided id")
			}
			group := result[0]

			//2. check if the user is already a member of this group - pending or member or admin or rejected
			members := group.Members
			if members != nil {
				for _, cMember := range members {
					if cMember.UserID == user.ID {
						switch cMember.Status {
						case "admin":
							return errors.New("the user is an admin for the group")
						case "member":
							return errors.New("the user is a member for the group")
						case "pending":
							return errors.New("the user is pending for the group")
						case "rejected":
							return errors.New("the user is rejected for the group")
						default:
							return errors.New("error creating a pending user")
						}
					}
				}
			}

			//3. check if the answers match the group questions
			if len(group.MembershipQuestions) != len(member.MemberAnswers) {
				return errors.New("member answers mismatch")
			}

			//4. now we can add the pending member
			var memberAns []memberAnswer
			if len(member.MemberAnswers) > 0 {
				for _, cAns := range member.MemberAnswers {
					memberAns = append(memberAns, memberAnswer{Question: cAns.Question, Answer: cAns.Answer})
				}
			}

			member.ID = uuid.NewString()
			member.DateCreated = time.Now().UTC()

			groupMembers := group.Members
			groupMembers = append(groupMembers, *member)
			saveFilter := bson.D{primitive.E{Key: "_id", Value: group.ID}}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "members", Value: groupMembers},
					primitive.E{Key: "date_updated", Value: time.Now()},
				},
				},
			}
			_, err = sa.db.groups.UpdateOneWithContext(sessionContext, saveFilter, update, nil)
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
	}

	return nil
}

//DeletePendingMember deletes a pending member from a specific group
func (sa *Adapter) DeletePendingMember(clientID string, groupID string, userID string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//1. first check if there is a group for the prvoided group id
		groupFilter := bson.D{primitive.E{Key: "_id", Value: groupID}, primitive.E{Key: "client_id", Value: clientID}}
		var result []*group
		err = sa.db.groups.FindWithContext(sessionContext, groupFilter, &result, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if result == nil || len(result) == 0 {
			//there is no a group for the provided id
			abortTransaction(sessionContext)
			return errors.New("there is no a group for the provided id")
		}
		group := result[0]

		//2. delete the pending member
		members := group.Members
		indexToRemove := -1
		if len(members) > 0 {
			for i, cMember := range members {
				if cMember.UserID == userID && cMember.Status == "pending" {
					indexToRemove = i
					break
				}
			}
		}
		if indexToRemove == -1 {
			return errors.New("there is no pending request for the user")
		}

		// delete it from the members list
		members = append(members[:indexToRemove], members[indexToRemove+1:]...)

		//save it
		saveFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "members", Value: members},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, saveFilter, update, nil)
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

	return nil
}

//CreateMemberUnchecked Created a member to a group
func (sa *Adapter) CreateMemberUnchecked(clientID string, current *model.User, group *model.Group, member *model.Member) error {
	if group != nil {
		if !group.IsGroupAdmin(current.ID) && !group.CanJoinAutomatically {
			log.Printf("error: storage.CreateMemberUnchecked() - current user is not admin of the group")
			return fmt.Errorf("current user is not admin of the group")
		}

		if member.ExternalID != "" && group.GetMemberByExternalID(member.ExternalID) != nil {
			log.Printf("error: storage.CreateMemberUnchecked() - member of group '%s' with external id %s already exists", group.Title, member.ExternalID)
			return fmt.Errorf("member of group '%s' with external id %s already exists", group.Title, member.ExternalID)
		}

		if member.UserID != "" && group.GetMemberByUserID(member.UserID) != nil {
			log.Printf("error: storage.CreateMemberUnchecked() - member of group '%s' with user id %s already exists", group.Title, member.UserID)
			return fmt.Errorf("member of group '%s' with user id %s already exists", group.Title, member.UserID)
		}

		if len(member.UserID) == 0 && len(member.ExternalID) == 0 {
			log.Printf("error: storage.CreateMemberUnchecked() - expected user_id or external_id")
			return fmt.Errorf("expected user_id or external_id")
		}

		member.ID = uuid.NewString()
		member.DateCreated = time.Now()
		member.MemberAnswers = group.CreateMembershipEmptyAnswers()

		// transaction
		saveFilter := bson.D{
			primitive.E{Key: "_id", Value: group.ID},
		}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
			}},
			primitive.E{Key: "$push", Value: bson.D{
				primitive.E{Key: "members", Value: member},
			}},
		}
		_, err := sa.db.groups.UpdateOne(saveFilter, update, nil)
		if err != nil {
			return err
		}

	}

	return nil
}

//DeleteMember deletes a member membership from a specific group
func (sa *Adapter) DeleteMember(clientID string, groupID string, userID string, force bool) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// get the member as we need to validate it
		pipeline := []bson.M{
			{"$unwind": "$members"},
			{"$match": bson.M{"_id": groupID, "members.user_id": userID, "client_id": clientID}},
		}
		var result []struct {
			ID     string `bson:"_id"`
			Member member `bson:"members"`
		}
		err = sa.db.groups.AggregateWithContext(sessionContext, pipeline, &result, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if result == nil || len(result) == 0 {
			abortTransaction(sessionContext)
			return errors.New("there is an issue processing the item")
		}
		resultItem := result[0]
		member := resultItem.Member
		if !(member.Status == "admin" || member.Status == "member") && !force {
			abortTransaction(sessionContext)
			return errors.New("you are not member/admin to the group")
		}

		//check if the member is admin, do not allow the group to become with 0 admins
		if member.Status == "admin" && !force {
			adminsCount, err := sa.findAdminsCount(sessionContext, groupID)
			if err != nil {
				abortTransaction(sessionContext)
				return err
			}
			if *adminsCount == 1 {
				abortTransaction(sessionContext)
				return errors.New("you are the only admin for the group, you need to set another person for amdin before to leave")
			}
		}

		// delete the member, also keep the group members count updated
		changeFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
			}},
			primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "members", Value: bson.M{"id": member.ID}}}},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, changeFilter, change, nil)
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

	return nil
}

//ApplyMembershipApproval applies a membership approval
func (sa *Adapter) ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//1. first check if there is a group for the provided membership id
		groupFilter := bson.D{primitive.E{Key: "members.id", Value: membershipID}, primitive.E{Key: "client_id", Value: clientID}}
		var result []*group
		err = sa.db.groups.FindWithContext(sessionContext, groupFilter, &result, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if result == nil || len(result) == 0 {
			//there is no a group for the provided membership id
			abortTransaction(sessionContext)
			return errors.New("there is no a group for the provided membership id")
		}
		group := result[0]

		//2. find the member
		memberIndex := -1
		var member member
		if len(group.Members) > 0 {
			for i, cMember := range group.Members {
				if cMember.ID == membershipID && cMember.Status == "pending" {
					member = cMember
					memberIndex = i
					break
				}
			}
		}
		if memberIndex == -1 {
			return errors.New("there is an issue with the reject member index")
		}

		//3. apply approve/deny
		groupMembers := group.Members
		now := time.Now()
		if approve {
			//apply approve
			member.DateUpdated = &now
			member.Status = "member"
			groupMembers[memberIndex] = member
		} else {
			//apply deny
			member.DateUpdated = &now
			member.Status = "rejected"
			member.RejectReason = rejectReason
			groupMembers[memberIndex] = member
		}

		saveFilter := bson.D{primitive.E{Key: "_id", Value: group.ID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "members", Value: groupMembers},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, saveFilter, update, nil)
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

	return nil
}

//DeleteMembership deletes a membership
func (sa *Adapter) DeleteMembership(clientID string, current *model.User, membershipID string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// get the member as we need to validate it
		pipeline := []bson.M{
			{"$unwind": "$members"},
			{"$match": bson.M{"members.id": membershipID, "client_id": clientID}},
		}
		var result []struct {
			GroupID string `bson:"_id"`
			Member  member `bson:"members"`
		}
		err = sa.db.groups.AggregateWithContext(sessionContext, pipeline, &result, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if result == nil || len(result) == 0 {
			abortTransaction(sessionContext)
			return errors.New("there is an issue processing the item")
		}
		resultItem := result[0]
		groupID := resultItem.GroupID
		member := resultItem.Member
		if member.UserID == current.ID {
			abortTransaction(sessionContext)
			return errors.New("you cannot remove yourself")
		}
		if !(member.Status == "admin" || member.Status == "member" || member.Status == "rejected") {
			abortTransaction(sessionContext)
			return errors.New("membership which is not member or admin or rejected cannot be removed from the group")
		}

		// delete the membership, also keep the group members count updated
		changeFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
			}},
			primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "members", Value: bson.M{"id": member.ID}}}},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, changeFilter, change, nil)
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

	return nil
}

//UpdateMembership updates a membership
func (sa *Adapter) UpdateMembership(clientID string, current *model.User, membershipID string, status string, dateAttended *time.Time) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// get the member as we need to validate it
		pipeline := []bson.M{
			{"$unwind": "$members"},
			{"$match": bson.M{"members.id": membershipID, "client_id": clientID}},
		}
		var result []struct {
			GroupID string `bson:"_id"`
			Member  member `bson:"members"`
		}
		err = sa.db.groups.AggregateWithContext(sessionContext, pipeline, &result, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if result == nil || len(result) == 0 {
			abortTransaction(sessionContext)
			return errors.New("there is an issue processing the item")
		}
		resultItem := result[0]
		member := resultItem.Member
		if member.UserID == current.ID {
			abortTransaction(sessionContext)
			return errors.New("you cannot update yourself")
		}
		//check only admin or member to be updated
		if !(member.Status == "admin" || member.Status == "member") {
			abortTransaction(sessionContext)
			return errors.New("only admin or member can be updated")
		}

		// update the membership
		changeFilter := bson.D{primitive.E{Key: "members.id", Value: membershipID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
				primitive.E{Key: "members.$.status", Value: status},
				primitive.E{Key: "members.$.date_updated", Value: time.Now()},
				primitive.E{Key: "members.$.date_attended", Value: dateAttended},
			}},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, changeFilter, change, nil)
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

	return nil
}

//FindEvents finds the events for a group
func (sa *Adapter) FindEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	filter := bson.D{
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID},
	}
	if filterByToMembers && current != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: []primitive.M{
			primitive.M{"to_members": primitive.Null{}},
			primitive.M{"to_members": primitive.M{"$exists": true, "$size": 0}},
			primitive.M{"to_members.user_id": current.ID},
			primitive.M{"member.user_id": current.ID},
		}})
	}

	var result []model.Event
	err := sa.db.events.Find(filter, &result, nil)
	return result, err
}

//CreateEvent creates a group event
func (sa *Adapter) CreateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) (*model.Event, error) {
	var creator *model.Creator
	if current != nil {
		creator = current.ToCreator()
	}
	event := model.Event{
		ClientID:      clientID,
		EventID:       eventID,
		GroupID:       groupID,
		DateCreated:   time.Now().UTC(),
		ToMembersList: toMemberList,
		Creator:       creator,
	}
	_, err := sa.db.events.InsertOne(event)
	if err != nil {
		return nil, err
	}

	if err == nil {
		sa.resetGroupUpdatedDate(clientID, groupID)
	}

	return &event, err
}

//CreateEventWithCreator creates a group event with predefined creator record
func (sa *Adapter) CreateEventWithCreator(clientID string, eventID string, groupID string, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	event := model.Event{
		ClientID:      clientID,
		EventID:       eventID,
		GroupID:       groupID,
		DateCreated:   time.Now().UTC(),
		ToMembersList: toMemberList,
		Creator:       creator,
	}
	_, err := sa.db.events.InsertOne(event)
	if err != nil {
		return nil, err
	}

	if err == nil {
		sa.resetGroupUpdatedDate(clientID, groupID)
	}

	return &event, err
}

// UpdateEvent updates a group event
func (sa *Adapter) UpdateEvent(clientID string, _ *model.User, eventID string, groupID string, toMemberList []model.ToMember) error {

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
	_, err := sa.db.events.UpdateOne(filter, change, nil)
	if err == nil {
		sa.resetGroupUpdatedDate(clientID, groupID)
	}
	return err
}

//DeleteEvent deletes a group event
func (sa *Adapter) DeleteEvent(clientID string, current *model.User, eventID string, groupID string) error {
	filter := bson.D{primitive.E{Key: "event_id", Value: eventID},
		primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID}}
	result, err := sa.db.events.DeleteOne(filter, nil)
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

	sa.resetGroupUpdatedDate(clientID, groupID)

	return nil
}

func (sa *Adapter) findAdminsCount(sessionContext mongo.SessionContext, groupID string) (*int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"_id": groupID}},
		{"$unwind": "$members"},
		{"$group": bson.M{"_id": "$members.status", "count": bson.M{"$sum": 1}}},
	}
	var result []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}
	err := sa.db.groups.AggregateWithContext(sessionContext, pipeline, &result, nil)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result) == 0 {
		//no data - return 0
		noDataCount := 0
		return &noDataCount, nil
	}

	for _, item := range result {
		if item.ID == "admin" {
			return &item.Count, nil
		}
	}
	//no data - return 0
	noDataCount := 0
	return &noDataCount, nil
}

// FindPosts Retrieves posts for a group
func (sa *Adapter) FindPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {

	group, errGr := sa.FindGroup(clientID, groupID)
	if group == nil {
		if errGr != nil {
			log.Printf("unable to find group with id %s: %s", groupID, errGr)
		} else {
			log.Printf("group does not exists %s", groupID)
		}
		return nil, errGr
	}

	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "group_id", Value: groupID},
	}

	if filterByToMembers {
		filter = append(filter, primitive.E{Key: "$or", Value: []primitive.M{
			primitive.M{"to_members": primitive.Null{}},
			primitive.M{"to_members": primitive.M{"$exists": true, "$size": 0}},
			primitive.M{"to_members.user_id": current.ID},
			primitive.M{"member.user_id": current.ID},
		}})
	}

	if filterPrivatePostsValue != nil {
		filter = append(filter, primitive.E{Key: "private", Value: *filterPrivatePostsValue})
	}

	paging := false
	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{"date_created", -1}})
	} else {
		findOptions.SetSort(bson.D{{"date_created", 1}})
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
			childPosts, err := sa.FindPostsByTopParentID(clientID, current, groupID, *post.ID, true, order)
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
			postMapping[*postID] = list[i]
		}
		for _, post := range list {
			var parentPost *model.Post
			if post.ParentID != nil {
				parentID := post.ParentID
				parentPost = postMapping[*parentID]
				repliesList := parentPost.Replies

				repliesList = append(repliesList, post)
				parentPost.Replies = repliesList
			}
		}
		for _, post := range list {
			if post.ParentID == nil {
				resultList = append(resultList, post)
			}
		}
	}

	return resultList, nil
}

// FindAllUserPosts Retrieves all user posts across all existing groups
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindAllUserPosts(clientID string, userID string) ([]model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "member.user_id", Value: userID},
	}

	var posts []model.Post
	err := sa.db.posts.Find(filter, &posts, nil)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

//FindPost Retrieves a post by groupID and postID
func (sa *Adapter) FindPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return sa.findPostWithContext(clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (sa *Adapter) findPostWithContext(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "_id", Value: postID},
	}

	if filterByToMembers {
		filter = append(filter, primitive.E{Key: "$or", Value: []primitive.M{
			primitive.M{"to_members": primitive.Null{}},
			primitive.M{"to_members": primitive.M{"$exists": true, "$size": 0}},
			primitive.M{"to_members.user_id": *userID},
		}})
	}

	if !skipMembershipCheck && userID != nil {
		group, err := sa.findGroupWithContext(clientID, groupID)
		if group == nil || err != nil || !group.IsGroupAdminOrMember(*userID) {
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
		group, err := sa.FindGroup(clientID, groupID)
		if group == nil || err != nil || !group.IsGroupAdminOrMember(current.ID) {
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
func (sa *Adapter) FindPostsByParentID(clientID string, userID string, groupID string, parentID string, skipMembershipCheck bool, filterByToMembers bool, recursive bool, order *string) ([]*model.Post, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "parent_id", Value: parentID},
	}

	if !skipMembershipCheck {
		group, err := sa.FindGroup(clientID, groupID)
		if group == nil || err != nil || !group.IsGroupAdminOrMember(userID) {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{"date_created", -1}})
	} else {
		findOptions.SetSort(bson.D{{"date_created", 1}})
	}

	var posts []*model.Post
	err := sa.db.posts.Find(filter, &posts, findOptions)
	if err != nil {
		return nil, err
	}

	if recursive {
		if len(posts) > 0 {
			for _, post := range posts {
				childPosts, err := sa.FindPostsByParentID(clientID, userID, groupID, *post.ID, true, filterByToMembers, recursive, order)
				if err == nil && childPosts != nil {
					for _, childPost := range childPosts {
						posts = append(posts, childPost)
					}
				}
			}
		}
	}

	return posts, nil
}

// FindPostsByTopParentID  Retrieves a post by groupID and top parent id
// This method doesn't construct tree hierarchy!
func (sa *Adapter) FindPostsByTopParentID(clientID string, current *model.User, groupID string, topParentID string, skipMembershipCheck bool, order *string) ([]*model.Post, error) {
	filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "top_parent_id", Value: topParentID}}

	if !skipMembershipCheck {
		group, err := sa.FindGroup(clientID, groupID)
		if group == nil || err != nil || !group.IsGroupAdminOrMember(current.ID) {
			return nil, fmt.Errorf("the user is not member or admin of the group")
		}
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.D{{"date_created", -1}})
	} else {
		findOptions.SetSort(bson.D{{"date_created", 1}})
	}

	var posts []*model.Post
	err := sa.db.posts.Find(filter, &posts, findOptions)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// CreatePost Created a post
func (sa *Adapter) CreatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error) {

	group, err := sa.FindGroup(clientID, post.GroupID)
	if group == nil || err != nil || !group.IsGroupAdminOrMember(current.ID) {
		return nil, fmt.Errorf("the user is not member or admin of the group")
	}

	if post.ClientID == nil { // Always required
		post.ClientID = &clientID
	}

	if post.ID == nil { // Always required
		id := uuid.New().String()
		post.ID = &id
	}

	if post.Replies != nil { // This is constructed only for GET all for group
		post.Replies = nil
	}

	if post.ParentID != nil {
		topPost, _ := sa.FindTopPostByParentID(clientID, current, group.ID, *post.ParentID, false)
		if topPost != nil && topPost.ParentID == nil {
			post.TopParentID = topPost.ID
		}
	}

	now := time.Now()
	post.DateCreated = &now
	post.DateUpdated = &now

	group, err = sa.FindGroup(clientID, post.GroupID)
	if group != nil && err == nil && group.IsGroupAdminOrMember(current.ID) {
		name := group.UserNameByID(current.ID) // Workaround due to missing name within the id token
		post.Creator = model.Creator{
			UserID: current.ID,
			Email:  current.Email,
			Name:   *name,
		}
	} else {
		return nil, fmt.Errorf("the user is not member or admin of the group")
	}

	_, err = sa.db.posts.InsertOne(post)

	if err == nil {
		sa.resetGroupUpdatedDate(clientID, post.GroupID)
	}

	return post, err
}

// UpdatePost Updates a post
func (sa *Adapter) UpdatePost(clientID string, userID string, post *model.Post) (*model.Post, error) {

	originalPost, _ := sa.FindPost(clientID, &userID, post.GroupID, *post.ID, true, true)
	if originalPost == nil {
		return nil, fmt.Errorf("unable to find post with id (%s) ", *post.ID)
	}
	if originalPost.Creator.UserID != userID {
		return nil, fmt.Errorf("only creator of the post can update it")
	}

	if post.ClientID == nil { // Always required
		post.ClientID = &clientID
	}

	if post.ID == nil { // Always required
		return nil, fmt.Errorf("Missing id")
	}

	now := time.Now()
	post.DateUpdated = &now

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
			primitive.E{Key: "to_members", Value: post.ToMembersList},
		},
		},
	}
	_, err := sa.db.posts.UpdateOne(filter, update, nil)

	if err == nil {
		sa.resetGroupUpdatedDate(clientID, post.GroupID)
	}

	return post, err
}

// ReportPostAsAbuse Report post as abuse
func (sa *Adapter) ReportPostAsAbuse(clientID string, userID string, group *model.Group, post *model.Post) error {

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

// DeletePost Deletes a post
func (sa *Adapter) DeletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return sa.deletePost(context.Background(), clientID, userID, groupID, postID, force)
}

func (sa *Adapter) deletePost(ctx context.Context, clientID string, userID string, groupID string, postID string, force bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	group, _ := sa.FindGroup(clientID, groupID)
	originalPost, _ := sa.FindPost(clientID, &userID, groupID, postID, true, !group.IsGroupAdmin(userID))
	if originalPost == nil {
		return fmt.Errorf("unable to find post with id (%s) ", postID)
	}

	if !force {
		if group == nil || originalPost == nil || (!group.IsGroupAdmin(userID) && originalPost.Creator.UserID != userID) {
			return fmt.Errorf("only creator of the post or group admin can delete it")
		}
	}

	childPosts, err := sa.FindPostsByParentID(clientID, userID, groupID, postID, true, !group.IsGroupAdmin(userID), false, nil)
	if len(childPosts) > 0 && err == nil {
		for _, post := range childPosts {
			sa.deletePost(ctx, clientID, userID, groupID, *post.ID, true)
		}
	}

	filter := bson.D{primitive.E{Key: "client_id", Value: clientID}, primitive.E{Key: "_id", Value: postID}}

	_, err = sa.db.posts.DeleteOne(filter, nil)

	if err == nil {
		sa.resetGroupUpdatedDate(clientID, groupID)
	}

	return err
}

//resetGroupUpdatedDate set the updated date to the current date time (now)
func (sa *Adapter) resetGroupUpdatedDate(clientID string, id string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// update the group
		filter := bson.D{
			primitive.E{Key: "_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID},
		}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now()},
			}},
		}
		_, err = sa.db.groups.UpdateOneWithContext(sessionContext, filter, update, nil)
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
	return nil
}

// FindAuthmanGroups finds all groups that are associated with Authman
func (sa *Adapter) FindAuthmanGroups(clientID string) ([]model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "authman_enabled", Value: true},
	}

	findOptions := options.Find()

	var list []group
	err := sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	result := make([]model.Group, len(list))
	if list != nil {
		for i, current := range list {
			item := constructGroup(current)
			result[i] = item
		}
	}

	return result, nil
}

// FindAuthmanGroupByKey Finds an Authman group by group long name
func (sa *Adapter) FindAuthmanGroupByKey(clientID string, id string) (*model.Group, error) {
	filter := bson.D{
		primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "id", Value: id},
	}

	findOptions := options.Find()

	var list []group
	err := sa.db.groups.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		first := constructGroup(list[0])
		return &first, nil
	}

	return nil, nil
}

//cacheManagedGroupConfigs caches the managed group configs from the DB
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

//FindManagedGroupConfig finds a managed group config by ID
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

//FindManagedGroupConfigs finds all managed group configs for a specified clientID
func (sa *Adapter) FindManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	return sa.getCachedManagedGroupConfigs(clientID)
}

//InsertManagedGroupConfig inserts a new managed group config
func (sa *Adapter) InsertManagedGroupConfig(config model.ManagedGroupConfig) error {
	_, err := sa.db.managedGroupConfigs.InsertOne(config)
	if err != nil {
		return err
	}

	return nil
}

//UpdateManagedGroupConfig updates an existing managed group config
func (sa *Adapter) UpdateManagedGroupConfig(config model.ManagedGroupConfig) error {
	filter := bson.M{"_id": config, "client_id": config.ClientID}
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

//DeleteManagedGroupConfig deletes an existing managed group config
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

//NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS}
	return &Adapter{db: db}
}

func abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error on aborting a transaction - %s", err)
	}
}

func constructGroup(gr group) model.Group {
	id := gr.ID
	clientID := gr.ClientID
	category := gr.Category
	title := gr.Title
	privacy := gr.Privacy
	hiddenForSearch := gr.HiddenForSearch
	description := gr.Description
	imageURL := gr.ImageURL
	webURL := gr.WebURL
	tags := gr.Tags
	membershipQuestions := gr.MembershipQuestions
	authmanEnabled := gr.AuthmanEnabled
	authmanGroup := gr.AuthmanGroup
	onlyAdminsCanCreatePolls := gr.OnlyAdminsCanCreatePolls
	blockNewMembershipRequests := gr.BlockNewMembershipRequests
	canJoinAutomatically := gr.CanJoinAutomatically
	attendanceGroup := gr.AttendanceGroup

	dateCreated := gr.DateCreated
	dateUpdated := gr.DateUpdated

	members := make([]model.Member, len(gr.Members))
	for i, current := range gr.Members {
		members[i] = constructMember(current)
	}

	return model.Group{ID: id, ClientID: clientID, Category: category, Title: title, Privacy: privacy,
		HiddenForSearch: hiddenForSearch, Description: description, ImageURL: imageURL, WebURL: webURL,
		Tags: tags, MembershipQuestions: membershipQuestions, DateCreated: dateCreated, DateUpdated: dateUpdated,
		Members: members, AuthmanEnabled: authmanEnabled, AuthmanGroup: authmanGroup,
		OnlyAdminsCanCreatePolls: onlyAdminsCanCreatePolls, BlockNewMembershipRequests: blockNewMembershipRequests,
		CanJoinAutomatically: canJoinAutomatically, AttendanceGroup: attendanceGroup,
	}
}

func constructMember(member member) model.Member {
	id := member.ID
	userID := member.UserID
	externalID := member.ExternalID
	netID := member.NetID
	name := member.Name
	email := member.Email
	photoURL := member.PhotoURL
	status := member.Status
	rejectReason := member.RejectReason
	dateCreated := member.DateCreated
	dateUpdated := member.DateUpdated
	dateAttended := member.DateAttended

	memberAnswers := make([]model.MemberAnswer, len(member.MemberAnswers))
	for i, current := range member.MemberAnswers {
		memberAnswers[i] = model.MemberAnswer{Question: current.Question, Answer: current.Answer}
	}

	return model.Member{ID: id, UserID: userID, ExternalID: externalID, NetID: netID, Name: name, Email: email, PhotoURL: photoURL,
		Status: status, RejectReason: rejectReason, DateCreated: dateCreated, DateUpdated: dateUpdated, MemberAnswers: memberAnswers,
		DateAttended: dateAttended,
	}
}

type storageListener struct {
	adapter *Adapter
	DefaultListenerImpl
}

func (sl *storageListener) OnManagedGroupConfigsChanged() {
	sl.adapter.cacheManagedGroupConfigs()
}

//Listener  listens for change data storage events
type Listener interface {
	OnManagedGroupConfigsChanged()
}

//DefaultListenerImpl default listener implementation
type DefaultListenerImpl struct{}

//OnManagedGroupConfigsChanged notifies managed group configs have been updated
func (d *DefaultListenerImpl) OnManagedGroupConfigsChanged() {}
