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
	"errors"
	"fmt"
	"groups/core/model"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindGroupsV3 finds groups with filter
func (sa *Adapter) FindGroupsV3(context TransactionContext, orgID string, filter model.GroupsFilter) ([]model.Group, error) {

	var err error
	var memberships model.MembershipCollection
	findOptions := options.Find()

	groupFilter, err := sa.buildGroupsFilter(orgID, context, filter)
	if err != nil {
		return nil, err
	}

	if filter.Order != nil && "desc" == *filter.Order {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: -1},
		})
	} else {
		findOptions.SetSort(bson.D{
			{Key: "title", Value: 1},
		})
	}

	var list []model.Group
	err = sa.db.groups.FindWithContext(context, groupFilter, &list, findOptions)
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

func (sa *Adapter) buildGroupsFilter(orgID string, context TransactionContext, filter model.GroupsFilter) (bson.D, error) {
	var groupIDs []string

	groupFilter := bson.D{primitive.E{Key: "org_id", Value: orgID}}

	groupIDMap := map[string]bool{}
	if len(filter.GroupIDs) > 0 {
		for _, groupID := range filter.GroupIDs {
			groupIDs = append(groupIDs, groupID)
			groupIDMap[groupID] = true
		}
	}

	// Credits to Ryan Oberlander suggest
	if filter.MemberUserID != nil || filter.MemberID != nil || filter.MemberExternalID != nil {
		// find group memberships
		memberships, err := sa.FindGroupMembershipsWithContext(context, orgID, model.MembershipFilter{
			ID:         filter.MemberID,
			UserID:     filter.MemberUserID,
			ExternalID: filter.MemberExternalID,
		})
		if err != nil {
			return nil, err
		}

		for _, membership := range memberships.Items {
			if len(groupIDMap) == 0 || !groupIDMap[membership.GroupID] {
				groupIDs = append(groupIDs, membership.GroupID)
				groupIDMap[membership.GroupID] = true
			}
		}
	}

	if len(groupIDs) > 0 {
		groupFilter = append(groupFilter, primitive.E{Key: "_id", Value: primitive.M{"$in": groupIDs}})
	}
	if len(filter.Tags) > 0 {
		groupFilter = append(groupFilter, primitive.E{Key: "tags", Value: primitive.M{"$in": filter.Tags}})
	}
	if filter.Category != nil {
		groupFilter = append(groupFilter, primitive.E{Key: "category", Value: *filter.Category})
	}
	if filter.Title != nil {
		groupFilter = append(groupFilter, primitive.E{Key: "title", Value: primitive.Regex{Pattern: *filter.Title, Options: "i"}})
	}
	if filter.Privacy != nil {
		groupFilter = append(groupFilter, primitive.E{Key: "privacy", Value: *filter.Privacy})
	}
	if filter.ResearchOpen != nil {
		if *filter.ResearchOpen {
			groupFilter = append(groupFilter, primitive.E{Key: "research_open", Value: true})
		} else {
			groupFilter = append(groupFilter, primitive.E{Key: "research_open", Value: primitive.M{"$ne": true}})
		}
	}

	if filter.ResearchGroup != nil {
		if *filter.ResearchGroup {
			groupFilter = append(groupFilter, primitive.E{Key: "research_group", Value: true})
		} else {
			groupFilter = append(groupFilter, primitive.E{Key: "research_group", Value: primitive.M{"$ne": true}})
		}
	}
	if filter.ResearchAnswers != nil {
		for outerKey, outerValue := range filter.ResearchAnswers {
			for innerKey, innerValue := range outerValue {
				groupFilter = append(groupFilter, bson.E{
					Key: "$or", Value: []bson.M{
						{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$elemMatch": bson.M{"$in": innerValue}}},
						{fmt.Sprintf("research_profile.%s.%s", outerKey, innerKey): bson.M{"$exists": false}},
					},
				})
			}
		}
	}
	if filter.Hidden != nil {
		if *filter.Hidden {
			groupFilter = append(groupFilter, primitive.E{Key: "hidden_for_search", Value: *filter.Hidden})
		} else {
			groupFilter = append(groupFilter, primitive.E{Key: "hidden_for_search", Value: primitive.M{"$ne": true}})
		}
	}

	if filter.Attributes != nil {
		innerGroupFilters := []bson.M{}
		for key, value := range filter.Attributes {
			if reflect.TypeOf(value).Kind() != reflect.Slice {
				innerGroupFilters = append(innerGroupFilters, bson.M{fmt.Sprintf("attributes.%s", key): value})
			} else {
				orSubCriterias := []bson.M{}
				var entryList []interface{} = value.([]interface{})
				for _, entry := range entryList {
					orSubCriterias = append(orSubCriterias, bson.M{fmt.Sprintf("attributes.%s", key): entry})
				}
				innerGroupFilters = append(innerGroupFilters, bson.M{"$or": orSubCriterias})
			}
		}
		if len(innerGroupFilters) > 0 {
			groupFilter = append(groupFilter, bson.E{
				Key: "$and", Value: innerGroupFilters,
			})
		}
	}

	if filter.DaysInactive != nil {
		pastTime := time.Now().Add(-time.Duration(*filter.DaysInactive) * 24 * time.Hour)
		groupFilter = append(groupFilter, primitive.E{Key: "date_updated", Value: bson.M{"$lt": pastTime}})
	}

	return groupFilter, nil
}

// CalculateGroupFilterStats Generates the stats for a given filter
func (sa *Adapter) CalculateGroupFilterStats(orgID string, current *model.User, filter model.StatsFilter) (*model.StatsResult, error) {
	var result *model.StatsResult
	err := sa.PerformTransaction(func(ctx TransactionContext) error {
		baseFilter, err := sa.buildGroupsFilter(orgID, ctx, filter.BaseFilter)
		if err != nil {
			return err
		}

		pipeline := bson.A{bson.D{{Key: "$match", Value: baseFilter}}}

		subFilters := bson.D{}
		for key, value := range filter.SubFilters {
			innerSubFilter := bson.A{}
			filter, err := sa.buildGroupsFilter(orgID, ctx, value)
			if err != nil {
				return err
			}
			if len(filter) > 0 {
				innerSubFilter = append(innerSubFilter, bson.D{{Key: "$match", Value: filter}})
			}
			innerSubFilter = append(innerSubFilter, bson.D{
				{Key: "$group",
					Value: bson.D{
						{Key: "_id", Value: primitive.Null{}},
						{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
					},
				},
			})
			subFilter := bson.E{Key: key,
				Value: innerSubFilter,
			}
			subFilters = append(subFilters, subFilter)
		}
		pipeline = append(pipeline, bson.D{{Key: "$facet", Value: subFilters}})

		project := bson.D{}
		for key := range filter.SubFilters {
			project = append(project, bson.E{Key: key,
				Value: bson.D{
					{Key: "$ifNull",
						Value: bson.A{
							bson.D{
								{Key: "$arrayElemAt",
									Value: bson.A{
										fmt.Sprintf("$%s", key),
										0,
									},
								},
							},
							bson.D{{Key: "count", Value: 0}},
						},
					},
				},
			})
		}
		pipeline = append(pipeline, bson.D{
			{Key: "$project", Value: project},
		})

		project = bson.D{}
		for key := range filter.SubFilters {
			project = append(project, bson.E{Key: key, Value: fmt.Sprintf("$%s.count", key)})
		}
		pipeline = append(pipeline, bson.D{
			{Key: "$project", Value: project},
		})

		for key := range filter.SubFilters {
			pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: fmt.Sprintf("$%s", key)}}}})
		}

		//str, _ := json.Marshal(pipeline) // To ensure the pipeline is valid
		//fmt.Printf("events stats aggregation pipeline: %s", string(str))

		var aggrResult []map[string]int64
		var statsResult map[string]int64
		err = sa.db.groups.AggregateWithContext(ctx, pipeline, &aggrResult, nil)
		if err != nil {
			return nil
		}

		if len(aggrResult) > 0 {
			statsResult = aggrResult[0]
		} else {
			statsResult = make(map[string]int64)
		}

		for key := range filter.SubFilters {
			if _, ok := statsResult[key]; !ok {
				statsResult[key] = 0 // Ensure all keys are present in the result
			}
		}

		result = &model.StatsResult{
			Stats: statsResult,
		}
		return nil
	})
	if err != nil {
		log.Printf("error calculating group filter stats: %v", err)
		return nil, err
	}

	return result, nil
}

