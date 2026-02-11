// client/src/utils/api.jsx

export class ApiError extends Error {
  constructor(message, { status, code, fields, requestId } = {}) {
    super(message || "Request failed");
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
    this.requestId = requestId;
  }
}

export const apiRequest = async (endpoint, options = {}) => {
  // Allow passing either "/path" OR full "http://..."
  const base =
    (import.meta.env.VITE_API_BASE_URL || "http://localhost:8080").replace(/\/+$/, "");
  const url =
    endpoint.startsWith("http://") || endpoint.startsWith("https://")
      ? endpoint
      : `${base}${endpoint.startsWith("/") ? "" : "/"}${endpoint}`;

  const headers = {
    "Content-Type": "application/json",

    // keep your existing auth/device headers
    "x-cf-device-id": localStorage.getItem("x-cf-device-id") || "",
    "x-cf-uid": localStorage.getItem("x-cf-uid") || "",
    "x-cf-bearer": localStorage.getItem("x-cf-bearer") || "",
    "x-cf-refresh": localStorage.getItem("x-cf-refresh") || "",

    ...options.headers,
  };

  const res = await fetch(url, {
    ...options,
    headers,
  });

  // ✅ Detect your NEW standardized backend errors without consuming the body
  const contentType = res.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    try {
      const body = await res.clone().json();

      // standardized response: { ok: true, data } OR { ok:false, error }
      if (body && typeof body === "object" && "ok" in body && body.ok === false) {
        const err = body.error || {};
        throw new ApiError(err.message || "Request failed", {
          status: res.status,
          code: err.code,
          fields: err.fields,
          requestId: err.requestId,
        });
      }
    } catch (e) {
      // If it's already an ApiError, rethrow.
      if (e instanceof ApiError) throw e;
      // otherwise ignore parse errors here and fall through to res.ok check
    }
  }

  // non-standard errors (plain 500 text, etc.)
  if (!res.ok) {
    throw new ApiError(res.statusText || "Request failed", { status: res.status });
  }

  // ✅ Keep returning Response so your existing code can still do: await res.json()
  return res;
};
