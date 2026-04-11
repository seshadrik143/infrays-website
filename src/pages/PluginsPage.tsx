import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CTABanner from '@/components/CTABanner'
import { Link } from 'react-router-dom'
import { useState } from 'react'
import { ArrowRight, Search, Package, Bell, LayoutDashboard, Code2 } from 'lucide-react'

type Plugin = {
  name: string
  desc: string
  category: string
  type: 'collector' | 'notifier' | 'dashboard_template'
  tags: string[]
}

const plugins: Plugin[] = [
  // ── Databases ───────────────────────────────────────────────
  { name: 'PostgreSQL', desc: 'Connections, locks, vacuum stats, table/index bloat, replication lag, query throughput', category: 'Database', type: 'collector', tags: ['postgres', 'sql', 'rdbms'] },
  { name: 'MySQL', desc: 'Queries/sec, slow queries, InnoDB buffer pool, connections, replication delay', category: 'Database', type: 'collector', tags: ['mysql', 'sql', 'rdbms'] },
  { name: 'MongoDB Collector', desc: 'Connections, opcounters, WiredTiger cache, network bytes, replication lag', category: 'Database', type: 'collector', tags: ['mongodb', 'nosql'] },
  { name: 'Redis / Valkey Collector', desc: 'Connections, memory, keyspace hit ratio, ops/s, evictions, replication lag', category: 'Database', type: 'collector', tags: ['redis', 'cache', 'valkey'] },
  { name: 'Cassandra Collector', desc: 'Read/write latency, pending tasks, compaction stats via JMX', category: 'Database', type: 'collector', tags: ['cassandra', 'nosql'] },
  { name: 'Elasticsearch Collector', desc: 'Cluster health, shard counts, index stats, JVM heap from REST API', category: 'Database', type: 'collector', tags: ['elasticsearch', 'search'] },
  { name: 'ClickHouse Collector', desc: 'Query throughput, merge stats, part counts from ClickHouse system tables', category: 'Database', type: 'collector', tags: ['clickhouse', 'analytics'] },
  { name: 'Oracle Database Collector', desc: 'Instance health, sessions, wait classes, redo log switches, tablespace usage', category: 'Database', type: 'collector', tags: ['oracle', 'sql', 'rdbms'] },
  { name: 'Microsoft SQL Server Collector', desc: 'Connections, batch requests, buffer cache, lock waits, deadlocks, per-DB file sizes', category: 'Database', type: 'collector', tags: ['mssql', 'sql', 'rdbms'] },
  { name: 'ScyllaDB Collector', desc: 'Read/write latency P99, compaction backlog, cache hit rate, node health via Prometheus', category: 'Database', type: 'collector', tags: ['scylladb', 'nosql'] },
  { name: 'CockroachDB Collector', desc: 'SQL queries/errors, Raft replication, storage live bytes, disk capacity via Prometheus', category: 'Database', type: 'collector', tags: ['cockroachdb', 'sql', 'distributed'] },
  { name: 'MinIO Collector', desc: 'Cluster health, disk usage, object count, S3 request rates via Prometheus endpoint', category: 'Database', type: 'collector', tags: ['minio', 'storage', 's3'] },

  // ── Message Queues ──────────────────────────────────────────
  { name: 'Kafka Collector', desc: 'Message/byte rates, under-replicated partitions, controller state, per-request latency via Jolokia', category: 'Message Queue', type: 'collector', tags: ['kafka', 'streaming'] },
  { name: 'Kafka Connect Collector', desc: 'Connector state (RUNNING/PAUSED/FAILED), per-connector task health via Connect REST API', category: 'Message Queue', type: 'collector', tags: ['kafka', 'kafka-connect'] },
  { name: 'RabbitMQ Collector', desc: 'Queue depth, message rates, consumer counts, node memory, channel stats', category: 'Message Queue', type: 'collector', tags: ['rabbitmq', 'amqp'] },
  { name: 'ActiveMQ Collector', desc: 'Broker memory/store usage, per-queue depth/consumer counts, enqueue/dequeue rates via Jolokia', category: 'Message Queue', type: 'collector', tags: ['activemq', 'jms'] },
  { name: 'NATS Collector', desc: 'Connections, message rates, JetStream streams/consumers/bytes via HTTP monitoring API', category: 'Message Queue', type: 'collector', tags: ['nats', 'messaging'] },

  // ── Web & Proxy ─────────────────────────────────────────────
  { name: 'NGINX Stub Status Collector', desc: 'Active connections, accepts, handled, requests/s, reading/writing/waiting workers', category: 'Web & Proxy', type: 'collector', tags: ['nginx', 'web', 'proxy'] },
  { name: 'NGINX VTS Collector', desc: 'Virtual Host Traffic Statistics — per-upstream/server-zone request and byte rates', category: 'Web & Proxy', type: 'collector', tags: ['nginx', 'vts'] },
  { name: 'HAProxy Collector', desc: 'Frontend/backend stats, queue depth, session rates from HAProxy stats socket', category: 'Web & Proxy', type: 'collector', tags: ['haproxy', 'lb'] },
  { name: 'Traefik Collector', desc: 'Request rates, entrypoint stats, service health from Traefik metrics endpoint', category: 'Web & Proxy', type: 'collector', tags: ['traefik', 'proxy'] },
  { name: 'Envoy Proxy Collector', desc: 'Per-cluster upstream health, request rates, latency; downstream connections via admin /stats', category: 'Web & Proxy', type: 'collector', tags: ['envoy', 'proxy', 'service-mesh'] },
  { name: 'Varnish Cache Collector', desc: 'Cache hit ratio, backend connection health, object counts, thread pool via varnishstat', category: 'Web & Proxy', type: 'collector', tags: ['varnish', 'cache'] },

  // ── Application Runtimes ────────────────────────────────────
  { name: 'JVM Collector', desc: 'Heap, GC pause, thread count via JMX — supports Tomcat, Spring Boot, and any JVM app', category: 'Runtime', type: 'collector', tags: ['jvm', 'java', 'spring'] },
  { name: 'Node.js Collector', desc: 'V8 heap, GC duration, event loop lag, HTTP request rates from prom-client /metrics', category: 'Runtime', type: 'collector', tags: ['nodejs', 'javascript'] },
  { name: 'PHP-FPM Collector', desc: 'Pool stats: idle/active processes, queue depth, accepted connections, slow requests', category: 'Runtime', type: 'collector', tags: ['php', 'php-fpm'] },
  { name: 'Apache Airflow Collector', desc: 'Scheduler/metadb health, DAG active/paused counts, dag_runs (24h), task instance states', category: 'Runtime', type: 'collector', tags: ['airflow', 'workflow', 'etl'] },
  { name: 'Celery / Flower Collector', desc: 'Worker online/offline, active task counts, task history success/failure via Flower REST API', category: 'Runtime', type: 'collector', tags: ['celery', 'python', 'tasks'] },
  { name: 'Apache Spark Collector', desc: 'Running apps, worker health, core/memory usage, per-app executor stats via Master REST API', category: 'Runtime', type: 'collector', tags: ['spark', 'bigdata'] },
  { name: 'Apache Solr Collector', desc: 'JVM heap, GC, core document counts, request handler throughput, cache hit ratios', category: 'Runtime', type: 'collector', tags: ['solr', 'search'] },
  { name: 'uWSGI Collector', desc: 'Worker states, request/exception counts, listen queue, avg response time via stats server', category: 'Runtime', type: 'collector', tags: ['uwsgi', 'python', 'wsgi'] },
  { name: 'Jenkins Collector', desc: 'Job health, build queue depth, executor availability via Jenkins JSON REST API', category: 'Runtime', type: 'collector', tags: ['jenkins', 'ci', 'cd'] },

  // ── Cloud & Infrastructure ──────────────────────────────────
  { name: 'AWS CloudWatch Collector', desc: 'EC2, RDS, ALB, Lambda metrics via CloudWatch GetMetricStatistics with SigV4 auth', category: 'Cloud', type: 'collector', tags: ['aws', 'cloudwatch', 'cloud'] },
  { name: 'Azure Monitor Collector', desc: 'VM, SQL Database, App Service metrics via Azure Monitor REST API with OAuth2', category: 'Cloud', type: 'collector', tags: ['azure', 'cloud'] },
  { name: 'GCP Cloud Monitoring Collector', desc: 'GCE, Cloud SQL, GKE metrics via Cloud Monitoring API with JWT service account auth', category: 'Cloud', type: 'collector', tags: ['gcp', 'google', 'cloud'] },
  { name: 'VMware vSphere Collector', desc: 'ESXi host connectivity, VM power states, datastore capacity via vCenter REST API', category: 'Cloud', type: 'collector', tags: ['vmware', 'vsphere', 'esxi', 'vm'] },
  { name: 'Proxmox VE Collector', desc: 'Cluster/node CPU, memory, disk, VM and LXC counts via PVE REST API', category: 'Cloud', type: 'collector', tags: ['proxmox', 'vm', 'hypervisor'] },

  // ── Service Discovery & Config ──────────────────────────────
  { name: 'Consul Collector', desc: 'Service health, catalog counts, Raft stats from Consul HTTP API', category: 'Service Discovery', type: 'collector', tags: ['consul', 'service-mesh'] },
  { name: 'etcd Collector', desc: 'Raft health, proposals, DB size, WAL fsync latency, peer network bytes', category: 'Service Discovery', type: 'collector', tags: ['etcd', 'k8s', 'distributed'] },
  { name: 'ZooKeeper Collector', desc: 'Connections, znodes, watchers, leader/follower state, latency via mntr command', category: 'Service Discovery', type: 'collector', tags: ['zookeeper', 'distributed'] },
  { name: 'HashiCorp Vault Collector', desc: 'Token TTL, lease counts, seal status, request rates from Vault API', category: 'Service Discovery', type: 'collector', tags: ['vault', 'secrets'] },
  { name: 'Vault Collector', desc: 'Seal status, token counts, lease expirations, Raft state via telemetry API', category: 'Service Discovery', type: 'collector', tags: ['vault', 'secrets', 'hashicorp'] },

  // ── Network & Security ──────────────────────────────────────
  { name: 'SNMP v2c Collector', desc: 'Poll any SNMP v2c device — switches, routers, UPS — with configurable OID list', category: 'Network', type: 'collector', tags: ['snmp', 'network', 'router'] },
  { name: 'WireGuard Collector', desc: 'Per-peer rx/tx bytes, last handshake time, active peer count via wg show dump', category: 'Network', type: 'collector', tags: ['wireguard', 'vpn', 'network'] },
  { name: 'OpenVPN Collector', desc: 'Connected clients, per-client bytes transferred, connection timestamps via management interface', category: 'Network', type: 'collector', tags: ['openvpn', 'vpn', 'network'] },
  { name: 'NFS Collector', desc: 'Client/server RPC call counts, NFSv3/v4 read-write ops from /proc/net/rpc/nfs[d]', category: 'Network', type: 'collector', tags: ['nfs', 'storage', 'network'] },
  { name: 'Ceph Collector', desc: 'Cluster health, OSD up/in counts, capacity usage, IOPS/throughput, per-pool stats via MGR API', category: 'Storage', type: 'collector', tags: ['ceph', 'storage', 'distributed'] },
  { name: 'Apache HTTP Server Collector', desc: 'mod_status: requests/sec, worker state, scoreboard, CPU load, connections', category: 'Web & Proxy', type: 'collector', tags: ['apache', 'httpd', 'web'] },

  // ── Notifiers ───────────────────────────────────────────────
  { name: 'Discord Notifier', desc: 'Send alert notifications to Discord channels via webhook with rich embed formatting', category: 'Notifications', type: 'notifier', tags: ['discord', 'chat'] },
  { name: 'Telegram Notifier', desc: 'Send alert notifications to Telegram chats or groups via Bot API', category: 'Notifications', type: 'notifier', tags: ['telegram', 'chat'] },
  { name: 'OpsGenie Notifier', desc: 'Create and close OpsGenie alerts with priority, tags, and responder assignment', category: 'Notifications', type: 'notifier', tags: ['opsgenie', 'alerting'] },
  { name: 'VictorOps / Splunk On-Call Notifier', desc: 'Send incident notifications to VictorOps (Splunk On-Call) with routing keys', category: 'Notifications', type: 'notifier', tags: ['victorops', 'splunk', 'oncall'] },
]

