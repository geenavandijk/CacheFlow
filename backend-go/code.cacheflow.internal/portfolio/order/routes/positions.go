package routes

import (
	"net/http"
	"strings"

	accountEntities "code.cacheflow.internal/account/entities"
	"code.cacheflow.internal/datafeed"
	datastores "code.cacheflow.internal/datastores/mongo"
	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	"code.cacheflow.internal/util/httpx"

	"github.com/massive-com/client-go/v2/rest/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Position struct {
	Ticker       string  `json:"ticker"`
	Shares       int64   `json:"shares"`
	AvgCost      float64 `json:"avg_cost"`
	CurrentPrice float64 `json:"current_price"`
	Unrealized   float64 `json:"unrealized"`
}

// GetPositions returns active positions (per-ticker lots) with unrealized PnL for a portfolio.
// Query params:
// - portfolio_uuid (required)
// - ticker (optional; if provided, filters to that ticker)
func GetPositions(res http.ResponseWriter, req *http.Request) {
	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", map[string]string{
			"email": email,
		}))
		return
	}

	accountCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Accounts)

	var account *accountEntities.AccountEntity
	if err := accountCollection.FindOne(req.Context(), bson.M{"email": email}).Decode(&account); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", map[string]string{
			"email": email,
		}))
		return
	}

	q := req.URL.Query()
	portfolioUUID := strings.TrimSpace(q.Get("portfolio_uuid"))
	if portfolioUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}
	filterTicker := strings.TrimSpace(q.Get("ticker"))

	ordersCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Orders)

	filter := bson.M{
		"account_id":     account.AccountID,
		"portfolio_uuid": portfolioUUID,
	}
	if filterTicker != "" {
		filter["ticker"] = filterTicker
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cur, err := ordersCollection.Find(req.Context(), filter, findOpts)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}
	defer cur.Close(req.Context())

	type lot struct {
		Qty       int64
		CostPerSh float64
	}

	lotsByTicker := map[string][]lot{}

	for cur.Next(req.Context()) {
		var o orderEntities.OrderEntity
		if err := cur.Decode(&o); err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to decode order", map[string]string{
				"email": email,
			}))
			return
		}
		if o.Ticker == nil || o.Side == nil || o.Quantity == nil || o.Price == nil {
			continue
		}
		t := *o.Ticker
		switch *o.Side {
		case "BUY":
			if *o.Quantity > 0 {
				lotsByTicker[t] = append(lotsByTicker[t], lot{
					Qty:       *o.Quantity,
					CostPerSh: *o.Price,
				})
			}
		case "SELL":
			toSell := *o.Quantity
			if toSell <= 0 {
				continue
			}
			lots := lotsByTicker[t]
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
			lotsByTicker[t] = lots
		default:
			continue
		}
	}

	if err := cur.Err(); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}

	client := datafeed.GetMassiveClient()
	var positions []Position

	for ticker, lots := range lotsByTicker {
		var totalQty int64
		var totalCost float64
		for _, l := range lots {
			if l.Qty > 0 {
				totalQty += l.Qty
				totalCost += float64(l.Qty) * l.CostPerSh
			}
		}
		if totalQty <= 0 {
			continue
		}
		avgCost := totalCost / float64(totalQty)

		last, err := client.GetLastTrade(req.Context(), &models.GetLastTradeParams{
			Ticker: ticker,
		})
		if err != nil || last.ErrorMessage != "" {
			// If price lookup fails, skip unrealized calculation for this ticker.
			continue
		}
		currentPrice := last.Results.Price
		unrealized := float64(totalQty)*(currentPrice-avgCost)

		positions = append(positions, Position{
			Ticker:       ticker,
			Shares:       totalQty,
			AvgCost:      avgCost,
			CurrentPrice: currentPrice,
			Unrealized:   unrealized,
		})
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]any{
		"portfolio_uuid": portfolioUUID,
		"positions":      positions,
	})
}

