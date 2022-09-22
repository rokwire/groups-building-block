package storage

import (
	"context"
	"errors"
	"fmt"
	"groups/core/model"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindGroupsV3 finds groups with filter
func (sa *Adapter) FindGroupsV3(clientID string, filter *model.GroupsFilter) ([]model.Group, error) {
	var groupIDs []string
	var err error
	var userID *string
	var memberships model.MembershipCollection

	groupFilter := bson.D{primitive.E{Key: "client_id", Value: clientID}}
	findOptions := options.Find()

	if filter != nil {
		groupIDMap := map[string]bool{}
		if len(filter.GroupIDs) > 0 {
			for _, groupID := range filter.GroupIDs {
				groupIDMap[groupID] = true
			}
		}

		if filter.MemberUserID == nil && filter.MemberExternalID != nil {
			var user model.User
			err := sa.db.users.Find(bson.D{
				{"client_id", clientID},
				{"external_id", filter.MemberExternalID},
			}, &user, nil)
			if err != nil {
				userID = &user.ID
			}
		}
		if userID == nil && filter.MemberUserID == nil && filter.MemberID != nil {
			membership, _ := sa.FindGroupMembershipByID(clientID, *filter.MemberID)
			if membership != nil {
				memberships = model.MembershipCollection{
					Items: []model.GroupMembership{*membership},
				}
				userID = &membership.UserID
				if len(groupIDMap) == 0 || groupIDMap[membership.GroupID] {
					groupIDs = append(groupIDs, membership.GroupID)
				}
			}
		}

		if filter.MemberUserID != nil {
			// find group memberships
			memberships, err = sa.FindGroupMemberships(clientID, model.MembershipFilter{
				UserID: filter.MemberUserID,
			})
			if err != nil {
				return nil, err
			}

			for _, membership := range memberships.Items {
				if len(groupIDMap) == 0 || groupIDMap[membership.GroupID] {
					groupIDs = append(groupIDs, membership.GroupID)
				}
			}
		}

		if filter.MemberExternalID != nil {
			// find group memberships
			memberships, err = sa.FindGroupMemberships(clientID, model.MembershipFilter{
				ExternalID: filter.MemberExternalID,
			})
			if err != nil {
				return nil, err
			}

			for _, membership := range memberships.Items {
				if len(groupIDMap) == 0 || groupIDMap[membership.GroupID] {
					groupIDs = append(groupIDs, membership.GroupID)
				}
			}
		}

		if len(filter.GroupIDs) > 0 {
			groupFilter = append(groupFilter, primitive.E{Key: "_id", Value: primitive.M{"$in": filter.GroupIDs}})
		}
		if len(groupIDs) > 0 && (filter.MemberID != nil || filter.MemberUserID != nil || filter.MemberExternalID != nil) {
			groupFilter = append(groupFilter, primitive.E{Key: "_id", Value: primitive.M{"$in": groupIDs}})
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
	}

	var list []model.Group
	err = sa.db.groups.Find(groupFilter, &list, findOptions)
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

// FindGroupMemberships finds the group membership for a given group
func (sa *Adapter) FindGroupMemberships(clientID string, filter model.MembershipFilter) (model.MembershipCollection, error) {
	return sa.FindGroupMembershipsWithContext(nil, clientID, filter)
}

// FindGroupMembershipsWithContext finds the group membership for a given group
func (sa *Adapter) FindGroupMembershipsWithContext(ctx TransactionContext, clientID string, filter model.MembershipFilter) (model.MembershipCollection, error) {

	if filter.ID == nil && len(filter.GroupIDs) == 0 && filter.UserID == nil && filter.ExternalID == nil && filter.Name == nil {
		log.Print("The memberships filter requires at least one of the listed filters to be set: ID, GroupsIDs, UserID, ExternalID or Name")
		return model.MembershipCollection{}, fmt.Errorf("the memberships filter requires at least one of the listed filters to be set: ID, GroupsIDs, UserID, ExternalID or Name")
	}

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
	err := sa.db.groupMemberships.FindWithContext(ctx, matchFilter, &result, &findOptions)
	return model.MembershipCollection{Items: result}, err
}

// FindGroupMembership finds the group membership for a given user and group
func (sa *Adapter) FindGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	return sa.FindGroupMembershipWithContext(nil, clientID, groupID, userID)
}

// FindGroupMembershipWithContext finds the group membership for a given user and group
func (sa *Adapter) FindGroupMembershipWithContext(ctx context.Context, clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	filter := bson.M{"client_id": clientID, "group_id": groupID, "user_id": userID}

	var result model.GroupMembership
	err := sa.db.groupMemberships.FindOneWithContext(ctx, filter, &result, nil)
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

// CreatePendingMembership creates a pending membership for a specific group
func (sa *Adapter) CreatePendingMembership(clientID string, user *model.User, group *model.Group, membership *model.GroupMembership) error {
	if membership != nil && group != nil {

		//1. check if the user is already a member of this group - pending or member or admin or rejected
		storedMembership, err := sa.FindGroupMembership(clientID, group.ID, user.ID)
		if err == nil && storedMembership != nil {
			switch storedMembership.Status {
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

		//2. check if the answers match the group questions
		if len(group.MembershipQuestions) != len(membership.MemberAnswers) {
			return errors.New("member answers mismatch")
		}

		membership.ID = uuid.NewString()
		membership.ClientID = clientID
		membership.GroupID = group.ID
		membership.DateCreated = time.Now().UTC()

		err = sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.groupMemberships.InsertOneWithContext(context, membership)
			if err != nil {
				return err
			}

			return sa.resetGroupState(context, clientID, membership.GroupID, true, true)
		})
		if err != nil {
			return err
		}
	}

	return nil
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

		return sa.resetGroupState(context, clientID, groupID, true, true)
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateMembershipUnchecked Created a member to a group
func (sa *Adapter) CreateMembershipUnchecked(clientID string, current *model.User, group *model.Group, membership *model.GroupMembership) error {
	if group != nil {
		existingMembership, err := sa.FindGroupMembership(clientID, group.ID, current.ID)
		if err != nil || existingMembership == nil || !existingMembership.IsAdmin() {
			log.Printf("error: storage.CreateMembershipUnchecked() - current user is not admin of the group")
			return fmt.Errorf("current user is not admin of the group")
		}

		existingMembership, _ = sa.FindGroupMembership(clientID, group.ID, membership.UserID)
		if existingMembership != nil {
			log.Printf("error: storage.CreateMembershipUnchecked() - member of group '%s' with user id %s already exists", group.Title, membership.UserID)
			return fmt.Errorf("member of group '%s' with user id %s already exists", group.Title, membership.UserID)
		}

		existingMembership, _ = sa.FindGroupMembership(clientID, group.ID, membership.ExternalID)
		if existingMembership != nil {
			log.Printf("error: storage.CreateMembershipUnchecked() - member of group '%s' with external id %s already exists", group.Title, membership.ExternalID)
			return fmt.Errorf("member of group '%s' with external id %s already exists", group.Title, membership.ExternalID)
		}

		if len(membership.UserID) == 0 && len(membership.ExternalID) == 0 {
			log.Printf("error: storage.CreateMembershipUnchecked() - expected user_id or external_id")
			return fmt.Errorf("expected user_id or external_id")
		}

		membership.ID = uuid.NewString()
		membership.ClientID = clientID
		membership.GroupID = group.ID
		membership.DateCreated = time.Now()
		membership.MemberAnswers = group.CreateMembershipEmptyAnswers()

		return sa.PerformTransaction(func(context TransactionContext) error {
			_, err := sa.db.groupMemberships.InsertOne(membership)
			if err != nil {
				return err
			}

			return sa.resetGroupState(context, clientID, membership.GroupID, true, true)
		})
	}

	return nil
}

// ApplyMembershipApproval applies a membership approval
func (sa *Adapter) ApplyMembershipApproval(clientID string, membershipID string, approve bool, rejectReason string) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		status := "rejected"
		if approve {
			status = "member"
		}

		var membership model.GroupMembership
		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "client_id", Value: clientID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: status},
				primitive.E{Key: "reject_reason", Value: rejectReason},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		err := sa.db.groupMemberships.FindOneAndUpdateWithContext(context, filter, update, &membership, nil)
		if err != nil {
			return err
		}

		return sa.resetGroupState(context, clientID, membership.GroupID, true, true)
	})
}

