import { useEffect, useState } from "react";
import { adminApi, AdminUser } from "../../api/admin";
import { useAdminAuth } from "../../lib/adminAuth";
import { StatusPill } from "./AdminCustomerList";

const ROLES = ["admin", "support", "sales", "engineer"];

export default function AdminUsers() {
  const { me } = useAdminAuth();
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [newTemp, setNewTemp] = useState<{ email: string; password: string } | null>(null);

  const load = () => {
    adminApi.listAdmins().then((r) => setAdmins(r.admins || [])).catch((e) => setErr(e.message));
  };
  useEffect(() => { load(); }, []);

  const onlyAdmins = me?.role === "admin";

  const setStatus = async (a: AdminUser, status: "active" | "disabled") => {
    if (!confirm(`${status === "disabled" ? "Disable" : "Reactivate"} ${a.email}?`)) return;
    try {
      await adminApi.setAdminStatus(a.id, status);
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const changeRole = async (a: AdminUser) => {
    const role = prompt(`Role for ${a.email} (admin/support/sales/engineer):`, a.role);
    if (!role || !ROLES.includes(role)) return;
    try {
      await adminApi.updateAdminRole(a.id, role);
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const resetPassword = async (a: AdminUser) => {
    if (!confirm(`Reset password for ${a.email}? They'll be signed out and given a new temporary password.`)) return;
    try {
      const r = await adminApi.resetAdminPassword(a.id);
      setNewTemp({ email: r.email, password: r.temp_password });
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const deleteAdmin = async (a: AdminUser) => {
    if (!confirm(`Delete ${a.email} permanently? This cannot be undone.`)) return;
    try {
      await adminApi.deleteAdmin(a.id);
      load();
    } catch (e: any) {
      alert(e.message);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Admin Users</h1>
        {onlyAdmins && (
          <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">Invite admin</button>
        )}
      </div>
      {!onlyAdmins && (
        <div className="card text-sm text-gray-400">
          Only users with the <code>admin</code> role can manage other admins.
        </div>
      )}
      {err && <div className="text-red-400 text-sm">{err}</div>}

      {newTemp && (
        <div className="card border-accent-500">
          <div className="font-medium mb-2">Temporary password for {newTemp.email} — copy now, won't be shown again:</div>
          <code className="block bg-ink-900 p-3 rounded text-xs font-mono break-all">{newTemp.password}</code>
          <p className="text-xs text-gray-400 mt-2">
            Share this with the user securely (Slack DM, password manager). They'll be required to change it on first login.
          </p>
          <button onClick={() => setNewTemp(null)} className="btn-secondary text-sm mt-3">Dismiss</button>
        </div>
      )}

      <div className="card p-0 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-ink-900">
            <tr className="text-left text-xs text-gray-400">
              <th className="px-4 py-2">Email</th>
              <th className="px-4 py-2">Role</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2">MFA</th>
              <th className="px-4 py-2">Last login</th>
              {onlyAdmins && <th className="px-4 py-2 text-right">Actions</th>}
            </tr>
          </thead>
          <tbody>
            {admins.map((a) => {
              const isSelf = me && me.id === a.id;
              return (
                <tr key={a.id} className="border-t border-ink-700">
                  <td className="px-4 py-2">
                    {a.email} {isSelf && <span className="text-xs text-gray-500">(you)</span>}
                  </td>
                  <td className="px-4 py-2 capitalize">{a.role}</td>
                  <td className="px-4 py-2"><StatusPill status={a.status} /></td>
                  <td className="px-4 py-2 text-xs">
                    {a.mfa_enrolled ? <span className="text-green-400">Enrolled</span> : <span className="text-amber-300">Not enrolled</span>}
                    {a.must_change_password && <span className="ml-2 text-amber-300">· Must change PW</span>}
                  </td>
                  <td className="px-4 py-2 text-xs text-gray-500">
                    {a.last_login ? new Date(a.last_login).toLocaleString() : "—"}
                  </td>
                  {onlyAdmins && (
                    <td className="px-4 py-2 text-right">
                      {!isSelf && (
                        <div className="flex gap-2 justify-end text-xs">
                          <button onClick={() => changeRole(a)} className="text-gray-300 hover:underline">Role</button>
                          <button onClick={() => resetPassword(a)} className="text-gray-300 hover:underline">Reset PW</button>
                          {a.status === "active" ? (
                            <button onClick={() => setStatus(a, "disabled")} className="text-amber-300 hover:underline">Disable</button>
                          ) : (
                            <button onClick={() => setStatus(a, "active")} className="text-green-300 hover:underline">Reactivate</button>
                          )}
                          <button onClick={() => deleteAdmin(a)} className="text-red-400 hover:underline">Delete</button>
                        </div>
                      )}
                    </td>
                  )}
                </tr>
              );
            })}
            {admins.length === 0 && (
              <tr><td colSpan={onlyAdmins ? 6 : 5} className="px-4 py-8 text-center text-gray-400">No admins.</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {showCreate && (
        <CreateAdminModal
          onClose={() => setShowCreate(false)}
          onCreated={(email, temp) => {
            setNewTemp({ email, password: temp });
            setShowCreate(false);
            load();
          }}
        />
      )}
    </div>
  );
}

function CreateAdminModal({
  onClose, onCreated,
}: {
  onClose: () => void;
  onCreated: (email: string, temp: string) => void;
}) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("support");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      const r = await adminApi.createAdmin({ email, role });
      onCreated(r.email, r.temp_password);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-6 z-10">
      <div className="card w-full max-w-md">
        <h2 className="font-medium mb-4">Invite admin</h2>
        <form onSubmit={submit} className="space-y-3">
          <div>
            <label className="label">Email</label>
            <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="input" />
          </div>
          <div>
            <label className="label">Role</label>
            <select value={role} onChange={(e) => setRole(e.target.value)} className="input">
              {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
            </select>
          </div>
          <p className="text-xs text-gray-400">
            A one-time temporary password will be generated. Share it with the user securely.
          </p>
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
