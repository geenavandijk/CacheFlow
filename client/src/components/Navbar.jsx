import { Link, useNavigate } from "react-router-dom";
import Logo from "./Logo";

export default function Navbar() {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem("x-cf-bearer");
    localStorage.removeItem("x-cf-refresh");
    localStorage.removeItem("x-cf-uid");

    window.dispatchEvent(new Event("authChange"));
    navigate("/");
  };

  return (
    <nav
      style={{
        padding: "1rem 2rem",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <Logo size="small" />

      <div style={{ display: "flex", gap: "0.75rem", alignItems: "center" }}>
        <Link to="/profile">Profile</Link>
        <Link to="/search">Search</Link>
        <Link to="/stock">Stock</Link>
        <Link to="/portfolio">Portfolio</Link>

        {/* ✅ NEW */}
        <Link to="/api">API</Link>

        <button
          onClick={handleLogout}
          className="secondary"
          style={{ marginLeft: "1rem", padding: "8px 16px" }}
        >
          Logout
        </button>
      </div>
    </nav>
  );
}
