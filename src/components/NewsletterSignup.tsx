import { useState } from 'react'
import { Mail, Check, ArrowRight } from 'lucide-react'

// NewsletterSignup — uses the existing Formspree form ID with a topic
// field tagged "newsletter" so submissions land in the same inbox as
// contact-form messages but are easy to filter.
//
// Better choice longer-term: ConvertKit / Buttondown / Resend.
// For now this works without adding a new service or API key.
const FORMSPREE_URL = 'https://formspree.io/f/xzdokqzo'

type State = 'idle' | 'sending' | 'sent' | 'error'

export default function NewsletterSignup() {
  const [email, setEmail] = useState('')
  const [state, setState] = useState<State>('idle')

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || state === 'sending' || state === 'sent') return
    setState('sending')
    try {
      const res = await fetch(FORMSPREE_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({
          email,
          topic: 'newsletter',
          // Honeypot-style _subject + _replyto for Formspree's own routing.
          _subject: 'infrays.org newsletter signup',
        }),
      })
      if (res.ok) {
        setState('sent')
      } else {
        setState('error')
      }
    } catch {
      setState('error')
    }
  }

  return (
    <div className="border border-white/10 rounded-2xl p-5"
      style={{ background: 'rgba(0,212,255,0.04)' }}>
      <div className="flex items-center gap-2 mb-2">
        <Mail className="w-4 h-4 text-cyan-400" />
        <h4 className="text-xs font-semibold text-white/70 uppercase tracking-widest">Newsletter</h4>
      </div>
      <p className="text-xs text-white/40 mb-4 leading-relaxed">
        ~1 email / month. Release notes, deep-dives, customer stories.
        No spam, no third-party tracking.
      </p>
      <form onSubmit={submit} className="flex flex-col gap-2">
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={state === 'sending' || state === 'sent'}
          placeholder="you@yourcompany.com"
          className="bg-black/30 border border-white/10 rounded-md px-3 py-2 text-sm text-white placeholder:text-white/30 focus:border-cyan-400 focus:outline-none disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={state === 'sending' || state === 'sent'}
          className="flex items-center justify-center gap-2 rounded-md bg-cyan-500 text-white font-semibold text-sm py-2 hover:bg-cyan-400 disabled:opacity-50 transition-colors"
        >
          {state === 'sent' ? (
            <>
              <Check className="w-4 h-4" />
              Subscribed
            </>
          ) : state === 'sending' ? (
            'Sending…'
          ) : (
            <>
              Subscribe
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </button>
        {state === 'error' && (
          <p className="text-xs text-red-400">Something went wrong. Try again or email contact@infrays.org.</p>
        )}
      </form>
    </div>
  )
}
