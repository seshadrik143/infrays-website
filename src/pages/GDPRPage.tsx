import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'

// GDPR / data-request page. Focused, single-purpose. Cross-links
// to the broader Privacy Policy. Standard SaaS expression of EU
// data subject rights + how to invoke them.
const EFFECTIVE = '2026-05-23'

export default function GDPRPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md">
            <span className="badge-cyan mb-4">Legal</span>
            <h1 className="text-5xl font-black tracking-tight mb-4">GDPR &amp; Data Requests</h1>
            <p className="text-sm text-white/40">Effective {EFFECTIVE} · For requests, email <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a></p>
          </div>
        </section>

        <section className="section py-12">
          <div className="container-md grid lg:grid-cols-[220px_1fr] gap-12">
            <aside className="hidden lg:block sticky top-24 self-start text-sm space-y-1">
              <a href="#scope" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">1. Scope</a>
              <a href="#role" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">2. Our role</a>
              <a href="#rights" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">3. Your rights</a>
              <a href="#how" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">4. How to request</a>
              <a href="#timeline" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">5. Timeline &amp; verification</a>
              <a href="#dpa" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">6. Data Processing Addendum</a>
              <a href="#contact" className="block px-3 py-1.5 text-white/40 hover:text-white/80 rounded-md">7. Contact &amp; complaints</a>
            </aside>

            <article className="prose-invert max-w-none">
              <section id="scope" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">1. Scope</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">This page documents how infraYS Solutions handles requests under the EU General Data Protection Regulation (GDPR), the UK Data Protection Act, and equivalent regulations in other jurisdictions (CCPA, LGPD, PIPEDA, etc.).</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">For the full description of what data we collect and why, see our <Link to="/privacy" className="text-cyan-400 hover:underline">Privacy Policy</Link>.</p>
              </section>

              <section id="role" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">2. Our role</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm"><strong className="text-white">Data Controller</strong> — for your account information (email, name, billing details, license history) we are the data controller under GDPR.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm"><strong className="text-white">Data Processor</strong> — for any operational data your self-hosted NodePulse install pushes to our managed services (currently: license heartbeat metadata only — IP, deployment ID, version), we act as data processor on your behalf.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm"><strong className="text-white">Sub-processors:</strong> Stripe (billing), Postmark (transactional email), Fly.io (hosting). See <Link to="/privacy#sharing" className="text-cyan-400 hover:underline">Privacy Policy §4</Link> for links.</p>
              </section>

              <section id="rights" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">3. Your rights</h2>
                <ul className="text-white/60 text-sm space-y-2 list-disc pl-5">
                  <li><strong className="text-white">Access</strong> — request a copy of personal data we hold about you.</li>
                  <li><strong className="text-white">Rectification</strong> — correct inaccurate or incomplete data.</li>
                  <li><strong className="text-white">Erasure</strong> ("right to be forgotten") — request deletion of your data. Some records (audit logs, financial records) are retained for legal compliance.</li>
                  <li><strong className="text-white">Restriction</strong> — limit how we process your data (e.g. pending an objection).</li>
                  <li><strong className="text-white">Portability</strong> — receive your data in a structured, machine-readable format (JSON).</li>
                  <li><strong className="text-white">Objection</strong> — object to processing based on legitimate interest.</li>
                  <li><strong className="text-white">Withdraw consent</strong> — if processing was based on consent, you can withdraw it at any time.</li>
                  <li><strong className="text-white">No automated decision-making</strong> — we do not perform profiling or automated decisions that produce legal effects.</li>
                </ul>
              </section>

              <section id="how" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">4. How to request</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">Email <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a> from the address associated with your account. Subject line: "GDPR access request" / "GDPR deletion request" / etc.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">Include: your account email, the type of request, and any clarification of what data you want or what you want done.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">There is no fee for the first request in any 12-month period. We may charge a reasonable fee for repeat or manifestly unfounded requests, as permitted by Art. 12(5) GDPR.</p>
              </section>

              <section id="timeline" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">5. Timeline &amp; verification</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">We will respond within <strong className="text-white">30 calendar days</strong> as required by Art. 12(3) GDPR. For complex requests, we may extend by up to 60 days and will notify you of the extension within the initial 30 days.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">We will verify your identity before fulfilling a request — typically by asking you to respond from the account email. For erasure or access requests involving sensitive data, we may request additional verification.</p>
              </section>

              <section id="dpa" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">6. Data Processing Addendum (DPA)</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">If you are an EU/EEA/UK-based business customer and need a signed DPA covering our processor relationship, contact us. We use Standard Contractual Clauses (SCCs) for international data transfers.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">We do not currently appoint a separate EU representative under Art. 27 GDPR. If your jurisdiction requires data residency that we cannot meet, the self-hosted NodePulse option keeps all data within your infrastructure.</p>
              </section>

              <section id="contact" className="mb-10 scroll-mt-24">
                <h2 className="text-xl font-bold text-white mb-4">7. Contact &amp; complaints</h2>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">For all GDPR matters: <a className="text-cyan-400" href="mailto:contact@infrays.org">contact@infrays.org</a>.</p>
                <p className="text-white/60 leading-relaxed mb-3 text-sm">You have the right to lodge a complaint with your local supervisory authority. For EU residents, find your authority at <a className="text-cyan-400" href="https://edpb.europa.eu/about-edpb/about-edpb/members_en" target="_blank" rel="noopener noreferrer">edpb.europa.eu</a>.</p>
              </section>

              <div className="mt-12 border-t border-white/10 pt-6 text-xs text-white/30">
                See also: <Link to="/privacy" className="text-cyan-400 hover:underline">Privacy Policy</Link> · <Link to="/terms" className="text-cyan-400 hover:underline">Terms of Service</Link> · <Link to="/security" className="text-cyan-400 hover:underline">Security</Link>
              </div>
            </article>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
