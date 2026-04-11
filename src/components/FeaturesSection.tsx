import { Link } from 'react-router-dom'
import {
  BarChart3, FileText, GitBranch, Brain, Bell, Shield,
  Activity, Globe, Cpu, Database, Cloud, Puzzle,
  ArrowRight, Network, Bug, BarChart2
} from 'lucide-react'

const features = [
  {
    icon: BarChart3,
    color: 'cyan',
    title: 'Unified Metrics',
    desc: 'Collect system, process, Docker, Kubernetes, custom, and StatsD metrics at up to 1M/sec. Built-in Prometheus remote_write compatibility.',
    tags: ['Prometheus', 'StatsD', 'OpenMetrics'],
  },
  {
    icon: FileText,
    color: 'purple',
    title: 'Structured Log Intelligence',
    desc: 'Ingest logs from any source with Grok parsing, full-text search, and real-time tailing. Integrates with Loki and any log backend.',
    tags: ['Loki', 'Grok', 'Full-text search'],
  },
  {
    icon: GitBranch,
    color: 'blue',
    title: 'Distributed Tracing',
    desc: 'OTLP-native trace ingestion, span analysis, and service dependency mapping. Powered by Grafana Tempo-compatible storage.',
    tags: ['OTLP', 'Tempo', 'Service Map'],
  },
  {
    icon: Cpu,
    color: 'orange',
    title: 'Continuous Profiling',
    desc: 'CPU and memory flame graphs at millisecond resolution. Click-to-zoom profiles, search by function, no code changes needed.',
    tags: ['pprof', 'Flame Graph', 'APM'],
  },
  {
    icon: Brain,
    color: 'pink',
    title: 'AI-Powered AIOps',
    desc: 'Anomaly detection with Z-score + Isolation Forest, predictive alerting, and LLM-powered root cause analysis. Spot problems before users do.',
    tags: ['Anomaly Detection', 'RCA', 'Forecasting'],
  },
  {
    icon: Bell,
    color: 'yellow',
    title: 'Intelligent Alerting',
    desc: 'Flap detection, alert correlation, baseline-aware thresholds, and on-call scheduling. Fire fewer, better alerts.',
    tags: ['PagerDuty', 'Slack', 'On-Call'],
  },
  {
    icon: Globe,
    color: 'teal',
    title: 'Synthetic Monitoring',
    desc: 'HTTP, TCP, DNS, SSL, multi-step API flows, and browser checks from global probe nodes. Public status page included.',
    tags: ['HTTP', 'Browser', 'Status Page'],
  },
  {
    icon: Activity,
    color: 'green',
    title: 'SLO & SLA Tracking',
    desc: 'Define error budgets, track burn rates, and generate compliance-ready SLA reports. Visualize uptime across your entire fleet.',
    tags: ['Error Budget', 'SLA Report', 'Burn Rate'],
  },
  {
    icon: Shield,
    color: 'indigo',
    title: 'Enterprise Security',
    desc: 'OIDC/SSO, RBAC with 11 granular permissions, mTLS, AES-256 encryption, GDPR erasure, SOC2/ISO27001 compliance controls.',
    tags: ['OIDC', 'RBAC', 'Compliance'],
  },
  {
    icon: Database,
    color: 'rose',
    title: 'Multi-Tenancy & HA',
    desc: 'Raft consensus, per-tenant data isolation, resource quotas, and unlimited agents per tenant. Built for enterprise scale.',
    tags: ['Raft HA', 'Multi-tenant', 'Quotas'],
  },
  {
    icon: Cloud,
    color: 'sky',
    title: 'Cloud Cost Intelligence',
    desc: 'Import and visualize AWS, Azure, GCP costs alongside infra metrics in the same dashboard. Identify waste instantly.',
    tags: ['AWS', 'Azure', 'GCP'],
  },
  {
    icon: Puzzle,
    color: 'violet',
    title: 'Plugin Ecosystem',
    desc: '67+ community plugins, custom collector SDK, webhook & dashboard templates, Terraform & Pulumi providers. Integrates with GitHub, Jira, ServiceNow.',
    tags: ['67 Plugins', 'SDK', 'Terraform'],
  },
  {
    icon: Network,
    color: 'amber',
    title: 'Auto-Discovery & Topology',
    desc: '/proc-based service scan, TCP flow tracking, and cloud metadata detection (AWS/GCP/Azure IMDSv2). Visualize your entire service topology.',
    tags: ['eBPF', 'Service Map', 'Cloud Detect'],
  },
  {
    icon: Bug,
    color: 'red',
    title: 'Error Tracking',
    desc: 'Automatic error grouping with SHA-256 fingerprinting, first/last seen tracking, stack trace capture, and error rate trends per service.',
    tags: ['Fingerprinting', 'Grouping', 'Stacks'],
  },
  {
    icon: BarChart2,
    color: 'lime',
    title: 'Alert Analytics & Coverage',
    desc: 'MTTR tracking, alert fatigue scoring, flap detection, and coverage scoring to measure how well your infrastructure is monitored.',
    tags: ['MTTR', 'Fatigue Score', 'Coverage'],
  },
]

