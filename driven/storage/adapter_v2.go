package storage

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"groups/core/model"
)

// GetGroupMembers Gets all group members
func (sa Adapter) GetGroupMembers(clientID string, groupID string, filter model.MembershipFilter) ([]model.Member, error) {
	innerMatch := bson.D{
		{"_id", groupID},
		{"client_id", clientID},
	}

	pipeline := bson.A{
		bson.D{
			{"$match", innerMatch},
		},
		bson.D{{"$unwind", bson.D{{"path", "$members"}}}},
		bson.D{{"$project", bson.D{{"members", 1}}}},
	}

	matchFilter := bson.D{}
	if filter.ID != nil {
		matchFilter = append(matchFilter, bson.E{"members.id", *filter.ID})
	}
	if filter.UserID != nil {
		matchFilter = append(matchFilter, bson.E{"members.user_id", *filter.UserID})
	}
	if filter.NetID != nil {
		matchFilter = append(matchFilter, bson.E{"members.net_id", *filter.NetID})
	}
	if filter.ExternalID != nil {
		matchFilter = append(matchFilter, bson.E{"members.external_id", *filter.ExternalID})
	}
	if filter.Statuses != nil {
		matchFilter = append(matchFilter, bson.E{"members.status", bson.D{{"$in", filter.Statuses}}})
	}
	if filter.Name != nil {
		matchFilter = append(matchFilter, bson.E{"members.name", primitive.Regex{fmt.Sprintf(`%s`, *filter.Name), "i"}})
	}

	if len(matchFilter) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", matchFilter}})
	}

	pipeline = append(pipeline, bson.D{{"$sort", bson.D{
		{"members.status", 1},
		{"members.name", 1},
	}}})

	if filter.Offset != nil {
		pipeline = append(pipeline, bson.D{{"$skip", *filter.Offset}})
	}
	if filter.Limit != nil {
		pipeline = append(pipeline, bson.D{{"$limit", *filter.Limit}})
	}

	var list []struct {
		ID     string       `json:"id" bson:"_id"`
		Member model.Member `json:"members" bson:"members"`
	}
	err := sa.db.groups.Aggregate(pipeline, &list, nil)
	if err != nil {
		return nil, err
	}

	var resultList []model.Member
	resultLength := len(list)
	if resultLength > 0 {
		resultList = make([]model.Member, resultLength)
		for index, member := range list {
			resultList[index] = member.Member
		}
	}

	return resultList, nil
}

// GetGroupStats Retrieves group stats
func (sa Adapter) GetGroupStats(clientID string, id string) (*model.GroupStats, error) {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{
			{"_id", id},
			{"client_id", clientID},
		}}},
		bson.D{{"$unwind", bson.D{{"path", "$members"}}}},
		bson.D{{"$project", bson.D{{"members", 1}}}},
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
							bson.D{{"$match", bson.D{{"members.status", "admin"}}}},
							bson.D{{"$count", "admins_count"}},
						},
					},
					{"member_count",
						bson.A{
							bson.D{{"$match", bson.D{{"members.status", "member"}}}},
							bson.D{{"$count", "member_count"}},
						},
					},
					{"pending_count",
						bson.A{
							bson.D{{"$match", bson.D{{"members.status", "pending"}}}},
							bson.D{{"$count", "pending_count"}},
						},
					},
					{"rejected_count",
						bson.A{
							bson.D{{"$match", bson.D{{"members.status", "rejected"}}}},
							bson.D{{"$count", "rejected_count"}},
						},
					},
					{"attendance_count",
						bson.A{
							bson.D{{"$match", bson.D{{"members.date_attended", bson.D{
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
	err := sa.db.groups.Aggregate(pipeline, &stats, nil)
	if err != nil {
		return nil, err
	}

	if len(stats) > 0 {
		return &stats[0], err
	}
	return nil, nil
}
