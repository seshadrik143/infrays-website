import { Outlet, NavLink, useNavigate } from "react-router-dom";
import { adminApi } from "../../api/admin";
import { useAdminAuth } from "../../lib/adminAuth";

const navItems = [
  { to: "/admin/dashboard", label: "Dashboard" },
  { to: "/admin/customers", label: "Customers" },
  { to: "/admin/audit", label: "Audit Log" },
];

export default function AdminLayout() {
  const { me, setMe } = useAdminAuth();
  const nav = useNavigate();

  const logout = async () => {
    try { await adminApi.logout(); } catch {}
    setMe(null);
    nav("/admin/login", { replace: true });
  };

  return (
    <div className="min-h-full flex flex-col">
      <header className="border-b border-ink-700 bg-ink-900">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded bg-red-500 flex items-center justify-center font-bold">A</div>
            <span className="font-semibold">infraYS Admin</span>
          </div>
          <div className="flex items-center gap-3 text-sm">
            <span className="text-gray-400">{me?.email}</span>
            <span className="px-2 py-1 rounded text-xs bg-red-900/40 text-red-300">{me?.role}</span>
            <button onClick={logout} className="btn-secondary text-xs">Sign out</button>
          </div>
        </div>
      </header>
      <div className="flex-1 max-w-6xl mx-auto w-full px-6 py-8 flex gap-8">
        <nav className="w-48 shrink-0 space-y-1">
          {navItems.map((it) => (
            <NavLink
              key={it.to}
              to={it.to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-md text-sm ${
                  isActive ? "bg-red-500 text-white" : "text-gray-300 hover:bg-ink-700"
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
