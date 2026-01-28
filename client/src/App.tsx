import { useEffect, useState } from "react";
import { fetchHealth, type HealthResponse } from "./lib/api";

export default function App() {
  const [data, setData] = useState<HealthResponse | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    fetchHealth()
      .then(setData)
      .catch((e) => setErr(e?.message ?? "Unknown error"));
  }, []);

  return (
    <div style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
      <h1>CacheFlow</h1>
      <p style={{ opacity: 0.8 }}>Frontend ↔ Backend connectivity check</p>

      {!data && !err && <p>Loading…</p>}

      {err && (
        <div style={{ marginTop: 16 }}>
          <p style={{ fontWeight: 700 }}>Backend not reachable</p>
          <pre style={{ padding: 12, background: "#111", borderRadius: 8, overflowX: "auto" }}>
            {err}
          </pre>
          <p>Make sure the Go server is running on :8080.</p>
        </div>
      )}

      {data && (
        <div style={{ marginTop: 16 }}>
          <p style={{ fontWeight: 700 }}>Backend OK ✅</p>
          <pre style={{ padding: 12, background: "#111", borderRadius: 8, overflowX: "auto" }}>
            {JSON.stringify(data, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
