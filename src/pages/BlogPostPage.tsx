import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CTABanner from '@/components/CTABanner'
import { Link, useParams, Navigate } from 'react-router-dom'
import { ArrowLeft, Calendar, Clock, ArrowRight } from 'lucide-react'

const posts: Record<string, {
  title: string
  category: string
  color: string
  date: string
  readTime: string
  excerpt: string
  body: string[]
}> = {
  'infrays-v1-launch': {
    title: 'Introducing infraYS v1.0: The Observability Platform Built for Teams That Never Sleep',
    category: 'Announcement',
    color: 'cyan',
    date: '2026-04-08',
    readTime: '8 min read',
    excerpt: 'After 27 development phases, infraYS v1.0 is here.',
    body: [
      'After 27 development phases and thousands of lines of Go, TypeScript, and Python, infraYS v1.0 is generally available. This post covers what we built, why we built it the way we did, and what comes next.',
      '## What is infraYS?',
      'infraYS is a unified observability platform. It replaces the typical monitoring stack — Prometheus + Grafana + Loki + Tempo + an APM tool + an alerting tool + an on-call tool — with a single binary agent and a unified server. You get metrics, logs, traces, continuous profiling, AI-powered anomaly detection, synthetic monitoring, SLO tracking, and cloud cost tracking from one install.',
      '## Why build this?',
      'Modern infrastructure monitoring is fragmented. Most teams run 5–8 separate tools, each with its own agent, its own storage, its own alerting logic, and its own billing. We built infraYS to collapse that into one coherent system — with a 12MB agent, < 30MB memory footprint, and an Apache 2.0 license.',
      '## What shipped in v0.33',
      'The v0.33 release includes everything from the first heartbeat to enterprise compliance: 34 phases covering metrics collection, structured log ingestion, distributed tracing, continuous profiling with flame graphs, AI anomaly detection, predictive alerting, flap detection, synthetic monitoring, SLO tracking, cloud cost import, OIDC SSO, RBAC, AES-256 encryption, Vault integration, GDPR erasure, SOC2/ISO27001 compliance reports, a 67-plugin ecosystem, Terraform and Pulumi providers, a fully managed SaaS mode with Stripe billing, and a 15-day self-hosted trial with license key enforcement.',
      '## What\'s next',
      'v0.34 will focus on the hosted cloud offering — streamlining the onboarding flow and expanding global synthetic probe coverage. Community contributions are welcome. The codebase is Apache 2.0 licensed.',
    ],
  },
  'ai-anomaly-detection': {
    title: 'How infraYS AI Catches Anomalies Before They Become Incidents',
    category: 'Engineering',
    color: 'purple',
    date: '2026-04-06',
    readTime: '12 min read',
    excerpt: 'Deep dive into the Z-score + Isolation Forest pipeline.',
    body: [
      'Every monitoring system can tell you something is broken. The hard part is telling you before it breaks. This post explains how infraYS AI works under the hood.',
      '## The Baseline Problem',
      'Most alerting systems use static thresholds: alert when CPU > 90%. This is fine for obvious failures but terrible for subtle degradation. CPU at 70% is fine on Monday morning but alarming at 3am on Sunday. Static thresholds produce false positives or miss real problems depending on what you set.',
      '## Rolling Baselines',
      'infraYS maintains a 2-week rolling ring buffer of hourly metric summaries per agent per metric. This gives us a rich statistical baseline: what is "normal" CPU usage for this agent at this hour on this day of the week?',
      '## Z-Score Detection',
      'We apply Z-score analysis against the rolling baseline. A Z-score above 3.0 (three standard deviations from the mean for that time slot) triggers an anomaly candidate. This alone handles most cases.',
      '## Isolation Forest',
      'For multivariate anomalies — where no single metric is extreme but the combination is unusual — we run an Isolation Forest algorithm. This catches cases like "low CPU + high memory + low disk I/O" which individually look fine but together indicate a hung process.',
      '## Flap Detection',
      'Anomalies that repeatedly fire and resolve within a 5-minute window are flagged as "flapping" and their notifications are suppressed. A flapping alert still appears in the UI — it just stops paging the on-call engineer every 2 minutes.',
      '## Predictive Alerting',
      'We run linear regression over the last 24 hours of a metric trend. If R² ≥ 0.5 (a reasonably confident trend), we project forward and flag metrics predicted to breach their threshold within the next 7 days. This turns reactive monitoring into proactive monitoring.',
    ],
  },
  'synthetic-monitoring-global-probes': {
    title: 'Synthetic Monitoring at Scale: HTTP, TCP, DNS, SSL, and Browser Checks',
    category: 'Engineering',
    color: 'green',
    date: '2026-04-05',
    readTime: '10 min read',
    excerpt: 'How we built the synthetic probe system.',
    body: [
      'Real-user monitoring tells you what your users experienced. Synthetic monitoring tells you what they\'ll experience before they try. This post covers how infraYS synthetic monitoring works.',
      '## Check Types',
      'infraYS supports HTTP checks (with full httptrace breakdown: DNS, TCP, TLS handshake, TTFB), TCP checks (raw port reachability), DNS checks (resolution time and answer validation), SSL/TLS checks (certificate expiry and chain validation), multi-step API flows (chained HTTP requests with response extraction), and headless browser checks.',
      '## httptrace Breakdown',
      'For HTTP checks, we use Go\'s net/http/httptrace package to instrument every phase of the connection lifecycle. The result is a latency breakdown that shows exactly where time is spent: DNS resolution, TCP connection, TLS handshake, time to first byte, and total transfer time.',
      '## Probe Nodes',
      'Checks run from probe nodes — lightweight binaries that register with the server and receive check schedules. You can run probe nodes in multiple regions, cloud accounts, or even on-premises networks. This gives you a distributed vantage point without needing a commercial CDN.',
      '## Status Page',
      'Every infraYS instance automatically generates a public status page at /status. It aggregates synthetic check results into service-level statuses. No template to configure — just define your checks and the status page builds itself.',
      '## Alert Integration',
      'When a probe check fails, it fires through the same alerting pipeline as metric alerts — meaning you get the same on-call routing, escalation policies, and notification channels for infrastructure and synthetic failures.',
    ],
  },
  'self-hosted-vs-saas-observability': {
    title: 'Self-Hosted vs SaaS Observability: Understanding the Cost Difference',
    category: 'Guide',
    color: 'orange',
    date: '2026-04-04',
    readTime: '7 min read',
    excerpt: 'A breakdown of what it actually costs to run a full observability stack.',
    body: [
      'The pitch for SaaS observability tools is compelling: no ops burden, no infrastructure to manage, just pay and plug in. But the economics get complicated at scale. Here\'s how to think about the real cost.',
      '## The SaaS Bill',
      'Most SaaS observability tools charge on a combination of hosts, metrics, log volume, and trace volume. A 50-agent deployment with moderate log volume can easily hit $3,000–8,000/month on commercial platforms. At 200 agents, you\'re looking at $15,000–40,000/month.',
      '## The Self-Hosted Cost',
      'Self-hosting infraYS on a modern cloud adds: storage for VictoriaMetrics (typically 1–5GB/day for 50 agents), compute for the server (2 CPU / 4GB RAM handles hundreds of agents), and ops time (typically 2–4 hours/month once running). For 50 agents, realistic cloud cost is $80–200/month.',
      '## Where SaaS Wins',
      'SaaS wins when: your team has no infra ops experience, you need managed uptime guarantees, your data retention requirements are long (> 1 year), or you need features that are hard to self-host (global probe networks, dedicated support SLAs).',
      '## Where Self-Hosting Wins',
      'Self-hosting wins when: you have a team that can handle ops, your data is sensitive and can\'t leave your network, you\'re cost-conscious at scale, or you want the flexibility to customize the stack.',
      '## The infraYS Middle Ground',
      'infraYS is designed for both. The Apache 2.0-licensed NodePulse server runs on your infrastructure at a fraction of the SaaS cost. The managed cloud tier ($49–$199/month) covers teams that want the economics of self-hosting with the ops simplicity of SaaS.',
    ],
  },
  'ebpf-autodiscovery': {
    title: 'Auto-Discovery Without eBPF: How We Map Services Using /proc',
    category: 'Engineering',
    color: 'teal',
    date: '2026-04-03',
    readTime: '9 min read',
    excerpt: 'Full eBPF requires kernel 5.8+. We built a /proc-based auto-discovery engine that works everywhere.',
    body: [
      'eBPF is the hot technology for service auto-discovery. But eBPF requires kernel 5.8+ and root privileges, which rules it out for a lot of infrastructure. infraYS auto-discovery works on kernel 3.x and above — and still gives you a complete service topology.',
      '## What /proc Tells Us',
      'Linux\'s /proc filesystem exposes everything: running processes, their open file descriptors, their listening sockets, their namespace memberships. By reading /proc/net/tcp, /proc/net/tcp6, and cross-referencing with /proc/*/fd, we can map every listening port to its owning process without any kernel hooks.',
      '## Service Detection',
      'We scan /proc every 30 seconds (configurable) and identify services by their listening port patterns. Port 5432 → PostgreSQL, port 6379 → Redis, port 9200 → Elasticsearch — we maintain a registry of 60+ well-known port/process mappings and also detect custom services by process name.',
      '## TCP Flow Tracking',
      '/proc/net/tcp shows established connections with source and destination. By correlating these with process IDs, we build a directed graph of which process is connecting to which — revealing service dependencies without any instrumentation.',
      '## Cloud Metadata',
      'On cloud instances, we hit the instance metadata service (AWS IMDSv2, GCP metadata, Azure IMDS) to enrich discovered services with cloud-native labels: instance ID, region, availability zone, instance type, and cloud provider.',
      '## The Topology Map',
      'The collected data feeds the service topology map in the infraYS dashboard — a live graph of services, their connections, latency (from Tempo spans when available), and error rates. No instrumentation required beyond installing the agent.',
    ],
  },
  'continuous-profiling-go': {
    title: 'Finding the Hidden 40ms: Continuous Profiling for Production Go Services',
    category: 'Case Study',
    color: 'pink',
    date: '2026-04-01',
    readTime: '6 min read',
    excerpt: 'A case study of how infraYS continuous profiling identified a hidden 40ms latency hotspot.',
    body: [
      'P99 latency was 40ms higher than it should have been. Logs showed nothing. Metrics showed nothing. We enabled continuous profiling and found the answer in 8 minutes.',
      '## The Problem',
      'A Go HTTP service was consistently running 40ms over its SLA target at P99. No errors, no panics, no obvious resource exhaustion. Traditional metrics — CPU, memory, goroutine count — all looked normal.',
      '## Enabling Continuous Profiling',
      'infraYS continuous profiling uses Go\'s runtime/pprof package to capture 10-second CPU profiles every 60 seconds. No code changes required — the ProfilerCollector runs as a goroutine inside the agent, profiles the target process by PID, and ships the pprof data to the server.',
      '## Reading the Flame Graph',
      'The flame graph immediately showed an unexpectedly wide bar for json.Unmarshal in a middleware function. The middleware was deserializing a large JSON config blob on every request to check a feature flag — a config that rarely changed.',
      '## The Fix',
      'Moving the deserialization to startup (cache once, read many) dropped P99 latency by 43ms. The fix was 4 lines of Go. Finding the problem took 8 minutes with a flame graph. Without profiling, it would have taken days of hypothesis-driven debugging.',
      '## Why Continuous vs On-Demand',
      'On-demand profiling is great for known problems. Continuous profiling catches the unknown ones — the hot path that only shows up under a specific traffic pattern, the allocator pressure that builds over 6 hours, the background goroutine that only misbehaves at 3am. infraYS collects profiles 24/7 so you always have the data when you need it.',
    ],
  },
}

