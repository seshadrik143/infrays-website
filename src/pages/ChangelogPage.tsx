import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CTABanner from '@/components/CTABanner'
import { CheckCircle2, Zap, Shield, Globe, Brain, Database, Puzzle, Cloud, Activity, Server, Code2, Package, GitBranch } from 'lucide-react'

const releases = [
  {
    version: 'v1.0.27',
    date: '2026-04-08',
    badge: 'Latest',
    badgeColor: 'badge-cyan',
    title: 'Alert Rule Library, Config Wizard & Coverage Score',
    icon: Activity,
    color: 'cyan',
    highlights: [
      'Alert Rule Library — browse, import, and customize pre-built alert rules',
      'Config Wizard — guided setup for collectors, alerting, and notifications',
      'Coverage Score — measure how well your infrastructure is monitored (0–100%)',
      'Alert coverage gaps highlighted directly in the dashboard',
    ],
  },
  {
    version: 'v1.0.26',
    date: '2026-04-07',
    badge: null,
    badgeColor: '',
    title: 'Enterprise Collector Plugins (20 plugins)',
    icon: Puzzle,
    color: 'violet',
    highlights: [
      '20 new enterprise-grade collector plugins added to the catalog',
      'ActiveMQ, Cassandra, Kafka, MongoDB, Elasticsearch plugins',
      'HAProxy, Vault, Consul, etcd, ZooKeeper collectors',
      'Total plugin catalog now at 67 community plugins',
    ],
  },
  {
    version: 'v1.0.25',
    date: '2026-04-07',
    badge: null,
    badgeColor: '',
    title: 'Tier 3 Enterprise Plugins (15 plugins)',
    icon: Puzzle,
    color: 'purple',
    highlights: [
      '15 Tier 3 enterprise collector plugins',
      'Oracle, MSSQL, Couchbase, InfluxDB, TimescaleDB support',
      'Nagios, Zabbix, Icinga compatibility bridges',
      'Custom exec protocol for any binary or script output',
    ],
  },
  {
    version: 'v1.0.24',
    date: '2026-04-06',
    badge: null,
    badgeColor: '',
    title: 'Tier 2 Collector Plugins (13 plugins)',
    icon: Puzzle,
    color: 'indigo',
    highlights: [
      '13 Tier 2 collector plugins: RabbitMQ, NATS, Memcached, etcd',
      'Windows WMI, JVM (JMX), PHP-FPM, Apache Tomcat collectors',
      'Plugin installation and management via npctl CLI',
    ],
  },
  {
    version: 'v1.0.23',
    date: '2026-04-06',
    badge: null,
    badgeColor: '',
    title: 'Tier 1 Collector Plugins (10 plugins)',
    icon: Code2,
    color: 'blue',
    highlights: [
      '10 Tier 1 plugins: MySQL, PostgreSQL, Redis, NGINX, HAProxy',
      'Docker extended metrics, Node.js runtime, Python runtime',
      'Plugin SDK v1.0.0 — write collectors in any language',
      'Notifier SDK for custom alert delivery channels',
    ],
  },
  {
    version: 'v1.0.21',
    date: '2026-04-08',
    badge: null,
    badgeColor: '',
    title: 'SaaS Mode & Commercial Launch',
    icon: Cloud,
    color: 'sky',
    highlights: [
      'Billing metering store — real-time agent/metric/log/span usage tracking',
      'Stripe integration — subscriptions, usage records, portal, invoices, HMAC webhooks',
      'Free / Starter ($49) / Pro ($199) / Enterprise pricing tiers',
      'Org provisioner — signup, trial periods, tier changes, onboarding checklist',
    ],
  },
  {
    version: 'v1.0.20',
    date: '2026-04-07',
    badge: null,
    badgeColor: '',
    title: 'Synthetic Monitoring & Global Probes',
    icon: Globe,
    color: 'teal',
    highlights: [
      'HTTP, TCP, DNS, SSL, multi-step API flows, and browser checks',
      'Global probe nodes — distributed monitoring from multiple regions',
      'httptrace-powered latency breakdown (DNS, TCP, TLS, TTFB)',
      'Public status page at /status — zero config, always live',
      'Alert integration — fires automatically on probe failure',
    ],
  },
  {
    version: 'v1.0.19',
    date: '2026-04-07',
    badge: null,
    badgeColor: '',
    title: 'Enterprise Hardening',
    icon: Shield,
    color: 'indigo',
    highlights: [
      'AES-256-GCM encryption service for secrets at rest',
      'Vault, AWS Secrets Manager, and GCP Secrets Manager integration',
      'Per-tenant resource quotas with live usage tracking',
      'GDPR erasure + data retention API (SOC2/ISO27001/GDPR controls)',
      'mTLS certificate generation for agent mutual authentication',
      'Compliance dashboard — 15 SOC2/ISO27001/GDPR controls',
      'PDF/HTML report export for compliance audits',
    ],
  },
  {
    version: 'v1.0.18',
    date: '2026-04-06',
    badge: null,
    badgeColor: '',
    title: 'Ecosystem & Extensibility',
    icon: Puzzle,
    color: 'violet',
    highlights: [
      'Plugin SDK v1.0.0 — Go interfaces + exec protocol for any language',
      'GitHub, Jira, ServiceNow, PagerDuty HMAC-verified webhook integrations',
      'Plugin marketplace — 17 seeded community plugins, BoltDB catalog',
      'Webhook template engine — Slack, Teams, PagerDuty, OpsGenie built-in',
      '67 pre-built dashboard JSON templates (system/web/db/cloud/SRE)',
      'Terraform provider — alert_rule, annotation, agents datasource',
      'Pulumi provider scaffold — AlertRule, Annotation, Dashboard resources',
      'GitHub Actions workflow for deploy annotations',
    ],
  },
  {
    version: 'v1.0.17',
    date: '2026-04-06',
    badge: null,
    badgeColor: '',
    title: 'eBPF Agent & Auto-Discovery',
    icon: Activity,
    color: 'green',
    highlights: [
      '/proc-based service scanning — discovers running services automatically',
      'TCP flow tracking with source/destination/PID mapping',
      'Cloud metadata detection — AWS/GCP/Azure IMDSv2',
      'Service topology map — visualize service dependencies as a graph',
      'Kubernetes operator — NodePulseAgent CRD + Go controller + RBAC',
      'Helm v0.2.0 — PDB, HPA templates, auto_discovery config block',
    ],
  },
  {
    version: 'v1.0.16',
    date: '2026-04-05',
    badge: null,
    badgeColor: '',
    title: 'Deep APM & Continuous Profiling',
    icon: Activity,
    color: 'orange',
    highlights: [
      'ProfilerCollector — runtime/pprof CPU and memory profiling',
      'Flame graph API — pprof → FlameNode JSON, click-to-zoom, search by function',
      'Error tracking — SHA-256 fingerprint dedup, first/last seen, error groups',
      'Service latency — P50/P95/P99 per service pair from Tempo spans',
      'Error rate per service edge in the topology map',
    ],
  },
  {
    version: 'v1.0.15',
    date: '2026-04-05',
    badge: null,
    badgeColor: '',
    title: 'Intelligent Alerting & AIOps',
    icon: Brain,
    color: 'pink',
    highlights: [
      'BaselineTracker — 2-week rolling hourly ring buffers, Z-score detection',
      'FlapDetector — 5-min window / 4-transition threshold, suppresses noisy alerts',
      'Correlator — groups related alerts within 60s per tenant into incidents',
      'Predictor — linear regression predicts threshold breaches up to 7 days ahead',
      'Alert analytics API — MTTR, fatigue scoring, correlations, baselines',
    ],
  },
  {
    version: 'v1.0.14',
    date: '2026-04-05',
    badge: null,
    badgeColor: '',
    title: 'Identity, SSO & Enterprise RBAC',
    icon: Shield,
    color: 'indigo',
    highlights: [
      'Custom RBAC roles — 11 permissions, BoltDB storage',
      'Built-in roles: admin, viewer, operator, developer',
      'Session tracking — revocation, TTL, purge',
      'OIDC/OAuth2 — Google, GitHub, Microsoft, custom IdP with CSRF nonces',
      'IP allowlist middleware — CIDR-based access control',
    ],
  },
  {
    version: 'v1.0.13',
    date: '2026-04-05',
    badge: null,
    badgeColor: '',
    title: 'OTel/OTLP, Prometheus Scrape & StatsD',
    icon: Database,
    color: 'blue',
    highlights: [
      'OTLP/HTTP JSON receivers — /v1/metrics, /v1/logs, /v1/traces',
      'Prometheus scrape endpoint at /metrics',
      'UDP StatsD listener on port 8125 with configurable flush interval',
      'Service map from Tempo CLIENT spans with latency + error rate',
    ],
  },
  {
    version: 'v1.0.11',
    date: '2026-04-05',
    badge: null,
    badgeColor: '',
    title: 'SMTP Email, GCP RSA JWT & Notifications UI',
    icon: Activity,
    color: 'yellow',
    highlights: [
      'Real SMTP email delivery — TLS port 465 + STARTTLS port 587',
      'Full GCP RS256 JWT signing — PKCS1 and PKCS8 PEM support',
      'Notification channel management UI',
      'SLO error budget and burn rate visualization',
      'React code splitting — bundle 1039 kB → 298 kB',
    ],
  },
  {
    version: 'v1.0.9',
    date: '2026-04-03',
    badge: null,
    badgeColor: '',
    title: 'Custom Collectors, Reports, Fleet Groups & Backup',
    icon: Package,
    color: 'rose',
    highlights: [
      'Custom script collectors — exec any binary, parse simple/JSON/influx output',
      'Scheduled reports — weekly summaries dispatched to Slack or webhook',
      'Fleet groups — label selector-based agent grouping with bulk commands',
      'OpenAPI 3.0 spec — 41 documented endpoints + Swagger UI at /api/v1/docs',
      'Backup/restore — streaming tar.gz via npctl CLI',
    ],
  },
  {
    version: 'v1.0.7',
    date: '2026-04-03',
    badge: null,
    badgeColor: '',
    title: 'High Availability, PostgreSQL & Cloud Costs',
    icon: Server,
    color: 'green',
    highlights: [
      'Raft consensus for HA — leader election, log replication',
      'PostgreSQL dual-mode storage — users, incidents, teams',
      'AWS / Azure / GCP cost import — visualize spend alongside metrics',
      'Multi-tenancy — per-tenant data isolation, tenant selector in dashboard',
    ],
  },
  {
    version: 'v1.0.5',
    date: '2026-04-02',
    badge: null,
    badgeColor: '',
    title: 'Dashboard Builder, Auto-Update & npctl CLI',
    icon: Code2,
    color: 'cyan',
    highlights: [
      'Dashboard builder with brush-zoom and annotation overlay',
      'Agent auto-update — semver check, SHA256 verify, atomic binary rename',
      'npctl CLI initial release — logs, oncall, annotations commands',
      'BoltDB annotation and dashboard stores',
    ],
  },
  {
    version: 'v1.0.4',
    date: '2026-04-01',
    badge: null,
    badgeColor: '',
    title: 'Distributed Tracing & Structured Logs',
    icon: GitBranch,
    color: 'blue',
    highlights: [
      'Grafana Tempo integration for distributed tracing',
      'Grafana Loki integration for log aggregation',
      'Grok parsing in the agent log collector',
      'Traces and Logs pages rewritten in dashboard',
    ],
  },
  {
    version: 'v1.0.3',
    date: '2026-03-31',
    badge: null,
    badgeColor: '',
    title: 'Docker, Network, K8s & AI Engine',
    icon: Brain,
    color: 'pink',
    highlights: [
      'Docker enhanced metrics and command channel',
      'Network collector — ping, DNS, TCP probing',
      'RTSP stream probe for video infrastructure',
      'Kubernetes collector — nodes, pods, namespaces',
      'Python AI engine — Z-score, Isolation Forest, RCA, LLM-powered analysis',
      'Microsoft Teams and PagerDuty notification channels',
      'Prometheus remote_write compatibility',
    ],
  },
  {
    version: 'v1.0.0',
    date: '2026-03-25',
    badge: 'Initial Release',
    badgeColor: 'badge-green',
    title: 'Core Agent & Server',
    icon: Zap,
    color: 'cyan',
    highlights: [
      'Single binary Go agent — system, process, Docker metrics',
      'VictoriaMetrics backend for high-throughput metrics storage',
      'JWT + API key authentication',
      'BoltDB persistent storage for all server data',
      'Alert engine with Slack, email, and webhook notifications',
      'React + TypeScript dashboard with real-time updates',
    ],
  },
]

