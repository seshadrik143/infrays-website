
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Book, Code2, Terminal, Server, Puzzle, Shield } from 'lucide-react'

const sections = [
  {
    icon: Terminal,
    title: 'Getting Started',
    color: 'cyan',
    articles: [
      { title: 'Quick Start Guide', href: '/install', desc: 'Deploy in 60 seconds' },
      { title: 'Architecture Overview', href: '#architecture', desc: 'How infraYS works' },
      { title: 'First Dashboard', href: '#dashboard', desc: 'Create your first view' },
      { title: 'Set Up Alerting', href: '#alerts', desc: 'Get notified when it matters' },
    ],
  },
  {
    icon: Code2,
    title: 'Agent Configuration',
    color: 'purple',
    articles: [
      { title: 'agent.yaml Reference', href: '#agent-config', desc: 'Full config options' },
      { title: 'Built-in Collectors', href: '#collectors', desc: 'System, Docker, K8s, RTSP...' },
      { title: 'Custom Collectors', href: '#custom', desc: 'Write your own scripts' },
      { title: 'Auto-Discovery', href: '#discovery', desc: '/proc-based service detection' },
    ],
  },
  {
    icon: Server,
    title: 'Server Setup',
    color: 'green',
    articles: [
      { title: 'config.yaml Reference', href: '#server-config', desc: 'Server configuration' },
      { title: 'PostgreSQL Mode', href: '#postgres', desc: 'Production data storage' },
      { title: 'High Availability', href: '#ha', desc: 'Raft consensus setup' },
      { title: 'Docker Compose', href: '#compose', desc: 'Full stack locally' },
    ],
  },
  {
    icon: Book,
    title: 'API Reference',
    color: 'orange',
    articles: [
      { title: 'Swagger UI', href: 'http://localhost:8080/api/v1/docs', desc: 'Interactive API explorer' },
      { title: 'Authentication', href: '#auth', desc: 'JWT + API keys' },
      { title: 'Metrics API', href: '#metrics-api', desc: 'Query and write metrics' },
      { title: 'Webhooks', href: '#webhooks', desc: 'Outbound event hooks' },
    ],
  },
  {
    icon: Puzzle,
    title: 'Plugin SDK',
    color: 'teal',
    articles: [
      { title: 'Plugin SDK v1.0', href: '#sdk', desc: 'Build custom collectors' },
      { title: 'Exec Protocol', href: '#exec', desc: 'Any language, any binary' },
      { title: 'Notifier SDK', href: '#notifier', desc: 'Custom alert channels' },
      { title: 'Community Plugins', href: '#marketplace', desc: 'Browse the marketplace' },
    ],
  },
  {
    icon: Shield,
    title: 'Enterprise & Security',
    color: 'indigo',
    articles: [
      { title: 'OIDC / SSO Setup', href: '#oidc', desc: 'Google, GitHub, Microsoft' },
      { title: 'RBAC Guide', href: '#rbac', desc: 'Roles and permissions' },
      { title: 'mTLS Configuration', href: '#mtls', desc: 'Agent mutual TLS' },
      { title: 'Compliance Reports', href: '#compliance', desc: 'SOC2, ISO27001, GDPR' },
    ],
  },
]

const colorMap: Record<string, { icon: string; badge: string }> = {
  cyan:   { icon: 'text-cyan-400 bg-cyan-500/10 border-cyan-500/20',   badge: 'badge-cyan' },
  purple: { icon: 'text-purple-400 bg-purple-500/10 border-purple-500/20', badge: 'badge-purple' },
  green:  { icon: 'text-green-400 bg-green-500/10 border-green-500/20',   badge: 'badge-green' },
  orange: { icon: 'text-orange-400 bg-orange-500/10 border-orange-500/20', badge: '' },
  teal:   { icon: 'text-teal-400 bg-teal-500/10 border-teal-500/20',     badge: '' },
  indigo: { icon: 'text-indigo-400 bg-indigo-500/10 border-indigo-500/20', badge: '' },
}

export default function DocsPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Documentation</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Everything you need to{' '}
              <span className="text-gradient-cyan">get running</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto mb-8">
              Full reference docs, how-to guides, and API specs.
            </p>
            <div className="flex items-center justify-center gap-4">
              <Link to="/install" className="btn-primary">
                Quick Start
                <ArrowRight className="w-4 h-4" />
              </Link>
              <a href="http://localhost:8080/api/v1/docs" target="_blank" rel="noopener noreferrer" className="btn-secondary">
                <Code2 className="w-4 h-4" />
                API Reference
              </a>
            </div>
          </div>
        </section>

        {/* Quick Links */}
        <section className="section py-12 border-b border-white/[0.06]"
          style={{ background: 'rgba(10,10,20,0.5)' }}>
          <div className="container-md">
            <div className="terminal rounded-2xl">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
                <span className="ml-auto text-xs text-white/20 font-mono">infrays — infraYS CLI</span>
              </div>
              <div className="p-6 grid md:grid-cols-2 gap-x-8 gap-y-2">
                {[
                  ['infrays agents list', 'List all registered agents'],
                  ['infrays alerts list', 'View active alerts'],
                  ['infrays slo summary', 'SLO burn rate summary'],
                  ['infrays logs tail --agent prod-01', 'Tail agent logs'],
                  ['infrays admin backup', 'Backup all data to tar.gz'],
                  ['infrays groups list', 'List agent groups'],
                  ['infrays annotations create', 'Add deployment annotation'],
                  ['infrays health', 'Check server health'],
                ].map(([cmd, desc]) => (
                  <div key={cmd} className="flex items-baseline gap-3">
                    <code className="text-cyan-400 text-xs font-mono flex-shrink-0">{cmd}</code>
                    <span className="text-xs text-white/30"># {desc}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Doc Sections */}
        <section className="section">
          <div className="container-lg">
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
              {sections.map((section) => {
                const c = colorMap[section.color]
                return (
                  <div key={section.title}
                    className="border border-white/[0.07] rounded-2xl p-6"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <div className={`w-10 h-10 rounded-xl border flex items-center justify-center mb-5 ${c.icon}`}>
                      <section.icon className="w-5 h-5" />
                    </div>
                    <h2 className="text-base font-bold text-white mb-4">{section.title}</h2>
                    <ul className="space-y-2">
                      {section.articles.map((article) => (
                        <li key={article.title}>
                          <Link to={article.href}
                            className="group flex items-center justify-between py-2 border-b border-white/[0.04] last:border-0 hover:border-white/10 transition-colors">
                            <div>
                              <div className="text-sm text-white/70 group-hover:text-white transition-colors">{article.title}</div>
                              <div className="text-xs text-white/30">{article.desc}</div>
                            </div>
                            <ArrowRight className="w-3.5 h-3.5 text-white/20 group-hover:text-cyan-400 flex-shrink-0 transition-colors" />
                          </Link>
                        </li>
                      ))}
                    </ul>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* infrays install */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md text-center">
            <h2 className="text-2xl font-black text-white mb-4">Install the infraYS CLI</h2>
            <p className="text-sm text-white/40 mb-8">Manage agents, alerts, SLOs, and more from your terminal.</p>
            <div className="terminal rounded-xl max-w-lg mx-auto">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
              </div>
              <div className="p-5 text-left">
                <p className="font-mono text-sm text-cyan-400">$ curl -fsSL https://get.infrays.org/ctl | sh</p>
                <p className="font-mono text-sm text-green-400 mt-2">✓ infrays installed to /usr/local/bin/infrays</p>
                <p className="font-mono text-sm text-white/40 mt-1">$ infrays --help</p>
              </div>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
