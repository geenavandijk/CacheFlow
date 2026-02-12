export default function Logo({ size = "large" }) {
  const fontSize = size === "large" ? "2rem" : "1.5rem";
  
  return (
    <div className="logo" style={{ fontSize }}>
      <span className="logo-cache">Cache</span>
      <span className="logo-flow">Flow</span>
    </div>
  );
}