// FindGroupMemberships finds the group membership for a given group
func (sa *Adapter) FindGroupMemberships(orgID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return sa.FindGroupMembershipsWithContext(nil, orgID, filter)
}

// FindGroupMembershipsWithContext finds the group membership for a given group
func (sa *Adapter) FindGroupMembershipsWithContext(ctx TransactionContext, orgID string, filter model.MembershipFilter) (model.MembershipCollection, error) {

	if filter.ID == nil && len(filter.GroupIDs) == 0 && filter.UserID == nil && filter.ExternalID == nil && filter.Name == nil {
		log.Print("The memberships filter requires at least one of the listed filters to be set: ID, GroupsIDs, UserID, ExternalID or Name")
		return model.MembershipCollection{}, fmt.Errorf("the memberships filter requires at least one of the listed filters to be set: ID, GroupsIDs, UserID, ExternalID or Name")
	}

	matchFilter := bson.D{
		bson.E{Key: "org_id", Value: orgID},
	}
	if len(filter.GroupIDs) > 0 {
		matchFilter = append(matchFilter, bson.E{Key: "group_id", Value: bson.M{"$in": filter.GroupIDs}})
	}
	if filter.ID != nil {
		matchFilter = append(matchFilter, bson.E{Key: "_id", Value: *filter.ID})
	}
	if filter.UserID != nil {
		matchFilter = append(matchFilter, bson.E{Key: "user_id", Value: *filter.UserID})
	} else if len(filter.UserIDs) > 0 {
		matchFilter = append(matchFilter, bson.E{Key: "user_id", Value: bson.D{{Key: "$in", Value: filter.UserIDs}}})
	}
	if filter.NetID != nil {
		matchFilter = append(matchFilter, bson.E{Key: "net_id", Value: *filter.NetID})
	} else if len(filter.NetIDs) > 0 {
		matchFilter = append(matchFilter, bson.E{Key: "net_id", Value: bson.D{{Key: "$in", Value: filter.NetIDs}}})
	}
	if filter.ExternalID != nil {
		matchFilter = append(matchFilter, bson.E{Key: "external_id", Value: *filter.ExternalID})
	}
	if filter.Statuses != nil {
		matchFilter = append(matchFilter, bson.E{Key: "status", Value: bson.D{{Key: "$in", Value: filter.Statuses}}})
	}
	if filter.Name != nil {
		matchFilter = append(matchFilter, bson.E{Key: "name", Value: primitive.Regex{Pattern: fmt.Sprintf(`%s`, *filter.Name), Options: "i"}})
	}

	findOptions := options.FindOptions{
		Sort: bson.D{
			{Key: "status", Value: 1},
			{Key: "name", Value: 1},
		},
	}
	if filter.Offset != nil {
		findOptions.Skip = filter.Offset
	}
	if filter.Limit != nil {
		findOptions.Limit = filter.Limit
	}

	var result []model.GroupMembership
	err := sa.db.groupMemberships.FindWithContext(ctx, matchFilter, &result, &findOptions)
	return model.MembershipCollection{Items: result}, err
}

