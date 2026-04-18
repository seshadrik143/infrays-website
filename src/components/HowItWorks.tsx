import { Download, Send, LayoutDashboard } from 'lucide-react'

const steps = [
  {
    step: '01',
    icon: Download,
    title: 'Install the Agent',
    desc: 'One curl command. The infraYS agent auto-registers with your server, discovers running services, and starts collecting immediately. Zero config required.',
    code: '$ curl -fsSL https://get.infrays.org/install | sudo bash',
    output: '→ Trial started — 15 days remaining ✓',
    outputColor: 'text-green-400',
    color: 'cyan',
  },
  {
    step: '02',
    icon: Send,
    title: 'Data Flows Automatically',
    desc: 'Metrics, logs, traces, and profiles stream to your infraYS server in real-time. Auto-discovery maps your services. AI baselines are built automatically.',
    code: '$ systemctl status nodepulse-agent',
    output: '✦ Metrics: 847/s  Logs: 12k/h  Traces: OK',
    outputColor: 'text-purple-400',
    color: 'purple',
  },
  {
    step: '03',
    icon: LayoutDashboard,
    title: 'Observe, Alert & Act',
    desc: 'Open the dashboard, create alert rules, set SLOs, and integrate with your incident workflow. Everything you need is in one place.',
    code: '$ npctl agents list',
    output: '✦ 0 anomalies  ✦ All SLOs met  ✦ 0 alerts',
    outputColor: 'text-green-400',
    color: 'green',
  },
]

const colorMap: Record<string, {
  text: string; bg: string; border: string; glowColor: string; stepText: string
}> = {
  cyan:   { text: 'text-cyan-400',   bg: 'rgba(0,212,255,0.08)',  border: 'rgba(0,212,255,0.25)',  glowColor: 'rgba(0,212,255,0.15)',  stepText: 'text-cyan-500/20' },
  purple: { text: 'text-purple-400', bg: 'rgba(168,85,247,0.08)', border: 'rgba(168,85,247,0.25)', glowColor: 'rgba(168,85,247,0.15)', stepText: 'text-purple-500/20' },
  green:  { text: 'text-green-400',  bg: 'rgba(16,185,129,0.08)', border: 'rgba(16,185,129,0.25)', glowColor: 'rgba(16,185,129,0.15)', stepText: 'text-green-500/20' },
}

