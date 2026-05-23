
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Terminal, Package, Cloud, Copy, CheckCircle2, ArrowRight, Key } from 'lucide-react'
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
    desc: 'Installs the NodePulse server, agent, and npctl CLI. A valid license key is required to start the server — sign up at license.infrays.org first to get one (15-day free trial included, no credit card). Works on any systemd-based Linux (amd64 / arm64).',
    steps: [
      {
        label: 'Download and install everything',
        code: 'curl -fsSL https://infrays.org/install.sh | sudo bash',
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
  // Kubernetes / Helm card removed temporarily — the Helm chart at
  // charts.infrays.org isn't published yet. Showing a broken `helm
  // repo add` command mid-page misleads customers. Restore this entry
  // once we publish charts (likely to ghcr.io OCI artifacts).
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
        code: `curl -fsSL https://infrays.org/docker-compose.yml -o docker-compose.yml
curl -fsSL https://infrays.org/.env.example -o .env
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
              Get NodePulse{' '}
              <span className="text-gradient-cyan">running</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto mb-8">
              Single binary. No Kubernetes required. Works on any
              systemd-based Linux. Pick the path below that matches
              your setup.
            </p>

            {/* License-required banner */}
            <div className="inline-flex items-center gap-6 border border-cyan-500/20 rounded-2xl px-6 py-4 text-sm"
              style={{ background: 'rgba(0,212,255,0.05)' }}>
              <div className="flex items-center gap-2 text-cyan-400 font-semibold">
                <Key className="w-4 h-4" />
                License key required
              </div>
              <span className="w-px h-4 bg-white/10" />
              <span className="text-white/50">15 days free with signup — no credit card</span>
              <span className="w-px h-4 bg-white/10" />
              <div className="flex items-center gap-2 text-white/50">
                <span>Get yours at</span>
                <a href="https://license.infrays.org/signup" className="text-cyan-400 hover:underline font-medium">license.infrays.org/signup</a>
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

        {/* License flow info */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(0,212,255,0.03)' }}>
          <div className="container-md">
            <div className="grid md:grid-cols-2 gap-8 items-start">
              <div>
                <div className="flex items-center gap-3 mb-4">
                  <Key className="w-5 h-5 text-cyan-400" />
                  <h2 className="text-xl font-black text-white">How licensing works</h2>
                </div>
                <p className="text-sm text-white/50 leading-relaxed mb-4">
                  NodePulse is commercial software and requires a valid license key to run.
                  Sign up at the licensing portal — you get a 15-day free trial key
                  immediately, no credit card. Pick a paid plan any time before the trial
                  ends to keep running.
                </p>
                <ul className="space-y-2 text-sm text-white/50">
                  {[
                    'No license = the server refuses to start',
                    'Trial key issued immediately on signup (15 days, full features)',
                    'Keys are bound to your account and your tier',
                    'License + tier visible in Dashboard → Settings → License',
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
                  Paste your enrollment token into the server config, restart, and the
                  signed license JWS is fetched automatically. Keys are issued instantly
                  at checkout.
                </p>
                <div className="terminal rounded-xl mb-4">
                  <div className="terminal-header">
                    <div className="terminal-dot bg-[#ff5f57]" />
                    <div className="terminal-dot bg-[#ffbd2e]" />
                    <div className="terminal-dot bg-[#28ca41]" />
                    <span className="ml-auto text-xs text-white/20 font-mono">bash</span>
                  </div>
                  <pre className="p-4 text-sm font-mono text-cyan-300 overflow-x-auto leading-relaxed">
                    <code>{`# /etc/nodepulse/license.yaml
issuer: https://license.infrays.org
enrollment_token: NP-ENROLL-...

# Restart the server to fetch + apply
sudo systemctl restart nodepulse-server`}</code>
                  </pre>
                </div>
                <a href="https://license.infrays.org/signup"
                  className="inline-flex items-center gap-2 btn-primary text-sm px-5 py-2.5">
                  Sign Up & Get a Key
                  <ArrowRight className="w-4 h-4" />
                </a>
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
                { title: 'Get a License', desc: 'Sign up and self-serve at the licensing portal, or contact sales.', href: 'https://license.infrays.org/signup', icon: '🔑' },
              ].map((item) => {
                const className = "group border border-white/[0.07] rounded-xl p-5 hover:border-cyan-500/20 transition-all"
                const style = { background: 'rgba(17,17,32,0.6)' }
                const inner = (
                  <>
                    <div className="text-2xl mb-3">{item.icon}</div>
                    <h3 className="text-sm font-bold text-white mb-2 group-hover:text-cyan-400 transition-colors">
                      {item.title} <ArrowRight className="w-3.5 h-3.5 inline ml-1" />
                    </h3>
                    <p className="text-xs text-white/40 leading-relaxed">{item.desc}</p>
                  </>
                )
                return item.href.startsWith('http') ? (
                  <a key={item.title} href={item.href} className={className} style={style}>{inner}</a>
                ) : (
                  <Link key={item.title} to={item.href} className={className} style={style}>{inner}</Link>
                )
              })}
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
