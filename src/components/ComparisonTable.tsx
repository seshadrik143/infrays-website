import { Check, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

const features = [
  { label: 'Open Source (MIT License)', desc: 'Full source code available, no license key' },
  { label: 'Single Agent Binary', desc: 'One 12MB binary, no runtime dependencies' },
  { label: 'Metrics + Logs + Traces + Profiles', desc: 'All four observability pillars in one agent' },
  { label: 'AI-Powered Anomaly Detection', desc: 'Z-score + Isolation Forest baseline analysis' },
  { label: 'Predictive Alerting', desc: 'Linear regression forecasts breaches up to 7 days ahead' },
  { label: 'Alert Flap Detection', desc: 'Suppresses noisy alerts with 5-min transition tracking' },
  { label: 'Alert Rule Library', desc: 'Browse, import, and customize pre-built alert rules' },
  { label: 'Alert Coverage Score', desc: 'Measure how well your infrastructure is monitored (0–100%)' },
  { label: 'Synthetic Monitoring', desc: 'HTTP, TCP, DNS, SSL, multi-step, browser checks' },
  { label: 'Public Status Page', desc: 'Auto-generated /status page from synthetic probe results' },
  { label: 'SLO / Error Budget Tracking', desc: 'Define, track, and report service levels' },
  { label: 'Continuous Profiling', desc: 'CPU/memory flame graphs, no instrumentation needed' },
  { label: 'Error Tracking', desc: 'SHA-256 fingerprint dedup, first/last seen, error groups' },
  { label: 'Auto-Discovery', desc: '/proc-based service scan + TCP flow map + cloud metadata' },
  { label: 'Service Topology Map', desc: 'Visualize service dependencies with latency and error rate' },
  { label: 'Multi-Tenancy', desc: 'Isolated data per tenant, per-team dashboards' },
  { label: 'On-Call Scheduling', desc: 'Built-in rotation, escalation, and notifications' },
  { label: 'Cloud Cost Tracking', desc: 'AWS, Azure, GCP cost import alongside metrics' },
  { label: 'Self-Hosted — Free Forever', desc: 'Run on your own infra with all features' },
  { label: '67 Community Plugins', desc: 'MySQL, Kafka, Cassandra, NGINX, Redis, and 60+ more' },
  { label: 'Custom Plugin SDK', desc: 'Write collectors in any language via exec protocol' },
  { label: '67 Dashboard Templates', desc: 'Pre-built dashboards for system, web, DB, cloud, SRE' },
  { label: 'Terraform / Pulumi Provider', desc: 'Infrastructure-as-code for your observability' },
  { label: 'OIDC / SSO Integration', desc: 'Google, GitHub, Microsoft, custom IdP' },
  { label: 'Granular RBAC', desc: '11 permissions, custom roles, session management' },
  { label: 'AES-256-GCM Encryption', desc: 'Secrets encrypted at rest, Vault/AWS SM/GCP SM integration' },
  { label: 'Compliance Reports', desc: 'SOC2, ISO27001, GDPR controls with PDF export' },
  { label: 'GDPR Erasure API', desc: 'Data subject erasure and configurable retention policies' },
  { label: 'Raft HA & Backup', desc: 'Leader election, failover, and streaming backup/restore' },
]

export default function ComparisonTable() {
  return (
    <section className="section relative overflow-hidden" id="features-list"
      style={{ background: 'rgba(6,6,14,0.5)' }}>
      <div className="container-lg relative z-10">
        {/* Header */}
        <div className="text-center mb-12">
          <span className="badge-green mb-5">What's Included</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Everything in{' '}
            <span className="text-gradient-cyan">one platform</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto leading-relaxed">
            infraYS ships with a complete observability stack out of the box —
            no extra tools, no plugins to buy, no surprise limits.
          </p>
        </div>

        {/* Feature checklist */}
        <div className="rounded-2xl overflow-hidden"
          style={{
            background: 'rgba(11,11,22,0.8)',
            border: '1px solid rgba(255,255,255,0.06)',
            backdropFilter: 'blur(10px)',
          }}>
          <div className="grid md:grid-cols-2">
            {features.map((f, i) => (
              <div key={f.label}
                className={`flex items-start gap-4 px-6 py-4 transition-colors hover:bg-white/[0.025]
                  ${i % 2 === 1 ? 'md:border-l border-white/[0.04]' : ''}
                  ${i < features.length - 2 ? 'border-b border-white/[0.04]' : ''}`}>
                <div className="w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5"
                  style={{ background: 'rgba(16,185,129,0.12)', border: '1px solid rgba(16,185,129,0.25)' }}>
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </div>
                <div>
                  <div className="text-sm font-semibold text-white/80">{f.label}</div>
                  <div className="text-xs text-white/30 mt-0.5 leading-relaxed">{f.desc}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Free forever note */}
        <div className="mt-6 rounded-2xl p-6 text-center"
          style={{
            background: 'rgba(16,185,129,0.04)',
            border: '1px solid rgba(16,185,129,0.15)',
          }}>
          <p className="text-sm text-white/55 leading-relaxed">
            <span className="text-green-400 font-bold">All features above are included when self-hosting.</span>{' '}
            infraYS is MIT licensed — download, deploy, and run it on your own infrastructure at no cost.
            Cloud plans add managed hosting and priority support.
          </p>
          <Link to="/pricing"
            className="inline-flex items-center gap-2 mt-4 text-sm text-cyan-400 hover:text-cyan-300 transition-colors font-medium">
            View pricing plans <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
