import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { adminApi, AdminCustomer } from "../../api/admin";

export default function AdminCustomerList() {
  const [items, setItems] = useState<AdminCustomer[]>([]);
  const [q, setQ] = useState("");
  const [err, setErr] = useState<string | null>(null);

  const load = (filter: string) => {
    adminApi.listCustomers(filter)
      .then((r) => setItems(r.customers || []))
      .catch((e) => setErr(e.message));
  };

  useEffect(() => { load(""); }, []);

  const onSearch = (e: React.FormEvent) => {
    e.preventDefault();
    load(q);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Customers</h1>
        <form onSubmit={onSearch} className="flex gap-2">
          <input value={q} onChange={(e) => setQ(e.target.value)}
            placeholder="Search email or company" className="input w-64" />
          <button className="btn-secondary text-sm" type="submit">Search</button>
        </form>
      </div>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      <div className="card p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-ink-900">
            <tr className="text-left text-xs text-gray-400">
              <th className="px-4 py-2">Email</th>
              <th className="px-4 py-2">Company</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2">Created</th>
            </tr>
          </thead>
          <tbody>
            {items.map((c) => (
              <tr key={c.id} className="border-t border-ink-700 hover:bg-ink-700/50">
                <td className="px-4 py-2">
                  <Link to={`/admin/customers/${c.id}`} className="link">{c.email}</Link>
                </td>
                <td className="px-4 py-2 text-gray-300">{c.company || "—"}</td>
                <td className="px-4 py-2">
                  <StatusPill status={c.status} />
                </td>
                <td className="px-4 py-2 text-xs text-gray-500">{new Date(c.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
            {items.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-400">No customers.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function StatusPill({ status }: { status: string }) {
  const color =
    status === "active" ? "bg-green-900/40 text-green-300" :
    status === "suspended" ? "bg-amber-900/40 text-amber-300" :
    status === "deleted" ? "bg-red-900/40 text-red-300" :
    "bg-gray-700 text-gray-300";
  return <span className={`inline-block px-2 py-1 rounded text-xs ${color}`}>{status}</span>;
}
