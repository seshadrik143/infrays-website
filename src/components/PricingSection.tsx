import { Link } from 'react-router-dom'
import { Check, ArrowRight, Zap } from 'lucide-react'

const tiers = [
  {
    name: 'Free',
    price: '$0',
    period: 'forever',
    desc: 'Perfect for personal projects, homelab, and evaluation.',
    color: 'default',
    features: [
      'Up to 3 agents',
      '1M metrics / month',
      '5GB log storage',
      '1GB trace storage',
      '7-day data retention',
      'Community support',
      'All collectors included',
      'Basic alerting',
    ],
    cta: 'Start Free',
    ctaHref: '/install',
    badge: null,
  },
  {
    name: 'Starter',
    price: '$49',
    period: 'per month',
    desc: 'For small teams and growing startups that need more scale.',
    color: 'cyan',
    features: [
      'Up to 25 agents',
      '50M metrics / month',
      '50GB log storage',
      '10GB trace storage',
      '30-day data retention',
      'Email + Slack support',
      'Synthetic monitoring (10 checks)',
      'SLO tracking',
      'RBAC (3 roles)',
      'API access',
    ],
    cta: 'Get Started',
    ctaHref: '/install',
    badge: 'Most Popular',
  },
  {
    name: 'Pro',
    price: '$199',
    period: 'per month',
    desc: 'For production workloads that demand full observability.',
    color: 'purple',
    features: [
      'Unlimited agents',
      'Unlimited metrics',
      '500GB log storage',
      '100GB trace storage',
      '90-day data retention',
      'Priority support (4h SLA)',
      'Unlimited synthetic checks',
      'Continuous profiling',
      'AI/AIOps suite',
      'Multi-tenancy',
      'Custom RBAC roles',
      'SSO / OIDC',
      'Compliance reports',
    ],
    cta: 'Start Pro Trial',
    ctaHref: '/install',
    badge: null,
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    period: 'contact us',
    desc: 'Dedicated support, custom SLAs, and on-premises deployment.',
    color: 'default',
    features: [
      'Everything in Pro',
      'On-premises deployment',
      'Dedicated support (1h SLA)',
      'Custom data retention',
      'Audit logging & SIEM',
      'SSO with custom IdP',
      'mTLS agent auth',
      'Secrets manager (Vault/AWS)',
      'Volume discounts',
      'Custom integrations',
    ],
    cta: 'Contact Sales',
    ctaHref: '/contact',
    badge: null,
  },
]

const colorStyles: Record<string, { card: string; badge: string; cta: string }> = {
  default: {
    card: 'border-white/[0.07]',
    badge: '',
    cta: 'btn-secondary w-full justify-center',
  },
  cyan: {
    card: 'border-cyan-500/30 shadow-[0_0_60px_rgba(0,212,255,0.08)]',
    badge: 'badge-cyan',
    cta: 'btn-primary w-full justify-center',
  },
  purple: {
    card: 'border-purple-500/20',
    badge: 'badge-purple',
    cta: 'btn-secondary w-full justify-center border-purple-500/30 hover:border-purple-400/50',
  },
}

export default function PricingSection() {
  return (
    <section className="section" id="pricing"
      style={{ background: 'rgba(8, 8, 16, 0.5)' }}>
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-14">
          <span className="badge-green mb-4">Pricing</span>
          <h2 className="text-4xl md:text-5xl font-black tracking-tight mb-5">
            Simple,{' '}
            <span className="text-gradient-cyan">transparent</span>{' '}
            pricing
          </h2>
          <p className="text-lg text-white/40 max-w-xl mx-auto">
            No per-seat pricing. No surprise bills. Start with a 15-day free trial,
            then a license key to keep running.
          </p>
        </div>

        {/* Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-5">
          {tiers.map((tier) => {
            const s = colorStyles[tier.color]
            return (
              <div key={tier.name}
                className={`border ${s.card} rounded-2xl p-6 flex flex-col relative`}
                style={{ background: tier.color === 'cyan'
                  ? 'linear-gradient(135deg, rgba(0,212,255,0.06), rgba(13,13,26,0.9))'
                  : 'rgba(17, 17, 32, 0.7)' }}>
                {tier.badge && (
                  <div className={`absolute -top-3 left-1/2 -translate-x-1/2 ${s.badge}`}>
                    <Zap className="w-3 h-3" />
                    {tier.badge}
                  </div>
                )}

                <div className="mb-6">
                  <h3 className="text-sm font-semibold text-white/50 uppercase tracking-widest mb-3">{tier.name}</h3>
                  <div className="flex items-baseline gap-1 mb-1">
                    <span className="text-4xl font-black text-white">{tier.price}</span>
                    {tier.price !== 'Custom' && (
                      <span className="text-sm text-white/30">/{tier.period}</span>
                    )}
                  </div>
                  {tier.price === 'Custom' && (
                    <span className="text-sm text-white/30">{tier.period}</span>
                  )}
                  <p className="text-xs text-white/40 mt-2 leading-relaxed">{tier.desc}</p>
                </div>

                <ul className="space-y-2.5 flex-1 mb-8">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5 text-sm text-white/60">
                      <Check className="w-4 h-4 text-green-400 flex-shrink-0 mt-0.5" />
                      {f}
                    </li>
                  ))}
                </ul>

                <Link to={tier.ctaHref} className={s.cta}>
                  {tier.cta}
                  <ArrowRight className="w-4 h-4" />
                </Link>
              </div>
            )
          })}
        </div>

        {/* Self-host note */}
        <div className="mt-10 text-center border border-white/[0.06] rounded-2xl p-6"
          style={{ background: 'rgba(16, 185, 129, 0.04)' }}>
          <p className="text-sm text-white/50">
            <span className="text-green-400 font-semibold">Self-hosting?</span>{' '}
            infraYS is Apache 2.0 licensed. Self-hosting starts with a free 15-day trial.
            After the trial, a license key is required. Cloud pricing applies only to our managed cloud service.
          </p>
        </div>
      </div>
    </section>
  )
}
