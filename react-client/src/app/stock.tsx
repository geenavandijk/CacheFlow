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
import { usePortfolio } from "./portfolio_context";
import { usePositions } from "./positions_context";

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

  const { selectedPortfolio } = usePortfolio();
  const { positionsByTicker, refreshPositions } = usePositions();

  const [side, setSide] = useState<"BUY" | "SELL">("BUY");
  const [quantity, setQuantity] = useState("");
  const [placing, setPlacing] = useState(false);
  const [orderError, setOrderError] = useState<string | null>(null);
  const [showConfetti, setShowConfetti] = useState(false);
  const [orders, setOrders] = useState<any[]>([]);
  const [ordersPage, setOrdersPage] = useState(1);
  const [ordersHasMore, setOrdersHasMore] = useState(false);
  const [ordersLoading, setOrdersLoading] = useState(false);

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

  const results = overview?.results;
  const position = ticker ? positionsByTicker[ticker] : undefined;
  
  const realtimePrice = position?.current_price ?? null;
  const displayData = useMemo(() => {
    if (!chartData.length || realtimePrice == null) return chartData;
    const last = { ...chartData[chartData.length - 1], close: realtimePrice };
    return [...chartData.slice(0, -1), last];
  }, [chartData, realtimePrice]);

  const latestClose = displayData.length > 0 ? displayData[displayData.length - 1].close : null;
  const firstClose = displayData.length > 0 ? displayData[0].close : null;

  useEffect(() => {
    setOrdersPage(1);
  }, [selectedPortfolio?.uuid, ticker]);

  useEffect(() => {
    const loadOrders = async () => {
      if (!selectedPortfolio || !ticker) {
        setOrders([]);
        setOrdersHasMore(false);
        return;
      }
      setOrdersLoading(true);
      try {
        const res = await api.get("/v1/orders", {
          params: {
            portfolio_uuid: selectedPortfolio.uuid,
            ticker,
            page: ordersPage,
            limit: 10,
          },
        });
        setOrders(res.data.orders ?? []);
        setOrdersHasMore(Boolean(res.data.has_more));
      } catch (err) {
        console.error("Failed to load orders for ticker", err);
        setOrders([]);
        setOrdersHasMore(false);
      } finally {
        setOrdersLoading(false);
      }
    };
    void loadOrders();
  }, [selectedPortfolio?.uuid, ticker, ordersPage]);
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

  return (
    <div className="flex flex-col max-w-6xl w-full mx-auto h-full bg-black text-white px-8 py-8 overflow-auto">
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
            {loadingAgg && displayData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-neutral-500">Loading chart…</div>
            ) : displayData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-neutral-500">No data for this range</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={displayData} margin={{ top: 8, right: 8, left: 8, bottom: 8 }}>
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

          <div className="mt-6 grid gap-6 md:grid-cols-2">
            <div className="rounded-xl border border-neutral-800 bg-neutral-900/40 px-5 py-4">
              <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500 mb-3">
                Position
              </h2>
              {selectedPortfolio ? (
                position ? (
                  <div className="space-y-1 text-sm">
                    <p>
                      <span className="text-neutral-500 mr-2">Shares:</span>
                      {position.shares}
                    </p>
                    <p>
                      <span className="text-neutral-500 mr-2">Avg cost:</span>
                      ${position.avg_cost.toFixed(2)}
                    </p>
                    <p>
                      <span className="text-neutral-500 mr-2">Unrealized:</span>
                      <span
                        className={
                          position.unrealized >= 0 ? "text-green-400" : "text-red-400"
                        }
                      >
                        ${position.unrealized.toFixed(2)}
                      </span>
                    </p>
                  </div>
                ) : (
                  <p className="text-sm text-neutral-500">
                    No active position in this ticker for the selected portfolio.
                  </p>
                )
              ) : (
                <p className="text-sm text-neutral-500">
                  Select a portfolio in the top nav to see your position.
                </p>
              )}
            </div>

            <div className="rounded-xl border border-neutral-800 bg-neutral-900/40 px-5 py-4">
              <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500 mb-3">
                New order
              </h2>
              {selectedPortfolio ? (
                <form
                  className="space-y-4"
                  onSubmit={async (e) => {
                    e.preventDefault();
                    setOrderError(null);
                    if (!quantity.trim()) {
                      setOrderError("Quantity is required.");
                      return;
                    }
                    const qty = parseInt(quantity, 10);
                    if (Number.isNaN(qty) || qty <= 0) {
                      setOrderError("Quantity must be a positive integer.");
                      return;
                    }
                    setPlacing(true);
                    try {
                      await api.post("/v1/order", {
                        ticker,
                        side,
                        quantity: qty,
                        portfolio_uuid: selectedPortfolio.uuid,
                      });
                      setQuantity("");
                      await refreshPositions();
                      setShowConfetti(true);
                      setTimeout(() => setShowConfetti(false), 1600);
                    } catch (err: any) {
                      const msg =
                        err?.response?.data?.error ??
                        err?.response?.data?.message ??
                        "Failed to place order.";
                      setOrderError(String(msg));
                    } finally {
                      setPlacing(false);
                    }
                  }}
                >
                  <div className="inline-flex rounded-full border border-neutral-700 bg-neutral-950/60 p-0.5 text-xs">
                    <button
                      type="button"
                      onClick={() => setSide("BUY")}
                      className={`px-4 py-1.5 rounded-full transition-colors ${
                        side === "BUY"
                          ? "bg-green-500 text-black"
                          : "text-neutral-300 hover:text-white"
                      }`}
                    >
                      Buy
                    </button>
                    <button
                      type="button"
                      onClick={() => setSide("SELL")}
                      className={`px-4 py-1.5 rounded-full transition-colors ${
                        side === "SELL"
                          ? "bg-red-500 text-black"
                          : "text-neutral-300 hover:text-white"
                      }`}
                    >
                      Sell
                    </button>
                  </div>

                  <div>
                    <label className="text-white text-sm font-regular">
                      Quantity
                    </label>
                    <input
                      type="number"
                      min="1"
                      step="1"
                      value={quantity}
                      onChange={(e) => setQuantity(e.target.value)}
                      className="w-full h-9 text-white mt-1 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm"
                    />
                  </div>

                  {(() => {
                    const qty = parseInt(quantity || "0", 10);
                    const validQty = !Number.isNaN(qty) && qty > 0;
                    const basePrice =
                      side === "SELL" && position
                        ? position.current_price ?? latestClose
                        : latestClose ?? position?.current_price ?? null;
                    const total =
                      validQty && basePrice != null ? qty * basePrice : null;
                    if (!total) return null;
                    return (
                      <p className="text-xs text-neutral-400">
                        {side === "BUY"
                          ? `$${total.toFixed(2)} cost`
                          : `$${total.toFixed(2)} returning`}
                      </p>
                    );
                  })()}

                  {orderError && (
                    <p className="text-xs text-red-400">{orderError}</p>
                  )}

                  <button
                    type="submit"
                    disabled={placing}
                    className="mt-1 inline-flex items-center justify-center rounded-full bg-orange-500 text-white text-xs font-medium px-6 py-2.5 hover:bg-orange-600 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    {placing && (
                      <span className="w-3 h-3 mr-2 border-2 border-orange-200 border-t-transparent rounded-full animate-spin" />
                    )}
                    {placing ? "Placing order…" : "Submit order"}
                  </button>
                </form>
              ) : (
                <p className="text-sm text-neutral-500">
                  Select a portfolio in the top nav to place an order.
                </p>
              )}
            </div>
          </div>

          {results.description && (
            <div className="mt-6 max-w-3xl">
              <h2 className="text-xs font-medium uppercase tracking-wider text-neutral-500 mb-2">
                About
              </h2>
              <p className="text-sm text-neutral-400 leading-relaxed">
                {results.description}
              </p>
            </div>
          )}

          {selectedPortfolio && (
            <div className="mt-8 max-w-3xl">
              <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500 mb-2">
                Order history
              </h2>
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-4 py-3">
                {ordersLoading ? (
                  <p className="text-sm text-neutral-500">Loading orders…</p>
                ) : orders.length === 0 ? (
                  <p className="text-sm text-neutral-500">
                    No orders yet for this ticker in the selected portfolio.
                  </p>
                ) : (
                  <div className="space-y-1 text-xs">
                    {orders.map((o) => {
                      const side = o.side ?? o.Side;
                      const qty = o.quantity ?? o.Quantity ?? 0;
                      const price = o.price ?? o.Price ?? 0;
                      const realized =
                        typeof o.realized === "number"
                          ? o.realized
                          : typeof o.Realized === "number"
                          ? o.Realized
                          : null;
                      const ts =
                        o.timestamp ??
                        o.Timestamp ??
                        o.created_at ??
                        o.CreatedAt ??
                        null;
                      const dt = ts ? new Date(ts) : null;
                      const dateStr = dt
                        ? dt.toLocaleString([], {
                            month: "short",
                            day: "2-digit",
                            hour: "2-digit",
                            minute: "2-digit",
                          })
                        : "";
                      const realizedPrefix =
                        realized != null && realized !== 0
                          ? realized > 0
                            ? "+"
                            : "-"
                          : "";
                      const realizedAbs =
                        realized != null
                          ? Math.abs(realized).toFixed(2)
                          : null;
                      const realizedText =
                        realizedAbs == null
                          ? "—"
                          : realizedPrefix === ""
                          ? `$${realizedAbs}`
                          : `${realizedPrefix}$${realizedAbs}`;
                      const realizedClass =
                        realized != null && realized !== 0
                          ? realized > 0
                            ? "text-green-400"
                            : "text-red-400"
                          : "text-neutral-400";

                      return (
                        <div
                          key={o.uuid ?? o.UUID ?? `${ticker}-${ts}-${side}`}
                          className="flex items-center justify-between py-1 border-b border-neutral-900 last:border-b-0"
                        >
                          <div className="flex flex-col">
                            <div className="flex items-center gap-2">
                              <span
                                className={
                                  side === "BUY"
                                    ? "text-[11px] font-semibold text-green-400"
                                    : "text-[11px] font-semibold text-red-400"
                                }
                              >
                                {side}
                              </span>
                              <span className="text-[11px] text-neutral-500">
                                {dateStr}
                              </span>
                            </div>
                            <span className="text-[11px] text-neutral-500">
                              {qty} @ ${Number(price).toFixed(2)}
                            </span>
                          </div>
                          <div className="flex flex-col items-end">
                            <span className={realizedClass}>{realizedText}</span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
                <div className="mt-3 flex justify-between items-center text-[11px] text-neutral-500">
                  <button
                    type="button"
                    disabled={ordersPage <= 1 || ordersLoading}
                    onClick={() => setOrdersPage((p) => Math.max(1, p - 1))}
                    className={`px-3 py-1 rounded-full border border-neutral-800 ${
                      ordersPage <= 1 || ordersLoading
                        ? "opacity-40 cursor-not-allowed"
                        : "hover:border-neutral-500 hover:text-neutral-200"
                    }`}
                  >
                    Previous
                  </button>
                  <span>Page {ordersPage}</span>
                  <button
                    type="button"
                    disabled={!ordersHasMore || ordersLoading}
                    onClick={() => setOrdersPage((p) => p + 1)}
                    className={`px-3 py-1 rounded-full border border-neutral-800 ${
                      !ordersHasMore || ordersLoading
                        ? "opacity-40 cursor-not-allowed"
                        : "hover:border-neutral-500 hover:text-neutral-200"
                    }`}
                  >
                    Next
                  </button>
                </div>
              </div>
            </div>
          )}
        </>
      )}
      {showConfetti && (
        <div className="pointer-events-none fixed inset-0 z-50 flex items-start justify-center">
          <div className="relative w-full max-w-xl h-40 mt-12">
            {Array.from({ length: 60 }).map((_, i) => (
              <span
                key={i}
                className="absolute w-1.5 h-3 rounded-sm"
                style={{
                  left: `${Math.random() * 100}%`,
                  backgroundColor:
                    ["#f97316", "#22c55e", "#38bdf8", "#e11d48"][
                      i % 4
                    ],
                  opacity: 0.9,
                  transform: `translateY(0px)`,
                  animation: `cf-fall 700ms ease-out forwards`,
                  animationDelay: `${Math.random() * 400}ms`,
                }}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
