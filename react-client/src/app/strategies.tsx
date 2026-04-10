import { useEffect, useState, useCallback } from "react";
import { usePortfolio } from "./portfolio_context";
import api from "../util/api";
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  ReferenceLine,
} from "recharts";

// ── Types ──────────────────────────────────────────────────────────────────────

type RuleType =
  | "RSI_CROSSES_ABOVE"
  | "RSI_CROSSES_BELOW"
  | "RSI_ABOVE"
  | "RSI_BELOW"
  | "EMA_CROSS_ABOVE"
  | "EMA_CROSS_BELOW"
  | "SMA_CROSS_ABOVE"
  | "SMA_CROSS_BELOW"
  | "PRICE_ABOVE_EMA"
  | "PRICE_BELOW_EMA"
  | "PRICE_ABOVE_SMA"
  | "PRICE_BELOW_SMA"
  | "MACD_CROSS_SIGNAL_ABOVE"
  | "MACD_CROSS_SIGNAL_BELOW"
  | "MACD_ABOVE_ZERO"
  | "MACD_BELOW_ZERO"
  | "PRICE_ABOVE_VWAP_PCT"
  | "PRICE_BELOW_VWAP_PCT";

type SellConditionType =
  | "TAKE_PROFIT"
  | "STOP_LOSS"
  | "TRAILING_STOP"
  | "INDICATOR";

interface Rule {
  type: RuleType;
  value: number;
  window: number;
  fast_window: number;
  slow_window: number;
  fast_period: number;
  slow_period: number;
  signal_period: number;
  vwap_deviation: number;
}

interface SellCondition {
  type: SellConditionType;
  percent: number;
  rule?: Rule;
}

interface Strategy {
  uuid: string;
  name: string;
  description: string;
  ticker: string;
  buy_rules: Rule[];
  sell_conditions: SellCondition[];
  portfolio_uuid: string;
  created_at: string;
}

interface BacktestTrade {
  type: "BUY" | "SELL";
  date: string;
  price: number;
  shares: number;
  value: number;
  pnl: number;
  pnl_percent: number;
  cash_after: number;
}

interface Backtest {
  uuid: string;
  strategy_uuid: string;
  ticker: string;
  from_date: string;
  to_date: string;
  initial_balance: number;
  final_balance: number;
  roi: number;
  total_trades: number;
  winning_trades: number;
  losing_trades: number;
  max_drawdown: number;
  trades: BacktestTrade[];
  created_at: string;
}

interface MonteCarloResult {
  num_simulations: number;
  initial_balance: number;
  median: number;
  p10: number;
  p25: number;
  p75: number;
  p90: number;
  worst: number;
  best: number;
  prob_profit: number;
  final_values: number[];
}

// ── Rule config ────────────────────────────────────────────────────────────────

const RULE_LABELS: Record<RuleType, string> = {
  RSI_CROSSES_ABOVE: "RSI crosses above value",
  RSI_CROSSES_BELOW: "RSI crosses below value",
  RSI_ABOVE: "RSI is above value",
  RSI_BELOW: "RSI is below value",
  EMA_CROSS_ABOVE: "Fast EMA crosses above Slow EMA (bullish)",
  EMA_CROSS_BELOW: "Fast EMA crosses below Slow EMA (bearish)",
  SMA_CROSS_ABOVE: "Fast SMA crosses above Slow SMA (bullish)",
  SMA_CROSS_BELOW: "Fast SMA crosses below Slow SMA (bearish)",
  PRICE_ABOVE_EMA: "Price crosses above EMA",
  PRICE_BELOW_EMA: "Price crosses below EMA",
  PRICE_ABOVE_SMA: "Price crosses above SMA",
  PRICE_BELOW_SMA: "Price crosses below SMA",
  MACD_CROSS_SIGNAL_ABOVE: "MACD crosses above signal line (bullish)",
  MACD_CROSS_SIGNAL_BELOW: "MACD crosses below signal line (bearish)",
  MACD_ABOVE_ZERO: "MACD histogram above zero",
  MACD_BELOW_ZERO: "MACD histogram below zero",
  PRICE_ABOVE_VWAP_PCT: "Price is X% above VWAP",
  PRICE_BELOW_VWAP_PCT: "Price is X% below VWAP",
};

// Fields required per rule type
const RULE_FIELDS: Record<RuleType, string[]> = {
  RSI_CROSSES_ABOVE: ["window", "value"],
  RSI_CROSSES_BELOW: ["window", "value"],
  RSI_ABOVE: ["window", "value"],
  RSI_BELOW: ["window", "value"],
  EMA_CROSS_ABOVE: ["fast_window", "slow_window"],
  EMA_CROSS_BELOW: ["fast_window", "slow_window"],
  SMA_CROSS_ABOVE: ["fast_window", "slow_window"],
  SMA_CROSS_BELOW: ["fast_window", "slow_window"],
  PRICE_ABOVE_EMA: ["window"],
  PRICE_BELOW_EMA: ["window"],
  PRICE_ABOVE_SMA: ["window"],
  PRICE_BELOW_SMA: ["window"],
  MACD_CROSS_SIGNAL_ABOVE: ["fast_period", "slow_period", "signal_period"],
  MACD_CROSS_SIGNAL_BELOW: ["fast_period", "slow_period", "signal_period"],
  MACD_ABOVE_ZERO: ["fast_period", "slow_period", "signal_period"],
  MACD_BELOW_ZERO: ["fast_period", "slow_period", "signal_period"],
  PRICE_ABOVE_VWAP_PCT: ["vwap_deviation"],
  PRICE_BELOW_VWAP_PCT: ["vwap_deviation"],
};

