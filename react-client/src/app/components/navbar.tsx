import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../provider/auth";
import Search from "./search";
import { usePortfolio } from "../portfolio_context";
import { useState } from "react";

const AppNavbar = () => {
  const { accountData, setIsAuthenticated } = useAuth();
  const { portfolios, selectedPortfolio, setSelectedPortfolioId, createPortfolio, updatePortfolio, deletePortfolio } =
    usePortfolio();
  const navigate = useNavigate();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit" | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [startingBalance, setStartingBalance] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogout = () => {
    localStorage.removeItem("x-cf-uid");
    localStorage.removeItem("x-cf-bearer");
    localStorage.removeItem("x-cf-refresh");
    setIsAuthenticated(false);
    navigate("/login");
  };

  const openCreate = () => {
    setFormMode("create");
    setEditingId(null);
    setName("");
    setDescription("");
    setStartingBalance("");
    setError(null);
    setIsModalOpen(true);
  };

  const openEdit = (id: string) => {
    const p = portfolios.find((x) => x.uuid === id);
    if (!p) return;
    setFormMode("edit");
    setEditingId(id);
    setName(p.name ?? "");
    setDescription(p.description ?? "");
    setStartingBalance(p.starting_balance != null ? String(p.starting_balance) : "");
    setError(null);
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setFormMode(null);
    setEditingId(null);
    setError(null);
  };

  const handleSelect = (id: string) => {
    setSelectedPortfolioId(id);
    closeModal();
  };

  const handleDelete = async (id: string) => {
    const p = portfolios.find((x) => x.uuid === id);
    if (!p) return;
    if (!window.confirm(`Delete portfolio "${p.name}"? This cannot be undone.`)) return;
    try {
      await deletePortfolio(id);
      if (selectedPortfolio && selectedPortfolio.uuid === id) {
        setSelectedPortfolioId(null);
      }
    } catch (err) {
      console.error("Failed to delete portfolio", err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim() || !description.trim() || (formMode === "create" && !startingBalance.trim())) {
      setError("All fields are required.");
      return;
    }

    setSaving(true);
    try {
      if (formMode === "create") {
        const start = parseFloat(startingBalance);
        if (Number.isNaN(start) || start <= 0) {
          setError("Starting balance must be a positive number.");
          setSaving(false);
          return;
        }
        await createPortfolio({
          name: name.trim(),
          description: description.trim(),
          starting_balance: start,
        });
      } else if (formMode === "edit" && editingId) {
        const p = portfolios.find((x) => x.uuid === editingId);
        const canEditBalance = p && p.starting_balance === p.current_balance;
        const payload: { uuid: string; name: string; description: string; starting_balance?: number } = {
          uuid: editingId,
          name: name.trim(),
          description: description.trim(),
        };
        if (canEditBalance && startingBalance.trim()) {
          const start = parseFloat(startingBalance);
          if (Number.isNaN(start) || start <= 0) {
            setError("Starting balance must be a positive number.");
            setSaving(false);
            return;
          }
          payload.starting_balance = start;
        }
        await updatePortfolio(payload);
      }
      closeModal();
    } catch (err) {
      console.error("Failed to save portfolio", err);
      setError("Failed to save portfolio.");
    } finally {
      setSaving(false);
    }
  };

  const portfolioButtonLabel = (() => {
    if (selectedPortfolio) return selectedPortfolio.name || "Portfolio";
    if (portfolios.length === 0) return "Create portfolio";
    return "Select portfolio";
  })();

  return (
    <>
      <nav className="flex items-center justify-between px-6 py-3 border-b border-neutral-800 bg-black/80 backdrop-blur-sm">
        <div className="flex items-center gap-2">
          <span className="h-7 w-7 rounded-full bg-orange-500 flex items-center justify-center text-xs font-bold text-black">
            CF
          </span>
          <span className="text-sm font-semibold text-white tracking-wide">CacheFlow</span>

          <button
            type="button"
            onClick={() => {
              if (portfolios.length === 0) {
                openCreate();
              } else {
                setFormMode(null);
                setIsModalOpen(true);
              }
            }}
            className="ml-3 rounded-full border border-neutral-800 px-3 py-1 text-[10px] uppercase tracking-[0.18em] text-neutral-300 hover:border-orange-500 hover:text-orange-300 transition-colors"
          >
            {portfolioButtonLabel}
          </button>
        </div>

        <div className="flex items-center mx-4">
          <Search />
        </div>

        <div className="flex items-center gap-4">
          <Link
            to="/client-app/dashboard"
            className="text-xs text-neutral-300 hover:text-white transition-colors"
          >
            Dashboard
          </Link>
          <Link
            to="/client-app/settings"
            className="text-xs text-neutral-300 hover:text-white transition-colors"
          >
            Settings
          </Link>

          <div className="flex items-center gap-2 text-xs text-neutral-400">
            {accountData && (
              <span className="hidden sm:inline">
                {accountData.first_name} {accountData.last_name}
              </span>
            )}
            <span className="h-6 w-px bg-neutral-800" />
            <button
              onClick={handleLogout}
              className="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-200 hover:border-orange-500 hover:text-orange-400 transition-colors"
            >
              Log out
            </button>
          </div>
        </div>
      </nav>

      {isModalOpen && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/70">
          <div className="w-full max-w-2xl rounded-2xl border border-neutral-800 bg-neutral-950/95 p-6 shadow-2xl">
            {formMode === null && (
              <>
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-sm font-semibold text-white tracking-wide">
                    Portfolios
                  </h2>
                  <button
                    onClick={closeModal}
                    className="text-neutral-500 hover:text-neutral-200 text-sm"
                  >
                    Close
                  </button>
                </div>

                <div className="flex items-center justify-between mb-3">
                  <button
                    type="button"
                    onClick={openCreate}
                    className="rounded-full border border-neutral-700 px-3 py-1.5 text-xs text-neutral-200 hover:border-orange-500 hover:text-orange-300 transition-colors"
                  >
                    + Create portfolio
                  </button>
                </div>

                <div className="flex flex-col gap-4 max-h-[360px] overflow-y-auto">
                  {portfolios.length === 0 && (
                    <p className="text-sm text-neutral-500">
                      You don&apos;t have any portfolios yet. Create one to get started.
                    </p>
                  )}

                  {portfolios.map((p) => {
                    const canEditBalance = p.starting_balance === p.current_balance;
                    return (
                      <div
                        key={p.uuid}
                        className="flex items-center justify-between rounded-xl border border-neutral-800 px-4 py-3 bg-neutral-900/40"
                      >
                        <button
                          type="button"
                          onClick={() => handleSelect(p.uuid)}
                          className="flex flex-col items-start text-left flex-1 mr-3"
                        >
                          <span className="text-sm font-medium text-white">
                            {p.name}
                          </span>
                          <span className="text-xs text-neutral-400 line-clamp-2">
                            {p.description}
                          </span>
                          <span className="text-[11px] text-neutral-500 mt-1">
                            Balance: ${p.current_balance.toFixed(2)}{" "}
                            <span className="text-neutral-600">
                              (start ${p.starting_balance.toFixed(2)})
                            </span>
                          </span>
                        </button>

                        <div className="flex flex-col items-end gap-2 text-xs">
                          <button
                            type="button"
                            onClick={() => openEdit(p.uuid)}
                            className="rounded-full border border-neutral-700 px-3 py-1 text-neutral-200 hover:border-orange-500 hover:text-orange-300 transition-colors"
                          >
                            Edit
                          </button>
                          <button
                            type="button"
                            onClick={() => handleDelete(p.uuid)}
                            className="rounded-full border border-neutral-800 px-3 py-1 text-neutral-500 hover:border-red-500 hover:text-red-400 transition-colors"
                          >
                            Delete
                          </button>
                          {!canEditBalance && (
                            <span className="text-[10px] text-neutral-500 max-w-[140px] text-right">
                              Balance has changed; starting balance is locked.
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </>
            )}

            {formMode !== null && (
              <>
                <div className="flex items-center justify-between mb-4">
                  <button
                    type="button"
                    onClick={() => {
                      setFormMode(null);
                      setError(null);
                    }}
                    className="text-xs text-neutral-400 hover:text-neutral-200"
                  >
                    ← Back
                  </button>
                  <button
                    onClick={closeModal}
                    className="text-neutral-500 hover:text-neutral-200 text-sm"
                  >
                    Close
                  </button>
                </div>
                <h3 className="text-sm font-semibold text-white mb-4">
                  {formMode === "edit" ? "Edit portfolio" : "Create portfolio"}
                </h3>

                <form onSubmit={handleSubmit} className="space-y-3">
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
                  {formMode === "create" && (
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
                  )}
                  {formMode === "edit" && editingId && (
                    <div>
                      <label className="text-white text-sm font-regular">
                        Starting balance
                      </label>
                      {(() => {
                        const p = portfolios.find((x) => x.uuid === editingId);
                        const canEditBalance = p && p.starting_balance === p.current_balance;
                        return (
                          <>
                            <input
                              type="number"
                              min="0"
                              step="0.01"
                              value={startingBalance}
                              onChange={(e) => setStartingBalance(e.target.value)}
                              disabled={!canEditBalance}
                              className={`w-full h-9 text-white mt-1 rounded-lg border px-2 bg-transparent active:ring-0 focus:ring-0 focus:outline-none text-sm ${
                                canEditBalance
                                  ? "border-neutral-800 focus:border-white focus:border-2 transition-all duration-300"
                                  : "border-neutral-800 opacity-60 cursor-not-allowed"
                              }`}
                            />
                            {!canEditBalance && (
                              <p className="text-[11px] text-neutral-500 mt-1">
                                Balance has changed; starting balance is locked.
                              </p>
                            )}
                          </>
                        );
                      })()}
                    </div>
                  )}

                  {error && (
                    <p className="text-xs text-red-400 mt-1">
                      {error}
                    </p>
                  )}

                  <div className="flex justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={closeModal}
                      className="rounded-full border border-neutral-700 px-4 py-2 text-xs text-neutral-300 hover:border-neutral-500 hover:text-white transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={saving}
                      className="rounded-full bg-orange-500 text-white text-xs font-medium px-5 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                      {saving ? "Saving…" : formMode === "edit" ? "Save changes" : "Create portfolio"}
                    </button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
};

export default AppNavbar;

