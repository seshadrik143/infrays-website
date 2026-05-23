import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { Shield, Lock, Key, FileCheck, Server, Globe, AlertTriangle, Mail } from 'lucide-react'

// Security posture page. Honest about what's in place today + what's
// in flight. For a monitoring product asking customers to install an
// agent on their servers, transparency about threat model + compliance
// status is essential trust signal.

const pillars = [
  {
    icon: Lock,
    title: 'Encryption',
    items: [
      'TLS 1.3 for all customer-facing endpoints (license.infrays.org, infrays.org)',
      'AES-256-GCM for secrets at rest in the NodePulse self-hosted product',
      'HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager integrations for secret injection',
      'Zero plaintext credentials in agent configs — secrets resolved at startup',
    ],
  },
  {
    icon: Key,
    title: 'Authentication',
    items: [
      'Admin portal: bcrypt cost 12 + mandatory TOTP MFA (Google Authenticator / 1Password / Authy compatible)',
      'Customer portal: bcrypt cost 12 + email verification on signup',
      'Per-account brute-force lockout (5 failed attempts → 15-minute cooldown)',
      'Per-IP rate limiting on auth endpoints',
      'Session cookies: HttpOnly, SameSite=Lax, Secure (TLS only), 7-day TTL',
    ],
  },
  {
    icon: Shield,
    title: 'Infrastructure',
    items: [
      'Hosted on Fly.io with TLS termination at the edge',
      'Web Application Firewall (Vercel + Fly built-in)',
      'HSTS + CSP + X-Frame-Options + nosniff defense-in-depth response headers',
      'Hashed asset bundles with immutable cache + content-hash filename versioning',
      'Auto-scaled to zero between requests; no idle attack surface',
    ],
  },
  {
    icon: FileCheck,
    title: 'Audit & Compliance',
    items: [
      'Tamper-evident audit log: every admin action + system event hash-chained (SHA-256)',
      'Audit-verify endpoint detects payload tampering, mid-chain deletions, sequence gaps',
      'Per-tenant data isolation enforced at every store call (verified by audit)',
      'GDPR data-request endpoint (see /gdpr)',
      'SOC 2 Type 1 — in flight (target: H2 2026)',
      'ISO 27001 — evaluation phase',
    ],
  },
  {
    icon: Server,
    title: 'Self-hosted advantage',
    items: [
      'Run NodePulse entirely on your own infrastructure — same license tier, no extra fee',
      'Your metrics, logs, traces never leave your network',
      'License refresh phones home daily but transmits only license ID + deployment ID + version',
      'Optional offline / air-gap mode (license valid until exp without refresh)',
      'No telemetry, no usage analytics, no third-party trackers',
    ],
  },
  {
    icon: Globe,
    title: 'Cloud (managed) data flows',
    items: [
      'Currently US-East-1 (Ashburn, VA) only',
      'Customer data: account email + bcrypt password hash + license records (no operational data)',
      'Sub-processors: Stripe (billing), Postmark (email), Fly.io (hosting) — see Privacy Policy §4',
      'For EU data residency: use the self-hosted product. Managed EU region is on the roadmap.',
    ],
  },
]

