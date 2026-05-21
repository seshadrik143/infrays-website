import { useEffect, useState } from "react";
import { api, Deployment } from "../api/client";
import { fmtDate } from "./Subscriptions";

export default function DeploymentsPage() {
  const [deps, setDeps] = useState<Deployment[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.listDeployments().then((r) => setDeps(r.deployments || [])).catch((e) => setErr(e.message));
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Deployments</h1>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      {deps.length === 0 && (
        <div className="card text-gray-400 text-sm">
          No deployments yet. Once a NodePulse server enrolls with one of your enrollment tokens it will show up here.
        </div>
      )}
      <div className="space-y-3">
        {deps.map((d) => (
          <div key={d.id} className="card">
            <div className="flex items-start justify-between">
              <div>
                <div className="font-medium">{d.deployment_name || d.deployment_id}</div>
                <div className="text-xs text-gray-500 mt-1 font-mono">{d.deployment_id}</div>
              </div>
              {d.flagged_for_review && (
                <span className="px-2 py-1 rounded text-xs bg-amber-900/40 text-amber-300">
                  Flagged
                </span>
              )}
            </div>
            <div className="grid grid-cols-3 gap-2 mt-4 text-xs text-gray-400">
              <div>First seen: {fmtDate(d.first_seen_at)}</div>
              <div>Last seen: {fmtDate(d.last_seen_at)}</div>
              <div>Version: {d.last_version || "—"}</div>
            </div>
            {d.flag_reason && (
              <div className="mt-2 text-xs text-amber-200">{d.flag_reason}</div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
