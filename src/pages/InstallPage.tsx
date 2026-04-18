
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Terminal, Package, Server, Cloud, Copy, CheckCircle2, ArrowRight, Key, Clock } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useState } from 'react'

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  return (
    <button onClick={copy}
      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all border border-white/[0.08] hover:border-cyan-500/30 hover:text-cyan-400"
      style={{ background: 'rgba(255,255,255,0.04)', color: copied ? '#22d3ee' : 'rgba(255,255,255,0.35)' }}>
      {copied ? <CheckCircle2 className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
      {copied ? 'Copied' : 'Copy'}
    </button>
  )
}

const methods = [
  {
    id: 'curl',
    icon: Terminal,
    title: 'One-Line Install (Recommended)',
    badge: 'Fastest',
    badgeColor: 'badge-cyan',
    desc: 'Installs the NodePulse server, agent, and npctl CLI. A 15-day free trial starts automatically — no license key required. Works on any systemd-based Linux (amd64 / arm64).',
    steps: [
      {
        label: 'Download and install everything',
        code: 'curl -fsSL https://get.infrays.org/install | sudo bash',
        lang: 'bash',
      },
      {
        label: 'Services start automatically',
        code: `# Check status
sudo systemctl status nodepulse-server
sudo systemctl status nodepulse-agent

# View logs
journalctl -u nodepulse-server -f
journalctl -u nodepulse-agent  -f`,
        lang: 'bash',
      },
      {
        label: 'Open the dashboard',
        code: `# Dashboard:  http://<your-host>:8080
# Default login: admin / changeme  ← change on first login
# Config:  /etc/nodepulse/server.yaml
#          /etc/nodepulse/agent.yaml
# License: /etc/nodepulse/license.yaml`,
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
    desc: 'Run the NodePulse agent as a sidecar or standalone container. Mounts host /proc for system metrics.',
    steps: [
      {
        label: 'Pull and run',
        code: `docker run -d \\
  --name nodepulse-agent \\
  --pid=host \\
  --network=host \\
  -v /proc:/host/proc:ro \\
  -v /sys:/host/sys:ro \\
  -v /var/run/docker.sock:/var/run/docker.sock:ro \\
  -e NODEPULSE_SERVER_URL=http://your-server:8080 \\
  -e NODEPULSE_API_KEY=your-api-key \\
  ghcr.io/nodepulserepo/nodepulse-agent:latest`,
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
        code: `helm repo add nodepulse https://charts.infrays.org
helm repo update`,
        lang: 'bash',
      },
      {
        label: 'Install chart',
        code: `helm install nodepulse nodepulse/nodepulse \\
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
# nodepulse-agent-xxxxx   1/1   Running   0   30s`,
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
    desc: 'Spin up the complete NodePulse stack — server, agent, VictoriaMetrics, and dashboard — with a single command.',
    steps: [
      {
        label: 'Download and start',
        code: `curl -fsSL https://get.infrays.org/docker-compose.yml -o docker-compose.yml
curl -fsSL https://get.infrays.org/.env.example -o .env
# Edit .env with your settings
docker compose up -d`,
        lang: 'bash',
      },
      {
        label: 'Access the dashboard',
        code: `# Dashboard:   http://localhost:8080
# VictoriaMetrics: http://localhost:8428
# Default login: admin / changeme`,
        lang: 'bash',
      },
    ],
  },
]

const requirements = [
  { label: 'OS', value: 'Linux (amd64 / arm64) · systemd required' },
  { label: 'Memory', value: '< 30MB RAM (agent)  ~256MB (server)' },
  { label: 'CPU', value: '< 1% single core at 10s interval' },
  { label: 'Go', value: 'Not required — single static binary' },
  { label: 'Ports', value: 'Server: 8080  VictoriaMetrics: 8428' },
  { label: 'Kernel', value: '≥ 4.4 recommended (for eBPF collectors)' },
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
            <p className="text-lg text-white/40 max-w-xl mx-auto mb-8">
              One command installs the server, agent, and CLI.
              Always installs the latest release — no version pinning needed.
            </p>

            {/* Trial banner */}
            <div className="inline-flex items-center gap-6 border border-cyan-500/20 rounded-2xl px-6 py-4 text-sm"
              style={{ background: 'rgba(0,212,255,0.05)' }}>
              <div className="flex items-center gap-2 text-cyan-400 font-semibold">
                <Clock className="w-4 h-4" />
                15-day free trial
              </div>
              <span className="w-px h-4 bg-white/10" />
              <span className="text-white/50">No credit card required at install</span>
              <span className="w-px h-4 bg-white/10" />
              <div className="flex items-center gap-2 text-white/50">
                <Key className="w-4 h-4" />
                <span>Get a license key at</span>
                <Link to="/pricing" className="text-cyan-400 hover:underline font-medium">infrays.org/pricing</Link>
              </div>
            </div>
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
                          <span className="ml-auto flex items-center gap-2">
                            <span className="text-xs text-white/20 font-mono">{step.lang}</span>
                            <CopyButton text={step.code} />
                          </span>
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

        {/* License / Trial info */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(0,212,255,0.03)' }}>
          <div className="container-md">
            <div className="grid md:grid-cols-2 gap-8 items-start">
              <div>
                <div className="flex items-center gap-3 mb-4">
                  <Clock className="w-5 h-5 text-cyan-400" />
                  <h2 className="text-xl font-black text-white">15-Day Free Trial</h2>
                </div>
                <p className="text-sm text-white/50 leading-relaxed mb-4">
                  After installation, NodePulse automatically starts a 15-day free trial.
                  No license key, no credit card — just install and go.
                </p>
                <ul className="space-y-2 text-sm text-white/50">
                  {[
                    'All features unlocked during trial',
                    'Warning emails at 7, 3, and 1 day before expiry',
                    'Trial status visible in Dashboard → Settings → License',
                    'API: GET /api/v1/license',
                  ].map((item) => (
                    <li key={item} className="flex items-center gap-2">
                      <CheckCircle2 className="w-3.5 h-3.5 text-cyan-500/60 flex-shrink-0" />
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <div className="flex items-center gap-3 mb-4">
                  <Key className="w-5 h-5 text-purple-400" />
                  <h2 className="text-xl font-black text-white">Activating a License Key</h2>
                </div>
                <p className="text-sm text-white/50 leading-relaxed mb-4">
                  After your trial, activate a license key to keep NodePulse running.
                  Keys are issued instantly at checkout.
                </p>
                <div className="terminal rounded-xl mb-4">
                  <div className="terminal-header">
                    <div className="terminal-dot bg-[#ff5f57]" />
                    <div className="terminal-dot bg-[#ffbd2e]" />
                    <div className="terminal-dot bg-[#28ca41]" />
                    <span className="ml-auto text-xs text-white/20 font-mono">bash</span>
                  </div>
                  <pre className="p-4 text-sm font-mono text-cyan-300 overflow-x-auto leading-relaxed">
                    <code>{`# Via API
curl -X POST http://localhost:8080/api/v1/license \\
  -H "Authorization: Bearer <token>" \\
  -d '{"key":"NPLIC-..."}'

# Or: Dashboard → Settings → License → Activate Key`}</code>
                  </pre>
                </div>
                <Link to="/pricing"
                  className="inline-flex items-center gap-2 btn-primary text-sm px-5 py-2.5">
                  View Pricing & Get a Key
                  <ArrowRight className="w-4 h-4" />
                </Link>
              </div>
            </div>
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
                { title: 'Get a License', desc: 'Browse plans and activate a license key after your trial.', href: '/pricing', icon: '🔑' },
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
