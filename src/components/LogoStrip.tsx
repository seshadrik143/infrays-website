// LogoStrip — "Trusted by" social-proof row. Phase C.5 rewrite:
// removed all fictional company names (Globex / Initech / Hooli / Pied
// Piper / SoylentNet) — they're trademarked properties of major media
// companies and using them as fake customer logos creates trademark
// risk even with "illustrative" labels.
//
// Until real customers consent to logo use, this strip displays sector
// labels instead of company names. Same conversion-signal function
// ("we serve real industries") without the legal exposure.
//
// To swap in real customer logos: replace each <Sector> with an <img>
// pointing at /logos/<file>.svg (drop SVG into public/logos/) — keep
// the muted styling.

function Sector({ children }: { children: React.ReactNode }) {
  return (
    <div className="text-sm sm:text-base font-semibold text-white/40 tracking-wide whitespace-nowrap select-none uppercase">
      {children}
    </div>
  )
}

export default function LogoStrip() {
  return (
    <section className="py-12 border-y border-white/[0.05]"
      style={{ background: 'rgba(8,8,18,0.5)' }}>
      <div className="max-w-7xl mx-auto px-6">
        <p className="text-center text-[11px] font-semibold uppercase tracking-[0.2em] text-white/30 mb-7">
          Used by infrastructure teams across
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
          <Sector>Fintech</Sector>
          <Sector>B2B SaaS</Sector>
          <Sector>E-commerce</Sector>
          <Sector>AI / ML platforms</Sector>
          <Sector>Streaming media</Sector>
          <Sector>Edge / IoT fleets</Sector>
          <Sector>Bare-metal &amp; colo</Sector>
        </div>
        <p className="text-center text-[10px] text-white/15 mt-5 font-medium">
          Customer logos appear here as customers consent to being named.
        </p>
      </div>
    </section>
  )
}
