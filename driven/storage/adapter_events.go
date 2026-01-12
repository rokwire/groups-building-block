package storage

import (
	"groups/core/model"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FindGroupMembershipStatusAndGroupTitle Find group membership status and group Title
func (sa *Adapter) FindGroupMembershipStatusAndGroupTitle(context TransactionContext, userID string) ([]model.GetGroupMembershipsResponse, error) {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: userID}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "groups"},
			{Key: "localField", Value: "group_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "group_info"},
		}}},
		bson.D{{Key: "$unwind", Value: "$group_info"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "title", Value: "$group_info.title"},
			{Key: "status", Value: "$status"},
			{Key: "groupId", Value: "$group_info._id"},
		}}},
	}

	// Define the results slice
	var results []model.GetGroupMembershipsResponse

	// Execute the aggregation pipeline
	err := sa.db.groupMemberships.AggregateWithContext(context, pipeline, &results, nil)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindGroupMembershipByGroupID Find group membership ids
func (sa *Adapter) FindGroupMembershipByGroupID(context TransactionContext, groupID string) ([]string, error) {
	filter := bson.D{primitive.E{Key: "group_id", Value: groupID}}

	// Define the results slice
	var results []model.GroupMembership

	// Execute the aggregation pipeline
	err := sa.db.groupMemberships.FindWithContext(context, filter, &results, nil)
	if err != nil {
		return nil, err
	}

	var userIDs []string
	for _, u := range results {
		if u.UserID != "" {
			userIDs = append(userIDs, u.UserID)
		}
	}

	return userIDs, nil
}

// GetGroupMembershipByUserID Find group membership by userID
func (sa *Adapter) GetGroupMembershipByUserID(userID string) ([]model.GroupMembership, error) {
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

	var groupMemberships []model.GroupMembership
	err := sa.db.groupMemberships.Find(filter, &groupMemberships, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "group membership", nil, err)
	}

	return groupMemberships, nil
}
