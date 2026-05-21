import { useEffect, useState } from "react";
import { useSearchParams, Link } from "react-router-dom";
import { api } from "../api/client";
import { AuthShell } from "./Login";

export default function VerifyEmailPage() {
  const [params] = useSearchParams();
  const token = params.get("token") || "";
  const [status, setStatus] = useState<"working" | "ok" | "error">("working");
  const [err, setErr] = useState<string>("");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setErr("Missing token");
      return;
    }
    api
      .verifyEmail(token)
      .then(() => setStatus("ok"))
      .catch((e) => {
        setStatus("error");
        setErr(e.message);
      });
  }, [token]);

  return (
    <AuthShell title="Verifying your email">
      {status === "working" && <p className="text-gray-400 text-sm text-center">Verifying…</p>}
      {status === "ok" && (
        <>
          <p className="text-green-400 text-sm text-center">Email verified.</p>
          <div className="mt-6 text-center">
            <Link to="/login" className="link">Sign in</Link>
          </div>
        </>
      )}
      {status === "error" && (
        <>
          <p className="text-red-400 text-sm text-center">{err}</p>
          <div className="mt-6 text-center">
            <Link to="/login" className="link">Back to sign in</Link>
          </div>
        </>
      )}
    </AuthShell>
  );
}
