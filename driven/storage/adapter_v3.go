package storage

import (
	"fmt"
	"groups/core/model"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindGroupsV3 finds groups with filter
func (sa *Adapter) FindGroupsV3(clientID string, filter *model.GroupsFilter) ([]model.Group, error) {
	groupIDs := []string{}
	var userID *string
	var memberships model.MembershipCollection

	if filter != nil && filter.MemberUserID == nil && filter.MemberExternalID != nil {
		var user model.User
		err := sa.db.users.Find(bson.D{
			{"client_id", clientID},
			{"_id", filter.MemberExternalID},
		}, &user, nil)
		if err != nil {
			userID = &user.ID
		}
	}
	if userID == nil && filter != nil && filter.MemberUserID == nil && filter.MemberID != nil {
		membership, _ := sa.FindGroupMembershipByID(clientID, *filter.MemberID)
		if membership != nil {
			userID = &membership.UserID
			groupIDs = append(groupIDs, membership.GroupID)
		}
	}

	if filter.MemberUserID != nil {
		// find group memberships
		memberships, err := sa.FindGroupMemberships(clientID, &model.MembershipFilter{})
		if err != nil {
			return nil, err
		}

		for _, membership := range memberships.Items {
			groupIDs = append(groupIDs, membership.GroupID)
		}
	}

	groupFilter := bson.D{primitive.E{Key: "client_id", Value: clientID}}
	if userID != nil {
		innerOrFilter := []bson.M{
			{"_id": bson.M{"$in": groupIDs}},
			{"privacy": bson.M{"$ne": "private"}},
		}

		if filter.Title != nil {
			innerOrFilter = append(innerOrFilter, primitive.M{"$and": []primitive.M{
				{"title": *filter.Title},
				{"hidden_for_search": false},
			}})
		}

		orFilter := primitive.E{Key: "$or", Value: innerOrFilter}

		groupFilter = append(groupFilter, orFilter)
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

	findOptions := options.Find()
	if filter.Order == nil || "asc" == *filter.Order {
		findOptions.SetSort(bson.D{
			{"category", 1},
			{"title", 1},
		})
	} else if filter.Order != nil && "desc" == *filter.Order {
		findOptions.SetSort(bson.D{
			{"category", -1},
			{"title", -1},
		})
	}
	if filter.Limit != nil {
		findOptions.SetLimit(*filter.Limit)
	}
	if filter.Offset != nil {
		findOptions.SetSkip(*filter.Offset)
	}

	var list []model.Group
	err := sa.db.groups.Find(groupFilter, &list, findOptions)
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

// FindGroupMemberships finds the group membership for a given group
func (sa *Adapter) FindGroupMemberships(clientID string, filter *model.MembershipFilter) (model.MembershipCollection, error) {

	matchFilter := bson.D{
		bson.E{"client_id", clientID},
	}
	if len(filter.GroupIDs) > 0 {
		matchFilter = append(matchFilter, bson.E{"group_id", bson.M{"$in": filter.GroupIDs}})
	}
	if filter.ID != nil {
		matchFilter = append(matchFilter, bson.E{"_id", *filter.ID})
	}
	if filter.UserID != nil {
		matchFilter = append(matchFilter, bson.E{"user_id", *filter.UserID})
	}
	if filter.NetID != nil {
		matchFilter = append(matchFilter, bson.E{"net_id", *filter.NetID})
	}
	if filter.ExternalID != nil {
		matchFilter = append(matchFilter, bson.E{"external_id", *filter.ExternalID})
	}
	if filter.Statuses != nil {
		matchFilter = append(matchFilter, bson.E{"status", bson.D{{"$in", filter.Statuses}}})
	}
	if filter.Name != nil {
		matchFilter = append(matchFilter, bson.E{"name", primitive.Regex{fmt.Sprintf(`%s`, *filter.Name), "i"}})
	}

	findOptions := options.FindOptions{
		Sort: bson.D{
			{"members.status", 1},
			{"members.name", 1},
		},
	}
	if filter.Offset != nil {
		findOptions.Skip = filter.Offset
	}
	if filter.Limit != nil {
		findOptions.Limit = filter.Limit
	}

	var result []model.GroupMembership
	err := sa.db.groupMemberships.Find(matchFilter, &result, &findOptions)
	return model.MembershipCollection{Items: result}, err
}

// FindGroupMembership finds the group membership for a given user and group
func (sa *Adapter) FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	filter := bson.M{"client_id": clientID, "group_id": groupID, "user_id": userID}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOne(filter, &result, nil)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// FindGroupMembershipByID finds the group membership by id
func (sa *Adapter) FindGroupMembershipByID(clientID string, id string) (*model.GroupMembership, error) {
	filter := bson.M{"client_id": clientID, "_id": id}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOne(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// FindUserGroupMemberships finds the group memberships for a given user
func (sa *Adapter) FindUserGroupMemberships(clientID string, userID string) (model.MembershipCollection, error) {
	filter := bson.M{"client_id": clientID, "user_id": userID}

	var result []model.GroupMembership
	err := sa.db.groupMemberships.Find(filter, &result, nil)

	return model.MembershipCollection{Items: result}, err
}

// CreateMissingGroupMembership creates a group membership if it does not exist by external ID
func (sa *Adapter) CreateMissingGroupMembership(membership *model.GroupMembership) error {
	transaction := func(context TransactionContext) error {
		filter := bson.M{"client_id": membership.ClientID, "group_id": membership.GroupID, "external_id": membership.ExternalID}

		var result []model.GroupMembership
		err := sa.db.groupMemberships.FindWithContext(context, filter, &result, nil)
		if err != nil {
			return err
		}
		if len(result) == 0 {
			if len(membership.ID) > 0 {
				membership.ID = uuid.NewString()
			}
			_, err = sa.db.groupMemberships.InsertOneWithContext(context, membership)
			return err
		}
		return nil
	}

	return sa.PerformTransaction(transaction)
}

// SaveGroupMembershipByExternalID creates or updates a group membership for a given external ID
func (sa *Adapter) SaveGroupMembershipByExternalID(clientID string, groupID string, externalID string, userID *string, status *string, admin *bool,
	email *string, name *string, memberAnswers []model.MemberAnswer, syncID *string) (*model.GroupMembership, error) {

	now := time.Now()

	filter := bson.M{"client_id": clientID, "group_id": groupID, "external_id": externalID}

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
	if admin != nil {
		update["admin"] = *admin
	}
	if syncID != nil {
		update["sync_id"] = *syncID
	}

	onInsert := bson.M{"_id": uuid.NewString(), "member_answers": memberAnswers, "date_created": now}

	upsert := true
	returnDoc := options.After
	opts := options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &returnDoc}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOneAndUpdate(filter, bson.M{"$set": update, "$setOnInsert": onInsert}, &result, &opts)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteGroupMembership deletes a group membership
func (sa *Adapter) DeleteGroupMembership(clientID string, userID string, groupID string) error {
	filter := bson.M{"client_id": clientID, "group_id": groupID, "user_id": userID}

	result, err := sa.db.groupMemberships.DeleteOne(filter, nil)
	if err != nil {
		return err
	}

	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occurred while deleting group membership for client_id=%s group_id=%s user_id=%s: %v", clientID, groupID, userID, err)
	}

	return nil
}

// DeleteUnsyncedGroupMemberships deletes group memberships that do not exist in the latest sync
func (sa *Adapter) DeleteUnsyncedGroupMemberships(clientID string, groupID string, syncID string, admin *bool) (int64, error) {
	filter := bson.M{"client_id": clientID, "group_id": groupID, "sync_id": bson.M{"$ne": syncID}}
	if admin != nil {
		if *admin {
			filter["admin"] = true
		} else {
			filter["admin"] = bson.M{"$ne": true}
		}
	}

	result, err := sa.db.groupMemberships.DeleteMany(filter, nil)
	if err != nil {
		return 0, err
	}

	deletedCount := result.DeletedCount
	return deletedCount, nil
}

// UpdateGroupSyncTimes updates a group uses group membership
func (sa *Adapter) UpdateGroupSyncTimes(context TransactionContext, clientID string, group *model.Group) error {
	filter := bson.D{primitive.E{Key: "_id", Value: group.ID}, primitive.E{Key: "client_id", Value: clientID}}
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
func (sa Adapter) GetGroupMembershipStats(clientID string, groupID string) (*model.GroupStats, error) {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{
			{"group_id", groupID},
			{"client_id", clientID},
		}}},
		bson.D{
			{"$facet",
				bson.D{
					{"total_count",
						bson.A{
							bson.D{{"$match", bson.D{{"_id", bson.D{{"$exists", true}}}}}},
							bson.D{{"$count", "total_count"}},
						},
					},
					{"admins_count",
						bson.A{
							bson.D{{"$match", bson.D{{"admin", true}}}},
							bson.D{{"$count", "admins_count"}},
						},
					},
					{"member_count",
						bson.A{
							bson.D{{"$match", bson.D{{"status", "member"}}}},
							bson.D{{"$count", "member_count"}},
						},
					},
					{"pending_count",
						bson.A{
							bson.D{{"$match", bson.D{{"status", "pending"}}}},
							bson.D{{"$count", "pending_count"}},
						},
					},
					{"rejected_count",
						bson.A{
							bson.D{{"$match", bson.D{{"status", "rejected"}}}},
							bson.D{{"$count", "rejected_count"}},
						},
					},
					{"attendance_count",
						bson.A{
							bson.D{{"$match", bson.D{{"date_attended", bson.D{
								{"$exists", true},
								{"$ne", nil},
							}}}}},
							bson.D{{"$count", "attendance_count"}},
						},
					},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"total_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$total_count.total_count",
									0,
								},
							},
						},
					},
					{"admins_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$admins_count.admins_count",
									0,
								},
							},
						},
					},
					{"member_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$member_count.member_count",
									0,
								},
							},
						},
					},
					{"pending_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$pending_count.pending_count",
									0,
								},
							},
						},
					},
					{"rejected_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$rejected_count.rejected_count",
									0,
								},
							},
						},
					},
					{"attendance_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
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
	err := sa.db.groupMemberships.Aggregate(pipeline, &stats, nil)
	if err != nil {
		return nil, err
	}

	if len(stats) > 0 {
		stat := stats[0]
		stat.MemberCount -= stat.AdminsCount
		return &stat, err
	}
	return nil, nil
}
