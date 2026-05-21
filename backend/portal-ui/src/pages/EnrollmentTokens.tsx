import { useEffect, useState } from "react";
import { api, EnrollmentToken, Subscription } from "../api/client";
import { fmtDate } from "./Subscriptions";
import { useAuth } from "../lib/auth";

export default function EnrollmentTokensPage() {
  const { me } = useAuth();
  const [tokens, setTokens] = useState<EnrollmentToken[]>([]);
  const [subs, setSubs] = useState<Subscription[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [newPlain, setNewPlain] = useState<string | null>(null);

  const reload = () => {
    Promise.all([api.listEnrollmentTokens(), api.listSubscriptions()])
      .then(([t, s]) => {
        setTokens(t.tokens || []);
        setSubs(s.subscriptions || []);
      })
      .catch((e) => setErr(e.message));
  };

  useEffect(() => { reload(); }, []);

  const revoke = async (id: string) => {
    if (!confirm("Revoke this enrollment token?")) return;
    try {
      await api.revokeEnrollmentToken(id);
      reload();
    } catch (e: any) {
      alert(e.message);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Enrollment Tokens</h1>
        {me?.email_verified && (
          <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">New token</button>
        )}
      </div>
      {!me?.email_verified && (
        <div className="card text-sm text-amber-200">
          Verify your email to create enrollment tokens.
        </div>
      )}
      {err && <div className="text-red-400 text-sm">{err}</div>}

      {newPlain && (
        <div className="card border-accent-500">
          <div className="font-medium mb-2">Save this token now — it won't be shown again.</div>
          <code className="block bg-ink-900 p-3 rounded text-xs font-mono break-all">{newPlain}</code>
          <button onClick={() => setNewPlain(null)} className="btn-secondary text-sm mt-3">Dismiss</button>
        </div>
      )}

      {tokens.length === 0 && !newPlain && (
        <div className="card text-gray-400 text-sm">No tokens yet.</div>
      )}

      <div className="space-y-2">
        {tokens.map((t) => {
          const consumed = t.consumed_at && !t.consumed_at.startsWith("0001");
          const expired = new Date(t.expires_at) < new Date();
          return (
            <div key={t.id} className="card">
              <div className="flex items-start justify-between">
                <div>
                  <div className="font-medium">{t.label || "(no label)"}</div>
                  <div className="text-xs text-gray-500 mt-1 font-mono">{t.id}</div>
                </div>
                {consumed ? (
                  <span className="px-2 py-1 rounded text-xs bg-gray-700 text-gray-300">Used</span>
                ) : expired ? (
                  <span className="px-2 py-1 rounded text-xs bg-red-900/40 text-red-300">Expired</span>
                ) : (
                  <button onClick={() => revoke(t.id)} className="text-xs text-red-400 hover:underline">Revoke</button>
                )}
              </div>
              <div className="grid grid-cols-2 gap-2 mt-3 text-xs text-gray-400">
                <div>Created: {fmtDate(t.created_at)}</div>
                <div>Expires: {fmtDate(t.expires_at)}</div>
                {consumed && t.consumed_by_deployment && (
                  <div className="col-span-2">Used by deployment: <code className="font-mono">{t.consumed_by_deployment}</code></div>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {showCreate && (
        <CreateTokenModal
          subs={subs}
          onClose={() => setShowCreate(false)}
          onCreated={(plain) => {
            setNewPlain(plain);
            setShowCreate(false);
            reload();
          }}
        />
      )}
    </div>
  );
}

function CreateTokenModal({
  subs, onClose, onCreated,
}: {
  subs: Subscription[];
  onClose: () => void;
  onCreated: (plain: string) => void;
}) {
  const [subID, setSubID] = useState(subs[0]?.id || "");
  const [label, setLabel] = useState("");
  const [ttl, setTtl] = useState(24);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      const r = await api.createEnrollmentToken({ subscription_id: subID, label, ttl_hours: ttl });
      onCreated(r.plaintext);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-6 z-10">
      <div className="card w-full max-w-md">
        <h2 className="font-medium mb-4">Create enrollment token</h2>
        <form onSubmit={submit} className="space-y-3">
          <div>
            <label className="label">Subscription</label>
            <select value={subID} onChange={(e) => setSubID(e.target.value)} className="input" required>
              {subs.map((s) => (
                <option key={s.id} value={s.id}>{s.tier} — {s.id}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="label">Label (optional)</label>
            <input value={label} onChange={(e) => setLabel(e.target.value)} className="input" placeholder="prod-cluster-east" />
          </div>
          <div>
            <label className="label">TTL (hours)</label>
            <input type="number" min={1} max={720} value={ttl} onChange={(e) => setTtl(parseInt(e.target.value) || 24)} className="input" />
          </div>
          {err && <div className="text-red-400 text-sm">{err}</div>}
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="btn-secondary text-sm">Cancel</button>
            <button type="submit" disabled={busy} className="btn-primary text-sm">{busy ? "Creating…" : "Create"}</button>
          </div>
        </form>
      </div>
    </div>
  );
}
