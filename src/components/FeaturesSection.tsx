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
    desc: 'Flap detection, alert correlation, baseline-aware thresholds, and on-call scheduling. Fire fewer, better alerts. Integrates with PagerDuty, Slack, Teams.',
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
    desc: '67+ community plugins, custom collector SDK, webhook & dashboard templates, and integrations with GitHub, Jira, ServiceNow, and PagerDuty. Terraform & Pulumi providers included.',
    tags: ['67 Plugins', 'SDK', 'Terraform', 'Pulumi'],
  },
  {
    icon: Network,
    color: 'amber',
    title: 'Auto-Discovery & Topology',
    desc: '/proc-based service scan, TCP flow tracking, and cloud metadata detection (AWS/GCP/Azure IMDSv2). Visualize your entire service topology automatically.',
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

const colorMap: Record<string, { bg: string; border: string; text: string; glow: string }> = {
  cyan:   { bg: 'bg-cyan-500/10',   border: 'border-cyan-500/20',   text: 'text-cyan-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(0,212,255,0.12)]' },
  purple: { bg: 'bg-purple-500/10', border: 'border-purple-500/20', text: 'text-purple-400', glow: 'group-hover:shadow-[0_0_30px_rgba(139,92,246,0.12)]' },
  blue:   { bg: 'bg-blue-500/10',   border: 'border-blue-500/20',   text: 'text-blue-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(59,130,246,0.12)]' },
  orange: { bg: 'bg-orange-500/10', border: 'border-orange-500/20', text: 'text-orange-400', glow: 'group-hover:shadow-[0_0_30px_rgba(249,115,22,0.12)]' },
  pink:   { bg: 'bg-pink-500/10',   border: 'border-pink-500/20',   text: 'text-pink-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(236,72,153,0.12)]' },
  yellow: { bg: 'bg-yellow-500/10', border: 'border-yellow-500/20', text: 'text-yellow-400', glow: 'group-hover:shadow-[0_0_30px_rgba(234,179,8,0.12)]' },
  teal:   { bg: 'bg-teal-500/10',   border: 'border-teal-500/20',   text: 'text-teal-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(20,184,166,0.12)]' },
  green:  { bg: 'bg-green-500/10',  border: 'border-green-500/20',  text: 'text-green-400',  glow: 'group-hover:shadow-[0_0_30px_rgba(16,185,129,0.12)]' },
  indigo: { bg: 'bg-indigo-500/10', border: 'border-indigo-500/20', text: 'text-indigo-400', glow: 'group-hover:shadow-[0_0_30px_rgba(99,102,241,0.12)]' },
  rose:   { bg: 'bg-rose-500/10',   border: 'border-rose-500/20',   text: 'text-rose-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(244,63,94,0.12)]' },
  sky:    { bg: 'bg-sky-500/10',    border: 'border-sky-500/20',    text: 'text-sky-400',    glow: 'group-hover:shadow-[0_0_30px_rgba(14,165,233,0.12)]' },
  violet: { bg: 'bg-violet-500/10', border: 'border-violet-500/20', text: 'text-violet-400', glow: 'group-hover:shadow-[0_0_30px_rgba(124,58,237,0.12)]' },
  amber:  { bg: 'bg-amber-500/10',  border: 'border-amber-500/20',  text: 'text-amber-400',  glow: 'group-hover:shadow-[0_0_30px_rgba(245,158,11,0.12)]' },
  red:    { bg: 'bg-red-500/10',    border: 'border-red-500/20',    text: 'text-red-400',    glow: 'group-hover:shadow-[0_0_30px_rgba(239,68,68,0.12)]' },
  lime:   { bg: 'bg-lime-500/10',   border: 'border-lime-500/20',   text: 'text-lime-400',   glow: 'group-hover:shadow-[0_0_30px_rgba(132,204,22,0.12)]' },
}

export default function FeaturesSection() {
  return (
    <section className="section" id="features">
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-16">
          <span className="badge-cyan mb-4">Platform Features</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Every signal. One platform.
          </h2>
          <p className="text-lg text-white/40 max-w-2xl mx-auto">
            infraYS brings metrics, logs, traces, alerting, and cost tracking together in a single agent
            and unified dashboard — at a fraction of the cost.
          </p>
        </div>

        {/* Feature Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
          {features.map((f) => {
            const c = colorMap[f.color]
            return (
              <div key={f.title}
                className={`feature-card group border border-white/[0.07] rounded-2xl p-6 ${c.glow} transition-all duration-300`}
                style={{ background: 'rgba(17, 17, 32, 0.6)' }}>
                <div className={`w-11 h-11 rounded-xl ${c.bg} border ${c.border} flex items-center justify-center mb-5`}>
                  <f.icon className={`w-5 h-5 ${c.text}`} />
                </div>
                <h3 className="text-base font-bold text-white mb-2">{f.title}</h3>
                <p className="text-sm text-white/40 leading-relaxed mb-4">{f.desc}</p>
                <div className="flex flex-wrap gap-2">
                  {f.tags.map((tag) => (
                    <span key={tag} className={`text-xs px-2.5 py-1 rounded-full ${c.bg} ${c.text} font-medium`}>
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )
          })}
        </div>

        <div className="text-center mt-12">
          <Link to="/features" className="btn-secondary">
            View All Features
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
