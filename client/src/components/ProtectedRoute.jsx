import { Navigate } from "react-router-dom";

export default function ProtectedRoute({ children }) {
  const bearer = localStorage.getItem("x-cf-bearer");
  if (!bearer) return <Navigate to="/login" replace />;
  return children;
}
