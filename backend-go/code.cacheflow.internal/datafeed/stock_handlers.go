package datafeed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/secrets"

	"github.com/charmbracelet/log"
)

const massiveBaseURL = "https://api.massive.com"

func GetTickerOverview(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: "DATAFEED (TO)",
	})
	ticker := req.URL.Query().Get("ticker")
	if ticker == "" {
		httpx.WriteError(res, req, httpx.BadRequest("ticker is required", map[string]string{"ticker": "missing"}))
		return
	}

	url := fmt.Sprintf("%s/v3/reference/tickers/%s?apiKey=%s", massiveBaseURL, ticker, secrets.MassiveMainApiKeyValue)
	httpReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, url, nil)
	if err != nil {
		logger.Error("failed to create request", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to request ticker details"))
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error("massive api request failed", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to fetch ticker details"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to read ticker response"))
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("massive api error", "status", resp.StatusCode, "body", string(body))
		httpx.WriteError(res, req, httpx.NotFound("Ticker not found or API error"))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("failed to parse response", "err", err)
		httpx.WriteError(res, req, httpx.Internal("invalid ticker response"))
		return
	}

	util.JSONResponse(res, http.StatusOK, result)
}

// timeframeToRange returns multiplier, timespan, from, to for Massive aggregates API.
func timeframeToRange(tf string) (multiplier, timespan, from, to string, ok bool) {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	switch tf {
	case "1d":
		// 5-minute candles for the most recent trading day (Friday if weekend)
		tradeDay := now
		switch tradeDay.Weekday() {
		case time.Saturday:
			tradeDay = tradeDay.AddDate(0, 0, -1)
		case time.Sunday:
			tradeDay = tradeDay.AddDate(0, 0, -2)
		}
		day := tradeDay.Format("2006-01-02")
		return "5", "minute", day, day, true
	case "1w":
		// 1-hour candles
		start := now.AddDate(0, 0, -7).Format("2006-01-02")
		return "1", "hour", start, today, true
	case "1m":
		// 4-hour candles
		start := now.AddDate(0, -1, 0).Format("2006-01-02")
		return "4", "hour", start, today, true
	case "3m":
		// 1-day candles
		start := now.AddDate(0, -3, 0).Format("2006-01-02")
		return "1", "day", start, today, true
	case "1y":
		// 1-day candles
		start := now.AddDate(-1, 0, 0).Format("2006-01-02")
		return "1", "day", start, today, true
	case "5y":
		// 1-week (7-day) candles
		start := now.AddDate(-5, 0, 0).Format("2006-01-02")
		return "1", "week", start, today, true
	case "all":
		// 1-month (30-day) candles
		return "1", "month", "1990-01-01", today, true
	default:
		return "", "", "", "", false
	}
}

func GetTickerAggregatesWithTimeframe(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: "DATAFEED (AG)",
	})
	ticker := req.URL.Query().Get("ticker")
	timeframe := req.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "1m"
	}
	mult, span, from, to, ok := timeframeToRange(timeframe)
	if !ok {
		httpx.WriteError(res, req, httpx.BadRequest("invalid timeframe; use 1d, 1w, 1m, 3m, 1y, 5y, all", nil))
		return
	}
	proxyAggregates(res, req, logger, ticker, mult, span, from, to)
}

func GetTickerAggregates(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: "DATAFEED (AG)",
	})
	ticker := req.URL.Query().Get("ticker")
	from := req.URL.Query().Get("from")
	to := req.URL.Query().Get("to")
	multiplier := req.URL.Query().Get("multiplier")
	timespan := req.URL.Query().Get("timespan")

	if ticker == "" || from == "" || to == "" || multiplier == "" || timespan == "" {
		httpx.WriteError(res, req, httpx.BadRequest("ticker, from, to, multiplier, timespan are required", nil))
		return
	}
	proxyAggregates(res, req, logger, ticker, multiplier, timespan, from, to)
}

