import { Link, useNavigate } from "react-router-dom";
import Logo from "./Logo";

export default function Navbar() {
  const navigate = useNavigate();

  const handleLogout = () => {
    // Clear tokens
    localStorage.removeItem('x-cf-bearer');
    localStorage.removeItem('x-cf-refresh');
    localStorage.removeItem('x-cf-uid');
    
    // Trigger auth state change
    window.dispatchEvent(new Event('authChange'));
    
    // Redirect to login
    navigate('/');
  };

  return (
    <nav style={{ 
      padding: "1rem 2rem", 
      display: "flex", 
      justifyContent: "space-between",
      alignItems: "center"
    }}>
      <Logo size="small" />
      
      <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
        <Link to="/profile">Profile</Link>
        <Link to="/search">Search</Link>
        <Link to="/stock">Stock</Link>
        <Link to="/portfolio">Portfolio</Link>
        <button 
          onClick={handleLogout} 
          className="secondary"
          style={{ marginLeft: '1rem', padding: '8px 16px' }}
        >
          Logout
        </button>
      </div>
    </nav>
  );
}