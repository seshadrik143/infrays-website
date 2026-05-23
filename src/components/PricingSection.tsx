import { Link } from 'react-router-dom'
import { useState } from 'react'
import { Check, ArrowRight, Zap, Key, Calendar } from 'lucide-react'

// PORTAL_SIGNUP — the customer self-service licensing portal. Used by
// paid-tier CTAs so customers can pay + receive an enrollment token
// without going through the contact form.
const PORTAL_SIGNUP = 'https://license.infrays.org/signup'

// Pricing is annual-effective-monthly when annual is selected:
//   monthly Starter:  $49/mo
//   annual  Starter:  $39/mo billed annually ($468/yr) — 20% discount
// Pro is similarly 20% off when annual.
type Tier = {
  name: string
  monthly: number | null     // null = Custom / Free indicators
  annual: number | null      // null = same as monthly
  freeLabel?: string         // overrides monthly/annual for Free
  customLabel?: string       // overrides for Enterprise
  desc: string
  color: 'default' | 'cyan' | 'purple'
  features: string[]
  cta: string
  ctaHref: string
  badge: string | null
}

const tiers: Tier[] = [
  {
    name: 'Free',
    monthly: 0,
    annual: 0,
    freeLabel: '$0',
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
    monthly: 49,
    annual: 39, // billed annually = 39 * 12 = $468/yr ≈ -20% off $588
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
    ctaHref: PORTAL_SIGNUP,
    badge: 'Most Popular',
  },
  {
    name: 'Pro',
    monthly: 199,
    annual: 159, // -20%
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
    ctaHref: PORTAL_SIGNUP,
    badge: null,
  },
  {
    name: 'Enterprise',
    monthly: null,
    annual: null,
    customLabel: 'Custom',
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
    cta: 'Book a Demo',
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
  // Monthly / Annual toggle. Defaults to annual since that's the
  // conversion-optimized default (20% cheaper effective rate; most
  // SaaS leans this way out of the box).
  const [billing, setBilling] = useState<'monthly' | 'annual'>('annual')

  return (
    <section className="section" id="pricing"
      style={{ background: 'rgba(8, 8, 16, 0.5)' }}>
      <div className="container-lg">
        {/* Header */}
        <div className="text-center mb-10">
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

        {/* Billing toggle — pill switcher centered above tier cards.
            Annual is the default + carries a savings badge to anchor
            visitors on the lower effective price. */}
        <div className="flex justify-center mb-10">
          <div className="inline-flex items-center gap-1 p-1 rounded-full border border-white/10"
            style={{ background: 'rgba(255,255,255,0.03)' }}>
            <button
              onClick={() => setBilling('monthly')}
              className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all ${
                billing === 'monthly'
                  ? 'bg-white/10 text-white shadow-sm'
                  : 'text-white/40 hover:text-white/70'
              }`}
            >
              Monthly
            </button>
            <button
              onClick={() => setBilling('annual')}
              className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all flex items-center gap-2 ${
                billing === 'annual'
                  ? 'bg-white/10 text-white shadow-sm'
                  : 'text-white/40 hover:text-white/70'
              }`}
            >
              Annual
              <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-green-500/20 text-green-400 font-semibold">
                Save 20%
              </span>
            </button>
          </div>
        </div>

        {/* Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-5">
          {tiers.map((tier) => {
            const s = colorStyles[tier.color]
            const showAnnual = billing === 'annual'
            const price = tier.freeLabel ?? tier.customLabel ?? (showAnnual && tier.annual !== null
              ? `$${tier.annual}`
              : tier.monthly !== null ? `$${tier.monthly}` : '—')
            const periodLabel = tier.freeLabel
              ? 'forever'
              : tier.customLabel
                ? 'contact us'
                : showAnnual
                  ? 'per month, billed annually'
                  : 'per month'

            return (
              <div key={tier.name}
                className={`border ${s.card} rounded-2xl p-6 pt-7 flex flex-col relative`}
                style={{ background: tier.color === 'cyan'
                  ? 'linear-gradient(135deg, rgba(0,212,255,0.06), rgba(13,13,26,0.9))'
                  : 'rgba(17, 17, 32, 0.7)' }}>
                {/* Badge — added whitespace-nowrap + z-10 + max-width
                    fits-content so it doesn't get clipped or wrapped on
                    narrow mobile widths where the card collapses to ~280px. */}
                {tier.badge && (
                  <div className={`absolute -top-3 left-1/2 -translate-x-1/2 whitespace-nowrap z-10 ${s.badge}`}>
                    <Zap className="w-3 h-3" />
                    {tier.badge}
                  </div>
                )}

                <div className="mb-6">
                  <h3 className="text-sm font-semibold text-white/50 uppercase tracking-widest mb-3">{tier.name}</h3>
                  <div className="flex items-baseline gap-1 mb-1">
                    <span className="text-4xl font-black text-white">{price}</span>
                    {!tier.customLabel && (
                      <span className="text-xs text-white/30">/{showAnnual ? 'mo' : 'mo'}</span>
                    )}
                  </div>
                  <span className="text-xs text-white/30">{periodLabel}</span>
                  {/* When annual is selected on a paid tier, show the
                      crossed-out monthly price so visitors see the
                      discount viscerally. */}
                  {showAnnual && tier.monthly !== null && tier.annual !== null && tier.monthly > 0 && (
                    <div className="text-xs text-white/30 mt-1">
                      <span className="line-through">${tier.monthly}/mo</span>{' '}
                      <span className="text-green-400">save ${(tier.monthly - tier.annual) * 12}/yr</span>
                    </div>
                  )}
                  <p className="text-xs text-white/40 mt-3 leading-relaxed">{tier.desc}</p>
                </div>

                <ul className="space-y-2.5 flex-1 mb-8">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5 text-sm text-white/60">
                      <Check className="w-4 h-4 text-green-400 flex-shrink-0 mt-0.5" />
                      {f}
                    </li>
                  ))}
                </ul>

                {tier.ctaHref.startsWith('http') ? (
                  <a href={tier.ctaHref} className={s.cta}>
                    {tier.cta}
                    {tier.cta === 'Book a Demo' ? <Calendar className="w-4 h-4" /> : <ArrowRight className="w-4 h-4" />}
                  </a>
                ) : (
                  <Link to={tier.ctaHref} className={s.cta}>
                    {tier.cta}
                    {tier.cta === 'Book a Demo' ? <Calendar className="w-4 h-4" /> : <ArrowRight className="w-4 h-4" />}
                  </Link>
                )}
              </div>
            )
          })}
        </div>

        {/* Self-host note */}
        <div className="mt-10 border border-white/[0.06] rounded-2xl p-6"
          style={{ background: 'rgba(16, 185, 129, 0.04)' }}>
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <p className="text-sm text-white/50 text-center sm:text-left">
              <span className="text-green-400 font-semibold">Self-hosting?</span>{' '}
              infraYS is Apache 2.0 licensed. Self-hosting starts with a free 15-day trial.
              After the trial, a license key is required. Cloud pricing applies only to our managed cloud service.
            </p>
            <div className="flex gap-2 flex-shrink-0">
              <a href={PORTAL_SIGNUP}
                className="btn-primary text-sm whitespace-nowrap flex items-center gap-2">
                <Key className="w-4 h-4" />
                Get License Key
              </a>
              <Link to="/contact"
                className="btn-secondary text-sm whitespace-nowrap">
                Talk to Sales
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
