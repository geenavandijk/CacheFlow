import { useState, useEffect } from "react";
import { useAuth } from "../provider/auth";
import api from "../util/api";
import type { RiskSettings } from "../util/account_data";
import { Slider } from "./components/slider";

const inputClass =
  "w-full h-10 text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none";

export const Settings = () => {
  const { accountData, callLoadInAccount } = useAuth();
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [budget, setBudget] = useState("");
  const [maxLossPercentage, setMaxLossPercentage] = useState("");
  const [riskTolerance, setRiskTolerance] = useState(5);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    if (!accountData) return;
    setFirstName(accountData.first_name ?? "");
    setLastName(accountData.last_name ?? "");
    const rs = accountData.risk_settings;
    if (rs?.budget != null) setBudget(String(rs.budget));
    if (rs?.max_loss_percentage != null) setMaxLossPercentage(String(rs.max_loss_percentage));
    if (rs?.risk_tolerance != null) setRiskTolerance(Math.min(10, Math.max(1, rs.risk_tolerance)));
  }, [accountData]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);
    setSaving(true);
    try {
      const payload: {
        first_name?: string;
        last_name?: string;
        risk_settings?: Partial<RiskSettings>;
      } = {};
      if (firstName.trim() !== (accountData?.first_name ?? "")) payload.first_name = firstName.trim();
      if (lastName.trim() !== (accountData?.last_name ?? "")) payload.last_name = lastName.trim();
      const budgetNum = budget.trim() === "" ? undefined : parseInt(budget, 10);
      const maxLossNum = maxLossPercentage.trim() === "" ? undefined : parseInt(maxLossPercentage, 10);
      payload.risk_settings = {
        risk_tolerance: riskTolerance,
      };
      if (budgetNum !== undefined && !Number.isNaN(budgetNum) && budgetNum >= 0) payload.risk_settings.budget = budgetNum;
      if (maxLossNum !== undefined && !Number.isNaN(maxLossNum)) payload.risk_settings.max_loss_percentage = Math.min(100, Math.max(0, maxLossNum));
      await api.patch("/v1/account/settings", payload);
      await callLoadInAccount();
      setMessage({ type: "success", text: "Settings saved." });
    } catch (err: unknown) {
      const msg = err && typeof err === "object" && "response" in err && err.response && typeof err.response === "object" && "data" in err.response && err.response.data && typeof err.response.data === "object" && "error" in err.response.data
        ? String((err.response.data as { error: unknown }).error)
        : "Failed to save settings.";
      setMessage({ type: "error", text: msg });
    } finally {
      setSaving(false);
    }
  };

  if (!accountData) {
    return (
      <div className="flex items-center justify-center w-full h-full bg-black text-white">
        <p>Loading account…</p>
      </div>
    );
  }

  const riskLabel = riskTolerance === 1 ? "Lowest" : riskTolerance === 10 ? "Highest" : "Medium";

  return (
    <div className="flex flex-col w-full h-full bg-black text-white px-8 py-8">
      <h1 className="text-2xl font-semibold mb-2">Settings</h1>
      <p className="text-sm text-neutral-400 mb-6">
        Update your profile and risk preferences.
      </p>

      <form onSubmit={handleSubmit} className="max-w-xl space-y-6">
        <div className="space-y-4 p-4 border border-neutral-800 rounded-lg bg-neutral-900/50">
          <h2 className="text-xs font-medium uppercase tracking-[0.16em] text-neutral-500">
            Profile
          </h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <label className="text-white text-sm font-regular">First name</label>
              <input
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                className={inputClass}
              />
            </div>
            <div>
              <label className="text-white text-sm font-regular">Last name</label>
              <input
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        <div className="space-y-4 p-4 border border-neutral-800 rounded-lg bg-neutral-900/50">
          <h2 className="text-xs font-medium uppercase tracking-[0.16em] text-neutral-500">
            Risk settings
          </h2>
          <div className="space-y-4">
            <div>
              <label className="text-white text-sm font-regular">Budget</label>
              <input
                type="number"
                placeholder="e.g. 1000000"
                value={budget}
                onChange={(e) => setBudget(e.target.value)}
                min={0}
                className={inputClass}
              />
            </div>
            <div>
              <label className="text-white text-sm font-regular">Max loss (%)</label>
              <input
                type="number"
                placeholder="0–100"
                value={maxLossPercentage}
                onChange={(e) => setMaxLossPercentage(e.target.value)}
                min={0}
                max={100}
                className={inputClass}
              />
            </div>
            <div className="mt-4">
              <Slider
                label="Risk tolerance"
                min={1}
                max={10}
                step={1}
                value={riskTolerance}
                onChange={setRiskTolerance}
                valueLabel={`${riskTolerance} — ${riskLabel} risk`}
              />
            </div>
          </div>
        </div>

        {message && (
          <p className={message.type === "success" ? "text-green-400 text-sm" : "text-red-400 text-sm"}>
            {message.text}
          </p>
        )}

        <button
          type="submit"
          disabled={saving}
          className="rounded-full hover:cursor-pointer bg-orange-500 text-white font-medium px-6 py-4 transition-all duration-300 hover:bg-orange-600 disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {saving ? "Saving…" : "Save settings"}
        </button>
      </form>
    </div>
  );
};