const colorMap: Record<string, { icon: string; line: string }> = {
  cyan:   { icon: 'text-cyan-400 bg-cyan-500/10 border-cyan-500/30',   line: 'bg-cyan-500/20' },
  purple: { icon: 'text-purple-400 bg-purple-500/10 border-purple-500/30', line: 'bg-purple-500/20' },
  blue:   { icon: 'text-blue-400 bg-blue-500/10 border-blue-500/30',   line: 'bg-blue-500/20' },
  orange: { icon: 'text-orange-400 bg-orange-500/10 border-orange-500/30', line: 'bg-orange-500/20' },
  pink:   { icon: 'text-pink-400 bg-pink-500/10 border-pink-500/30',   line: 'bg-pink-500/20' },
  yellow: { icon: 'text-yellow-400 bg-yellow-500/10 border-yellow-500/30', line: 'bg-yellow-500/20' },
  teal:   { icon: 'text-teal-400 bg-teal-500/10 border-teal-500/30',   line: 'bg-teal-500/20' },
  green:  { icon: 'text-green-400 bg-green-500/10 border-green-500/30', line: 'bg-green-500/20' },
  indigo: { icon: 'text-indigo-400 bg-indigo-500/10 border-indigo-500/30', line: 'bg-indigo-500/20' },
  rose:   { icon: 'text-rose-400 bg-rose-500/10 border-rose-500/30',   line: 'bg-rose-500/20' },
  sky:    { icon: 'text-sky-400 bg-sky-500/10 border-sky-500/30',     line: 'bg-sky-500/20' },
  violet: { icon: 'text-violet-400 bg-violet-500/10 border-violet-500/30', line: 'bg-violet-500/20' },
}

