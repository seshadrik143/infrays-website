import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'

const integrations = [
  { name: 'Prometheus', category: 'Metrics', color: '#e6522c' },
  { name: 'OpenTelemetry', category: 'Observability', color: '#4ca4e0' },
  { name: 'Kubernetes', category: 'Orchestration', color: '#326ce5' },
  { name: 'Docker', category: 'Containers', color: '#2496ed' },
  { name: 'Grafana Tempo', category: 'Tracing', color: '#f5810e' },
  { name: 'Grafana Loki', category: 'Logging', color: '#f5810e' },
  { name: 'VictoriaMetrics', category: 'Storage', color: '#621dee' },
  { name: 'AWS CloudWatch', category: 'Cloud', color: '#ff9900' },
  { name: 'Azure Monitor', category: 'Cloud', color: '#0078d4' },
  { name: 'Google Cloud', category: 'Cloud', color: '#4285f4' },
  { name: 'PagerDuty', category: 'Alerting', color: '#06ac38' },
  { name: 'Slack', category: 'Notifications', color: '#4a154b' },
  { name: 'Microsoft Teams', category: 'Notifications', color: '#6264a7' },
  { name: 'GitHub', category: 'CI/CD', color: '#6e40c9' },
  { name: 'Jira', category: 'Ticketing', color: '#0052cc' },
  { name: 'ServiceNow', category: 'ITSM', color: '#62d84e' },
  { name: 'Terraform', category: 'IaC', color: '#7b42bc' },
  { name: 'Pulumi', category: 'IaC', color: '#8a3391' },
  { name: 'StatsD', category: 'Metrics', color: '#44cc11' },
  { name: 'Elasticsearch', category: 'Search', color: '#00bfb3' },
  { name: 'MySQL', category: 'Database', color: '#4479a1' },
  { name: 'PostgreSQL', category: 'Database', color: '#336791' },
  { name: 'Redis', category: 'Cache', color: '#dc382d' },
  { name: 'NGINX', category: 'Web', color: '#009639' },
]

export default function IntegrationsGrid() {
  return (
    <section className="section" id="integrations"
      style={{ background: 'rgba(8, 8, 16, 0.8)' }}>
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-purple mb-4">Integrations</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Works with your{' '}
            <span className="text-gradient-purple">entire stack</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto">
            infraYS connects to everything you already use. Native OTLP support means any
            OpenTelemetry-compatible tool works out of the box.
          </p>
        </div>

        {/* Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3 mb-10">
          {integrations.map((integration) => (
            <div key={integration.name}
              className="group border border-white/[0.07] rounded-xl p-4 text-center hover:border-white/20 transition-all duration-300 cursor-default hover:-translate-y-1"
              style={{ background: 'rgba(17, 17, 32, 0.6)' }}>
              {/* Color dot */}
              <div className="w-8 h-8 rounded-full mx-auto mb-3 flex items-center justify-center text-sm font-bold text-white"
                style={{ background: `${integration.color}25`, border: `1px solid ${integration.color}40` }}>
                <span style={{ color: integration.color }}>{integration.name[0]}</span>
              </div>
              <div className="text-xs font-semibold text-white/70 group-hover:text-white transition-colors leading-tight">
                {integration.name}
              </div>
              <div className="text-xs text-white/25 mt-0.5">{integration.category}</div>
            </div>
          ))}
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
