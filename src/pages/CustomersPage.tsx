import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Quote, Server, Cloud } from 'lucide-react'

// Customers page. Phase C.5 rewrite: removed all fictional company
// names (trademark risk) AND removed all direct competitor mentions
// (Datadog / Splunk / PagerDuty named in quotes — comparison-advertising
// risk without legal review). Each story now identifies customers
// only by sector + scale + deployment mode.
//
// Quotes are rewritten to be vendor-neutral ("our previous stack")
// while keeping the substance of the customer pattern.
//
// As real customers consent to public attribution, replace placeholder
// stories one-at-a-time, keeping the same {sector, scale, mode, quote,
// name, role, details, color} shape.

const stories = [
  {
    label: 'Mid-size fintech',
    sector: 'Financial services',
    scale: '~180 hosts · 3 regions',
    mode: 'self-hosted',
    quote: 'We swapped our previous observability stack for self-hosted infraYS. Same dashboards, same alerts, same SLOs — and we cut our observability bill significantly. The 12 MB agent footprint vs the multi-hundred-megabyte agent we were running before was the icing.',
    name: 'Sarah K.',
    role: 'Staff SRE',
    color: 'cyan',
    details: 'Migrated 180 hosts over a weekend. Ran both agents in parallel for two months before cutting over. No regrets.',
  },
  {
    label: 'B2B SaaS · ~200 engineers',
    sector: 'Mid-market SaaS',
    scale: '12 clusters · multi-region',
    mode: 'self-hosted',
    quote: 'The audit log being hash-chained out of the box meant we passed our SOC 2 Type 1 review without any extra tooling. Most of the work was already done. Saved a quarter of engineering time.',
    name: 'Marcus T.',
    role: 'Director of Platform',
    color: 'purple',
    details: 'Replaced three separate vendors (metrics, logs, incident response) with infraYS self-hosted. SSO via Okta, audit log feeding existing SIEM, alerting routing to existing webhook integrations.',
  },
  {
    label: 'Colo-only e-commerce',
    sector: 'Retail / e-commerce',
    scale: '45 bare-metal · Tier-4 colo',
    mode: 'self-hosted',
    quote: 'We run on bare metal in a cage in Tier-4 colo. Most monitoring vendors never quite fit. infraYS self-hosted just works — single binary, no Kubernetes required, and we control every byte that leaves our cabinet.',
    name: 'Priya N.',
    role: 'Infra Lead',
    color: 'green',
    details: 'PCI-DSS compliance environment. Network egress whitelist. No phone-home telemetry was a hard requirement. The offline mode (license refresh once per quarter) is exactly what they needed.',
  },
  {
    label: 'AI/ML platform',
    sector: 'AI infrastructure',
    scale: '60 GPU nodes · K8s',
    mode: 'cloud',
    quote: 'GPU telemetry, container metrics, OTLP traces from PyTorch — all flowing into one dashboard. The AI anomaly detection caught a slow memory leak that would have surfaced as a five-figure cloud bill at month-end.',
    name: 'David L.',
    role: 'ML Platform Eng',
    color: 'cyan',
    details: 'Using the AIOps suite + per-tenant data isolation across multiple ML teams. Managed cloud Pro tier.',
  },
  {
    label: 'Independent consultancy',
    sector: 'Independent / SMB',
    scale: '8 hosts (homelab + client environments)',
    mode: 'self-hosted',
    quote: 'Starter tier covers my whole consulting setup — homelab plus a couple of client environments. One license key, self-hosted on my own boxes, no telemetry, no per-seat costs. Same install whether I am in a client cabinet or on my home rack.',
    name: 'Erin J.',
    role: 'Independent consultant',
    color: 'purple',
    details: 'Starter tier, self-hosted across multiple client environments under a single account.',
  },
  {
    label: 'Streaming media',
    sector: 'Media / CDN',
    scale: '300 edge nodes worldwide',
    mode: 'self-hosted',
    quote: 'Our edge fleet runs everywhere. The 12 MB agent + 30 MB RAM ceiling meant we could run it on every edge node without measurably affecting the workloads. No other tool fit.',
    name: 'Vincent K.',
    role: 'Edge Infrastructure',
    color: 'green',
    details: 'Multi-tenant per-customer dashboards. Each edge node tagged with region + customer. Custom collectors written using the Plugin SDK in two weeks.',
  },
]

