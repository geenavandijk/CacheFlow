package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	portfolioEntities "code.cacheflow.internal/portfolio/management/entities"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/ptr"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreatePortfolioBody struct {
	Name *string `json:"name"`
	Description *string `json:"description"`
	StartingBalance *float64 `json:"starting_balance"`
}

func CreatePortfolio(res http.ResponseWriter, req *http.Request) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          "ORDER (EO)",
	})

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

	var body CreatePortfolioBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Error("failed to decode request body", "err", err)
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	if body.Name == nil || strings.TrimSpace(*body.Name) == "" {
		logger.Error("name is required")
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "name is required",
		})
		return
	}

	if body.Description == nil || strings.TrimSpace(*body.Description) == "" {
		logger.Error("description is required")
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "description is required",
		})
		return
	}

	if body.StartingBalance == nil || *body.StartingBalance <= 0 {
		logger.Error("starting balance is required and must be greater than 0")
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "starting balance is required and must be greater than 0",
		})
		return
	}

	db := datastores.GetMongoDatabase(req.Context())

	portfolioCollection := db.Collection(datastores.Portfolios)
	
	// check if a portfolio with the same name and same account ID already exists
	var existingPortfolio portfolioEntities.PortfolioEntity
	err = portfolioCollection.FindOne(req.Context(), bson.M{"name": *body.Name, "account_id": account.AccountID}).Decode(&existingPortfolio)
	if err != nil && err != mongo.ErrNoDocuments {
		httpx.WriteError(res, req, httpx.BadRequest("failed to check for existing portfolio", map[string]string{
			"name": *body.Name,
			"account_id": *account.AccountID,
		}))
		return
	}

	if err == nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio already exists", map[string]string{
			"name": *body.Name,
			"account_id": *account.AccountID,
		}))
		return
	}

	portfolio := &portfolioEntities.PortfolioEntity{
		UUID: ptr.String(primitive.NewObjectID().Hex()),
		AccountID: account.AccountID,
		Name: body.Name,
		Description: body.Description,
		StartingBalance: body.StartingBalance,
		CurrentBalance: body.StartingBalance,
		Orders: []*string{},
		Watchlists: []*portfolioEntities.WatchlistEntity{},
		CreatedAt: ptr.Time(time.Now()),
		UpdatedAt: ptr.Time(time.Now()),
	}

	if _, err := portfolioCollection.InsertOne(req.Context(), portfolio); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to create portfolio", map[string]string{
			"name": *body.Name,
			"account_id": *account.AccountID,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusCreated, portfolio)
}

func GetPortfolios(res http.ResponseWriter, req *http.Request) {

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

	portfolioCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Portfolios)

	var portfolios []*portfolioEntities.PortfolioEntity
	cursor, err := portfolioCollection.Find(req.Context(), bson.M{"account_id": account.AccountID})
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get portfolios", map[string]string{
			"account_id": *account.AccountID,
		}))
		return
	}
	defer cursor.Close(req.Context())
	
	for cursor.Next(req.Context()) {
		var portfolio portfolioEntities.PortfolioEntity
		if err := cursor.Decode(&portfolio); err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to decode portfolio", map[string]string{
				"account_id": *account.AccountID,
			}))
			return
		}
		portfolios = append(portfolios, &portfolio)
	}

	if err := cursor.Err(); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get portfolios", map[string]string{
			"account_id": *account.AccountID,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, portfolios)
}

type DeletePortfolioBody struct {
	UUID *string `json:"uuid"`
}

func DeletePortfolio(res http.ResponseWriter, req *http.Request) {

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

	var body DeletePortfolioBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}
	
	if body.UUID == nil || strings.TrimSpace(*body.UUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	portfolioCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Portfolios)

	_, err = portfolioCollection.DeleteOne(req.Context(), bson.M{"uuid": *body.UUID, "account_id": account.AccountID})
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to delete portfolio", map[string]string{
			"email": email,
			"uuid": *body.UUID,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]string{
		"message": "portfolio deleted successfully",
	})
}