const colorMap: Record<string, { bg: string; border: string; text: string; shadow: string }> = {
  cyan:   { bg: 'rgba(0,212,255,0.08)',    border: 'rgba(0,212,255,0.2)',    text: 'text-cyan-400',   shadow: '0 0 30px rgba(0,212,255,0.08)' },
  purple: { bg: 'rgba(168,85,247,0.08)',   border: 'rgba(168,85,247,0.2)',   text: 'text-purple-400', shadow: '0 0 30px rgba(168,85,247,0.08)' },
  blue:   { bg: 'rgba(59,130,246,0.08)',   border: 'rgba(59,130,246,0.2)',   text: 'text-blue-400',   shadow: '0 0 30px rgba(59,130,246,0.08)' },
  orange: { bg: 'rgba(249,115,22,0.08)',   border: 'rgba(249,115,22,0.2)',   text: 'text-orange-400', shadow: '0 0 30px rgba(249,115,22,0.08)' },
  pink:   { bg: 'rgba(236,72,153,0.08)',   border: 'rgba(236,72,153,0.2)',   text: 'text-pink-400',   shadow: '0 0 30px rgba(236,72,153,0.08)' },
  yellow: { bg: 'rgba(234,179,8,0.08)',    border: 'rgba(234,179,8,0.2)',    text: 'text-yellow-400', shadow: '0 0 30px rgba(234,179,8,0.08)' },
  teal:   { bg: 'rgba(20,184,166,0.08)',   border: 'rgba(20,184,166,0.2)',   text: 'text-teal-400',   shadow: '0 0 30px rgba(20,184,166,0.08)' },
  green:  { bg: 'rgba(16,185,129,0.08)',   border: 'rgba(16,185,129,0.2)',   text: 'text-green-400',  shadow: '0 0 30px rgba(16,185,129,0.08)' },
  indigo: { bg: 'rgba(99,102,241,0.08)',   border: 'rgba(99,102,241,0.2)',   text: 'text-indigo-400', shadow: '0 0 30px rgba(99,102,241,0.08)' },
  rose:   { bg: 'rgba(244,63,94,0.08)',    border: 'rgba(244,63,94,0.2)',    text: 'text-rose-400',   shadow: '0 0 30px rgba(244,63,94,0.08)' },
  sky:    { bg: 'rgba(14,165,233,0.08)',   border: 'rgba(14,165,233,0.2)',   text: 'text-sky-400',    shadow: '0 0 30px rgba(14,165,233,0.08)' },
  violet: { bg: 'rgba(124,58,237,0.08)',   border: 'rgba(124,58,237,0.2)',   text: 'text-violet-400', shadow: '0 0 30px rgba(124,58,237,0.08)' },
  amber:  { bg: 'rgba(245,158,11,0.08)',   border: 'rgba(245,158,11,0.2)',   text: 'text-amber-400',  shadow: '0 0 30px rgba(245,158,11,0.08)' },
  red:    { bg: 'rgba(239,68,68,0.08)',    border: 'rgba(239,68,68,0.2)',    text: 'text-red-400',    shadow: '0 0 30px rgba(239,68,68,0.08)' },
  lime:   { bg: 'rgba(132,204,22,0.08)',   border: 'rgba(132,204,22,0.2)',   text: 'text-lime-400',   shadow: '0 0 30px rgba(132,204,22,0.08)' },
}

export default function FeaturesSection() {
  return (
    <section className="section" id="features">
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-16">
          <span className="badge-cyan mb-5">Platform Features</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Every signal.{' '}
            <span className="text-gradient-cyan">One platform.</span>
          </h2>
          <p className="text-lg text-white/40 max-w-2xl mx-auto leading-relaxed">
            infraYS brings metrics, logs, traces, alerting, and cost tracking together in a single agent
            and unified dashboard — at a fraction of the cost.
          </p>
        </div>

        {/* Feature Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((f) => {
            const c = colorMap[f.color]
            return (
              <div key={f.title}
                className="group relative rounded-2xl p-6 transition-all duration-350 cursor-default overflow-hidden"
                style={{
                  background: 'rgba(13,13,24,0.7)',
                  border: `1px solid rgba(255,255,255,0.06)`,
                  backdropFilter: 'blur(8px)',
                }}>
                {/* Hover glow overlay */}
                <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-400 pointer-events-none rounded-2xl"
                  style={{ background: `radial-gradient(circle at 30% 30%, ${c.bg}, transparent 70%)` }} />
                {/* Hover border */}
                <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-400 pointer-events-none rounded-2xl"
                  style={{ boxShadow: `inset 0 0 0 1px ${c.border}, ${c.shadow}` }} />

                <div className="relative z-10">
                  {/* Icon */}
                  <div className="w-11 h-11 rounded-xl flex items-center justify-center mb-5 transition-transform duration-300 group-hover:scale-110"
                    style={{
                      background: c.bg,
                      border: `1px solid ${c.border}`,
                    }}>
                    <f.icon className={`w-5 h-5 ${c.text}`} />
                  </div>

                  <h3 className="text-base font-bold text-white mb-2 group-hover:text-white transition-colors">{f.title}</h3>
                  <p className="text-sm text-white/40 leading-relaxed mb-4">{f.desc}</p>

                  <div className="flex flex-wrap gap-2">
                    {f.tags.map((tag) => (
                      <span key={tag}
                        className={`text-xs px-2.5 py-1 rounded-full font-medium ${c.text}`}
                        style={{ background: c.bg, border: `1px solid ${c.border}` }}>
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        <div className="text-center mt-12 flex flex-wrap justify-center gap-4">
          <Link to="/features" className="btn-secondary">
            View All Features
            <ArrowRight className="w-4 h-4" />
          </Link>
          <Link to="/plugins" className="btn-secondary">
            Browse Plugin Catalog
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
