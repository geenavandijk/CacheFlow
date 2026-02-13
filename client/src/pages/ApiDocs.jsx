const endpoints = [
  {
    group: "Auth",
    method: "POST",
    path: "/account/create",
    auth: "No",
    body: "{ email, password }",
    returns: "201 { ok:true, data:{ id, email } }",
    errors: "400 bad_request, 409 conflict",
  },
  {
    group: "Auth",
    method: "POST",
    path: "/auth/login",
    auth: "No",
    body: "{ email, password }",
    returns: "200 { ok:true, data:{ bearer, refresh, uid } }",
    errors: "400 bad_request, 401 unauthorized",
  },
  {
    group: "Auth",
    method: "POST",
    path: "/auth/logout",
    auth: "Yes",
    body: "{}",
    returns: "200 { ok:true, data:{ message } }",
    errors: "401 unauthorized",
  },
  {
    group: "Profile",
    method: "GET",
    path: "/profile",
    auth: "Yes",
    body: "—",
    returns: "200 { ok:true, data:{ profile } }",
    errors: "401 unauthorized, 404 not_found",
  },
  {
    group: "Profile",
    method: "PUT",
    path: "/profile",
    auth: "Yes",
    body: "{ riskTolerance, currentAge, ... }",
    returns: "200 { ok:true, data:{ profile } }",
    errors: "400 bad_request, 401 unauthorized",
  },
];

export default function ApiDocs() {
  return (
    <div>
      <h1 style={{ fontSize: 26, fontWeight: 800, marginBottom: 6 }}>
        CacheFlow API Endpoints
      </h1>
      <p style={{ marginBottom: 16, opacity: 0.85 }}>
        Auth + Profile endpoints. Responses use <code>{"{ ok, data }"}</code> or{" "}
        <code>{"{ ok:false, error }"}</code>.
      </p>

      <div style={{ overflowX: "auto", border: "1px solid #ddd", borderRadius: 10 }}>
        <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 900 }}>
          <thead>
            <tr style={{ background: "#f6f6f6" }}>
              {["Group", "Method", "Path", "Auth", "Body", "Returns", "Errors"].map(
                (h) => (
                  <th
                    key={h}
                    style={{
                      textAlign: "left",
                      padding: "10px 12px",
                      borderBottom: "1px solid #ddd",
                      fontSize: 13,
                    }}
                  >
                    {h}
                  </th>
                )
              )}
            </tr>
          </thead>
          <tbody>
            {endpoints.map((e, i) => (
              <tr key={i}>
                <td style={cell}>{e.group}</td>
                <td style={cell}>
                  <code>{e.method}</code>
                </td>
                <td style={cell}>
                  <code>{e.path}</code>
                </td>
                <td style={cell}>{e.auth}</td>
                <td style={cell}>
                  <code>{e.body}</code>
                </td>
                <td style={cell}>{e.returns}</td>
                <td style={cell}>{e.errors}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

const cell = {
  padding: "10px 12px",
  borderBottom: "1px solid #eee",
  verticalAlign: "top",
  fontSize: 13,
};
