import { Download, Send, LayoutDashboard, ArrowRight } from 'lucide-react'

const steps = [
  {
    step: '01',
    icon: Download,
    title: 'Install the Agent',
    desc: 'One curl command. The infraYS agent auto-registers with your server, discovers running services, and starts collecting immediately. Zero config required.',
    code: 'curl -fsSL https://get.infrays.dev | sh',
    color: 'cyan',
  },
  {
    step: '02',
    icon: Send,
    title: 'Data Flows Automatically',
    desc: 'Metrics, logs, traces, and profiles stream to your infraYS server in real-time. Auto-discovery maps your services. AI baselines are built automatically.',
    code: 'infrays-agent status\n✦ Metrics: 847/s  Logs: 12k/h  Traces: OK',
    color: 'purple',
  },
  {
    step: '03',
    icon: LayoutDashboard,
    title: 'Observe, Alert & Act',
    desc: 'Open the dashboard, create alert rules, set SLOs, and integrate with your incident workflow. Everything you need is in one place.',
    code: '✦ 0 anomalies  ✦ All SLOs met  ✦ 0 open alerts',
    color: 'green',
  },
]

const colorMap: Record<string, { text: string; bg: string; border: string; lineColor: string }> = {
  cyan:   { text: 'text-cyan-400',   bg: 'bg-cyan-500/10',   border: 'border-cyan-500/20',   lineColor: '#00d4ff' },
  purple: { text: 'text-purple-400', bg: 'bg-purple-500/10', border: 'border-purple-500/20', lineColor: '#8b5cf6' },
  green:  { text: 'text-green-400',  bg: 'bg-green-500/10',  border: 'border-green-500/20',  lineColor: '#10b981' },
}

export default function HowItWorks() {
  return (
    <section className="section" id="how-it-works"
      style={{ background: 'rgba(10, 10, 20, 0.5)' }}>
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-16">
          <span className="badge-purple mb-4">Quick Start</span>
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
          {/* Connector line */}
          <div className="hidden lg:block absolute top-16 left-[calc(16.67%)] right-[calc(16.67%)] h-px"
            style={{ background: 'linear-gradient(90deg, #00d4ff30, #8b5cf630, #10b98130)' }} />

          <div className="grid lg:grid-cols-3 gap-8">
            {steps.map((step, i) => {
              const c = colorMap[step.color]
              return (
                <div key={step.step} className="relative">
                  {/* Arrow between steps (mobile) */}
                  {i < steps.length - 1 && (
                    <div className="lg:hidden flex justify-center my-4">
                      <ArrowRight className="w-5 h-5 text-white/20 rotate-90" />
                    </div>
                  )}

                  <div className="border border-white/[0.07] rounded-2xl p-7 h-full"
                    style={{ background: 'rgba(17, 17, 32, 0.7)' }}>
                    {/* Step number + icon */}
                    <div className="flex items-center gap-4 mb-6">
                      <div className={`w-12 h-12 rounded-xl ${c.bg} border ${c.border} flex items-center justify-center`}>
                        <step.icon className={`w-5 h-5 ${c.text}`} />
                      </div>
                      <span className="text-4xl font-black text-white/[0.05]">{step.step}</span>
                    </div>

                    <h3 className="text-lg font-bold text-white mb-3">{step.title}</h3>
                    <p className="text-sm text-white/40 leading-relaxed mb-5">{step.desc}</p>

                    {/* Code snippet */}
                    <div className="terminal rounded-lg">
                      <div className="px-4 py-3">
                        {step.code.split('\n').map((line, li) => (
                          <p key={li} className={`font-mono text-xs ${li === 0 ? c.text : 'text-green-400'}`}>
                            {line}
                          </p>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* Architecture visual */}
        <div className="mt-16 border border-white/[0.07] rounded-2xl p-8 text-center"
          style={{ background: 'rgba(17, 17, 32, 0.5)' }}>
          <h3 className="text-sm font-semibold text-white/30 uppercase tracking-widest mb-8">Architecture</h3>
          <div className="flex flex-col md:flex-row items-center justify-center gap-4 flex-wrap">
            {[
              { label: 'Your Servers', sub: 'Bare metal / VMs', color: 'text-cyan-400' },
              { label: '→', sub: '', color: 'text-white/20' },
              { label: 'infraYS Agent', sub: '12MB binary', color: 'text-purple-400' },
              { label: '→', sub: '', color: 'text-white/20' },
              { label: 'infraYS Server', sub: 'BoltDB + VictoriaMetrics', color: 'text-green-400' },
              { label: '→', sub: '', color: 'text-white/20' },
              { label: 'Dashboard + API', sub: 'React + REST', color: 'text-orange-400' },
            ].map((node, i) => (
              node.label === '→' ? (
                <span key={i} className="text-2xl text-white/15 hidden md:block">→</span>
              ) : (
                <div key={i} className="px-5 py-3 rounded-xl border border-white/[0.07]"
                  style={{ background: 'rgba(255,255,255,0.03)' }}>
                  <div className={`text-sm font-bold ${node.color}`}>{node.label}</div>
                  <div className="text-xs text-white/25 mt-0.5">{node.sub}</div>
                </div>
              )
            ))}
          </div>
          <div className="mt-6 text-xs text-white/20">
            Also supports: Kubernetes / Helm · Docker Compose · Multi-region HA · OTLP sources · StatsD
          </div>
        </div>
      </div>
    </section>
  )
}
