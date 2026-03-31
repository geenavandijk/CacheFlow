package entities

import (
	"time"
)

type WatchlistEntity struct {
	UUID    *string    `json:"uuid" bson:"uuid"`
	Name    *string    `json:"name" bson:"name"`
	Index   *int32     `json:"index" bson:"index"`
	Tickers []*string  `json:"tickers" bson:"tickers"`
	CreatedAt *time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at"`
}

type PortfolioEntity struct {

	// Unique identifier for the portfolio
	UUID *string `json:"uuid" bson:"uuid"`
	AccountID *string `json:"account_id" bson:"account_id"`

	// General information about the portfolio
	Name *string `json:"name" bson:"name"`
	Description *string `json:"description" bson:"description"`

	// Starting balance of the portfolio
	StartingBalance *float64 `json:"starting_balance" bson:"starting_balance"`
	CurrentBalance *float64 `json:"current_balance" bson:"current_balance"`

	// List of order UUIDs that are associated with the portfolio
	Orders []*string `json:"orders" bson:"orders"`

	// Watchlists scoped to this portfolio
	Watchlists []*WatchlistEntity `json:"watchlists,omitempty" bson:"watchlists,omitempty"`

	// Timestamp for when the portfolio was created / updated
	CreatedAt *time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at"`
}