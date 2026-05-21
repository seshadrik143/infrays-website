import { useEffect, useState } from "react";
import { adminApi, AuditEntry } from "../../api/admin";

export default function AdminAudit() {
  const [items, setItems] = useState<AuditEntry[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [limit, setLimit] = useState(100);

  useEffect(() => {
    adminApi.listAudit(limit)
      .then((r) => setItems(r.entries || []))
      .catch((e) => setErr(e.message));
  }, [limit]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Audit Log</h1>
        <select value={limit} onChange={(e) => setLimit(parseInt(e.target.value))} className="input w-32">
          <option value={50}>Last 50</option>
          <option value={100}>Last 100</option>
          <option value={250}>Last 250</option>
          <option value={500}>Last 500</option>
        </select>
      </div>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      <div className="card p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-ink-900">
            <tr className="text-left text-xs text-gray-400">
              <th className="px-3 py-2">Seq</th>
              <th className="px-3 py-2">Time</th>
              <th className="px-3 py-2">Event</th>
              <th className="px-3 py-2">Actor</th>
              <th className="px-3 py-2">Customer</th>
              <th className="px-3 py-2">Hash</th>
            </tr>
          </thead>
          <tbody>
            {items.map((e) => (
              <tr key={e.seq} className="border-t border-ink-700 text-xs">
                <td className="px-3 py-2 font-mono text-gray-500">{e.seq}</td>
                <td className="px-3 py-2 font-mono">{new Date(e.created_at).toLocaleString()}</td>
                <td className="px-3 py-2 text-gray-200">{e.event_type}</td>
                <td className="px-3 py-2 text-gray-400">{e.actor}</td>
                <td className="px-3 py-2 text-gray-400 font-mono">{e.customer_id || "—"}</td>
                <td className="px-3 py-2 font-mono text-gray-600">{e.hash?.slice(0, 12) || "—"}</td>
              </tr>
            ))}
            {items.length === 0 && (
              <tr><td colSpan={6} className="px-3 py-8 text-center text-gray-400">No entries.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
