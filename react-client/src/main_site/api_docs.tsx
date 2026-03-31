export default function ApiDocsPage() {
  const rows = [
    {
      group: "Search",
      method: "GET",
      path: "/v1/companies/search?q=QUERY",
      returns: "Array of ticker results",
      notes: "Searches company tickers in MongoDB. Max 10 results.",
    },
    {
      group: "Stocks",
      method: "GET",
      path: "/v1/datafeed/stock?ticker=TSLA",
      returns: "Ticker overview",
      notes: "Gets stock overview data.",
    },
    {
      group: "Stocks",
      method: "GET",
      path: "/v1/datafeed/stock/aggregates",
      returns: "Aggregates data",
      notes: "Gets stock aggregate data.",
    },
    {
      group: "Auth",
      method: "POST",
      path: "/oauth2/token",
      returns: "Token response",
      notes: "OAuth token endpoint.",
    },
    {
      group: "Portfolio",
      method: "GET",
      path: "/v1/portfolios",
      returns: "Portfolio list",
      notes: "Gets user portfolios.",
    },
  ];

  return (
    <div className="w-full min-h-screen bg-black text-white pt-28 px-4">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold mb-2">API Documentation</h1>
        <p className="text-white/70 mb-6">Quick reference for current endpoints.</p>

        <div className="overflow-x-auto rounded-xl border border-white/20 bg-white/5">
          <table className="w-full min-w-[900px]">
            <thead>
              <tr className="border-b border-white/20 bg-white/10">
                <th className="text-left px-4 py-3">Group</th>
                <th className="text-left px-4 py-3">Method</th>
                <th className="text-left px-4 py-3">Path</th>
                <th className="text-left px-4 py-3">Returns</th>
                <th className="text-left px-4 py-3">Notes</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row, i) => (
                <tr key={i} className="border-b border-white/10">
                  <td className="px-4 py-3">{row.group}</td>
                  <td className="px-4 py-3">{row.method}</td>
                  <td className="px-4 py-3">{row.path}</td>
                  <td className="px-4 py-3">{row.returns}</td>
                  <td className="px-4 py-3">{row.notes}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}