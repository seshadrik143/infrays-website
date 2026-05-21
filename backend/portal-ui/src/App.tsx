import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "./lib/auth";
import { AdminAuthProvider, useAdminAuth } from "./lib/adminAuth";
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
import AdminLoginPage from "./pages/admin/AdminLogin";
import AdminLayout from "./pages/admin/AdminLayout";
import AdminDashboard from "./pages/admin/AdminDashboard";
import AdminCustomerList from "./pages/admin/AdminCustomerList";
import AdminCustomerDetail from "./pages/admin/AdminCustomerDetail";
import AdminAudit from "./pages/admin/AdminAudit";

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

      <Route path="/admin/*" element={<AdminAuthProvider><AdminRoutes /></AdminAuthProvider>} />

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

function RequireAdmin({ children }: { children: JSX.Element }) {
  const { me, loading } = useAdminAuth();
  if (loading) return <div className="p-12 text-center text-gray-400">Loading…</div>;
  if (!me || !me.mfa_verified) return <Navigate to="/admin/login" replace />;
  return children;
}

function AdminRoutes() {
  return (
    <Routes>
      <Route path="login" element={<AdminLoginPage />} />
      <Route element={<RequireAdmin><AdminLayout /></RequireAdmin>}>
        <Route path="dashboard" element={<AdminDashboard />} />
        <Route path="customers" element={<AdminCustomerList />} />
        <Route path="customers/:id" element={<AdminCustomerDetail />} />
        <Route path="audit" element={<AdminAudit />} />
      </Route>
      <Route path="*" element={<Navigate to="/admin/dashboard" replace />} />
    </Routes>
  );
}
