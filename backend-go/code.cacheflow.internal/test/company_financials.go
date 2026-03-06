package test

import (
	"context"
	"log"

	"code.cacheflow.internal/datafeed"
	"github.com/massive-com/client-go/v2/rest/models"
)

func GetCompanySnapshot(company string) {
	client := datafeed.GetMassiveClient()

	params := models.GetTickerSnapshotParams{
		Ticker:     company,
		Locale:     "us",
		MarketType: "stocks",
	}

	res, err := client.GetTickerSnapshot(context.Background(), &params)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(res)
}