import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'

// Terms of Service boilerplate — NOT legal advice. Standard SaaS template
// adapted for infraYS's open-source + paid-license + cloud model.
// Replace company contact + jurisdiction-specific lines once legal
// counsel reviews.
const EFFECTIVE = '2026-05-23'

const sections = [
  {
    id: 'acceptance',
    title: '1. Acceptance',
    body: [
      'By creating an account at license.infrays.org, accessing infrays.org, or running the NodePulse software ("the Software"), you agree to these Terms. If you don\'t agree, don\'t use the service.',
      'If you are using infraYS on behalf of a company or other entity, "you" refers to that entity and you confirm you have authority to bind it.',
    ],
  },
  {
    id: 'license',
    title: '2. Software License',
    body: [
      'The NodePulse open-source core is licensed under the Apache License 2.0. You may use, modify, and redistribute it under those terms.',
      'Enterprise modules in the Software are licensed under the Functional Source License v1.1 with a 2-year MIT change date. See the LICENSING.md file in the NodePulse source repository for the canonical terms.',
      'Your subscription with infraYS grants you a non-exclusive, non-transferable, revocable right to use the Enterprise features for the duration of the subscription period.',
    ],
  },
  {
    id: 'acceptable-use',
    title: '3. Acceptable Use',
    body: [
      'You agree NOT to use infraYS to: violate any law or regulation; infringe any third party\'s rights; transmit malware or illegal content; reverse-engineer our managed services in an attempt to recreate them; perform load tests or penetration tests against our infrastructure without written permission.',
      'You are responsible for ensuring your use complies with all laws applicable to you and your end users.',
    ],
  },
  {
    id: 'accounts',
    title: '4. Accounts & Security',
    body: [
      'You are responsible for safeguarding your account credentials and enrollment tokens. Share enrollment tokens via secure channels only — they grant license access to whoever holds them.',
      'You must notify us immediately at contact@infrays.org if you suspect unauthorized access.',
      'We may suspend or terminate accounts that violate these Terms or that show signs of compromise pending investigation.',
    ],
  },
  {
    id: 'payment',
    title: '5. Payment & Subscriptions',
    body: [
      'Paid subscriptions are billed in advance via Stripe. Prices listed on infrays.org/pricing are in USD unless otherwise noted.',
      'Subscriptions auto-renew at the end of each billing period unless canceled. You can cancel at any time from the customer portal; cancellation takes effect at the end of the current period.',
      'Refunds: we offer a pro-rated refund within 14 days of initial purchase. After 14 days, all sales are final. Contact contact@infrays.org for refund requests.',
      'We may change pricing with 30 days notice via email. Price changes do not apply to subscription periods already paid for.',
    ],
  },
  {
    id: 'data',
    title: '6. Your Data',
    body: [
      'You retain all rights to data you store or process with the Software. We do not claim ownership of your metrics, logs, traces, or any operational data.',
      'For self-hosted NodePulse, your data stays on your infrastructure. We have no access.',
      'For the managed licensing service (license.infrays.org), see our Privacy Policy for what we collect and why.',
    ],
  },
  {
    id: 'warranty',
    title: '7. Warranties & Disclaimer',
    body: [
      'THE SOFTWARE AND SERVICES ARE PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING WITHOUT LIMITATION WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, OR NON-INFRINGEMENT.',
      'We do not warrant that the service will be uninterrupted, error-free, or that defects will be corrected. We do not warrant any specific results from use.',
    ],
  },
  {
    id: 'liability',
    title: '8. Limitation of Liability',
    body: [
      'TO THE MAXIMUM EXTENT PERMITTED BY LAW, INFRAYS\'S TOTAL LIABILITY ARISING OUT OF OR RELATED TO THESE TERMS WILL NOT EXCEED THE GREATER OF (A) THE AMOUNT YOU PAID INFRAYS IN THE 12 MONTHS PRECEDING THE CLAIM OR (B) USD 100.',
      'NEITHER PARTY WILL BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES.',
      'These limitations apply even if a remedy fails of its essential purpose.',
    ],
  },
  {
    id: 'indemnification',
    title: '9. Indemnification',
    body: [
      'You agree to defend, indemnify, and hold infraYS harmless from any claim arising out of your use of the Software in violation of these Terms or applicable law.',
    ],
  },
  {
    id: 'termination',
    title: '10. Termination',
    body: [
      'You may terminate your account at any time via the customer portal. We may terminate or suspend your account for material breach of these Terms with notice (immediate for security-critical breaches).',
      'Upon termination, your right to use the Enterprise features ends. The Apache 2.0 open-source core remains licensed to you under those terms.',
      'Sections 6, 7, 8, 9, and 12 survive termination.',
    ],
  },
  {
    id: 'changes',
    title: '11. Changes',
    body: [
      'We may update these Terms. Material changes will be announced via email to active customers. Continuing to use the service after a change constitutes acceptance.',
      `Last updated: ${EFFECTIVE}`,
    ],
  },
  {
    id: 'governing-law',
    title: '12. Governing Law',
    body: [
      'These Terms are governed by the laws of India, without regard to its conflict-of-laws principles. Any dispute will be submitted to the exclusive jurisdiction of courts located in Hyderabad, Telangana, India.',
      'Nothing in these Terms prevents either party from seeking injunctive relief in any court of competent jurisdiction.',
    ],
  },
  {
    id: 'contact',
    title: '13. Contact',
    body: [
      'For questions about these Terms:',
      'contact@infrays.org',
      'infraYS Solutions, c/o Postal Address (to be added), India.',
    ],
  },
]

export default function TermsPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <span className="badge-cyan mb-4">Legal</span>
            <h1 className="text-5xl font-black tracking-tight mb-4">Terms of Service</h1>
            <p className="text-sm text-white/40">Effective {EFFECTIVE} · Questions? <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a></p>
          </div>
        </section>

        <section className="section py-12">
          <div className="container-md grid lg:grid-cols-[220px_1fr] gap-12">
            <aside className="hidden lg:block sticky top-24 self-start text-sm space-y-1">
              {sections.map((s) => (
                <a key={s.id} href={`#${s.id}`}
                  className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md transition-colors">
                  {s.title}
                </a>
              ))}
            </aside>

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
                See also: <Link to="/privacy" className="text-cyan-400 hover:underline">Privacy Policy</Link> · <Link to="/gdpr" className="text-cyan-400 hover:underline">GDPR / Data Requests</Link> · <Link to="/security" className="text-cyan-400 hover:underline">Security</Link>
              </div>
            </article>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
