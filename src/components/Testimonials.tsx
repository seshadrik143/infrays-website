import { Link } from 'react-router-dom'
import { MessageSquare, ArrowRight, Quote } from 'lucide-react'

// Testimonials. Names + roles + companies are PLACEHOLDERS — replace
// one at a time with real customer quotes (with permission) as they
// come in. Until then, the quotes themselves are written to be
// realistic and specific (not generic praise), so they still convey
// "this product solves a real problem" even when readers know they're
// illustrative.
//
// When replacing: keep the structure {quote, name, role, company,
// initials, color}. Initials drive the avatar fallback.
const testimonials = [
  {
    quote: '"We swapped our previous observability stack for self-hosted infraYS. Same dashboards, same alerts, same SLOs — and we cut our observability bill significantly. The 12 MB agent footprint was the icing."',
    name: 'Sarah K.',
    role: 'Staff SRE',
    company: 'mid-size fintech (placeholder)',
    initials: 'SK',
    color: 'cyan',
  },
  {
    quote: '"The audit log being hash-chained out of the box meant we passed our SOC 2 Type 1 review without any extra tooling. Most of the work was already done. Saved a quarter of engineering time, easily."',
    name: 'Marcus T.',
    role: 'Director of Platform',
    company: 'B2B SaaS, ~200 engineers (placeholder)',
    initials: 'MT',
    color: 'purple',
  },
  {
    quote: '"We run on bare metal in a cage in Tier-4 colo. Most monitoring vendors never quite fit. infraYS self-hosted just works — single binary, no Kubernetes required, and we control every byte that leaves our cabinet."',
    name: 'Priya N.',
    role: 'Infra Lead',
    company: 'colo-only e-commerce (placeholder)',
    initials: 'PN',
    color: 'green',
  },
]

const colorMap: Record<string, { bg: string; border: string; text: string; glow: string }> = {
  cyan:   { bg: 'rgba(0,212,255,0.08)',  border: 'rgba(0,212,255,0.2)',  text: 'text-cyan-400',   glow: '0 0 20px rgba(0,212,255,0.1)' },
  purple: { bg: 'rgba(168,85,247,0.08)', border: 'rgba(168,85,247,0.2)', text: 'text-purple-400', glow: '0 0 20px rgba(168,85,247,0.1)' },
  green:  { bg: 'rgba(16,185,129,0.08)', border: 'rgba(16,185,129,0.2)', text: 'text-green-400',  glow: '0 0 20px rgba(16,185,129,0.1)' },
}

export default function Testimonials() {
  return (
    <section className="section relative overflow-hidden" id="community">
      {/* Background mesh */}
      <div className="absolute inset-0 mesh-bg pointer-events-none opacity-40" />

      <div className="container-lg relative z-10">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-cyan mb-5">From customers</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Built for teams.{' '}
            <span className="text-gradient-cyan">Loved by engineers.</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto leading-relaxed">
            Quotes below are illustrative until we have permission to publish real customer names.
            Real customers tell us essentially the same thing.
          </p>
        </div>

        {/* Testimonial cards */}
        <div className="grid md:grid-cols-3 gap-5 mb-10">
          {testimonials.map((t) => {
            const c = colorMap[t.color]
            return (
              <div key={t.name}
                className="group relative rounded-2xl p-7 flex flex-col transition-all duration-350 hover:-translate-y-1"
                style={{
                  background: 'rgba(13,13,24,0.7)',
                  border: '1px solid rgba(255,255,255,0.06)',
                  backdropFilter: 'blur(10px)',
                }}>
                {/* Hover glow */}
                <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-400 pointer-events-none"
                  style={{ boxShadow: `inset 0 0 0 1px ${c.border}, ${c.glow}` }} />

                <div className="relative z-10 flex flex-col h-full">
                  <Quote className={`w-6 h-6 ${c.text} opacity-50 mb-3`} />
                  <p className="text-sm text-white/70 leading-relaxed mb-6 flex-1">{t.quote}</p>

                  <div className="flex items-center gap-3 pt-4 border-t border-white/[0.06]">
                    <div className="w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm flex-shrink-0"
                      style={{ background: c.bg, border: `1px solid ${c.border}`, color: c.text.replace('text-', '').includes('cyan') ? '#22d3ee' : c.text.replace('text-', '').includes('purple') ? '#a855f7' : '#10b981' }}>
                      {t.initials}
                    </div>
                    <div className="min-w-0">
                      <div className="text-sm font-semibold text-white truncate">{t.name}</div>
                      <div className="text-xs text-white/40 truncate">{t.role} · {t.company}</div>
                    </div>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Talk-to-us banner */}
        <div className="rounded-2xl p-8 flex flex-col md:flex-row items-center justify-between gap-6"
          style={{
            background: 'linear-gradient(135deg, rgba(0,212,255,0.06), rgba(168,85,247,0.04))',
            border: '1px solid rgba(0,212,255,0.12)',
          }}>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0"
              style={{ background: 'rgba(0,212,255,0.1)', border: '1px solid rgba(0,212,255,0.2)' }}>
              <MessageSquare className="w-6 h-6 text-cyan-400" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Talk to us before you buy</h3>
              <p className="text-sm text-white/40 mt-1">
                Sizing questions, deployment questions, security questions — we answer all of them.
              </p>
            </div>
          </div>
          <Link to="/contact"
            className="btn-secondary whitespace-nowrap flex-shrink-0 flex items-center gap-2">
            <MessageSquare className="w-4 h-4" />
            Contact Sales
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
