import AppNavbar from "./components/navbar";
import { Dashboard } from "./dashboard";
import { Profile } from "./profile";
import { Settings } from "./settings";
import { Stock } from "./stock";
import { Strategies } from "./strategies";
import { useEffect } from "react";
import { PortfolioProvider } from "./portfolio_context";
import { PositionsProvider } from "./positions_context";
import { WatchlistsProvider } from "./watchlists_context";

const AppViewController = () => {
  const currentTab = window.location.pathname.split("/").pop() || "dashboard";

  useEffect(() => {
    const pageNames: Record<string, string> = {
      dashboard: "Dashboard",
      profile: "Profile",
      settings: "Settings",
      stock: "Stock",
      strategies: "Strategies",
    };
    const pageName = pageNames[currentTab] ?? "Dashboard";
    document.title = `CacheFlow | ${pageName}`;
  }, [currentTab]);

  let component = null;

  switch (currentTab) {
    case "dashboard":
      component = <Dashboard />;
      break;
    case "profile":
      component = <Profile />;
      break;
    case "settings":
      component = <Settings />;
      break;
    case "stock":
      component = <Stock />;
      break;
    case "strategies":
      component = <Strategies />;
      break;
    default:
      component = <Dashboard />;
      break;
  }

  return (
    <PortfolioProvider>
      <PositionsProvider>
        <WatchlistsProvider>
          <div className="flex flex-col w-full h-screen bg-black">
            <AppNavbar />
            {component}
          </div>
        </WatchlistsProvider>
      </PositionsProvider>
    </PortfolioProvider>
  );
};

export default AppViewController;