type UpdatePortfolioBody struct {
	UUID *string `json:"uuid"`
	Name *string `json:"name"`
	Description *string `json:"description"`
	StartingBalance *float64 `json:"starting_balance"`
}

func UpdatePortfolio(res http.ResponseWriter, req *http.Request) {

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

	var body UpdatePortfolioBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.UUID == nil || strings.TrimSpace(*body.UUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("uuid is required", map[string]string{
			"email": email,
		}))
		return
	}
	
	if body.Name == nil || strings.TrimSpace(*body.Name) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("name is required", map[string]string{
			"email": email,
		}))
		return
	}
	
	if body.Description == nil || strings.TrimSpace(*body.Description) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("description is required", map[string]string{
			"email": email,
		}))
		return	
	}

	portfolioCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Portfolios)

	// Load existing portfolio to determine if starting/current balance can be changed.
	var existing portfolioEntities.PortfolioEntity
	err = portfolioCollection.FindOne(req.Context(), bson.M{"uuid": *body.UUID, "account_id": account.AccountID}).Decode(&existing)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
			"uuid":  *body.UUID,
		}))
		return
	}

	update := bson.M{
		"name":        *body.Name,
		"description": *body.Description,
		"updated_at":  time.Now(),
	}

	// Only allow changing starting/current balance if they are currently equal (no activity yet).
	if body.StartingBalance != nil &&
		existing.StartingBalance != nil &&
		existing.CurrentBalance != nil &&
		*existing.StartingBalance == *existing.CurrentBalance {
		update["starting_balance"] = *body.StartingBalance
		update["current_balance"] = *body.StartingBalance
	}

	_, err = portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{"uuid": *body.UUID, "account_id": account.AccountID},
		bson.M{"$set": update},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to update portfolio", map[string]string{
			"email": email,
			"uuid":  *body.UUID,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]string{
		"message": "portfolio updated successfully",
	})
}

// GetBuyingPower calculates buying power as: starting_balance - total cost basis of all active positions.
// It walks orders for the portfolio and uses per-ticker FIFO lots to determine remaining invested capital.
func GetBuyingPower(res http.ResponseWriter, req *http.Request) {

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

	portfolioUUID := strings.TrimSpace(req.URL.Query().Get("portfolio_uuid"))
	if portfolioUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	var portfolio portfolioEntities.PortfolioEntity
	if err := portfolioCollection.FindOne(req.Context(), bson.M{"uuid": portfolioUUID, "account_id": account.AccountID}).Decode(&portfolio); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
		}))
		return
	}

	if portfolio.StartingBalance == nil {
		httpx.WriteJSON(res, http.StatusOK, map[string]any{
			"portfolio_uuid":   portfolioUUID,
			"starting_balance": 0,
			"invested":         0,
			"buying_power":     0,
		})
		return
	}

	type lot struct {
		Qty       int64
		CostPerSh float64
	}

	ordersCollection := db.Collection(datastores.Orders)

	filter := bson.M{
		"account_id":     account.AccountID,
		"portfolio_uuid": portfolioUUID,
	}

	cursor, err := ordersCollection.Find(req.Context(), filter, nil)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}
	defer cursor.Close(req.Context())

	lotsByTicker := map[string][]lot{}

	for cursor.Next(req.Context()) {
		var o orderEntities.OrderEntity
		if err := cursor.Decode(&o); err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to decode order", map[string]string{
				"email": email,
			}))
			return
		}
		if o.Ticker == nil || o.Side == nil || o.Quantity == nil || o.Price == nil {
			continue
		}
		ticker := *o.Ticker
		switch *o.Side {
		case "BUY":
			if *o.Quantity > 0 {
				lotsByTicker[ticker] = append(lotsByTicker[ticker], lot{
					Qty:       *o.Quantity,
					CostPerSh: *o.Price,
				})
			}
		case "SELL":
			toSell := *o.Quantity
			if toSell <= 0 {
				continue
			}
			lots := lotsByTicker[ticker]
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
			lotsByTicker[ticker] = lots
		default:
			continue
		}
	}

	if err := cursor.Err(); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}

	var invested float64
	for _, lots := range lotsByTicker {
		for _, l := range lots {
			if l.Qty > 0 {
				invested += float64(l.Qty) * l.CostPerSh
			}
		}
	}

	start := *portfolio.StartingBalance
	buyingPower := start - invested
	if buyingPower < 0 {
		buyingPower = 0
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]any{
		"portfolio_uuid":   portfolioUUID,
		"starting_balance": start,
		"invested":         invested,
		"buying_power":     buyingPower,
	})
}

