import { createContext, useContext, useEffect, useState } from "react";
import api from "../util/api";
import { usePortfolio } from "./portfolio_context";

export interface Position {
  ticker: string;
  shares: number;
  avg_cost: number;
  current_price: number;
  unrealized: number;
}

interface PositionsContextValue {
  positions: Position[];
  positionsByTicker: Record<string, Position>;
  totalUnrealized: number;
  totalMarketValue: number;
  activePositionsCount: number;
  refreshPositions: () => Promise<void>;
}

const PositionsContext = createContext<PositionsContextValue | undefined>(
  undefined,
);

export const PositionsProvider = ({ children }: { children: React.ReactNode }) => {
  const { selectedPortfolio } = usePortfolio();
  const [positions, setPositions] = useState<Position[]>([]);

  const refreshPositions = async () => {
    if (!selectedPortfolio) {
      setPositions([]);
      return;
    }
    const res = await api.get("/v1/portfolio/positions", {
      params: { portfolio_uuid: selectedPortfolio.uuid },
    });
    const list = (res.data.positions ?? []) as Position[];
    setPositions(list);
  };

  useEffect(() => {
    let timer: number | undefined;
    const poll = async () => {
      try {
        await refreshPositions();
      } catch (err) {
        console.error("Failed to load positions", err);
        setPositions([]);
      }
    };
    if (selectedPortfolio) {
      void poll();
      timer = window.setInterval(poll, 1000);
    } else {
      setPositions([]);
    }
    return () => {
      if (timer !== undefined) {
        clearInterval(timer);
      }
    };
  }, [selectedPortfolio?.uuid]);

  const positionsByTicker: Record<string, Position> = {};
  let totalUnrealized = 0;
  let totalMarketValue = 0;
  for (const p of positions) {
    positionsByTicker[p.ticker.toUpperCase()] = p;
    totalUnrealized += p.unrealized ?? 0;
    totalMarketValue += (p.current_price ?? 0) * (p.shares ?? 0);
  }
  const activePositionsCount = positions.length;

  return (
    <PositionsContext.Provider
      value={{
        positions,
        positionsByTicker,
        totalUnrealized,
        totalMarketValue,
        activePositionsCount,
        refreshPositions,
      }}
    >
      {children}
    </PositionsContext.Provider>
  );
};

export const usePositions = () => {
  const ctx = useContext(PositionsContext);
  if (!ctx) {
    throw new Error("usePositions must be used within a PositionsProvider");
  }
  return ctx;
};

