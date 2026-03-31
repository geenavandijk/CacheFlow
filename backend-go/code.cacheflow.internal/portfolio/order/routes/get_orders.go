package routes

import (
	"net/http"
	"strconv"
	"strings"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	orderEntities "code.cacheflow.internal/portfolio/order/entities"
	"code.cacheflow.internal/util/httpx"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOrders returns paginated orders for a portfolio, optionally filtered by ticker.
// Query params:
// - portfolio_uuid (required)
// - ticker (optional)
// - page (optional, default 1)
// - limit (optional, default 20, max 100)
func GetOrders(res http.ResponseWriter, req *http.Request) {
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

	ticker := strings.TrimSpace(q.Get("ticker"))

	page := 1
	if v := strings.TrimSpace(q.Get("page")); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	limit := 20
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ordersCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Orders)

	filter := bson.M{
		"account_id":     account.AccountID,
		"portfolio_uuid": portfolioUUID,
	}
	if ticker != "" {
		filter["ticker"] = ticker
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cur, err := ordersCollection.Find(req.Context(), filter, findOpts)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}
	defer cur.Close(req.Context())

	var orders []*orderEntities.OrderEntity
	for cur.Next(req.Context()) {
		var o orderEntities.OrderEntity
		if err := cur.Decode(&o); err != nil {
			httpx.WriteError(res, req, httpx.BadRequest("failed to decode order", map[string]string{
				"email": email,
			}))
			return
		}
		orders = append(orders, &o)
	}
	if err := cur.Err(); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("failed to get orders", map[string]string{
			"email": email,
		}))
		return
	}

	response := map[string]any{
		"page":     page,
		"limit":    limit,
		"has_more": len(orders) == limit,
		"orders":   orders,
	}

	httpx.WriteJSON(res, http.StatusOK, response)
}