// Watchlist management

type CreateWatchlistBody struct {
	PortfolioUUID *string   `json:"portfolio_uuid"`
	Name          *string   `json:"name"`
	Tickers       *[]string `json:"tickers"`
}

func CreateWatchlist(res http.ResponseWriter, req *http.Request) {
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

	var body CreateWatchlistBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.PortfolioUUID == nil || strings.TrimSpace(*body.PortfolioUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	if body.Name == nil || strings.TrimSpace(*body.Name) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("name is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	var portfolio portfolioEntities.PortfolioEntity
	if err := portfolioCollection.FindOne(req.Context(), bson.M{
		"uuid":       *body.PortfolioUUID,
		"account_id": account.AccountID,
	}).Decode(&portfolio); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
		}))
		return
	}

	var nextIndex int32 = 0
	if len(portfolio.Watchlists) > 0 {
		for _, wl := range portfolio.Watchlists {
			if wl != nil && wl.Index != nil && *wl.Index >= nextIndex {
				nextIndex = *wl.Index + 1
			}
		}
	}

	now := time.Now()
	var tickers []*string
	if body.Tickers != nil {
		for _, t := range *body.Tickers {
			tt := strings.ToUpper(strings.TrimSpace(t))
			if tt == "" {
				continue
			}
			tickers = append(tickers, ptr.String(tt))
		}
	}

	watchlist := &portfolioEntities.WatchlistEntity{
		UUID:      ptr.String(primitive.NewObjectID().Hex()),
		Name:      body.Name,
		Index:     ptr.Int32(nextIndex),
		Tickers:   tickers,
		CreatedAt: ptr.Time(now),
		UpdatedAt: ptr.Time(now),
	}

	_, err := portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{"uuid": *body.PortfolioUUID, "account_id": account.AccountID},
		bson.M{"$push": bson.M{"watchlists": watchlist}},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to create watchlist", map[string]string{
			"email": email,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusCreated, watchlist)
}

func GetWatchlists(res http.ResponseWriter, req *http.Request) {
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

	portfolioUUID := strings.TrimSpace(req.URL.Query().Get("portfolio_uuid"))
	if portfolioUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	var portfolio portfolioEntities.PortfolioEntity
	if err := portfolioCollection.FindOne(req.Context(), bson.M{
		"uuid":       portfolioUUID,
		"account_id": account.AccountID,
	}).Decode(&portfolio); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
		}))
		return
	}

	watchlists := portfolio.Watchlists
	// sort in-memory by Index to ensure stable order
	if len(watchlists) > 1 {
		for i := 0; i < len(watchlists); i++ {
			for j := i + 1; j < len(watchlists); j++ {
				if watchlists[i] == nil || watchlists[j] == nil {
					continue
				}
				ii, jj := int32(0), int32(0)
				if watchlists[i].Index != nil {
					ii = *watchlists[i].Index
				}
				if watchlists[j].Index != nil {
					jj = *watchlists[j].Index
				}
				if jj < ii {
					watchlists[i], watchlists[j] = watchlists[j], watchlists[i]
				}
			}
		}
	}

	httpx.WriteJSON(res, http.StatusOK, watchlists)
}

type UpdateWatchlistBody struct {
	PortfolioUUID *string   `json:"portfolio_uuid"`
	WatchlistUUID *string   `json:"watchlist_uuid"`
	Name          *string   `json:"name"`
	Tickers       *[]string `json:"tickers"`
}

