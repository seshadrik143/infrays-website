import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, Subscription, Deployment } from "../api/client";
import { useAuth } from "../lib/auth";

export default function DashboardPage() {
  const { me } = useAuth();
  const [subs, setSubs] = useState<Subscription[]>([]);
  const [deps, setDeps] = useState<Deployment[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.listSubscriptions(), api.listDeployments()])
      .then(([s, d]) => {
        setSubs(s.subscriptions || []);
        setDeps(d.deployments || []);
      })
      .catch((e) => setErr(e.message));
  }, []);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold">Welcome{me?.name ? `, ${me.name}` : ""}.</h1>
        <p className="text-gray-400 mt-1 text-sm">Manage your infraYS subscription and deployments.</p>
      </div>

      {err && <div className="text-red-400 text-sm">{err}</div>}

      <div className="grid grid-cols-3 gap-4">
        <StatCard title="Subscriptions" value={subs.length} link="/subscriptions" />
        <StatCard title="Deployments" value={deps.length} link="/deployments" />
        <StatCard
          title="Active tier"
          value={subs.find((s) => s.status === "active" || s.status === "trialing")?.tier || "—"}
        />
      </div>

      <div className="card">
        <h2 className="font-medium mb-3">Quick start</h2>
        <ol className="text-sm text-gray-300 space-y-2 list-decimal list-inside">
          <li><Link className="link" to="/subscriptions">Pick a subscription</Link></li>
          <li><Link className="link" to="/enrollment-tokens">Create an enrollment token</Link></li>
          <li>Paste it into your NodePulse server config under <code>license.enrollment_token</code></li>
          <li>Restart NodePulse — it will fetch a signed license from license.infrays.org</li>
        </ol>
      </div>
    </div>
  );
}

function StatCard({ title, value, link }: { title: string; value: string | number; link?: string }) {
  const inner = (
    <div className="card hover:border-accent-500 transition-colors h-full">
      <div className="text-xs text-gray-400 uppercase tracking-wider">{title}</div>
      <div className="text-3xl font-semibold mt-2">{value}</div>
    </div>
  );
  return link ? <Link to={link}>{inner}</Link> : inner;
}
