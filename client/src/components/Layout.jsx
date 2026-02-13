import Navbar from "./Navbar";

export default function Layout({ children }) {
  return (
    <div className="min-h-screen">
      <Navbar />
      <main style={{ maxWidth: 1100, margin: "0 auto", padding: "24px 16px" }}>
        {children}
      </main>
    </div>
  );
}