// FindGroupMembership finds the group membership for a given user and group
func (sa *Adapter) FindGroupMembership(orgID string, groupID string, userID string) (*model.GroupMembership, error) {
	return sa.FindGroupMembershipWithContext(nil, orgID, groupID, userID)
}

// FindGroupMembershipWithContext finds the group membership for a given user and group
func (sa *Adapter) FindGroupMembershipWithContext(context TransactionContext, orgID string, groupID string, userID string) (*model.GroupMembership, error) {
	filter := bson.M{"org_id": orgID, "group_id": groupID, "user_id": userID}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// FindGroupMembershipByID finds the group membership by id
func (sa *Adapter) FindGroupMembershipByID(orgID string, id string) (*model.GroupMembership, error) {
	filter := bson.M{"org_id": orgID, "_id": id}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOne(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// FindUserGroupMemberships finds the group memberships for a given user
func (sa *Adapter) FindUserGroupMemberships(orgID string, userID string) (model.MembershipCollection, error) {
	return sa.FindUserGroupMembershipsWithContext(nil, orgID, userID)
}

// FindUserGroupMembershipsWithContext finds the group memberships for a given user with context
func (sa *Adapter) FindUserGroupMembershipsWithContext(ctx TransactionContext, orgID string, userID string) (model.MembershipCollection, error) {
	filter := bson.M{"org_id": orgID, "user_id": userID}

	var result []model.GroupMembership
	err := sa.db.groupMemberships.FindWithContext(ctx, filter, &result, nil)

	return model.MembershipCollection{Items: result}, err
}

// CreatePendingMembership creates a pending membership for a specific group
func (sa *Adapter) CreatePendingMembership(orgID string, user *model.User, group *model.Group, membership *model.GroupMembership) error {
	if membership != nil && group != nil {

		//1. check if the user is already a member of this group - pending or member or admin or rejected
		storedMembership, err := sa.FindGroupMembership(orgID, group.ID, user.ID)
		if err == nil && storedMembership != nil {
			switch storedMembership.Status {
			case "admin":
				return errors.New("the user is an admin for the group")
			case "member":
				return errors.New("the user is a member for the group")
			case "pending":
				return errors.New("the user is pending for the group")
			case "rejected":
				return errors.New("the user is denied for the group")
			default:
				return errors.New("error creating a pending user")
			}
		}

		//2. check if the answers match the group questions
		if len(group.MembershipQuestions) != len(membership.MemberAnswers) {
			return errors.New("member answers mismatch")
		}

		membership.ID = uuid.NewString()
		membership.OrgID = orgID
		membership.GroupID = group.ID
		membership.DateCreated = time.Now().UTC()

		err = sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.groupMemberships.InsertOneWithContext(context, membership)
			if err != nil {
				return err
			}

			return sa.UpdateGroupStats(context, orgID, membership.GroupID, false, true, false, true)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// SingleMembershipOperation wraps single membership operation for possible updates
type SingleMembershipOperation struct {
	OrgID      string
	GroupID    string
	ExternalID string
	UserID     *string
	Status     *string
	Email      *string
	Name       *string
	Answers    []model.MemberAnswer
	SyncID     *string
}

// BulkUpdateGroupMembershipsByExternalID Bulk update with a list of memberships
func (sa *Adapter) BulkUpdateGroupMembershipsByExternalID(orgID string, groupID string, saveOperations []SingleMembershipOperation, updateGroupStats bool) error {
	now := time.Now()

	var updateModels []mongo.WriteModel
	upsert := true
	for _, operation := range saveOperations {
		filter := bson.M{"org_id": operation.OrgID, "group_id": operation.GroupID, "external_id": operation.ExternalID}
		update := bson.M{"date_updated": now}
		if operation.UserID != nil {
			update["user_id"] = *operation.UserID
		}
		if operation.Name != nil {
			update["name"] = *operation.Name
		}
		if operation.Email != nil {
			update["email"] = *operation.Email
		}
		if operation.Status != nil {
			update["status"] = *operation.Status
		}
		if operation.SyncID != nil {
			update["sync_id"] = *operation.SyncID
		}
		onInsert := bson.M{"_id": uuid.NewString(), "member_answers": operation.Answers, "date_created": now}
		updateModels = append(updateModels, &mongo.UpdateOneModel{
			Filter: filter,
			Update: bson.M{"$set": update, "$setOnInsert": onInsert},
			Upsert: &upsert,
		})
	}

	if len(updateModels) > 0 {
		return sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.groupMemberships.BulkWrite(updateModels, nil)
			if err != nil {
				return err
			}

			if updateGroupStats {
				return sa.UpdateGroupStats(context, orgID, groupID, false, false, true, true)
			}

			return nil
		})
	}

	return nil
}

// SaveGroupMembershipByExternalID creates or updates a group membership for a given external ID
func (sa *Adapter) SaveGroupMembershipByExternalID(orgID string, groupID string, externalID string, userID *string, status *string,
	email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string, updateGroupStats bool) (*model.GroupMembership, error) {

	now := time.Now()

	filter := bson.M{"org_id": orgID, "group_id": groupID, "external_id": externalID}

	update := bson.M{"date_updated": now}
	if userID != nil {
		update["user_id"] = *userID
	}
	if name != nil {
		update["name"] = *name
	}
	if email != nil {
		update["email"] = *email
	}
	if status != nil {
		update["status"] = *status
	}
	if syncID != nil {
		update["sync_id"] = *syncID
	}

	var result model.GroupMembership
	err := sa.PerformTransaction(func(context TransactionContext) error {
		onInsert := bson.M{"_id": uuid.NewString(), "member_answers": memberAnswers, "date_created": now}

		upsert := true
		returnDoc := options.After
		opts := options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &returnDoc}

		err := sa.db.groupMemberships.FindOneAndUpdateWithContext(context, filter, bson.M{"$set": update, "$setOnInsert": onInsert}, &result, &opts)
		if err != nil {
			return err
		}

		if updateGroupStats {
			return sa.UpdateGroupStats(context, orgID, groupID, false, false, true, true)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateMembership Created a member to a group
func (sa *Adapter) CreateMembership(orgID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	if group != nil {

		if len(membership.UserID) == 0 && len(membership.ExternalID) == 0 {
			log.Printf("error: storage.CreateMembership() - expected user_id or external_id")
			return fmt.Errorf("expected user_id or external_id")
		}

		existingMembership, err := sa.FindGroupMembership(orgID, group.ID, current.ID)
		if err != nil || existingMembership == nil || !existingMembership.IsAdmin() {
			log.Printf("error: storage.CreateMembership() - current user is not admin of the group")
			return fmt.Errorf("current user is not admin of the group")
		}

		existingMembership, _ = sa.FindGroupMembership(orgID, group.ID, membership.UserID)
		if existingMembership != nil {
			log.Printf("error: storage.CreateMembership() - member of group '%s' with user id %s already exists", group.Title, membership.UserID)
			return fmt.Errorf("member of group '%s' with user id %s already exists", group.Title, membership.UserID)
		}

		existingMembership, _ = sa.FindGroupMembership(orgID, group.ID, membership.ExternalID)
		if existingMembership != nil {
			log.Printf("error: storage.CreateMembership() - member of group '%s' with external id %s already exists", group.Title, membership.ExternalID)
			return fmt.Errorf("member of group '%s' with external id %s already exists", group.Title, membership.ExternalID)
		}

		membership.ID = uuid.NewString()
		membership.OrgID = orgID
		membership.GroupID = group.ID
		membership.DateCreated = time.Now()
		membership.MemberAnswers = group.CreateMembershipEmptyAnswers()

		return sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.groupMemberships.InsertOne(membership)
			if err != nil {
				return err
			}

			return sa.UpdateGroupStats(context, orgID, membership.GroupID, false, true, false, true)
		})
	}

	return nil
}

// CreateMemberships Created multiple members to a group
func (sa *Adapter) CreateMemberships(context TransactionContext, orgID string, current *model.User, group *model.Group, memberships []model.GroupMembership) error {
	now := time.Now()
	upsert := true

	var objects []mongo.WriteModel
	for index := range memberships {
		if memberships[index].UserID != "" && memberships[index].ExternalID != "" && memberships[index].Email != "" && memberships[index].Status != "" {
			objects = append(objects, &mongo.UpdateOneModel{
				Filter: bson.M{
					"group_id": group.ID,
					"user_id":  memberships[index].UserID,
				},
				Update: bson.M{
					"$setOnInsert": bson.M{
						"_id":          uuid.NewString(),
						"org_id":       orgID,
						"group_id":     group.ID,
						"user_id":      memberships[index].UserID,
						"external_id":  memberships[index].ExternalID,
						"net_id":       memberships[index].NetID,
						"name":         memberships[index].Name,
						"email":        memberships[index].Email,
						"date_created": now,
					},
					"$set": bson.M{
						"date_updated": now,
						"status":       memberships[index].Status,
					},
				},
				Upsert: &upsert,
			})
		}
	}

	if len(objects) > 0 {
		_, err := sa.db.groupMemberships.BulkWriteWithContext(context, objects, nil)
		return err
	}

	return nil
}

// ApplyMembershipApproval applies a membership approval
func (sa *Adapter) ApplyMembershipApproval(orgID string, membershipID string, approve bool, rejectReason string) (*model.GroupMembership, error) {
	var membership model.GroupMembership
	err := sa.PerformTransaction(func(context TransactionContext) error {
		status := "rejected"
		if approve {
			status = "member"
		}

		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "org_id", Value: orgID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: status},
				primitive.E{Key: "reject_reason", Value: rejectReason},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		after := options.After
		err := sa.db.groupMemberships.FindOneAndUpdateWithContext(context, filter, update, &membership, &options.FindOneAndUpdateOptions{ReturnDocument: &after})
		if err != nil {
			return err
		}

		sa.UpdateGroupStats(context, orgID, membership.GroupID, false, true, false, true)

		return err
	})
	return &membership, err
}

// UpdateMembership updates a membership
func (sa *Adapter) UpdateMembership(orgID string, _ *model.User, membershipID string, membership *model.GroupMembership) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "org_id", Value: orgID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: membership.Status},
				primitive.E{Key: "reject_reason", Value: membership.RejectReason},
				primitive.E{Key: "date_attended", Value: membership.DateAttended},
				primitive.E{Key: "notifications_preferences", Value: membership.NotificationsPreferences},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		var membership model.GroupMembership
		err := sa.db.groupMemberships.FindOneAndUpdateWithContext(context, filter, update, &membership, nil)
		if err != nil {
			return err
		}

		return sa.UpdateGroupStats(context, orgID, membership.GroupID, false, true, false, true)
	})

}

