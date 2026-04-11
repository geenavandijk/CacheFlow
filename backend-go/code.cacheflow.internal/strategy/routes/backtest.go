package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	strategyEntities "code.cacheflow.internal/strategy/entities"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/secrets"

	"github.com/pborman/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

const massiveBase = "https://api.massive.com"

// ── Helpers: Massive API calls ────────────────────────────────────────────────

type dailyBar struct {
	Date  string
	Open  float64
	High  float64
	Low   float64
	Close float64
	Vol   float64
	VWAP  float64 // vw field from Massive
	TsMs  int64
}

func msToDate(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02")
}

func fetchDailyBars(ctx *http.Request, ticker, from, to string) ([]dailyBar, error) {
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/1/day/%s/%s?apiKey=%s&sort=asc&limit=5000&adjusted=true",
		massiveBase, ticker, from, to, secrets.MassiveMainApiKeyValue)

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw struct {
		Results []struct {
			T  int64   `json:"t"`
			O  float64 `json:"o"`
			H  float64 `json:"h"`
			L  float64 `json:"l"`
			C  float64 `json:"c"`
			V  float64 `json:"v"`
			VW float64 `json:"vw"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	bars := make([]dailyBar, 0, len(raw.Results))
	for _, r := range raw.Results {
		bars = append(bars, dailyBar{
			Date:  msToDate(r.T),
			Open:  r.O,
			High:  r.H,
			Low:   r.L,
			Close: r.C,
			Vol:   r.V,
			VWAP:  r.VW,
			TsMs:  r.T,
		})
	}
	return bars, nil
}

// fetchSimpleIndicator fetches RSI / EMA / SMA values and returns date→value map.
func fetchSimpleIndicator(ctx *http.Request, kind, ticker, from, to string, window int) (map[string]float64, error) {
	url := fmt.Sprintf(
		"%s/v1/indicators/%s/%s?timespan=day&window=%d&order=asc&limit=5000&adjusted=true&timestamp.gte=%s&timestamp.lte=%s&apiKey=%s",
		massiveBase, kind, ticker, window, from, to, secrets.MassiveMainApiKeyValue)

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw struct {
		Results struct {
			Values []struct {
				Timestamp int64   `json:"timestamp"`
				Value     float64 `json:"value"`
			} `json:"values"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	m := make(map[string]float64, len(raw.Results.Values))
	for _, v := range raw.Results.Values {
		m[msToDate(v.Timestamp)] = v.Value
	}
	return m, nil
}

type macdPoint struct {
	Value     float64
	Signal    float64
	Histogram float64
}

func fetchMACDIndicator(ctx *http.Request, ticker, from, to string, fast, slow, signal int) (map[string]macdPoint, error) {
	url := fmt.Sprintf(
		"%s/v1/indicators/macd/%s?timespan=day&fast_period=%d&slow_period=%d&signal_period=%d&order=asc&limit=5000&adjusted=true&timestamp.gte=%s&timestamp.lte=%s&apiKey=%s",
		massiveBase, ticker, fast, slow, signal, from, to, secrets.MassiveMainApiKeyValue)

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw struct {
		Results struct {
			Values []struct {
				Timestamp int64   `json:"timestamp"`
				Value     float64 `json:"value"`
				Signal    float64 `json:"signal"`
				Histogram float64 `json:"histogram"`
			} `json:"values"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	m := make(map[string]macdPoint, len(raw.Results.Values))
	for _, v := range raw.Results.Values {
		m[msToDate(v.Timestamp)] = macdPoint{Value: v.Value, Signal: v.Signal, Histogram: v.Histogram}
	}
	return m, nil
}

// ── Backtest Engine ───────────────────────────────────────────────────────────

type indicatorStore struct {
	simple map[string]map[string]float64 // "RSI_14" → date → value
	macd   map[string]map[string]macdPoint
}

func (s *indicatorStore) get(key, date string) (float64, bool) {
	if m, ok := s.simple[key]; ok {
		if v, ok := m[date]; ok {
			return v, true
		}
	}
	return 0, false
}

func (s *indicatorStore) getMACD(key, date string) (macdPoint, bool) {
	if m, ok := s.macd[key]; ok {
		if v, ok := m[date]; ok {
			return v, true
		}
	}
	return macdPoint{}, false
}

func loadIndicators(req *http.Request, strategy *strategyEntities.StrategyEntity, from, to string) (*indicatorStore, error) {
	store := &indicatorStore{
		simple: make(map[string]map[string]float64),
		macd:   make(map[string]map[string]macdPoint),
	}

	// Gather all rules including indicator-based sell conditions
	allRules := make([]strategyEntities.Rule, len(strategy.BuyRules))
	copy(allRules, strategy.BuyRules)
	for _, sc := range strategy.SellConditions {
		if sc.Type == strategyEntities.SellIndicator && sc.Rule != nil {
			allRules = append(allRules, *sc.Rule)
		}
	}

	for _, rule := range allRules {
		switch rule.Type {
		case strategyEntities.RuleRSICrossAbove, strategyEntities.RuleRSICrossBelow,
			strategyEntities.RuleRSIAbove, strategyEntities.RuleRSIBelow:
			w := rule.Window
			if w == 0 {
				w = 14
			}
			key := fmt.Sprintf("RSI_%d", w)
			if _, ok := store.simple[key]; !ok {
				vals, err := fetchSimpleIndicator(req, "rsi", strategy.Ticker, from, to, w)
				if err != nil {
					return nil, fmt.Errorf("fetch RSI_%d: %w", w, err)
				}
				store.simple[key] = vals
			}

		case strategyEntities.RuleEMACrossAbove, strategyEntities.RuleEMACrossBelow:
			for _, w := range []int{rule.FastWindow, rule.SlowWindow} {
				if w == 0 {
					continue
				}
				key := fmt.Sprintf("EMA_%d", w)
				if _, ok := store.simple[key]; !ok {
					vals, err := fetchSimpleIndicator(req, "ema", strategy.Ticker, from, to, w)
					if err != nil {
						return nil, fmt.Errorf("fetch EMA_%d: %w", w, err)
					}
					store.simple[key] = vals
				}
			}

		case strategyEntities.RuleSMACrossAbove, strategyEntities.RuleSMACrossBelow:
			for _, w := range []int{rule.FastWindow, rule.SlowWindow} {
				if w == 0 {
					continue
				}
				key := fmt.Sprintf("SMA_%d", w)
				if _, ok := store.simple[key]; !ok {
					vals, err := fetchSimpleIndicator(req, "sma", strategy.Ticker, from, to, w)
					if err != nil {
						return nil, fmt.Errorf("fetch SMA_%d: %w", w, err)
					}
					store.simple[key] = vals
				}
			}

		case strategyEntities.RulePriceAboveEMA, strategyEntities.RulePriceBelowEMA:
			w := rule.Window
			if w == 0 {
				w = 9
			}
			key := fmt.Sprintf("EMA_%d", w)
			if _, ok := store.simple[key]; !ok {
				vals, err := fetchSimpleIndicator(req, "ema", strategy.Ticker, from, to, w)
				if err != nil {
					return nil, fmt.Errorf("fetch EMA_%d: %w", w, err)
				}
				store.simple[key] = vals
			}

		case strategyEntities.RulePriceAboveSMA, strategyEntities.RulePriceBelowSMA:
			w := rule.Window
			if w == 0 {
				w = 20
			}
			key := fmt.Sprintf("SMA_%d", w)
			if _, ok := store.simple[key]; !ok {
				vals, err := fetchSimpleIndicator(req, "sma", strategy.Ticker, from, to, w)
				if err != nil {
					return nil, fmt.Errorf("fetch SMA_%d: %w", w, err)
				}
				store.simple[key] = vals
			}

		case strategyEntities.RuleMACDCrossSignalAbove, strategyEntities.RuleMACDCrossSignalBelow,
			strategyEntities.RuleMACDAboveZero, strategyEntities.RuleMACDBelowZero:
			fast, slow, sig := rule.FastPeriod, rule.SlowPeriod, rule.SignalPeriod
			if fast == 0 {
				fast = 12
			}
			if slow == 0 {
				slow = 26
			}
			if sig == 0 {
				sig = 9
			}
			key := fmt.Sprintf("MACD_%d_%d_%d", fast, slow, sig)
			if _, ok := store.macd[key]; !ok {
				vals, err := fetchMACDIndicator(req, strategy.Ticker, from, to, fast, slow, sig)
				if err != nil {
					return nil, fmt.Errorf("fetch MACD: %w", err)
				}
				store.macd[key] = vals
			}
		}
	}

	return store, nil
}

// evaluateRule checks a single rule against the current and previous bar.
// prevDate is "" when bar is the first bar (no previous available).
func evaluateRule(rule strategyEntities.Rule, bar *dailyBar, prevDate string, store *indicatorStore) bool {
	hasPrev := prevDate != ""

	switch rule.Type {
	// ── RSI ──────────────────────────────────────────────────────────────────
	case strategyEntities.RuleRSIAbove:
		w := rule.Window
		if w == 0 {
			w = 14
		}
		v, ok := store.get(fmt.Sprintf("RSI_%d", w), bar.Date)
		return ok && v > rule.Value

	case strategyEntities.RuleRSIBelow:
		w := rule.Window
		if w == 0 {
			w = 14
		}
		v, ok := store.get(fmt.Sprintf("RSI_%d", w), bar.Date)
		return ok && v < rule.Value

	case strategyEntities.RuleRSICrossAbove:
		if !hasPrev {
			return false
		}
		w := rule.Window
		if w == 0 {
			w = 14
		}
		key := fmt.Sprintf("RSI_%d", w)
		curr, okC := store.get(key, bar.Date)
		prev, okP := store.get(key, prevDate)
		return okC && okP && prev <= rule.Value && curr > rule.Value

	case strategyEntities.RuleRSICrossBelow:
		if !hasPrev {
			return false
		}
		w := rule.Window
		if w == 0 {
			w = 14
		}
		key := fmt.Sprintf("RSI_%d", w)
		curr, okC := store.get(key, bar.Date)
		prev, okP := store.get(key, prevDate)
		return okC && okP && prev >= rule.Value && curr < rule.Value

	// ── EMA crossover ─────────────────────────────────────────────────────────
	case strategyEntities.RuleEMACrossAbove:
		if !hasPrev || rule.FastWindow == 0 || rule.SlowWindow == 0 {
			return false
		}
		fKey := fmt.Sprintf("EMA_%d", rule.FastWindow)
		sKey := fmt.Sprintf("EMA_%d", rule.SlowWindow)
		fCurr, ok1 := store.get(fKey, bar.Date)
		sCurr, ok2 := store.get(sKey, bar.Date)
		fPrev, ok3 := store.get(fKey, prevDate)
		sPrev, ok4 := store.get(sKey, prevDate)
		return ok1 && ok2 && ok3 && ok4 && fPrev <= sPrev && fCurr > sCurr

	case strategyEntities.RuleEMACrossBelow:
		if !hasPrev || rule.FastWindow == 0 || rule.SlowWindow == 0 {
			return false
		}
		fKey := fmt.Sprintf("EMA_%d", rule.FastWindow)
		sKey := fmt.Sprintf("EMA_%d", rule.SlowWindow)
		fCurr, ok1 := store.get(fKey, bar.Date)
		sCurr, ok2 := store.get(sKey, bar.Date)
		fPrev, ok3 := store.get(fKey, prevDate)
		sPrev, ok4 := store.get(sKey, prevDate)
		return ok1 && ok2 && ok3 && ok4 && fPrev >= sPrev && fCurr < sCurr

	// ── SMA crossover ─────────────────────────────────────────────────────────
	case strategyEntities.RuleSMACrossAbove:
		if !hasPrev || rule.FastWindow == 0 || rule.SlowWindow == 0 {
			return false
		}
		fKey := fmt.Sprintf("SMA_%d", rule.FastWindow)
		sKey := fmt.Sprintf("SMA_%d", rule.SlowWindow)
		fCurr, ok1 := store.get(fKey, bar.Date)
		sCurr, ok2 := store.get(sKey, bar.Date)
		fPrev, ok3 := store.get(fKey, prevDate)
		sPrev, ok4 := store.get(sKey, prevDate)
		return ok1 && ok2 && ok3 && ok4 && fPrev <= sPrev && fCurr > sCurr

	case strategyEntities.RuleSMACrossBelow:
		if !hasPrev || rule.FastWindow == 0 || rule.SlowWindow == 0 {
			return false
		}
		fKey := fmt.Sprintf("SMA_%d", rule.FastWindow)
		sKey := fmt.Sprintf("SMA_%d", rule.SlowWindow)
		fCurr, ok1 := store.get(fKey, bar.Date)
		sCurr, ok2 := store.get(sKey, bar.Date)
		fPrev, ok3 := store.get(fKey, prevDate)
		sPrev, ok4 := store.get(sKey, prevDate)
		return ok1 && ok2 && ok3 && ok4 && fPrev >= sPrev && fCurr < sCurr

	// ── Price vs EMA/SMA ─────────────────────────────────────────────────────
	case strategyEntities.RulePriceAboveEMA:
		w := rule.Window
		if w == 0 {
			w = 9
		}
		ema, ok := store.get(fmt.Sprintf("EMA_%d", w), bar.Date)
		return ok && bar.Close > ema

	case strategyEntities.RulePriceBelowEMA:
		w := rule.Window
		if w == 0 {
			w = 9
		}
		ema, ok := store.get(fmt.Sprintf("EMA_%d", w), bar.Date)
		return ok && bar.Close < ema

	case strategyEntities.RulePriceAboveSMA:
		w := rule.Window
		if w == 0 {
			w = 20
		}
		sma, ok := store.get(fmt.Sprintf("SMA_%d", w), bar.Date)
		return ok && bar.Close > sma

	case strategyEntities.RulePriceBelowSMA:
		w := rule.Window
		if w == 0 {
			w = 20
		}
		sma, ok := store.get(fmt.Sprintf("SMA_%d", w), bar.Date)
		return ok && bar.Close < sma

	// ── MACD ─────────────────────────────────────────────────────────────────
	case strategyEntities.RuleMACDCrossSignalAbove:
		if !hasPrev {
			return false
		}
		fast, slow, sig := rule.FastPeriod, rule.SlowPeriod, rule.SignalPeriod
		if fast == 0 {
			fast, slow, sig = 12, 26, 9
		}
		key := fmt.Sprintf("MACD_%d_%d_%d", fast, slow, sig)
		curr, ok1 := store.getMACD(key, bar.Date)
		prev, ok2 := store.getMACD(key, prevDate)
		return ok1 && ok2 && prev.Value <= prev.Signal && curr.Value > curr.Signal

	case strategyEntities.RuleMACDCrossSignalBelow:
		if !hasPrev {
			return false
		}
		fast, slow, sig := rule.FastPeriod, rule.SlowPeriod, rule.SignalPeriod
		if fast == 0 {
			fast, slow, sig = 12, 26, 9
		}
		key := fmt.Sprintf("MACD_%d_%d_%d", fast, slow, sig)
		curr, ok1 := store.getMACD(key, bar.Date)
		prev, ok2 := store.getMACD(key, prevDate)
		return ok1 && ok2 && prev.Value >= prev.Signal && curr.Value < curr.Signal

	case strategyEntities.RuleMACDAboveZero:
		fast, slow, sig := rule.FastPeriod, rule.SlowPeriod, rule.SignalPeriod
		if fast == 0 {
			fast, slow, sig = 12, 26, 9
		}
		key := fmt.Sprintf("MACD_%d_%d_%d", fast, slow, sig)
		p, ok := store.getMACD(key, bar.Date)
		return ok && p.Histogram > 0

	case strategyEntities.RuleMACDBelowZero:
		fast, slow, sig := rule.FastPeriod, rule.SlowPeriod, rule.SignalPeriod
		if fast == 0 {
			fast, slow, sig = 12, 26, 9
		}
		key := fmt.Sprintf("MACD_%d_%d_%d", fast, slow, sig)
		p, ok := store.getMACD(key, bar.Date)
		return ok && p.Histogram < 0

	// ── VWAP ─────────────────────────────────────────────────────────────────
	case strategyEntities.RulePriceAboveVWAP:
		if bar.VWAP == 0 {
			return false
		}
		dev := rule.VWAPDeviation
		threshold := bar.VWAP * (1 + dev/100)
		return bar.Close > threshold

	case strategyEntities.RulePriceBelowVWAP:
		if bar.VWAP == 0 {
			return false
		}
		dev := rule.VWAPDeviation
		threshold := bar.VWAP * (1 - dev/100)
		return bar.Close < threshold
	}

	return false
}

func allBuyRulesMet(rules []strategyEntities.Rule, bar *dailyBar, prevDate string, store *indicatorStore) bool {
	if len(rules) == 0 {
		return false
	}
	for _, rule := range rules {
		if !evaluateRule(rule, bar, prevDate, store) {
			return false
		}
	}
	return true
}

type backtestResult struct {
	FinalBalance  float64
	ROI           float64
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	MaxDrawdown   float64
	Trades        []strategyEntities.BacktestTrade
}

func runBacktestEngine(req *http.Request, strategy *strategyEntities.StrategyEntity, ticker, from, to string, initialBalance float64) (*backtestResult, error) {
	bars, err := fetchDailyBars(req, ticker, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bars: %w", err)
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("no price data for %s between %s and %s", ticker, from, to)
	}

	store, err := loadIndicators(req, strategy, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to load indicators: %w", err)
	}

	cash := initialBalance
	var shares float64
	var buyPrice float64
	var peakPrice float64
	var trades []strategyEntities.BacktestTrade

	peakEquity := initialBalance
	maxDrawdown := 0.0

	for i, bar := range bars {
		bar := bar // copy

		prevDate := ""
		if i > 0 {
			prevDate = bars[i-1].Date
		}

		equity := cash + shares*bar.Close
		if equity > peakEquity {
			peakEquity = equity
		}
		if peakEquity > 0 {
			dd := (peakEquity - equity) / peakEquity * 100
			if dd > maxDrawdown {
				maxDrawdown = dd
			}
		}

		if shares == 0 {
			// Check buy conditions (all must be met)
			if allBuyRulesMet(strategy.BuyRules, &bar, prevDate, store) {
				qty := math.Floor(cash / bar.Close)
				if qty >= 1 {
					cost := qty * bar.Close
					cash -= cost
					shares = qty
					buyPrice = bar.Close
					peakPrice = bar.Close
					trades = append(trades, strategyEntities.BacktestTrade{
						Type:      "BUY",
						Date:      bar.Date,
						Price:     bar.Close,
						Shares:    qty,
						Value:     cost,
						CashAfter: cash,
					})
				}
			}
		} else {
			// Update trailing stop peak
			if bar.Close > peakPrice {
				peakPrice = bar.Close
			}

			shouldSell := false

			// No sell conditions → hold to end
			if len(strategy.SellConditions) > 0 {
				for _, sc := range strategy.SellConditions {
					switch sc.Type {
					case strategyEntities.SellTakeProfit:
						pnlPct := (bar.Close - buyPrice) / buyPrice * 100
						if pnlPct >= sc.Percent {
							shouldSell = true
						}
					case strategyEntities.SellStopLoss:
						pnlPct := (bar.Close - buyPrice) / buyPrice * 100
						if pnlPct <= -sc.Percent {
							shouldSell = true
						}
					case strategyEntities.SellTrailingStop:
						if peakPrice > 0 {
							drawdown := (peakPrice - bar.Close) / peakPrice * 100
							if drawdown >= sc.Percent {
								shouldSell = true
							}
						}
					case strategyEntities.SellIndicator:
						if sc.Rule != nil && evaluateRule(*sc.Rule, &bar, prevDate, store) {
							shouldSell = true
						}
					}
					if shouldSell {
						break
					}
				}

				// Force close on last bar
				if i == len(bars)-1 {
					shouldSell = true
				}
			}

			if shouldSell {
				proceeds := shares * bar.Close
				pnl := (bar.Close - buyPrice) * shares
				pnlPct := (bar.Close - buyPrice) / buyPrice * 100
				cash += proceeds
				trades = append(trades, strategyEntities.BacktestTrade{
					Type:       "SELL",
					Date:       bar.Date,
					Price:      bar.Close,
					Shares:     shares,
					Value:      proceeds,
					PnL:        pnl,
					PnLPercent: pnlPct,
					CashAfter:  cash,
				})
				shares = 0
				buyPrice = 0
				peakPrice = 0
			}
		}
	}

	// Mark-to-market any open position
	lastPrice := bars[len(bars)-1].Close
	finalBalance := cash + shares*lastPrice
	roi := (finalBalance - initialBalance) / initialBalance * 100

	winning, losing := 0, 0
	for _, t := range trades {
		if t.Type == "SELL" {
			if t.PnL >= 0 {
				winning++
			} else {
				losing++
			}
		}
	}

	return &backtestResult{
		FinalBalance:  finalBalance,
		ROI:           roi,
		TotalTrades:   len(trades),
		WinningTrades: winning,
		LosingTrades:  losing,
		MaxDrawdown:   maxDrawdown,
		Trades:        trades,
	}, nil
}

// ── HTTP Handlers ─────────────────────────────────────────────────────────────

type runBacktestBody struct {
	StrategyUUID   string  `json:"strategy_uuid"`
	Ticker         string  `json:"ticker"`
	FromDate       string  `json:"from_date"`
	ToDate         string  `json:"to_date"`
	InitialBalance float64 `json:"initial_balance"`
}

func RunBacktest(res http.ResponseWriter, req *http.Request) {
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

	var body runBacktestBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("invalid request body", nil))
		return
	}

	body.Ticker = strings.ToUpper(strings.TrimSpace(body.Ticker))
	if body.StrategyUUID == "" || body.Ticker == "" || body.FromDate == "" || body.ToDate == "" || body.InitialBalance <= 0 {
		httpx.WriteError(res, req, httpx.BadRequest("strategy_uuid, ticker, from_date, to_date, initial_balance are required", nil))
		return
	}

	// Load strategy
	var strategy strategyEntities.StrategyEntity
	if err := db.Collection(datastores.Strategies).FindOne(req.Context(),
		bson.M{"uuid": body.StrategyUUID, "account_id": *account.AccountID}).Decode(&strategy); err != nil {
		httpx.WriteError(res, req, httpx.NotFound("strategy not found"))
		return
	}

	// Use the strategy ticker unless the caller overrides
	ticker := body.Ticker
	if ticker == "" {
		ticker = strategy.Ticker
	}

	result, err := runBacktestEngine(req, &strategy, ticker, body.FromDate, body.ToDate, body.InitialBalance)
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal(fmt.Sprintf("backtest failed: %s", err.Error())))
		return
	}

	record := strategyEntities.BacktestEntity{
		UUID:           uuid.New(),
		StrategyUUID:   body.StrategyUUID,
		Ticker:         ticker,
		FromDate:       body.FromDate,
		ToDate:         body.ToDate,
		InitialBalance: body.InitialBalance,
		FinalBalance:   result.FinalBalance,
		ROI:            result.ROI,
		TotalTrades:    result.TotalTrades,
		WinningTrades:  result.WinningTrades,
		LosingTrades:   result.LosingTrades,
		MaxDrawdown:    result.MaxDrawdown,
		Trades:         result.Trades,
		AccountID:      *account.AccountID,
		PortfolioUUID:  strategy.PortfolioUUID,
		CreatedAt:      time.Now().UTC(),
	}

	if _, err := db.Collection(datastores.Backtests).InsertOne(req.Context(), record); err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to save backtest"))
		return
	}

	httpx.WriteJSON(res, http.StatusOK, record)
}

func GetBacktests(res http.ResponseWriter, req *http.Request) {
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

	strategyUUID := strings.TrimSpace(req.URL.Query().Get("strategy_uuid"))
	if strategyUUID == "" {
		httpx.WriteError(res, req, httpx.BadRequest("strategy_uuid is required", nil))
		return
	}

	cur, err := db.Collection(datastores.Backtests).Find(req.Context(),
		bson.M{"strategy_uuid": strategyUUID, "account_id": *account.AccountID})
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to fetch backtests"))
		return
	}
	defer cur.Close(req.Context())

	var backtests []strategyEntities.BacktestEntity
	if err := cur.All(req.Context(), &backtests); err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to decode backtests"))
		return
	}
	if backtests == nil {
		backtests = []strategyEntities.BacktestEntity{}
	}

	httpx.WriteJSON(res, http.StatusOK, map[string]any{"backtests": backtests})
}
