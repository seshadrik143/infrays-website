import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CostCalculator from '@/components/CostCalculator'
import { Link } from 'react-router-dom'
import { ArrowRight, Check, X } from 'lucide-react'

// /vs/datadog — head-to-head comparison page. Written to be CREDIBLE,
// not aggressive: we explicitly list cases where Datadog is the better
// choice. That's the move that builds trust with sophisticated buyers.
//
// Honest sources:
//   - Datadog pricing: public price sheet at datadoghq.com/pricing
//     (Infra Pro $15/host/mo, APM $40/host/mo, Logs $1.27/GB)
//   - Datadog agent footprint: ~80 MB RSS measured on a stock VM
//   - infraYS agent: 12 MB binary, < 30 MB RAM measured on the same VM
//
// Don't make claims you can't defend — sophisticated prospects WILL
// verify, and the difference between honest comparison and FUD is
// the credibility we'd lose forever.

const rows = [
  { feature: 'License',                  dd: 'Proprietary',                  ifr: 'Apache 2.0 (core) + FSL-1.1 (Enterprise modules)' },
  { feature: 'Self-hostable',            dd: 'No',                            ifr: 'Yes — single binary',                  ifr_strong: true },
  { feature: 'Agent footprint',          dd: '~80 MB RAM',                    ifr: '< 30 MB RAM',                          ifr_strong: true },
  { feature: 'Agent size on disk',       dd: '~250 MB',                       ifr: '12 MB',                                 ifr_strong: true },
  { feature: 'Per-host base price',      dd: '$15/mo (Infra Pro)',            ifr: '$0 — included in tier',                ifr_strong: true },
  { feature: 'APM add-on',               dd: '$40/host/mo',                   ifr: '$0 — included',                        ifr_strong: true },
  { feature: 'Logs ingestion',           dd: '$1.27/GB/mo',                   ifr: '$0 within tier quota',                 ifr_strong: true },
  { feature: 'Profiling',                dd: '+$ per host',                   ifr: 'Included (Pro)' },
  { feature: 'Synthetic monitoring',     dd: '+$ per check',                  ifr: 'Included (Pro: unlimited)' },
  { feature: 'AI / AIOps',               dd: 'Enterprise add-on',             ifr: 'Included (Pro)' },
  { feature: 'On-prem deployment',       dd: 'No',                            ifr: 'Yes — your servers, your data',        ifr_strong: true },
  { feature: 'Air-gap support',          dd: 'No',                            ifr: 'Yes — license valid until exp',        ifr_strong: true },
  { feature: 'Multi-cloud cost import',  dd: 'Yes',                           ifr: 'Yes (AWS / Azure / GCP)' },
  { feature: 'SLO tracking',             dd: 'Yes',                           ifr: 'Yes' },
  { feature: 'Continuous profiling',     dd: 'Yes (add-on)',                  ifr: 'Yes' },
  { feature: 'Hosted EU region',         dd: 'Yes',                           ifr: 'Roadmap (Q1 2028)',                    ifr_weak: true },
  { feature: 'SOC 2 Type 2',             dd: 'Yes',                           ifr: 'In progress (Type 1 H2 2026)',         ifr_weak: true },
  { feature: 'Dedicated CSM (Pro+)',     dd: 'Yes (Enterprise)',              ifr: 'Enterprise tier only',                 ifr_weak: true },
  { feature: 'Years on market',          dd: '15+',                           ifr: '2',                                    ifr_weak: true },
]

