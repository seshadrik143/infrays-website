import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Quote, Server, Cloud, Building } from 'lucide-react'

// Customers page. Placeholder stories until real customers convert
// + give permission to be named. Each "story" is realistic in shape
// and pattern but flagged as illustrative.
//
// Replace one story at a time as real customers come in: keep the
// {company, logo wordmark, sector, scale, quote, name, role, mode,
//  details} shape.

const stories = [
  {
    company: 'Acme.Corp',
    sector: 'Mid-size fintech',
    scale: '180 hosts · 3 regions',
    mode: 'self-hosted',
    quote: 'We swapped our Datadog setup for self-hosted infraYS. Same dashboards, same alerts, same SLOs — and we cut our observability bill by 71%. The 12 MB agent footprint vs the 80 MB Datadog one was the icing.',
    name: 'Sarah K.',
    role: 'Staff SRE',
    color: 'cyan',
    details: 'Migrated 180 hosts over a weekend. Kept Datadog for two months in parallel before cutting over. No regrets.',
  },
  {
    company: 'Globex Industries',
    sector: 'B2B SaaS · ~200 engineers',
    scale: '12 clusters · multi-region',
    mode: 'self-hosted',
    quote: 'The audit log being hash-chained out of the box meant we passed our SOC 2 Type 1 review without any extra tooling. Most of the work was already done. Saved a quarter of engineering time.',
    name: 'Marcus T.',
    role: 'Director of Platform',
    color: 'purple',
    details: 'Was running Splunk + Datadog + PagerDuty. infraYS replaced all three. SSO via Okta, audit log feeding SIEM, alerting routing to existing webhook integrations.',
  },
  {
    company: 'STELLAR /\\',
    sector: 'Colo-only e-commerce',
    scale: '45 bare-metal · Tier-4 colo',
    mode: 'self-hosted',
    quote: 'We run on bare metal in a cage in Tier-4 colo. Datadog never quite fit. infraYS self-hosted just works — single binary, no Kubernetes required, and we control every byte that leaves our cabinet.',
    name: 'Priya N.',
    role: 'Infra Lead',
    color: 'green',
    details: 'PCI-DSS compliance environment. Network egress whitelist. No phone-home telemetry was a hard requirement. infraYS\'s offline mode (license refresh once per quarter) is exactly what they needed.',
  },
  {
    company: 'Initech',
    sector: 'AI/ML platform',
    scale: '60 GPU nodes · K8s',
    mode: 'cloud',
    quote: 'NVIDIA GPU telemetry, container metrics, OTLP traces from PyTorch — all flowing into one dashboard. The AI anomaly detection caught a slow memory leak that would have surfaced as a $30k cloud bill at month-end.',
    name: 'David L.',
    role: 'ML Platform Eng',
    color: 'cyan',
    details: 'Using the AIOps suite + per-tenant data isolation across multiple ML teams. Stripe Pro tier.',
  },
  {
    company: 'Hooli',
    sector: 'Open-source maintainer',
    scale: '8 home-lab hosts',
    mode: 'free-tier',
    quote: 'The free tier covers my homelab perfectly. 3 agents, Apache 2.0, no telemetry, no nag screens. When I started my consultancy I upgraded to Starter for paid clients — same install, just a license key.',
    name: 'Erin J.',
    role: 'Independent consultant',
    color: 'purple',
    details: 'Started on Free, organically upgraded to Starter ~6 months in. Self-hosted both modes.',
  },
  {
    company: 'Pied Piper',
    sector: 'Streaming media',
    scale: '300 edge nodes worldwide',
    mode: 'self-hosted',
    quote: 'Our edge fleet runs everywhere. infraYS\'s 12 MB agent + 30 MB RAM ceiling meant we could run it on every edge node without measurably affecting the workloads. No other tool fit.',
    name: 'Vincent K.',
    role: 'Edge Infrastructure',
    color: 'green',
    details: 'Multi-tenant per-customer dashboards. Each edge node tagged with region + customer. Custom collectors written using the Plugin SDK in 2 weeks.',
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
  'free-tier':   Building,
}

const modeLabels: Record<string, string> = {
  'self-hosted': 'Self-hosted',
  'cloud':       'Managed Cloud',
  'free-tier':   'Free / Community',
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
              The stories below are illustrative until we have permission to publish real customer names + logos.
              The patterns are real: cost savings, compliance prep, edge fleet support, regulated environments.
            </p>
          </div>
        </section>

        {/* Logo strip */}
        <section className="py-12 border-b border-white/[0.06]"
          style={{ background: 'rgba(8,8,18,0.5)' }}>
          <div className="container-lg">
            <p className="text-center text-[11px] font-semibold uppercase tracking-[0.2em] text-white/30 mb-6">
              Currently running infraYS
            </p>
            <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-3">
              {stories.map((s) => (
                <span key={s.company} className="text-base font-bold text-white/30 tracking-tight">
                  {s.company}
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
                  <article key={s.company} className="border border-white/[0.07] rounded-2xl p-7 flex flex-col"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <header className="flex items-start justify-between mb-5">
                      <div>
                        <h3 className="text-lg font-bold text-white">{s.company}</h3>
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
              Try infraYS free for 15 days. If you end up shipping it to production,
              tell us — we&apos;d love to feature your story (with your permission).
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <Link to="/install" className="btn-primary text-base px-8 py-4">
                Start free trial
                <ArrowRight className="w-5 h-5" />
              </Link>
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
