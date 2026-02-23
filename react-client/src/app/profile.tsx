import { useAuth } from "../provider/auth";

export const Profile = () => {
  const { accountData } = useAuth();

  if (!accountData) {
    return (
      <div className="flex items-center justify-center w-full h-full bg-black text-white">
        <p>[We need to have account data working or else this will show]
            <br></br>LOADING...
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col w-full h-full bg-black text-white px-8 py-8">
      <h1 className="text-2xl font-semibold mb-2">
        Profile{accountData.first_name ? `: ${accountData.first_name}` : ""}
      </h1>

      <div className="mt-4 space-y-2 p-4 border border-neutral-800 rounded-lg bg-neutral-900/50">
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
  );
};