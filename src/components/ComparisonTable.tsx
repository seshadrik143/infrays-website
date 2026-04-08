import { Check, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

const features = [
  { label: 'Open Source (MIT License)', desc: 'Full source code available, no license key' },
  { label: 'Single Agent Binary', desc: 'One 12MB binary, no runtime dependencies' },
  { label: 'Metrics + Logs + Traces + Profiles', desc: 'All four observability pillars in one agent' },
  { label: 'AI-Powered Anomaly Detection', desc: 'Z-score + Isolation Forest baseline analysis' },
  { label: 'Synthetic Monitoring', desc: 'HTTP, TCP, DNS, SSL, browser checks' },
  { label: 'SLO / Error Budget Tracking', desc: 'Define, track, and report service levels' },
  { label: 'Continuous Profiling', desc: 'CPU/memory flame graphs, no instrumentation needed' },
  { label: 'Multi-Tenancy', desc: 'Isolated data per tenant, per-team dashboards' },
  { label: 'On-Call Scheduling', desc: 'Built-in rotation, escalation, and notifications' },
  { label: 'Cloud Cost Tracking', desc: 'AWS, Azure, GCP cost import alongside metrics' },
  { label: 'Self-Hosted — Free Forever', desc: 'Run on your own infra with all features' },
  { label: 'Custom Plugin SDK', desc: 'Write collectors in any language' },
  { label: 'Terraform / Pulumi Provider', desc: 'Infrastructure-as-code for your observability' },
  { label: 'OIDC / SSO Integration', desc: 'Google, GitHub, Microsoft, custom IdP' },
  { label: 'Granular RBAC', desc: '11 permissions, custom roles, session management' },
  { label: 'Compliance Reports', desc: 'SOC2, ISO27001, GDPR controls built in' },
]

export default function ComparisonTable() {
  return (
    <section className="section" id="features-list">
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-12">
          <span className="badge-green mb-4">What's Included</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Everything in one platform
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto">
            infraYS ships with a complete observability stack out of the box —
            no extra tools, no plugins to buy, no surprise limits.
          </p>
        </div>

        {/* Feature checklist */}
        <div className="border border-white/[0.07] rounded-2xl overflow-hidden"
          style={{ background: 'rgba(17, 17, 32, 0.7)' }}>
          <div className="grid md:grid-cols-2">
            {features.map((f, i) => (
              <div key={f.label}
                className={`flex items-start gap-4 px-6 py-4 border-b border-white/[0.04] hover:bg-white/[0.02] transition-colors
                  ${i % 2 === 1 ? 'md:border-l border-white/[0.04]' : ''}
                  ${i >= features.length - 2 ? 'border-b-0' : ''}`}>
                <div className="w-6 h-6 rounded-full bg-green-500/15 border border-green-500/30 flex items-center justify-center flex-shrink-0 mt-0.5">
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </div>
                <div>
                  <div className="text-sm font-semibold text-white/85">{f.label}</div>
                  <div className="text-xs text-white/35 mt-0.5">{f.desc}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Free forever note */}
        <div className="mt-6 border border-green-500/20 rounded-2xl p-6 text-center"
          style={{ background: 'rgba(16, 185, 129, 0.04)' }}>
          <p className="text-sm text-white/60">
            <span className="text-green-400 font-bold">All features above are included when self-hosting.</span>{' '}
            infraYS is MIT licensed — download, deploy, and run it on your own infrastructure at no cost.
            Cloud plans add managed hosting and priority support.
          </p>
          <Link to="/pricing" className="inline-flex items-center gap-2 mt-4 text-sm text-cyan-400 hover:text-cyan-300 transition-colors font-medium">
            View pricing plans <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
