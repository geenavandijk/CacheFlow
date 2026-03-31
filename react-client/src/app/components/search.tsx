import { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import { useNavigate } from "react-router-dom";
import { Star } from "lucide-react";
import { useWatchlists } from "../watchlists_context";

interface Company {
    name: string;
    ticker: string;
}

export default function Search() {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<Company[]>([]);
    const [starTicker, setStarTicker] = useState<string | null>(null);
    const [starMode, setStarMode] = useState<"select" | "create">("select");
    const [newWatchlistName, setNewWatchlistName] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [saving, setSaving] = useState(false);
    const navigate = useNavigate();
    const {
        watchlists,
        createWatchlist,
        updateWatchlist,
        refreshWatchlists,
    } = useWatchlists();

    useEffect(() => {
        if (!query) return setResults([]);

        const timeout = setTimeout(async () => {
            const res = await fetch(`/v1/companies/search?q=${encodeURIComponent(query)}`);
            const data = await res.json();
            setResults(data);
        }, 300); // debounce

        return () => clearTimeout(timeout);
    }, [query]);

    const handleSelect = (ticker: string) => {
        setQuery('');
        navigate(`/client-app/stock?ticker=${ticker}`);
    };

    const tickerInAnyWatchlist = (ticker: string) => {
        const t = ticker.toUpperCase();
        return watchlists.some(w => w.tickers.includes(t));
    };

    const toggleTickerInWatchlist = async (watchlistId: string, ticker: string) => {
        const t = ticker.toUpperCase();
        const wl = watchlists.find(w => w.uuid === watchlistId);
        if (!wl) return;
        const has = wl.tickers.includes(t);
        const nextTickers = has
            ? wl.tickers.filter(x => x !== t)
            : [...wl.tickers, t];
        await updateWatchlist({ uuid: watchlistId, tickers: nextTickers });
        await refreshWatchlists();
    };

    const handleCreateWatchlistWithTicker = async () => {
        if (!starTicker) return;
        if (!newWatchlistName.trim()) {
            setError("Name is required.");
            return;
        }
        setSaving(true);
        setError(null);
        try {
            await createWatchlist({
                name: newWatchlistName.trim(),
                tickers: [starTicker.toUpperCase()],
            });
            await refreshWatchlists();
            setStarTicker(null);
            setNewWatchlistName("");
        } catch (err) {
            console.error("Failed to create watchlist from search", err);
            setError("Failed to create watchlist.");
        } finally {
            setSaving(false);
        }
    };

    return (
        <>
            <div className="relative w-96">
                <input
                    className="border-b border-neutral-400 w-96 bg-transparent px-1 py-0.5 text-[10px] uppercase tracking-[0.18em] text-neutral-400 outline-none placeholder:text-neutral-400"
                    type="text"
                    placeholder="SEARCH"
                    value={query}
                    onChange={e => setQuery(e.target.value)}
                />
                {query && (
                    <ul className="absolute top-full left-0 mt-1 w-72 bg-black border border-neutral-900 rounded-xl shadow-md z-50">
                        {results.map(company => {
                            const inWatchlist = tickerInAnyWatchlist(company.ticker);
                            return (
                                <li
                                    key={company.ticker}
                                    className="px-3 py-1.5 text-[10px] uppercase tracking-[0.18em] text-neutral-400 hover:bg-neutral-800 flex items-center justify-between gap-2"
                                >
                                    <button
                                        type="button"
                                        onClick={() => handleSelect(company.ticker)}
                                        className="flex-1 text-left"
                                    >
                                        {company.ticker} — {company.name}
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setStarTicker(company.ticker.toUpperCase());
                                            setStarMode("select");
                                            setNewWatchlistName("");
                                            setError(null);
                                        }}
                                        className="p-1 text-yellow-400 hover:text-yellow-300"
                                    >
                                        <Star
                                            size={14}
                                            fill={inWatchlist ? "#facc15" : "none"}
                                            strokeWidth={1.5}
                                        />
                                    </button>
                                </li>
                            );
                        })}
                    </ul>
                )}
            </div>

            {starTicker &&
                createPortal(
                    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black">
                        <div className="w-full max-w-sm rounded-2xl border border-neutral-800 bg-neutral-950 p-5 shadow-2xl">
                            <div className="flex items-center justify-between mb-3">
                                <h2 className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">
                                    {starMode === "select" ? "Add to watchlist" : "Create watchlist"}
                                </h2>
                                <button
                                    type="button"
                                    onClick={() => setStarTicker(null)}
                                    className="text-[11px] text-neutral-500 hover:text-neutral-200"
                                >
                                    Close
                                </button>
                            </div>

                            <p className="text-xs text-neutral-300 mb-3">
                                {starTicker}
                            </p>

                            {starMode === "select" ? (
                                <>
                                    {watchlists.length > 0 ? (
                                        <div className="mb-3 space-y-1">
                                            {watchlists.map(wl => {
                                                const has = wl.tickers.includes(starTicker);
                                                return (
                                                    <button
                                                        key={wl.uuid}
                                                        type="button"
                                                        onClick={async () => {
                                                            await toggleTickerInWatchlist(wl.uuid, starTicker);
                                                        }}
                                                        className="w-full flex items-center justify-between rounded-lg border border-neutral-800 bg-neutral-900/60 px-3 py-2 text-[11px] text-neutral-200 hover:border-orange-500 transition-colors"
                                                    >
                                                        <span>{wl.name}</span>
                                                        <span className="text-neutral-500">
                                                            {has ? "Remove" : "Add"}
                                                        </span>
                                                    </button>
                                                );
                                            })}
                                        </div>
                                    ) : (
                                        <p className="text-[11px] text-neutral-500 mb-3">
                                            You do not have any watchlists yet.
                                        </p>
                                    )}
                                    <button
                                        type="button"
                                        className="mt-2 w-full rounded-full border border-neutral-700 px-4 py-2 text-[11px] text-neutral-200 hover:border-neutral-500 hover:text-white transition-colors"
                                        onClick={() => {
                                            setStarMode("create");
                                            setError(null);
                                        }}
                                    >
                                        Create new watchlist
                                    </button>
                                </>
                            ) : (
                                <div className="mt-1">
                                    <button
                                        type="button"
                                        className="mb-3 text-[11px] text-neutral-400 hover:text-neutral-200"
                                        onClick={() => {
                                            setStarMode("select");
                                            setError(null);
                                        }}
                                    >
                                        ← Back to watchlists
                                    </button>
                                    <input
                                        type="text"
                                        value={newWatchlistName}
                                        onChange={e => setNewWatchlistName(e.target.value)}
                                        placeholder="Watchlist name"
                                        className="w-full h-8 text-xs text-white rounded-lg border border-neutral-800 bg-transparent px-2 mb-2 focus:border-white focus:border-2 outline-none"
                                    />
                                    {error && (
                                        <p className="text-[10px] text-red-400 mb-1">
                                            {error}
                                        </p>
                                    )}
                                    <button
                                        type="button"
                                        disabled={saving}
                                        onClick={handleCreateWatchlistWithTicker}
                                        className="w-full rounded-full bg-orange-500 text-white text-[10px] font-medium px-3 py-2 hover:bg-orange-600 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                                    >
                                        {saving ? "Creating…" : "Create and add"}
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>,
                    document.body,
                )}
        </>
    );
}