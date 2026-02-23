import LoadingScreen from '../app/loading_screen';
import type { LoadInAccountData } from '../util/account_data';
import api from '../util/api';
import React, { useState, useEffect, createContext, useContext } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';

const AuthContext = createContext<{
  isAuthenticated: boolean | null;
  setIsAuthenticated: React.Dispatch<React.SetStateAction<boolean | null>>;
  callLoadInAccount: () => Promise<void>;
  checkAuthentication: () => Promise<void>;
  accountData: LoadInAccountData | null;
  setAccountData: React.Dispatch<React.SetStateAction<LoadInAccountData | null>>;
}>({
  isAuthenticated: null,
  setIsAuthenticated: () => {},
  callLoadInAccount: async () => {},
  checkAuthentication: async () => {},
  accountData: null,
  setAccountData: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const [accountData, setAccountData] = useState<LoadInAccountData | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  const deviceID = localStorage.getItem('x-cf-device-id');
  const username = localStorage.getItem('x-cf-uid');

  const callLoadInAccount = async () => {
    try {
      const response = await api.get('/v1/account/loadin', {});
      const loadInData = response.data as LoadInAccountData;
      setAccountData(loadInData);
    } catch (error) {
      console.error('Error calling load in account:', error);
    }
  };

  const refreshToken = async () => {

    try{

        const response = await api.post('/oauth2/token', {
            password: localStorage.getItem('x-cf-refresh'),
            grant_type: 'refresh',
            scope: "external",
            username: username
        });

        if(response.status === 200) {
            await callLoadInAccount();
            setIsAuthenticated(true);
            localStorage.setItem('x-cf-bearer', response.data.access_token);
            localStorage.setItem('x-cf-refresh', response.data.refresh_token);
        } else {
            setIsAuthenticated(false);
            if (location.pathname.includes('/client-app')) {
              navigate('/login');
            }
        }

    } catch(error) {
        console.error('Error refreshing token:', error);
        setIsAuthenticated(false);
        if (location.pathname.includes('/client-app')) {
          navigate('/login');
        }
    }

  };

  const checkAuthentication = async () => {
    if (deviceID && username) {
      try {
        const response = await api.post('/v1/account/verify-token', {}, {
          headers: {
            'x-cf-device-id': deviceID,
            'x-cf-uid': username,
            'x-cf-auth-scope': 'external',
            'x-cf-bearer': localStorage.getItem('x-cf-bearer'),
            'x-cf-refresh': localStorage.getItem('x-cf-refresh'),
          },
        //   withCredentials: true,
        });

        if (response.status === 200) {
          if(response.data.access_token) {
            localStorage.setItem('x-cf-bearer', response.data.access_token);
          }
          if(response.data.refresh_token) {
            localStorage.setItem('x-cf-refresh', response.data.refresh_token);
          }
          await callLoadInAccount();
          setIsAuthenticated(true);
        } else {
          await refreshToken();
        }
      } catch (error) {
        await refreshToken();
      }
    } else {
      setIsAuthenticated(false);
      if (location.pathname.includes('/client-app')) {
        navigate('/login');
      }
    }
  };

useEffect(() => {
  if (import.meta.env.VITE_NODE_ENV === "development") {
    setIsAuthenticated(true);
    console.log("working in dev mode")
  } else {
    checkAuthentication();
  }
}, []);

  if (isAuthenticated === null && window.location.pathname.includes('/client-app')) {
    return <LoadingScreen />;
  }

  if (!isAuthenticated && window.location.pathname.includes('/client-app')) {
    return <Navigate to="/login" />;
  }

  return <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticated, callLoadInAccount, checkAuthentication, accountData, setAccountData }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}