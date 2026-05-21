import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Layout from "./components/Layout";
import LoginPage from "./pages/Login";
import SignupPage from "./pages/Signup";
import VerifyEmailPage from "./pages/VerifyEmail";
import ResetPasswordPage from "./pages/ResetPassword";
import DashboardPage from "./pages/Dashboard";
import SubscriptionsPage from "./pages/Subscriptions";
import DeploymentsPage from "./pages/Deployments";
import EnrollmentTokensPage from "./pages/EnrollmentTokens";
import LicensesPage from "./pages/Licenses";
import AccountPage from "./pages/Account";

function RequireAuth({ children }: { children: JSX.Element }) {
  const { me, loading } = useAuth();
  if (loading) return <div className="p-12 text-center text-gray-400">Loading…</div>;
  if (!me) return <Navigate to="/login" replace />;
  return children;
}

function RedirectAuthed({ children }: { children: JSX.Element }) {
  const { me, loading } = useAuth();
  if (loading) return null;
  if (me) return <Navigate to="/dashboard" replace />;
  return children;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<RedirectAuthed><LoginPage /></RedirectAuthed>} />
      <Route path="/signup" element={<RedirectAuthed><SignupPage /></RedirectAuthed>} />
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />

      <Route element={<RequireAuth><Layout /></RequireAuth>}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/subscriptions" element={<SubscriptionsPage />} />
        <Route path="/deployments" element={<DeploymentsPage />} />
        <Route path="/enrollment-tokens" element={<EnrollmentTokensPage />} />
        <Route path="/licenses" element={<LicensesPage />} />
        <Route path="/account" element={<AccountPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
