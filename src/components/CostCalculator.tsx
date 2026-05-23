import { useState, useMemo } from 'react'
import { Calculator, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

// CostCalculator — back-of-envelope "what would I save vs Datadog"
// estimator. Numbers based on Datadog's public pricing (Infrastructure
// Pro $15/host/mo + APM $40/host/mo + Logs $1.27/GB ingested) so we
// can defend the math, with conservative averages that won't be
// embarrassing if a hostile prospect verifies.
//
// infraYS pricing is from tiers.ts in PricingSection. We pick the
// tier whose host-count window the user lands in.

function infraysCostMonthly(hosts: number): { tier: string; price: number } {
  if (hosts <= 3) return { tier: 'Free', price: 0 }
  if (hosts <= 25) return { tier: 'Starter (annual)', price: 39 }
  // Pro starts at 25+ hosts and is "unlimited" by tier rules. We use
  // the annual-billed monthly rate ($159) for both lower + higher
  // host counts, since Pro doesn't tier by host count.
  return { tier: 'Pro (annual)', price: 159 }
}

// Datadog public list pricing as of mid-2025 (in USD). These are the
// "Pro" SKUs at the standard mid-tier — Datadog Enterprise is more
// expensive but most prospects compare against Pro. Update when their
// pricing page changes.
const DATADOG_INFRA_PER_HOST    = 15   // Infrastructure Pro
const DATADOG_APM_PER_HOST      = 40   // APM (avg)
const DATADOG_LOGS_PER_GB_MONTH = 1.27 // Log ingestion + 15 day retention

export default function CostCalculator() {
  const [hosts, setHosts] = useState(25)
  const [logsGBPerMonth, setLogsGBPerMonth] = useState(100)

  const calc = useMemo(() => {
    const datadog =
      hosts * (DATADOG_INFRA_PER_HOST + DATADOG_APM_PER_HOST) +
      logsGBPerMonth * DATADOG_LOGS_PER_GB_MONTH
    const infraysTier = infraysCostMonthly(hosts)
    const savings = datadog - infraysTier.price
    const savingsPct = datadog > 0 ? Math.round((savings / datadog) * 100) : 0
    return {
      datadog,
      infrays: infraysTier.price,
      infraysTier: infraysTier.tier,
      savings,
      savingsPct,
    }
  }, [hosts, logsGBPerMonth])

  return (
    <section className="section py-16 border-y border-white/[0.06]"
      style={{ background: 'rgba(8,8,16,0.5)' }}>
      <div className="container-md">
        <div className="text-center mb-10">
          <span className="badge-cyan mb-4">Cost calculator</span>
          <h2 className="text-3xl md:text-4xl font-black tracking-tight mb-4">
            How much would you save vs{' '}
            <span className="text-gradient-cyan">Datadog</span>?
          </h2>
          <p className="text-white/40 text-sm max-w-xl mx-auto">
            Move the sliders. Math is from Datadog&apos;s public pricing
            (Infra Pro $15 + APM $40 per host, $1.27/GB logs).
            Self-serve infraYS pricing is the annual-effective monthly rate.
          </p>
        </div>

        <div className="grid lg:grid-cols-2 gap-6">
          {/* Inputs */}
          <div className="border border-white/10 rounded-2xl p-6"
            style={{ background: 'rgba(17,17,32,0.7)' }}>
            <div className="flex items-center gap-3 mb-6">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-cyan-500/10 border border-cyan-500/20">
                <Calculator className="w-4 h-4 text-cyan-400" />
              </div>
              <h3 className="font-bold text-white">Your usage</h3>
            </div>

            <div className="space-y-6">
              <div>
                <div className="flex justify-between items-baseline mb-2">
                  <label className="text-sm text-white/60">Hosts (servers + containers)</label>
                  <span className="text-lg font-bold text-cyan-400 tabular-nums">{hosts}</span>
                </div>
                <input
                  type="range" min={1} max={500} step={1}
                  value={hosts}
                  onChange={(e) => setHosts(parseInt(e.target.value, 10))}
                  className="w-full accent-cyan-400"
                />
                <div className="flex justify-between text-[10px] text-white/30 mt-1">
                  <span>1</span><span>250</span><span>500+</span>
                </div>
              </div>

              <div>
                <div className="flex justify-between items-baseline mb-2">
                  <label className="text-sm text-white/60">Log volume / month (GB)</label>
                  <span className="text-lg font-bold text-cyan-400 tabular-nums">{logsGBPerMonth} GB</span>
                </div>
                <input
                  type="range" min={0} max={2000} step={10}
                  value={logsGBPerMonth}
                  onChange={(e) => setLogsGBPerMonth(parseInt(e.target.value, 10))}
                  className="w-full accent-cyan-400"
                />
                <div className="flex justify-between text-[10px] text-white/30 mt-1">
                  <span>0</span><span>1 TB</span><span>2 TB</span>
                </div>
              </div>
            </div>
          </div>

          {/* Output */}
          <div className="border border-cyan-500/30 rounded-2xl p-6 relative overflow-hidden"
            style={{ background: 'linear-gradient(135deg, rgba(0,212,255,0.04), rgba(13,13,26,0.9))' }}>
            <h3 className="font-bold text-white mb-6">Monthly cost comparison</h3>

            <div className="space-y-4">
              <div className="flex items-baseline justify-between pb-3 border-b border-white/[0.06]">
                <span className="text-sm text-white/50">Datadog</span>
                <span className="text-2xl font-black text-white/80 tabular-nums">
                  ${calc.datadog.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                </span>
              </div>

              <div className="flex items-baseline justify-between pb-3 border-b border-white/[0.06]">
                <div>
                  <div className="text-sm text-cyan-400 font-semibold">infraYS</div>
                  <div className="text-[10px] text-white/30 uppercase tracking-wider mt-0.5">{calc.infraysTier}</div>
                </div>
                <span className="text-2xl font-black text-cyan-400 tabular-nums">
                  ${calc.infrays.toLocaleString()}
                </span>
              </div>

              <div className="flex items-baseline justify-between pt-2">
                <span className="text-sm font-semibold text-white">You save</span>
                <div className="text-right">
                  <div className="text-3xl font-black text-green-400 tabular-nums">
                    ${calc.savings.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                  </div>
                  <div className="text-xs text-green-400/80 mt-0.5">/ month · {calc.savingsPct}% off</div>
                </div>
              </div>

              <div className="text-xs text-white/30 pt-3 border-t border-white/[0.06]">
                Annual savings: <strong className="text-white/60">${(calc.savings * 12).toLocaleString(undefined, { maximumFractionDigits: 0 })}/year</strong>
              </div>
            </div>

            <Link to="/install" className="btn-primary w-full justify-center mt-6">
              Start free trial
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>

        <p className="text-[10px] text-white/25 text-center mt-6 max-w-2xl mx-auto">
          Disclaimer: Datadog pricing is from their public price sheet; effective
          discounted enterprise pricing will be lower. Actual savings depend on
          your config (retention, custom metrics, sampling). Numbers are
          indicative, not a contractual quote.
        </p>
      </div>
    </section>
  )
}
