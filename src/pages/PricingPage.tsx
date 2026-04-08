
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import PricingSection from '@/components/PricingSection'
import CTABanner from '@/components/CTABanner'
import { Check } from 'lucide-react'

const faq = [
  {
    q: 'Is infraYS really free for self-hosted?',
    a: 'Yes. The entire codebase is MIT licensed. Self-hosting gives you all features — metrics, logs, traces, profiling, AI, multi-tenancy, RBAC — with no license key or call-home required. The paid cloud plans are only for our managed hosting service.',
  },
  {
    q: 'How are metrics counted?',
    a: 'A metric is a unique time-series data point sent in a 1-minute window. System CPU = 1 metric. If you have 8 cores, that\'s 8 metrics. Our Free tier is generous for personal use; Starter covers most small teams comfortably.',
  },
  {
    q: 'Can I upgrade or downgrade mid-month?',
    a: 'Yes. Upgrades are prorated and take effect immediately. Downgrades take effect at the end of your billing period. No penalties, no questions asked.',
  },
  {
    q: 'What payment methods do you accept?',
    a: 'All major credit cards via Stripe. Enterprise customers can pay via invoice/wire transfer on annual contracts.',
  },
  {
    q: 'What happens when I exceed my quota?',
    a: 'We soft-throttle — metrics above quota are sampled rather than dropped entirely. You\'ll get an email warning and can upgrade at any time. We never silently drop your data.',
  },
  {
    q: 'Do you offer discounts for startups or open-source projects?',
    a: 'Yes. OSS projects with a public repository get 50% off Pro. Startups under $1M ARR get 30% off. Contact us at hello@infrays.dev.',
  },
]

export default function PricingPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Header */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-green mb-4">Pricing</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Pay for what you use.{' '}
              <span className="text-gradient-cyan">Nothing more.</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              Transparent pricing, no hidden fees, no per-seat costs.
              Self-host forever free with the MIT-licensed codebase.
            </p>
          </div>
        </section>

        <PricingSection />

        {/* Feature comparison */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(10,10,20,0.5)' }}>
          <div className="container-lg">
            <h2 className="text-2xl font-black text-center text-white mb-10">Full plan comparison</h2>
            <div className="overflow-x-auto rounded-2xl border border-white/[0.07]"
              style={{ background: 'rgba(17,17,32,0.7)' }}>
              <table className="w-full min-w-[600px] text-sm">
                <thead>
                  <tr className="border-b border-white/[0.07]">
                    <th className="text-left px-6 py-4 text-white/30 font-medium w-[40%]">Feature</th>
                    {['Free', 'Starter', 'Pro', 'Enterprise'].map((t) => (
                      <th key={t} className="px-4 py-4 text-center font-bold text-white/70 border-l border-white/[0.05]">{t}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {[
                    ['Agents', '3', '25', 'Unlimited', 'Unlimited'],
                    ['Metrics/month', '1M', '50M', 'Unlimited', 'Unlimited'],
                    ['Log storage', '5GB', '50GB', '500GB', 'Custom'],
                    ['Trace storage', '1GB', '10GB', '100GB', 'Custom'],
                    ['Data retention', '7 days', '30 days', '90 days', 'Custom'],
                    ['Collectors', '✓ All', '✓ All', '✓ All', '✓ All'],
                    ['Custom collectors', '✓', '✓', '✓', '✓'],
                    ['Alerting', '✓ Basic', '✓ Full', '✓ AI-powered', '✓ AI-powered'],
                    ['Synthetic monitoring', '—', '10 checks', 'Unlimited', 'Unlimited'],
                    ['Continuous profiling', '—', '—', '✓', '✓'],
                    ['AI / AIOps', '—', 'Basic', '✓ Full', '✓ Full'],
                    ['SLO tracking', '—', '✓', '✓', '✓'],
                    ['Multi-tenancy', '—', '—', '✓', '✓'],
                    ['RBAC', '2 roles', '3 roles', 'Custom roles', 'Custom roles'],
                    ['SSO / OIDC', '—', '—', '✓', '✓'],
                    ['Compliance reports', '—', '—', '✓', '✓'],
                    ['On-premises deploy', '—', '—', '—', '✓'],
                    ['Support', 'Community', 'Email + Slack', 'Priority (4h)', 'Dedicated (1h)'],
                  ].map(([feature, ...vals]) => (
                    <tr key={feature} className="border-b border-white/[0.04] hover:bg-white/[0.02]">
                      <td className="px-6 py-3 text-white/60">{feature}</td>
                      {vals.map((val, i) => (
                        <td key={i} className="px-4 py-3 text-center text-white/50 border-l border-white/[0.04]">
                          {val === '✓' ? <Check className="w-4 h-4 text-green-400 mx-auto" /> :
                           val === '—' ? <span className="text-white/15">—</span> :
                           val.startsWith('✓') ? <span className="text-green-400 text-xs">{val}</span> :
                           <span className="text-xs">{val}</span>}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        {/* FAQ */}
        <section className="section py-16">
          <div className="container-md">
            <h2 className="text-3xl font-black text-center text-white mb-12">Frequently asked questions</h2>
            <div className="space-y-4">
              {faq.map((item) => (
                <div key={item.q} className="border border-white/[0.07] rounded-xl p-6"
                  style={{ background: 'rgba(17,17,32,0.6)' }}>
                  <h3 className="text-sm font-bold text-white mb-3">{item.q}</h3>
                  <p className="text-sm text-white/45 leading-relaxed">{item.a}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
