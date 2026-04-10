package entities

import "time"

// BacktestTrade is a single trade event in a backtest simulation.
type BacktestTrade struct {
	Type       string  `json:"type" bson:"type"`               // "BUY" or "SELL"
	Date       string  `json:"date" bson:"date"`               // YYYY-MM-DD
	Price      float64 `json:"price" bson:"price"`
	Shares     float64 `json:"shares" bson:"shares"`
	Value      float64 `json:"value" bson:"value"`             // price * shares
	PnL        float64 `json:"pnl" bson:"pnl"`                 // realized P&L (SELL only)
	PnLPercent float64 `json:"pnl_percent" bson:"pnl_percent"` // realized P&L % (SELL only)
	CashAfter  float64 `json:"cash_after" bson:"cash_after"`   // cash remaining after trade
}

// BacktestEntity is a saved backtest result.
type BacktestEntity struct {
	UUID           string          `json:"uuid" bson:"uuid"`
	StrategyUUID   string          `json:"strategy_uuid" bson:"strategy_uuid"`
	Ticker         string          `json:"ticker" bson:"ticker"`
	FromDate       string          `json:"from_date" bson:"from_date"`
	ToDate         string          `json:"to_date" bson:"to_date"`
	InitialBalance float64         `json:"initial_balance" bson:"initial_balance"`
	FinalBalance   float64         `json:"final_balance" bson:"final_balance"`
	ROI            float64         `json:"roi" bson:"roi"`
	TotalTrades    int             `json:"total_trades" bson:"total_trades"`
	WinningTrades  int             `json:"winning_trades" bson:"winning_trades"`
	LosingTrades   int             `json:"losing_trades" bson:"losing_trades"`
	MaxDrawdown    float64         `json:"max_drawdown" bson:"max_drawdown"`
	Trades         []BacktestTrade `json:"trades" bson:"trades"`
	AccountID      string          `json:"account_id" bson:"account_id"`
	PortfolioUUID  string          `json:"portfolio_uuid" bson:"portfolio_uuid"`
	CreatedAt      time.Time       `json:"created_at" bson:"created_at"`
}
