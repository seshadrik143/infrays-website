
import Navbar from '@/components/Navbar'
import Footer from '@/components/Footer'
import { Link } from 'react-router-dom'
import { ArrowRight, Calendar, Clock } from 'lucide-react'

const posts = [
  {
    slug: 'infrays-v1-launch',
    title: 'Introducing infraYS v1.0: The Observability Platform Built for Teams That Never Sleep',
    excerpt: 'After 21 development phases, infraYS v1.0 is here. Metrics, logs, traces, profiling, AI-powered AIOps, synthetic monitoring, SLO tracking, and SaaS billing — all from a single 12MB agent.',
    date: '2026-04-08',
    readTime: '8 min read',
    category: 'Announcement',
    color: 'cyan',
    featured: true,
  },
  {
    slug: 'ai-anomaly-detection',
    title: 'How infraYS AI Catches Anomalies Before They Become Incidents',
    excerpt: 'Deep dive into the Z-score + Isolation Forest pipeline, rolling 2-week baselines, flap detection, and how predictive alerting uses linear regression to forecast threshold breaches.',
    date: '2026-04-06',
    readTime: '12 min read',
    category: 'Engineering',
    color: 'purple',
    featured: false,
  },
  {
    slug: 'synthetic-monitoring-global-probes',
    title: 'Synthetic Monitoring at Scale: HTTP, TCP, DNS, SSL, and Browser Checks',
    excerpt: 'How we built the synthetic probe system — from httptrace-powered HTTP checks to headless browser tests — and why a distributed probe network beats a single origin.',
    date: '2026-04-05',
    readTime: '10 min read',
    category: 'Engineering',
    color: 'green',
    featured: false,
  },
  {
    slug: 'self-hosted-vs-saas-observability',
    title: 'Self-Hosted vs SaaS Observability: Understanding the Cost Difference',
    excerpt: 'A breakdown of what it actually costs to run a full observability stack — infrastructure, storage, compute — when you self-host vs. pay for a managed SaaS solution.',
    date: '2026-04-04',
    readTime: '7 min read',
    category: 'Guide',
    color: 'orange',
    featured: false,
  },
  {
    slug: 'ebpf-autodiscovery',
    title: 'Auto-Discovery Without eBPF: How We Map Services Using /proc',
    excerpt: 'Full eBPF requires kernel 5.8+. We built a /proc-based auto-discovery engine that works everywhere — including decade-old bare metal servers — and still gives you a complete service topology.',
    date: '2026-04-03',
    readTime: '9 min read',
    category: 'Engineering',
    color: 'teal',
    featured: false,
  },
  {
    slug: 'continuous-profiling-go',
    title: 'Finding the Hidden 40ms: Continuous Profiling for Production Go Services',
    excerpt: 'A case study of how infraYS continuous profiling identified a hot path that was adding 40ms to every HTTP request — without any code instrumentation.',
    date: '2026-04-01',
    readTime: '6 min read',
    category: 'Case Study',
    color: 'pink',
    featured: false,
  },
]

const colorMap: Record<string, string> = {
  cyan:   'badge-cyan',
  purple: 'bg-purple-500/10 text-purple-400 border border-purple-500/20',
  green:  'badge-green',
  orange: 'bg-orange-500/10 text-orange-400 border border-orange-500/20',
  teal:   'bg-teal-500/10 text-teal-400 border border-teal-500/20',
  pink:   'bg-pink-500/10 text-pink-400 border border-pink-500/20',
}

export default function BlogPage() {
  const featured = posts.find((p) => p.featured)
  const rest = posts.filter((p) => !p.featured)

  return (
    <>
      <Navbar />
      <main className="pt-24">
        {/* Header */}
        <section className="hero-bg section py-16 border-b border-white/[0.06]">
          <div className="container-md text-center">
            <span className="badge-cyan mb-4">Blog</span>
            <h1 className="text-5xl font-black tracking-tight mb-5">
              Engineering insights &{' '}
              <span className="text-gradient-cyan">release notes</span>
            </h1>
            <p className="text-lg text-white/40 max-w-xl mx-auto">
              How we build infraYS, observability best practices, and deep technical dives.
            </p>
          </div>
        </section>

        <section className="section">
          <div className="container-lg space-y-6">
            {/* Featured post */}
            {featured && (
              <Link to={`/blog/${featured.slug}`}
                className="group block border border-white/[0.07] rounded-2xl p-8 hover:border-cyan-500/20 transition-all"
                style={{ background: 'linear-gradient(135deg, rgba(0,212,255,0.05), rgba(17,17,32,0.8))' }}>
                <div className="flex items-start justify-between gap-4 mb-4">
                  <span className={`badge ${colorMap[featured.color]}`}>{featured.category}</span>
                  <span className="badge-cyan text-xs">Featured</span>
                </div>
                <h2 className="text-2xl font-black text-white mb-3 group-hover:text-cyan-300 transition-colors leading-snug">
                  {featured.title}
                </h2>
                <p className="text-white/45 leading-relaxed mb-6">{featured.excerpt}</p>
                <div className="flex items-center gap-4 text-xs text-white/30">
                  <span className="flex items-center gap-1.5"><Calendar className="w-3.5 h-3.5" />{featured.date}</span>
                  <span className="flex items-center gap-1.5"><Clock className="w-3.5 h-3.5" />{featured.readTime}</span>
                  <span className="ml-auto text-cyan-400 group-hover:translate-x-1 transition-transform flex items-center gap-1">
                    Read more <ArrowRight className="w-3.5 h-3.5" />
                  </span>
                </div>
              </Link>
            )}

            {/* Rest */}
            <div className="grid md:grid-cols-2 gap-5">
              {rest.map((post) => (
                <Link key={post.slug} to={`/blog/${post.slug}`}
                  className="group border border-white/[0.07] rounded-2xl p-6 hover:border-white/15 transition-all flex flex-col"
                  style={{ background: 'rgba(17,17,32,0.7)' }}>
                  <span className={`badge ${colorMap[post.color]} w-fit mb-4`}>{post.category}</span>
                  <h2 className="text-base font-bold text-white mb-3 group-hover:text-cyan-300 transition-colors leading-snug flex-1">
                    {post.title}
                  </h2>
                  <p className="text-sm text-white/40 leading-relaxed mb-5 line-clamp-2">{post.excerpt}</p>
                  <div className="flex items-center gap-4 text-xs text-white/25">
                    <span className="flex items-center gap-1.5"><Calendar className="w-3 h-3" />{post.date}</span>
                    <span className="flex items-center gap-1.5"><Clock className="w-3 h-3" />{post.readTime}</span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
