import { createContext, useContext, useEffect, useState } from "react";
import api from "../util/api";
import { usePortfolio } from "./portfolio_context";

export interface Watchlist {
  uuid: string;
  name: string;
  index: number;
  tickers: string[];
}

interface WatchlistsContextValue {
  watchlists: Watchlist[];
  loading: boolean;
  error: string | null;
  refreshWatchlists: () => Promise<void>;
  createWatchlist: (input: {
    name: string;
    tickers?: string[];
  }) => Promise<void>;
  updateWatchlist: (input: {
    uuid: string;
    name?: string;
    tickers?: string[];
  }) => Promise<void>;
  deleteWatchlist: (uuid: string) => Promise<void>;
  reorderWatchlists: (order: string[]) => Promise<void>;
}

const WatchlistsContext = createContext<WatchlistsContextValue | undefined>(
  undefined,
);

export const WatchlistsProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { selectedPortfolio } = usePortfolio();
  const [watchlists, setWatchlists] = useState<Watchlist[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refreshWatchlists = async () => {
    if (!selectedPortfolio) {
      setWatchlists([]);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await api.get("/v1/portfolio/watchlists", {
        params: { portfolio_uuid: selectedPortfolio.uuid },
      });
      if (!Array.isArray(res.data)) {
        // In case the backend returns an error object or null instead of an array.
        setWatchlists([]);
        return;
      }
      const data = (res.data as any[]).map((w) => ({
        uuid: w.uuid,
        name: w.name,
        index: w.index ?? 0,
        tickers: (w.tickers ?? []).map((t: string) =>
          (t ?? "").toUpperCase(),
        ),
      })) as Watchlist[];
      // ensure sorted by index
      data.sort((a, b) => a.index - b.index);
      setWatchlists(data);
    } catch (err) {
      console.error("Failed to load watchlists", err);
      setError("Failed to load watchlists");
      setWatchlists([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refreshWatchlists();
  }, [selectedPortfolio?.uuid]);

  const createWatchlist: WatchlistsContextValue["createWatchlist"] = async (
    input,
  ) => {
    if (!selectedPortfolio) return;
    await api.post("/v1/portfolio/watchlist", {
      portfolio_uuid: selectedPortfolio.uuid,
      name: input.name,
      tickers: input.tickers,
    });
    await refreshWatchlists();
  };

  const updateWatchlist: WatchlistsContextValue["updateWatchlist"] = async (
    input,
  ) => {
    if (!selectedPortfolio) return;
    await api.put("/v1/portfolio/watchlist", {
      portfolio_uuid: selectedPortfolio.uuid,
      watchlist_uuid: input.uuid,
      name: input.name,
      tickers: input.tickers,
    });
    await refreshWatchlists();
  };

  const deleteWatchlist: WatchlistsContextValue["deleteWatchlist"] = async (
    uuid,
  ) => {
    if (!selectedPortfolio) return;
    await api.delete("/v1/portfolio/watchlist", {
      data: {
        portfolio_uuid: selectedPortfolio.uuid,
        watchlist_uuid: uuid,
      },
    });
    await refreshWatchlists();
  };

  const reorderWatchlists: WatchlistsContextValue["reorderWatchlists"] = async (
    order,
  ) => {
    if (!selectedPortfolio) return;
    await api.put("/v1/portfolio/watchlists/reorder", {
      portfolio_uuid: selectedPortfolio.uuid,
      order,
    });
    await refreshWatchlists();
  };

  return (
    <WatchlistsContext.Provider
      value={{
        watchlists,
        loading,
        error,
        refreshWatchlists,
        createWatchlist,
        updateWatchlist,
        deleteWatchlist,
        reorderWatchlists,
      }}
    >
      {children}
    </WatchlistsContext.Provider>
  );
};

export const useWatchlists = () => {
  const ctx = useContext(WatchlistsContext);
  if (!ctx) {
    throw new Error("useWatchlists must be used within a WatchlistsProvider");
  }
  return ctx;
};

