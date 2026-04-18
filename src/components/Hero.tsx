import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, Terminal, Zap, Shield, Activity } from 'lucide-react'

const terminalLines = [
  { delay: 0,    text: '$ curl -fsSL https://get.infrays.org/install | sudo bash', color: 'text-white/60' },
  { delay: 900,  text: '→ Fetching latest release: v0.33.0...',                    color: 'text-cyan-400' },
  { delay: 1800, text: '→ Installing server + agent + npctl...',                   color: 'text-cyan-400' },
  { delay: 2700, text: '→ Trial started — 15 days remaining ✓',                    color: 'text-green-400' },
  { delay: 3600, text: '$ systemctl status nodepulse-server nodepulse-agent',      color: 'text-white/60' },
  { delay: 4500, text: '✦ Collecting 847 metrics/s  Latency: 2ms',                 color: 'text-purple-400' },
]

const pills = [
  { icon: Shield, label: 'Apache 2.0', color: 'text-green-400' },
  { icon: Zap, label: '< 30MB RAM', color: 'text-cyan-400' },
  { icon: Activity, label: 'Single Binary', color: 'text-purple-400' },
]

export default function Hero() {
  const [visibleLines, setVisibleLines] = useState<number[]>([])

  useEffect(() => {
    const timers: ReturnType<typeof setTimeout>[] = []
    terminalLines.forEach((line, i) => {
      const t = setTimeout(() => setVisibleLines((prev) => [...prev, i]), line.delay + 300)
      timers.push(t)
    })
    return () => timers.forEach(clearTimeout)
  }, [])

  return (
    <section className="hero-bg relative min-h-screen flex items-center pt-16 overflow-hidden">
      {/* Grid overlay */}
      <div className="absolute inset-0 grid-bg opacity-40 pointer-events-none" />

      {/* Floating orbs */}
      <div className="orb w-[700px] h-[700px] top-[-200px] left-[-200px] opacity-[0.07]"
        style={{ background: 'radial-gradient(circle, #00d4ff, transparent 70%)' }} />
      <div className="orb w-[600px] h-[600px] bottom-[-150px] right-[-100px] opacity-[0.06] animate-float"
        style={{ background: 'radial-gradient(circle, #a855f7, transparent 70%)', animationDelay: '2s' }} />
      <div className="orb w-[400px] h-[400px] top-[30%] right-[20%] opacity-[0.05] animate-float"
        style={{ background: 'radial-gradient(circle, #10b981, transparent 70%)', animationDelay: '4s' }} />

      <div className="max-w-7xl mx-auto px-6 py-24 w-full relative z-10">
        <div className="grid lg:grid-cols-2 gap-16 items-center">

          {/* ── Left column ─────────────────────────────────── */}
          <div>
            {/* Badge */}
            <div className="animate-fade-up mb-7">
              <span className="badge-cyan">
                <span className="relative flex w-2 h-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75" />
                  <span className="relative inline-flex rounded-full w-2 h-2 bg-cyan-400" />
                </span>
                v1.0 — Now Available
              </span>
            </div>

            {/* Headline */}
            <h1 className="animate-fade-up delay-100 text-6xl lg:text-7xl font-black leading-[1.05] tracking-tight mb-6">
              Observe{' '}
              <span className="text-gradient-brand">Everything.</span>
              <br />
              Miss{' '}
              <span className="text-gradient-cyan">Nothing.</span>
            </h1>

            <p className="animate-fade-up delay-200 text-xl text-white/45 leading-relaxed mb-8 max-w-xl">
              The unified observability stack — metrics, logs, traces, profiling, and AI — all from a{' '}
              <span className="text-white/70 font-medium">single 12MB agent</span>.
              From bare metal to Kubernetes.
            </p>

            {/* Pills */}
            <div className="animate-fade-up delay-300 flex flex-wrap gap-3 mb-10">
              {pills.map((p) => (
                <div key={p.label}
                  className="flex items-center gap-2 px-4 py-2 rounded-full border border-white/[0.08] text-sm font-medium"
                  style={{ background: 'rgba(255,255,255,0.03)', backdropFilter: 'blur(8px)' }}>
                  <p.icon className={`w-3.5 h-3.5 ${p.color}`} />
                  <span className="text-white/60">{p.label}</span>
                </div>
              ))}
            </div>

            {/* CTA buttons */}
            <div className="animate-fade-up delay-400 flex flex-wrap gap-4 mb-10">
              <Link to="/install" className="btn-primary text-base px-8 py-4">
                Deploy Free Now
                <ArrowRight className="w-5 h-5" />
              </Link>
              <Link to="/docs" className="btn-secondary text-base px-8 py-4">
                Read the Docs
              </Link>
            </div>

            {/* Bottom social proof */}
            <div className="animate-fade-up delay-500 flex items-center gap-5 text-sm text-white/30">
              <span>Self-hostable</span>
              <span className="w-px h-4 bg-white/10" />
              <span>No telemetry</span>
            </div>
          </div>

          {/* ── Right column — Terminal ──────────────────────── */}
          <div className="animate-fade-up delay-300">
            {/* Glow halo behind terminal */}
            <div className="relative">
              <div className="absolute -inset-4 rounded-3xl opacity-30 blur-2xl pointer-events-none"
                style={{ background: 'radial-gradient(ellipse at 50% 50%, rgba(0,212,255,0.25), rgba(168,85,247,0.15), transparent)' }} />

              <div className="terminal relative">
                {/* Traffic lights */}
                <div className="terminal-header">
                  <span className="terminal-dot" style={{ color: '#ff5f57', background: '#ff5f57' }} />
                  <span className="terminal-dot" style={{ color: '#ffbd2e', background: '#ffbd2e' }} />
                  <span className="terminal-dot" style={{ color: '#28ca41', background: '#28ca41' }} />
                  <div className="flex items-center gap-2 ml-auto">
                    <Terminal className="w-3.5 h-3.5 text-white/25" />
                    <span className="text-xs text-white/25 font-mono">infraYS — bash</span>
                  </div>
                </div>

                <div className="p-6 space-y-2.5 min-h-[230px]">
                  {terminalLines.map((line, i) =>
                    visibleLines.includes(i) ? (
                      <p key={i} className={`font-mono text-sm leading-relaxed ${line.color} animate-fade-up`}>
                        {line.text}
                        {i === visibleLines[visibleLines.length - 1] && (
                          <span className="ml-1 inline-block w-[7px] h-[14px] bg-cyan-400 animate-[blink_1s_step-end_infinite] align-middle opacity-90" />
                        )}
                      </p>
                    ) : null
                  )}
                </div>
              </div>
            </div>

            {/* Stat cards */}
            <div className="mt-4 grid grid-cols-3 gap-3">
              {[
                { label: 'Binary Size', value: '~12MB', color: 'text-cyan-400', glow: 'rgba(0,212,255,0.15)' },
                { label: 'Agent RAM',   value: '< 30MB', color: 'text-purple-400', glow: 'rgba(168,85,247,0.15)' },
                { label: 'License',    value: 'Apache 2.0', color: 'text-green-400', glow: 'rgba(16,185,129,0.15)' },
              ].map((stat) => (
                <div key={stat.label}
                  className="gradient-border text-center py-4 px-3 transition-all duration-300 hover:scale-105"
                  style={{ boxShadow: `0 0 20px ${stat.glow}` }}>
                  <div className={`text-2xl font-black ${stat.color} mb-1`}>{stat.value}</div>
                  <div className="text-xs text-white/30">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* ── Open standards strip ─────────────────────────── */}
        <div className="mt-24 pt-10 border-t border-white/[0.05]">
          <p className="text-center text-[11px] font-semibold text-white/20 uppercase tracking-[0.2em] mb-7">
            Built on open standards
          </p>
          <div className="flex flex-wrap justify-center items-center gap-x-12 gap-y-3">
            {['OpenTelemetry', 'Prometheus', 'Grafana Tempo', 'Grafana Loki', 'VictoriaMetrics', 'Kubernetes'].map((tech) => (
              <span key={tech}
                className="text-sm font-semibold text-white/20 hover:text-white/50 transition-colors duration-300 cursor-default tracking-wide">
                {tech}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