// UpdateMemberships Updates multiple memberships for userids in a group
func (sa *Adapter) UpdateMemberships(orgID string, user *model.User, groupID string, operation model.MembershipMultiUpdate) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.D{
			primitive.E{Key: "group_id", Value: groupID},
			primitive.E{Key: "user_id", Value: bson.M{"$in": operation.UserIDs}},
		}
		operarions := bson.D{}
		if operation.Status != nil {
			operarions = append(operarions, primitive.E{Key: "status", Value: *operation.Status})
		}
		if operation.Reason != nil {
			operarions = append(operarions, primitive.E{Key: "reject_reason", Value: *operation.Reason})
		}
		if operation.DateAttended != nil {
			operarions = append(operarions, primitive.E{Key: "date_attended", Value: *operation.DateAttended})
		}
		if len(operarions) > 0 {
			operarions = append(operarions, primitive.E{Key: "date_updated", Value: time.Now()})
			update := bson.D{
				primitive.E{Key: "$set", Value: operarions},
			}
			_, err := sa.db.groupMemberships.UpdateManyWithContext(context, filter, update, nil)
			if err != nil {
				return err
			}

			return sa.UpdateGroupStats(context, orgID, groupID, false, true, false, true)
		}
		return nil
	})
}

// DeleteMembership deletes a member membership from a specific group
func (sa *Adapter) DeleteMembership(orgID string, groupID string, userID string) error {
	return sa.DeleteMembershipWithContext(nil, orgID, groupID, userID, false)
}

