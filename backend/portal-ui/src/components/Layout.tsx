import { Outlet, NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../lib/auth";
import { api } from "../api/client";

const navItems = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/subscriptions", label: "Subscriptions" },
  { to: "/deployments", label: "Deployments" },
  { to: "/enrollment-tokens", label: "Enrollment Tokens" },
  { to: "/licenses", label: "Licenses" },
  { to: "/account", label: "Account" },
];

export default function Layout() {
  const { me, setMe } = useAuth();
  const nav = useNavigate();

  const logout = async () => {
    try { await api.logout(); } catch {}
    setMe(null);
    nav("/login", { replace: true });
  };

  return (
    <div className="min-h-full flex flex-col">
      <header className="border-b border-ink-700 bg-ink-800">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded bg-accent-500 flex items-center justify-center font-bold">i</div>
            <span className="font-semibold">infraYS Portal</span>
          </div>
          <div className="flex items-center gap-3 text-sm">
            <span className="text-gray-400">{me?.email}</span>
            <button onClick={logout} className="btn-secondary text-xs">Sign out</button>
          </div>
        </div>
      </header>

      {me && !me.email_verified && <VerifyEmailBanner />}

      <div className="flex-1 max-w-6xl mx-auto w-full px-6 py-8 flex gap-8">
        <nav className="w-48 shrink-0 space-y-1">
          {navItems.map((it) => (
            <NavLink
              key={it.to}
              to={it.to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-md text-sm ${
                  isActive
                    ? "bg-accent-500 text-white"
                    : "text-gray-300 hover:bg-ink-700"
                }`
              }
            >
              {it.label}
            </NavLink>
          ))}
        </nav>
        <main className="flex-1 min-w-0">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

function VerifyEmailBanner() {
  const resend = async () => {
    try {
      await api.resendVerification();
      alert("Verification email re-sent.");
    } catch (e: any) {
      alert(e.message);
    }
  };
  return (
    <div className="bg-amber-900/40 border-b border-amber-800 text-amber-200 text-sm">
      <div className="max-w-6xl mx-auto px-6 py-2 flex items-center justify-between">
        <span>Verify your email to create enrollment tokens.</span>
        <button onClick={resend} className="underline text-amber-100">Resend verification</button>
      </div>
    </div>
  );
}
