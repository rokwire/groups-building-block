package storage

import (
	"context"
	"fmt"
	"groups/core"
	"groups/core/model"
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

	users   *collectionWrapper
	enums   *collectionWrapper
	groups  *collectionWrapper
	events  *collectionWrapper
	configs *collectionWrapper

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
	configs := &collectionWrapper{database: m, coll: db.Collection("configs")}
	err = m.applyConfigChecks(configs)
	if err != nil {
		return err
	}

	//apply multi-tenant
	err = m.applyMultiTenantChecks(client, users, groups, events)
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
	m.configs = configs

	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	log.Println("apply users checks.....")

	//drop the previous index - external_id
	err := users.DropIndex("external_id_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}

	//drop the previous index - client_id
	err = users.DropIndex("client_id_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}

	//create compound index
	err = users.AddIndex(bson.D{primitive.E{Key: "external_id", Value: 1}, primitive.E{Key: "client_id", Value: 1}}, true)
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

	//drop the previous index - title
	err := groups.DropIndex("title_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}
	//drop the previous index - client_id
	err = groups.DropIndex("client_id_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}
	func (m *database) applyConfigChecks(config *collectionWrapper) error {
		log.Println("apply config checks.....")
	}

	//create compound index
	err = groups.AddIndex(bson.D{primitive.E{Key: "title", Value: 1}, primitive.E{Key: "client_id", Value: 1}}, true)
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

	log.Println("groups checks passed")
	return nil
}

func (m *database) applyEventsChecks(events *collectionWrapper) error {
	log.Println("apply events checks.....")

	//drop the previous compound index
	err := events.DropIndex("event_id_1_group_id_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}
	//drop the previous index - client_id
	err = events.DropIndex("client_id_1")
	if err != nil {
		//just log
		log.Printf("%s", err.Error())
	}

	//create compound index
	err = events.AddIndex(bson.D{primitive.E{Key: "event_id", Value: 1},
		primitive.E{Key: "group_id", Value: 1},
		primitive.E{Key: "client_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("events checks passed")
	return nil
}

func (m *database) applyMultiTenantChecks(client *mongo.Client, users *collectionWrapper, groups *collectionWrapper, events *collectionWrapper) error {
	log.Println("apply multi-tenant checks.....")

	// transaction
	err := client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		//apply users collection
		var usersList []model.User
		err = users.FindWithContext(sessionContext, bson.D{}, &usersList, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if len(usersList) > 0 {
			for _, u := range usersList {
				if len(u.ClientID) == 0 {
					log.Printf("USERS - SET CLIENT ID for %s", u.Email)

					_, err = users.UpdateOneWithContext(sessionContext,
						bson.D{primitive.E{Key: "_id", Value: u.ID}},
						bson.D{
							primitive.E{Key: "$set", Value: bson.D{
								primitive.E{Key: "client_id", Value: "edu.illinois.rokwire"}},
							}},
						nil)
					if err != nil {
						abortTransaction(sessionContext)
						return err
					}
				}
			}
		}

		//apply groups collection
		var groupsList []group
		err = groups.FindWithContext(sessionContext, bson.D{}, &groupsList, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if len(groupsList) > 0 {
			for _, gr := range groupsList {
				if len(gr.ClientID) == 0 {
					log.Printf("GROUPS - SET CLIENT ID for %s", gr.Title)

					_, err = groups.UpdateOneWithContext(sessionContext,
						bson.D{primitive.E{Key: "_id", Value: gr.ID}},
						bson.D{
							primitive.E{Key: "$set", Value: bson.D{
								primitive.E{Key: "client_id", Value: "edu.illinois.rokwire"}},
							}},
						nil)
					if err != nil {
						abortTransaction(sessionContext)
						return err
					}
				}
			}
		}

		//apply events collection
		var eventsList []event
		err = events.FindWithContext(sessionContext, bson.D{}, &eventsList, nil)
		if err != nil {
			abortTransaction(sessionContext)
			return err
		}
		if len(eventsList) > 0 {
			for _, ev := range eventsList {
				if len(ev.ClientID) == 0 {
					log.Printf("EVENTS - SET CLIENT ID for %s", ev.EventID)

					_, err = events.UpdateOneWithContext(sessionContext,
						bson.D{primitive.E{Key: "event_id", Value: ev.EventID}},
						bson.D{
							primitive.E{Key: "$set", Value: bson.D{
								primitive.E{Key: "client_id", Value: "edu.illinois.rokwire"}},
							}},
						nil)
					if err != nil {
						abortTransaction(sessionContext)
						return err
					}
				}
			}
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

	log.Println("multi-tenant checks passed")
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