// DeleteMembershipWithContext deletes a member membership from a specific group with context
func (sa *Adapter) DeleteMembershipWithContext(ctx TransactionContext, orgID string, groupID string, userID string, forsed bool) error {

	deleteWrapper := func(context TransactionContext) error {
		currentMembership, _ := sa.FindGroupMembershipWithContext(context, orgID, groupID, userID)
		if currentMembership != nil {

			if !forsed {
				if currentMembership.IsAdmin() {
					adminMemberships, _ := sa.FindGroupMembershipsWithContext(context, orgID, model.MembershipFilter{
						GroupIDs: []string{groupID},
						Statuses: []string{"admin"},
					})
					if len(adminMemberships.Items) <= 1 {
						log.Printf("sa.DeleteMembership() - there must be at least two admins in order to delete ")
						return fmt.Errorf("there must be at least two admins in order to delete ")
					}
				}
			}

			filter := bson.D{
				primitive.E{Key: "group_id", Value: groupID},
				primitive.E{Key: "user_id", Value: userID},
				primitive.E{Key: "org_id", Value: orgID},
			}
			_, err := sa.db.groupMemberships.DeleteOneWithContext(context, filter, nil)
			if err != nil {
				log.Printf("error deleting membership - %s", err)
				return err
			}
			return sa.UpdateGroupStats(context, orgID, groupID, false, true, false, true)
		}
		return nil
	}

	if ctx != nil {
		return deleteWrapper(ctx)
	}
	return sa.PerformTransaction(func(transactionContext TransactionContext) error {
		return deleteWrapper(transactionContext)
	})
}