func UpdateWatchlist(res http.ResponseWriter, req *http.Request) {
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

	var body UpdateWatchlistBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.PortfolioUUID == nil || strings.TrimSpace(*body.PortfolioUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}
	if body.WatchlistUUID == nil || strings.TrimSpace(*body.WatchlistUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("watchlist_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	update := bson.M{
		"watchlists.$.updated_at": time.Now(),
	}
	if body.Name != nil && strings.TrimSpace(*body.Name) != "" {
		update["watchlists.$.name"] = strings.TrimSpace(*body.Name)
	}
	if body.Tickers != nil {
		var tickers []*string
		for _, t := range *body.Tickers {
			tt := strings.ToUpper(strings.TrimSpace(t))
			if tt == "" {
				continue
			}
			tickers = append(tickers, ptr.String(tt))
		}
		update["watchlists.$.tickers"] = tickers
	}

	_, err := portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{
			"uuid":             *body.PortfolioUUID,
			"account_id":       account.AccountID,
			"watchlists.uuid":  *body.WatchlistUUID,
		},
		bson.M{"$set": update},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to update watchlist", map[string]string{
			"email": email,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]string{
		"message": "watchlist updated successfully",
	})
}

type DeleteWatchlistBody struct {
	PortfolioUUID *string `json:"portfolio_uuid"`
	WatchlistUUID *string `json:"watchlist_uuid"`
}

func DeleteWatchlist(res http.ResponseWriter, req *http.Request) {
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

	var body DeleteWatchlistBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.PortfolioUUID == nil || strings.TrimSpace(*body.PortfolioUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}
	if body.WatchlistUUID == nil || strings.TrimSpace(*body.WatchlistUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("watchlist_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	_, err := portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{"uuid": *body.PortfolioUUID, "account_id": account.AccountID},
		bson.M{
			"$pull": bson.M{
				"watchlists": bson.M{"uuid": *body.WatchlistUUID},
			},
		},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to delete watchlist", map[string]string{
			"email": email,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]string{
		"message": "watchlist deleted successfully",
	})
}

type ReorderWatchlistsBody struct {
	PortfolioUUID *string   `json:"portfolio_uuid"`
	Order         *[]string `json:"order"`
}

func ReorderWatchlists(res http.ResponseWriter, req *http.Request) {
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

	var body ReorderWatchlistsBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to decode request body", map[string]string{
			"email": email,
		}))
		return
	}

	if body.PortfolioUUID == nil || strings.TrimSpace(*body.PortfolioUUID) == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", map[string]string{
			"email": email,
		}))
		return
	}
	if body.Order == nil {
		httpx.WriteError(res, req, httpx.BadRequest("order is required", map[string]string{
			"email": email,
		}))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())
	portfolioCollection := db.Collection(datastores.Portfolios)

	var portfolio portfolioEntities.PortfolioEntity
	if err := portfolioCollection.FindOne(req.Context(), bson.M{
		"uuid":       *body.PortfolioUUID,
		"account_id": account.AccountID,
	}).Decode(&portfolio); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio not found", map[string]string{
			"email": email,
		}))
		return
	}

	order := *body.Order
	indexByID := make(map[string]int32, len(order))
	for i, id := range order {
		indexByID[id] = int32(i)
	}

	for _, wl := range portfolio.Watchlists {
		if wl == nil || wl.UUID == nil {
			continue
		}
		if idx, ok := indexByID[*wl.UUID]; ok {
			wl.Index = ptr.Int32(idx)
		}
	}

	_, err := portfolioCollection.UpdateOne(
		req.Context(),
		bson.M{"uuid": *body.PortfolioUUID, "account_id": account.AccountID},
		bson.M{
			"$set": bson.M{
				"watchlists": portfolio.Watchlists,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to reorder watchlists", map[string]string{
			"email": email,
		}))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]string{
		"message": "watchlists reordered successfully",
	})
}