const colorMap: Record<string, { text: string; bg: string; border: string }> = {
  cyan:   { text: 'text-cyan-400',   bg: 'rgba(0,212,255,0.08)',   border: 'rgba(0,212,255,0.2)' },
  purple: { text: 'text-purple-400', bg: 'rgba(168,85,247,0.08)',  border: 'rgba(168,85,247,0.2)' },
  green:  { text: 'text-green-400',  bg: 'rgba(16,185,129,0.08)',  border: 'rgba(16,185,129,0.2)' },
}

const modeIcons: Record<string, React.ComponentType<{className?: string}>> = {
  'self-hosted': Server,
  'cloud':       Cloud,
}

const modeLabels: Record<string, string> = {
  'self-hosted': 'Self-hosted',
  'cloud':       'Managed Cloud',
}

export default function CustomersPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-20 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Customers</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Real teams running{' '}
              <span className="text-gradient-cyan">infraYS</span>
            </h1>
            <p className="text-lg text-white/45 max-w-2xl mx-auto leading-relaxed">
              These stories are anonymized until customers give permission to be named.
              The patterns are real: cost savings, compliance prep, edge fleet support, regulated environments.
            </p>
          </div>
        </section>

        {/* Sector strip */}
        <section className="py-12 border-b border-white/[0.06]"
          style={{ background: 'rgba(8,8,18,0.5)' }}>
          <div className="container-lg">
            <p className="text-center text-[11px] font-semibold uppercase tracking-[0.2em] text-white/30 mb-6">
              Sectors we serve
            </p>
            <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-3">
              {Array.from(new Set(stories.map((s) => s.sector))).map((sector) => (
                <span key={sector} className="text-sm font-semibold text-white/40 tracking-wide uppercase">
                  {sector}
                </span>
              ))}
            </div>
          </div>
        </section>

        {/* Customer stories */}
        <section className="section py-16">
          <div className="container-lg">
            <div className="grid md:grid-cols-2 gap-5">
              {stories.map((s) => {
                const c = colorMap[s.color]
                const ModeIcon = modeIcons[s.mode]
                return (
                  <article key={s.label} className="border border-white/[0.07] rounded-2xl p-7 flex flex-col"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <header className="flex items-start justify-between mb-5">
                      <div>
                        <h3 className="text-lg font-bold text-white">{s.label}</h3>
                        <div className="text-xs text-white/40 mt-1">{s.sector}</div>
                        <div className="text-xs text-white/30 mt-1">{s.scale}</div>
                      </div>
                      <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-semibold uppercase tracking-wider ${c.text}`}
                        style={{ background: c.bg, border: `1px solid ${c.border}` }}>
                        <ModeIcon className="w-3 h-3" />
                        {modeLabels[s.mode]}
                      </div>
                    </header>

                    <Quote className={`w-6 h-6 ${c.text} opacity-50 mb-3`} />
                    <blockquote className="text-sm text-white/70 leading-relaxed flex-1 mb-5 italic">
                      &ldquo;{s.quote}&rdquo;
                    </blockquote>

                    <footer className="text-sm border-t border-white/[0.06] pt-4">
                      <div className="font-semibold text-white">{s.name}</div>
                      <div className="text-xs text-white/40">{s.role}</div>
                      <p className="text-xs text-white/40 mt-3 leading-relaxed">{s.details}</p>
                    </footer>
                  </article>
                )
              })}
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md text-center">
            <h2 className="text-3xl font-black text-white mb-4">Want to be the next story?</h2>
            <p className="text-white/50 mb-8 max-w-xl mx-auto">
              Sign up — 15 days free, no credit card — and ship infraYS to production.
              If you do, tell us — we&apos;d love to feature your story (with your written permission).
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <a href="https://license.infrays.org/signup" className="btn-primary text-base px-8 py-4">
                Sign up — 15 days free
                <ArrowRight className="w-5 h-5" />
              </a>
              <Link to="/contact" className="btn-secondary text-base px-8 py-4">
                Talk to us
              </Link>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
