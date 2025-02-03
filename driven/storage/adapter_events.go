package storage

import (
	"groups/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FindAdminGroupsForEvent Finds all groups for an event where the user is admin
func (sa *Adapter) FindAdminGroupsForEvent(context TransactionContext, clientID string, current *model.User, eventID string) ([]string, error) {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "event_id", Value: eventID},
			{Key: "client_id", Value: clientID},
		}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "group_memberships"},
					{Key: "localField", Value: "group_id"},
					{Key: "foreignField", Value: "group_id"},
					{Key: "as", Value: "membership"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$membership"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "membership.user_id", Value: current.ID}}}},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: primitive.Null{}},
					{Key: "group_ids", Value: bson.D{{Key: "$push", Value: "$membership.group_id"}}},
				},
			},
		},
	}

	type aggregator struct {
		GroupIDs []string `bson:"group_ids"`
	}
	var result []aggregator

	err := sa.db.events.AggregateWithContext(context, pipeline, &result, nil)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0].GroupIDs, err
	}

	return nil, nil
}

// FindAdminGroupsIDs Finds all groups where the current user is admin
func (sa *Adapter) FindAdminGroupsIDs(context TransactionContext, clientID string, current *model.User) ([]string, error) {
	pipeline := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "user_id", Value: current.ID},
					{Key: "status", Value: "admin"},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: primitive.Null{}},
					{Key: "group_ids", Value: bson.D{{Key: "$push", Value: "$group_id"}}},
				},
			},
		},
	}

	type aggregator struct {
		GroupIDs []string `bson:"group_ids"`
	}
	var result []aggregator

	err := sa.db.groupMemberships.AggregateWithContext(context, pipeline, &result, nil)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0].GroupIDs, err
	}

	return nil, nil
}

// UpdateGroupMappingsForEvent Updates group mappings for an event
func (sa *Adapter) UpdateGroupMappingsForEvent(context TransactionContext, clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	var result []string

	wrapper := func(context TransactionContext) error {
		// 1. Construct mappings for lookups
		adminIDMappings := map[string]bool{}
		adminGroupIDs, err := sa.FindAdminGroupsIDs(context, clientID, current)
		if err != nil {
			return err
		}
		for _, groupID := range adminGroupIDs {
			adminIDMappings[groupID] = true
		}

		currentAdminIDMappings := map[string]bool{}
		currentAdminGroupIDs, err := sa.FindAdminGroupsForEvent(context, clientID, current, eventID)
		if err != nil {
			return err
		}
		for _, groupID := range currentAdminGroupIDs {
			currentAdminIDMappings[groupID] = true
		}

		newGroupIDsMapping := map[string]bool{}
		for _, groupID := range groupIDs {
			newGroupIDsMapping[groupID] = true
		}

		for _, groupID := range groupIDs {
			if adminIDMappings[groupID] {
				result = append(result, groupID)
			}
		}

		// 2. Construct mappings for remove
		var groupIDsForRemove []string
		for _, groupID := range currentAdminGroupIDs {
			if currentAdminIDMappings[groupID] && !newGroupIDsMapping[groupID] {
				groupIDsForRemove = append(groupIDsForRemove, groupID)
			}
		}
		if len(groupIDsForRemove) > 0 {
			_, err := sa.db.events.DeleteManyWithContext(context, bson.D{
				{Key: "event_id", Value: eventID},
				{Key: "group_id", Value: bson.M{"$in": groupIDsForRemove}},
				{Key: "client_id", Value: clientID},
			}, nil)
			if err != nil {
				return err
			}
		}

		var eventsForAdd []interface{}
		for _, groupID := range groupIDs {
			if _, ok := currentAdminIDMappings[groupID]; !ok {
				if _, innerOK := adminIDMappings[groupID]; innerOK {
					eventsForAdd = append(eventsForAdd, model.Event{
						ClientID: clientID,
						GroupID:  groupID,
						EventID:  eventID,
						Creator: &model.Creator{
							UserID: current.ID,
							Name:   current.Name,
							Email:  current.Email,
						},
						DateCreated: time.Now(),
					})
				}
			}
			if !currentAdminIDMappings[groupID] && adminIDMappings[groupID] {

			}
		}
		if len(eventsForAdd) > 0 {
			_, err := sa.db.events.InsertManyWithContext(context, eventsForAdd, nil)
			if err != nil {
				return err
			}
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

	return result, nil
}

// FindEventUserIDs Find all linked users for group event
func (sa *Adapter) FindEventUserIDs(context TransactionContext, eventID string) ([]string, error) {

	var list []struct {
		List []string `bson:"list"`
	}
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "event_id", Value: eventID}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "group_memberships"},
					{Key: "localField", Value: "group_id"},
					{Key: "foreignField", Value: "group_id"},
					{Key: "as", Value: "member"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$member"}}}},
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$member.user_id"}}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.M{"$ne": nil}},
			{Key: "_id", Value: bson.M{"$ne": ""}},
		}}},
		bson.D{{Key: "$group",
			Value: bson.D{
				{Key: "_id", Value: primitive.Null{}},
				{Key: "list", Value: bson.D{{Key: "$addToSet", Value: "$_id"}}},
			},
		}},
	}

	err := sa.db.events.AggregateWithContext(context, pipeline, &list, nil)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0].List, nil
	}

	return nil, nil

}

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

// FindGroupsEvents Find group ID and event ID
func (sa *Adapter) FindGroupsEvents(context TransactionContext, eventIDs []string) ([]model.GetGroupsEvents, error) {
	filter := bson.D{}

	if len(eventIDs) > 0 {
		filter = append(filter, bson.E{Key: "event_id", Value: bson.M{"$in": eventIDs}})
	}

	var groupsEvents []model.GetGroupsEvents
	err := sa.db.events.Find(filter, &groupsEvents, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "groups events", nil, err)
	}

	return groupsEvents, nil
}

// GetEventByUserID Find events by userID
func (sa *Adapter) GetEventByUserID(userID string) ([]model.Event, error) {
	filter := bson.D{primitive.E{Key: "to_members.user_id", Value: userID}}

	var events []model.Event
	err := sa.db.events.Find(filter, &events, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "events", nil, err)
	}

	return events, nil
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

// GetPostsByUserID Find posts by userID
func (sa *Adapter) GetPostsByUserID(userID string) ([]model.Post, error) {
	filter := bson.D{primitive.E{Key: "member.user_id", Value: userID}}

	var posts []model.Post
	err := sa.db.posts.Find(filter, &posts, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "posts", nil, err)
	}

	return posts, nil
}

// FindGroupsByGroupIDs Find group by group ID
func (sa *Adapter) FindGroupsByGroupIDs(groupIDs []string) ([]model.Group, error) {
	filter := bson.D{}

	if len(groupIDs) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": groupIDs}})
	}

	var groups []model.Group
	err := sa.db.groups.Find(filter, &groups, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "groups", nil, err)
	}

	return groups, nil
}
