import { useEffect, useState } from "react";
import api from "../util/api";

type Timeframe = "1d" | "1w" | "1m" | "3m" | "1y" | "all";

interface Bar {
  t: number; // unix ms
  c: number;
}

export interface ChartPoint {
  t: number;
  value: number;
}

interface Order {
  ticker: string;
  side: "BUY" | "SELL";
  quantity: number;
  price: number;
  timestampMs: number;
}

async function fetchAllOrders(portfolioUUID: string): Promise<Order[]> {
  const all: Order[] = [];
  let page = 1;
  for (;;) {
    const res = await api.get("/v1/orders", {
      params: { portfolio_uuid: portfolioUUID, page, limit: 100 },
    });
    const rawOrders = (res.data.orders ?? []) as any[];
    for (const o of rawOrders) {
      const ts =
        o.timestamp ?? o.Timestamp ?? o.created_at ?? o.CreatedAt ?? null;
      const ms = ts ? new Date(ts as string).getTime() : 0;
      all.push({
        ticker: ((o.ticker ?? o.Ticker ?? "") as string).toUpperCase(),
        side: (o.side ?? o.Side ?? "BUY") as "BUY" | "SELL",
        quantity: Number(o.quantity ?? o.Quantity ?? 0),
        price: Number(o.price ?? o.Price ?? 0),
        timestampMs: ms,
      });
    }
    if (!res.data.has_more) break;
    page++;
  }
  return all;
}

async function fetchBars(ticker: string, timeframe: Timeframe): Promise<Bar[]> {
  try {
    const res = await api.get("/v1/datafeed/stock/aggregates", {
      params: { ticker, timeframe },
    });
    return ((res.data.results ?? []) as any[]).map((r: any) => ({
      t: r.t as number,
      c: r.c as number,
    }));
  } catch {
    return [];
  }
}

export function usePortfolioChart(
  portfolio: { uuid: string; starting_balance: number } | null,
  timeframe: Timeframe,
) {
  const [data, setData] = useState<ChartPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!portfolio) {
      setData([]);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);
    setData([]);

    const run = async () => {
      try {
        const orders = await fetchAllOrders(portfolio.uuid);
        if (cancelled) return;

        orders.sort((a, b) => a.timestampMs - b.timestampMs);

        const tickers = Array.from(
          new Set(orders.map((o) => o.ticker).filter(Boolean)),
        );

        if (tickers.length === 0) {
          if (!cancelled) {
            setData([]);
            setLoading(false);
          }
          return;
        }

        const barsMap: Record<string, Bar[]> = {};
        await Promise.all(
          tickers.map(async (t) => {
            barsMap[t] = await fetchBars(t, timeframe);
          }),
        );
        if (cancelled) return;

        // Unified sorted timestamp list across all tickers
        const allTimestamps = Array.from(
          new Set(
            Object.values(barsMap).flatMap((bars) => bars.map((b) => b.t)),
          ),
        ).sort((a, b) => a - b);

        if (allTimestamps.length === 0) {
          if (!cancelled) {
            setData([]);
            setLoading(false);
          }
          return;
        }

        const chartStart = allTimestamps[0];

        // Replay orders that happened before the chart window to get initial state
        let cash = portfolio.starting_balance;
        const shares: Record<string, number> = {};
        let orderIdx = 0;

        while (
          orderIdx < orders.length &&
          orders[orderIdx].timestampMs < chartStart
        ) {
          const o = orders[orderIdx++];
          if (o.side === "BUY") {
            cash -= o.quantity * o.price;
            shares[o.ticker] = (shares[o.ticker] ?? 0) + o.quantity;
          } else {
            cash += o.quantity * o.price;
            shares[o.ticker] = Math.max(
              0,
              (shares[o.ticker] ?? 0) - o.quantity,
            );
          }
        }

        // Sorted bars per ticker for carry-forward price lookups
        const sortedBars: Record<string, Bar[]> = {};
        for (const ticker of tickers) {
          sortedBars[ticker] = [...(barsMap[ticker] ?? [])].sort(
            (a, b) => a.t - b.t,
          );
        }

        const tickerBarIdx: Record<string, number> = {};
        const tickerLastClose: Record<string, number> = {};
        for (const ticker of tickers) {
          tickerBarIdx[ticker] = 0;
          tickerLastClose[ticker] = 0;
        }

        const chartPoints: ChartPoint[] = [];

        for (const ts of allTimestamps) {
          // Apply any orders that occurred at or before this bar
          while (
            orderIdx < orders.length &&
            orders[orderIdx].timestampMs <= ts
          ) {
            const o = orders[orderIdx++];
            if (o.side === "BUY") {
              cash -= o.quantity * o.price;
              shares[o.ticker] = (shares[o.ticker] ?? 0) + o.quantity;
            } else {
              cash += o.quantity * o.price;
              shares[o.ticker] = Math.max(
                0,
                (shares[o.ticker] ?? 0) - o.quantity,
              );
            }
          }

          // Advance carry-forward close price for each ticker
          for (const ticker of tickers) {
            const bars = sortedBars[ticker];
            while (
              tickerBarIdx[ticker] < bars.length &&
              bars[tickerBarIdx[ticker]].t <= ts
            ) {
              tickerLastClose[ticker] = bars[tickerBarIdx[ticker]].c;
              tickerBarIdx[ticker]++;
            }
          }

          // Portfolio value = cash + sum(shares * last known price)
          let positionValue = 0;
          for (const [ticker, qty] of Object.entries(shares)) {
            if (qty > 0) {
              positionValue += qty * (tickerLastClose[ticker] ?? 0);
            }
          }

          chartPoints.push({ t: ts, value: Math.max(0, cash + positionValue) });
        }

        if (!cancelled) {
          setData(chartPoints);
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          console.error("Failed to build portfolio chart", err);
          setError("Failed to load chart data");
          setLoading(false);
        }
      }
    };

    void run();
    return () => {
      cancelled = true;
    };
  }, [portfolio?.uuid, portfolio?.starting_balance, timeframe]);

  return { data, loading, error };
}
