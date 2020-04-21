package storage

import (
	"groups/core"
	"groups/core/model"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
