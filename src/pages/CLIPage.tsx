import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CTABanner from '@/components/CTABanner'
import { Terminal, Copy, CheckCircle2, ArrowRight } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

const commands = [
  {
    cmd: 'infrays agents list',
    desc: 'List all registered agents with status, version, OS, CPU, memory, and last-seen time.',
    example: `AGENT ID     HOSTNAME      STATUS  VERSION  OS            CPU  MEMORY    METRICS  LAST SEEN
prod-web-01  web-server-1  online  1.0.0    ubuntu 24.04  8    31.1 GiB  48293    3s ago
prod-db-01   db-server-1   online  1.0.0    ubuntu 22.04  16   63.9 GiB  12847    5s ago`,
    color: 'cyan',
  },
  {
    cmd: 'infrays agents get <id>',
    desc: 'Get full details for a single agent including tenant, version, CPU, memory, and registration times.',
    example: `Agent ID:      prod-web-01
Tenant ID:     default
Hostname:      web-server-1
Status:        online
Version:       1.0.0
OS:            ubuntu 24.04
CPU Cores:     8
Memory:        31.1 GiB
Metrics Recv:  48293
First Seen:    2026-04-01 10:00:00
Last Seen:     2026-04-11 08:15:03`,
    color: 'cyan',
  },
  {
    cmd: 'infrays alerts list [--all]',
    desc: 'List active alerts. Pass --all to include resolved alerts.',
    example: `ID          AGENT        SEVERITY  STATE    RULE               FIRED
alert-001   prod-web-01  critical  firing   High CPU Usage     2m ago
alert-002   prod-db-01   warning   firing   Disk Usage > 80%   15m ago`,
    color: 'red',
  },
  {
    cmd: 'infrays slo list [--window 7] [--target 99] [--failing]',
    desc: 'List per-agent SLO status for a given window (days) and target percentage.',
    example: `Target: 99.0%  |  Window: 7 days

AGENT ID     HOSTNAME      STATUS  UPTIME    DOWNTIME  INCIDENTS  SLO
prod-web-01  web-server-1  online  99.987%   1m 48s    1          PASS
prod-db-01   db-server-1   online  100.000%  —         0          PASS

2 passing, 0 failing (total: 2)`,
    color: 'green',
  },
  {
    cmd: 'infrays slo summary [--window 7] [--target 99]',
    desc: 'Fleet-wide SLO summary with averages, downtime totals, and pass/fail counts.',
    example: `SLO Summary — 7-day window, 99.0% target
─────────────────────────────────────────
Total agents:    2
Meeting SLO:     2
Breaching SLO:   0
Avg uptime:      99.994%
Total downtime:  1m 48s
Total incidents: 1`,
    color: 'green',
  },
  {
    cmd: 'infrays health',
    desc: 'Check server health and VictoriaMetrics connectivity.',
    example: `Server:          ok (http://localhost:8080)
Version:         1.0.0
VictoriaMetrics: ok`,
    color: 'green',
  },
  {
    cmd: 'infrays config get <agent-id>',
    desc: 'Fetch the current YAML config for an agent.',
    example: `agent_id: prod-web-01
server_url: http://infrays-server:8080
collectors:
  system: { enabled: true, interval: 10 }
  docker: { enabled: true }
  logs:
    sources:
      - path: /var/log/nginx/access.log`,
    color: 'purple',
  },
  {
    cmd: 'infrays config push <agent-id> <file>',
    desc: 'Push a new YAML config file to an agent. Agent reloads on next check-in.',
    example: `Config pushed to prod-web-01 ✓`,
    color: 'purple',
  },
  {
    cmd: 'infrays logs tail [--agent <id>] [--level <lvl>] [--n <count>] [--q <text>]',
    desc: 'Tail recent log entries. Filter by agent, level (info/warn/error), count, or full-text search.',
    example: `TIMESTAMP            AGENT        LEVEL  SOURCE               MESSAGE
2026-04-11 08:14:44  prod-web-01  error  /var/log/nginx/err   upstream connect error
2026-04-11 08:14:22  prod-web-01  warn   /var/log/app.log     high latency detected`,
    color: 'yellow',
  },
  {
    cmd: 'infrays oncall current',
    desc: 'Show who is currently on-call across all configured schedules.',
    example: `Schedule:  Platform Team
On-Call:   Jane Smith
Email:     jane@company.com
Phone:     +1-555-0100
Since:     2026-04-10 09:00:00
Until:     2026-04-14 09:00:00`,
    color: 'orange',
  },
  {
    cmd: 'infrays oncall list',
    desc: 'List all configured on-call schedules with member count and description.',
    example: `ID              NAME             MEMBERS  DESCRIPTION
platform-team   Platform Team    4        24/7 infrastructure on-call
sre-rotation    SRE Rotation     3        Business hours SRE team`,
    color: 'orange',
  },
  {
    cmd: 'infrays annotations list',
    desc: 'List all annotations (deployment markers) with time, title, tags, and creator.',
    example: `ID                   TIME                 TITLE             TAGS           CREATED BY
1744329600000000000  2026-04-11 08:00:00  Deploy v2.1.0     deploy,prod    ci-bot
1744243200000000000  2026-04-10 08:00:00  DB migration      db,migration   admin`,
    color: 'teal',
  },
  {
    cmd: 'infrays annotations create --title <t> [--desc <d>] [--tags <a,b>]',
    desc: 'Create a deployment annotation visible as a marker on all metric charts.',
    example: `Annotation created: ID=1744329600000000000`,
    color: 'teal',
  },
  {
    cmd: 'infrays groups list',
    desc: 'List all agent groups with label selector, agent count, and description.',
    example: `ID            NAME             AGENTS  SELECTOR           DESCRIPTION
prod-web      Production Web   12      env=prod,role=web  Production web tier
staging       Staging Fleet    4       env=staging        All staging agents`,
    color: 'blue',
  },
  {
    cmd: 'infrays groups command --id <id> --action <act> [--params key=val,...]',
    desc: 'Send a bulk command (restart, update, config-reload) to all agents matching a group.',
    example: `Command dispatched to 12 agents in group prod-web ✓`,
    color: 'blue',
  },
  {
    cmd: 'infrays admin backup [--out <file>]',
    desc: 'Download a complete server backup as a tar.gz archive. Includes all BoltDB stores.',
    example: `Backup saved to nodepulse-backup-20260411-081503.tar.gz (26450 bytes)`,
    color: 'rose',
  },
  {
    cmd: 'infrays admin restore --file <file>',
    desc: 'Restore server data from a backup archive. Restart the server to reopen DB files.',
    example: `Restored: agents.db, alerts.db, annotations.db, groups.db
Total: 4 files restored. Restart the server to apply.`,
    color: 'rose',
  },
  {
    cmd: 'infrays report latest',
    desc: 'Show the current monitoring report — agent counts, alerts, SLO status, and dispatch note.',
    example: `NodePulse Report — 7-day summary
Generated: 2026-04-11 08:15 UTC
Window:    2026-04-04 → 2026-04-11
────────────────────────────────
Agents : 2 total  |  2 online  |  0 offline
Alerts : 2 firing  |  14 resolved
SLO    : target 99.0%  |  2 passing, 0 failing`,
    color: 'purple',
  },
  {
    cmd: 'infrays report trigger',
    desc: 'Dispatch the current report to all configured notification channels immediately.',
    example: `Report triggered — dispatched to 2 channel(s) ✓`,
    color: 'purple',
  },
  {
    cmd: 'infrays cluster',
    desc: 'Show Raft cluster and HA status — leader, node state, term, and log index.',
    example: `Mode:   raft
State:  leader
Leader: infrays-server-01:8081
Term:   42
Index:  18341`,
    color: 'indigo',
  },
  {
    cmd: 'infrays completion [bash|zsh|fish]',
    desc: 'Generate shell completion scripts for bash, zsh, or fish.',
    example: `# Add to ~/.bashrc:
source <(infrays completion bash)`,
    color: 'white',
  },
  {
    cmd: 'infrays version',
    desc: 'Print the CLI version.',
    example: `npctl v0.33.0`,
    color: 'white',
  },
]

