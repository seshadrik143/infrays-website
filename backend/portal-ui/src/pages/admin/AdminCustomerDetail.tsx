import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { adminApi, CustomerDetail } from "../../api/admin";
import { StatusPill } from "./AdminCustomerList";

export default function AdminCustomerDetail() {
  const { id = "" } = useParams();
  const [data, setData] = useState<CustomerDetail | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [newPlain, setNewPlain] = useState<string | null>(null);

  const load = () => {
    adminApi.getCustomer(id).then(setData).catch((e) => setErr(e.message));
  };
  useEffect(() => { load(); }, [id]);

  if (err) return <div className="text-red-400">{err}</div>;
  if (!data) return <div className="text-gray-400">Loading…</div>;
  const c = data.customer;

  const updateStatus = async (status: string) => {
    if (!confirm(`Set status to "${status}"?`)) return;
    setBusy(true);
    try {
      await adminApi.updateCustomer(id, { status });
      load();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setBusy(false);
    }
  };

  const flagDep = async (depID: string, flagged: boolean) => {
    const reason = flagged ? (prompt("Reason for flagging:") || "") : "";
    if (flagged && !reason) return;
    try {
      await adminApi.flagDeployment(depID, { flagged, reason });
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const revokeLicense = async (jti: string) => {
    const reason = prompt("Revocation reason:");
    if (!reason) return;
    try {
      await adminApi.revokeLicense(jti, reason);
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const issueToken = async (subID: string) => {
    const label = prompt("Token label (e.g. prod-cluster):") || "";
    try {
      const r = await adminApi.createEnrollmentToken(id, {
        subscription_id: subID, label, ttl_hours: 24,
      });
      setNewPlain(r.plaintext);
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <Link to="/admin/customers" className="link text-sm">← Customers</Link>
        <h1 className="text-2xl font-semibold mt-2">{c.email}</h1>
        <div className="flex items-center gap-2 mt-2 text-sm text-gray-400">
          <StatusPill status={c.status} />
          {c.company && <span>· {c.company}</span>}
          {c.name && <span>· {c.name}</span>}
        </div>
      </div>

      <div className="card">
        <h2 className="font-medium mb-3">Account actions</h2>
        <div className="flex gap-2">
          {c.status !== "active" && (
            <button onClick={() => updateStatus("active")} disabled={busy} className="btn-secondary text-sm">Reactivate</button>
          )}
          {c.status !== "suspended" && (
            <button onClick={() => updateStatus("suspended")} disabled={busy} className="btn-secondary text-sm">Suspend</button>
          )}
        </div>
      </div>

      {newPlain && (
        <div className="card border-accent-500">
          <div className="font-medium mb-2">Enrollment token — copy now, won't be shown again:</div>
          <code className="block bg-ink-900 p-3 rounded text-xs font-mono break-all">{newPlain}</code>
          <button onClick={() => setNewPlain(null)} className="btn-secondary text-sm mt-3">Dismiss</button>
        </div>
      )}

      <Section title="Subscriptions">
        {(data.subscriptions || []).length === 0 && <Empty>No subscriptions.</Empty>}
        {(data.subscriptions || []).map((s) => (
          <div key={s.id} className="flex items-center justify-between p-3 border-b border-ink-700 text-sm last:border-0">
            <div>
              <span className="capitalize font-medium">{s.tier}</span>
              <span className="text-xs text-gray-500 ml-2">{s.id}</span>
              {s.manual_offline && <span className="ml-2 px-2 py-0.5 rounded text-xs bg-gray-700 text-gray-300">offline</span>}
            </div>
            <div className="flex items-center gap-3">
              <StatusPill status={s.status} />
              <button onClick={() => issueToken(s.id)} className="btn-secondary text-xs">Issue token</button>
            </div>
          </div>
        ))}
      </Section>

      <Section title="Deployments">
        {(data.deployments || []).length === 0 && <Empty>No deployments.</Empty>}
        {(data.deployments || []).map((d) => (
          <div key={d.id} className="flex items-center justify-between p-3 border-b border-ink-700 text-sm last:border-0">
            <div>
              <div className="font-medium">{d.deployment_name || d.deployment_id}</div>
              <div className="text-xs text-gray-500 font-mono">{d.deployment_id}</div>
              {d.flag_reason && <div className="text-xs text-amber-200 mt-1">{d.flag_reason}</div>}
            </div>
            <div>
              {d.flagged_for_review ? (
                <button onClick={() => flagDep(d.deployment_id, false)} className="btn-secondary text-xs">Unflag</button>
              ) : (
                <button onClick={() => flagDep(d.deployment_id, true)} className="btn-secondary text-xs">Flag</button>
              )}
            </div>
          </div>
        ))}
      </Section>

      <Section title="Enrollment tokens">
        {(data.enrollment_tokens || []).length === 0 && <Empty>No tokens.</Empty>}
        {(data.enrollment_tokens || []).map((t) => (
          <div key={t.id} className="flex items-center justify-between p-3 border-b border-ink-700 text-sm last:border-0">
            <div>
              <div className="font-medium">{t.label || "(no label)"}</div>
              <div className="text-xs text-gray-500 font-mono">{t.id}</div>
            </div>
            <div className="text-xs text-gray-400">
              {t.consumed_at && !t.consumed_at.startsWith("0001") ? "Used" : `Expires ${new Date(t.expires_at).toLocaleDateString()}`}
            </div>
          </div>
        ))}
      </Section>

      <Section title="Licenses">
        {(data.licenses || []).length === 0 && <Empty>No licenses.</Empty>}
        {(data.licenses || []).map((l) => (
          <div key={l.jti} className="flex items-center justify-between p-3 border-b border-ink-700 text-sm last:border-0">
            <div>
              <div className="capitalize font-medium">{l.tier}</div>
              <div className="text-xs text-gray-500 font-mono">{l.license_id} · dep {l.deployment_id}</div>
            </div>
            <div className="flex items-center gap-2">
              {l.revoked ? (
                <span className="px-2 py-1 rounded text-xs bg-red-900/40 text-red-300">Revoked</span>
              ) : (
                <button onClick={() => revokeLicense(l.jti)} className="text-xs text-red-400 hover:underline">Revoke</button>
              )}
            </div>
          </div>
        ))}
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="card p-0 overflow-hidden">
      <div className="px-4 py-2 bg-ink-900 text-xs uppercase tracking-wider text-gray-400">{title}</div>
      <div>{children}</div>
    </div>
  );
}

function Empty({ children }: { children: React.ReactNode }) {
  return <div className="p-4 text-gray-500 text-sm">{children}</div>;
}
