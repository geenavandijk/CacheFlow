import { useLocation } from "react-router-dom";
import AppNavbar from "./components/navbar";
import { Dashboard } from "./dashboard";
import { Profile } from "./profile";
import { useEffect } from "react";

const AppViewController = () => {

  const location = useLocation();

  const currentTab = window.location.pathname.split("/").pop() || "dashboard";

  useEffect(() => {
      const pageNames: Record<string, string> = {
          dashboard: 'Dashboard',
          profile: 'Profile'
      };
      const pageName = pageNames[currentTab] ?? 'Dashboard';
      document.title = `CacheFlow | ${pageName}`;
  }, [currentTab]);

  let component = null;

  switch(currentTab) {
      case "dashboard":
          component = <Dashboard />;
          break;
      case "profile":
          component = <Profile />;
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