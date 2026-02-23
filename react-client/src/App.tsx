import { useEffect } from 'react'
import './App.css'
import { v4 as uuidv4 } from 'uuid';
import { AuthProvider } from './provider/auth';
import { Route, Routes } from 'react-router-dom';
import MainSiteViewController from './main_site/main_site_view_controller';
import AppViewController from './app/view_controller';

function App() {



  useEffect(() => {
    // check if device_id is set, if not set to a new uuid
    const deviceId = localStorage.getItem("x-cf-device-id");
    if (!deviceId) {
      localStorage.setItem("x-cf-device-id", uuidv4());
    }
  }, []);

  return (
    
      <AuthProvider>
        <Routes>
          <Route path="/" element={<MainSiteViewController />} />
          <Route path="/login" element={<MainSiteViewController />} />
          <Route path="/signup" element={<MainSiteViewController />} />
          <Route path="/verify" element={<MainSiteViewController />} />
          <Route path="/client-app/profile" element={<AppViewController />} />
          <Route path="/client-app/dashboard" element={<AppViewController />} />
        </Routes>
      </AuthProvider>
  
  )
}

export default App