// UpdateMembership updates a membership
func (sa *Adapter) UpdateMembership(clientID string, _ *model.User, membershipID string, membership *model.GroupMembership) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "client_id", Value: clientID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: membership.Status},
				primitive.E{Key: "reject_reason", Value: membership.RejectReason},
				primitive.E{Key: "date_attended", Value: membership.DateAttended},
				primitive.E{Key: "date_updated", Value: time.Now()},
			},
			},
		}
		var membership model.GroupMembership
		err := sa.db.groupMemberships.FindOneAndUpdateWithContext(context, filter, update, &membership, nil)
		if err != nil {
			return err
		}

		return sa.resetGroupState(context, clientID, membership.GroupID, true, true)
	})

}

// DeleteMembership deletes a member membership from a specific group
func (sa *Adapter) DeleteMembership(clientID string, groupID string, userID string) error {

	return sa.PerformTransaction(func(context TransactionContext) error {
		currentMembership, _ := sa.FindGroupMembershipWithContext(context, clientID, groupID, userID)
		if currentMembership != nil {

			if currentMembership.IsAdmin() {
				adminMemberships, _ := sa.FindGroupMembershipsWithContext(context, clientID, model.MembershipFilter{
					GroupIDs: []string{groupID},
					Statuses: []string{"admin"},
				})
				if len(adminMemberships.Items) <= 1 {
					log.Printf("sa.DeleteMembership() - there must be at least two admins in order to delete ")
					return fmt.Errorf("there must be at least two admins in order to delete ")
				}
			}

			filter := bson.D{
				primitive.E{Key: "group_id", Value: groupID},
				primitive.E{Key: "user_id", Value: userID},
				primitive.E{Key: "client_id", Value: clientID},
			}
			_, err := sa.db.groupMemberships.DeleteOneWithContext(context, filter, nil)
			if err != nil {
				log.Printf("error deleting membership - %s", err)
				return err
			}
			return sa.resetGroupState(context, clientID, groupID, true, true)
		}
		return nil
	})
}