func proxyAggregates(res http.ResponseWriter, req *http.Request, logger *log.Logger, ticker, multiplier, timespan, from, to string) {
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/%s/%s/%s/%s?apiKey=%s&sort=asc&limit=5000",
		massiveBaseURL, ticker, multiplier, timespan, from, to, secrets.MassiveMainApiKeyValue)

	httpReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, url, nil)
	if err != nil {
		logger.Error("failed to create request", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to request aggregates"))
		return
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error("massive api request failed", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to fetch aggregates"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response", "err", err)
		httpx.WriteError(res, req, httpx.Internal("failed to read aggregates response"))
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("massive api error", "status", resp.StatusCode, "body", string(body))
		httpx.WriteError(res, req, httpx.NotFound("Aggregates not found or API error"))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("failed to parse response", "err", err)
		httpx.WriteError(res, req, httpx.Internal("invalid aggregates response"))
		return
	}

	util.JSONResponse(res, http.StatusOK, result)
}

// GetTickerSnapshots returns latest price and intraday change for one or more tickers.
// Query params:
// - ticker: single ticker, OR
// - tickers: comma-separated list of tickers
func GetTickerSnapshots(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: "DATAFEED (SN)",
	})

	// Support either ?ticker= or ?tickers=aapl,msft
	q := req.URL.Query()
	single := strings.TrimSpace(q.Get("ticker"))
	multi := strings.TrimSpace(q.Get("tickers"))

	var tickers []string
	if multi != "" {
		for _, t := range strings.Split(multi, ",") {
			tt := strings.ToUpper(strings.TrimSpace(t))
			if tt != "" {
				tickers = append(tickers, tt)
			}
		}
	} else if single != "" {
		tickers = []string{strings.ToUpper(single)}
	}

	if len(tickers) == 0 {
		httpx.WriteError(res, req, httpx.BadRequest("ticker or tickers is required", map[string]string{
			"ticker":  "missing",
			"tickers": "missing",
		}))
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}

	type Snapshot struct {
		Ticker         string   `json:"ticker"`
		LastPrice      float64  `json:"last_price"`
		PrevClose      *float64 `json:"prev_close,omitempty"`
		Change         *float64 `json:"change,omitempty"`
		ChangePercent  *float64 `json:"change_percent,omitempty"`
		Raw            any      `json:"raw,omitempty"`
	}

	var snapshots []Snapshot

	for _, t := range tickers {
		url := fmt.Sprintf("%s/v2/snapshot/locale/us/markets/stocks/tickers/%s?apiKey=%s",
			massiveBaseURL, t, secrets.MassiveMainApiKeyValue)

		httpReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, url, nil)
		if err != nil {
			logger.Error("failed to create snapshot request", "ticker", t, "err", err)
			continue
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			logger.Error("snapshot request failed", "ticker", t, "err", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Error("failed to read snapshot response", "ticker", t, "err", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			logger.Error("snapshot api error", "ticker", t, "status", resp.StatusCode, "body", string(body))
			continue
		}

		var parsed struct {
			Status string                 `json:"status"`
			Ticker map[string]any         `json:"ticker"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			logger.Error("failed to parse snapshot response", "ticker", t, "err", err)
			continue
		}

		var lastPrice float64
		var prevClosePtr *float64
		var changePtr *float64
		var changePctPtr *float64

		if lastTradeRaw, ok := parsed.Ticker["lastTrade"]; ok {
			if m, ok := lastTradeRaw.(map[string]any); ok {
				if p, ok := m["p"].(float64); ok {
					lastPrice = p
				}
			}
		}
		if lastPrice == 0 {
			if dayRaw, ok := parsed.Ticker["day"]; ok {
				if m, ok := dayRaw.(map[string]any); ok {
					if c, ok := m["c"].(float64); ok {
						lastPrice = c
					}
				}
			}
		}

		if prevDayRaw, ok := parsed.Ticker["prevDay"]; ok {
			if m, ok := prevDayRaw.(map[string]any); ok {
				if c, ok := m["c"].(float64); ok {
					prevClose := c
					prevClosePtr = &prevClose
					if lastPrice != 0 {
						ch := lastPrice - c
						cp := (ch / c) * 100
						changePtr = &ch
						changePctPtr = &cp
					}
				}
			}
		}

		snapshots = append(snapshots, Snapshot{
			Ticker:        t,
			LastPrice:     lastPrice,
			PrevClose:     prevClosePtr,
			Change:        changePtr,
			ChangePercent: changePctPtr,
		})
	}

	util.JSONResponse(res, http.StatusOK, map[string]any{
		"snapshots": snapshots,
	})
}
