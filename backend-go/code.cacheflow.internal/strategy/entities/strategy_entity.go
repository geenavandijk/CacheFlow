package entities

import "time"

// RuleType is the kind of indicator condition to evaluate.
type RuleType string

const (
	RuleRSICrossAbove        RuleType = "RSI_CROSSES_ABOVE"
	RuleRSICrossBelow        RuleType = "RSI_CROSSES_BELOW"
	RuleRSIAbove             RuleType = "RSI_ABOVE"
	RuleRSIBelow             RuleType = "RSI_BELOW"
	RuleEMACrossAbove        RuleType = "EMA_CROSS_ABOVE"
	RuleEMACrossBelow        RuleType = "EMA_CROSS_BELOW"
	RuleSMACrossAbove        RuleType = "SMA_CROSS_ABOVE"
	RuleSMACrossBelow        RuleType = "SMA_CROSS_BELOW"
	RulePriceAboveEMA        RuleType = "PRICE_ABOVE_EMA"
	RulePriceBelowEMA        RuleType = "PRICE_BELOW_EMA"
	RulePriceAboveSMA        RuleType = "PRICE_ABOVE_SMA"
	RulePriceBelowSMA        RuleType = "PRICE_BELOW_SMA"
	RuleMACDCrossSignalAbove RuleType = "MACD_CROSS_SIGNAL_ABOVE"
	RuleMACDCrossSignalBelow RuleType = "MACD_CROSS_SIGNAL_BELOW"
	RuleMACDAboveZero        RuleType = "MACD_ABOVE_ZERO"
	RuleMACDBelowZero        RuleType = "MACD_BELOW_ZERO"
	RulePriceAboveVWAP       RuleType = "PRICE_ABOVE_VWAP_PCT"
	RulePriceBelowVWAP       RuleType = "PRICE_BELOW_VWAP_PCT"
)

// Rule is a single indicator-based buy/sell condition.
type Rule struct {
	Type          RuleType `json:"type" bson:"type"`
	Value         float64  `json:"value" bson:"value"`                   // threshold (RSI level, etc.)
	Window        int      `json:"window" bson:"window"`                  // period for RSI / single EMA / single SMA
	FastWindow    int      `json:"fast_window" bson:"fast_window"`        // fast period for EMA/SMA crossover
	SlowWindow    int      `json:"slow_window" bson:"slow_window"`        // slow period for EMA/SMA crossover
	FastPeriod    int      `json:"fast_period" bson:"fast_period"`        // MACD fast EMA
	SlowPeriod    int      `json:"slow_period" bson:"slow_period"`        // MACD slow EMA
	SignalPeriod  int      `json:"signal_period" bson:"signal_period"`    // MACD signal line
	VWAPDeviation float64  `json:"vwap_deviation" bson:"vwap_deviation"` // % above/below VWAP
}

// SellConditionType defines how a position exit is triggered.
type SellConditionType string

const (
	SellTakeProfit   SellConditionType = "TAKE_PROFIT"
	SellStopLoss     SellConditionType = "STOP_LOSS"
	SellTrailingStop SellConditionType = "TRAILING_STOP"
	SellIndicator    SellConditionType = "INDICATOR"
)

// SellCondition defines when to exit a position.
type SellCondition struct {
	Type    SellConditionType `json:"type" bson:"type"`
	Percent float64           `json:"percent" bson:"percent"`             // for TP / SL / trailing stop (%)
	Rule    *Rule             `json:"rule,omitempty" bson:"rule,omitempty"` // for indicator-based exits
}

// StrategyEntity is the persisted strategy document.
type StrategyEntity struct {
	UUID           string          `json:"uuid" bson:"uuid"`
	Name           string          `json:"name" bson:"name"`
	Description    string          `json:"description" bson:"description"`
	Ticker         string          `json:"ticker" bson:"ticker"`
	BuyRules       []Rule          `json:"buy_rules" bson:"buy_rules"`
	SellConditions []SellCondition `json:"sell_conditions" bson:"sell_conditions"`
	PortfolioUUID  string          `json:"portfolio_uuid" bson:"portfolio_uuid"`
	AccountID      string          `json:"account_id" bson:"account_id"`
	CreatedAt      time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" bson:"updated_at"`
}