// DeleteMembershipByID deletes a membership by ID
func (sa *Adapter) DeleteMembershipByID(clientID string, current *model.User, membershipID string) error {
	return sa.PerformTransaction(func(context TransactionContext) error {
		membership, err := sa.FindGroupMembershipByID(clientID, membershipID)
		if err != nil || membership == nil {
			return fmt.Errorf("membership %s not found", membershipID)
		}

		filter := bson.D{primitive.E{Key: "_id", Value: membershipID}, primitive.E{Key: "client_id", Value: clientID}}
		_, err = sa.db.groupMemberships.DeleteManyWithContext(context, filter, nil)
		if err != nil {
			return err
		}

		return sa.resetGroupState(context, clientID, membership.GroupID, true, true)
	})
}

// DeleteUnsyncedGroupMemberships deletes group memberships that do not exist in the latest sync
func (sa *Adapter) DeleteUnsyncedGroupMemberships(clientID string, groupID string, syncID string, admin *bool) (int64, error) {
	var deletedCount int64 = 0
	err := sa.PerformTransaction(func(context TransactionContext) error {
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
			return err
		}

		deletedCount = result.DeletedCount
		if deletedCount > 0 {
			return sa.resetGroupState(context, clientID, groupID, true, true)
		}

		return nil
	})
	return deletedCount, err
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
func (sa Adapter) GetGroupMembershipStats(context TransactionContext, clientID string, groupID string) (*model.GroupStats, error) {
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
							bson.D{{"$match", bson.D{
								{"status", bson.D{{"$in", []string{"member", "admin"}}}},
							}}},
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
					{"members_added_last_24hours",
						bson.A{
							bson.D{{"$match", bson.D{
								{"status", "member"},
								{"date_created", bson.M{"$gt": time.Now().Add(-24 * time.Hour)}},
							}}},
							bson.D{{"$count", "members_added_last_24hours"}},
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
					{"members_added_last_24hours",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$members_added_last_24hours.members_added_last_24hours",
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
