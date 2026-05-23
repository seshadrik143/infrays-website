import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { X } from 'lucide-react'

// Minimal cookie / privacy disclosure banner. The site does NOT set
// tracking cookies and does NOT use third-party analytics — so this
// is mostly a transparency notice rather than a consent gate. Still
// shows because EU + India PDPB-style regulations expect at least
// disclosure on every first-party site that handles any cookies
// (even essential session cookies on subdomains).
//
// Storage: dismissed state goes into localStorage under a versioned
// key. Bumping the version (e.g. v2) will re-show the banner to
// every visitor — use that if the disclosure ever changes materially.

const DISMISS_KEY = 'infrays.cookie-banner.dismissed.v1'

export default function CookieBanner() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    // Run only on the client to avoid SSR mismatches.
    try {
      if (!localStorage.getItem(DISMISS_KEY)) {
        setVisible(true)
      }
    } catch {
      // localStorage may be disabled (private browsing, embedded
      // contexts). Default to showing the banner once per session.
      setVisible(true)
    }
  }, [])

  const dismiss = () => {
    setVisible(false)
    try { localStorage.setItem(DISMISS_KEY, '1') } catch { /* noop */ }
  }

  if (!visible) return null

  return (
    <div
      role="dialog"
      aria-label="Cookie and privacy disclosure"
      className="fixed bottom-4 inset-x-4 md:bottom-6 md:inset-x-auto md:right-6 md:max-w-md z-50 border border-white/10 rounded-2xl backdrop-blur-md p-5 shadow-2xl"
      style={{ background: 'rgba(8,8,18,0.92)' }}
    >
      <button
        onClick={dismiss}
        aria-label="Dismiss"
        className="absolute top-3 right-3 text-white/30 hover:text-white/70 transition-colors"
      >
        <X className="w-4 h-4" />
      </button>
      <p className="text-sm text-white/70 leading-relaxed mb-3 pr-6">
        We use only essential cookies (login session, your dismiss preference for this banner). No third-party trackers, no advertising cookies, no behavioral analytics.
      </p>
      <div className="flex items-center gap-3 text-xs">
        <button
          onClick={dismiss}
          className="px-3 py-1.5 rounded-md bg-cyan-500 text-white font-semibold hover:bg-cyan-400 transition-colors"
        >
          Got it
        </button>
        <Link to="/privacy" className="text-cyan-400 hover:text-cyan-300 underline-offset-2 hover:underline">
          Privacy Policy
        </Link>
      </div>
    </div>
  )
}
