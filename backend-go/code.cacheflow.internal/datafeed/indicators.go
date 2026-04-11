package datafeed

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/secrets"

	"encoding/json"
)

// proxyIndicator is a generic proxy for RSI / EMA / SMA indicator endpoints.
// Route: GET /v1/datafeed/indicators/{kind}?ticker=X&timespan=day&window=14&from=YYYY-MM-DD&to=YYYY-MM-DD
func proxyIndicator(kind string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		ticker := strings.ToUpper(strings.TrimSpace(q.Get("ticker")))
		if ticker == "" {
			httpx.WriteError(res, req, httpx.BadRequest("ticker is required", nil))
			return
		}

		timespan := q.Get("timespan")
		if timespan == "" {
			timespan = "day"
		}
		window := q.Get("window")
		if window == "" {
			window = "14"
		}
		from := q.Get("from")
		to := q.Get("to")
		order := q.Get("order")
		if order == "" {
			order = "asc"
		}
		limit := q.Get("limit")
		if limit == "" {
			limit = "5000"
		}

		u := fmt.Sprintf("%s/v1/indicators/%s/%s?timespan=%s&window=%s&order=%s&limit=%s&adjusted=true&apiKey=%s",
			massiveBaseURL, kind, ticker, timespan, window, order, limit, secrets.MassiveMainApiKeyValue)
		if from != "" {
			u += "&timestamp.gte=" + from
		}
		if to != "" {
			u += "&timestamp.lte=" + to
		}

		httpReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, u, nil)
		if err != nil {
			httpx.WriteError(res, req, httpx.Internal("failed to create request"))
			return
		}

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			httpx.WriteError(res, req, httpx.Internal("failed to fetch indicator"))
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			httpx.WriteError(res, req, httpx.Internal("failed to read response"))
			return
		}

		if resp.StatusCode != http.StatusOK {
			httpx.WriteError(res, req, httpx.NotFound("indicator data not found"))
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			httpx.WriteError(res, req, httpx.Internal("invalid response"))
			return
		}

		util.JSONResponse(res, http.StatusOK, result)
	}
}

// GetRSI proxies RSI data from Massive.
func GetRSI(res http.ResponseWriter, req *http.Request) {
	proxyIndicator("rsi")(res, req)
}

// GetEMA proxies EMA data from Massive.
func GetEMA(res http.ResponseWriter, req *http.Request) {
	proxyIndicator("ema")(res, req)
}

// GetSMA proxies SMA data from Massive.
func GetSMA(res http.ResponseWriter, req *http.Request) {
	proxyIndicator("sma")(res, req)
}

// GetMACD proxies MACD data from Massive.
// Route: GET /v1/datafeed/indicators/macd?ticker=X&timespan=day&fast_period=12&slow_period=26&signal_period=9&from=...&to=...
func GetMACD(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	ticker := strings.ToUpper(strings.TrimSpace(q.Get("ticker")))
	if ticker == "" {
		httpx.WriteError(res, req, httpx.BadRequest("ticker is required", nil))
		return
	}

	timespan := q.Get("timespan")
	if timespan == "" {
		timespan = "day"
	}
	fastPeriod := q.Get("fast_period")
	if fastPeriod == "" {
		fastPeriod = "12"
	}
	slowPeriod := q.Get("slow_period")
	if slowPeriod == "" {
		slowPeriod = "26"
	}
	signalPeriod := q.Get("signal_period")
	if signalPeriod == "" {
		signalPeriod = "9"
	}
	from := q.Get("from")
	to := q.Get("to")
	order := q.Get("order")
	if order == "" {
		order = "asc"
	}

	u := fmt.Sprintf(
		"%s/v1/indicators/macd/%s?timespan=%s&fast_period=%s&slow_period=%s&signal_period=%s&order=%s&limit=5000&adjusted=true&apiKey=%s",
		massiveBaseURL, ticker, timespan, fastPeriod, slowPeriod, signalPeriod, order, secrets.MassiveMainApiKeyValue)
	if from != "" {
		u += "&timestamp.gte=" + from
	}
	if to != "" {
		u += "&timestamp.lte=" + to
	}

	httpReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, u, nil)
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to create request"))
		return
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to fetch MACD"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpx.WriteError(res, req, httpx.Internal("failed to read response"))
		return
	}

	if resp.StatusCode != http.StatusOK {
		httpx.WriteError(res, req, httpx.NotFound("MACD data not found"))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		httpx.WriteError(res, req, httpx.Internal("invalid response"))
		return
	}

	util.JSONResponse(res, http.StatusOK, result)
}
