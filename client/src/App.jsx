import { useEffect, useState } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Navbar from "./components/Navbar";
import Login from "./pages/Login";
import Signup from "./pages/Signup";
import Profile from "./pages/Profile";
import Search from "./pages/Search";
import Stock from "./pages/Stock";
import Portfolio from "./pages/Portfolio";

function generateDeviceId() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function ProtectedRoute({ children }) {
  const isAuthenticated = localStorage.getItem('x-cf-bearer');
  return isAuthenticated ? children : <Navigate to="/" replace />;
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('x-cf-bearer'));

  useEffect(() => {
    if (!localStorage.getItem('x-cf-device-id')) {
      localStorage.setItem('x-cf-device-id', generateDeviceId());
    }

    // Listen for custom auth events
    const handleAuthChange = () => {
      setIsAuthenticated(!!localStorage.getItem('x-cf-bearer'));
    };

    window.addEventListener('authChange', handleAuthChange);
    return () => window.removeEventListener('authChange', handleAuthChange);
  }, []);

  return (
    <Router>
      {isAuthenticated && <Navbar />}
      
      <div style={{ padding: "1rem" }}>
        <Routes>
          <Route path="/" element={<Login />} />
          <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
          <Route path="/search" element={<ProtectedRoute><Search /></ProtectedRoute>} />
          <Route path="/stock" element={<ProtectedRoute><Stock /></ProtectedRoute>} />
          <Route path="/portfolio" element={<ProtectedRoute><Portfolio /></ProtectedRoute>} />
          <Route path="/signup" element={<Signup />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;