export default function Toast({ toast, onClose }) {
  if (!toast) return null;

  return (
    <div
      style={{
        position: "fixed",
        right: 16,
        bottom: 16,
        padding: "12px 14px",
        borderRadius: 10,
        background: "#111",
        color: "white",
        maxWidth: 360,
        boxShadow: "0 10px 30px rgba(0,0,0,0.25)",
        zIndex: 9999,
      }}
      role="status"
    >
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
        <div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>
            {toast.title || "Notice"}
          </div>
          <div style={{ opacity: 0.9 }}>{toast.message}</div>
        </div>
        <button
          onClick={onClose}
          style={{
            background: "transparent",
            border: 0,
            color: "white",
            fontSize: 16,
            cursor: "pointer",
          }}
          aria-label="Close"
        >
          ×
        </button>
      </div>
    </div>
  );
}
