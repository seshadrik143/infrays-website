import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Menu, X, Zap, ChevronDown } from 'lucide-react'

const navLinks = [
  {
    label: 'Product',
    children: [
      { to: '/features',     label: 'Features',       desc: 'Full platform overview' },
      { to: '/plugins',      label: 'Plugin Catalog',  desc: '52+ collector & notifier plugins' },
      { to: '/integrations', label: 'Integrations',   desc: 'Full tech stack coverage' },
      { to: '/enterprise',   label: 'Enterprise',     desc: 'Security, compliance & HA' },
      { to: '/install',      label: 'Installation',   desc: 'Up in 60 seconds' },
    ],
  },
  { to: '/pricing', label: 'Pricing' },
  { to: '/contact', label: 'Contact' },
  {
    label: 'Developers',
    children: [
      { to: '/docs',       label: 'Documentation', desc: 'Guides & API reference' },
      { to: '/cli',        label: 'CLI Reference',  desc: '20+ commands, shell completion' },
      { to: '/changelog',  label: 'Changelog',      desc: '27 phases of features shipped' },
    ],
  },
  { to: '/blog', label: 'Blog' },
]

export default function Navbar() {
  const [scrolled, setScrolled]       = useState(false)
  const [mobileOpen, setMobileOpen]   = useState(false)
  const [openDropdown, setOpenDropdown] = useState<string | null>(null)

  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 20)
    window.addEventListener('scroll', handler)
    return () => window.removeEventListener('scroll', handler)
  }, [])

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${
        scrolled ? 'py-0' : 'py-0'
      }`}
      style={scrolled ? {
        background: 'rgba(3,3,8,0.88)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255,255,255,0.07)',
        boxShadow: '0 8px 40px rgba(0,0,0,0.6), 0 1px 0 rgba(0,212,255,0.04)',
      } : {
        background: 'transparent',
      }}
    >
      <nav className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2.5 group shrink-0">
          <img
            src="/icons/infrays-logo.svg"
            alt="infraYS"
            className="h-9 transition-opacity duration-300 group-hover:opacity-90"
          />
        </Link>

        {/* Desktop nav */}
        <div className="hidden md:flex items-center gap-0.5">
          {navLinks.map((link) =>
            link.children ? (
              <div key={link.label} className="relative"
                onMouseEnter={() => setOpenDropdown(link.label)}
                onMouseLeave={() => setOpenDropdown(null)}>
                <button className="btn-ghost flex items-center gap-1.5 text-sm">
                  {link.label}
                  <ChevronDown className={`w-3.5 h-3.5 transition-transform duration-200 ${
                    openDropdown === link.label ? 'rotate-180 text-cyan-400' : ''
                  }`} />
                </button>

                {/* Dropdown */}
                <div className={`absolute top-full left-0 pt-3 w-60 transition-all duration-200 ${
                  openDropdown === link.label
                    ? 'opacity-100 translate-y-0 pointer-events-auto'
                    : 'opacity-0 -translate-y-2 pointer-events-none'
                }`}>
                  <div className="rounded-xl p-1.5" style={{
                    background: 'rgba(10,10,20,0.97)',
                    backdropFilter: 'blur(20px)',
                    border: '1px solid rgba(255,255,255,0.08)',
                    boxShadow: '0 24px 60px rgba(0,0,0,0.7), 0 0 0 1px rgba(0,212,255,0.04)',
                  }}>
                    {link.children.map((child) => (
                      <Link key={child.to} to={child.to}
                        className="flex flex-col px-3.5 py-3 rounded-lg hover:bg-white/[0.05] transition-colors group/item">
                        <span className="text-sm font-medium text-white/80 group-hover/item:text-cyan-400 transition-colors">
                          {child.label}
                        </span>
                        <span className="text-xs text-white/30 mt-0.5">{child.desc}</span>
                      </Link>
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <Link key={link.to} to={link.to!} className="btn-ghost text-sm">
                {link.label}
              </Link>
            )
          )}
        </div>

        {/* Right CTAs */}
        <div className="hidden md:flex items-center gap-3">
          <Link to="/install" className="btn-primary text-sm py-2 px-5">
            Get Started Free
          </Link>
        </div>

        {/* Mobile hamburger */}
        <button className="md:hidden p-2 rounded-lg text-white/60 hover:text-white hover:bg-white/[0.06] transition-all"
          onClick={() => setMobileOpen(!mobileOpen)}>
          {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
      </nav>

      {/* Mobile menu */}
      {mobileOpen && (
        <div className="md:hidden px-6 py-6 space-y-5" style={{
          background: 'rgba(8,8,18,0.98)',
          backdropFilter: 'blur(20px)',
          borderTop: '1px solid rgba(255,255,255,0.06)',
        }}>
          {navLinks.map((link) =>
            link.children ? (
              <div key={link.label}>
                <p className="text-[11px] font-semibold text-white/25 uppercase tracking-[0.15em] mb-3">
                  {link.label}
                </p>
                <div className="space-y-1 pl-1">
                  {link.children.map((child) => (
                    <Link key={child.to} to={child.to}
                      onClick={() => setMobileOpen(false)}
                      className="block py-2 text-sm text-white/60 hover:text-white transition-colors">
                      {child.label}
                    </Link>
                  ))}
                </div>
              </div>
            ) : (
              <Link key={link.to} to={link.to!}
                onClick={() => setMobileOpen(false)}
                className="block text-sm font-medium text-white/70 hover:text-white transition-colors py-1">
                {link.label}
              </Link>
            )
          )}
          <div className="pt-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(255,255,255,0.06)' }}>
            <Link to="/install" onClick={() => setMobileOpen(false)} className="btn-primary justify-center">
              Get Started Free
            </Link>
          </div>
        </div>
      )}
    </header>
  )
}
