import { useState, useEffect, useMemo } from "react";
import { useSearchParams, Link } from "react-router-dom";
import api from "../util/api";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

const TIMEFRAMES = ["1d", "1w", "1m", "3m", "1y", "5y", "all"] as const;
type Timeframe = (typeof TIMEFRAMES)[number];

interface TickerResults {
  active?: boolean;
  name?: string;
  ticker?: string;
  description?: string;
  market_cap?: number;
  primary_exchange?: string;
  locale?: string;
  currency_name?: string;
  homepage_url?: string;
  list_date?: string;
  branding?: { logo_url?: string; icon_url?: string };
  address?: { city?: string; state?: string };
}

interface AggregatesResult {
  results?: Array<{ t: number; o: number; h: number; l: number; c: number; v?: number }>;
  ticker?: string;
  resultsCount?: number;
}

function formatPrice(n: number): string {
  if (n >= 1e12) return `$${(n / 1e12).toFixed(2)}T`;
  if (n >= 1e9) return `$${(n / 1e9).toFixed(2)}B`;
  if (n >= 1e6) return `$${(n / 1e6).toFixed(2)}M`;
  if (n >= 1e3) return `$${(n / 1e3).toFixed(2)}K`;
  return `$${n.toFixed(2)}`;
}

function formatDate(ms: number, timeframe: Timeframe): string {
  const d = new Date(ms);
  if (timeframe === "1d") return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  return d.toLocaleDateString([], { month: "short", day: "numeric", year: timeframe === "1y" || timeframe === "5y" || timeframe === "all" ? "2-digit" : undefined });
}

