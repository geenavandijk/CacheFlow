import { useState } from "react";
import { Link } from "react-router-dom";
import { MailCheck } from "lucide-react";
import { getApiBaseUrl } from "../util/api";

const AuthCreate = () => {
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const [currentPage, setCurrentPage] = useState<"personal-info" | "email-sent">("personal-info");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreateAccount = async () => {
    if (isSubmitting) return;
    setIsSubmitting(true);
    setError(null);

    try {
      const response = await fetch(`${getApiBaseUrl()}/v1/account/create`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          first_name: firstName || undefined,
          last_name: lastName || undefined,
          email,
          password,
        }),
      });

      if (!response.ok) {
        const data = await response.json().catch(() => null);
        throw new Error(data?.error || "Failed to create account");
      }

      setCurrentPage("email-sent");
    } catch (err: any) {
      setError(err.message || "Something went wrong");
    } finally {
      setIsSubmitting(false);
    }
  };

  if(currentPage === "personal-info") {
    return (
        <div className="flex flex-row items-center justify-center h-full w-full bg-black">
          <div className="hidden md:flex flex-col w-1/2 h-full border-r border-neutral-800 items-center justify-center">
            <h1 className="text-white text-4xl font-bold">Create Image</h1>
          </div>
    
          <div className="flex flex-col w-full md:w-1/2 h-full justify-center px-12">
          <div className="flex flex-col w-full xl:w-[36rem]">
          <h1 className="text-white text-lg font-regular">
              Create an account
            </h1>
    
            <div className="flex flex-row items-center mt-8 gap-x-4"> 
    
                <div className="flex flex-col w-full">
                <h1 className="text-white text-sm font-regular">First Name</h1>
                <input
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
                />
                </div>  
    
                <div className="flex flex-col w-full">
                <h1 className="text-white text-sm font-regular">Last Name</h1>
                <input
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                    className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
                    />
                </div>
            </div>
    
            <h1 className="text-white text-sm font-regular mt-6">Email</h1>
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
    
            <div className="flex flex-row items-center mt-8 gap-x-4">
    
    
            <button
              className="rounded-full w-full hover:cursor-pointer bg-orange-500 text-white font-medium px-6 py-4 transition-all duration-300 hover:bg-orange-600 disabled:opacity-60 disabled:cursor-not-allowed"
              onClick={handleCreateAccount}
              disabled={isSubmitting}
            >
              {isSubmitting ? "Creating..." : "Create your account"}
            </button>
    
            </div>
    
            {error && (
              <p className="text-red-500 text-sm mt-4">
                {error}
              </p>
            )}

            <h1 className="text-white text-sm font-regular mt-8">
                Already have an account? <Link
                to="/login"
                className="text-orange-500 underline hover:cursor-pointer">Log in</Link>
            </h1>
    
            <h1 className="text-neutral-400 text-xs font-regular mt-8">
                By signing up, you agree to our Privacy Policy and Terms of Service.
            </h1>
    
          </div>
          </div>
        </div>
      );
  }

  if (currentPage === "email-sent") {
    return (
      <div className="flex flex-row items-center justify-center h-full w-full bg-black">
        <div className="hidden md:flex flex-col w-1/2 h-full border-r border-neutral-800 items-center justify-center">
          <h1 className="text-white text-4xl font-bold">Verify your email</h1>
        </div>

        <div className="flex flex-col w-full md:w-1/2 h-full justify-center px-12">
          <div className="flex flex-col w-full xl:w-[36rem] items-center text-center">
            <MailCheck className="w-12 h-12 text-green-400 mb-6" />
            <h1 className="text-white text-2xl font-semibold">
              Check your email
            </h1>
            <p className="text-neutral-400 text-sm mt-3 max-w-md">
              We&apos;ve sent a verification link to{" "}
              <span className="text-white font-medium">{email}</span>. <br />
              Click the link in the email to verify your account.
            </p>
            <p className="text-neutral-500 text-xs mt-6">
              Didn&apos;t get the email? Check your spam folder or try again in a few minutes.
            </p>

            <button
              className="mt-8 rounded-full px-6 py-3 bg-neutral-900 text-white border border-neutral-700 hover:border-white transition-all duration-300"
              onClick={() => setCurrentPage("personal-info")}
            >
              Back to sign up
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-row items-center justify-center h-full w-full bg-black">
      <div className="hidden md:flex flex-col w-1/2 h-full border-r border-neutral-800 items-center justify-center">
        <h1 className="text-white text-4xl font-bold">Create Image</h1>
      </div>

      <div className="flex flex-col w-full md:w-1/2 h-full justify-center px-12">
      <div className="flex flex-col w-full xl:w-[36rem]">
      <h1 className="text-white text-lg font-regular">
          Create an account
        </h1>

        <div className="flex flex-row items-center mt-8 gap-x-4"> 

            <div className="flex flex-col w-full">
            <h1 className="text-white text-sm font-regular">First Name</h1>
            <input
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
            />
            </div>  

            <div className="flex flex-col w-full">
            <h1 className="text-white text-sm font-regular">Last Name</h1>
            <input
            type="text"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
                className="w- h-10 w-full text-white mt-2 rounded-lg border border-neutral-800 focus:border-white focus:border-2 transition-all duration-300 p-2 active:ring-0 focus:ring-0 focus:outline-none"
                />
            </div>
        </div>

        <h1 className="text-white text-sm font-regular mt-6">Email</h1>
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

        <div className="flex flex-row items-center mt-8 gap-x-4">


        <button className="rounded-full w-full hover:cursor-pointer bg-orange-500 text-white font-medium px-6 py-4 transition-all duration-300 hover:bg-orange-600">
          Create your account
        </button>

        </div>

        <h1 className="text-white text-sm font-regular mt-8">
            Already have an account? <Link
            to="/login"
            className="text-orange-500 underline hover:cursor-pointer">Log in</Link>
        </h1>

        <h1 className="text-neutral-400 text-xs font-regular mt-8">
            By signing up, you agree to our Privacy Policy and Terms of Service.
        </h1>

      </div>
      </div>
    </div>
  );
};

export default AuthCreate;
