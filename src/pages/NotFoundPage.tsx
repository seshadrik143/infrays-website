import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Home, FileText, Mail } from 'lucide-react'

// Branded 404 page. The previous behavior was Vercel's default
// plain-text "404 NOT_FOUND" page (we saw this in the earlier QA
// pass). This component renders in the SPA when no route matches —
// see App.tsx's catch-all <Route path="*">.

export default function NotFoundPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24 min-h-[70vh] flex items-center">
        <div className="container-md text-center">
          {/* Big 404 glyph */}
          <div className="text-[120px] sm:text-[180px] font-black leading-none mb-3 text-gradient-cyan select-none">
            404
          </div>
          <h1 className="text-2xl sm:text-3xl font-black text-white mb-4">
            We can&apos;t find that page.
          </h1>
          <p className="text-white/50 max-w-md mx-auto mb-10 leading-relaxed">
            It might have moved, been removed, or never existed.
            Here are some places you might be looking for instead.
          </p>

          <div className="grid sm:grid-cols-3 gap-4 max-w-2xl mx-auto mb-10">
            <Link to="/" className="border border-white/[0.07] rounded-xl p-5 hover:border-cyan-500/30 transition-colors group">
              <Home className="w-5 h-5 text-cyan-400 mx-auto mb-3" />
              <div className="text-sm font-bold text-white mb-1 group-hover:text-cyan-400 transition-colors">Home</div>
              <div className="text-xs text-white/40">Start here.</div>
            </Link>

            <Link to="/docs" className="border border-white/[0.07] rounded-xl p-5 hover:border-cyan-500/30 transition-colors group">
              <FileText className="w-5 h-5 text-cyan-400 mx-auto mb-3" />
              <div className="text-sm font-bold text-white mb-1 group-hover:text-cyan-400 transition-colors">Docs</div>
              <div className="text-xs text-white/40">Setup + reference.</div>
            </Link>

            <Link to="/contact" className="border border-white/[0.07] rounded-xl p-5 hover:border-cyan-500/30 transition-colors group">
              <Mail className="w-5 h-5 text-cyan-400 mx-auto mb-3" />
              <div className="text-sm font-bold text-white mb-1 group-hover:text-cyan-400 transition-colors">Contact</div>
              <div className="text-xs text-white/40">Tell us what you needed.</div>
            </Link>
          </div>

          <a href="https://license.infrays.org/signup" className="btn-primary text-sm">
            Or sign up — 15 days free
            <ArrowRight className="w-4 h-4" />
          </a>
        </div>
      </main>
      <Footer />
    </>
  )
}