export default function ChangelogPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Changelog</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              34 phases.{' '}
              <span className="text-gradient-cyan">One platform.</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              Every feature shipped — from the first agent heartbeat to enterprise compliance,
              SaaS billing, and a 67-plugin ecosystem.
            </p>
          </div>
        </section>

        {/* Timeline */}
        <section className="section py-16">
          <div className="container-md">
            <div className="relative">
              {/* Vertical line */}
              <div className="absolute left-[19px] top-0 bottom-0 w-px bg-white/[0.06] hidden md:block" />

              <div className="space-y-10">
                {releases.map((rel) => {
                  const c = colorMap[rel.color] ?? colorMap['cyan']
                  return (
                    <div key={rel.version} className="relative md:pl-14">
                      {/* Icon dot on timeline */}
                      <div className={`hidden md:flex absolute left-0 top-1 w-10 h-10 rounded-full border items-center justify-center ${c.icon}`}>
                        <rel.icon className="w-4 h-4" />
                      </div>

                      <div className="border border-white/[0.07] rounded-2xl p-6"
                        style={{ background: 'rgba(17,17,32,0.7)' }}>
                        {/* Header */}
                        <div className="flex flex-wrap items-center gap-3 mb-3">
                          <span className="font-mono text-xs text-white/30">{rel.version}</span>
                          {rel.badge && (
                            <span className={rel.badgeColor}>{rel.badge}</span>
                          )}
                          <span className="text-xs text-white/25 ml-auto">{rel.date}</span>
                        </div>

                        <h3 className="text-base font-bold text-white mb-4">{rel.title}</h3>

                        <ul className="space-y-2">
                          {rel.highlights.map((h) => (
                            <li key={h} className="flex items-start gap-3 text-sm text-white/55">
                              <CheckCircle2 className="w-4 h-4 text-green-400 flex-shrink-0 mt-0.5" />
                              {h}
                            </li>
                          ))}
                        </ul>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </section>

        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
