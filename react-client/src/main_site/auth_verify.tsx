import React, { useEffect, useState } from "react";
import { CheckCircle2, AlertCircle } from "lucide-react";
import { getApiBaseUrl } from "../util/api";

type Status = "loading" | "success" | "error";

const AuthVerify: React.FC = () => {
  const [status, setStatus] = useState<Status>("loading");
  const [message, setMessage] = useState<string>("Verifying your account...");

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const secret = params.get("secret");

    if (!secret) {
      setStatus("error");
      setMessage("Missing verification token in URL.");
      return;
    }

    const verify = async () => {
      try {
        const res = await fetch(
          `${getApiBaseUrl()}/v1/account/verify?secret=${encodeURIComponent(
            secret,
          )}`,
        );

        if (!res.ok) {
          const data = await res.json().catch(() => null);
          throw new Error(data?.error || "Verification failed");
        }

        setStatus("success");
        setMessage("Your email has been verified. You can now log in.");
      } catch (err: any) {
        setStatus("error");
        setMessage(err.message || "Verification failed. Please try again.");
      }
    };

    verify();
  }, []);

  const Icon = status === "success" ? CheckCircle2 : AlertCircle;
  const iconColor =
    status === "success" ? "text-green-400" : status === "loading" ? "text-blue-400" : "text-red-400";

  return (
    <div className="flex flex-row items-center justify-center h-full w-full bg-black">
      <div className="hidden md:flex flex-col w-1/2 h-full border-r border-neutral-800 items-center justify-center">
        <h1 className="text-white text-4xl font-bold">Verify email</h1>
      </div>

      <div className="flex flex-col w-full md:w-1/2 h-full justify-center px-12">
        <div className="flex flex-col w-full xl:w-[36rem] items-center text-center">
          <Icon className={`w-12 h-12 mb-6 ${iconColor}`} />
          <h1 className="text-white text-2xl font-semibold">
            {status === "loading"
              ? "Verifying your email..."
              : status === "success"
                ? "Email verified"
                : "Verification failed"}
          </h1>
          <p className="text-neutral-400 text-sm mt-3 max-w-md">{message}</p>
        </div>
      </div>
    </div>
  );
};

export default AuthVerify;

