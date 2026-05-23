import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Heart, Shield, Code2, Zap } from 'lucide-react'

// About page — story + principles + current state. Founder-led product,
// so the page is written in first-person plural without pretending
// there's a 50-person team. Honest > aspirational here.

const principles = [
  {
    icon: Code2,
    title: 'Open source first',
    body: 'The core is Apache 2.0. Enterprise modules are FSL-1.1 with a 2-year MIT change date. Self-hosting is a real option, not a fallback.',
  },
  {
    icon: Shield,
    title: 'No telemetry, no trackers',
    body: 'The website has zero third-party analytics. The self-hosted product never phones home for usage data. Your data stays yours.',
  },
  {
    icon: Zap,
    title: 'Boring is a feature',
    body: 'Single binary. No Kubernetes required. Works on a Raspberry Pi or a bare-metal cabinet. Most of the complexity in observability is accidental — we cut it.',
  },
  {
    icon: Heart,
    title: 'No dark patterns',
    body: 'No bait pricing. No upsell screens. No "are you sure?" friction on cancel. Cancel from your portal in two clicks. Refund within 14 days, no questions.',
  },
]

const timeline = [
  { date: '2026 Q4', label: 'First commit. CPU + memory collector. Single binary.', kind: 'shipped' },
  { date: '2027 Q1', label: 'Logs + traces + distributed profiling. Plugin SDK v1.', kind: 'shipped' },
  { date: '2027 Q2', label: 'Enterprise modules: SSO, RBAC, AES, Vault, GDPR, audit log.', kind: 'shipped' },
  { date: '2027 Q3', label: 'Open core launch — Apache 2.0 + FSL-1.1 split.', kind: 'shipped' },
  { date: '2027 Q4', label: 'Cloud licensing service at license.infrays.org.', kind: 'shipped' },
  { date: '2028 Q1', label: 'Hosted EU region. SOC 2 Type 1. Public bug-bounty.', kind: 'roadmap' },
]

export default function AboutPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-20 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">About us</span>
            <h1 className="text-5xl font-black tracking-tight mb-6">
              We&apos;re tired of the{' '}
              <span className="text-gradient-cyan">observability bill</span>
            </h1>
            <p className="text-lg text-white/50 max-w-2xl mx-auto leading-relaxed">
              infraYS started as a side project — one engineer who&apos;d spent
              years gluing Prometheus + Grafana + Loki + Tempo + Datadog +
              an APM tool + an on-call rotation together. We thought:
              this could be one binary.
            </p>
          </div>
        </section>

        {/* The story */}
        <section className="section py-16 border-b border-white/[0.06]">
          <div className="container-md space-y-6">
            <h2 className="text-3xl font-black text-white">The short version</h2>
            <p className="text-white/60 leading-relaxed">
              Modern infrastructure monitoring is fragmented. The typical team runs five to eight separate tools — each with its own agent, its own storage, its own alerting, its own billing. The agents alone consume hundreds of MB of memory per host. The bill from Datadog or Splunk grows linearly with infrastructure, and they know it.
            </p>
            <p className="text-white/60 leading-relaxed">
              We started infraYS with the boring hypothesis that <strong className="text-white">one binary could do almost all of it</strong>, at roughly 5% of the cost. After 34 development phases, that hypothesis turned out to be true.
            </p>
            <p className="text-white/60 leading-relaxed">
              We are not VC-backed (yet). We are not trying to be the next CrowdStrike. We&apos;re building a tool we want to use ourselves, charging fairly for what it costs to run, and keeping the option to self-host so customers in regulated industries or air-gapped environments aren&apos;t locked out.
            </p>
          </div>
        </section>

        {/* Principles */}
        <section className="section py-16 border-b border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-lg">
            <h2 className="text-3xl font-black text-white text-center mb-12">What we believe</h2>
            <div className="grid md:grid-cols-2 gap-5">
              {principles.map((p) => {
                const Icon = p.icon
                return (
                  <div key={p.title} className="border border-white/[0.07] rounded-2xl p-6"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <div className="w-10 h-10 rounded-xl flex items-center justify-center mb-4 bg-cyan-500/10 border border-cyan-500/20">
                      <Icon className="w-5 h-5 text-cyan-400" />
                    </div>
                    <h3 className="text-base font-bold text-white mb-3">{p.title}</h3>
                    <p className="text-sm text-white/60 leading-relaxed">{p.body}</p>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* Timeline */}
        <section className="section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <h2 className="text-3xl font-black text-white text-center mb-12">Where we are</h2>
            <ol className="space-y-5">
              {timeline.map((entry) => (
                <li key={entry.date} className="flex gap-5 items-start">
                  <div className="flex flex-col items-center flex-shrink-0">
                    <div className={`w-3 h-3 rounded-full ring-4 ${entry.kind === 'shipped' ? 'bg-green-400 ring-green-400/20' : 'bg-cyan-400 ring-cyan-400/20'}`} />
                    <div className="w-px flex-1 bg-white/10 mt-1 min-h-[20px]" />
                  </div>
                  <div className="pb-4">
                    <div className="text-xs font-mono text-white/30 mb-1">{entry.date}</div>
                    <div className="text-sm text-white/70">{entry.label}</div>
                    {entry.kind === 'roadmap' && (
                      <span className="inline-block mt-1 text-[10px] uppercase tracking-wider text-cyan-400 font-semibold">Roadmap</span>
                    )}
                  </div>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* Honesty */}
        <section className="section py-16 border-b border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md">
            <h2 className="text-3xl font-black text-white mb-6">What we&apos;re not</h2>
            <ul className="text-white/60 text-sm space-y-3 leading-relaxed">
              <li><strong className="text-white">We&apos;re not a Datadog killer.</strong> Datadog is great for what it does. If you have unlimited budget and want every feature out of the box, use Datadog. We compete on cost, footprint, and self-hostability.</li>
              <li><strong className="text-white">We&apos;re not finished.</strong> v0.33 is a real, complete platform — but we&apos;re not at v1.0 yet. We don&apos;t pretend to be enterprise-grade in places we&apos;re not. The <Link to="/security" className="text-cyan-400 hover:underline">Security page</Link> is honest about what compliance we have and don&apos;t.</li>
              <li><strong className="text-white">We&apos;re not VC-funded.</strong> This is a deliberately patient project. Costs are low because the team is small and the infrastructure is efficient. That&apos;s the moat.</li>
              <li><strong className="text-white">We&apos;re not for everyone.</strong> If you want hand-holding, dedicated CSMs, and quarterly business reviews, the Enterprise tier (with paid support) is the right fit. The Free / Starter tiers are for teams who like reading docs.</li>
            </ul>
          </div>
        </section>

        {/* CTA */}
        <section className="section py-16">
          <div className="container-md text-center">
            <h2 className="text-3xl font-black text-white mb-4">Try the product</h2>
            <p className="text-white/50 mb-8">
              The best way to evaluate infraYS is to deploy it.
              15 days free, no credit card.
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