const colorBadge: Record<string, string> = {
  cyan:   'text-cyan-400 bg-cyan-500/10',
  red:    'text-red-400 bg-red-500/10',
  green:  'text-green-400 bg-green-500/10',
  purple: 'text-purple-400 bg-purple-500/10',
  yellow: 'text-yellow-400 bg-yellow-500/10',
  orange: 'text-orange-400 bg-orange-500/10',
  teal:   'text-teal-400 bg-teal-500/10',
  blue:   'text-blue-400 bg-blue-500/10',
  rose:   'text-rose-400 bg-rose-500/10',
  indigo: 'text-indigo-400 bg-indigo-500/10',
  white:  'text-white/50 bg-white/5',
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <button
      className="text-white/30 hover:text-white/70 transition-colors p-1"
      onClick={() => {
        navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      }}>
      {copied ? <CheckCircle2 className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
    </button>
  )
}

export default function CLIPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">CLI Reference</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              <span className="text-gradient-cyan">infrays</span> — the CLI
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto mb-8">
              Manage agents, alerts, SLOs, logs, groups, backups, and on-call rotations
              from your terminal. 20+ commands, API key auth, shell completion included.
            </p>

            {/* Install */}
            <div className="terminal rounded-xl max-w-lg mx-auto mb-6">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
                <span className="ml-auto text-xs text-white/20 font-mono">install</span>
              </div>
              <div className="p-5 text-left flex items-center justify-between">
                <p className="font-mono text-sm text-cyan-400">curl -fsSL https://get.infrays.org/install | sudo bash</p>
                <CopyButton text="curl -fsSL https://get.infrays.org/install | sudo bash" />
              </div>
            </div>

            {/* Flags */}
            <div className="terminal rounded-xl max-w-lg mx-auto">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
                <span className="ml-auto text-xs text-white/20 font-mono">global flags</span>
              </div>
              <div className="p-5 text-left space-y-2">
                <p className="font-mono text-sm">
                  <span className="text-yellow-400">-server</span>
                  <span className="text-white/40"> string  </span>
                  <span className="text-white/50">Server URL (env: INFRAYS_SERVER)</span>
                </p>
                <p className="font-mono text-sm">
                  <span className="text-yellow-400">-key   </span>
                  <span className="text-white/40"> string  </span>
                  <span className="text-white/50">API key  (env: INFRAYS_KEY)</span>
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Commands */}
        <section className="section py-16">
          <div className="container-lg">
            <div className="space-y-5">
              {commands.map((c) => {
                const badge = colorBadge[c.color] ?? colorBadge['white']
                return (
                  <div key={c.cmd}
                    className="border border-white/[0.07] rounded-2xl overflow-hidden"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    {/* Command header */}
                    <div className="flex items-start justify-between gap-4 px-6 py-4 border-b border-white/[0.05]">
                      <div className="flex items-start gap-4">
                        <Terminal className="w-4 h-4 text-white/20 flex-shrink-0 mt-1" />
                        <div>
                          <code className={`text-sm font-mono font-semibold px-2 py-0.5 rounded ${badge}`}>
                            {c.cmd}
                          </code>
                          <p className="text-sm text-white/45 mt-2">{c.desc}</p>
                        </div>
                      </div>
                      <CopyButton text={c.cmd.split(' ')[0] + ' ' + c.cmd.split(' ').slice(1).join(' ')} />
                    </div>
                    {/* Example output */}
                    <div className="p-5">
                      <p className="text-xs text-white/20 uppercase tracking-widest mb-3">Example output</p>
                      <pre className="font-mono text-xs text-white/55 leading-relaxed whitespace-pre-wrap overflow-x-auto">
                        {c.example}
                      </pre>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* Shell completion */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md">
            <h2 className="text-2xl font-black text-white mb-6 text-center">Shell Completion</h2>
            <div className="grid md:grid-cols-3 gap-4">
              {['bash', 'zsh', 'fish'].map((shell) => (
                <div key={shell}
                  className="terminal rounded-xl">
                  <div className="terminal-header">
                    <div className="terminal-dot bg-[#ff5f57]" />
                    <div className="terminal-dot bg-[#ffbd2e]" />
                    <div className="terminal-dot bg-[#28ca41]" />
                    <span className="ml-auto text-xs text-white/20 font-mono">{shell}</span>
                  </div>
                  <div className="p-4 text-left space-y-1">
                    <p className="font-mono text-xs text-cyan-400">
                      infrays completion {shell}
                    </p>
                    {shell === 'bash' && <p className="font-mono text-xs text-white/30"># → source &lt;(infrays completion bash)</p>}
                    {shell === 'zsh' && <p className="font-mono text-xs text-white/30"># → infrays completion zsh &gt; ~/.zfunc/_infrays</p>}
                    {shell === 'fish' && <p className="font-mono text-xs text-white/30"># → infrays completion fish | source</p>}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
