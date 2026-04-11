
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Terminal, Package, Server, Cloud, Copy, CheckCircle2, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

const methods = [
  {
    id: 'curl',
    icon: Terminal,
    title: 'Quick Install (Recommended)',
    badge: 'Fastest',
    badgeColor: 'badge-cyan',
    desc: 'One command installs and starts the agent. Works on Ubuntu, Debian, RHEL, CentOS, Fedora, and Alpine.',
    steps: [
      {
        label: 'Download and install',
        code: 'curl -fsSL https://get.infrays.dev | sh',
        lang: 'bash',
      },
      {
        label: 'Configure your server endpoint',
        code: `# Edit /etc/infrays/agent.yaml
server_url: "http://your-server:8080"
api_key: "your-api-key"
agent_name: "production-01"`,
        lang: 'yaml',
      },
      {
        label: 'Start the agent',
        code: `sudo systemctl enable infrays-agent
sudo systemctl start infrays-agent
sudo systemctl status infrays-agent`,
        lang: 'bash',
      },
    ],
  },
  {
    id: 'docker',
    icon: Package,
    title: 'Docker',
    badge: 'Container',
    badgeColor: 'badge-purple',
    desc: 'Run the agent as a sidecar or standalone container. Mounts host proc for system metrics.',
    steps: [
      {
        label: 'Pull and run',
        code: `docker run -d \\
  --name infrays-agent \\
  --pid=host \\
  --network=host \\
  -v /proc:/host/proc:ro \\
  -v /sys:/host/sys:ro \\
  -v /var/run/docker.sock:/var/run/docker.sock:ro \\
  -e INFRAYS_SERVER_URL=http://your-server:8080 \\
  -e INFRAYS_API_KEY=your-api-key \\
  infrays/agent:latest`,
        lang: 'bash',
      },
    ],
  },
  {
    id: 'kubernetes',
    icon: Server,
    title: 'Kubernetes / Helm',
    badge: 'K8s',
    badgeColor: 'badge-green',
    desc: 'Deploy as a DaemonSet across all nodes. Includes auto-discovery for pods and services.',
    steps: [
      {
        label: 'Add Helm repo',
        code: `helm repo add infrays https://charts.infrays.dev
helm repo update`,
        lang: 'bash',
      },
      {
        label: 'Install chart',
        code: `helm install infrays-agent infrays/agent \\
  --namespace monitoring \\
  --create-namespace \\
  --set server.url=http://your-server:8080 \\
  --set server.apiKey=your-api-key \\
  --set autodiscovery.enabled=true`,
        lang: 'bash',
      },
      {
        label: 'Verify pods are running',
        code: `kubectl get pods -n monitoring
# infrays-agent-xxxxx   1/1   Running   0   30s`,
        lang: 'bash',
      },
    ],
  },
  {
    id: 'compose',
    icon: Cloud,
    title: 'Docker Compose (Full Stack)',
    badge: 'All-in-one',
    badgeColor: 'badge-cyan',
    desc: 'Spin up the complete infraYS stack — server, agent, VictoriaMetrics, and dashboard — locally.',
    steps: [
      {
        label: 'Clone and start',
        code: `git clone https://github.com/seshadrik143/infrays-website.git
cd infrays-website
docker compose up -d`,
        lang: 'bash',
      },
      {
        label: 'Access the dashboard',
        code: `# Dashboard:   http://localhost:5176
# Server API:  http://localhost:8080
# Metrics DB:  http://localhost:8428
# Default login: admin / admin123`,
        lang: 'bash',
      },
    ],
  },
]

const requirements = [
  { label: 'OS', value: 'Linux (amd64 / arm64) · macOS · Windows WSL2' },
  { label: 'Memory', value: '< 30MB RAM (agent only)' },
  { label: 'CPU', value: '< 1% single core at 10s interval' },
  { label: 'Go', value: 'Not required — single static binary' },
  { label: 'Ports', value: 'Agent outbound: 8080 · Server: 8080, 8428 · Dashboard: 5176' },
  { label: 'Kernel', value: '≥ 4.4 recommended (for eBPF auto-discovery)' },
]

