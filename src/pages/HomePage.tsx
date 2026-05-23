import Navbar from '@/components/Navbar'
import Hero from '@/components/Hero'
import LogoStrip from '@/components/LogoStrip'
import StatsBar from '@/components/StatsBar'
import FeaturesSection from '@/components/FeaturesSection'
import HowItWorks from '@/components/HowItWorks'
import ComparisonTable from '@/components/ComparisonTable'
import IntegrationsGrid from '@/components/IntegrationsGrid'
import Testimonials from '@/components/Testimonials'
import PricingSection from '@/components/PricingSection'
import CTABanner from '@/components/CTABanner'
import Footer from '@/components/Footer'

export default function HomePage() {
  return (
    <>
      <Navbar />
      <main>
        <Hero />
        {/* LogoStrip sits immediately under Hero for social proof
            above the fold-ish (depends on viewport). */}
        <LogoStrip />
        <StatsBar />
        <FeaturesSection />
        <HowItWorks />
        <ComparisonTable />
        <IntegrationsGrid />
        <Testimonials />
        <PricingSection />
        <CTABanner />
      </main>
      <Footer />
    </>
  )
}
