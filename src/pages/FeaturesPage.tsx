
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import FeaturesSection from '@/components/FeaturesSection'
import CTABanner from '@/components/CTABanner'
import HowItWorks from '@/components/HowItWorks'

const pillars = [
  {
    title: 'Observe',
    desc: 'Metrics · Logs · Traces · Profiles',
    detail: 'The four pillars of observability, unified in a single agent and dashboard. No more context-switching between tools.',
    color: 'cyan',
  },
  {
    title: 'Alert',
    desc: 'Intelligent · Predictive · On-Call',
    detail: 'AI-powered alerting that knows your baseline, suppresses flaps, correlates incidents, and routes to the right person.',
    color: 'purple',
  },
  {
    title: 'Act',
    desc: 'Incidents · SLOs · Automation',
    detail: 'Close the loop from detection to resolution. SLO burn rates, incident timelines, and IaC-ready automation.',
    color: 'green',
  },
  {
    title: 'Scale',
    desc: 'HA · Multi-tenant · Enterprise',
    detail: 'Raft consensus, per-tenant isolation, OIDC SSO, RBAC, compliance reports. Built for the largest infrastructure teams.',
    color: 'orange',
  },
]

const colorMap: Record<string, string> = {
  cyan:   'from-cyan-500/20 to-cyan-500/0 border-cyan-500/20 text-cyan-400',
  purple: 'from-purple-500/20 to-purple-500/0 border-purple-500/20 text-purple-400',
  green:  'from-green-500/20 to-green-500/0 border-green-500/20 text-green-400',
  orange: 'from-orange-500/20 to-orange-500/0 border-orange-500/20 text-orange-400',
}

export default function FeaturesPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Platform Features</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              The complete{' '}
              <span className="text-gradient-cyan">observability stack</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              infraYS was built from the ground up to replace a full monitoring stack — not add to it.
              Everything you need, nothing you don't.
            </p>
          </div>
        </section>

        {/* 4 Pillars */}
        <section className="section py-16 border-b border-white/[0.06]"
          style={{ background: 'rgba(10,10,20,0.5)' }}>
          <div className="container-lg">
            <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-5">
              {pillars.map((p) => {
                const c = colorMap[p.color]
                return (
                  <div key={p.title}
                    className={`border rounded-2xl p-6 bg-gradient-to-b ${c}`}
                    style={{ background: 'rgba(17,17,32,0.8)' }}>
                    <h3 className={`text-2xl font-black mb-1 ${c.split(' ').pop()}`}>{p.title}</h3>
                    <p className="text-xs text-white/50 font-medium mb-4">{p.desc}</p>
                    <p className="text-sm text-white/40 leading-relaxed">{p.detail}</p>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        <FeaturesSection />
        <HowItWorks />
        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
