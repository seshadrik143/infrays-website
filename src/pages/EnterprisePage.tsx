import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import CTABanner from '@/components/CTABanner'
import { Link } from 'react-router-dom'
import {
  Shield, Lock, Users, FileText, Server, Key, Globe,
  CheckCircle2, ArrowRight, Database, Cpu, Cloud
} from 'lucide-react'

const pillars = [
  {
    icon: Lock,
    color: 'indigo',
    title: 'Encryption & Secrets',
    items: [
      'AES-256-GCM encryption for all secrets at rest',
      'HashiCorp Vault integration for secret injection',
      'AWS Secrets Manager & GCP Secrets Manager support',
      'Zero plaintext credentials in agent configs',
    ],
  },
  {
    icon: Users,
    color: 'purple',
    title: 'Identity & Access',
    items: [
      'OIDC/SSO with Google, GitHub, Microsoft, custom IdP',
      'RBAC — 11 granular permissions, custom roles',
      'Session tracking with per-session revocation',
      'IP allowlist — CIDR-based access control middleware',
      'mTLS — mutual TLS for agent-to-server authentication',
    ],
  },
  {
    icon: FileText,
    color: 'cyan',
    title: 'Compliance',
    items: [
      'SOC 2 Type II control mapping (15 controls)',
      'ISO 27001 control evidence collection',
      'GDPR data subject erasure API',
      'Configurable data retention policies',
      'PDF/HTML compliance report export',
      'Audit log — every admin action recorded',
    ],
  },
  {
    icon: Server,
    color: 'green',
    title: 'High Availability',
    items: [
      'Raft consensus — leader election & log replication',
      'Zero-downtime server failover',
      'Multi-region deployment support',
      'Per-tenant resource quotas with live enforcement',
      'Backup & restore — streaming tar.gz via CLI',
    ],
  },
  {
    icon: Database,
    color: 'orange',
    title: 'Multi-Tenancy',
    items: [
      'Full data isolation per tenant',
      'Per-tenant agent registration and metrics',
      'Tenant-scoped API keys and dashboards',
      'Usage metering per tenant (agents/metrics/logs/spans)',
      'Admin tenant management API',
    ],
  },
  {
    icon: Globe,
    color: 'teal',
    title: 'Enterprise Integrations',
    items: [
      'Jira — automatic ticket creation on alert fire',
      'ServiceNow — incident creation via Table API',
      'PagerDuty — Events v2 with HMAC verification',
      'GitHub — deploy event annotations',
      'Terraform provider — full IaC for observability config',
      'Pulumi provider for all resource types',
    ],
  },
]

const complianceControls = [
  { framework: 'SOC 2', control: 'CC6.1', desc: 'Logical and physical access controls' },
  { framework: 'SOC 2', control: 'CC6.2', desc: 'Authentication mechanisms enforced' },
  { framework: 'SOC 2', control: 'CC7.2', desc: 'Security incidents are monitored' },
  { framework: 'SOC 2', control: 'CC8.1', desc: 'Change management processes' },
  { framework: 'ISO 27001', control: 'A.9.1', desc: 'Access control policy' },
  { framework: 'ISO 27001', control: 'A.10.1', desc: 'Cryptographic controls' },
  { framework: 'ISO 27001', control: 'A.12.1', desc: 'Operational procedures' },
  { framework: 'GDPR', control: 'Art. 17', desc: 'Right to erasure implemented' },
  { framework: 'GDPR', control: 'Art. 25', desc: 'Data protection by design' },
  { framework: 'GDPR', control: 'Art. 32', desc: 'Security of processing (AES-256)' },
]

const colorMap: Record<string, { bg: string; border: string; text: string }> = {
  indigo: { bg: 'bg-indigo-500/10', border: 'border-indigo-500/20', text: 'text-indigo-400' },
  purple: { bg: 'bg-purple-500/10', border: 'border-purple-500/20', text: 'text-purple-400' },
  cyan:   { bg: 'bg-cyan-500/10',   border: 'border-cyan-500/20',   text: 'text-cyan-400' },
  green:  { bg: 'bg-green-500/10',  border: 'border-green-500/20',  text: 'text-green-400' },
  orange: { bg: 'bg-orange-500/10', border: 'border-orange-500/20', text: 'text-orange-400' },
  teal:   { bg: 'bg-teal-500/10',   border: 'border-teal-500/20',   text: 'text-teal-400' },
}