const colorBadge: Record<string, string> = {
  cyan:   'badge-cyan',
  purple: 'bg-purple-500/10 text-purple-400 border border-purple-500/20 text-xs px-2.5 py-1 rounded-full font-medium',
  green:  'badge-green',
  orange: 'bg-orange-500/10 text-orange-400 border border-orange-500/20 text-xs px-2.5 py-1 rounded-full font-medium',
  teal:   'bg-teal-500/10 text-teal-400 border border-teal-500/20 text-xs px-2.5 py-1 rounded-full font-medium',
  pink:   'bg-pink-500/10 text-pink-400 border border-pink-500/20 text-xs px-2.5 py-1 rounded-full font-medium',
}

export default function BlogPostPage() {
  const { slug } = useParams<{ slug: string }>()
  const post = slug ? posts[slug] : null

  if (!post) {
    return <Navigate to="/blog" replace />
  }

  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Header */}
        <section className="hero-bg section py-12 border-b border-white/[0.06]">
          <div className="container-md">
            <Link to="/blog"
              className="inline-flex items-center gap-2 text-sm text-white/40 hover:text-white/70 transition-colors mb-8">
              <ArrowLeft className="w-4 h-4" />
              Back to Blog
            </Link>

            <div className="flex items-center gap-3 mb-5">
              <span className={colorBadge[post.color]}>{post.category}</span>
              <span className="flex items-center gap-1.5 text-xs text-white/30">
                <Calendar className="w-3.5 h-3.5" />{post.date}
              </span>
              <span className="flex items-center gap-1.5 text-xs text-white/30">
                <Clock className="w-3.5 h-3.5" />{post.readTime}
              </span>
            </div>

            <h1 className="text-4xl md:text-5xl font-black tracking-tight leading-tight text-white">
              {post.title}
            </h1>
          </div>
        </section>

        {/* Body */}
        <section className="section py-12">
          <div className="container-md">
            <div className="prose-infrays max-w-none">
              {post.body.map((block, i) => {
                if (block.startsWith('## ')) {
                  return (
                    <h2 key={i} className="text-2xl font-black text-white mt-10 mb-4">
                      {block.slice(3)}
                    </h2>
                  )
                }
                return (
                  <p key={i} className="text-white/60 leading-relaxed mb-5 text-base">
                    {block}
                  </p>
                )
              })}
            </div>

            {/* Back + next actions */}
            <div className="mt-16 pt-8 border-t border-white/[0.06] flex flex-col sm:flex-row gap-4">
              <Link to="/blog" className="btn-secondary">
                <ArrowLeft className="w-4 h-4" />
                All Posts
              </Link>
              <Link to="/install" className="btn-primary ml-auto">
                Try infraYS Free
                <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        </section>

        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
