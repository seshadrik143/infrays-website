import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { adminApi, AdminCustomer } from "../../api/admin";

export default function AdminCustomerList() {
  const [items, setItems] = useState<AdminCustomer[]>([]);
  const [q, setQ] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [busy, setBusy] = useState(false);

  const load = (filter: string) => {
    adminApi.listCustomers(filter)
      .then((r) => {
        setItems(r.customers || []);
        setSelected(new Set()); // clear after reload
      })
      .catch((e) => setErr(e.message));
  };

  useEffect(() => { load(""); }, []);

  const onSearch = (e: React.FormEvent) => {
    e.preventDefault();
    load(q);
  };

  const toggle = (id: string) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  };

  const toggleAll = () => {
    if (selected.size === items.length) setSelected(new Set());
    else setSelected(new Set(items.map((c) => c.id)));
  };

  const bulkStatus = async (status: "suspended" | "active") => {
    const ids = Array.from(selected);
    const verb = status === "suspended" ? "Suspend" : "Reactivate";
    if (!confirm(`${verb} ${ids.length} customer(s)?`)) return;
    setBusy(true);
    try {
      const r = await adminApi.bulkCustomerStatus({ ids, status });
      alert(`${verb}d ${r.success} customer(s).${r.failed ? ` ${r.failed} failed.` : ""}`);
      load(q);
    } catch (e: any) {
      alert(e.message);
    } finally {
      setBusy(false);
    }
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

      {selected.size > 0 && (
        <div className="card bg-ink-900 border-accent-500 flex items-center justify-between">
          <div className="text-sm">
            <span className="font-medium">{selected.size}</span> selected
          </div>
          <div className="flex gap-2">
            <button disabled={busy} onClick={() => bulkStatus("suspended")} className="btn-secondary text-xs text-amber-300">Suspend</button>
            <button disabled={busy} onClick={() => bulkStatus("active")} className="btn-secondary text-xs text-green-300">Reactivate</button>
            <button onClick={() => setSelected(new Set())} className="btn-secondary text-xs">Clear</button>
          </div>
        </div>
      )}

      <div className="card p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-ink-900">
            <tr className="text-left text-xs text-gray-400">
              <th className="px-3 py-2 w-8">
                <input
                  type="checkbox"
                  checked={items.length > 0 && selected.size === items.length}
                  onChange={toggleAll}
                  aria-label="Select all"
                />
              </th>
              <th className="px-4 py-2">Email</th>
              <th className="px-4 py-2">Company</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2">Created</th>
            </tr>
          </thead>
          <tbody>
            {items.map((c) => (
              <tr key={c.id} className="border-t border-ink-700 hover:bg-ink-700/50">
                <td className="px-3 py-2">
                  <input
                    type="checkbox"
                    checked={selected.has(c.id)}
                    onChange={() => toggle(c.id)}
                    aria-label={`Select ${c.email}`}
                  />
                </td>
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
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">No customers.</td></tr>
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
