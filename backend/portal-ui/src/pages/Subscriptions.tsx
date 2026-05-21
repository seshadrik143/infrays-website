import { useEffect, useState } from "react";
import { api, Subscription } from "../api/client";

export default function SubscriptionsPage() {
  const [subs, setSubs] = useState<Subscription[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [billingBusy, setBillingBusy] = useState(false);

  useEffect(() => {
    api.listSubscriptions().then((r) => setSubs(r.subscriptions || [])).catch((e) => setErr(e.message));
  }, []);

  const openBilling = async () => {
    setBillingBusy(true);
    try {
      const r = await api.billingPortalURL();
      window.location.href = r.url;
    } catch (e: any) {
      alert(e.message);
    } finally {
      setBillingBusy(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Subscriptions</h1>
        <button onClick={openBilling} disabled={billingBusy} className="btn-secondary text-sm">
          {billingBusy ? "Opening…" : "Manage billing"}
        </button>
      </div>
      {err && <div className="text-red-400 text-sm">{err}</div>}
      {subs.length === 0 && <div className="card text-gray-400 text-sm">No subscriptions yet.</div>}
      <div className="space-y-3">
        {subs.map((s) => (
          <div key={s.id} className="card">
            <div className="flex items-start justify-between">
              <div>
                <div className="font-medium capitalize">{s.tier}</div>
                <div className="text-xs text-gray-400 mt-1">
                  {s.manual_offline ? "Offline / sales-issued" : s.stripe_subscription_id || "—"}
                </div>
              </div>
              <Badge status={s.status} />
            </div>
            <div className="grid grid-cols-2 gap-2 mt-4 text-xs text-gray-400">
              <div>Period: {fmtDate(s.current_period_start)} → {fmtDate(s.current_period_end)}</div>
              {s.trial_end && !isZero(s.trial_end) && <div>Trial ends: {fmtDate(s.trial_end)}</div>}
              {s.cancel_at && !isZero(s.cancel_at) && <div>Cancels: {fmtDate(s.cancel_at)}</div>}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function Badge({ status }: { status: string }) {
  const color =
    status === "active" ? "bg-green-900/40 text-green-300" :
    status === "trialing" ? "bg-blue-900/40 text-blue-300" :
    status === "past_due" ? "bg-amber-900/40 text-amber-300" :
    status === "canceled" ? "bg-gray-700 text-gray-300" :
    "bg-red-900/40 text-red-300";
  return <span className={`inline-block px-2 py-1 rounded text-xs ${color}`}>{status}</span>;
}

export function fmtDate(iso: string) {
  if (!iso || isZero(iso)) return "—";
  return new Date(iso).toLocaleDateString();
}

export function isZero(iso: string) {
  return !iso || iso.startsWith("0001-01-01");
}
