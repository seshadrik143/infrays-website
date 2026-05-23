import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'

// Privacy Policy boilerplate — NOT legal advice. Standard SaaS template
// with infraYS-specific data flows called out. Replace company contact
// + jurisdiction-specific lines once legal counsel reviews.
//
// Source-of-truth: this file. Effective-date string is hand-maintained.
const EFFECTIVE = '2026-05-23'

const sections = [
  {
    id: 'overview',
    title: '1. Overview',
    body: [
      'infraYS Solutions ("infraYS", "we", "us") provides an observability platform for infrastructure teams. This Privacy Policy explains what data we collect, why we collect it, and how you can control it.',
      'This policy applies to our marketing website (infrays.org) and the managed licensing service (license.infrays.org). The self-hosted NodePulse product runs entirely on your own infrastructure — we do not receive your metrics, logs, traces, or any operational data unless you explicitly enable our cloud services.',
    ],
  },
  {
    id: 'data-we-collect',
    title: '2. Data We Collect',
    body: [
      'When you visit infrays.org, we collect: pages visited, referring URL, user agent, IP address (truncated to /24 within 24 hours), and any forms you submit (contact form via Formspree → contact@infrays.org).',
      'When you sign up at license.infrays.org, we collect: email address, hashed password (bcrypt), name and company (optional), IP address of last login, and timestamps for account events.',
      'When you make a paid purchase, Stripe collects payment details on their hosted checkout page. We do not see card numbers or full billing addresses; we only receive a Stripe customer ID + subscription metadata.',
      'When your self-hosted NodePulse install phones home to refresh its license, we receive: license ID, deployment ID, NodePulse version, IP address of the request. We do not receive your metrics, logs, traces, alert rules, or any operational data.',
    ],
  },
  {
    id: 'why',
    title: '3. Why We Collect It',
    body: [
      'Account data — to authenticate you, send service emails (verification, password reset), and enforce subscription limits.',
      'Operational telemetry from the licensing service — to detect anomalous enrollment patterns, prevent abuse, and improve reliability.',
      'Web analytics — minimal, server-side only. We do not use Google Analytics or third-party trackers on infrays.org.',
      'Transactional email — Postmark sends our verification + reset + invoice emails. Postmark sees the recipient email + message body for the duration of delivery.',
    ],
  },
  {
    id: 'sharing',
    title: '4. Who We Share It With',
    body: [
      'Stripe — payment processing only (their Privacy Policy: https://stripe.com/privacy)',
      'Postmark — transactional email delivery only (https://postmarkapp.com/privacy-policy)',
      'Fly.io — our hosting provider (https://fly.io/legal/privacy-policy)',
      'We do not sell your data. We do not share it with advertising networks. We do not embed third-party trackers.',
      'We may disclose data if required by valid legal process; we will challenge overly broad requests where lawful and notify you unless prohibited.',
    ],
  },
  {
    id: 'retention',
    title: '5. Retention',
    body: [
      'Customer account records: retained for the lifetime of the subscription + 90 days after cancellation, then deleted (subject to legal retention requirements).',
      'License JWS records: retained indefinitely for audit purposes. Contains no personal data beyond your customer ID + email.',
      'Audit log entries: retained 12 months, then archived to cold storage for an additional 24 months, then deleted.',
      'Web server logs: retained 30 days for security monitoring; IP addresses are anonymized after 24 hours.',
    ],
  },
  {
    id: 'rights',
    title: '6. Your Rights (GDPR, CCPA, and others)',
    body: [
      'Right to access — request a copy of your data. Email contact@infrays.org with the subject "Data access request".',
      'Right to deletion — request erasure of your account. Note that some records (audit logs, financial records) are retained for legal reasons.',
      'Right to portability — receive your data in a machine-readable format (JSON).',
      'Right to correct — log into the customer portal to update your own profile, or email us.',
      'Right to opt out of sale — we do not sell personal information, so this right doesn\'t change anything practical, but we acknowledge it.',
      'We respond to verified rights requests within 30 days as required by GDPR (EU) and CCPA (California).',
    ],
  },
  {
    id: 'transfers',
    title: '7. International Data Transfers',
    body: [
      'We operate from Fly.io\'s us-east-1 (Ashburn, VA) region. If you are in the EU, EEA, or UK, your data is transferred to and processed in the United States under Standard Contractual Clauses (SCCs).',
      'We do not currently offer EU-region hosting. If your jurisdiction requires data residency, please contact us before signing up — for those cases, the self-hosted NodePulse option keeps all data on your infrastructure.',
    ],
  },
  {
    id: 'children',
    title: '8. Children',
    body: [
      'infraYS is a B2B SaaS product. We do not knowingly collect data from individuals under 16. If you believe a child has provided us data, contact us and we will delete it.',
    ],
  },
  {
    id: 'changes',
    title: '9. Changes to This Policy',
    body: [
      'We will update this page when we change how we handle data. Material changes will be announced via email to active customers + a banner at the top of license.infrays.org.',
      `Last updated: ${EFFECTIVE}`,
    ],
  },
  {
    id: 'contact',
    title: '10. Contact',
    body: [
      'For privacy questions, data requests, or to invoke any of the rights listed above:',
      'contact@infrays.org',
      'infraYS Solutions, c/o Postal Address (to be added), India.',
      'EU Representative: not currently appointed; pending appointment per Art. 27 GDPR.',
    ],
  },
]

export default function PrivacyPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <span className="badge-cyan mb-4">Legal</span>
            <h1 className="text-5xl font-black tracking-tight mb-4">Privacy Policy</h1>
            <p className="text-sm text-white/40">Effective {EFFECTIVE} · For data requests email <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a></p>
          </div>
        </section>

        <section className="section py-12">
          <div className="container-md grid lg:grid-cols-[220px_1fr] gap-12">
            {/* TOC */}
            <aside className="hidden lg:block sticky top-24 self-start text-sm space-y-1">
              {sections.map((s) => (
                <a key={s.id} href={`#${s.id}`}
                  className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md transition-colors">
                  {s.title}
                </a>
              ))}
            </aside>

            {/* Content */}
            <article className="prose-invert max-w-none">
              {sections.map((s) => (
                <section key={s.id} id={s.id} className="mb-10 scroll-mt-24">
                  <h2 className="text-xl font-bold text-white mb-4">{s.title}</h2>
                  {s.body.map((p, i) => (
                    <p key={i} className="text-white/60 leading-relaxed mb-3 text-sm">{p}</p>
                  ))}
                </section>
              ))}

              <div className="mt-12 border-t border-white/10 pt-6 text-xs text-white/30">
                See also: <Link to="/terms" className="text-cyan-400 hover:underline">Terms of Service</Link> · <Link to="/gdpr" className="text-cyan-400 hover:underline">GDPR / Data Requests</Link> · <Link to="/security" className="text-cyan-400 hover:underline">Security</Link>
              </div>
            </article>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
