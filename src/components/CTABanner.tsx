import { Link } from 'react-router-dom'
import { ArrowRight, Github } from 'lucide-react'

export default function CTABanner() {
  return (
    <section className="section py-20">
      <div className="container-md">
        <div className="relative rounded-3xl overflow-hidden border border-white/10"
          style={{
            background: 'linear-gradient(135deg, rgba(0,212,255,0.1) 0%, rgba(139,92,246,0.1) 50%, rgba(16,185,129,0.05) 100%)',
          }}>
          <div className="absolute inset-0 overflow-hidden">
            <div className="absolute -top-20 -left-20 w-64 h-64 rounded-full blur-3xl opacity-20"
              style={{ background: '#00d4ff' }} />
            <div className="absolute -bottom-20 -right-20 w-64 h-64 rounded-full blur-3xl opacity-15"
              style={{ background: '#8b5cf6' }} />
          </div>
          <div className="absolute inset-0 grid-bg opacity-30" />
          <div className="relative px-8 py-16 text-center">
            <span className="badge-cyan mb-6">Open Source & Free to Start</span>
            <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
              Ready to see your{' '}
              <br className="hidden md:block" />
              infrastructure clearly?
            </h2>
            <p className="text-lg text-white/50 max-w-lg mx-auto mb-10">
              Deploy infraYS in 60 seconds. No credit card required.
              Full observability stack — metrics, logs, traces, profiling, AI — free forever on self-hosted.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link to="/install" className="btn-primary text-base px-8 py-4">
                Deploy Now — It's Free
                <ArrowRight className="w-5 h-5" />
              </Link>
              <a href="https://github.com/seshadrik143/infrays" target="_blank" rel="noopener noreferrer"
                className="btn-secondary text-base px-8 py-4">
                <Github className="w-5 h-5" />
                Star on GitHub
              </a>
            </div>
            <p className="mt-8 text-xs text-white/25">
              MIT Licensed · Self-hosted · No telemetry sent without consent
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}