export default function EnterprisePage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-indigo mb-4">Enterprise</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Built for teams that{' '}
              <span className="text-gradient-brand">can't afford downtime</span>
            </h1>
            <p className="text-lg text-white/40 max-w-2xl mx-auto mb-8">
              SOC 2, ISO 27001, and GDPR controls out of the box. OIDC SSO, granular RBAC,
              AES-256 encryption, Vault integration, and Raft HA — all self-hosted, all open source.
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <a href="mailto:hello@infrays.dev" className="btn-primary">
                Talk to Sales
                <ArrowRight className="w-4 h-4" />
              </a>
              <Link to="/install" className="btn-secondary">
                Deploy Now
              </Link>
            </div>
          </div>
        </section>

        {/* Security pillars */}
        <section className="section py-16">
          <div className="container-lg">
            <div className="text-center mb-12">
              <span className="badge-indigo mb-4">Security & Compliance</span>
              <h2 className="text-4xl font-black tracking-tight mb-4">
                Every control. Every framework.
              </h2>
              <p className="text-white/40 max-w-xl mx-auto">
                infraYS ships enterprise security controls that take months to build elsewhere —
                ready on day one.
              </p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
              {pillars.map((p) => {
                const c = colorMap[p.color]
                return (
                  <div key={p.title}
                    className="border border-white/[0.07] rounded-2xl p-6"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <div className={`w-11 h-11 rounded-xl ${c.bg} border ${c.border} flex items-center justify-center mb-5`}>
                      <p.icon className={`w-5 h-5 ${c.text}`} />
                    </div>
                    <h3 className="text-base font-bold text-white mb-4">{p.title}</h3>
                    <ul className="space-y-2.5">
                      {p.items.map((item) => (
                        <li key={item} className="flex items-start gap-2.5 text-sm text-white/55">
                          <CheckCircle2 className="w-4 h-4 text-green-400 flex-shrink-0 mt-0.5" />
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* Compliance controls table */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(10,10,20,0.5)' }}>
          <div className="container-lg">
            <div className="text-center mb-10">
              <h2 className="text-3xl font-black text-white mb-3">Built-in Compliance Controls</h2>
              <p className="text-white/40">Ready to export as PDF for your next audit</p>
            </div>

            <div className="border border-white/[0.07] rounded-2xl overflow-hidden"
              style={{ background: 'rgba(17,17,32,0.8)' }}>
              <div className="grid grid-cols-3 text-xs font-semibold text-white/30 uppercase tracking-wider px-6 py-3 border-b border-white/[0.06]">
                <div>Framework</div>
                <div>Control</div>
                <div>Description</div>
              </div>
              {complianceControls.map((ctrl, i) => (
                <div key={`${ctrl.framework}-${ctrl.control}`}
                  className={`grid grid-cols-3 px-6 py-4 text-sm ${i < complianceControls.length - 1 ? 'border-b border-white/[0.04]' : ''} hover:bg-white/[0.02] transition-colors`}>
                  <div>
                    <span className={`text-xs px-2 py-1 rounded-full font-medium ${
                      ctrl.framework === 'SOC 2' ? 'bg-indigo-500/15 text-indigo-400' :
                      ctrl.framework === 'ISO 27001' ? 'bg-purple-500/15 text-purple-400' :
                      'bg-green-500/15 text-green-400'
                    }`}>{ctrl.framework}</span>
                  </div>
                  <div className="font-mono text-white/60">{ctrl.control}</div>
                  <div className="text-white/50">{ctrl.desc}</div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Architecture callouts */}
        <section className="section py-16 border-t border-white/[0.06]">
          <div className="container-lg">
            <div className="grid md:grid-cols-3 gap-6">
              {[
                {
                  icon: Key,
                  color: 'text-yellow-400',
                  bg: 'bg-yellow-500/10 border-yellow-500/20',
                  title: 'Zero Secrets in Config',
                  desc: 'All credentials injected at runtime from Vault, AWS SM, or GCP SM. No plaintext in YAML or environment variables.',
                },
                {
                  icon: Cpu,
                  color: 'text-cyan-400',
                  bg: 'bg-cyan-500/10 border-cyan-500/20',
                  title: 'Raft Consensus HA',
                  desc: 'Leader election and log replication across nodes. Survive datacenter partitions. Automatic leader failover in < 5 seconds.',
                },
                {
                  icon: Cloud,
                  color: 'text-purple-400',
                  bg: 'bg-purple-500/10 border-purple-500/20',
                  title: 'On-Premises or Cloud',
                  desc: 'Deploy on your VPC, bare metal, or Kubernetes. No call-home, no telemetry. Your data stays in your environment.',
                },
              ].map((item) => (
                <div key={item.title}
                  className="border border-white/[0.07] rounded-2xl p-6 text-center"
                  style={{ background: 'rgba(17,17,32,0.7)' }}>
                  <div className={`w-12 h-12 rounded-2xl border ${item.bg} flex items-center justify-center mx-auto mb-5`}>
                    <item.icon className={`w-6 h-6 ${item.color}`} />
                  </div>
                  <h3 className="font-bold text-white mb-2">{item.title}</h3>
                  <p className="text-sm text-white/45 leading-relaxed">{item.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md text-center">
            <Shield className="w-10 h-10 text-indigo-400 mx-auto mb-5" />
            <h2 className="text-3xl font-black text-white mb-4">Ready for your security review?</h2>
            <p className="text-white/40 max-w-xl mx-auto mb-8">
              We'll provide architecture diagrams, security questionnaire responses, and a dedicated
              solution engineer to support your procurement process.
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <a href="mailto:hello@infrays.dev" className="btn-primary">
                Request Security Package
                <ArrowRight className="w-4 h-4" />
              </a>
              <Link to="/docs" className="btn-secondary">
                <FileText className="w-4 h-4" />
                Read Security Docs
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
