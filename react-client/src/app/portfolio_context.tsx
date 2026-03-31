import { createContext, useContext, useEffect, useState } from "react";
import api from "../util/api";

export interface Portfolio {
  uuid: string;
  name: string;
  description: string;
  starting_balance: number;
  current_balance: number;
}

interface PortfolioContextValue {
  portfolios: Portfolio[];
  selectedPortfolio: Portfolio | null;
  setSelectedPortfolioId: (id: string | null) => void;
  refreshPortfolios: () => Promise<void>;
  createPortfolio: (input: { name: string; description: string; starting_balance: number }) => Promise<void>;
  updatePortfolio: (input: { uuid: string; name: string; description: string; starting_balance?: number }) => Promise<void>;
  deletePortfolio: (uuid: string) => Promise<void>;
}

const PortfolioContext = createContext<PortfolioContextValue | undefined>(undefined);

const STORAGE_KEY = "cf-selected-portfolio";

export const PortfolioProvider = ({ children }: { children: React.ReactNode }) => {
  const [portfolios, setPortfolios] = useState<Portfolio[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const selectedPortfolio = portfolios.find((p) => p.uuid === selectedId) ?? null;

  const refreshPortfolios = async () => {
    const res = await api.get("/v1/portfolios");
    const data = (res.data as any[]).map((p) => ({
      uuid: p.uuid,
      name: p.name,
      description: p.description,
      starting_balance: p.starting_balance,
      current_balance: p.current_balance,
    })) as Portfolio[];
    setPortfolios(data);

    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && data.some((p) => p.uuid === stored)) {
      setSelectedId(stored);
      return;
    }
    if (!stored && data.length > 0 && !selectedId) {
      setSelectedId(data[0].uuid);
      localStorage.setItem(STORAGE_KEY, data[0].uuid);
      return;
    }
    if (data.length === 0) {
      setSelectedId(null);
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  useEffect(() => {
    refreshPortfolios().catch((err) => {
      console.error("Failed to load portfolios", err);
    });
  }, []);

  const setSelectedPortfolioId = (id: string | null) => {
    setSelectedId(id);
    if (id) {
      localStorage.setItem(STORAGE_KEY, id);
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  const createPortfolio: PortfolioContextValue["createPortfolio"] = async (input) => {
    await api.post("/v1/portfolio", {
      name: input.name,
      description: input.description,
      starting_balance: input.starting_balance,
    });
    await refreshPortfolios();
  };

  const updatePortfolio: PortfolioContextValue["updatePortfolio"] = async (input) => {
    const payload: any = {
      uuid: input.uuid,
      name: input.name,
      description: input.description,
    };
    if (typeof input.starting_balance === "number") {
      payload.starting_balance = input.starting_balance;
    }
    await api.put("/v1/portfolio", payload);
    await refreshPortfolios();
  };

  const deletePortfolio: PortfolioContextValue["deletePortfolio"] = async (uuid: string) => {
    await api.delete("/v1/portfolio", {
      data: { uuid },
    });
    await refreshPortfolios();
  };

  return (
    <PortfolioContext.Provider
      value={{
        portfolios,
        selectedPortfolio,
        setSelectedPortfolioId,
        refreshPortfolios,
        createPortfolio,
        updatePortfolio,
        deletePortfolio,
      }}
    >
      {children}
    </PortfolioContext.Provider>
  );
};

export const usePortfolio = () => {
  const ctx = useContext(PortfolioContext);
  if (!ctx) {
    throw new Error("usePortfolio must be used within a PortfolioProvider");
  }
  return ctx;
};

