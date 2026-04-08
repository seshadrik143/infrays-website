import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, Play, Terminal, CheckCircle2, Github } from 'lucide-react'

const terminalLines = [
  { delay: 0,    text: '$ curl -fsSL https://get.infrays.dev | sh',        color: 'text-white/70' },
  { delay: 800,  text: '→ Downloading infraYS agent v1.0.0...',             color: 'text-cyan-400' },
  { delay: 1600, text: '→ Installing to /usr/local/bin/infrays-agent',       color: 'text-cyan-400' },
  { delay: 2400, text: '→ Agent registered ✓  Server connected ✓',          color: 'text-green-400' },
  { delay: 3200, text: '$ infrays-agent start',                             color: 'text-white/70' },
  { delay: 4000, text: '✦ Collecting metrics/s  Latency: 2ms',              color: 'text-purple-400' },
]

const highlights = [
  'No vendor lock-in',
  'Single binary agent',
  'Open-source (MIT)',
  '< 30MB memory',
]

export default function Hero() {
  const [visibleLines, setVisibleLines] = useState<number[]>([])

  useEffect(() => {
    const timers: ReturnType<typeof setTimeout>[] = []
    terminalLines.forEach((line, i) => {
      const t = setTimeout(() => {
        setVisibleLines((prev) => [...prev, i])
      }, line.delay + 500)
      timers.push(t)
    })
    return () => timers.forEach(clearTimeout)
  }, [])

  return (
    <section className="hero-bg grid-bg relative min-h-screen flex items-center pt-16 overflow-hidden">
      {/* Background orbs */}
      <div className="absolute top-1/4 left-1/4 w-[500px] h-[500px] rounded-full opacity-10 blur-[120px] animate-pulse-slow"
        style={{ background: 'radial-gradient(circle, #00d4ff, transparent)' }} />
      <div className="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] rounded-full opacity-10 blur-[100px] animate-pulse-slow"
        style={{ background: 'radial-gradient(circle, #8b5cf6, transparent)' }} />

      <div className="max-w-7xl mx-auto px-6 py-24 w-full">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Left */}
          <div>
            <div className="animate-fade-up mb-6">
              <span className="badge-cyan">
                <span className="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse" />
                v1.0 — Now Available
              </span>
            </div>

            <h1 className="animate-fade-up delay-100 text-5xl lg:text-6xl font-black leading-[1.1] tracking-tight mb-6">
              Observe{' '}
              <span className="text-gradient-brand">Everything.</span>
              <br />
              Miss{' '}
              <span className="text-gradient-cyan">Nothing.</span>
            </h1>

            <p className="animate-fade-up delay-200 text-lg text-white/50 leading-relaxed mb-8 max-w-xl">
              Metrics, logs, traces, continuous profiling, and AI-powered insights — all from a single
              lightweight agent. From bare metal to Kubernetes, infraYS gives you complete visibility.
            </p>

            <ul className="animate-fade-up delay-300 flex flex-wrap gap-x-6 gap-y-2 mb-10">
              {highlights.map((h) => (
                <li key={h} className="flex items-center gap-2 text-sm text-white/50">
                  <CheckCircle2 className="w-4 h-4 text-green-400 flex-shrink-0" />
                  {h}
                </li>
              ))}
            </ul>

            <div className="animate-fade-up delay-400 flex flex-wrap gap-4 mb-10">
              <Link to="/install" className="btn-primary">
                Get Started Free
                <ArrowRight className="w-4 h-4" />
              </Link>
              <Link to="/docs" className="btn-secondary">
                <Play className="w-4 h-4 text-cyan-400" />
                View Docs
              </Link>
            </div>

            {/* Honest badges */}
            <div className="animate-fade-up delay-500 flex items-center gap-6">
              <a href="https://github.com/seshadrik143/infrays" target="_blank" rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-white/40 hover:text-white/60 transition-colors">
                <Github className="w-4 h-4" />
                Open source on GitHub
              </a>
              <div className="w-px h-4 bg-white/10" />
              <div className="text-sm text-white/40">
                <span className="font-semibold text-white/70">MIT</span> licensed
              </div>
              <div className="w-px h-4 bg-white/10" />
              <div className="text-sm text-white/40">
                Self-hostable
              </div>
            </div>
          </div>

          {/* Right — Terminal */}
          <div className="animate-fade-up delay-300 lg:block">
            <div className="terminal shadow-[0_30px_80px_rgba(0,0,0,0.6)]">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
                <div className="flex items-center gap-2 ml-auto">
                  <Terminal className="w-3.5 h-3.5 text-white/30" />
                  <span className="text-xs text-white/30 font-mono">infraYS — bash</span>
                </div>
              </div>
              <div className="p-6 space-y-2 min-h-[220px]">
                {terminalLines.map((line, i) => (
                  visibleLines.includes(i) && (
                    <p key={i} className={`font-mono text-sm ${line.color} animate-fade-up`}>
                      {line.text}
                      {i === visibleLines[visibleLines.length - 1] && (
                        <span className="ml-0.5 inline-block w-2 h-4 bg-cyan-400 animate-[blink_1s_step-end_infinite] align-middle" />
                      )}
                    </p>
                  )
                ))}
              </div>
            </div>

            {/* Stat cards — only factual technical specs */}
            <div className="mt-4 grid grid-cols-3 gap-3">
              {[
                { label: 'Agent Size', value: '~12MB', color: 'text-cyan-400' },
                { label: 'Agent Memory', value: '< 30MB', color: 'text-purple-400' },
                { label: 'License', value: 'MIT', color: 'text-green-400' },
              ].map((stat) => (
                <div key={stat.label}
                  className="border-gradient rounded-xl p-4 text-center"
                  style={{ background: 'rgba(13, 13, 26, 0.8)' }}>
                  <div className={`text-xl font-black ${stat.color} mb-1`}>{stat.value}</div>
                  <div className="text-xs text-white/30">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Open source pitch — no fake company names */}
        <div className="mt-20 pt-10 border-t border-white/[0.05]">
          <p className="text-center text-xs text-white/25 uppercase tracking-widest mb-6">
            Built on open standards
          </p>
          <div className="flex flex-wrap justify-center items-center gap-x-10 gap-y-4">
            {['OpenTelemetry', 'Prometheus', 'Grafana Tempo', 'Grafana Loki', 'VictoriaMetrics', 'Kubernetes'].map((tech) => (
              <span key={tech} className="text-sm font-semibold text-white/20 hover:text-white/40 transition-colors cursor-default tracking-wide">
                {tech}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