const FIELD_LABELS: Record<string, string> = {
  window: "Period",
  value: "Threshold value",
  fast_window: "Fast period",
  slow_window: "Slow period",
  fast_period: "Fast EMA",
  slow_period: "Slow EMA",
  signal_period: "Signal period",
  vwap_deviation: "% deviation from VWAP",
};

const FIELD_DEFAULTS: Record<string, number> = {
  window: 14,
  value: 50,
  fast_window: 9,
  slow_window: 21,
  fast_period: 12,
  slow_period: 26,
  signal_period: 9,
  vwap_deviation: 0.1,
};

function emptyRule(): Rule {
  return {
    type: "RSI_CROSSES_ABOVE",
    value: 55,
    window: 14,
    fast_window: 9,
    slow_window: 21,
    fast_period: 12,
    slow_period: 26,
    signal_period: 9,
    vwap_deviation: 0.1,
  };
}

function emptySellCondition(): SellCondition {
  return { type: "TAKE_PROFIT", percent: 10 };
}

function ruleDesc(rule: Rule): string {
  const label = RULE_LABELS[rule.type] ?? rule.type;
  const fields = RULE_FIELDS[rule.type] ?? [];
  const parts = fields.map((f) => {
    const v = (rule as any)[f] as number;
    if (f === "value") return `${v}`;
    if (f === "vwap_deviation") return `${v}%`;
    return `${v}`;
  });
  return `${label} (${parts.join(", ")})`;
}

function sellCondDesc(sc: SellCondition): string {
  switch (sc.type) {
    case "TAKE_PROFIT":
      return `Take profit at +${sc.percent}%`;
    case "STOP_LOSS":
      return `Stop loss at -${sc.percent}%`;
    case "TRAILING_STOP":
      return `Trailing stop: ${sc.percent}% from peak`;
    case "INDICATOR":
      return sc.rule ? `Indicator exit: ${ruleDesc(sc.rule)}` : "Indicator exit";
  }
}

// ── Shared input style ─────────────────────────────────────────────────────────

const inputCls =
  "w-full h-9 text-white rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm";

const selectCls =
  "w-full h-9 text-white rounded-lg border border-neutral-800 bg-neutral-950 px-2 text-sm focus:outline-none focus:border-white";

// ── Rule editor ────────────────────────────────────────────────────────────────