export default function SecurityPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Security</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Built for teams that can&apos;t afford{' '}
              <span className="text-gradient-cyan">surprises</span>
            </h1>
            <p className="text-lg text-white/40 max-w-2xl mx-auto">
              infraYS is an observability platform — so the bar for our own security posture is high.
              Here&apos;s what we do, what we&apos;re working on, and how to reach us.
            </p>
          </div>
        </section>

        {/* Pillars */}
        <section className="section py-16">
          <div className="container-lg">
            <div className="grid md:grid-cols-2 gap-5">
              {pillars.map((p) => {
                const Icon = p.icon
                return (
                  <div key={p.title} className="border border-white/[0.07] rounded-2xl p-6"
                    style={{ background: 'rgba(17,17,32,0.7)' }}>
                    <div className="w-10 h-10 rounded-xl flex items-center justify-center mb-4 bg-cyan-500/10 border border-cyan-500/20">
                      <Icon className="w-5 h-5 text-cyan-400" />
                    </div>
                    <h3 className="text-base font-bold text-white mb-4">{p.title}</h3>
                    <ul className="space-y-2">
                      {p.items.map((item) => (
                        <li key={item} className="text-sm text-white/60 leading-relaxed flex gap-2">
                          <span className="text-cyan-400 flex-shrink-0 mt-0.5">·</span>
                          <span>{item}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* Threat model */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md">
            <h2 className="text-3xl font-black text-white mb-6">Our threat model</h2>
            <div className="space-y-4 text-sm text-white/60 leading-relaxed">
              <p>
                <strong className="text-white">What we defend against:</strong>{' '}
                credential stuffing (per-account lockout), session hijacking
                (HttpOnly + SameSite cookies), MFA bypass attempts (TOTP enforced on
                every admin login), CSRF (SameSite=Lax + JSON-only API), clickjacking
                (X-Frame-Options: DENY), MIME sniffing attacks (X-Content-Type-Options:
                nosniff), tampering with the audit log (hash-chained entries).
              </p>
              <p>
                <strong className="text-white">What we don&apos;t defend against:</strong>{' '}
                attacks against the customer&apos;s own infrastructure where NodePulse
                is installed. That&apos;s the customer&apos;s perimeter to defend.
                We do publish hardening guidance for self-hosted deployments in our
                setup guide (deployment §11 Production Security Checklist).
              </p>
              <p>
                <strong className="text-white">What we&apos;re explicit about NOT having yet:</strong>{' '}
                SOC 2 Type 2 attestation (Type 1 in flight), ISO 27001 certificate,
                a public bug-bounty program (responsible disclosure via security@ works
                — see below), an EU-region managed deployment, automated penetration
                testing on the managed service.
              </p>
            </div>
          </div>
        </section>

        {/* Responsible disclosure */}
        <section className="section py-16 border-t border-white/[0.06]">
          <div className="container-md">
            <div className="border border-amber-500/20 rounded-2xl p-8"
              style={{ background: 'rgba(245,158,11,0.05)' }}>
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 bg-amber-500/10 border border-amber-500/20">
                  <AlertTriangle className="w-5 h-5 text-amber-400" />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-bold text-white mb-3">Found a vulnerability?</h3>
                  <p className="text-sm text-white/60 leading-relaxed mb-4">
                    Email <a className="text-amber-400 font-mono" href="mailto:contact@infrays.org">contact@infrays.org</a>{' '}
                    with subject line <code className="text-amber-400 bg-black/30 px-1.5 py-0.5 rounded">[SECURITY]</code>.
                    We aim to acknowledge within 48 hours and to ship a fix within 30 days for
                    critical issues. We&apos;ll credit you in the changelog unless you ask us not to.
                  </p>
                  <p className="text-sm text-white/60 leading-relaxed mb-2">
                    Please <strong className="text-white">don&apos;t</strong> publicly disclose until we&apos;ve had a chance
                    to fix. Please <strong className="text-white">don&apos;t</strong> run automated scans against
                    license.infrays.org without notifying us first — we&apos;ll allowlist
                    your IP for the test window.
                  </p>
                  <p className="text-xs text-white/40 mt-4">
                    A machine-readable disclosure policy is published at{' '}
                    <a className="text-amber-400" href="/.well-known/security.txt">/.well-known/security.txt</a>.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Contact CTAs */}
        <section className="section py-16 border-t border-white/[0.06]"
          style={{ background: 'rgba(8,8,16,0.5)' }}>
          <div className="container-md grid md:grid-cols-2 gap-5">
            <a href="mailto:contact@infrays.org?subject=Security%20question"
              className="border border-white/[0.07] rounded-2xl p-6 hover:border-cyan-500/30 transition-colors group">
              <Mail className="w-5 h-5 text-cyan-400 mb-3" />
              <h3 className="text-base font-bold text-white mb-2 group-hover:text-cyan-400 transition-colors">Security questions</h3>
              <p className="text-sm text-white/60">For pre-sales security review, DPA requests, or compliance questionnaires.</p>
            </a>

            <Link to="/contact"
              className="border border-white/[0.07] rounded-2xl p-6 hover:border-cyan-500/30 transition-colors group">
              <Shield className="w-5 h-5 text-cyan-400 mb-3" />
              <h3 className="text-base font-bold text-white mb-2 group-hover:text-cyan-400 transition-colors">Enterprise security review</h3>
              <p className="text-sm text-white/60">For SOC 2 reports, ISO 27001 status, custom security agreements.</p>
            </Link>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