export default function HowItWorks() {
  return (
    <section className="section relative overflow-hidden" id="how-it-works"
      style={{ background: 'rgba(6,6,14,0.6)' }}>
      {/* Dot background */}
      <div className="absolute inset-0 dot-bg opacity-30 pointer-events-none" />

      <div className="container-lg relative z-10">
        {/* Header */}
        <div className="text-center mb-16">
          <span className="badge-purple mb-5">Quick Start</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Up and running in{' '}
            <span className="text-gradient-cyan">60 seconds</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto">
            No Helm charts to debug. No Prometheus scrape configs. No week-long onboarding.
            Just one command.
          </p>
        </div>

        {/* Steps */}
        <div className="relative">
          {/* Connector dashes (desktop) */}
          <div className="hidden lg:flex absolute top-[52px] left-[calc(33.33%+32px)] right-[calc(33.33%+32px)] items-center justify-center pointer-events-none">
            <div className="flex-1 h-px" style={{
              background: 'linear-gradient(90deg, rgba(0,212,255,0.3), rgba(168,85,247,0.3), rgba(16,185,129,0.3))',
              maskImage: 'repeating-linear-gradient(90deg, #000 0 8px, transparent 8px 16px)',
              WebkitMaskImage: 'repeating-linear-gradient(90deg, #000 0 8px, transparent 8px 16px)',
            }} />
          </div>

          <div className="grid lg:grid-cols-3 gap-6">
            {steps.map((step) => {
              const c = colorMap[step.color]
              return (
                <div key={step.step} className="group relative">
                  <div
                    className="rounded-2xl p-7 h-full transition-all duration-350 hover:-translate-y-2"
                    style={{
                      background: 'rgba(13,13,24,0.8)',
                      border: '1px solid rgba(255,255,255,0.06)',
                      backdropFilter: 'blur(10px)',
                    }}>
                    {/* Hover glow */}
                    <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-400 pointer-events-none"
                      style={{ boxShadow: `0 0 0 1px ${c.border}, 0 20px 60px ${c.glowColor}` }} />

                    {/* Icon + step number */}
                    <div className="flex items-center gap-4 mb-6 relative z-10">
                      <div className="w-12 h-12 rounded-xl flex items-center justify-center transition-transform duration-300 group-hover:scale-110"
                        style={{ background: c.bg, border: `1px solid ${c.border}` }}>
                        <step.icon className={`w-5 h-5 ${c.text}`} />
                      </div>
                      <span className="text-5xl font-black text-white/[0.04] select-none">{step.step}</span>
                    </div>

                    <h3 className="text-lg font-bold text-white mb-3 relative z-10">{step.title}</h3>
                    <p className="text-sm text-white/40 leading-relaxed mb-5 relative z-10">{step.desc}</p>

                    {/* Mini terminal */}
                    <div className="rounded-xl overflow-hidden relative z-10" style={{
                      background: '#040410',
                      border: '1px solid rgba(255,255,255,0.07)',
                    }}>
                      <div className="px-4 py-3 border-b border-white/[0.05] flex items-center gap-1.5">
                        <div className="w-2 h-2 rounded-full bg-white/10" />
                        <div className="w-2 h-2 rounded-full bg-white/10" />
                        <div className="w-2 h-2 rounded-full bg-white/10" />
                      </div>
                      <div className="px-4 py-3 space-y-1.5">
                        <p className={`font-mono text-xs ${c.text}`}>{step.code}</p>
                        <p className={`font-mono text-xs ${step.outputColor}`}>{step.output}</p>
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* Architecture diagram */}
        <div className="mt-16 rounded-2xl p-8" style={{
          background: 'rgba(10,10,20,0.6)',
          border: '1px solid rgba(255,255,255,0.06)',
          backdropFilter: 'blur(8px)',
        }}>
          <h3 className="text-xs font-semibold text-white/25 uppercase tracking-[0.2em] text-center mb-8">
            Architecture
          </h3>
          <div className="flex flex-col md:flex-row items-center justify-center gap-3 flex-wrap">
            {[
              { label: 'Your Servers', sub: 'Bare metal / VMs', color: 'text-cyan-400',   border: 'rgba(0,212,255,0.2)',    bg: 'rgba(0,212,255,0.06)' },
              null,
              { label: 'infraYS Agent', sub: '12MB binary',      color: 'text-purple-400', border: 'rgba(168,85,247,0.2)',   bg: 'rgba(168,85,247,0.06)' },
              null,
              { label: 'infraYS Server', sub: 'BoltDB + VictoriaMetrics', color: 'text-green-400', border: 'rgba(16,185,129,0.2)', bg: 'rgba(16,185,129,0.06)' },
              null,
              { label: 'Dashboard + API', sub: 'React + REST', color: 'text-orange-400',  border: 'rgba(245,158,11,0.2)',   bg: 'rgba(245,158,11,0.06)' },
            ].map((node, i) =>
              node === null ? (
                <span key={i} className="text-xl font-bold text-white/10 hidden md:block">→</span>
              ) : (
                <div key={i} className="px-5 py-3.5 rounded-xl text-center transition-transform hover:scale-105"
                  style={{ background: node.bg, border: `1px solid ${node.border}` }}>
                  <div className={`text-sm font-bold ${node.color}`}>{node.label}</div>
                  <div className="text-xs text-white/25 mt-0.5">{node.sub}</div>
                </div>
              )
            )}
          </div>
          <p className="mt-6 text-center text-xs text-white/20">
            Also supports: Kubernetes / Helm · Docker Compose · Multi-region HA · OTLP · StatsD
          </p>
        </div>
      </div>
    </section>
  )
}
