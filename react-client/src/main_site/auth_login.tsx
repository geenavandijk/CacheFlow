import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../provider/auth";
import api from "../util/api";

const AuthLogin = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();

    const { setIsAuthenticated, callLoadInAccount } = useAuth();

  const handleLogin = async () => {
    try {
        localStorage.setItem("x-cf-uid", email);
      const response = await api.post("/oauth2/token", {
        password: password,
        grant_type: "password",
        scope: "external",
        username: email,
      }, {
        headers: {
          "Content-Type": "application/json",
          "x-cf-device-id": localStorage.getItem("x-cf-device-id"),
          "x-cf-uid": email,
          "x-cf-bearer": localStorage.getItem("x-cf-bearer"),
          "x-cf-refresh": localStorage.getItem("x-cf-refresh"),
        }
      });

      if (response.data?.success) {
        localStorage.setItem("x-cf-uid", email);
        localStorage.setItem("x-cf-bearer", response.data?.access_token);
        localStorage.setItem("x-cf-refresh", response.data?.refresh_token);

        await callLoadInAccount();
        setIsAuthenticated(true);
        navigate("/client-app/dashboard");

      } else {

        // add toast error
        setError(response?.data?.error)
      }
    } catch (error: any) {

        // add toast error
        setError("Invalid email or password")
    }
  };

  return (
    <div className="flex flex-row items-center justify-center h-full w-full bg-black">
      <div className="hidden md:flex flex-col w-1/2 h-full border-r border-neutral-800 items-center justify-center">
        <h1 className="text-white text-4xl font-bold">Login Image</h1>
      </div>

      <div className="flex flex-col w-full md:w-1/2 h-full justify-center px-12">
      <div className="flex flex-col w-full xl:w-[36rem]">
      <h1 className="text-white text-lg font-regular">
          Log in to CacheFlow
        </h1>

        <h1 className="text-white text-sm font-regular mt-10">Email</h1>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
        />

        <h1 className="text-white text-sm font-regular mt-4">Password</h1>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
        />

        {
            error && (
                <p className="text-red-500 text-sm mt-4">
                    {error}
                </p>
            )
        }

        <div className="flex flex-row items-center mt-8 gap-x-4">


        <button
        onClick={handleLogin}
        className="rounded-full w-1/2 hover:cursor-pointer bg-orange-500 text-white font-medium px-6 py-4 transition-all duration-300 hover:bg-orange-600">
          Log in
        </button>

        <button className="rounded-full w-1/2 hover:cursor-pointer border border-1 border-orange-500 text-white font-medium px-6 py-4 transition-all duration-300 hover:bg-orange-600">
          Need help
        </button>

        </div>

        <h1 className="text-white text-sm font-regular mt-8">
            Not on CacheFlow? <Link
            to="/signup"
            className="text-orange-500 underline hover:cursor-pointer">Sign up</Link>
        </h1>

        <h1 className="text-neutral-400 text-xs font-regular mt-8">
            By signing in, you agree to our Privacy Policy and Terms of Service.
        </h1>

      </div>
      </div>
    </div>
  );
};

export default AuthLogin;
