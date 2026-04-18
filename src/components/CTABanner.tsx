import { Link } from 'react-router-dom'
import { ArrowRight, Zap } from 'lucide-react'

export default function CTABanner() {
  return (
    <section className="section py-24 relative overflow-hidden">
      {/* Background orbs */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[400px] rounded-full blur-[100px] opacity-15"
          style={{ background: 'radial-gradient(ellipse, #00d4ff, #a855f7, transparent)' }} />
        <div className="absolute -top-20 -left-20 w-72 h-72 rounded-full blur-[80px] opacity-10"
          style={{ background: '#00d4ff' }} />
        <div className="absolute -bottom-20 -right-20 w-72 h-72 rounded-full blur-[80px] opacity-10"
          style={{ background: '#a855f7' }} />
      </div>

      <div className="container-md relative z-10">
        <div className="relative rounded-3xl overflow-hidden"
          style={{
            background: 'linear-gradient(135deg, rgba(0,212,255,0.07) 0%, rgba(13,13,26,0.9) 40%, rgba(168,85,247,0.07) 100%)',
            border: '1px solid rgba(255,255,255,0.08)',
            boxShadow: '0 0 0 1px rgba(0,212,255,0.06), 0 40px 100px rgba(0,0,0,0.6)',
          }}>

          {/* Subtle grid inside */}
          <div className="absolute inset-0 grid-bg opacity-20 pointer-events-none" />

          {/* Top gradient line */}
          <div className="absolute top-0 left-[10%] right-[10%] h-px"
            style={{ background: 'linear-gradient(90deg, transparent, rgba(0,212,255,0.4), rgba(168,85,247,0.4), transparent)' }} />

          <div className="relative px-8 py-20 text-center">
            {/* Icon */}
            <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-6"
              style={{
                background: 'linear-gradient(135deg, rgba(0,212,255,0.15), rgba(168,85,247,0.1))',
                border: '1px solid rgba(0,212,255,0.2)',
                boxShadow: '0 0 30px rgba(0,212,255,0.15)',
              }}>
              <Zap className="w-7 h-7 text-cyan-400 fill-cyan-400/30" />
            </div>

            <span className="badge-cyan mb-6">Open Source & Free to Start</span>

            <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5 mt-2">
              Ready to see your infrastructure{' '}
              <span className="text-gradient-brand">clearly?</span>
            </h2>
            <p className="text-lg text-white/45 max-w-lg mx-auto mb-10 leading-relaxed">
              Deploy infraYS in 60 seconds. No credit card. Full observability stack —
              metrics, logs, traces, profiling, AI — free forever on self-hosted.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-10">
              <Link to="/install" className="btn-primary text-base px-9 py-4">
                Deploy Now — It's Free
                <ArrowRight className="w-5 h-5" />
              </Link>
            </div>

            <p className="text-xs text-white/20 tracking-wide">
              Apache 2.0 · 15-day free trial · No telemetry sent without consent
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}
