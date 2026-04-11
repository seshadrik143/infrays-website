import React, { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import FeaturesPage from './pages/FeaturesPage'
import InstallPage from './pages/InstallPage'
import PricingPage from './pages/PricingPage'
import DocsPage from './pages/DocsPage'
import BlogPage from './pages/BlogPage'
import IntegrationsPage from './pages/IntegrationsPage'
import ChangelogPage from './pages/ChangelogPage'
import EnterprisePage from './pages/EnterprisePage'
import CLIPage from './pages/CLIPage'
import BlogPostPage from './pages/BlogPostPage'
import PluginsPage from './pages/PluginsPage'
import ContactPage from './pages/ContactPage'


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
  // Scroll to top on route change
  useEffect(() => {
    window.scrollTo(0, 0)
  })
  return null
}

export default function App() {
  return (
    <ErrorBoundary>
      <ScrollToTop />
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
      </Routes>
    </ErrorBoundary>
  )
}
