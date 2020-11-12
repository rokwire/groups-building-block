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
}

type member struct {
	ID            string         `bson:"id"`
	UserID        string         `bson:"user_id"`
	Name          string         `bson:"name"`
	Email         string         `bson:"email"`
	PhotoURL      string         `bson:"photo_url"`
	Status        string         `bson:"status"` //pending, member, admin
	MemberAnswers []memberAnswer `bson:"member_answers"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type memberAnswer struct {
	Question string `bson:"question"`
	Answer   string `bson:"answer"`
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

//FindUser finds the user for the provided external id
func (sa *Adapter) FindUser(externalID string) (*model.User, error) {
	filter := bson.D{primitive.E{Key: "external_id", Value: externalID}}
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
func (sa *Adapter) CreateUser(externalID string, email string, isMemberOf *[]string) (*model.User, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	dateCreated := time.Now()
	user := model.User{ID: id.String(), ExternalID: externalID, Email: email,
		IsMemberOf: isMemberOf, DateCreated: dateCreated}
	_, err = sa.db.users.InsertOne(&user)
	if err != nil {
		return nil, err
	}

	//return the inserted item
	return &user, nil
}

//SaveUser saves the user
func (sa *Adapter) SaveUser(user *model.User) error {
	filter := bson.D{primitive.E{Key: "_id", Value: user.ID}}

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
func (sa *Adapter) CreateGroup(title string, description *string, category string, tags []string, privacy string,
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
		group := group{ID: insertedID, Title: title, Description: description, Category: category,
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
func (sa *Adapter) UpdateGroup(id string, category string, title string, privacy string, description *string,
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
		filter := bson.D{primitive.E{Key: "_id", Value: id}}
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

//FindGroup finds group by id
func (sa *Adapter) FindGroup(id string) (*model.Group, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
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
func (sa *Adapter) FindGroups(category *string) ([]model.Group, error) {
	filter := bson.D{}
	if category != nil {
		filter = bson.D{primitive.E{Key: "category", Value: category}}
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

//FindUserGroups finds the user groups
func (sa *Adapter) FindUserGroups(userID string) ([]model.Group, error) {
	filter := bson.D{primitive.E{Key: "members.user_id", Value: userID}}

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
func (sa *Adapter) CreatePendingMember(groupID string, userID string, name string, email string, photoURL string, memberAnswers []string) error {
	//TODO
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
	group := model.Group{ID: groupID}
	dateCreated := member.DateCreated
	dateUpdated := member.DateUpdated

	memberAnswers := make([]model.MemberAnswer, len(member.MemberAnswers))
	for i, current := range member.MemberAnswers {
		memberAnswers[i] = model.MemberAnswer{Question: current.Question, Answer: current.Answer}
	}

	return model.Member{ID: id, User: user, Name: name, Email: email, PhotoURL: photoURL,
		Status: status, Group: group, DateCreated: dateCreated, DateUpdated: dateUpdated, MemberAnswers: memberAnswers}
}
