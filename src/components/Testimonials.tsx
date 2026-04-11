import { MessageSquare, ArrowRight, Users, Star, GitPullRequest } from 'lucide-react'

const cards = [
  {
    icon: Star,
    color: 'cyan',
    title: 'Early Adopter',
    desc: 'Deploy infraYS in your infrastructure today and be part of shaping the roadmap. Your feedback drives every release.',
  },
  {
    icon: GitPullRequest,
    color: 'purple',
    title: 'Share Your Experience',
    desc: "Once you've used infraYS in production, we'd love to feature your story. Real engineers, real results.",
  },
  {
    icon: Users,
    color: 'green',
    title: 'Open Source Community',
    desc: 'Contribute collectors, plugins, dashboard templates, or bug fixes. Every contribution makes the platform better.',
  },
]

const colorMap: Record<string, { bg: string; border: string; text: string; glow: string }> = {
  cyan:   { bg: 'rgba(0,212,255,0.08)',  border: 'rgba(0,212,255,0.2)',  text: 'text-cyan-400',   glow: '0 0 20px rgba(0,212,255,0.1)' },
  purple: { bg: 'rgba(168,85,247,0.08)', border: 'rgba(168,85,247,0.2)', text: 'text-purple-400', glow: '0 0 20px rgba(168,85,247,0.1)' },
  green:  { bg: 'rgba(16,185,129,0.08)', border: 'rgba(16,185,129,0.2)', text: 'text-green-400',  glow: '0 0 20px rgba(16,185,129,0.1)' },
}

export default function Testimonials() {
  return (
    <section className="section relative overflow-hidden" id="community">
      {/* Background mesh */}
      <div className="absolute inset-0 mesh-bg pointer-events-none opacity-40" />

      <div className="container-lg relative z-10">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-cyan mb-5">Community</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Built for teams.{' '}
            <span className="text-gradient-cyan">Loved by engineers.</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto leading-relaxed">
            infraYS is open source and actively growing. Be among the first teams to deploy it
            and shape the platform with your feedback.
          </p>
        </div>

        {/* Cards */}
        <div className="grid md:grid-cols-3 gap-5 mb-10">
          {cards.map((card) => {
            const c = colorMap[card.color]
            return (
              <div key={card.title}
                className="group relative rounded-2xl p-7 text-center transition-all duration-350 hover:-translate-y-2"
                style={{
                  background: 'rgba(13,13,24,0.7)',
                  border: '1px solid rgba(255,255,255,0.06)',
                  backdropFilter: 'blur(10px)',
                }}>
                {/* Hover glow */}
                <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-400 pointer-events-none"
                  style={{ boxShadow: `inset 0 0 0 1px ${c.border}, ${c.glow}` }} />

                <div className="relative z-10">
                  <div className="w-14 h-14 rounded-2xl mx-auto mb-5 flex items-center justify-center transition-transform duration-300 group-hover:scale-110"
                    style={{ background: c.bg, border: `1px solid ${c.border}` }}>
                    <card.icon className={`w-6 h-6 ${c.text}`} />
                  </div>
                  <h3 className="text-base font-bold text-white mb-3">{card.title}</h3>
                  <p className="text-sm text-white/40 leading-relaxed">{card.desc}</p>
                </div>
              </div>
            )
          })}
        </div>

        {/* Community join banner */}
        <div className="rounded-2xl p-8 flex flex-col md:flex-row items-center justify-between gap-6"
          style={{
            background: 'linear-gradient(135deg, rgba(0,212,255,0.06), rgba(168,85,247,0.04))',
            border: '1px solid rgba(0,212,255,0.12)',
          }}>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0"
              style={{ background: 'rgba(0,212,255,0.1)', border: '1px solid rgba(0,212,255,0.2)' }}>
              <MessageSquare className="w-6 h-6 text-cyan-400" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Join the conversation</h3>
              <p className="text-sm text-white/40 mt-1">
                Ask questions, share configs, and connect with other infraYS users.
              </p>
            </div>
          </div>
          <a href="/contact"
            className="btn-secondary whitespace-nowrap flex-shrink-0 flex items-center gap-2">
            <MessageSquare className="w-4 h-4" />
            Join Community
            <ArrowRight className="w-4 h-4" />
          </a>
        </div>
      </div>
    </section>
  )
}