// DeleteMembershipByID deletes a membership by ID
func (sa *Adapter) DeleteMembershipByID(orgID string, current *model.User, membershipID string) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		membership, err := sa.FindGroupMembershipByID(orgID, membershipID)
		if err != nil || membership == nil {
			return fmt.Errorf("membership %s not found", membershipID)
		}

		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "org_id", Value: orgID}}
		_, err = sa.db.groupMemberships.DeleteManyWithContext(context, filter, nil)
		if err != nil {
			return err
		}

		return sa.UpdateGroupStats(context, orgID, membership.GroupID, false, true, false, true)
	})
}

// DeleteUnsyncedGroupMemberships deletes group memberships that do not exist in the latest sync
func (sa *Adapter) DeleteUnsyncedGroupMemberships(orgID string, groupID string, syncID string) (int64, error) {
	var deletedCount int64 = 0
	err := sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.M{
			"org_id":   orgID,
			"group_id": groupID,
			"sync_id":  bson.M{"$ne": syncID},
			"status":   bson.M{"$ne": "admin"},
		}

		result, err := sa.db.groupMemberships.DeleteMany(filter, nil)
		if err != nil {
			return err
		}

		deletedCount = result.DeletedCount
		if deletedCount > 0 {
			return sa.UpdateGroupStats(context, orgID, groupID, false, false, true, true)
		}

		return nil
	})
	return deletedCount, err
}

