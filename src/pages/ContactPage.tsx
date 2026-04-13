import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { useState } from 'react'
import { Mail, MessageSquare, ArrowRight, CheckCircle2, Copy } from 'lucide-react'

const SALES_EMAIL = 'sales@infrays.org'

const topics = [
  'Sales & Pricing',
  'Enterprise Plan',
  'Technical Question',
  'Partnership',
  'Other',
]

export default function ContactPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [topic, setTopic] = useState(topics[0])
  const [message, setMessage] = useState('')
  const [copied, setCopied] = useState(false)

  function handleSend(e: React.FormEvent) {
    e.preventDefault()
    const subject = encodeURIComponent(`[infraYS] ${topic} — ${name}`)
    const body = encodeURIComponent(
      `Hi infraYS team,\n\n${message}\n\n---\nName: ${name}\nEmail: ${email}\nTopic: ${topic}`
    )
    window.location.href = `mailto:${SALES_EMAIL}?subject=${subject}&body=${body}`
  }

  function copyEmail() {
    navigator.clipboard.writeText(SALES_EMAIL)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Hero */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Contact</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Let's talk{' '}
              <span className="text-gradient-cyan">observability</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              Questions about pricing, enterprise plans, or technical setup?
              Drop us a message and we'll get back to you within 24 hours.
            </p>
          </div>
        </section>

        <section className="section py-16">
          <div className="container-md">
            <div className="grid md:grid-cols-5 gap-10">

              {/* Left — info */}
              <div className="md:col-span-2 space-y-8">
                {/* Email card */}
                <div className="border border-white/[0.07] rounded-2xl p-6"
                  style={{ background: 'rgba(17,17,32,0.7)' }}>
                  <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center mb-4">
                    <Mail className="w-5 h-5 text-cyan-400" />
                  </div>
                  <h3 className="text-sm font-bold text-white mb-1">Email us directly</h3>
                  <p className="text-xs text-white/40 mb-4">We reply within 24 hours on business days.</p>
                  <div className="flex items-center gap-2 bg-white/[0.04] border border-white/10 rounded-xl px-4 py-3">
                    <span className="text-sm font-mono text-cyan-400 flex-1 truncate">{SALES_EMAIL}</span>
                    <button onClick={copyEmail} className="text-white/30 hover:text-white/70 transition-colors flex-shrink-0">
                      {copied
                        ? <CheckCircle2 className="w-4 h-4 text-green-400" />
                        : <Copy className="w-4 h-4" />}
                    </button>
                  </div>
                </div>

                {/* What to expect */}
                <div className="border border-white/[0.07] rounded-2xl p-6"
                  style={{ background: 'rgba(17,17,32,0.7)' }}>
                  <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center mb-4">
                    <MessageSquare className="w-5 h-5 text-purple-400" />
                  </div>
                  <h3 className="text-sm font-bold text-white mb-4">What we can help with</h3>
                  <ul className="space-y-2.5">
                    {[
                      'Pricing & plan selection',
                      'Enterprise deployment',
                      'Self-hosted setup help',
                      'Plugin & SDK questions',
                      'Integration support',
                    ].map(item => (
                      <li key={item} className="flex items-center gap-2.5 text-sm text-white/50">
                        <CheckCircle2 className="w-4 h-4 text-green-400 flex-shrink-0" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              {/* Right — form */}
              <div className="md:col-span-3">
                <form onSubmit={handleSend}
                  className="border border-white/[0.07] rounded-2xl p-8 space-y-5"
                  style={{ background: 'rgba(17,17,32,0.7)' }}>

                  <div className="grid sm:grid-cols-2 gap-5">
                    <div>
                      <label className="block text-xs font-semibold text-white/40 uppercase tracking-wider mb-2">
                        Your Name *
                      </label>
                      <input
                        required
                        type="text"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        placeholder="Jane Smith"
                        className="w-full bg-white/[0.04] border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-white/25 focus:outline-none focus:border-cyan-500/40 transition-colors"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-white/40 uppercase tracking-wider mb-2">
                        Your Email *
                      </label>
                      <input
                        required
                        type="email"
                        value={email}
                        onChange={e => setEmail(e.target.value)}
                        placeholder="jane@company.com"
                        className="w-full bg-white/[0.04] border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-white/25 focus:outline-none focus:border-cyan-500/40 transition-colors"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-white/40 uppercase tracking-wider mb-2">
                      Topic
                    </label>
                    <select
                      value={topic}
                      onChange={e => setTopic(e.target.value)}
                      className="w-full bg-white/[0.04] border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-cyan-500/40 transition-colors"
                      style={{ background: 'rgba(17,17,32,0.9)' }}>
                      {topics.map(t => (
                        <option key={t} value={t}>{t}</option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-white/40 uppercase tracking-wider mb-2">
                      Message *
                    </label>
                    <textarea
                      required
                      rows={5}
                      value={message}
                      onChange={e => setMessage(e.target.value)}
                      placeholder="Tell us about your infrastructure, team size, and what you're trying to monitor..."
                      className="w-full bg-white/[0.04] border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-white/25 focus:outline-none focus:border-cyan-500/40 transition-colors resize-none"
                    />
                  </div>

                  <button type="submit" className="btn-primary w-full justify-center">
                    Send Message
                    <ArrowRight className="w-4 h-4" />
                  </button>

                  <p className="text-xs text-white/25 text-center">
                    Clicking "Send Message" will open your email client with the form pre-filled.
                  </p>
                </form>
              </div>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
