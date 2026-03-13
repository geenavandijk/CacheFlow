package mongo

import (
	"context"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes() {
	db := GetMongoDatabase(context.Background())
	if db == nil {
		log.Error("mongo database is nil, could not create indexes")
		return
	}

	// Orders collection
	ordersCollection := db.Collection("orders")

	orderIndexes := []mongodriver.IndexModel{
		{
			Keys:    bson.D{{Key: "ticker", Value: 1}},
			Options: options.Index().SetName("ticker_1"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_-1"),
		},
	}

	_, err := ordersCollection.Indexes().CreateMany(context.Background(), orderIndexes)
	if err != nil {
		log.Error("failed to create order indexes", "err", err)
	} else {
		log.Info("order indexes ensured")
	}

	// Portfolio collection
	portfolioCollection := db.Collection("portfolios")

	portfolioIndexes := []mongodriver.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_-1"),
		},
	}

	_, err = portfolioCollection.Indexes().CreateMany(context.Background(), portfolioIndexes)
	if err != nil {
		log.Error("failed to create portfolio indexes", "err", err)
	} else {
		log.Info("portfolio indexes ensured")
	}
}