// UpdateGroupSyncTimes updates a group uses group membership
func (sa *Adapter) UpdateGroupSyncTimes(context TransactionContext, orgID string, group *model.Group) error {

	filter := bson.D{primitive.E{Key: "_id", Value: group.ID}, primitive.E{Key: "org_id", Value: orgID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "sync_start_time", Value: group.SyncStartTime},
			primitive.E{Key: "sync_end_time", Value: group.SyncEndTime},
		}},
	}

	res, err := sa.db.groups.UpdateOneWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("group could not be found for id: %s", group.ID)
	}

	return nil
}

// GetGroupMembershipStats Retrieves group membership stats
func (sa Adapter) GetGroupMembershipStats(context TransactionContext, orgID string, groupID string) (*model.GroupStats, error) {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "group_id", Value: groupID},
			{Key: "org_id", Value: orgID},
		}}},
		bson.D{
			{Key: "$facet",
				Value: bson.D{
					{Key: "total_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{
								{Key: "status", Value: bson.D{{Key: "$in", Value: []string{"member", "admin"}}}},
							}}},
							bson.D{{Key: "$count", Value: "total_count"}},
						},
					},
					{Key: "admins_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "admin"}}}},
							bson.D{{Key: "$count", Value: "admins_count"}},
						},
					},
					{Key: "member_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "member"}}}},
							bson.D{{Key: "$count", Value: "member_count"}},
						},
					},
					{Key: "members_added_last_24hours",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{
								{Key: "status", Value: "member"},
								{Key: "date_created", Value: bson.M{"$gt": time.Now().Add(-24 * time.Hour)}},
							}}},
							bson.D{{Key: "$count", Value: "members_added_last_24hours"}},
						},
					},
					{Key: "pending_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "pending"}}}},
							bson.D{{Key: "$count", Value: "pending_count"}},
						},
					},
					{Key: "rejected_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "rejected"}}}},
							bson.D{{Key: "$count", Value: "rejected_count"}},
						},
					},
					{Key: "attendance_count",
						Value: bson.A{
							bson.D{{Key: "$match", Value: bson.D{{Key: "date_attended", Value: bson.D{
								{Key: "$exists", Value: true},
								{Key: "$ne", Value: nil},
							}}}}},
							bson.D{{Key: "$count", Value: "attendance_count"}},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "total_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$total_count.total_count",
									0,
								},
							},
						},
					},
					{Key: "admins_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$admins_count.admins_count",
									0,
								},
							},
						},
					},
					{Key: "member_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$member_count.member_count",
									0,
								},
							},
						},
					},
					{Key: "members_added_last_24hours",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$members_added_last_24hours.members_added_last_24hours",
									0,
								},
							},
						},
					},
					{Key: "pending_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$pending_count.pending_count",
									0,
								},
							},
						},
					},
					{Key: "rejected_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$rejected_count.rejected_count",
									0,
								},
							},
						},
					},
					{Key: "attendance_count",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$attendance_count.attendance_count",
									0,
								},
							},
						},
					},
				},
			},
		},
	}

	var stats []model.GroupStats
	err := sa.db.groupMemberships.AggregateWithContext(context, pipeline, &stats, nil)
	if err != nil {
		return nil, err
	}

	if len(stats) > 0 {
		stat := stats[0]
		return &stat, err
	}
	return nil, nil
}

// FindAllGroupsUnsecured finds all groups
func (sa *Adapter) FindAllGroupsUnsecured() ([]model.Group, error) {

	var list []model.Group
	err := sa.db.groups.Find(bson.D{}, &list, nil)
	if err != nil {
		return nil, err
	}

	return list, err
}

// FindAllPostsUnsecured finds all posts
func (sa *Adapter) FindAllPostsUnsecured() ([]model.Post, error) {

	var list []model.Post
	err := sa.db.posts.Find(bson.D{}, &list, nil)
	if err != nil {
		return nil, err
	}

	return list, err
}
