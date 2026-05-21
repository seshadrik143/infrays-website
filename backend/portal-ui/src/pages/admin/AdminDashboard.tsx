import { useEffect, useState } from "react";
import { adminApi, AdminCustomer, AuditEntry } from "../../api/admin";
import { Link } from "react-router-dom";

export default function AdminDashboard() {
  const [customers, setCustomers] = useState<AdminCustomer[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([adminApi.listCustomers(), adminApi.listAudit(15)])
      .then(([c, a]) => {
        setCustomers(c.customers || []);
        setAudit(a.entries || []);
      })
      .catch((e) => setErr(e.message));
  }, []);

  const recentCustomers = customers.slice(0, 5);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Admin Dashboard</h1>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Total Customers" value={customers.length} link="/admin/customers" />
        <StatCard label="Active" value={customers.filter((c) => c.status === "active").length} />
        <StatCard label="Suspended" value={customers.filter((c) => c.status === "suspended").length} />
      </div>

      <div className="card">
        <h2 className="font-medium mb-3">Recent customers</h2>
        {recentCustomers.length === 0 && <div className="text-gray-400 text-sm">No customers yet.</div>}
        <div className="space-y-1">
          {recentCustomers.map((c) => (
            <Link key={c.id} to={`/admin/customers/${c.id}`}
              className="block px-3 py-2 rounded hover:bg-ink-700 text-sm">
              <div className="flex items-center justify-between">
                <span>{c.email}</span>
                <span className="text-xs text-gray-500">{c.company || "—"}</span>
              </div>
            </Link>
          ))}
        </div>
      </div>

      <div className="card">
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-medium">Recent audit events</h2>
          <Link to="/admin/audit" className="link text-sm">View all</Link>
        </div>
        <div className="space-y-1">
          {audit.map((e) => (
            <div key={e.seq} className="text-xs flex justify-between border-b border-ink-700 pb-1">
              <span className="font-mono text-gray-400">{new Date(e.created_at).toLocaleString()}</span>
              <span className="text-gray-300">{e.event_type}</span>
              <span className="text-gray-500">{e.actor}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, link }: { label: string; value: number; link?: string }) {
  const inner = (
    <div className="card hover:border-red-500 transition-colors h-full">
      <div className="text-xs text-gray-400 uppercase tracking-wider">{label}</div>
      <div className="text-3xl font-semibold mt-2">{value}</div>
    </div>
  );
  return link ? <Link to={link}>{inner}</Link> : inner;
}
