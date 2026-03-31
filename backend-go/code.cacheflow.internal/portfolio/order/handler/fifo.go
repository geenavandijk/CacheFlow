package handler

import (
	"context"

	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	datastores "code.cacheflow.internal/datastores/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fifoLot struct {
	Qty       int64
	CostPerSh float64
}

// buildLotsForTicker builds FIFO lots for a given ticker + portfolio from all existing orders,
// ordered by timestamp ascending. It mutates lots based on historical BUY/SELL activity.
func buildLotsForTicker(ctx context.Context, ticker, portfolioUUID string) ([]fifoLot, error) {
	ordersCollection := datastores.GetMongoDatabase(ctx).Collection(datastores.Orders)

	filter := bson.M{
		"ticker":         ticker,
		"portfolio_uuid": portfolioUUID,
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cur, err := ordersCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var lots []fifoLot

	for cur.Next(ctx) {
		var o orderEntities.OrderEntity
		if err := cur.Decode(&o); err != nil {
			return nil, err
		}
		if o.Side == nil || o.Quantity == nil || o.Price == nil {
			continue
		}

		switch *o.Side {
		case "BUY":
			if *o.Quantity > 0 {
				lots = append(lots, fifoLot{
					Qty:       *o.Quantity,
					CostPerSh: *o.Price,
				})
			}
		case "SELL":
			toSell := *o.Quantity
			if toSell <= 0 {
				continue
			}
			for i := 0; i < len(lots) && toSell > 0; i++ {
				if lots[i].Qty == 0 {
					continue
				}
				if lots[i].Qty <= toSell {
					toSell -= lots[i].Qty
					lots[i].Qty = 0
				} else {
					lots[i].Qty -= toSell
					toSell = 0
				}
			}
		default:
			continue
		}
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return lots, nil
}

// CalculateRealizedPnLForSell takes a SELL order (quantity + price) and returns the realized PnL
// using FIFO based on all existing orders for that ticker + portfolio.
func CalculateRealizedPnLForSell(ctx context.Context, ticker, portfolioUUID string, sellQty int64, sellPrice float64) (float64, error) {
	if ticker == "" || portfolioUUID == "" || sellQty <= 0 {
		return 0, nil
	}

	lots, err := buildLotsForTicker(ctx, ticker, portfolioUUID)
	if err != nil {
		return 0, err
	}

	toSell := sellQty
	var realized float64

	for i := 0; i < len(lots) && toSell > 0; i++ {
		if lots[i].Qty == 0 {
			continue
		}
		if lots[i].Qty <= toSell {
			shares := lots[i].Qty
			realized += float64(shares) * (sellPrice - lots[i].CostPerSh)
			toSell -= shares
			lots[i].Qty = 0
		} else {
			shares := toSell
			realized += float64(shares) * (sellPrice - lots[i].CostPerSh)
			lots[i].Qty -= toSell
			toSell = 0
		}
	}

	return realized, nil
}

