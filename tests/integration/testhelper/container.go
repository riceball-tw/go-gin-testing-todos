package testhelper

import (
	"context"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupTestContainer(ctx context.Context) (*mongodb.MongoDBContainer, *mongo.Database, error) {
	mongodbContainer, err := mongodb.Run(ctx, "mongo:latest")
	if err != nil {
		return nil, nil, err
	}

	endpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		return nil, nil, err
	}

	return mongodbContainer, client.Database("test_db"), nil
}
