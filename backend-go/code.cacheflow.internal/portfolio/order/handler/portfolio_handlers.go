package handler

import (
	"context"

	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	datastores "code.cacheflow.internal/datastores/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetActiveShares walks through all orders for a given ticker + portfolio,
// ordered by timestamp, and applies a FIFO algorithm to determine how many
// shares remain open (i.e. not fully sold).
func GetActiveShares(ticker string, portfolioUUID string) (int64, error) {
	if ticker == "" || portfolioUUID == "" {
		return 0, nil
	}

	ordersCollection := datastores.GetMongoDatabase(context.Background()).Collection(datastores.Orders)

	// Fetch all orders for this portfolio + ticker ordered by time (oldest first)
	filter := bson.M{
		"ticker":         ticker,
		"portfolio_uuid": portfolioUUID,
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cur, err := ordersCollection.Find(context.Background(), filter, findOpts)
	if err != nil {
		return 0, err
	}
	defer cur.Close(context.Background())

	// FIFO lots: each BUY creates a lot, SELL consumes from the earliest lots
	var lots []int64

	for cur.Next(context.Background()) {
		var o orderEntities.OrderEntity
		if err := cur.Decode(&o); err != nil {
			return 0, err
		}

		if o.Side == nil || o.Quantity == nil {
			continue
		}

		switch *o.Side {
		case "BUY":
			if *o.Quantity > 0 {
				lots = append(lots, *o.Quantity)
			}
		case "SELL":
			toSell := *o.Quantity
			if toSell <= 0 {
				continue
			}

			for i := 0; i < len(lots) && toSell > 0; i++ {
				if lots[i] == 0 {
					continue
				}
				if lots[i] <= toSell {
					toSell -= lots[i]
					lots[i] = 0
				} else {
					lots[i] -= toSell
					toSell = 0
				}
			}
		default:
			// ignore unknown sides
			continue
		}
	}

	if err := cur.Err(); err != nil {
		return 0, err
	}

	var remaining int64
	for _, q := range lots {
		if q > 0 {
			remaining += q
		}
	}

	return remaining, nil
}
