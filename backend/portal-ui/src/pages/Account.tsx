import { useEffect, useState } from "react";
import { api } from "../api/client";
import { useAuth } from "../lib/auth";

export default function AccountPage() {
  const { me, setMe } = useAuth();
  const [name, setName] = useState(me?.name || "");
  const [company, setCompany] = useState(me?.company || "");
  const [profileSaved, setProfileSaved] = useState(false);
  const [profileErr, setProfileErr] = useState<string | null>(null);

  useEffect(() => {
    setName(me?.name || "");
    setCompany(me?.company || "");
  }, [me]);

  const saveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileErr(null);
    setProfileSaved(false);
    try {
      const r = await api.updateAccount({ name, company });
      setMe(r);
      setProfileSaved(true);
    } catch (e: any) {
      setProfileErr(e.message);
    }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Account</h1>

      <div className="card">
        <h2 className="font-medium mb-4">Profile</h2>
        <form onSubmit={saveProfile} className="space-y-3">
          <div>
            <label className="label">Email</label>
            <input value={me?.email || ""} disabled className="input opacity-60" />
          </div>
          <div>
            <label className="label">Name</label>
            <input value={name} onChange={(e) => setName(e.target.value)} className="input" />
          </div>
          <div>
            <label className="label">Company</label>
            <input value={company} onChange={(e) => setCompany(e.target.value)} className="input" />
          </div>
          {profileErr && <div className="text-red-400 text-sm">{profileErr}</div>}
          {profileSaved && <div className="text-green-400 text-sm">Saved.</div>}
          <button type="submit" className="btn-primary text-sm">Save</button>
        </form>
      </div>

      <ChangePasswordCard />
    </div>
  );
}

function ChangePasswordCard() {
  const [oldPw, setOldPw] = useState("");
  const [newPw, setNewPw] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [ok, setOk] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setOk(false);
    setBusy(true);
    try {
      await api.changePassword({ old_password: oldPw, new_password: newPw });
      setOk(true);
      setOldPw("");
      setNewPw("");
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="card">
      <h2 className="font-medium mb-4">Change password</h2>
      <form onSubmit={submit} className="space-y-3">
        <div>
          <label className="label">Current password</label>
          <input type="password" required value={oldPw} onChange={(e) => setOldPw(e.target.value)} className="input" />
        </div>
        <div>
          <label className="label">New password</label>
          <input type="password" required minLength={8} value={newPw} onChange={(e) => setNewPw(e.target.value)} className="input" />
        </div>
        {err && <div className="text-red-400 text-sm">{err}</div>}
        {ok && <div className="text-green-400 text-sm">Password updated. Other sessions have been signed out.</div>}
        <button type="submit" disabled={busy} className="btn-primary text-sm">{busy ? "Saving…" : "Change password"}</button>
      </form>
    </div>
  );
}