export const Stock = () => {
  const [searchParams] = useSearchParams();
  const ticker = searchParams.get("ticker")?.toUpperCase() ?? "";

  const [overview, setOverview] = useState<{ results?: TickerResults } | null>(null);
  const [aggregates, setAggregates] = useState<AggregatesResult | null>(null);
  const [timeframe, setTimeframe] = useState<Timeframe>("1m");
  const [loadingOverview, setLoadingOverview] = useState(true);
  const [loadingAgg, setLoadingAgg] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!ticker) {
      setLoadingOverview(false);
      setLoadingAgg(false);
      return;
    }
    setError(null);
    setLoadingOverview(true);
    api
      .get("/v1/datafeed/stock", { params: { ticker } })
      .then((res) => setOverview(res.data))
      .catch((err) => setError(err.response?.data?.error ?? "Failed to load ticker"))
      .finally(() => setLoadingOverview(false));
  }, [ticker]);

  useEffect(() => {
    document.title = ticker ? `CacheFlow | ${ticker}` : "CacheFlow | Stock";
  }, [ticker]);

  useEffect(() => {
    if (!ticker) return;
    setLoadingAgg(true);
    api
      .get("/v1/datafeed/stock/aggregates", { params: { ticker, timeframe } })
      .then((res) => setAggregates(res.data))
      .catch(() => setAggregates({ results: [] }))
      .finally(() => setLoadingAgg(false));
  }, [ticker, timeframe]);

  const chartData = useMemo(() => {
    const results = aggregates?.results ?? [];
    return results.map((r) => ({
      ...r,
      time: r.t,
      dateLabel: formatDate(r.t, timeframe),
      close: r.c,
      volume: r.v ?? 0,
    }));
  }, [aggregates, timeframe]);

  const latestClose = chartData.length > 0 ? chartData[chartData.length - 1].close : null;
  const firstClose = chartData.length > 0 ? chartData[0].close : null;
  const pctChange = latestClose != null && firstClose != null && firstClose !== 0
    ? (((latestClose - firstClose) / firstClose) * 100).toFixed(2)
    : null;

  if (!ticker) {
    return (
      <div className="flex flex-col w-full h-full bg-black text-white px-8 py-8">
        <h1 className="text-2xl font-semibold mb-2">Stock</h1>
        <p className="text-neutral-400">No ticker specified. Use <code className="text-orange-500">/client-app/stock?ticker=AAPL</code></p>
        <Link to="/client-app/dashboard" className="mt-4 text-orange-500 hover:underline">Back to Dashboard</Link>
      </div>
    );
  }

  const results = overview?.results;

  return (
    <div className="flex flex-col w-full h-full bg-black text-white px-8 py-8 overflow-auto">
      <div className="flex items-center gap-3 mb-4">
        <Link to="/client-app/dashboard" className="text-neutral-400 hover:text-white text-sm transition-colors">← Dashboard</Link>
        <span className="text-neutral-600">|</span>
        <h1 className="text-2xl font-semibold">
          {results?.name ?? ticker}
          {results?.ticker && <span className="text-neutral-500 font-normal ml-2">({results.ticker})</span>}
        </h1>
      </div>

      {error && (
        <p className="text-red-400 text-sm mb-4">{error}</p>
      )}

      {loadingOverview && !results && (
        <p className="text-neutral-400">Loading…</p>
      )}

      {results && (
        <>
          <div className="flex flex-wrap items-baseline gap-4 mb-2">
            {latestClose != null && (
              <span className="text-2xl font-semibold">${latestClose.toFixed(2)}</span>
            )}
            {pctChange != null && (
              <span className={Number(pctChange) >= 0 ? "text-green-400" : "text-red-400"}>
                {Number(pctChange) >= 0 ? "+" : ""}{pctChange}%
              </span>
            )}
          </div>
          <div className="flex flex-wrap gap-6 text-sm text-neutral-400 mb-6">
            {results.market_cap != null && <span>Market cap {formatPrice(results.market_cap)}</span>}
            {results.primary_exchange && <span>{results.primary_exchange}</span>}
            {results.currency_name && <span>{results.currency_name.toUpperCase()}</span>}
          </div>

          <div className="flex gap-2 mb-4">
            {TIMEFRAMES.map((tf) => (
              <button
                key={tf}
                onClick={() => setTimeframe(tf)}
                className={`rounded-full px-3 py-1.5 text-sm font-medium transition-all duration-200 ${
                  timeframe === tf
                    ? "bg-orange-500 text-black"
                    : "border border-neutral-700 text-neutral-300 hover:border-neutral-500 hover:text-white"
                }`}
              >
                {tf.toUpperCase()}
              </button>
            ))}
          </div>

          <div className="flex-1 min-h-[320px] w-full rounded-lg border border-neutral-800 bg-neutral-900/30 p-4">
            {loadingAgg && chartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-neutral-500">Loading chart…</div>
            ) : chartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-neutral-500">No data for this range</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData} margin={{ top: 8, right: 8, left: 8, bottom: 8 }}>
                  <defs>
                    <linearGradient id="fillPrice" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="rgb(249 115 22)" stopOpacity={0.3} />
                      <stop offset="100%" stopColor="rgb(249 115 22)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis
                    dataKey="dateLabel"
                    stroke="#71717a"
                    tick={{ fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    domain={["auto", "auto"]}
                    stroke="#71717a"
                    tick={{ fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={(v) => `$${v}`}
                  />
                  <Tooltip
                    contentStyle={{ backgroundColor: "#18181b", border: "1px solid #27272a", borderRadius: "8px" }}
                    labelStyle={{ color: "#a1a1aa" }}
                    formatter={(value: number | undefined) => [value != null ? `$${Number(value).toFixed(2)}` : "—", "Close"]}
                    labelFormatter={(_, payload) => payload[0]?.payload?.dateLabel ?? ""}
                  />
                  <Area
                    type="monotone"
                    dataKey="close"
                    stroke="rgb(249 115 22)"
                    strokeWidth={2}
                    fill="url(#fillPrice)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>

          {results.description && (
            <div className="mt-6 max-w-3xl">
              <h2 className="text-xs font-medium uppercase tracking-wider text-neutral-500 mb-2">About</h2>
              <p className="text-sm text-neutral-400 leading-relaxed">{results.description}</p>
            </div>
          )}
        </>
      )}
    </div>
  );
};
