import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Menu, X, Zap, ChevronDown } from 'lucide-react'

const navLinks = [
  {
    label: 'Product',
    children: [
      { to: '/features', label: 'Features', desc: 'Full platform overview' },
      { to: '/integrations', label: 'Integrations', desc: '67+ plugins & connectors' },
      { to: '/enterprise', label: 'Enterprise', desc: 'Security, compliance & HA' },
      { to: '/install', label: 'Installation', desc: 'Up in 60 seconds' },
    ],
  },
  { to: '/pricing', label: 'Pricing' },
  {
    label: 'Developers',
    children: [
      { to: '/docs', label: 'Documentation', desc: 'Guides & API reference' },
      { to: '/cli', label: 'CLI Reference', desc: '20+ commands, shell completion' },
      { to: '/changelog', label: 'Changelog', desc: '27 phases of features shipped' },
    ],
  },
  { to: '/blog', label: 'Blog' },
]

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [openDropdown, setOpenDropdown] = useState<string | null>(null)

  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 20)
    window.addEventListener('scroll', handler)
    return () => window.removeEventListener('scroll', handler)
  }, [])

  return (
    <header
      className="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
      style={scrolled ? {
        background: 'rgba(6,6,16,0.92)',
        backdropFilter: 'blur(16px)',
        WebkitBackdropFilter: 'blur(16px)',
        borderBottom: '1px solid rgba(255,255,255,0.06)',
        boxShadow: '0 4px 30px rgba(0,0,0,0.5)',
      } : undefined}
    >
      <nav className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center"
            style={{ background: 'linear-gradient(135deg, #00d4ff, #8b5cf6)' }}>
            <Zap className="w-4 h-4 text-white fill-white" />
          </div>
          <span className="text-lg font-bold tracking-tight text-white">
            infra<span className="text-gradient-cyan">YS</span>
          </span>
        </Link>

        <div className="hidden md:flex items-center gap-1">
          {navLinks.map((link) =>
            link.children ? (
              <div key={link.label} className="relative"
                onMouseEnter={() => setOpenDropdown(link.label)}
                onMouseLeave={() => setOpenDropdown(null)}>
                <button className="btn-ghost flex items-center gap-1">
                  {link.label}
                  <ChevronDown className={`w-3.5 h-3.5 transition-transform ${openDropdown === link.label ? 'rotate-180' : ''}`} />
                </button>
                {openDropdown === link.label && (
                  <div className="absolute top-full left-0 pt-2 w-56">
                    <div className="rounded-xl p-2" style={{
                      background: 'rgba(13,13,26,0.95)',
                      backdropFilter: 'blur(16px)',
                      border: '1px solid rgba(255,255,255,0.08)',
                      boxShadow: '0 20px 60px rgba(0,0,0,0.6)',
                    }}>
                      {link.children.map((child) => (
                        <Link key={child.to} to={child.to}
                          className="flex flex-col px-3 py-2.5 rounded-lg hover:bg-white/5 transition-colors">
                          <span className="text-sm font-medium text-white/90 hover:text-cyan-400 transition-colors">{child.label}</span>
                          <span className="text-xs text-white/40 mt-0.5">{child.desc}</span>
                        </Link>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <Link key={link.to} to={link.to!} className="btn-ghost">
                {link.label}
              </Link>
            )
          )}
        </div>

        <div className="hidden md:flex items-center gap-3">
          <a href="https://github.com/seshadrik143/infrays" target="_blank" rel="noopener noreferrer"
            className="btn-ghost flex items-center gap-2 text-sm">
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
            </svg>
            GitHub
          </a>
          <Link to="/install" className="btn-primary text-sm py-2.5 px-5">
            Get Started Free
          </Link>
        </div>

        <button className="md:hidden p-2 rounded-lg text-white/60 hover:text-white hover:bg-white/5 transition-all"
          onClick={() => setMobileOpen(!mobileOpen)}>
          {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
      </nav>

      {mobileOpen && (
        <div className="md:hidden px-6 py-6 space-y-4" style={{
          background: 'rgba(13,13,26,0.97)',
          backdropFilter: 'blur(16px)',
          borderTop: '1px solid rgba(255,255,255,0.06)',
        }}>
          {navLinks.map((link) =>
            link.children ? (
              <div key={link.label}>
                <p className="text-xs font-semibold text-white/30 uppercase tracking-widest mb-2">{link.label}</p>
                {link.children.map((child) => (
                  <Link key={child.to} to={child.to}
                    onClick={() => setMobileOpen(false)}
                    className="block py-2 text-sm text-white/70 hover:text-white transition-colors pl-2">
                    {child.label}
                  </Link>
                ))}
              </div>
            ) : (
              <Link key={link.to} to={link.to!}
                onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm text-white/70 hover:text-white transition-colors font-medium">
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
