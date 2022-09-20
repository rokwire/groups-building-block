package storage

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"groups/core/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetGroupMembers Gets all group members
func (sa Adapter) GetGroupMembers(clientID string, groupID string, filter *model.MembershipFilter) ([]model.Member, error) {
	membershipFilter := bson.D{
		{"client_id", clientID},
		{"group_id", groupID},
	}

	if filter.ID != nil {
		membershipFilter = append(membershipFilter, bson.E{"id", *filter.ID})
	}
	if filter.UserID != nil {
		membershipFilter = append(membershipFilter, bson.E{"user_id", *filter.UserID})
	} else if filter.UserIDs != nil {
		membershipFilter = append(membershipFilter, bson.E{"user_id", bson.D{{"$in", filter.UserIDs}}})
	}
	if filter.NetID != nil {
		membershipFilter = append(membershipFilter, bson.E{"net_id", *filter.NetID})
	}
	if filter.ExternalID != nil {
		membershipFilter = append(membershipFilter, bson.E{"external_id", *filter.ExternalID})
	}
	if filter.Statuses != nil {
		membershipFilter = append(membershipFilter, bson.E{"status", bson.D{{"$in", filter.Statuses}}})
	}
	if filter.Name != nil {
		membershipFilter = append(membershipFilter, bson.E{"name", primitive.Regex{fmt.Sprintf(`%s`, *filter.Name), "i"}})
	}

	opts := options.Find()
	opts.Sort = bson.D{
		{"status", 1},
		{"name", 1},
	}
	if filter.Offset != nil {
		opts.Skip = filter.Offset
	}
	if filter.Limit != nil {
		opts.Limit = filter.Limit
	}

	var memberships []model.GroupMembership
	err := sa.db.groupMemberships.Find(membershipFilter, &memberships, opts)
	if err != nil {
		return nil, err
	}

	members := make([]model.Member, len(memberships))
	for index, membership := range memberships {
		members[index] = membership.ToMember()
	}

	return members, nil
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
