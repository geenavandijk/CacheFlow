import { useState, type FormEvent } from "react";

type Ticker = {
  ticker: string;
  name: string;
};

export default function SearchPage() {
  const [q, setQ] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<Ticker[]>([]);
  const [error, setError] = useState<string | null>(null);

  const base = (import.meta.env.VITE_API_BASE_URL || "http://localhost:8080").replace(/\/+$/, "");

  async function runSearch(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const query = q.trim();

    if (!query) {
      setResults([]);
      setError(null);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${base}/v1/companies/search?q=${encodeURIComponent(query)}`);
      const body = await res.json();
      setResults(Array.isArray(body) ? body : []);
    } catch {
      setError("Search failed");
      setResults([]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full min-h-screen bg-black text-white pt-28 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Company Search</h1>

        <form onSubmit={runSearch} className="flex gap-3 mb-6">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search ticker or company name. Example: AAPL"
            className="flex-1 rounded-lg px-4 py-3 text-black"
          />
          <button
            type="submit"
            disabled={loading}
            className="bg-orange-500 hover:bg-orange-600 px-5 py-3 rounded-lg font-medium"
          >
            {loading ? "Searching..." : "Search"}
          </button>
        </form>

        {error && <div className="text-red-400 mb-4">{error}</div>}

        <div className="space-y-3">
          {results.length === 0 && !loading && !error ? (
            <div className="text-white/70">No results yet. Try AAPL, MSFT, GOOGL.</div>
          ) : null}

          {results.map((r) => (
            <div
              key={r.ticker}
              className="border border-white/20 rounded-xl p-4 flex items-center justify-between bg-white/5"
            >
              <div>
                <div className="font-bold text-lg">{r.ticker}</div>
                <div className="text-white/80">{r.name}</div>
              </div>

              <button
                onClick={() => navigator.clipboard.writeText(r.ticker)}
                className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg"
              >
                Copy
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}