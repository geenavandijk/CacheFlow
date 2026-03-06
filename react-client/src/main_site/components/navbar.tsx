import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Menu, X } from "lucide-react";
import { AnimatePresence, motion } from "framer-motion";
import { useAuth } from "../../provider/auth";

const MainSiteNavbar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const { isAuthenticated } = useAuth();

  interface Tab {
    name: string;
    path: string;
  }

  const tabs: Tab[] = [
    { name: "Home", path: "/" },
  ];

  const currentTab = location.pathname;

  const goTo = (path: string) => {
    navigate(path);
    setMobileMenuOpen(false);
  };

  return (
    <div className="z-10">
      <div className=" w-full h-20 flex items-center justify-between px-4">
        <div className="flex items-center lg:w-1/3">
          <button  onClick={() => navigate("/")} className="hover:cursor-pointer focus:outline-none">
            <h1 className="text-2xl font-bold text-orange-500">Cache<span className="text-blue-600">Flow</span></h1>
          </button>
        </div>

        {/* Desktop: center tabs + login */}
        <div className="hidden lg:w-full md:flex flex-row items-center space-x-2">

          <div className="flex items-center justify-center mx-auto rounded-full bg-black/20 p-2">
          {tabs.map((tab) => (
            <button
              key={tab.path}
              onClick={() => navigate(tab.path)}
              className={`font-medium text-white hover:cursor-pointer px-4 py-2 w-28 rounded-full transition-all duration-300 ${
                currentTab === tab.path
                  ? "bg-white/10 border border-white/10"
                  : "text- hover:bg-primary"
              }`}
            >
              {tab.name}
            </button>
          ))}
          </div>
        </div>

        <div className="flex lg:w-1/3 items-center justify-end gap-2">
          <div className="hidden md:flex rounded-full p-2 flex-row items-center space-x-2">
            <button
              onClick={() => isAuthenticated ? goTo("/client-app/profile") : goTo("/login")}
              className="bg-orange-500 ml-auto hover:cursor-pointer hover:bg-orange-600 transition-all duration-300 font-medium text-white px-6 py-2 rounded-full"
            >
              {
                isAuthenticated ? "Dashboard" : "Login"
              }
            </button>
          </div>

          {/* Mobile: menu icon (opacity style) */}
          <button
            aria-label="Open menu"
            onClick={() => setMobileMenuOpen(true)}
            className="md:hidden p-2 rounded-full text-white/80 hover:text-white hover:bg-white/10 transition-all duration-200"
          >
            <Menu className="w-8 h-8" strokeWidth={1.5} />
          </button>
        </div>
      </div>

      {/* Full-screen mobile menu modal */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 z-50 bg-black/95 backdrop-blur-md md:hidden"
          >
            {/* Top bar: spacer + X (z-10 so X stays above the full-screen content and is clickable) */}
            <div className="relative z-10 flex items-center justify-between h-20 px-4">
              <div className="w-16" aria-hidden />
              <button
                aria-label="Close menu"
                onClick={() => setMobileMenuOpen(false)}
                className="p-2 rounded-full text-white/80 hover:text-white hover:bg-white/10 transition-all duration-200"
              >
                <X className="w-8 h-8" strokeWidth={1.5} />
              </button>
            </div>

            {/* Centered tabs - big writing */}
            <motion.div
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 8 }}
              transition={{ duration: 0.2, delay: 0.05 }}
              className="absolute inset-0 flex flex-col items-center justify-center gap-8 px-6"
            >
              {tabs.map((tab, i) => (
                <motion.button
                  key={tab.path}
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2, delay: 0.05 + i * 0.04 }}
                  onClick={() => goTo(tab.path)}
                  className={`text-3xl sm:text-4xl font-medium transition-colors ${
                    currentTab === tab.path
                      ? "text-primary"
                      : "text-white hover:text-primary"
                  }`}
                >
                  {tab.name}
                </motion.button>
              ))}
              <motion.button
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.2, delay: 0.05 + tabs.length * 0.04 }}
                onClick={() => isAuthenticated ? goTo("/client-app/profile") : goTo("/login")}
                className="text-3xl sm:text-4xl font-medium text-white hover:text-primary transition-colors mt-4"
              >
                {
                  isAuthenticated ? "Dashboard" : "Login"
                }
              </motion.button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default MainSiteNavbar;
