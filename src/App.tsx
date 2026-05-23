import React, { useEffect, Suspense, lazy } from 'react'
import { Routes, Route, useLocation } from 'react-router-dom'

// HomePage is loaded eagerly — it's the most-hit route, splitting it
// would just add a Suspense flicker on the entry path. Every other
// page is lazy-loaded so visitors only download what they navigate to.
//
// Vite's manualChunks isn't needed — React.lazy + Suspense gives us
// per-route chunks automatically, and Vite's build splits them by
// import boundary.
import HomePage from './pages/HomePage'
import CookieBanner from './components/CookieBanner'

const FeaturesPage     = lazy(() => import('./pages/FeaturesPage'))
const InstallPage      = lazy(() => import('./pages/InstallPage'))
const PricingPage      = lazy(() => import('./pages/PricingPage'))
const DocsPage         = lazy(() => import('./pages/DocsPage'))
const BlogPage         = lazy(() => import('./pages/BlogPage'))
const IntegrationsPage = lazy(() => import('./pages/IntegrationsPage'))
const ChangelogPage    = lazy(() => import('./pages/ChangelogPage'))
const EnterprisePage   = lazy(() => import('./pages/EnterprisePage'))
const CLIPage          = lazy(() => import('./pages/CLIPage'))
const BlogPostPage     = lazy(() => import('./pages/BlogPostPage'))
const PluginsPage      = lazy(() => import('./pages/PluginsPage'))
const ContactPage      = lazy(() => import('./pages/ContactPage'))
const PrivacyPage      = lazy(() => import('./pages/PrivacyPage'))
const TermsPage        = lazy(() => import('./pages/TermsPage'))
const GDPRPage         = lazy(() => import('./pages/GDPRPage'))
const SecurityPage     = lazy(() => import('./pages/SecurityPage'))
const AboutPage        = lazy(() => import('./pages/AboutPage'))
const CustomersPage    = lazy(() => import('./pages/CustomersPage'))
const NotFoundPage     = lazy(() => import('./pages/NotFoundPage'))


class ErrorBoundary extends React.Component<{children: React.ReactNode}, {error: Error | null}> {
  constructor(props: {children: React.ReactNode}) {
    super(props)
    this.state = { error: null }
  }
  static getDerivedStateFromError(error: Error) {
    return { error }
  }
  render() {
    if (this.state.error) {
      return (
        <div style={{ background: '#060610', color: '#fff', padding: '40px', minHeight: '100vh', fontFamily: 'monospace' }}>
          <h1 style={{ color: '#ff4444', marginBottom: '20px' }}>Runtime Error</h1>
          <pre style={{ color: '#ffaa00', whiteSpace: 'pre-wrap' }}>{this.state.error.message}</pre>
          <pre style={{ color: '#888', marginTop: '20px', fontSize: '12px', whiteSpace: 'pre-wrap' }}>{this.state.error.stack}</pre>
        </div>
      )
    }
    return this.props.children
  }
}

function ScrollToTop() {
  const { pathname } = useLocation()
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [pathname])
  return null
}

// PageFallback — what shows while a lazy-loaded chunk is being fetched.
// Minimal so it doesn't visually conflict with whatever page is loading.
// Brand-color spinner + dark background = no jarring white flash.
function PageFallback() {
  return (
    <div style={{
      minHeight: '60vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: '#060610',
    }}>
      <div style={{
        width: '32px',
        height: '32px',
        border: '2px solid rgba(0,212,255,0.15)',
        borderTopColor: '#22d3ee',
        borderRadius: '50%',
        animation: 'spin 0.9s linear infinite',
      }} />
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  )
}

export default function App() {
  return (
    <ErrorBoundary>
      <ScrollToTop />
      <Suspense fallback={<PageFallback />}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/features" element={<FeaturesPage />} />
          <Route path="/install" element={<InstallPage />} />
          <Route path="/pricing" element={<PricingPage />} />
          <Route path="/docs" element={<DocsPage />} />
          <Route path="/blog" element={<BlogPage />} />
          <Route path="/integrations" element={<IntegrationsPage />} />
          <Route path="/changelog" element={<ChangelogPage />} />
          <Route path="/enterprise" element={<EnterprisePage />} />
          <Route path="/cli" element={<CLIPage />} />
          <Route path="/blog/:slug" element={<BlogPostPage />} />
          <Route path="/plugins" element={<PluginsPage />} />
          <Route path="/contact" element={<ContactPage />} />
          <Route path="/privacy" element={<PrivacyPage />} />
          <Route path="/terms" element={<TermsPage />} />
          <Route path="/gdpr" element={<GDPRPage />} />
          <Route path="/security" element={<SecurityPage />} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/customers" element={<CustomersPage />} />
          {/* /vs/datadog removed in Phase C.5 — competitor comparison
              advertising without legal counsel is too much exposure. */}
          {/* Catch-all 404 — must be LAST so any path that doesn't match
              an earlier route falls through here. */}
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Suspense>
      <CookieBanner />
    </ErrorBoundary>
  )
}
