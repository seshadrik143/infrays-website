import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { AuthShell } from "./Login";

// Single page that handles both modes:
//   - no ?token=: shows the "send reset link" form
//   - ?token=...: shows the "set new password" form
export default function ResetPasswordPage() {
  const [params] = useSearchParams();
  const token = params.get("token");
  return token ? <ConfirmReset token={token} /> : <RequestReset />;
}

function RequestReset() {
  const [email, setEmail] = useState("");
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    try {
      await api.requestPasswordReset(email);
      setDone(true);
    } finally {
      setBusy(false);
    }
  };

  if (done) {
    return (
      <AuthShell title="Check your email">
        <p className="text-sm text-gray-300">
          If an account exists for <strong>{email}</strong>, we sent a password reset link.
        </p>
        <div className="mt-6 text-center">
          <Link to="/login" className="link text-sm">Back to sign in</Link>
        </div>
      </AuthShell>
    );
  }

  return (
    <AuthShell title="Reset your password">
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="label">Email</label>
          <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="input" />
        </div>
        <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "Sending…" : "Send reset link"}</button>
        <div className="text-center text-sm text-gray-400">
          <Link to="/login" className="link">Back to sign in</Link>
        </div>
      </form>
    </AuthShell>
  );
}

function ConfirmReset({ token }: { token: string }) {
  const [pw, setPw] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      await api.resetPassword({ token, new_password: pw });
      setDone(true);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  if (done) {
    return (
      <AuthShell title="Password reset">
        <p className="text-sm text-green-400 text-center">Your password has been updated.</p>
        <div className="mt-6 text-center">
          <Link to="/login" className="link">Sign in</Link>
        </div>
      </AuthShell>
    );
  }

  return (
    <AuthShell title="Set a new password">
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="label">New password</label>
          <input type="password" required minLength={8} value={pw} onChange={(e) => setPw(e.target.value)} className="input" />
        </div>
        {err && <div className="text-red-400 text-sm">{err}</div>}
        <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "Saving…" : "Set password"}</button>
      </form>
    </AuthShell>
  );
}
