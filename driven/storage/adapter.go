package storage

import (
	"context"
	"errors"
	"fmt"
	"groups/core"
	"groups/core/model"
	"log"
	"strconv"
	"time"

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
	Description         *string  `bson:"description"`
	ImageURL            *string  `bson:"image_url"`
	WebURL              *string  `bson:"web_url"`
	MembersCount        int      `bson:"members_count"` //to be supported up to date
	Tags                []string `bson:"tags"`
	MembershipQuestions []string `bson:"membership_questions"`

	Members []member `bson:"members"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`

	ClientID string `bson:"client_id"`
}

type member struct {
	ID            string         `bson:"id"`
	UserID        string         `bson:"user_id"`
	Name          string         `bson:"name"`
	Email         string         `bson:"email"`
	PhotoURL      string         `bson:"photo_url"`
	Status        string         `bson:"status"` //pending, member, admin, reject
	RejectReason  string         `bson:"reject_reason"`
	MemberAnswers []memberAnswer `bson:"member_answers"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type memberAnswer struct {
	Question string `bson:"question"`
	Answer   string `bson:"answer"`
}

type event struct {
	EventID     string    `bson:"event_id"`
	GroupID     string    `bson:"group_id"`
	DateCreated time.Time `bson:"date_created"`
	Comments    []comment `bson:"comments"`

	ClientID string `bson:"client_id"`
}

type comment struct {
	MemberID    string    `bson:"member_id"`
	Text        string    `bson:"text"`
	DateCreated time.Time `bson:"date_created"`
}

//Adapter implements the Storage interface
type Adapter struct {
	db *database
}

//Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

//SetStorageListener sets listener for the storage
func (sa *Adapter) SetStorageListener(storageListener core.StorageListener) {
	sa.db.listener = storageListener
}

//FindUser finds the user for the provided external id and client id
func (sa *Adapter) FindUser(clientID string, externalID string) (*model.User, error) {
	filter := bson.D{primitive.E{Key: "client_id", Value: clientID},
		primitive.E{Key: "external_id", Value: externalID}}
	var result []*model.User
	err := sa.db.users.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, nil
	}
	return result[0], nil
}

//CreateUser creates an user
func (sa *Adapter) CreateUser(clientID string, externalID string, email string, isMemberOf *[]string) (*model.User, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	dateCreated := time.Now()
	user := model.User{ID: id.String(), ClientID: clientID, ExternalID: externalID, Email: email,
		IsMemberOf: isMemberOf, DateCreated: dateCreated}
	_, err = sa.db.users.InsertOne(&user)
	if err != nil {
		return nil, err
	}

	//return the inserted item
	return &user, nil
}

//SaveUser saves the user
func (sa *Adapter) SaveUser(clientID string, user *model.User) error {
	filter := bson.D{primitive.E{Key: "_id", Value: user.ID}}

	//clientID - no need ...

	dateUpdated := time.Now()
	user.DateUpdated = &dateUpdated

	err := sa.db.users.ReplaceOne(filter, user, nil)
	if err != nil {
		return err
	}
	return nil
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
func (sa *Adapter) CreateGroup(clientID string, title string, description *string, category string, tags []string, privacy string,
	creatorUserID string, creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	var insertedID string

	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//1. check if the category value is one of the enums list
		categoryFilter := bson.D{primitive.E{Key: "values", Value: category}}
		var categoriesResult []enumItem
		err = sa.db.enums.FindWithContext(sessionContext, categoryFilter, &categoriesResult, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if len(categoriesResult) == 0 {
			abortTransaction(sessionContext)
			return errors.New("the provided category must be one of the categories list")
		}

		//2. insert the group and the admin member
		now := time.Now()

		memberID, _ := uuid.NewUUID()
		adminMember := member{ID: memberID.String(), UserID: creatorUserID, Name: creatorName, Email: creatorEmail,
			PhotoURL: creatorPhotoURL, Status: "admin", DateCreated: now}

		members := []member{adminMember}

		groupID, _ := uuid.NewUUID()
		insertedID = groupID.String()
		group := group{ID: insertedID, ClientID: clientID, Title: title, Description: description, Category: category,
			Tags: tags, Privacy: privacy, MembersCount: 1, Members: members, DateCreated: now}
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
		return nil, err
	}

	return &insertedID, nil
}

//UpdateGroup updates a group.
func (sa *Adapter) UpdateGroup(clientID string, id string, category string, title string, privacy string, description *string,
	imageURL *string, webURL *string, tags []string, membershipQuestions []string) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//1. check if the category value is one of the enums list
		categoryFilter := bson.D{primitive.E{Key: "values", Value: category}}
		var categoriesResult []enumItem
		err = sa.db.enums.FindWithContext(sessionContext, categoryFilter, &categoriesResult, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if len(categoriesResult) == 0 {
			abortTransaction(sessionContext)
			return errors.New("the provided category must be one of the categories list")
		}

		//2. update the group
		filter := bson.D{primitive.E{Key: "_id", Value: id},
			primitive.E{Key: "client_id", Value: clientID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "category", Value: category},
				primitive.E{Key: "title", Value: title},
				primitive.E{Key: "privacy", Value: privacy},
				primitive.E{Key: "description", Value: description},
				primitive.E{Key: "image_url", Value: imageURL},
				primitive.E{Key: "web_url", Value: webURL},
				primitive.E{Key: "tags", Value: tags},
				primitive.E{Key: "membership_questions", Value: membershipQuestions},
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

//FindGroup finds group by id and client id
func (sa *Adapter) FindGroup(clientID string, id string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: id},
		primitive.E{Key: "client_id", Value: clientID}}
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

//FindGroups finds groups
func (sa *Adapter) FindGroups(clientID string, category *string) ([]model.Group, error) {
	filter := bson.D{primitive.E{Key: "client_id", Value: clientID}}
	if category != nil {
		filter = bson.D{primitive.E{Key: "category", Value: category},
			primitive.E{Key: "client_id", Value: clientID}}
	}

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

//CreatePendingMember creates a pending member for a specific group
func (sa *Adapter) CreatePendingMember(clientID string, groupID string, userID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error {
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

		//2. check if the user is already a member of this group - pending or member or admin or rejected
		members := group.Members
		if members != nil {
			for _, cMember := range members {
				if cMember.UserID == userID {
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
		if len(group.MembershipQuestions) != len(memberAnswers) {
			return errors.New("member answers mismatch")
		}

		//4. now we can add the pending member
		now := time.Now()
		memberID, _ := uuid.NewUUID()
		var memberAns []memberAnswer
		if len(memberAnswers) > 0 {
			for _, cAns := range memberAnswers {
				memberAns = append(memberAns, memberAnswer{Question: cAns.Question, Answer: cAns.Answer})
			}
		}
		pendingMember := member{ID: memberID.String(), UserID: userID, Name: name, Email: email,
			PhotoURL: photoURL, Status: "pending", MemberAnswers: memberAns, DateCreated: now}
		groupMembers := group.Members
		groupMembers = append(groupMembers, pendingMember)
		saveFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "members", Value: groupMembers},
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

//DeleteMember deletes a member membership from a specific group
func (sa *Adapter) DeleteMember(clientID string, groupID string, userID string) error {
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
			ID           string `bson:"_id"`
			MembersCount int    `bson:"members_count"`
			Member       member `bson:"members"`
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
		membersCount := resultItem.MembersCount
		member := resultItem.Member
		if !(member.Status == "admin" || member.Status == "member") {
			abortTransaction(sessionContext)
			return errors.New("you are not member/admin to the group")
		}

		//check if the member is admin, do not allow the group to become with 0 admins
		if member.Status == "admin" {
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
		membersCount-- //keep the members count updated
		changeFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "members_count", Value: membersCount}}},
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
		membersCount := group.MembersCount
		groupMembers := group.Members
		now := time.Now()
		if approve {
			//apply approve
			member.DateUpdated = &now
			member.Status = "member"
			membersCount = membersCount + 1
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
				primitive.E{Key: "members_count", Value: membersCount},
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
func (sa *Adapter) DeleteMembership(currentUserID string, membershipID string) error {
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
			{"$match": bson.M{"members.id": membershipID}},
		}
		var result []struct {
			GroupID      string `bson:"_id"`
			MembersCount int    `bson:"members_count"`
			Member       member `bson:"members"`
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
		membersCount := resultItem.MembersCount
		member := resultItem.Member
		if member.UserID == currentUserID {
			abortTransaction(sessionContext)
			return errors.New("you cannot remove yourself")
		}
		if !(member.Status == "admin" || member.Status == "member" || member.Status == "rejected") {
			abortTransaction(sessionContext)
			return errors.New("membership which is not member or admin or rejected cannot be removed from the group")
		}

		// delete the membership, also keep the group members count updated
		membersCount-- //keep the members count updated
		changeFilter := bson.D{primitive.E{Key: "_id", Value: groupID}}
		change := bson.D{
			primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "members_count", Value: membersCount}}},
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
func (sa *Adapter) UpdateMembership(currentUserID string, membershipID string, status string) error {
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
			{"$match": bson.M{"members.id": membershipID}},
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
		if member.UserID == currentUserID {
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
				primitive.E{Key: "members.$.status", Value: status},
				primitive.E{Key: "members.$.date_updated", Value: time.Now()},
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
func (sa *Adapter) FindEvents(clientID string, groupID string) ([]model.Event, error) {
	filter := bson.D{primitive.E{Key: "group_id", Value: groupID},
		primitive.E{Key: "client_id", Value: clientID}}
	var result []event
	err := sa.db.events.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return make([]model.Event, 0), nil
	}

	resList := make([]model.Event, len(result))
	for i, e := range result {
		group := model.Group{ID: groupID}
		resList[i] = model.Event{EventID: e.EventID, Group: group, DateCreated: e.DateCreated}
	}

	return resList, nil
}

//CreateEvent creates a group event
func (sa *Adapter) CreateEvent(clientID string, eventID string, groupID string) error {
	event := event{ClientID: clientID, EventID: eventID, GroupID: groupID, DateCreated: time.Now()}
	_, err := sa.db.events.InsertOne(event)
	if err != nil {
		return err
	}

	return nil
}

//DeleteEvent deletes a group event
func (sa *Adapter) DeleteEvent(clientID string, eventID string, groupID string) error {
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
	category := gr.Category
	title := gr.Title
	privacy := gr.Privacy
	description := gr.Description
	imageURL := gr.ImageURL
	webURL := gr.WebURL
	membersCount := gr.MembersCount
	tags := gr.Tags
	membershipQuestions := gr.MembershipQuestions

	dateCreated := gr.DateCreated
	dateUpdated := gr.DateUpdated

	members := make([]model.Member, len(gr.Members))
	for i, current := range gr.Members {
		members[i] = constructMember(id, current)
	}

	return model.Group{ID: id, Category: category, Title: title, Privacy: privacy,
		Description: description, ImageURL: imageURL, WebURL: webURL, MembersCount: membersCount,
		Tags: tags, MembershipQuestions: membershipQuestions, DateCreated: dateCreated, DateUpdated: dateUpdated,
		Members: members}
}

func constructMember(groupID string, member member) model.Member {
	id := member.ID
	user := model.User{ID: member.UserID}
	name := member.Name
	email := member.Email
	photoURL := member.PhotoURL
	status := member.Status
	rejectReason := member.RejectReason
	group := model.Group{ID: groupID}
	dateCreated := member.DateCreated
	dateUpdated := member.DateUpdated

	memberAnswers := make([]model.MemberAnswer, len(member.MemberAnswers))
	for i, current := range member.MemberAnswers {
		memberAnswers[i] = model.MemberAnswer{Question: current.Question, Answer: current.Answer}
	}

	return model.Member{ID: id, User: user, Name: name, Email: email, PhotoURL: photoURL,
		Status: status, RejectReason: rejectReason, Group: group, DateCreated: dateCreated, DateUpdated: dateUpdated, MemberAnswers: memberAnswers}
}
