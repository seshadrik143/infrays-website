import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { AuthShell } from "./Login";

export default function SignupPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [company, setCompany] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      await api.signup({ email, password, name, company });
      setDone(true);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  if (done) {
    return (
      <AuthShell title="Check your email">
        <p className="text-sm text-gray-300">
          We sent a verification link to <strong>{email}</strong>. Click it to activate your account.
        </p>
        <div className="mt-6 text-center">
          <Link to="/login" className="link text-sm">Back to sign in</Link>
        </div>
      </AuthShell>
    );
  }

  return (
    <AuthShell title="Create your infraYS account">
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="label">Email</label>
          <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="input" />
        </div>
        <div>
          <label className="label">Password</label>
          <input type="password" required minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} className="input" />
          <p className="text-xs text-gray-500 mt-1">At least 8 characters.</p>
        </div>
        <div>
          <label className="label">Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} className="input" />
        </div>
        <div>
          <label className="label">Company</label>
          <input value={company} onChange={(e) => setCompany(e.target.value)} className="input" />
        </div>
        {err && <div className="text-red-400 text-sm">{err}</div>}
        <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "Creating…" : "Create account"}</button>
        <div className="text-center text-sm text-gray-400">
          Already have an account? <Link to="/login" className="link">Sign in</Link>
        </div>
      </form>
    </AuthShell>
  );
}