export default function InstallPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Installation Guide</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Deploy in{' '}
              <span className="text-gradient-cyan">60 seconds</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              Choose your installation method. The agent is a single static binary —
              no dependencies, no runtime, no surprises.
            </p>
          </div>
        </section>

        {/* System Requirements */}
        <section className="section py-12 border-b border-white/[0.06]"
          style={{ background: 'rgba(10, 10, 20, 0.5)' }}>
          <div className="container-md">
            <h2 className="text-sm font-semibold text-white/30 uppercase tracking-widest mb-6">System Requirements</h2>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
              {requirements.map((r) => (
                <div key={r.label} className="flex gap-3 border border-white/[0.06] rounded-xl px-4 py-3"
                  style={{ background: 'rgba(17,17,32,0.6)' }}>
                  <span className="text-xs font-semibold text-white/30 w-16 flex-shrink-0 pt-0.5">{r.label}</span>
                  <span className="text-xs text-white/60">{r.value}</span>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Install Methods */}
        <section className="section">
          <div className="container-md space-y-10">
            {methods.map((method) => (
              <div key={method.id} className="border border-white/[0.07] rounded-2xl overflow-hidden"
                style={{ background: 'rgba(17,17,32,0.7)' }}>
                {/* Header */}
                <div className="flex items-center gap-4 px-6 py-5 border-b border-white/[0.06]">
                  <div className="w-10 h-10 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center">
                    <method.icon className="w-5 h-5 text-white/60" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h2 className="text-base font-bold text-white">{method.title}</h2>
                      <span className={method.badgeColor}>{method.badge}</span>
                    </div>
                    <p className="text-sm text-white/40 mt-0.5">{method.desc}</p>
                  </div>
                </div>

                {/* Steps */}
                <div className="p-6 space-y-6">
                  {method.steps.map((step, si) => (
                    <div key={si}>
                      <div className="flex items-center gap-2.5 mb-3">
                        <div className="w-6 h-6 rounded-full bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center flex-shrink-0">
                          <span className="text-xs font-bold text-cyan-400">{si + 1}</span>
                        </div>
                        <span className="text-sm font-medium text-white/70">{step.label}</span>
                      </div>
                      <div className="terminal rounded-xl">
                        <div className="terminal-header">
                          <div className="terminal-dot bg-[#ff5f57]" />
                          <div className="terminal-dot bg-[#ffbd2e]" />
                          <div className="terminal-dot bg-[#28ca41]" />
                          <span className="ml-auto text-xs text-white/20 font-mono">{step.lang}</span>
                        </div>
                        <pre className="p-5 text-sm font-mono text-cyan-300 overflow-x-auto leading-relaxed">
                          <code>{step.code}</code>
                        </pre>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Next Steps */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md">
            <h2 className="text-2xl font-black text-white mb-8">What's next?</h2>
            <div className="grid md:grid-cols-3 gap-5">
              {[
                { title: 'Read the Docs', desc: 'Full configuration reference, collector guides, and API docs.', href: '/docs', icon: '📖' },
                { title: 'Set Up Alerts', desc: 'Configure alert rules, on-call schedules, and integrations.', href: '/docs#alerts', icon: '🔔' },
                { title: 'Join the Community', desc: 'Get help, share dashboards, and contribute plugins.', href: '/contact', icon: '💬' },
              ].map((item) => (
                <Link key={item.title} to={item.href}
                  className="group border border-white/[0.07] rounded-xl p-5 hover:border-cyan-500/20 transition-all"
                  style={{ background: 'rgba(17,17,32,0.6)' }}>
                  <div className="text-2xl mb-3">{item.icon}</div>
                  <h3 className="text-sm font-bold text-white mb-2 group-hover:text-cyan-400 transition-colors">
                    {item.title} <ArrowRight className="w-3.5 h-3.5 inline ml-1" />
                  </h3>
                  <p className="text-xs text-white/40 leading-relaxed">{item.desc}</p>
                </Link>
              ))}
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
