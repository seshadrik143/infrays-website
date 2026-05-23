import { Link } from 'react-router-dom'
import { Zap } from 'lucide-react'
import NewsletterSignup from './NewsletterSignup'

// Removed Plugin SDK / API Reference from Developers — both pointed at
// /docs#sdk and /docs#api anchors that don't exist, so clicking did
// nothing useful. Restore once docs sections are written.
//
// Legal: kept as a placeholder pointing at /contact for now. Real
// /privacy and /terms pages need to be authored before Stripe Live +
// any compliance review. Tracked separately.
//
// Socials removed — we don't have real Twitter / LinkedIn / Discord
// accounts yet, so href='#' was dead-clickbait. Restore as
// <a href="..."> entries when accounts exist.
const footerLinks = {
  Product: [
    { href: '/features', label: 'Features' },
    { href: '/enterprise', label: 'Enterprise' },
    { href: '/plugins', label: 'Plugin Catalog' },
    { href: '/pricing', label: 'Pricing' },
    { href: '/integrations', label: 'Integrations' },
    { href: '/install', label: 'Installation' },
  ],
  Developers: [
    { href: '/docs', label: 'Documentation' },
    { href: '/cli', label: 'CLI Reference' },
    { href: '/changelog', label: 'Changelog' },
  ],
  Company: [
    { href: '/blog', label: 'Blog' },
    { href: '/changelog', label: "What's New" },
    { href: '/contact', label: 'Contact' },
  ],
  Legal: [
    { href: '/privacy', label: 'Privacy Policy' },
    { href: '/terms', label: 'Terms of Service' },
    { href: '/gdpr', label: 'GDPR / Data Requests' },
    { href: '/security', label: 'Security' },
  ],
}

export default function Footer() {
  return (
    <footer className="border-t border-white/[0.06]" style={{ background: '#08080f' }}>
      <div className="max-w-7xl mx-auto px-6 py-16">
        <div className="grid grid-cols-2 md:grid-cols-6 gap-10 mb-16">
          {/* Brand */}
          <div className="col-span-2">
            <Link to="/" className="flex items-center gap-2.5 mb-4">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center"
                style={{ background: 'linear-gradient(135deg, #00d4ff, #8b5cf6)' }}>
                <Zap className="w-4 h-4 text-white fill-white" />
              </div>
              <span className="text-lg font-bold">infra<span className="text-gradient-cyan">YS</span></span>
            </Link>
            <p className="text-sm text-white/40 leading-relaxed mb-6 max-w-xs">
              The unified observability platform for modern infrastructure teams. Open-source core, enterprise ready.
            </p>
            {/* Newsletter signup — Formspree-routed (same form ID used
                by the contact page) with topic='newsletter' so we can
                filter in Zoho. Real ESP later. */}
            <div className="max-w-xs">
              <NewsletterSignup />
            </div>
          </div>

          {/* Links */}
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h4 className="text-xs font-semibold text-white/30 uppercase tracking-widest mb-4">{category}</h4>
              <ul className="space-y-2.5">
                {links.map((link) => (
                  <li key={link.label}>
                    {(link.href.startsWith('http') || link.href.startsWith('mailto')) ? (
                      <a href={link.href}
                        target={link.href.startsWith('http') ? '_blank' : undefined}
                        rel={link.href.startsWith('http') ? 'noopener noreferrer' : undefined}
                        className="text-sm text-white/50 hover:text-white/90 transition-colors">
                        {link.label}
                      </a>
                    ) : (
                      <Link to={link.href}
                        className="text-sm text-white/50 hover:text-white/90 transition-colors">
                        {link.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom bar */}
        <div className="flex flex-col md:flex-row items-center justify-between gap-4 pt-8 border-t border-white/[0.06]">
          <p className="text-xs text-white/25">
            © 2026 infraYS. All rights reserved.
          </p>
          <div className="flex items-center gap-6">
            <span className="badge-green text-xs">
              <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
              All Systems Operational
            </span>
            <span className="text-xs text-white/25">v0.33.0 — Apache 2.0</span>
          </div>
        </div>
      </div>
    </footer>
  )
}
