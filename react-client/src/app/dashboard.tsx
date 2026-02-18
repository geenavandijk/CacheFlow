import AppNavbar from "./components/navbar";
import { useAuth } from "../provider/auth";

export const Dashboard = () => {
  const { accountData } = useAuth();

  return (
    <div className="flex flex-col w-full h-full bg-black text-white">
      <div className="flex-1 flex flex-col px-8 py-8">
        <h1 className="text-2xl font-semibold">
          Welcome{accountData ? `, ${accountData.first_name}` : ""}.
        </h1>
        <p className="mt-2 text-sm text-neutral-400">
          This is your CacheFlow dashboard. You&apos;ll see your account overview and
          product data here as we build things out.
        </p>

        {accountData && (
          <div className="mt-6 rounded-xl border border-neutral-800 bg-neutral-900/40 px-5 py-4 text-sm">
            <h2 className="text-xs font-medium uppercase tracking-[0.16em] text-neutral-500">
              Account
            </h2>
            <div className="mt-3 space-y-1">
              <p>
                <span className="text-neutral-500 mr-2">Name:</span>
                {accountData.first_name} {accountData.last_name}
              </p>
              <p>
                <span className="text-neutral-500 mr-2">Email:</span>
                {accountData.email}
              </p>
              <p>
                <span className="text-neutral-500 mr-2">Account ID:</span>
                <span className="font-mono text-xs text-neutral-300">
                  {accountData.account_id}
                </span>
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

