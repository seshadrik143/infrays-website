const specs = [
  { value: '~12MB',  label: 'Agent Binary Size',   color: 'text-cyan-400',   glow: 'rgba(0,212,255,0.2)' },
  { value: '< 30MB', label: 'Agent RAM Usage',      color: 'text-purple-400', glow: 'rgba(168,85,247,0.2)' },
  { value: '10s',    label: 'Default Scrape Interval', color: 'text-green-400', glow: 'rgba(16,185,129,0.2)' },
  { value: 'MIT',    label: 'License',              color: 'text-orange-400', glow: 'rgba(245,158,11,0.2)' },
  { value: '67+',    label: 'Collector Plugins',    color: 'text-pink-400',   glow: 'rgba(236,72,153,0.2)' },
  { value: 'v1.0',   label: 'Current Release',      color: 'text-cyan-400',   glow: 'rgba(0,212,255,0.2)' },
]

export default function StatsBar() {
  return (
    <section className="py-16 px-6 border-y border-white/[0.05] relative overflow-hidden"
      style={{ background: 'rgba(8,8,18,0.7)' }}>
      {/* Subtle shimmer layer */}
      <div className="absolute inset-0 animate-shimmer pointer-events-none" />

      <div className="max-w-7xl mx-auto relative z-10">
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6 lg:gap-0">
          {specs.map((spec, i) => (
            <div key={i}
              className={`text-center group ${
                i > 0 ? 'lg:border-l border-white/[0.06]' : ''
              } px-4`}>
              <div
                className={`text-3xl lg:text-4xl font-black mb-1 transition-all duration-300 ${spec.color}`}
                style={{ textShadow: `0 0 20px ${spec.glow}` }}>
                {spec.value}
              </div>
              <div className="text-xs text-white/30 font-medium tracking-wide">{spec.label}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
