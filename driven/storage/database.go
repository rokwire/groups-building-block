package storage

import (
	"context"
	"groups/core"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client

	users  *collectionWrapper
	enums  *collectionWrapper
	groups *collectionWrapper
	events *collectionWrapper

	listener core.StorageListener
}

func (m *database) start() error {
	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(m.mongoDBName)

	users := &collectionWrapper{database: m, coll: db.Collection("users")}
	err = m.applyUsersChecks(users)
	if err != nil {
		return err
	}

	enums := &collectionWrapper{database: m, coll: db.Collection("enums")}
	err = m.applyEnumsChecks(enums)
	if err != nil {
		return err
	}

	groups := &collectionWrapper{database: m, coll: db.Collection("groups")}
	err = m.applyGroupsChecks(groups)
	if err != nil {
		return err
	}

	events := &collectionWrapper{database: m, coll: db.Collection("events")}
	err = m.applyEventsChecks(events)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.users = users
	m.enums = enums
	m.groups = groups
	m.events = events

	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	log.Println("apply users checks.....")

	//add index - unique
	err := users.AddIndex(bson.D{primitive.E{Key: "external_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	err = users.AddIndex(bson.D{primitive.E{Key: "client_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("users checks passed")
	return nil
}

func (m *database) applyEnumsChecks(enums *collectionWrapper) error {
	log.Println("apply enums checks.....")

	log.Println("enums checks passed")
	return nil
}

func (m *database) applyGroupsChecks(groups *collectionWrapper) error {
	log.Println("apply groups checks.....")

	err := groups.AddIndex(bson.D{primitive.E{Key: "title", Value: 1}}, true)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "category", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "members.id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "members.user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	err = groups.AddIndex(bson.D{primitive.E{Key: "client_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("groups checks passed")
	return nil
}

func (m *database) applyEventsChecks(events *collectionWrapper) error {
	log.Println("apply events checks.....")

	//compound index
	err := events.AddIndex(bson.D{primitive.E{Key: "event_id", Value: 1}, primitive.E{Key: "group_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	err = events.AddIndex(bson.D{primitive.E{Key: "client_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("events checks passed")
	return nil
}

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}

	//do nothing for now
}
