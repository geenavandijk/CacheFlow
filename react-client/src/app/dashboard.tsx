import { useAuth } from "../provider/auth";
import { usePortfolio } from "./portfolio_context";
import { usePositions } from "./positions_context";
import { useWatchlists } from "./watchlists_context";
import { useWatchlistMarketData } from "./use_watchlist_market_data";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../util/api";
import { Hourglass, Star, GripVertical } from "lucide-react";
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  YAxis,
} from "recharts";

const TIMEFRAMES: Array<"1d" | "1w" | "1m" | "3m" | "1y" | "all"> = [
  "1d",
  "1w",
  "1m",
  "3m",
  "1y",
  "all",
];

export const Dashboard = () => {
  const { accountData } = useAuth();
  const { portfolios, selectedPortfolio, createPortfolio } = usePortfolio();
  const { positions, totalMarketValue, activePositionsCount } = usePositions();
  const {
    watchlists,
    loading: watchlistsLoading,
    createWatchlist,
    updateWatchlist,
    deleteWatchlist,
    reorderWatchlists,
  } = useWatchlists();
  const navigate = useNavigate();

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [startingBalance, setStartingBalance] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [savingCreate, setSavingCreate] = useState(false);

  const [timeframe, setTimeframe] =
    useState<"1d" | "1w" | "1m" | "3m" | "1y" | "all">("1m");
  const [buyingPower, setBuyingPower] = useState<number | null>(null);
  const [buyingPowerLoading, setBuyingPowerLoading] = useState(false);

  const [nameByTicker, setNameByTicker] = useState<Record<string, string>>({});
  const [orders, setOrders] = useState<any[]>([]);
  const [ordersPage, setOrdersPage] = useState(1);
  const [ordersHasMore, setOrdersHasMore] = useState(false);
  const [ordersLoading, setOrdersLoading] = useState(false);

  const [isWatchlistCreateOpen, setIsWatchlistCreateOpen] = useState(false);
  const [watchlistName, setWatchlistName] = useState("");
  const [watchlistError, setWatchlistError] = useState<string | null>(null);
  const [savingWatchlist, setSavingWatchlist] = useState(false);

  const [isReorderOpen, setIsReorderOpen] = useState(false);
  const [reorderItems, setReorderItems] = useState<string[]>([]);

  const [activeWatchlistForAdd, setActiveWatchlistForAdd] = useState<string | null>(null);
  const [addSearchQuery, setAddSearchQuery] = useState("");
  const [addSearchResults, setAddSearchResults] = useState<{ name: string; ticker: string }[]>([]);
  const [addSearchLoading, setAddSearchLoading] = useState(false);
  const [addSearchError, setAddSearchError] = useState<string | null>(null);

  const [activeWatchlistForManage, setActiveWatchlistForManage] = useState<string | null>(null);

  useEffect(() => {
    const loadBuyingPower = async () => {
      if (!selectedPortfolio) {
        setBuyingPower(null);
        return;
      }
      setBuyingPowerLoading(true);
      try {
        const res = await api.get("/v1/portfolio/buying-power", {
          params: { portfolio_uuid: selectedPortfolio.uuid },
        });
        setBuyingPower(res.data.buying_power ?? null);
      } catch (err) {
        console.error("Failed to load buying power", err);
        setBuyingPower(null);
      } finally {
        setBuyingPowerLoading(false);
      }
    };
    loadBuyingPower();
  }, [selectedPortfolio]);

  useEffect(() => {
    if (!activeWatchlistForAdd) return;
    if (!addSearchQuery) {
      setAddSearchResults([]);
      return;
    }
    let cancelled = false;
    setAddSearchLoading(true);
    setAddSearchError(null);
    const timeout = window.setTimeout(async () => {
      try {
        const res = await fetch(
          `/v1/companies/search?q=${encodeURIComponent(addSearchQuery)}`,
        );
        if (cancelled) return;
        const data = (await res.json()) as { name: string; ticker: string }[];
        setAddSearchResults(data);
      } catch (err) {
        console.error("Failed to search companies for watchlist add", err);
        if (!cancelled) {
          setAddSearchError("Failed to search tickers.");
        }
      } finally {
        if (!cancelled) {
          setAddSearchLoading(false);
        }
      }
    }, 300);
    return () => {
      cancelled = true;
      window.clearTimeout(timeout);
    };
  }, [activeWatchlistForAdd, addSearchQuery]);

  const handleOpenCreate = () => {
    setName("");
    setDescription("");
    setStartingBalance("");
    setCreateError(null);
    setIsCreateOpen(true);
  };

  const handleSubmitCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError(null);
    if (!name.trim() || !description.trim() || !startingBalance.trim()) {
      setCreateError("All fields are required.");
      return;
    }
    const start = parseFloat(startingBalance);
    if (Number.isNaN(start) || start <= 0) {
      setCreateError("Starting balance must be a positive number.");
      return;
    }
    setSavingCreate(true);
    try {
      await createPortfolio({
        name: name.trim(),
        description: description.trim(),
        starting_balance: start,
      });
      setIsCreateOpen(false);
    } catch (err) {
      console.error("Failed to create portfolio", err);
      setCreateError("Failed to create portfolio.");
    } finally {
      setSavingCreate(false);
    }
  };

  const showCreateCta = !selectedPortfolio && portfolios.length === 0;

  useEffect(() => {
    const loadNames = async () => {
      const tickersToLoad = Array.from(
        new Set(
          positions
            .map((p) => p.ticker.toUpperCase())
            .filter((t) => t && !nameByTicker[t]),
        ),
      );
      if (tickersToLoad.length === 0) return;
      try {
        const results = await Promise.all(
          tickersToLoad.map((t) =>
            api
              .get("/v1/datafeed/stock", { params: { ticker: t } })
              .then((res) => ({
                ticker: t,
                name: (res.data?.results?.name as string | undefined) ?? t,
              }))
              .catch(() => ({ ticker: t, name: t })),
          ),
        );
        setNameByTicker((prev) => {
          const next = { ...prev };
          for (const r of results) {
            next[r.ticker] = r.name;
          }
          return next;
        });
      } catch (err) {
        // ignore name errors; ticker symbols still show
        console.error("Failed to load ticker names", err);
      }
    };
    void loadNames();
  }, [positions, nameByTicker]);

  useEffect(() => {
    // reset pagination when portfolio changes
    setOrdersPage(1);
  }, [selectedPortfolio?.uuid]);

  useEffect(() => {
    const loadOrders = async () => {
      if (!selectedPortfolio) {
        setOrders([]);
        setOrdersHasMore(false);
        return;
      }
      setOrdersLoading(true);
      try {
        const res = await api.get("/v1/orders", {
          params: {
            portfolio_uuid: selectedPortfolio.uuid,
            page: ordersPage,
            limit: 10,
          },
        });
        setOrders(res.data.orders ?? []);
        setOrdersHasMore(Boolean(res.data.has_more));
      } catch (err) {
        console.error("Failed to load orders", err);
        setOrders([]);
        setOrdersHasMore(false);
      } finally {
        setOrdersLoading(false);
      }
    };
    void loadOrders();
  }, [selectedPortfolio?.uuid, ordersPage]);

  const allWatchlistTickers = Array.from(
    new Set(watchlists.flatMap((w) => w.tickers)),
  );
  const marketData = useWatchlistMarketData(allWatchlistTickers);

  return (
    <div className="flex flex-col max-w-6xl w-full mx-auto h-full bg-black text-white overflow-y-auto">
      <div className="flex-1 flex flex-col px-8 py-8 gap-6">
        <div>
          <h1 className="text-2xl font-semibold">
            Welcome{accountData ? `, ${accountData.first_name}` : ""}.
          </h1>
          <p className="mt-2 text-sm text-neutral-400">
            This is your CacheFlow dashboard.
          </p>
        </div>

        {showCreateCta && (
          <div className="mt-4 rounded-xl flex flex-col items-center justify-center border border-dashed border-neutral-800 bg-neutral-900/30 px-6 py-12 w-full">
            <Hourglass size={64} className="text-white" />
            <h2 className="text-xl font-semibold text-white mt-4">
              Create a portfolio to get started
            </h2>
            <p className="mt-2 text-md text-neutral-400">
              Set an initial budget and start tracking positions and PnL.
            </p>
            <button
              type="button"
              onClick={handleOpenCreate}
              className="mt-8 inline-flex items-center rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-4 hover:bg-orange-600 transition-colors"
            >
              Create portfolio
            </button>
          </div>
        )}

        {selectedPortfolio && (
          <>
            <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-6 py-5">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-sm font-semibold text-white">
                  {selectedPortfolio.name}
                </h2>
                <span className="text-xs text-neutral-500">
                  Portfolio overview (chart placeholder)
                </span>
              </div>
              <div className="h-52 rounded-lg bg-gradient-to-b from-neutral-800 to-neutral-950 border border-neutral-800 flex items-center justify-center text-neutral-600 text-xs">
                Price / equity chart coming soon
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                {TIMEFRAMES.map((tf) => (
                  <button
                    key={tf}
                    type="button"
                    onClick={() => setTimeframe(tf)}
                    className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
                      timeframe === tf
                        ? "bg-orange-500 text-black"
                        : "border border-neutral-700 text-neutral-300 hover:border-neutral-500 hover:text-white"
                    }`}
                  >
                    {tf.toUpperCase()}
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-6 py-5 max-w-md">
                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">
                  Current Balance
                </h2>
                <div className="mt-3 flex items-baseline gap-3">
                  <span className="text-xl font-semibold">
                    {buyingPower != null
                      ? `$${(buyingPower + (totalMarketValue ?? 0)).toFixed(2)}`
                      : "—"}
                  </span>
                </div>
                <p className="mt-1 text-[11px] text-neutral-500">
                  Buying power + market value of positions.
                </p>
              </div>
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-6 py-5 max-w-md">
                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">
                  Buying Power
                </h2>
                <div className="mt-3 flex items-baseline gap-3">
                  <span className="text-xl font-semibold">
                    {buyingPowerLoading
                      ? "—"
                      : buyingPower != null
                      ? `$${buyingPower.toFixed(2)}`
                      : "—"}
                  </span>
                </div>
              </div>
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-6 py-5 max-w-md">
                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">
                  Active Positions
                </h2>
                <div className="mt-3 flex items-baseline gap-3">
                  <span className="text-xl font-semibold">
                    {activePositionsCount}
                  </span>
                </div>
                <p className="mt-1 text-[11px] text-neutral-500">
                  Number of tickers with open positions.
                </p>
              </div>
            </div>
            {activePositionsCount > 0 && (
              <div className="mt-6">
                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500 mb-2">
                  Active positions
                </h2>
                <div className="space-y-2">
                  {positions.map((p) => {
                    const ticker = p.ticker.toUpperCase();
                    const companyName = nameByTicker[ticker] ?? ticker;
                    const unreal = p.unrealized ?? 0;
                    const absUnreal = Math.abs(unreal).toFixed(2);
                    const prefix =
                      unreal > 0 ? "+" : unreal < 0 ? "-" : "";
                    const unrealStr =
                      prefix === ""
                        ? `$${absUnreal}`
                        : `${prefix}$${absUnreal}`;
                    const worth = (p.current_price ?? 0) * (p.shares ?? 0);
                    const unrealClass =
                      unreal > 0
                        ? "text-green-400"
                        : unreal < 0
                        ? "text-red-400"
                        : "text-neutral-300";

                    return (
                      <button
                        key={ticker}
                        type="button"
                        onClick={() =>
                          navigate(
                            `/client-app/stock?ticker=${encodeURIComponent(
                              ticker,
                            )}`,
                          )
                        }
                        className="w-full rounded-xl border border-neutral-800 bg-neutral-900/40 px-4 py-3 flex items-center justify-between text-left hover:border-orange-500 transition-colors"
                      >
                        <div>
                          <div className="flex items-baseline gap-2">
                            <span className="text-sm font-semibold text-white">
                              {ticker}
                            </span>
                            <span className="text-xs text-neutral-400">
                              {companyName}
                            </span>
                          </div>
                          <div className="mt-1 text-[11px] text-neutral-500">
                            {p.shares} shares @ ${p.avg_cost.toFixed(2)}
                          </div>
                        </div>
                        <div className="flex flex-col items-end gap-0.5 text-xs">
                          <span className={unrealClass}>{unrealStr}</span>
                          <div className="flex items-center gap-1 text-neutral-300">
                            <span className="text-[11px]">
                              ${worth.toFixed(2)} total
                            </span>
                            <span className="text-neutral-500 text-sm">
                              ›
                            </span>
                          </div>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
            <div className="mt-8">
              <div className="flex items-center justify-between mb-2">
                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">
                  Watchlists
                </h2>
                <div className="flex items-center gap-2 text-[11px]">
                  <button
                    type="button"
                    onClick={() => {
                      setWatchlistName("");
                      setWatchlistError(null);
                      setIsWatchlistCreateOpen(true);
                    }}
                    className="rounded-full border border-neutral-700 px-3 py-1 text-neutral-200 hover:border-orange-500 hover:text-orange-300 transition-colors"
                  >
                    + Create
                  </button>
                  {watchlists.length > 1 && (
                    <button
                      type="button"
                      onClick={() => {
                        setReorderItems(watchlists.map((w) => w.uuid));
                        setIsReorderOpen(true);
                      }}
                      className="rounded-full border border-neutral-700 px-3 py-1 text-neutral-200 hover:border-neutral-500 hover:text-white transition-colors"
                    >
                      Reorder
                    </button>
                  )}
                </div>
              </div>
              <div className="rounded- py-3">
                {watchlistsLoading ? (
                  <p className="text-sm text-neutral-500">Loading watchlists…</p>
                ) : watchlists.length === 0 ? (
                  <p className="text-sm text-neutral-500">
                    You don&apos;t have any watchlists yet.
                  </p>
                ) : (
                  <div className="space-y-3">
                    {watchlists.map((wl) => (
                      <div
                        key={wl.uuid}
                        className="rounded-lg border border-neutral-800 bg-neutral-950/60 px-4 py-3"
                      >
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-2">
                            <Star size={12} className="text-yellow-400" />
                            <span className="text-sm font-medium text-white">
                              {wl.name}
                            </span>
                            <span className="text-[11px] text-neutral-500">
                              {wl.tickers.length} tickers
                            </span>
                          </div>
                          <div className="flex items-center gap-3 text-[11px]">
                            <button
                              type="button"
                              onClick={() => {
                                setActiveWatchlistForAdd(wl.uuid);
                                setAddSearchQuery("");
                                setAddSearchResults([]);
                                setAddSearchError(null);
                              }}
                              className="text-neutral-500 hover:text-neutral-200"
                            >
                              Add tickers
                            </button>
                            <button
                              type="button"
                              disabled={wl.tickers.length === 0}
                              onClick={() => {
                                if (wl.tickers.length === 0) return;
                                setActiveWatchlistForManage(wl.uuid);
                              }}
                              className={`${
                                wl.tickers.length === 0
                                  ? "text-neutral-700 cursor-not-allowed"
                                  : "text-neutral-500 hover:text-neutral-200"
                              }`}
                            >
                              Manage
                            </button>
                            <button
                              type="button"
                              onClick={() => {
                                if (
                                  window.confirm(
                                    `Delete watchlist "${wl.name}"?`,
                                  )
                                ) {
                                  void deleteWatchlist(wl.uuid);
                                }
                              }}
                              className="text-neutral-500 hover:text-red-400"
                            >
                              Delete
                            </button>
                          </div>
                        </div>
                        {wl.tickers.length === 0 ? (
                          <p className="text-[11px] text-neutral-500">
                            No tickers in this watchlist yet.
                          </p>
                        ) : (
                          <div className="space-y-2">
                            {wl.tickers.map((t) => {
                              const ticker = t.toUpperCase();
                              const md = marketData[ticker];
                              const snapshot = md?.snapshot ?? null;
                              const series = md?.series ?? [];
                              const last =
                                snapshot?.lastPrice ??
                                (series.length
                                  ? series[series.length - 1].c
                                  : null);
                              const change = snapshot?.change ?? null;
                              const changePct =
                                snapshot?.changePercent ?? null;
                              const changePrefix =
                                change != null && change !== 0
                                  ? change > 0
                                    ? "+"
                                    : "-"
                                  : "";
                              const changeAbs =
                                change != null
                                  ? Math.abs(change).toFixed(2)
                                  : null;
                              const changeText =
                                changeAbs == null
                                  ? "—"
                                  : changePrefix === ""
                                  ? `$${changeAbs}`
                                  : `${changePrefix}$${changeAbs}`;
                              const changePctText =
                                changePct != null
                                  ? `${changePct > 0 ? "+" : ""}${changePct.toFixed(
                                      2,
                                    )}%`
                                  : null;
                              const changeClass =
                                change != null && change !== 0
                                  ? change > 0
                                    ? "text-green-400"
                                    : "text-red-400"
                                  : "text-neutral-300";

                              const name =
                                nameByTicker[ticker] ?? ticker;

                              return (
                                <button
                                  key={ticker}
                                  type="button"
                                  onClick={() =>
                                    navigate(
                                      `/client-app/stock?ticker=${encodeURIComponent(
                                        ticker,
                                      )}`,
                                    )
                                  }
                                  className="w-full h-20 flex items-center justify-between gap-3 rounded-md border border-neutral-900 bg-neutral-900/60 px-3 py-2 hover:border-orange-500 transition-colors"
                                >
                                  <div className="flex w-64 max-w-64 flex-col items-start text-left min-w-[80px]">
                                    <span className="text-xs font-semibold text-white">
                                      {ticker}
                                    </span>
                                    <span className="text-[11px] text-neutral-400 line-clamp-1">
                                      {name}
                                    </span>
                                  </div>
                                  <div className="flex- h-10 mx-auto justify-center items-center">
                                    {series.length > 0 ? (
                                      <ResponsiveContainer
                                        width={200}
                                        height="100%"
                                      >
<AreaChart data={series}>
  <defs>{/* … existing gradient … */}</defs>
  <YAxis
    hide
    domain={['dataMin', 'dataMax']} // or a padded version
  />
  <Area
    type="monotone"
    dataKey="c"
    stroke="rgb(249 115 22)"
    strokeWidth={1.5}
    fill={`url(#wl-${ticker}-fill)`}
  />
</AreaChart>
                                      </ResponsiveContainer>
                                    ) : (
                                      <div className="w-full h-full flex items-center justify-center text-[10px] text-neutral-600">
                                        No data
                                      </div>
                                    )}
                                  </div>
                                  <div className="flex flex-col  items-end min-w-[90px]">
                                    <span className="text-xs text-white">
                                      {last != null
                                        ? `$${last.toFixed(2)}`
                                        : "—"}
                                    </span>
                                    <span
                                      className={`text-[11px] ${changeClass}`}
                                    >
                                      {changeText}
                                      {changePctText &&
                                        ` (${changePctText})`}
                                    </span>
                                  </div>
                                </button>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
            <div className="mt-8">
              <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500 mb-2">
                Order history
              </h2>
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 px-4 py-3">
                {ordersLoading ? (
                  <p className="text-sm text-neutral-500">Loading orders…</p>
                ) : orders.length === 0 ? (
                  <p className="text-sm text-neutral-500">
                    No orders yet for this portfolio.
                  </p>
                ) : (
                  <div className="space-y-1 text-xs">
                    {orders.map((o) => {
                      const side = o.side ?? o.Side;
                      const ticker = (o.ticker ?? o.Ticker ?? "").toUpperCase();
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
                              <span className="text-[11px] text-neutral-300">
                                {ticker}
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
          </>
        )}
      </div>

      {isCreateOpen && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/70">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950/95 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white tracking-wide">
                Create portfolio
              </h2>
              <button
                onClick={() => setIsCreateOpen(false)}
                className="text-neutral-500 hover:text-neutral-200 text-sm"
              >
                Close
              </button>
            </div>

            <form onSubmit={handleSubmitCreate} className="space-y-3">
              <div>
                <label className="text-white text-sm font-regular">
                  Name
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full h-9 text-white mt-1 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm"
                />
              </div>
              <div>
                <label className="text-white text-sm font-regular">
                  Description
                </label>
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full h-9 text-white mt-1 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm"
                />
              </div>
              <div>
                <label className="text-white text-sm font-regular">
                  Starting balance
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={startingBalance}
                  onChange={(e) => setStartingBalance(e.target.value)}
                  className="w-full h-9 text-white mt-1 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm"
                />
              </div>

              {createError && (
                <p className="text-xs text-red-400 mt-1">
                  {createError}
                </p>
              )}

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={savingCreate}
                  className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {savingCreate ? "Creating…" : "Create portfolio"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {isWatchlistCreateOpen && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white tracking-wide">
                Create watchlist
              </h2>
              <button
                onClick={() => setIsWatchlistCreateOpen(false)}
                className="text-neutral-500 hover:text-neutral-200 text-sm"
              >
                Close
              </button>
            </div>
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                setWatchlistError(null);
                if (!watchlistName.trim()) {
                  setWatchlistError("Name is required.");
                  return;
                }
                setSavingWatchlist(true);
                try {
                  await createWatchlist({
                    name: watchlistName.trim(),
                  });
                  setIsWatchlistCreateOpen(false);
                } catch (err) {
                  console.error("Failed to create watchlist", err);
                  setWatchlistError("Failed to create watchlist.");
                } finally {
                  setSavingWatchlist(false);
                }
              }}
              className="space-y-3"
            >
              <div>
                <label className="text-white text-sm font-regular">
                  Name
                </label>
                <input
                  type="text"
                  value={watchlistName}
                  onChange={(e) => setWatchlistName(e.target.value)}
                  className="w-full h-9 text-white mt-1 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm"
                />
              </div>
              {watchlistError && (
                <p className="text-xs text-red-400 mt-1">{watchlistError}</p>
              )}
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsWatchlistCreateOpen(false)}
                  className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={savingWatchlist}
                  className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {savingWatchlist ? "Creating…" : "Create watchlist"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {isReorderOpen && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white tracking-wide">
                Reorder watchlists
              </h2>
              <button
                onClick={() => setIsReorderOpen(false)}
                className="text-neutral-500 hover:text-neutral-200 text-sm"
              >
                Close
              </button>
            </div>
            <div className="space-y-2 mb-4">
              {reorderItems.map((id) => {
                const wl = watchlists.find((w) => w.uuid === id);
                if (!wl) return null;
                return (
                  <div
                    key={id}
                    className="flex items-center justify-between gap-2 rounded-lg border border-neutral-800 bg-neutral-900/80 px-3 py-2"
                  >
                    <div className="flex items-center gap-2">
                      <GripVertical size={14} className="text-neutral-500" />
                      <span className="text-xs text-white">{wl.name}</span>
                    </div>
                    <div className="flex items-center gap-1 text-[11px]">
                      <button
                        type="button"
                        onClick={() => {
                          setReorderItems((prev) => {
                            const arr = [...prev];
                            const idx = arr.indexOf(id);
                            if (idx <= 0) return prev;
                            const tmp = arr[idx - 1];
                            arr[idx - 1] = arr[idx];
                            arr[idx] = tmp;
                            return arr;
                          });
                        }}
                        className="px-2 py-1 rounded-full border border-neutral-700 text-neutral-300 hover:border-neutral-500 hover:text-white"
                      >
                        ↑
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          setReorderItems((prev) => {
                            const arr = [...prev];
                            const idx = arr.indexOf(id);
                            if (idx === -1 || idx >= arr.length - 1) return prev;
                            const tmp = arr[idx + 1];
                            arr[idx + 1] = arr[idx];
                            arr[idx] = tmp;
                            return arr;
                          });
                        }}
                        className="px-2 py-1 rounded-full border border-neutral-700 text-neutral-300 hover:border-neutral-500 hover:text-white"
                      >
                        ↓
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setIsReorderOpen(false)}
                className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={async () => {
                  try {
                    await reorderWatchlists(reorderItems);
                    setIsReorderOpen(false);
                  } catch (err) {
                    console.error("Failed to reorder watchlists", err);
                  }
                }}
                className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors"
              >
                Save order
              </button>
            </div>
          </div>
        </div>
      )}

      {activeWatchlistForAdd && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white tracking-wide">
                Add tickers to watchlist
              </h2>
              <button
                onClick={() => {
                  setActiveWatchlistForAdd(null);
                  setAddSearchQuery("");
                  setAddSearchResults([]);
                  setAddSearchError(null);
                }}
                className="text-neutral-500 hover:text-neutral-200 text-sm"
              >
                Close
              </button>
            </div>
            <div className="mb-3">
              <input
                type="text"
                value={addSearchQuery}
                onChange={(e) => setAddSearchQuery(e.target.value)}
                placeholder="Search tickers"
                className="w-full h-9 text-sm text-white rounded-lg border border-neutral-800 bg-transparent px-2 focus:border-white focus:border-2 outline-none"
              />
            </div>
            {addSearchError && (
              <p className="text-xs text-red-400 mb-2">{addSearchError}</p>
            )}
            <div className="max-h-64 overflow-y-auto space-y-1 text-sm">
              {addSearchLoading ? (
                <p className="text-xs text-neutral-500">Searching…</p>
              ) : addSearchResults.length === 0 ? (
                <p className="text-xs text-neutral-500">No results.</p>
              ) : (
                addSearchResults.map((c) => {
                  const ticker = c.ticker.toUpperCase();
                  return (
                    <button
                      key={ticker}
                      type="button"
                      onClick={async () => {
                        const wl = watchlists.find(
                          (w) => w.uuid === activeWatchlistForAdd,
                        );
                        if (!wl) return;
                        if (wl.tickers.includes(ticker)) return;
                        const nextTickers = [...wl.tickers, ticker];
                        await updateWatchlist({
                          uuid: wl.uuid,
                          tickers: nextTickers,
                        });
                        if (!nameByTicker[ticker]) {
                          setNameByTicker((prev) => ({
                            ...prev,
                            [ticker]: c.name,
                          }));
                        }
                      }}
                      className="w-full flex items-center justify-between rounded-lg border border-neutral-800 bg-neutral-900/60 px-3 py-2 text-[11px] text-neutral-200 hover:border-orange-500 transition-colors"
                    >
                      <span>
                        {ticker} — {c.name}
                      </span>
                    </button>
                  );
                })
              )}
            </div>
          </div>
        </div>
      )}

      {activeWatchlistForManage && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white tracking-wide">
                Manage watchlist tickers
              </h2>
              <button
                onClick={() => setActiveWatchlistForManage(null)}
                className="text-neutral-500 hover:text-neutral-200 text-sm"
              >
                Close
              </button>
            </div>
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {(() => {
                const wl = watchlists.find(
                  (w) => w.uuid === activeWatchlistForManage,
                );
                if (!wl || wl.tickers.length === 0) {
                  return (
                    <p className="text-xs text-neutral-500">
                      This watchlist has no tickers.
                    </p>
                  );
                }
                return wl.tickers.map((t) => {
                  const ticker = t.toUpperCase();
                  const name = nameByTicker[ticker] ?? ticker;
                  return (
                    <div
                      key={ticker}
                      className="flex items-center justify-between rounded-lg border border-neutral-800 bg-neutral-900/60 px-3 py-2 text-[11px] text-neutral-200"
                    >
                      <span>
                        {ticker} — {name}
                      </span>
                      <button
                        type="button"
                        onClick={async () => {
                          const current = watchlists.find(
                            (w) => w.uuid === activeWatchlistForManage,
                          );
                          if (!current) return;
                          const nextTickers = current.tickers.filter(
                            (x) => x.toUpperCase() !== ticker,
                          );
                          await updateWatchlist({
                            uuid: current.uuid,
                            tickers: nextTickers,
                          });
                        }}
                        className="text-red-400 hover:text-red-300"
                      >
                        Remove
                      </button>
                    </div>
                  );
                });
              })()}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

