package entities

import (
	"time"
)

type OrderEntity struct {
	UUID *string `json:"uuid" bson:"uuid"`
	Ticker *string `json:"ticker" bson:"ticker"`
	Side *string `json:"side" bson:"side"`
	Quantity *int64 `json:"quantity" bson:"quantity"`
	Price *float64 `json:"price" bson:"price"`
	TotalCost *float64 `json:"total_cost" bson:"total_cost"`
	Realized *float64 `json:"realized" bson:"realized"`
	Timestamp *time.Time `json:"timestamp" bson:"timestamp"`
	AccountID *string `json:"account_id" bson:"account_id"`
	PortfolioUUID *string `json:"portfolio_uuid" bson:"portfolio_uuid"`
}