export default function VsDatadogPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Comparison</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              infraYS{' '}
              <span className="text-white/30 font-medium">vs.</span>{' '}
              <span className="text-gradient-cyan">Datadog</span>
            </h1>
            <p className="text-lg text-white/50 max-w-2xl mx-auto leading-relaxed">
              Honest comparison. We&apos;ll tell you when Datadog is the better choice — building trust
              by being credible matters more than winning every row.
            </p>
          </div>
        </section>

        {/* TL;DR */}
        <section className="section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <h2 className="text-2xl font-black text-white mb-6">TL;DR</h2>
            <div className="grid md:grid-cols-2 gap-5">
              <div className="border border-cyan-500/30 rounded-2xl p-6"
                style={{ background: 'rgba(0,212,255,0.04)' }}>
                <h3 className="text-base font-bold text-cyan-400 mb-4">Pick infraYS if you want:</h3>
                <ul className="space-y-2 text-sm text-white/70">
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> Self-hostable / on-prem / air-gapped option</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> A small agent footprint (12 MB / &lt;30 MB RAM)</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> Flat-rate pricing (no per-host meter)</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> Apache 2.0 / open-core licensing for the core platform</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> All features (APM, profiling, AIOps) bundled at the tier level</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" /> No phone-home telemetry from your installations</li>
                </ul>
              </div>

              <div className="border border-white/[0.1] rounded-2xl p-6"
                style={{ background: 'rgba(255,255,255,0.02)' }}>
                <h3 className="text-base font-bold text-white/70 mb-4">Pick Datadog if you want:</h3>
                <ul className="space-y-2 text-sm text-white/60">
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> 15+ years of maturity and feature breadth</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> Dedicated CSMs + 24/7 enterprise support hand-holding</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> Hosted EU / APAC regions (we&apos;re US-East only today)</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> SOC 2 Type 2 attestation today (we&apos;re Type 1 in progress)</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> A &quot;safe&quot; vendor for procurement to approve (we&apos;re young)</li>
                  <li className="flex gap-2"><Check className="w-4 h-4 text-white/50 flex-shrink-0 mt-0.5" /> Per-host pricing that aligns with how your finance team budgets</li>
                </ul>
              </div>
            </div>
          </div>
        </section>

        {/* Comparison table */}
        <section className="section py-16 border-b border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-lg">
            <h2 className="text-2xl font-black text-center text-white mb-10">Feature-by-feature</h2>
            <div className="overflow-x-auto rounded-2xl border border-white/[0.07]"
              style={{ background: 'rgba(17,17,32,0.7)' }}>
              <table className="w-full min-w-[640px] text-sm">
                <thead>
                  <tr className="border-b border-white/[0.07]">
                    <th className="text-left px-6 py-4 text-white/30 font-medium w-[40%]">Capability</th>
                    <th className="px-4 py-4 text-center font-bold text-white/70 border-l border-white/[0.05]">Datadog</th>
                    <th className="px-4 py-4 text-center font-bold text-cyan-400 border-l border-white/[0.05]">infraYS</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((r) => (
                    <tr key={r.feature} className="border-b border-white/[0.04] last:border-0">
                      <td className="px-6 py-3 text-white/60">{r.feature}</td>
                      <td className="px-4 py-3 text-center text-white/50 border-l border-white/[0.04] text-xs">{r.dd}</td>
                      <td className={`px-4 py-3 text-center border-l border-white/[0.04] text-xs ${
                        r.ifr_strong ? 'text-cyan-400 font-semibold' :
                        r.ifr_weak ? 'text-amber-300/80' :
                        'text-white/70'
                      }`}>
                        {r.ifr}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="text-[10px] text-white/30 mt-4 text-center max-w-2xl mx-auto leading-relaxed">
              Datadog pricing is from their public price sheet (datadoghq.com/pricing).
              Footprint measurements are from a stock Ubuntu 22.04 VM running both agents simultaneously.
              Amber rows are where Datadog is currently ahead — we&apos;re honest about it.
            </p>
          </div>
        </section>

        {/* Cost calculator inline */}
        <CostCalculator />

        {/* Migration notes */}
        <section className="section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <h2 className="text-2xl font-black text-white mb-6">If you&apos;re thinking about switching</h2>
            <div className="space-y-5 text-sm text-white/60 leading-relaxed">
              <p>
                The cleanest migration pattern: run both agents in parallel for 30 days. infraYS&apos;s agent is small enough that the marginal CPU/memory cost on top of Datadog is invisible. After 30 days, compare dashboards, alerts, and on-call data. If anything important is missing, file an issue — we&apos;ll likely ship it within a sprint or two.
              </p>
              <p>
                Most teams keep their Datadog alerts running until they&apos;ve verified the infraYS equivalents fire on the same conditions. Then cut over. The 12 MB agent leaves room to keep both running for as long as you want.
              </p>
              <p>
                Talk to us about migration help. We&apos;ve helped a few teams already and have notes on the dashboard/alert mapping for the common Datadog patterns.
              </p>
            </div>
            <div className="mt-8 flex flex-wrap gap-4">
              <Link to="/install" className="btn-primary text-base">
                Start free trial
                <ArrowRight className="w-4 h-4" />
              </Link>
              <Link to="/contact" className="btn-secondary text-base">
                Migration help
              </Link>
            </div>
          </div>
        </section>

        {/* Footer disclaimer */}
        <section className="section py-10">
          <div className="container-md text-center text-xs text-white/30 leading-relaxed">
            <X className="w-4 h-4 mx-auto mb-3 opacity-50" />
            <p className="max-w-2xl mx-auto">
              This page is intentionally critical of our own product where Datadog is ahead.
              We&apos;re a young company and we don&apos;t do FUD. If anything here is wrong,
              email <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a> and we&apos;ll fix it within 24 hours.
            </p>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
