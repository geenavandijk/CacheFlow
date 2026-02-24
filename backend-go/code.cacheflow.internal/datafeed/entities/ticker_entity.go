package entities

import (

)

type TickerEntity struct {
	Ticker string `json:"ticker" bson:"ticker"`
	Name string `json:"name" bson:"name"`
	Market string `json:"market" bson:"market"`
	Locale string `json:"locale" bson:"locale"`
	CurrencySymbol string `json:"currency_symbol" bson:"currency_symbol"`
	CurrencyName string `json:"currency_name" bson:"currency_name"`
	BaseCurrencySymbol string `json:"base_currency_symbol" bson:"base_currency_symbol"`
	BaseCurrencyName string `json:"base_currency_name" bson:"base_currency_name"`
}