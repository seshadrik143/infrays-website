import { useEffect, useState } from "react";
import { api, Subscription } from "../api/client";

// Plans shown in the self-serve picker. `tier` is the identifier sent
// to the backend — it MUST match a tier configured in the issuer's
// NP_STRIPE_PRICE_MAP (e.g. "professional"), otherwise checkout returns
// "unknown plan". Prices are display-only (Stripe is the source of
// truth); keep them in sync with docs/LICENSING.md and infrays.org/pricing.
type Plan = {
  tier: string;
  name: string;
  monthly: string;
  annual: string;
  blurb: string;
  highlight?: boolean;
};

const PLANS: Plan[] = [
  {
    tier: "starter",
    name: "Starter",
    monthly: "$49/mo",
    annual: "$39/mo billed annually",
    blurb: "For small teams getting started with NodePulse.",
  },
  {
    tier: "professional",
    name: "Pro",
    monthly: "$199/mo",
    annual: "$159/mo billed annually",
    blurb: "Production monitoring with higher limits and full features.",
    highlight: true,
  },
];

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

  // A customer still on the signup trial (no paid subscription, or a
  // subscription in "trialing" status) should be nudged to pick a plan.
  const trialing = subs.find((s) => s.status === "trialing");
  const hasPaid = subs.some((s) => s.status === "active" || s.status === "past_due");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Subscriptions</h1>
        <button onClick={openBilling} disabled={billingBusy} className="btn-secondary text-sm">
          {billingBusy ? "Opening…" : "Manage billing"}
        </button>
      </div>
      {err && <div className="text-red-400 text-sm">{err}</div>}

      {trialing?.trial_end && !isZero(trialing.trial_end) && (
        <div className="card border border-blue-800/60 bg-blue-950/30 text-sm text-blue-200">
          Your free trial ends {fmtDate(trialing.trial_end)}. Choose a plan below to keep
          NodePulse running without interruption.
        </div>
      )}

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

      {/* Plan picker: shown until the customer has a paid subscription.
          Existing paid customers change plans via "Manage billing". */}
      {!hasPaid && <PlanPicker onError={setErr} />}
    </div>
  );
}

function PlanPicker({ onError }: { onError: (msg: string) => void }) {
  const [interval, setInterval] = useState<"month" | "annual">("month");
  const [busyTier, setBusyTier] = useState<string | null>(null);

  const subscribe = async (tier: string) => {
    setBusyTier(tier);
    try {
      const r = await api.createCheckoutSession({ tier, interval });
      window.location.href = r.url;
    } catch (e: any) {
      onError(e.message);
      setBusyTier(null);
    }
  };

  return (
    <div className="space-y-4 pt-2">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">Choose a plan</h2>
        <div className="inline-flex rounded-md border border-gray-700 text-xs overflow-hidden">
          <button
            onClick={() => setInterval("month")}
            className={`px-3 py-1.5 ${interval === "month" ? "bg-gray-700 text-white" : "text-gray-400"}`}
          >
            Monthly
          </button>
          <button
            onClick={() => setInterval("annual")}
            className={`px-3 py-1.5 ${interval === "annual" ? "bg-gray-700 text-white" : "text-gray-400"}`}
          >
            Annual
          </button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {PLANS.map((p) => (
          <div
            key={p.tier}
            className={`card flex flex-col ${p.highlight ? "border border-indigo-600/70" : ""}`}
          >
            <div className="flex items-center justify-between">
              <div className="font-semibold">{p.name}</div>
              {p.highlight && (
                <span className="text-[10px] uppercase tracking-wide text-indigo-300">Popular</span>
              )}
            </div>
            <div className="mt-2 text-xl font-semibold">
              {interval === "month" ? p.monthly : p.annual.split(" ")[0] + "/mo"}
            </div>
            <div className="text-xs text-gray-400">
              {interval === "annual" ? "billed annually" : "billed monthly"}
            </div>
            <p className="mt-3 text-sm text-gray-400 flex-1">{p.blurb}</p>
            <button
              onClick={() => subscribe(p.tier)}
              disabled={busyTier !== null}
              className="btn-primary text-sm mt-4"
            >
              {busyTier === p.tier ? "Redirecting…" : `Subscribe to ${p.name}`}
            </button>
          </div>
        ))}
      </div>

      <div className="text-xs text-gray-500">
        Need higher limits, on-prem, or dedicated support?{" "}
        <a className="text-indigo-300 hover:underline" href="mailto:contact@infrays.org">
          Contact sales
        </a>{" "}
        about Enterprise.
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
