import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../lib/auth";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const { setMe } = useAuth();
  const nav = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      const me = await api.login({ email, password });
      setMe(me);
      nav("/dashboard", { replace: true });
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthShell title="Sign in to infraYS">
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="label">Email</label>
          <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="input" />
        </div>
        <div>
          <label className="label">Password</label>
          <input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} className="input" />
        </div>
        {err && <div className="text-red-400 text-sm">{err}</div>}
        <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "Signing in…" : "Sign in"}</button>
        <div className="text-center text-sm text-gray-400">
          <Link to="/reset-password" className="link">Forgot password?</Link>
          {" · "}
          <Link to="/signup" className="link">Create account</Link>
        </div>
      </form>
    </AuthShell>
  );
}

export function AuthShell({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="min-h-full flex items-center justify-center p-6">
      <div className="card w-full max-w-md">
        <h1 className="text-xl font-semibold mb-6 text-center">{title}</h1>
        {children}
      </div>
    </div>
  );
}
