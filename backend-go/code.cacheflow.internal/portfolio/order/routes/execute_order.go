package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	"code.cacheflow.internal/datafeed"
	datastores "code.cacheflow.internal/datastores/mongo"
	portfolioEntities "code.cacheflow.internal/portfolio/management/entities"
	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	orderHandler "code.cacheflow.internal/portfolio/order/handler"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/ptr"

	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/pborman/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type ExecuteOrderBody struct {
	Ticker *string `json:"ticker"`
	Side *string `json:"side"`
	Quantity *int64 `json:"quantity"`
	PortfolioUUID *string `json:"portfolio_uuid"`
}

func ExecuteOrder(res http.ResponseWriter, req *http.Request) {

	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", map[string]string{
			"email": email,
		}))
		return
	}

	accountCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Accounts)

	var account *accountEntities.AccountEntity
	err := accountCollection.FindOne(req.Context(), bson.M{"email": email}).Decode(&account)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", map[string]string{
			"email": email,
		}))
		return
	}

	var body ExecuteOrderBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.Ticker == nil || strings.TrimSpace(*body.Ticker) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("ticker is required", map[string]string{
			"email": email,
		}))
		return
	}

	// side is either BUY or SELL
	if body.Side == nil || (*body.Side != "BUY" && *body.Side != "SELL") {
		httpx.WriteError(res, req, httpx.BadRequest("side is required and must be either BUY or SELL", map[string]string{
			"email": email,
		}))
		return
	}

	if body.Quantity == nil || *body.Quantity <= 0 {
		httpx.WriteError(res, req, httpx.BadRequest("quantity is required and must be greater than 0", map[string]string{
			"email": email,
		}))
		return
	}

	if body.PortfolioUUID == nil || strings.TrimSpace(*body.PortfolioUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	c := datafeed.GetMassiveClient()

	resp, err := c.GetLastTrade(
		req.Context(),
		&models.GetLastTradeParams{
			Ticker: *body.Ticker,
		},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get stock snapshot", map[string]string{
			"email": email,
		}))
		return
	}

	if resp.ErrorMessage != "" {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get stock snapshot", map[string]string{
			"email": email,
		}))
		return
	}

	currentPrice := resp.Results.Price

	totalCost := float64(*body.Quantity) * currentPrice

	portfolioCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Portfolios)

	var portfolio portfolioEntities.PortfolioEntity
	err = portfolioCollection.FindOne(req.Context(), bson.M{"uuid": *body.PortfolioUUID, "account_id": account.AccountID}).Decode(&portfolio)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
		}))
		return	
	}

	// make sure portfolio belongs to account
	if *portfolio.AccountID != *account.AccountID {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio does not belong to account", map[string]string{
			"email": email,
		}))
		return
	}

	var realizedPtr *float64

	if *body.Side == "BUY" {
		if *portfolio.CurrentBalance < totalCost {
			httpx.WriteError(res, req, httpx.BadRequest("insufficient funds", map[string]string{
				"email": email,
			}))
			return
		}
	} else {

		// check if the portfolio has enough shares to sell
		activeShares, err := orderHandler.GetActiveShares(*body.Ticker, *body.PortfolioUUID)
		if err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to get active shares", map[string]string{
				"email": email,
			}))
			return
		}

		if activeShares < *body.Quantity {
			httpx.WriteError(res, req, httpx.BadRequest("insufficient shares", map[string]string{
				"email": email,
			}))
			return
		}

		// calculate realized PnL for this sell using FIFO
		realized, err := orderHandler.CalculateRealizedPnLForSell(req.Context(), *body.Ticker, *body.PortfolioUUID, *body.Quantity, currentPrice)
		if err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to calculate realized PnL", map[string]string{
				"email": email,
			}))
			return
		}
		realizedPtr = &realized
	}

	// create order
	order := &orderEntities.OrderEntity{
		UUID: ptr.String(uuid.NewRandom().String()),
		Ticker: body.Ticker,
		Side: body.Side,
		Quantity: body.Quantity,
		Price: &currentPrice,
		TotalCost: &totalCost,
		Realized: realizedPtr,
		Timestamp: ptr.Time(time.Now()),
		AccountID: account.AccountID,
		PortfolioUUID: body.PortfolioUUID,
	}

	ordersCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Orders)

	_, err = ordersCollection.InsertOne(req.Context(), order)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to create order", map[string]string{
			"email": email,
		}))
		return
	}

	// Update portfolio current balance: decrease on BUY, increase on SELL by trade notional.
	delta := totalCost
	if *body.Side == "BUY" {
		delta = -totalCost
	}

	_, err = portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{"uuid": *body.PortfolioUUID, "account_id": account.AccountID},
		bson.M{
			"$inc": bson.M{"current_balance": delta},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to update portfolio balance", map[string]string{
			"email": email,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusCreated, order)
}