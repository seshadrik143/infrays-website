import { useEffect, useState } from "react";
import { api, License } from "../api/client";
import { fmtDate } from "./Subscriptions";

export default function LicensesPage() {
  const [items, setItems] = useState<License[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.listLicenses().then((r) => setItems(r.licenses || [])).catch((e) => setErr(e.message));
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Licenses</h1>
      <p className="text-sm text-gray-400">Active licenses issued to your deployments.</p>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      {items.length === 0 && (
        <div className="card text-gray-400 text-sm">
          No licenses yet. They are minted automatically when a deployment redeems an enrollment token.
        </div>
      )}
      <div className="space-y-3">
        {items.map((l) => (
          <div key={l.jti} className="card">
            <div className="flex items-start justify-between">
              <div>
                <div className="font-medium capitalize">{l.tier}</div>
                <div className="text-xs text-gray-500 mt-1 font-mono">license: {l.license_id}</div>
                <div className="text-xs text-gray-500 font-mono">deployment: {l.deployment_id}</div>
              </div>
              <div className="flex items-center gap-2">
                {l.revoked && <span className="px-2 py-1 rounded text-xs bg-red-900/40 text-red-300">Revoked</span>}
                <a
                  href={api.offlineLicenseURL(l.license_id)}
                  className="btn-secondary text-xs"
                  download
                >
                  Download JWS
                </a>
              </div>
            </div>
            <div className="grid grid-cols-3 gap-2 mt-3 text-xs text-gray-400">
              <div>Issued: {fmtDate(l.issued_at)}</div>
              <div>Expires: {fmtDate(l.expires_at)}</div>
              <div>Grace until: {fmtDate(l.grace_until)}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
