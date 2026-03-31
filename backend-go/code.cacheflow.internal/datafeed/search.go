package datafeed

import (
	"context"
	"net/http"
	"os"

	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/datafeed/entities"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SearchCompanies(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: "DATAFEED (SEARCH)",
	})

	q := req.URL.Query().Get("q")
	if q == "" {
		util.JSONResponse(res, http.StatusOK, []entities.TickerEntity{})
		return
	}

	filter := bson.M{
		"$or": bson.A{
			bson.M{"name":   bson.M{"$regex": q, "$options": "i"}},
			bson.M{"ticker": bson.M{"$regex": q, "$options": "i"}},
		},
	}

	cursor, err := datastores.GetMongoDatabase(req.Context()).Collection("tickers").Find(
		context.TODO(),
		filter,
		options.Find().SetLimit(10),
	)
	if err != nil {
		logger.Error("failed to search tickers", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to search tickers"))
		return
	}

	var results []entities.TickerEntity
	if err := cursor.All(context.TODO(), &results); err != nil {
		logger.Error("failed to decode results", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to decode results"))
		return
	}

	if results == nil {
		results = []entities.TickerEntity{}
	}

	util.JSONResponse(res, http.StatusOK, results)
}