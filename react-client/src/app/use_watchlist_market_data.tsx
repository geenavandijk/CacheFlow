import { useEffect, useState } from "react";
import api from "../util/api";

interface Snapshot {
  ticker: string;
  lastPrice: number | null;
  change: number | null;
  changePercent: number | null;
}

export interface MiniSeriesPoint {
  t: number;
  c: number;
}

export interface WatchlistMarketEntry {
  snapshot: Snapshot | null;
  series: MiniSeriesPoint[];
}

type MarketMap = Record<string, WatchlistMarketEntry>;

export const useWatchlistMarketData = (tickers: string[]): MarketMap => {
  const [data, setData] = useState<MarketMap>({});

  useEffect(() => {
    const normalized = Array.from(
      new Set(
        tickers
          .map((t) => (t ?? "").toUpperCase())
          .filter((t) => t.length > 0),
      ),
    );
    if (normalized.length === 0) {
      setData({});
      return;
    }

    let cancelled = false;

    const fetchOnce = async () => {
      try {
        // Snapshots in one batched call
        const snapRes = await api.get("/v1/datafeed/snapshots", {
          params: { tickers: normalized.join(",") },
        });
        const snapshotsRaw = (snapRes.data?.snapshots ?? []) as any[];
        const byTickerSnap: Record<string, Snapshot> = {};
        for (const s of snapshotsRaw) {
          const t = (s.ticker ?? "").toUpperCase();
          if (!t) continue;
          byTickerSnap[t] = {
            ticker: t,
            lastPrice:
              typeof s.last_price === "number" ? s.last_price : null,
            change: typeof s.change === "number" ? s.change : null,
            changePercent:
              typeof s.change_percent === "number" ? s.change_percent : null,
          };
        }

        // Daily aggregates per ticker over recent window (small series)
        const aggResults = await Promise.all(
          normalized.map((t) =>
            (async () => {
              try {
                const now = new Date();
                const to = now.toISOString().slice(0, 10);
                const fromDate = new Date(
                  now.getTime() - 30 * 24 * 60 * 60 * 1000,
                );
                const from = fromDate.toISOString().slice(0, 10);
                const res = await api.get(
                  "/v1/datafeed/stock/aggregates-range",
                  {
                    params: {
                      ticker: t,
                      multiplier: 1,
                      timespan: "day",
                      from,
                      to,
                    },
                  },
                );
                return {
                  ticker: t,
                  results: (res.data?.results ?? []) as any[],
                };
              } catch {
                return { ticker: t, results: [] as any[] };
              }
            })(),
          ),
        );

        if (cancelled) return;

        setData((prev) => {
          const next: MarketMap = { ...prev };
          for (const t of normalized) {
            const snap = byTickerSnap[t] ?? null;
            const agg = aggResults.find((a) => a.ticker === t);
            const series: MiniSeriesPoint[] = (agg?.results ?? []).map(
              (r) => ({
                t: r.t as number,
                c: r.c as number,
              }),
            );
            next[t] = {
              snapshot: snap,
              series,
            };
          }
          return next;
        });
      } catch (err) {
        // On failure, keep existing data but log
        console.error("Failed to load watchlist market data", err);
      }
    };

    void fetchOnce();
    const id = window.setInterval(fetchOnce, 3000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [tickers.join(",")]);

  return data;
};

