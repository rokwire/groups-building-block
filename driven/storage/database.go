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

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.users = users
	m.enums = enums
	m.groups = groups

	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	log.Println("apply users checks.....")

	//add index - unique
	err := users.AddIndex(bson.D{primitive.E{Key: "external_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("users checks passed")
	return nil
}

func (m *database) applyEnumsChecks(enums *collectionWrapper) error {
	log.Println("apply enums checks.....")

	//add initial group categories data if not already added
	fFilter := bson.D{primitive.E{Key: "_id", Value: "categories"}}
	var result []enumItem
	err := enums.Find(fFilter, &result, nil)
	if err != nil {
		return err
	}
	hasData := result != nil && len(result) > 0
	if !hasData {
		log.Println("there is no group categories, so add initial data")

		data := []string{"Academic/Pre-Professional", "Athletic/Recreation", "Club Sports",
			"Creative/Media/Performing Arts", "Cultural/Ethnic", "Graduate",
			"Honorary", "International", "Other Social",
			"Political", "Religious", "Residence Hall",
			"Rights/Freedom Issues", "ROTC", "Service/Philanthropy",
			"Social Fraternity/Sorority", "University Student Governance/Council/Committee"}

		categoriesData := enumItem{ID: "categories", Values: data}
		_, err = enums.InsertOne(&categoriesData)
		if err != nil {
			return err
		}
	} else {
		log.Println("there is group categories data, so do nothing")
	}

	log.Println("enums checks passed")
	return nil
}

func (m *database) applyGroupsChecks(groups *collectionWrapper) error {
	log.Println("apply groups checks.....")

	err := groups.AddIndex(bson.D{primitive.E{Key: "category", Value: 1}}, false)
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

	log.Println("groups checks passed")
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
