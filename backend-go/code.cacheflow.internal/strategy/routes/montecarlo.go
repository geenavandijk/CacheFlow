package routes

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strings"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	strategyEntities "code.cacheflow.internal/strategy/entities"
	"code.cacheflow.internal/util/httpx"

	"go.mongodb.org/mongo-driver/bson"
)

type monteCarloBody struct {
	BacktestUUID   string  `json:"backtest_uuid"`
	NumSimulations int     `json:"num_simulations"`
	InitialBalance float64 `json:"initial_balance"`
}

type MonteCarloResult struct {
	NumSimulations int       `json:"num_simulations"`
	InitialBalance float64   `json:"initial_balance"`
	Median         float64   `json:"median"`
	P10            float64   `json:"p10"`
	P25            float64   `json:"p25"`
	P75            float64   `json:"p75"`
	P90            float64   `json:"p90"`
	Worst          float64   `json:"worst"`
	Best           float64   `json:"best"`
	ProbProfit     float64   `json:"prob_profit"` // % of sims that ended above initial balance
	FinalValues    []float64 `json:"final_values"` // all simulation outcomes (for distribution chart)
}

func percentileF(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func RunMonteCarlo(res http.ResponseWriter, req *http.Request) {
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

	var body monteCarloBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("invalid request body", nil))
		return
	}

	body.BacktestUUID = strings.TrimSpace(body.BacktestUUID)
	if body.BacktestUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("backtest_uuid is required", nil))
		return
	}
	if body.NumSimulations <= 0 || body.NumSimulations > 10000 {
		body.NumSimulations = 1000
	}

	var bt strategyEntities.BacktestEntity
	if err := db.Collection(datastores.Backtests).FindOne(req.Context(),
		bson.M{"uuid": body.BacktestUUID, "account_id": *account.AccountID}).Decode(&bt); err != nil {
		httpx.WriteError(res, req, httpx.NotFound("backtest not found"))
		return
	}

	initialBalance := body.InitialBalance
	if initialBalance <= 0 {
		initialBalance = bt.InitialBalance
	}

	// Extract per-trade return multipliers from SELL trades
	var returnMultipliers []float64
	for _, t := range bt.Trades {
		if t.Type == "SELL" {
			// The fraction of capital that was in the trade: value = shares * buy_price
			// We approximate: each trade uses available cash; return multiplier from pnl_percent
			mult := 1 + t.PnLPercent/100
			returnMultipliers = append(returnMultipliers, mult)
		}
	}

	if len(returnMultipliers) == 0 {
		httpx.WriteError(res, req, httpx.BadRequest("no completed trades in this backtest", nil))
		return
	}

	// Run N simulations: shuffle trade order each time, compound returns
	finalValues := make([]float64, body.NumSimulations)
	shuffled := make([]float64, len(returnMultipliers))

	for i := 0; i < body.NumSimulations; i++ {
		copy(shuffled, returnMultipliers)
		rand.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })

		equity := initialBalance
		for _, mult := range shuffled {
			// Each trade: invest a fixed fraction of equity (simplified: 100% per trade)
			equity *= mult
			if equity < 0 {
				equity = 0
			}
		}
		finalValues[i] = equity
	}

	sort.Float64s(finalValues)

	profitCount := 0
	for _, v := range finalValues {
		if v > initialBalance {
			profitCount++
		}
	}

	result := MonteCarloResult{
		NumSimulations: body.NumSimulations,
		InitialBalance: initialBalance,
		Median:         percentileF(finalValues, 0.50),
		P10:            percentileF(finalValues, 0.10),
		P25:            percentileF(finalValues, 0.25),
		P75:            percentileF(finalValues, 0.75),
		P90:            percentileF(finalValues, 0.90),
		Worst:          finalValues[0],
		Best:           finalValues[len(finalValues)-1],
		ProbProfit:     float64(profitCount) / float64(body.NumSimulations) * 100,
		FinalValues:    finalValues,
	}

	httpx.WriteJSON(res, http.StatusOK, result)
}
