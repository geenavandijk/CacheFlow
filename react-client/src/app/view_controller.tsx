import { useLocation } from "react-router-dom";
import AppNavbar from "./components/navbar";
import { Dashboard } from "./dashboard";
import { useEffect } from "react";

const AppViewController = () => {

  const location = useLocation();

  const currentTab = window.location.pathname.split("/")[1];

  useEffect(() => {
      const pageNames: Record<string, string> = {
          dashboard: 'Dashboard',
      };
      const pageName = pageNames[currentTab] ?? 'Dashboard';
      document.title = `CacheFlow | ${pageName}`;
  }, [currentTab]);

  let component = null;

  switch(currentTab) {
      case "dashboard":
          component = <Dashboard />;
          break;
      default:
          component = <Dashboard />;
          break;
  }

  return (
    <div className="flex flex-col w-full h-screen bg-black">
      <AppNavbar />
      {component}
    </div>
  );
};

export default AppViewController;