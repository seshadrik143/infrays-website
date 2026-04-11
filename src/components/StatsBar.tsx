// Technical specifications — factual, measurable, verifiable
const specs = [
  { value: '~12MB',  label: 'Agent Binary Size' },
  { value: '< 30MB', label: 'Agent RAM Usage' },
  { value: '10s',    label: 'Default Scrape Interval' },
  { value: 'MIT',    label: 'License' },
  { value: '67+',    label: 'Collector Plugins' },
  { value: 'v1.0',   label: 'Current Release' },
]

export default function StatsBar() {
  return (
    <section className="py-14 px-6 border-y border-white/[0.05]"
      style={{ background: 'rgba(13, 13, 26, 0.6)' }}>
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-8">
          {specs.map((spec, i) => (
            <div key={i} className={`text-center ${i > 0 ? 'lg:border-l border-white/[0.06] lg:pl-8' : ''}`}>
              <div className="text-3xl font-black text-gradient-cyan mb-1">{spec.value}</div>
              <div className="text-xs text-white/35 font-medium">{spec.label}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
