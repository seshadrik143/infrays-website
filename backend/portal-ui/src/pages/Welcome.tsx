import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";

// Welcome is the post-checkout landing page. Stripe redirects here as
// <AppURL>/welcome?session_id=cs_... after a successful payment. The
// subscription is created asynchronously by the customer.subscription
// webhook, so we poll listSubscriptions until a paid subscription
// shows up (or we give up and tell the user it's processing).
export default function WelcomePage() {
  const [state, setState] = useState<"waiting" | "ready" | "slow">("waiting");

  useEffect(() => {
    let cancelled = false;
    let tries = 0;
    const poll = async () => {
      tries++;
      try {
        const r = await api.listSubscriptions();
        const paid = (r.subscriptions || []).some(
          (s) => s.status === "active" || s.status === "trialing" || s.status === "past_due"
        );
        if (paid) {
          if (!cancelled) setState("ready");
          return;
        }
      } catch {
        // ignore transient errors while the webhook lands
      }
      if (cancelled) return;
      if (tries >= 10) {
        setState("slow");
        return;
      }
      setTimeout(poll, 2000);
    };
    poll();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="max-w-xl mx-auto py-16 px-4 text-center space-y-6">
      <div className="text-4xl">🎉</div>
      <h1 className="text-2xl font-semibold">Subscription active</h1>

      {state === "waiting" && (
        <p className="text-gray-400">
          Thanks for subscribing! We're finalizing your account — this usually takes a few
          seconds.
        </p>
      )}
      {state === "ready" && (
        <p className="text-gray-400">
          Your subscription is live. Next, create an enrollment token to connect a deployment.
        </p>
      )}
      {state === "slow" && (
        <p className="text-gray-400">
          Payment received. Your subscription is still processing — it should appear on the
          Subscriptions page shortly. No further action is needed.
        </p>
      )}

      <div className="flex items-center justify-center gap-3">
        <Link to="/enrollment-tokens" className="btn-primary text-sm">
          Create enrollment token
        </Link>
        <Link to="/subscriptions" className="btn-secondary text-sm">
          View subscriptions
        </Link>
      </div>
    </div>
  );
}
