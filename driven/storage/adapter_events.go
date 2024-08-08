package storage

import (
	"groups/core/model"
	"time"

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
