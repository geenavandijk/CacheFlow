import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../provider/auth";

const AppNavbar = () => {
  const { accountData, setIsAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem("x-cf-uid");
    localStorage.removeItem("x-cf-bearer");
    localStorage.removeItem("x-cf-refresh");
    setIsAuthenticated(false);
    navigate("/login");
  };

  return (
    <nav className="flex items-center justify-between px-6 py-3 border-b border-neutral-800 bg-black/80 backdrop-blur-sm">
      <div className="flex items-center gap-2">
        <span className="h-7 w-7 rounded-full bg-orange-500 flex items-center justify-center text-xs font-bold text-black">
          CF
        </span>
        <span className="text-sm font-semibold text-white tracking-wide">
          CacheFlow
        </span>
        <span className="ml-3 rounded-full border border-neutral-800 px-2 py-0.5 text-[10px] uppercase tracking-[0.18em] text-neutral-400">
          Client
        </span>
      </div>

      <div className="flex items-center gap-4">
        <Link
          to="/client-app"
          className="text-xs text-neutral-300 hover:text-white transition-colors"
        >
          Dashboard
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
  );
};

export default AppNavbar;