function RuleEditor({
  rule,
  onChange,
  onRemove,
  label,
}: {
  rule: Rule;
  onChange: (r: Rule) => void;
  onRemove: () => void;
  label: string;
}) {
  const fields = RULE_FIELDS[rule.type] ?? [];

  const handleTypeChange = (t: RuleType) => {
    const defaults = { ...emptyRule(), type: t };
    // Preserve common fields
    onChange(defaults);
  };

  const setField = (field: string, val: number) => {
    onChange({ ...rule, [field]: val });
  };

  return (
    <div className="rounded-lg border border-neutral-800 bg-neutral-900/40 p-3 space-y-2">
      <div className="flex items-center justify-between mb-1">
        <span className="text-[11px] font-semibold uppercase tracking-widest text-neutral-500">
          {label}
        </span>
        <button
          type="button"
          onClick={onRemove}
          className="text-[11px] text-neutral-600 hover:text-red-400 transition-colors"
        >
          Remove
        </button>
      </div>

      <div>
        <label className="text-[11px] text-neutral-500">Condition</label>
        <select
          value={rule.type}
          onChange={(e) => handleTypeChange(e.target.value as RuleType)}
          className={selectCls}
        >
          {(Object.keys(RULE_LABELS) as RuleType[]).map((t) => (
            <option key={t} value={t}>
              {RULE_LABELS[t]}
            </option>
          ))}
        </select>
      </div>

      <div className="grid grid-cols-2 gap-2">
        {fields.map((field) => (
          <div key={field}>
            <label className="text-[11px] text-neutral-500">
              {FIELD_LABELS[field] ?? field}
            </label>
            <input
              type="number"
              step={field === "vwap_deviation" ? "0.01" : "1"}
              min={field === "vwap_deviation" ? "0" : "1"}
              value={(rule as any)[field] ?? FIELD_DEFAULTS[field]}
              onChange={(e) => setField(field, parseFloat(e.target.value) || 0)}
              className={inputCls}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

// ── Sell condition editor ──────────────────────────────────────────────────────

function SellConditionEditor({
  condition,
  onChange,
  onRemove,
  label,
}: {
  condition: SellCondition;
  onChange: (sc: SellCondition) => void;
  onRemove: () => void;
  label: string;
}) {
  return (
    <div className="rounded-lg border border-neutral-800 bg-neutral-900/40 p-3 space-y-2">
      <div className="flex items-center justify-between mb-1">
        <span className="text-[11px] font-semibold uppercase tracking-widest text-neutral-500">
          {label}
        </span>
        <button
          type="button"
          onClick={onRemove}
          className="text-[11px] text-neutral-600 hover:text-red-400 transition-colors"
        >
          Remove
        </button>
      </div>

      <div>
        <label className="text-[11px] text-neutral-500">Exit type</label>
        <select
          value={condition.type}
          onChange={(e) =>
            onChange({ ...condition, type: e.target.value as SellConditionType })
          }
          className={selectCls}
        >
          <option value="TAKE_PROFIT">Take profit (% gain)</option>
          <option value="STOP_LOSS">Stop loss (% loss)</option>
          <option value="TRAILING_STOP">Trailing stop (% from peak)</option>
          <option value="INDICATOR">Indicator-based exit</option>
        </select>
      </div>

      {condition.type !== "INDICATOR" && (
        <div>
          <label className="text-[11px] text-neutral-500">
            {condition.type === "TAKE_PROFIT"
              ? "Profit target (%)"
              : condition.type === "STOP_LOSS"
                ? "Max loss (%)"
                : "Trailing % from peak"}
          </label>
          <input
            type="number"
            step="0.1"
            min="0.1"
            value={condition.percent}
            onChange={(e) =>
              onChange({ ...condition, percent: parseFloat(e.target.value) || 0 })
            }
            className={inputCls}
          />
        </div>
      )}

      {condition.type === "INDICATOR" && (
        <RuleEditor
          rule={condition.rule ?? emptyRule()}
          onChange={(r) => onChange({ ...condition, rule: r })}
          onRemove={() => {}}
          label="Exit indicator"
        />
      )}
    </div>
  );
}

// ── Strategy form (create / edit) ─────────────────────────────────────────────

function StrategyForm({
  initial,
  portfolioUUID,
  onSave,
  onCancel,
}: {
  initial?: Strategy;
  portfolioUUID: string;
  onSave: (s: Strategy) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(initial?.name ?? "");
  const [description, setDescription] = useState(initial?.description ?? "");
  const [ticker, setTicker] = useState(initial?.ticker ?? "");
  const [buyRules, setBuyRules] = useState<Rule[]>(
    initial?.buy_rules?.length ? initial.buy_rules : [emptyRule()],
  );
  const [sellConditions, setSellConditions] = useState<SellCondition[]>(
    initial?.sell_conditions ?? [],
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!name.trim() || !ticker.trim()) {
      setError("Name and ticker are required.");
      return;
    }
    if (buyRules.length === 0) {
      setError("Add at least one buy rule.");
      return;
    }
    setSaving(true);
    try {
      let result: Strategy;
      if (initial) {
        await api.put("/v1/strategy", {
          uuid: initial.uuid,
          name: name.trim(),
          description: description.trim(),
          ticker: ticker.trim().toUpperCase(),
          buy_rules: buyRules,
          sell_conditions: sellConditions,
        });
        result = {
          ...initial,
          name: name.trim(),
          description: description.trim(),
          ticker: ticker.trim().toUpperCase(),
          buy_rules: buyRules,
          sell_conditions: sellConditions,
        };
      } else {
        const res = await api.post("/v1/strategy", {
          name: name.trim(),
          description: description.trim(),
          ticker: ticker.trim().toUpperCase(),
          buy_rules: buyRules,
          sell_conditions: sellConditions,
          portfolio_uuid: portfolioUUID,
        });
        result = res.data as Strategy;
      }
      onSave(result);
    } catch (err: any) {
      setError(err?.response?.data?.message ?? "Failed to save strategy.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/80 overflow-y-auto py-8 px-4">
      <div className="w-full max-w-2xl rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-sm font-semibold text-white">
            {initial ? "Edit strategy" : "New strategy"}
          </h2>
          <button
            type="button"
            onClick={onCancel}
            className="text-neutral-500 hover:text-neutral-200 text-sm"
          >
            Cancel
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Basic info */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-neutral-400">Strategy name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. RSI + EMA Trend"
                className={inputCls + " mt-1"}
              />
            </div>
            <div>
              <label className="text-xs text-neutral-400">Ticker</label>
              <input
                type="text"
                value={ticker}
                onChange={(e) => setTicker(e.target.value.toUpperCase())}
                placeholder="e.g. AAPL"
                className={inputCls + " mt-1"}
              />
            </div>
          </div>
          <div>
            <label className="text-xs text-neutral-400">
              Description (optional)
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Brief description of this strategy"
              className={inputCls + " mt-1"}
            />
          </div>

          {/* Buy rules */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-semibold uppercase tracking-widest text-orange-500">
                Buy conditions (ALL must be true)
              </span>
              <button
                type="button"
                onClick={() => setBuyRules((r) => [...r, emptyRule()])}
                className="text-[11px] rounded-full border border-neutral-700 px-3 py-1 text-neutral-300 hover:border-orange-500 hover:text-orange-300 transition-colors"
              >
                + Add rule
              </button>
            </div>
            <div className="space-y-2">
              {buyRules.map((rule, i) => (
                <RuleEditor
                  key={i}
                  rule={rule}
                  label={`Buy rule ${i + 1}`}
                  onChange={(r) => {
                    const next = [...buyRules];
                    next[i] = r;
                    setBuyRules(next);
                  }}
                  onRemove={() =>
                    setBuyRules((r) => r.filter((_, idx) => idx !== i))
                  }
                />
              ))}
            </div>
          </div>

          {/* Sell conditions */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-semibold uppercase tracking-widest text-green-500">
                Exit conditions (optional — any triggers exit)
              </span>
              <button
                type="button"
                onClick={() =>
                  setSellConditions((s) => [...s, emptySellCondition()])
                }
                className="text-[11px] rounded-full border border-neutral-700 px-3 py-1 text-neutral-300 hover:border-green-500 hover:text-green-300 transition-colors"
              >
                + Add exit
              </button>
            </div>
            {sellConditions.length === 0 && (
              <p className="text-[11px] text-neutral-600">
                No exit conditions — strategy holds position through the full
                backtest period.
              </p>
            )}
            <div className="space-y-2">
              {sellConditions.map((sc, i) => (
                <SellConditionEditor
                  key={i}
                  condition={sc}
                  label={`Exit ${i + 1}`}
                  onChange={(s) => {
                    const next = [...sellConditions];
                    next[i] = s;
                    setSellConditions(next);
                  }}
                  onRemove={() =>
                    setSellConditions((s) => s.filter((_, idx) => idx !== i))
                  }
                />
              ))}
            </div>
          </div>

          {error && <p className="text-xs text-red-400">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onCancel}
              className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60"
            >
              {saving ? "Saving…" : initial ? "Save changes" : "Create strategy"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ── Backtest modal ─────────────────────────────────────────────────────────────

function BacktestModal({
  strategy,
  onClose,
  onComplete,
}: {
  strategy: Strategy;
  onClose: () => void;
  onComplete: (bt: Backtest) => void;
}) {
  const [ticker, setTicker] = useState(strategy.ticker);
  const [from, setFrom] = useState(() => {
    const d = new Date();
    d.setFullYear(d.getFullYear() - 1);
    return d.toISOString().split("T")[0];
  });
  const [to, setTo] = useState(() => new Date().toISOString().split("T")[0]);
  const [balance, setBalance] = useState("10000");
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    const bal = parseFloat(balance);
    if (!ticker.trim() || !from || !to || isNaN(bal) || bal <= 0) {
      setError("All fields are required and balance must be positive.");
      return;
    }
    setRunning(true);
    try {
      const res = await api.post("/v1/strategy/backtest", {
        strategy_uuid: strategy.uuid,
        ticker: ticker.toUpperCase().trim(),
        from_date: from,
        to_date: to,
        initial_balance: bal,
      });
      onComplete(res.data as Backtest);
    } catch (err: any) {
      setError(
        err?.response?.data?.message ?? "Backtest failed. Check your parameters.",
      );
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 px-4">
      <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-white">
            Run backtest — {strategy.name}
          </h2>
          <button
            onClick={onClose}
            className="text-neutral-500 hover:text-neutral-200 text-sm"
          >
            Close
          </button>
        </div>
        <form onSubmit={handleRun} className="space-y-3">
          <div>
            <label className="text-xs text-neutral-400">Ticker</label>
            <input
              type="text"
              value={ticker}
              onChange={(e) => setTicker(e.target.value.toUpperCase())}
              className={inputCls + " mt-1"}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-neutral-400">From date</label>
              <input
                type="date"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
                className={inputCls + " mt-1"}
              />
            </div>
            <div>
              <label className="text-xs text-neutral-400">To date</label>
              <input
                type="date"
                value={to}
                onChange={(e) => setTo(e.target.value)}
                className={inputCls + " mt-1"}
              />
            </div>
          </div>
          <div>
            <label className="text-xs text-neutral-400">
              Starting balance ($)
            </label>
            <input
              type="number"
              min="1"
              step="0.01"
              value={balance}
              onChange={(e) => setBalance(e.target.value)}
              className={inputCls + " mt-1"}
            />
          </div>
          {error && <p className="text-xs text-red-400">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={running}
              className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60"
            >
              {running ? "Running…" : "Run backtest"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ── Backtest results modal ─────────────────────────────────────────────────────

function BacktestResultsModal({
  backtest,
  onClose,
  onMonteCarlo,
}: {
  backtest: Backtest;
  onClose: () => void;
  onMonteCarlo: (bt: Backtest) => void;
}) {
  const roiPos = backtest.roi >= 0;
  // Equity at each trade point: after BUY = cash_after + position value; after SELL = cash_after
  const equityCurve = backtest.trades.map((t) => ({
    date: t.date,
    value:
      t.type === "BUY"
        ? Math.round((t.cash_after + t.shares * t.price) * 100) / 100
        : Math.round(t.cash_after * 100) / 100,
  }));

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/80 overflow-y-auto py-8 px-4">
      <div className="w-full max-w-2xl rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-sm font-semibold text-white">
              Backtest results — {backtest.ticker}
            </h2>
            <p className="text-[11px] text-neutral-500">
              {backtest.from_date} → {backtest.to_date} · $
              {backtest.initial_balance.toLocaleString()} starting balance
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-neutral-500 hover:text-neutral-200 text-sm"
          >
            Close
          </button>
        </div>

        {/* Summary stats */}
        <div className="grid grid-cols-3 gap-3 mb-4">
          {[
            {
              label: "Final balance",
              val: `$${backtest.final_balance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
              col: "",
            },
            {
              label: "ROI",
              val: `${roiPos ? "+" : ""}${backtest.roi.toFixed(2)}%`,
              col: roiPos ? "text-green-400" : "text-red-400",
            },
            {
              label: "Max drawdown",
              val: `-${backtest.max_drawdown.toFixed(2)}%`,
              col: "text-red-400",
            },
            {
              label: "Total trades",
              val: String(backtest.total_trades),
              col: "",
            },
            {
              label: "Winning trades",
              val: String(backtest.winning_trades),
              col: "text-green-400",
            },
            {
              label: "Losing trades",
              val: String(backtest.losing_trades),
              col: "text-red-400",
            },
          ].map((s) => (
            <div
              key={s.label}
              className="rounded-lg border border-neutral-800 bg-neutral-900/40 px-3 py-2"
            >
              <p className="text-[10px] uppercase tracking-widest text-neutral-500">
                {s.label}
              </p>
              <p className={`text-sm font-semibold mt-0.5 ${s.col || "text-white"}`}>
                {s.val}
              </p>
            </div>
          ))}
        </div>

        {/* Equity curve */}
        {equityCurve.length > 1 && (
          <div className="h-36 mb-4">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart
                data={equityCurve}
                margin={{ top: 2, right: 0, left: 0, bottom: 0 }}
              >
                <defs>
                  <linearGradient id="bt-fill" x1="0" y1="0" x2="0" y2="1">
                    <stop
                      offset="5%"
                      stopColor={roiPos ? "#4ade80" : "#f87171"}
                      stopOpacity={0.18}
                    />
                    <stop
                      offset="95%"
                      stopColor={roiPos ? "#4ade80" : "#f87171"}
                      stopOpacity={0}
                    />
                  </linearGradient>
                </defs>
                <XAxis
                  dataKey="date"
                  tick={{ fill: "#525252", fontSize: 9 }}
                  axisLine={false}
                  tickLine={false}
                  interval="preserveStartEnd"
                />
                <YAxis
                  hide
                  domain={[
                    (min: number) => min * 0.97,
                    (max: number) => max * 1.03,
                  ]}
                />
                <Tooltip
                  contentStyle={{
                    background: "#171717",
                    border: "1px solid #404040",
                    borderRadius: 8,
                    fontSize: 11,
                  }}
                  formatter={(v: number) => [
                    `$${v.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
                    "Equity",
                  ]}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke={roiPos ? "#4ade80" : "#f87171"}
                  strokeWidth={1.5}
                  fill="url(#bt-fill)"
                  dot={false}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}

        {/* Trade log */}
        <div className="max-h-52 overflow-y-auto space-y-1 mb-4">
          {backtest.trades.map((t, i) => (
            <div
              key={i}
              className="flex items-center justify-between text-[11px] border-b border-neutral-900 py-1 last:border-0"
            >
              <div className="flex items-center gap-2">
                <span
                  className={
                    t.type === "BUY"
                      ? "text-green-400 font-semibold"
                      : "text-red-400 font-semibold"
                  }
                >
                  {t.type}
                </span>
                <span className="text-neutral-400">{t.date}</span>
                <span className="text-neutral-300">
                  {t.shares} shares @ ${t.price.toFixed(2)}
                </span>
              </div>
              {t.type === "SELL" && (
                <span
                  className={t.pnl >= 0 ? "text-green-400" : "text-red-400"}
                >
                  {t.pnl >= 0 ? "+" : ""}${t.pnl.toFixed(2)} (
                  {t.pnl_percent >= 0 ? "+" : ""}
                  {t.pnl_percent.toFixed(2)}%)
                </span>
              )}
            </div>
          ))}
          {backtest.trades.length === 0 && (
            <p className="text-sm text-neutral-500">
              No trades were executed in this period.
            </p>
          )}
        </div>

        <div className="flex justify-between items-center">
          <p className="text-[10px] text-neutral-600">
            Saved{" "}
            {new Date(backtest.created_at).toLocaleDateString([], {
              month: "short",
              day: "numeric",
              year: "numeric",
            })}
          </p>
          <button
            type="button"
            onClick={() => onMonteCarlo(backtest)}
            className="rounded-full border border-neutral-700 px-4 py-1.5 text-xs text-neutral-300 hover:border-orange-500 hover:text-orange-300 transition-colors"
          >
            Monte Carlo simulation →
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Monte Carlo modal ──────────────────────────────────────────────────────────

function MonteCarloModal({
  backtest,
  onClose,
}: {
  backtest: Backtest;
  onClose: () => void;
}) {
  const [numSims, setNumSims] = useState("1000");
  const [balance, setBalance] = useState(String(backtest.initial_balance));
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<MonteCarloResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);
    setRunning(true);
    try {
      const res = await api.post("/v1/strategy/montecarlo", {
        backtest_uuid: backtest.uuid,
        num_simulations: parseInt(numSims) || 1000,
        initial_balance: parseFloat(balance) || backtest.initial_balance,
      });
      setResult(res.data as MonteCarloResult);
    } catch (err: any) {
      setError(err?.response?.data?.message ?? "Monte Carlo failed.");
    } finally {
      setRunning(false);
    }
  };

  // Build histogram data from final_values
  const histData = (() => {
    if (!result?.final_values?.length) return [];
    const vals = result.final_values;
    const min = vals[0];
    const max = vals[vals.length - 1];
    const buckets = 40;
    const step = (max - min) / buckets || 1;
    const counts = new Array(buckets).fill(0);
    for (const v of vals) {
      const idx = Math.min(Math.floor((v - min) / step), buckets - 1);
      counts[idx]++;
    }
    return counts.map((count, i) => ({
      x: Math.round((min + i * step) * 100) / 100,
      count,
    }));
  })();

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/80 overflow-y-auto py-8 px-4">
      <div className="w-full max-w-xl rounded-2xl border border-neutral-800 bg-neutral-950 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-sm font-semibold text-white">
              Monte Carlo simulation
            </h2>
            <p className="text-[11px] text-neutral-500">
              Randomised trade order from backtest on {backtest.ticker}
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-neutral-500 hover:text-neutral-200 text-sm"
          >
            Close
          </button>
        </div>

        {!result && (
          <form onSubmit={handleRun} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-neutral-400">Simulations</label>
                <input
                  type="number"
                  min="100"
                  max="10000"
                  value={numSims}
                  onChange={(e) => setNumSims(e.target.value)}
                  className={inputCls + " mt-1"}
                />
              </div>
              <div>
                <label className="text-xs text-neutral-400">
                  Starting balance ($)
                </label>
                <input
                  type="number"
                  min="1"
                  step="0.01"
                  value={balance}
                  onChange={(e) => setBalance(e.target.value)}
                  className={inputCls + " mt-1"}
                />
              </div>
            </div>
            {error && <p className="text-xs text-red-400">{error}</p>}
            <div className="flex justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={onClose}
                className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={running}
                className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60"
              >
                {running ? "Simulating…" : "Run Monte Carlo"}
              </button>
            </div>
          </form>
        )}

        {result && (
          <div className="space-y-4">
            {/* Stats grid */}
            <div className="grid grid-cols-2 gap-2">
              {[
                {
                  label: "Prob. of profit",
                  val: `${result.prob_profit.toFixed(1)}%`,
                  col:
                    result.prob_profit >= 50 ? "text-green-400" : "text-red-400",
                },
                {
                  label: "Median outcome",
                  val: `$${result.median.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                  col:
                    result.median >= result.initial_balance
                      ? "text-green-400"
                      : "text-red-400",
                },
                {
                  label: "10th percentile",
                  val: `$${result.p10.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                  col: "text-red-400",
                },
                {
                  label: "90th percentile",
                  val: `$${result.p90.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                  col: "text-green-400",
                },
                {
                  label: "Worst case",
                  val: `$${result.worst.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                  col: "text-red-400",
                },
                {
                  label: "Best case",
                  val: `$${result.best.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                  col: "text-green-400",
                },
              ].map((s) => (
                <div
                  key={s.label}
                  className="rounded-lg border border-neutral-800 bg-neutral-900/40 px-3 py-2"
                >
                  <p className="text-[10px] uppercase tracking-widest text-neutral-500">
                    {s.label}
                  </p>
                  <p className={`text-sm font-semibold mt-0.5 ${s.col}`}>
                    {s.val}
                  </p>
                </div>
              ))}
            </div>

            {/* Distribution histogram */}
            {histData.length > 0 && (
              <div>
                <p className="text-[10px] uppercase tracking-widest text-neutral-500 mb-2">
                  Final balance distribution ({result.num_simulations} simulations)
                </p>
                <div className="h-32">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart
                      data={histData}
                      margin={{ top: 2, right: 0, left: 0, bottom: 0 }}
                    >
                      <XAxis
                        dataKey="x"
                        tickFormatter={(v: number) =>
                          `$${(v / 1000).toFixed(0)}k`
                        }
                        tick={{ fill: "#525252", fontSize: 9 }}
                        axisLine={false}
                        tickLine={false}
                        interval="preserveStartEnd"
                      />
                      <YAxis hide />
                      <ReferenceLine
                        x={result.initial_balance}
                        stroke="#f97316"
                        strokeDasharray="4 2"
                      />
                      <Tooltip
                        contentStyle={{
                          background: "#171717",
                          border: "1px solid #404040",
                          borderRadius: 8,
                          fontSize: 11,
                        }}
                        formatter={(v: number, _: string, p: any) => [
                          `${v} sims`,
                          `$${(p.payload.x as number).toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
                        ]}
                      />
                      <Area
                        type="monotone"
                        dataKey="count"
                        stroke="#f97316"
                        strokeWidth={1.5}
                        fill="#f97316"
                        fillOpacity={0.12}
                        dot={false}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
                <p className="text-[9px] text-neutral-600 mt-1 text-center">
                  Orange line = starting balance (${result.initial_balance.toLocaleString()})
                </p>
              </div>
            )}

            <div className="flex justify-between items-center pt-2">
              <button
                type="button"
                onClick={() => setResult(null)}
                className="text-xs text-neutral-500 hover:text-neutral-300 transition-colors"
              >
                ← Run again
              </button>
              <button
                type="button"
                onClick={onClose}
                className="rounded-full border border-neutral-700 px-4 py-1.5 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// ── Main Strategies page ───────────────────────────────────────────────────────

export const Strategies = () => {
  const { selectedPortfolio } = usePortfolio();

  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [editingStrategy, setEditingStrategy] = useState<Strategy | null>(null);

  // Selected strategy detail
  const [selectedStrategy, setSelectedStrategy] = useState<Strategy | null>(null);
  const [backtests, setBacktests] = useState<Backtest[]>([]);
  const [backtestsLoading, setBacktestsLoading] = useState(false);

  // Modals
  const [showBacktestModal, setShowBacktestModal] = useState(false);
  const [viewingBacktest, setViewingBacktest] = useState<Backtest | null>(null);
  const [mcBacktest, setMcBacktest] = useState<Backtest | null>(null);

  const loadStrategies = useCallback(async () => {
    if (!selectedPortfolio) {
      setStrategies([]);
      return;
    }
    setLoading(true);
    try {
      const res = await api.get("/v1/strategies", {
        params: { portfolio_uuid: selectedPortfolio.uuid },
      });
      setStrategies(res.data.strategies ?? []);
    } catch {
      setStrategies([]);
    } finally {
      setLoading(false);
    }
  }, [selectedPortfolio?.uuid]);

  useEffect(() => {
    void loadStrategies();
  }, [loadStrategies]);

  const loadBacktests = useCallback(async (strategyUUID: string) => {
    setBacktestsLoading(true);
    try {
      const res = await api.get("/v1/strategy/backtests", {
        params: { strategy_uuid: strategyUUID },
      });
      setBacktests(res.data.backtests ?? []);
    } catch {
      setBacktests([]);
    } finally {
      setBacktestsLoading(false);
    }
  }, []);

  const handleSelectStrategy = (s: Strategy) => {
    setSelectedStrategy(s);
    void loadBacktests(s.uuid);
  };

  const handleDelete = async (uuid: string) => {
    if (!window.confirm("Delete this strategy and all its backtests?")) return;
    try {
      await api.delete("/v1/strategy", { data: { uuid } });
      if (selectedStrategy?.uuid === uuid) {
        setSelectedStrategy(null);
        setBacktests([]);
      }
      void loadStrategies();
    } catch {
      // ignore
    }
  };

  const handleFormSave = (s: Strategy) => {
    setShowForm(false);
    setEditingStrategy(null);
    void loadStrategies();
    handleSelectStrategy(s);
  };

  const handleBacktestComplete = (bt: Backtest) => {
    setShowBacktestModal(false);
    setBacktests((b) => [bt, ...b]);
    setViewingBacktest(bt);
  };

  if (!selectedPortfolio) {
    return (
      <div className="flex flex-col max-w-6xl w-full mx-auto h-full bg-black text-white overflow-y-auto px-8 py-8">
        <p className="text-sm text-neutral-500">
          Select a portfolio to view strategies.
        </p>
      </div>
    );
  }

  return (
    <div className="flex max-w-6xl w-full mx-auto h-full bg-black text-white overflow-hidden">
      {/* ── Left: strategy list ── */}
      <div className="w-72 min-w-72 border-r border-neutral-800 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-4 py-4 border-b border-neutral-800">
          <h1 className="text-sm font-semibold text-white">Strategies</h1>
          <button
            type="button"
            onClick={() => {
              setEditingStrategy(null);
              setShowForm(true);
            }}
            className="rounded-full bg-orange-500 text-black text-[11px] font-semibold px-3 py-1.5 hover:bg-orange-400 transition-colors"
          >
            + New
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-2 py-2 space-y-1">
          {loading && (
            <p className="text-xs text-neutral-500 px-2 pt-2">Loading…</p>
          )}
          {!loading && strategies.length === 0 && (
            <div className="px-2 pt-4 text-center">
              <p className="text-xs text-neutral-500">
                No strategies yet. Create one to get started.
              </p>
            </div>
          )}
          {strategies.map((s) => (
            <button
              key={s.uuid}
              type="button"
              onClick={() => handleSelectStrategy(s)}
              className={`w-full text-left rounded-lg px-3 py-2.5 transition-colors ${
                selectedStrategy?.uuid === s.uuid
                  ? "bg-orange-500/10 border border-orange-500/30"
                  : "hover:bg-neutral-900/60 border border-transparent"
              }`}
            >
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-white">{s.name}</span>
                <span className="rounded-full bg-neutral-800 text-neutral-300 text-[10px] px-1.5 py-0.5">
                  {s.ticker}
                </span>
              </div>
              <p className="text-[11px] text-neutral-500 mt-0.5">
                {s.buy_rules.length} buy rule{s.buy_rules.length !== 1 ? "s" : ""}
                {s.sell_conditions.length > 0
                  ? ` · ${s.sell_conditions.length} exit condition${s.sell_conditions.length !== 1 ? "s" : ""}`
                  : " · hold to end"}
              </p>
            </button>
          ))}
        </div>
      </div>

      {/* ── Right: strategy detail ── */}
      <div className="flex-1 overflow-y-auto px-6 py-6">
        {!selectedStrategy ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-neutral-600 text-sm">
              Select a strategy to view details and run backtests.
            </p>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <h2 className="text-lg font-semibold text-white">
                    {selectedStrategy.name}
                  </h2>
                  <span className="rounded-full bg-neutral-800 text-neutral-300 text-xs px-2 py-0.5">
                    {selectedStrategy.ticker}
                  </span>
                </div>
                {selectedStrategy.description && (
                  <p className="text-sm text-neutral-400 mt-1">
                    {selectedStrategy.description}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setEditingStrategy(selectedStrategy);
                    setShowForm(true);
                  }}
                  className="rounded-full border border-neutral-700 px-3 py-1.5 text-xs text-neutral-300 hover:border-orange-500 hover:text-orange-300 transition-colors"
                >
                  Edit
                </button>
                <button
                  type="button"
                  onClick={() => handleDelete(selectedStrategy.uuid)}
                  className="rounded-full border border-neutral-800 px-3 py-1.5 text-xs text-neutral-500 hover:border-red-500 hover:text-red-400 transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>

            {/* Buy rules */}
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-widest text-orange-500 mb-2">
                Buy conditions (all required)
              </h3>
              <div className="space-y-1.5">
                {selectedStrategy.buy_rules.map((r, i) => (
                  <div
                    key={i}
                    className="rounded-lg border border-neutral-800 bg-neutral-900/30 px-3 py-2 text-xs text-neutral-300"
                  >
                    <span className="text-neutral-600 mr-1.5">
                      {i + 1}.
                    </span>
                    {ruleDesc(r)}
                  </div>
                ))}
              </div>
            </div>

            {/* Sell conditions */}
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-widest text-green-500 mb-2">
                Exit conditions
              </h3>
              {selectedStrategy.sell_conditions.length === 0 ? (
                <p className="text-xs text-neutral-600">
                  None — holds position through backtest period.
                </p>
              ) : (
                <div className="space-y-1.5">
                  {selectedStrategy.sell_conditions.map((sc, i) => (
                    <div
                      key={i}
                      className="rounded-lg border border-neutral-800 bg-neutral-900/30 px-3 py-2 text-xs text-neutral-300"
                    >
                      <span className="text-neutral-600 mr-1.5">
                        {i + 1}.
                      </span>
                      {sellCondDesc(sc)}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Backtest section */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-xs font-semibold uppercase tracking-widest text-neutral-500">
                  Backtests
                </h3>
                <button
                  type="button"
                  onClick={() => setShowBacktestModal(true)}
                  className="rounded-full bg-orange-500 text-black text-[11px] font-semibold px-3 py-1.5 hover:bg-orange-400 transition-colors"
                >
                  Run backtest
                </button>
              </div>

              {backtestsLoading && (
                <p className="text-xs text-neutral-500">Loading backtests…</p>
              )}
              {!backtestsLoading && backtests.length === 0 && (
                <p className="text-xs text-neutral-600">
                  No backtests yet. Run your first one above.
                </p>
              )}
              <div className="space-y-2">
                {backtests.map((bt) => {
                  const roiPos = bt.roi >= 0;
                  return (
                    <div
                      key={bt.uuid}
                      className="rounded-lg border border-neutral-800 bg-neutral-900/30 px-4 py-3 flex items-center justify-between"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-white">
                            {bt.ticker}
                          </span>
                          <span className="text-[11px] text-neutral-500">
                            {bt.from_date} → {bt.to_date}
                          </span>
                        </div>
                        <div className="flex items-center gap-3 mt-0.5">
                          <span className="text-[11px] text-neutral-400">
                            ${bt.initial_balance.toLocaleString()} →{" "}
                            ${bt.final_balance.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                          </span>
                          <span
                            className={`text-[11px] font-semibold ${roiPos ? "text-green-400" : "text-red-400"}`}
                          >
                            {roiPos ? "+" : ""}
                            {bt.roi.toFixed(2)}%
                          </span>
                          <span className="text-[11px] text-neutral-600">
                            {bt.total_trades} trades
                          </span>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => setViewingBacktest(bt)}
                          className="text-[11px] text-neutral-500 hover:text-neutral-200 transition-colors"
                        >
                          View
                        </button>
                        <button
                          type="button"
                          onClick={() => setMcBacktest(bt)}
                          className="text-[11px] text-neutral-500 hover:text-orange-300 transition-colors"
                        >
                          Monte Carlo
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ── Modals ── */}
      {showForm && (
        <StrategyForm
          initial={editingStrategy ?? undefined}
          portfolioUUID={selectedPortfolio.uuid}
          onSave={handleFormSave}
          onCancel={() => {
            setShowForm(false);
            setEditingStrategy(null);
          }}
        />
      )}

      {showBacktestModal && selectedStrategy && (
        <BacktestModal
          strategy={selectedStrategy}
          onClose={() => setShowBacktestModal(false)}
          onComplete={handleBacktestComplete}
        />
      )}

      {viewingBacktest && (
        <BacktestResultsModal
          backtest={viewingBacktest}
          onClose={() => setViewingBacktest(null)}
          onMonteCarlo={(bt) => {
            setViewingBacktest(null);
            setMcBacktest(bt);
          }}
        />
      )}

      {mcBacktest && (
        <MonteCarloModal
          backtest={mcBacktest}
          onClose={() => setMcBacktest(null)}
        />
      )}
    </div>
  );
};
