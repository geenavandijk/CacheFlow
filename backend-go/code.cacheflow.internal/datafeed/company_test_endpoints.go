package datafeed

import (
	"context"
	"net/http"

	tickerEntities "code.cacheflow.internal/datafeed/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCompanyData(res http.ResponseWriter, req *http.Request) {

	// localhost:8080/v1/datafeed/company?ticker={Whatever user wants}
	// get the ticker from the request
	ticker := req.URL.Query().Get("ticker")

	// variable for tickers collection in the database
	tickersCollection := datastores.GetMongoDatabase(context.Background()).Collection("tickers")

	// Create empty pointer variable (null by derfault)
	var tickerObject *tickerEntities.TickerEntity

	// Query database for ticker, if found it will automatically assign the value to the pointer
	// 
	err := tickersCollection.FindOne(context.Background(), bson.M{
		"ticker": ticker,
	}, nil).Decode(&tickerObject)

	// if error is not null, see what happened 
	if err != nil {
		
		if err == mongo.ErrNoDocuments { // couldnt find the company with the given ticker
			httpx.WriteError(res, req, httpx.NotFound("Ticker not found"))
			return
		} else { // something is wrong with mongo's connection, SEVERE ERROR
			httpx.WriteError(res, req, httpx.Internal("Contact support, this should work"))
			return
		}
	}

	// RETURN ALL COMPANY DATA
	util.JSONResponse(res, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": tickerObject,
	})
}

