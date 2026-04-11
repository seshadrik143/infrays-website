import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'

const categories: { name: string; color: string; items: string[] }[] = [
  {
    name: 'Observability',
    color: 'cyan',
    items: ['OpenTelemetry', 'Prometheus', 'VictoriaMetrics', 'Grafana Tempo', 'Grafana Loki', 'StatsD', 'OTLP/HTTP'],
  },
  {
    name: 'Containers & Orchestration',
    color: 'blue',
    items: ['Kubernetes', 'Docker', 'Helm', 'NodePulseAgent CRD', 'containerd', 'Podman'],
  },
  {
    name: 'Cloud & Secrets',
    color: 'sky',
    items: ['AWS CloudWatch', 'AWS Secrets Manager', 'Azure Monitor', 'GCP Monitoring', 'GCP Secrets Manager', 'HashiCorp Vault'],
  },
  {
    name: 'Databases',
    color: 'orange',
    items: ['PostgreSQL', 'MySQL', 'Redis', 'MongoDB', 'Cassandra', 'Elasticsearch', 'Couchbase', 'InfluxDB', 'TimescaleDB', 'MSSQL', 'Oracle'],
  },
  {
    name: 'Message Queues',
    color: 'yellow',
    items: ['Apache Kafka', 'RabbitMQ', 'ActiveMQ', 'NATS', 'ZooKeeper'],
  },
  {
    name: 'Web & Proxy',
    color: 'green',
    items: ['NGINX', 'HAProxy', 'Apache Tomcat', 'PHP-FPM'],
  },
  {
    name: 'Alerting & Incident',
    color: 'red',
    items: ['PagerDuty', 'Slack', 'Microsoft Teams', 'Email (SMTP)', 'OpsGenie', 'Webhooks'],
  },
  {
    name: 'Ticketing & ITSM',
    color: 'indigo',
    items: ['Jira', 'ServiceNow', 'GitHub Issues'],
  },
  {
    name: 'CI/CD & IaC',
    color: 'violet',
    items: ['GitHub Actions', 'Terraform Provider', 'Pulumi Provider'],
  },
  {
    name: 'Service Discovery',
    color: 'teal',
    items: ['Consul', 'etcd', 'AWS IMDSv2', 'GCP Metadata', 'Azure Instance Metadata'],
  },
  {
    name: 'Runtimes & APM',
    color: 'pink',
    items: ['JVM (JMX)', 'Node.js', 'Python', 'Go pprof', 'Windows WMI'],
  },
  {
    name: 'Legacy Monitoring',
    color: 'rose',
    items: ['Nagios', 'Zabbix', 'Icinga', 'Memcached', 'RTSP / Video'],
  },
]

const colorMap: Record<string, { badge: string; dot: string; card: string }> = {
  cyan:   { badge: 'bg-cyan-500/15 text-cyan-400 border-cyan-500/25',        dot: 'bg-cyan-400',    card: 'border-cyan-500/15' },
  blue:   { badge: 'bg-blue-500/15 text-blue-400 border-blue-500/25',        dot: 'bg-blue-400',    card: 'border-blue-500/15' },
  sky:    { badge: 'bg-sky-500/15 text-sky-400 border-sky-500/25',           dot: 'bg-sky-400',     card: 'border-sky-500/15' },
  orange: { badge: 'bg-orange-500/15 text-orange-400 border-orange-500/25',  dot: 'bg-orange-400',  card: 'border-orange-500/15' },
  yellow: { badge: 'bg-yellow-500/15 text-yellow-400 border-yellow-500/25',  dot: 'bg-yellow-400',  card: 'border-yellow-500/15' },
  green:  { badge: 'bg-green-500/15 text-green-400 border-green-500/25',     dot: 'bg-green-400',   card: 'border-green-500/15' },
  red:    { badge: 'bg-red-500/15 text-red-400 border-red-500/25',           dot: 'bg-red-400',     card: 'border-red-500/15' },
  indigo: { badge: 'bg-indigo-500/15 text-indigo-400 border-indigo-500/25',  dot: 'bg-indigo-400',  card: 'border-indigo-500/15' },
  violet: { badge: 'bg-violet-500/15 text-violet-400 border-violet-500/25',  dot: 'bg-violet-400',  card: 'border-violet-500/15' },
  teal:   { badge: 'bg-teal-500/15 text-teal-400 border-teal-500/25',        dot: 'bg-teal-400',    card: 'border-teal-500/15' },
  pink:   { badge: 'bg-pink-500/15 text-pink-400 border-pink-500/25',        dot: 'bg-pink-400',    card: 'border-pink-500/15' },
  rose:   { badge: 'bg-rose-500/15 text-rose-400 border-rose-500/25',        dot: 'bg-rose-400',    card: 'border-rose-500/15' },
}

const totalIntegrations = categories.reduce((acc, c) => acc + c.items.length, 0)

export default function IntegrationsGrid() {
  return (
    <section className="section" id="integrations"
      style={{ background: 'rgba(8, 8, 16, 0.8)' }}>
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-purple mb-4">Integrations</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            {totalIntegrations}+ integrations.{' '}
            <span className="text-gradient-purple">Zero vendor lock-in.</span>
          </h2>
          <p className="text-lg text-white/40 max-w-2xl mx-auto">
            Native OTLP means every OpenTelemetry-compatible tool works out of the box.
            67 community collector plugins, cloud secrets managers, IaC providers, and full ITSM integration.
          </p>
        </div>

        {/* Category grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5 mb-10">
          {categories.map((cat) => {
            const c = colorMap[cat.color]
            return (
              <div key={cat.name}
                className={`border ${c.card} rounded-2xl p-5`}
                style={{ background: 'rgba(17,17,32,0.7)' }}>
                <div className="flex items-center gap-2 mb-4">
                  <span className={`text-xs px-2.5 py-1 rounded-full border font-medium ${c.badge}`}>
                    {cat.name}
                  </span>
                  <span className="text-xs text-white/25 ml-auto">{cat.items.length}</span>
                </div>
                <div className="flex flex-wrap gap-2">
                  {cat.items.map((item) => (
                    <span key={item}
                      className="flex items-center gap-1.5 text-xs text-white/55 bg-white/[0.04] border border-white/[0.06] px-2.5 py-1 rounded-full hover:text-white/80 hover:border-white/10 transition-colors cursor-default">
                      <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${c.dot}`} />
                      {item}
                    </span>
                  ))}
                </div>
              </div>
            )
          })}
        </div>

        {/* Plugin SDK CTA */}
        <div className="border border-white/[0.07] rounded-2xl p-8 flex flex-col md:flex-row items-center justify-between gap-6"
          style={{ background: 'linear-gradient(135deg, rgba(139,92,246,0.08), rgba(0,212,255,0.04))' }}>
          <div>
            <h3 className="text-lg font-bold text-white mb-2">Build your own integration</h3>
            <p className="text-sm text-white/40 max-w-md">
              The infraYS Plugin SDK lets you write custom collectors in any language via a simple exec protocol.
              Publish to the community marketplace.
            </p>
          </div>
          <Link to="/docs#sdk" className="btn-secondary whitespace-nowrap flex-shrink-0">
            Plugin SDK Docs
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </section>
  )
}
