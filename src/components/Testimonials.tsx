import { MessageSquare, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function Testimonials() {
  return (
    <section className="section" id="testimonials">
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-cyan mb-4">Community</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Built for teams.{' '}
            <span className="text-gradient-cyan">Loved by engineers.</span>
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto">
            infraYS is open source and actively growing. Be among the first teams to deploy it
            and share your experience with the community.
          </p>
        </div>

        {/* Early adopter CTA */}
        <div className="grid md:grid-cols-3 gap-5 mb-12">
          {[
            {
              icon: '🚀',
              title: 'Early Adopter',
              desc: 'Deploy infraYS in your infrastructure today and be part of shaping the roadmap. Your feedback drives the next release.',
              color: 'cyan',
            },
            {
              icon: '⭐',
              title: 'Share Your Experience',
              desc: 'Once you\'ve used infraYS in production, we\'d love to feature your story. Real engineers, real results.',
              color: 'purple',
            },
            {
              icon: '🤝',
              title: 'Open Source Community',
              desc: 'Contribute collectors, plugins, dashboard templates, or bug fixes. Every contribution makes the platform better for everyone.',
              color: 'green',
            },
          ].map((card) => (
            <div key={card.title}
              className="border border-white/[0.07] rounded-2xl p-6 text-center"
              style={{ background: 'rgba(17, 17, 32, 0.6)' }}>
              <div className="text-4xl mb-4">{card.icon}</div>
              <h3 className="text-base font-bold text-white mb-3">{card.title}</h3>
              <p className="text-sm text-white/40 leading-relaxed">{card.desc}</p>
            </div>
          ))}
        </div>

        {/* Community links */}
        <div className="border border-white/[0.07] rounded-2xl p-8 flex flex-col md:flex-row items-center justify-between gap-6"
          style={{ background: 'linear-gradient(135deg, rgba(0,212,255,0.05), rgba(139,92,246,0.04))' }}>
          <div className="flex items-center gap-4">
            <MessageSquare className="w-8 h-8 text-cyan-400 flex-shrink-0" />
            <div>
              <h3 className="text-base font-bold text-white">Join the conversation</h3>
              <p className="text-sm text-white/40 mt-1">
                Ask questions, share configs, and connect with other infraYS users on GitHub Discussions.
              </p>
            </div>
          </div>
          <a href="https://github.com/seshadrik143/infrays" target="_blank" rel="noopener noreferrer"
            className="btn-secondary whitespace-nowrap flex-shrink-0">
            GitHub Discussions
            <ArrowRight className="w-4 h-4" />
          </a>
        </div>
      </div>
    </section>
  )
}
