package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	strategyEntities "code.cacheflow.internal/strategy/entities"
	"code.cacheflow.internal/util/httpx"

	"github.com/pborman/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// ── Create ────────────────────────────────────────────────────────────────────

type createStrategyBody struct {
	Name           string                              `json:"name"`
	Description    string                              `json:"description"`
	Ticker         string                              `json:"ticker"`
	BuyRules       []strategyEntities.Rule             `json:"buy_rules"`
	SellConditions []strategyEntities.SellCondition    `json:"sell_conditions"`
	PortfolioUUID  string                              `json:"portfolio_uuid"`
}

func CreateStrategy(res http.ResponseWriter, req *http.Request) {
	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", nil))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())

	var account accountEntities.AccountEntity
	if err := db.Collection(datastores.Accounts).FindOne(req.Context(), bson.M{"email": email}).Decode(&account); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", nil))
		return
	}

	var body createStrategyBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("invalid request body", nil))
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Ticker = strings.ToUpper(strings.TrimSpace(body.Ticker))

	if body.Name == "" || body.Ticker == "" || body.PortfolioUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("name, ticker, and portfolio_uuid are required", nil))
		return
	}
	if len(body.BuyRules) == 0 {
		httpx.WriteError(res, req, httpx.BadRequest("at least one buy rule is required", nil))
		return
	}

	now := time.Now().UTC()
	strategy := strategyEntities.StrategyEntity{
		UUID:           uuid.New(),
		Name:           body.Name,
		Description:    body.Description,
		Ticker:         body.Ticker,
		BuyRules:       body.BuyRules,
		SellConditions: body.SellConditions,
		PortfolioUUID:  body.PortfolioUUID,
		AccountID:      *account.AccountID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := db.Collection(datastores.Strategies).InsertOne(req.Context(), strategy); err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to save strategy"))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, strategy)
}

// ── List ──────────────────────────────────────────────────────────────────────

func GetStrategies(res http.ResponseWriter, req *http.Request) {
	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", nil))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())

	var account accountEntities.AccountEntity
	if err := db.Collection(datastores.Accounts).FindOne(req.Context(), bson.M{"email": email}).Decode(&account); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", nil))
		return
	}

	portfolioUUID := strings.TrimSpace(req.URL.Query().Get("portfolio_uuid"))
	if portfolioUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("portfolio_uuid is required", nil))
		return
	}

	filter := bson.M{"account_id": *account.AccountID, "portfolio_uuid": portfolioUUID}
	cur, err := db.Collection(datastores.Strategies).Find(req.Context(), filter)
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to fetch strategies"))
		return
	}
	defer cur.Close(req.Context())

	var strategies []strategyEntities.StrategyEntity
	if err := cur.All(req.Context(), &strategies); err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to decode strategies"))
		return
	}
	if strategies == nil {
		strategies = []strategyEntities.StrategyEntity{}
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]any{"strategies": strategies})
}

// ── Update ────────────────────────────────────────────────────────────────────

type updateStrategyBody struct {
	UUID           string                              `json:"uuid"`
	Name           string                              `json:"name"`
	Description    string                              `json:"description"`
	Ticker         string                              `json:"ticker"`
	BuyRules       []strategyEntities.Rule             `json:"buy_rules"`
	SellConditions []strategyEntities.SellCondition    `json:"sell_conditions"`
}

func UpdateStrategy(res http.ResponseWriter, req *http.Request) {
	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", nil))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())

	var account accountEntities.AccountEntity
	if err := db.Collection(datastores.Accounts).FindOne(req.Context(), bson.M{"email": email}).Decode(&account); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", nil))
		return
	}

	var body updateStrategyBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("invalid request body", nil))
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Ticker = strings.ToUpper(strings.TrimSpace(body.Ticker))

	if body.UUID == "" || body.Name == "" || body.Ticker == "" {
		httpx.WriteError(res, req, httpx.BadRequest("uuid, name, and ticker are required", nil))
		return
	}
	if len(body.BuyRules) == 0 {
		httpx.WriteError(res, req, httpx.BadRequest("at least one buy rule is required", nil))
		return
	}

	filter := bson.M{"uuid": body.UUID, "account_id": *account.AccountID}
	update := bson.M{"$set": bson.M{
		"name":            body.Name,
		"description":     body.Description,
		"ticker":          body.Ticker,
		"buy_rules":       body.BuyRules,
		"sell_conditions": body.SellConditions,
		"updated_at":      time.Now().UTC(),
	}}

	result, err := db.Collection(datastores.Strategies).UpdateOne(req.Context(), filter, update)
	if err != nil || result.MatchedCount == 0 {
		httpx.WriteError(res, req, httpx.NotFound("strategy not found"))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]any{"updated": true})
}

// ── Delete ────────────────────────────────────────────────────────────────────

type deleteStrategyBody struct {
	UUID string `json:"uuid"`
}

func DeleteStrategy(res http.ResponseWriter, req *http.Request) {
	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", nil))
		return
	}

	db := datastores.GetMongoDatabase(req.Context())

	var account accountEntities.AccountEntity
	if err := db.Collection(datastores.Accounts).FindOne(req.Context(), bson.M{"email": email}).Decode(&account); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", nil))
		return
	}

	var body deleteStrategyBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("invalid request body", nil))
		return
	}

	if body.UUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("uuid is required", nil))
		return
	}

	filter := bson.M{"uuid": body.UUID, "account_id": *account.AccountID}
	result, err := db.Collection(datastores.Strategies).DeleteOne(req.Context(), filter)
	if err != nil || result.DeletedCount == 0 {
		httpx.WriteError(res, req, httpx.NotFound("strategy not found"))
		return
	}

	// Also delete associated backtests
	_, _ = db.Collection(datastores.Backtests).DeleteMany(req.Context(), bson.M{"strategy_uuid": body.UUID})

	httpx.WriteJSON(res, http.StatusOK, map[string]any{"deleted": true})
}
