
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import IntegrationsGrid from '@/components/IntegrationsGrid'
import CTABanner from '@/components/CTABanner'

export default function IntegrationsPage() {
  return (
    <>
      <Navbar />
      <main className="pt-24">
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-purple mb-4">Integrations</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Works with your{' '}
              <span className="text-gradient-purple">entire stack</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              Native OTLP support means any OpenTelemetry-compatible tool works out of the box.
              100+ built-in integrations with no extra config.
            </p>
          </div>
        </section>
        <IntegrationsGrid />
        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
