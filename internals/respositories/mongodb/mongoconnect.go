package mongodb

import (
	"context"
	"log"
	"schoolmanagementGRPC/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func CreateMongoClient() (*mongo.Client, error) {
	ctx := context.Background()

	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("username:password@mongodb://localhost:27017"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Unable to connect to database")
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Unable to ping th database")
	}

	log.Println("Connected to mongodb successfully")
	return client, nil
}
