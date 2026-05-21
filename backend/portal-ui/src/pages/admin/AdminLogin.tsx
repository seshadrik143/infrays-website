import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi } from "../../api/admin";
import { useAdminAuth } from "../../lib/adminAuth";
import { AuthShell } from "../Login";

type Stage = "credentials" | "mfa" | "enroll";

export default function AdminLoginPage() {
  const [stage, setStage] = useState<Stage>("credentials");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [enrollSecret, setEnrollSecret] = useState("");
  const [enrollURL, setEnrollURL] = useState("");
  const { refresh } = useAdminAuth();
  const nav = useNavigate();

  const submitCredentials = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      const r = await adminApi.login({ email, password });
      if (!r.mfa_enrolled) {
        // First login → enroll MFA.
        const setup = await adminApi.mfaSetup();
        setEnrollSecret(setup.secret);
        setEnrollURL(setup.otpauth_url);
        setStage("enroll");
      } else {
        setStage("mfa");
      }
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  const submitMFA = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      await adminApi.mfaChallenge(code);
      await refresh();
      nav("/admin/dashboard", { replace: true });
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  const submitEnroll = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      await adminApi.mfaSetupVerify(code);
      await refresh();
      nav("/admin/dashboard", { replace: true });
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  };

  if (stage === "credentials") {
    return (
      <AuthShell title="infraYS — Admin">
        <form onSubmit={submitCredentials} className="space-y-4">
          <div>
            <label className="label">Email</label>
            <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="input" />
          </div>
          <div>
            <label className="label">Password</label>
            <input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} className="input" />
          </div>
          {err && <div className="text-red-400 text-sm">{err}</div>}
          <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "…" : "Continue"}</button>
        </form>
      </AuthShell>
    );
  }

  if (stage === "mfa") {
    return (
      <AuthShell title="Two-factor code">
        <form onSubmit={submitMFA} className="space-y-4">
          <p className="text-sm text-gray-400">Enter the 6-digit code from your authenticator app.</p>
          <div>
            <label className="label">Code</label>
            <input value={code} onChange={(e) => setCode(e.target.value)} className="input font-mono tracking-widest text-center text-lg" maxLength={6} pattern="[0-9]{6}" required />
          </div>
          {err && <div className="text-red-400 text-sm">{err}</div>}
          <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "…" : "Verify"}</button>
        </form>
      </AuthShell>
    );
  }

  // stage === "enroll"
  return (
    <AuthShell title="Set up two-factor">
      <form onSubmit={submitEnroll} className="space-y-4">
        <p className="text-sm text-gray-300">
          Scan this with Google Authenticator / 1Password / Authy:
        </p>
        <div className="card bg-ink-900">
          <div className="text-xs text-gray-400 mb-1">Secret (manual entry)</div>
          <code className="font-mono text-xs break-all">{enrollSecret}</code>
          <div className="text-xs text-gray-400 mt-3 mb-1">URI</div>
          <code className="font-mono text-xs break-all">{enrollURL}</code>
        </div>
        <div>
          <label className="label">Confirm code from app</label>
          <input value={code} onChange={(e) => setCode(e.target.value)} className="input font-mono tracking-widest text-center text-lg" maxLength={6} pattern="[0-9]{6}" required />
        </div>
        {err && <div className="text-red-400 text-sm">{err}</div>}
        <button type="submit" disabled={busy} className="btn-primary w-full">{busy ? "…" : "Enable MFA"}</button>
      </form>
    </AuthShell>
  );
}