const allCategories = ['All', ...Array.from(new Set(plugins.map(p => p.category))).sort()]
const allTypes = ['All', 'collector', 'notifier', 'dashboard_template']

const typeIcon: Record<string, React.ElementType> = {
  collector: Package,
  notifier: Bell,
  dashboard_template: LayoutDashboard,
}

const typeLabel: Record<string, string> = {
  collector: 'Collector',
  notifier: 'Notifier',
  dashboard_template: 'Dashboard',
}

const typeBadge: Record<string, string> = {
  collector: 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20',
  notifier: 'bg-purple-500/10 text-purple-400 border border-purple-500/20',
  dashboard_template: 'bg-orange-500/10 text-orange-400 border border-orange-500/20',
}

const categoryColor: Record<string, string> = {
  Database: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  'Message Queue': 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  'Web & Proxy': 'bg-green-500/10 text-green-400 border-green-500/20',
  Runtime: 'bg-pink-500/10 text-pink-400 border-pink-500/20',
  Cloud: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  'Service Discovery': 'bg-teal-500/10 text-teal-400 border-teal-500/20',
  Network: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
  Storage: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  Notifications: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
}

export default function PluginsPage() {
  const [search, setSearch] = useState('')
  const [activeCategory, setActiveCategory] = useState('All')

  const filtered = plugins.filter(p => {
    const matchCat = activeCategory === 'All' || p.category === activeCategory
    const q = search.toLowerCase()
    const matchSearch = !q || p.name.toLowerCase().includes(q) || p.desc.toLowerCase().includes(q) || p.tags.some(t => t.includes(q)) || p.category.toLowerCase().includes(q)
    return matchCat && matchSearch
  })

  const collectorCount = plugins.filter(p => p.type === 'collector').length
  const notifierCount = plugins.filter(p => p.type === 'notifier').length

  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Plugin Catalog</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              {plugins.length} plugins.{' '}
              <span className="text-gradient-cyan">Ready to install.</span>
            </h1>
            <p className="text-lg text-white/40 max-w-2xl mx-auto">
              {collectorCount} collector plugins, {notifierCount} notification channels.
              One command to install. Write your own in any language using the Plugin SDK.
            </p>

            {/* Stats row */}
            <div className="flex flex-wrap justify-center gap-6 mt-8">
              {[
                { icon: Package, label: `${collectorCount} Collectors`, color: 'text-cyan-400' },
                { icon: Bell, label: `${notifierCount} Notifiers`, color: 'text-purple-400' },
                { icon: Code2, label: 'SDK for any language', color: 'text-green-400' },
              ].map(({ icon: Icon, label, color }) => (
                <div key={label} className="flex items-center gap-2 text-sm text-white/50">
                  <Icon className={`w-4 h-4 ${color}`} />
                  {label}
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Filter bar */}
        <section className="sticky top-16 z-40 border-b border-white/[0.06] py-4"
          style={{ background: 'rgba(6,6,16,0.95)', backdropFilter: 'blur(12px)' }}>
          <div className="container-lg flex flex-col sm:flex-row gap-4 items-start sm:items-center">
            {/* Search */}
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/30" />
              <input
                type="text"
                placeholder="Search plugins..."
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="w-full bg-white/[0.05] border border-white/10 rounded-xl pl-9 pr-4 py-2 text-sm text-white placeholder-white/30 focus:outline-none focus:border-cyan-500/40"
              />
            </div>

            {/* Category pills */}
            <div className="flex flex-wrap gap-2">
              {allCategories.map(cat => (
                <button
                  key={cat}
                  onClick={() => setActiveCategory(cat)}
                  className={`text-xs px-3 py-1.5 rounded-full border transition-all ${
                    activeCategory === cat
                      ? 'bg-cyan-500/20 border-cyan-500/40 text-cyan-300'
                      : 'bg-white/[0.04] border-white/10 text-white/50 hover:text-white/80 hover:border-white/20'
                  }`}>
                  {cat}
                </button>
              ))}
            </div>

            <span className="text-xs text-white/25 ml-auto flex-shrink-0">
              {filtered.length} of {plugins.length}
            </span>
          </div>
        </section>

        {/* Plugin grid */}
        <section className="section py-10">
          <div className="container-lg">
            {filtered.length === 0 ? (
              <div className="text-center py-20">
                <Package className="w-10 h-10 text-white/20 mx-auto mb-4" />
                <p className="text-white/40">No plugins match your search.</p>
              </div>
            ) : (
              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filtered.map(plugin => {
                  const TypeIcon = typeIcon[plugin.type] ?? Package
                  const catStyle = categoryColor[plugin.category] ?? 'bg-white/5 text-white/50 border-white/10'
                  return (
                    <div key={plugin.name}
                      className="border border-white/[0.07] rounded-2xl p-5 hover:border-white/15 transition-all group"
                      style={{ background: 'rgba(17,17,32,0.7)' }}>
                      {/* Header */}
                      <div className="flex items-start justify-between gap-3 mb-3">
                        <div className="flex items-center gap-3">
                          <div className="w-9 h-9 rounded-xl bg-white/[0.05] border border-white/10 flex items-center justify-center flex-shrink-0">
                            <TypeIcon className="w-4 h-4 text-white/40" />
                          </div>
                          <h3 className="text-sm font-bold text-white leading-snug group-hover:text-cyan-300 transition-colors">
                            {plugin.name}
                          </h3>
                        </div>
                      </div>

                      {/* Description */}
                      <p className="text-xs text-white/45 leading-relaxed mb-4">
                        {plugin.desc}
                      </p>

                      {/* Footer badges */}
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className={`text-xs px-2 py-0.5 rounded-full border ${catStyle}`}>
                          {plugin.category}
                        </span>
                        <span className={`text-xs px-2 py-0.5 rounded-full ${typeBadge[plugin.type]}`}>
                          {typeLabel[plugin.type]}
                        </span>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </section>

        {/* Build your own */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md text-center">
            <Code2 className="w-10 h-10 text-cyan-400 mx-auto mb-5" />
            <h2 className="text-3xl font-black text-white mb-4">Build your own plugin</h2>
            <p className="text-white/40 max-w-xl mx-auto mb-8">
              The infraYS Plugin SDK uses a simple exec protocol — write a collector script in Python,
              Ruby, Bash, Rust, or any language. Output metrics in simple key=value format and the
              agent takes care of the rest.
            </p>
            <div className="terminal rounded-xl max-w-lg mx-auto mb-8">
              <div className="terminal-header">
                <div className="terminal-dot bg-[#ff5f57]" />
                <div className="terminal-dot bg-[#ffbd2e]" />
                <div className="terminal-dot bg-[#28ca41]" />
                <span className="ml-auto text-xs text-white/20 font-mono">my_collector.sh</span>
              </div>
              <div className="p-5 text-left space-y-1">
                <p className="font-mono text-xs text-white/30">#!/bin/bash</p>
                <p className="font-mono text-xs text-cyan-400">my_service_queue_depth <span className="text-green-400">42</span></p>
                <p className="font-mono text-xs text-cyan-400">my_service_latency_ms <span className="text-green-400">12.4</span></p>
                <p className="font-mono text-xs text-cyan-400">my_service_errors_total <span className="text-green-400">3</span></p>
              </div>
            </div>
            <div className="flex flex-wrap justify-center gap-4">
              <Link to="/docs#sdk" className="btn-primary">
                Plugin SDK Docs
                <ArrowRight className="w-4 h-4" />
              </Link>
              <a href="https://github.com/seshadrik143/infrays-website"
                target="_blank" rel="noopener noreferrer"
                className="btn-secondary">
                Contribute a Plugin
              </a>
            </div>
          </div>
        </section>

        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
