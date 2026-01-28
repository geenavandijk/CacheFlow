package mongo

import (
	"context" // Context package for managing multiple requests
	"os"
	"strings"
	"sync" // Once package for ensuring that a function is only called once

	"github.com/charmbracelet/log" // log is a simple logging package for Go
	"go.mongodb.org/mongo-driver/mongo"         // MongoDB package for connecting to the MongoDB database
	"go.mongodb.org/mongo-driver/mongo/options" // Options package for configuring the client
)

// MARK: Global Variables
var (
    globalMongoClient   *mongo.Client
    globalMongoDatabase *mongo.Database
    once                sync.Once
)

// MARK: GetMongoClient
func GetMongoClient() *mongo.Client {
    return globalMongoClient
}

// MARK: GetMongoDatabase
func GetMongoDatabase(ctx context.Context) *mongo.Database {
    return globalMongoDatabase
}

// MARK: ConnectDB
func ConnectDB(uri string) *mongo.Client {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: true,
        ReportCaller:    true,
        TimeFormat:      "2006-01-02 15:04:05",
        Prefix:          "DATABASE",
    })

    once.Do(func() {
      // uri mongodb://localhost:27017
        serverAPI := options.ServerAPI(options.ServerAPIVersion1)
        opts := options.Client().
        ApplyURI(uri).
        SetMaxPoolSize(100).
        SetMinPoolSize(10).
        SetServerAPIOptions(serverAPI)
    

        client, err := mongo.Connect(context.TODO(), opts)
        if err != nil {
            logger.Fatal(err)
        }

        if err = client.Ping(context.TODO(), nil); err != nil {
            logger.Fatal(err)
        }

        logger.Info("Successfully connected to MongoDB")

        globalMongoClient = client
        if strings.Contains(uri, "production") {
            globalMongoDatabase = client.Database("production")
        } else {
            globalMongoDatabase = client.Database("development")
        }
    })

    return globalMongoClient
}