// LogoStrip — "Trusted by teams at" social-proof row. These are
// friendly placeholders until real customer permissions land. Rendered
// as inline SVG wordmarks at low opacity so they read as a "trusted by"
// signal without overstating.
//
// To replace with real customer logos: drop SVG files into public/logos/,
// then swap each <Wordmark> entry below for an <img> tag pointing at
// /logos/<file>.svg. Keep the muted styling and the row layout.

import type { ComponentProps } from 'react'

// Wordmark — neutral monospace-ish glyph treatment so all 6 placeholders
// read as a coherent visual row even though they're nothing more than
// styled <span>s. Real-customer logos can replace these one-by-one.
function Wordmark({ children, className = '', ...rest }: ComponentProps<'div'>) {
  return (
    <div
      className={`text-base sm:text-lg font-bold text-white/30 tracking-tight whitespace-nowrap select-none ${className}`}
      {...rest}
    >
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
          Trusted by infrastructure teams at
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-5">
          <Wordmark>Acme<span className="text-cyan-400/50">.</span>Corp</Wordmark>
          <Wordmark>Globex Industries</Wordmark>
          <Wordmark>Initech</Wordmark>
          <Wordmark>STELLAR<span className="text-purple-400/50"> /\</span></Wordmark>
          <Wordmark>Hooli</Wordmark>
          <Wordmark>Pied Piper</Wordmark>
          <Wordmark>Soylent<span className="text-green-400/50">Net</span></Wordmark>
        </div>
        {/* Honesty note: these are placeholders. Visible in source for
            anyone curious. When real customers convert, we replace them
            one at a time and remove this note. */}
        <p className="text-center text-[10px] text-white/15 mt-5 font-medium">
          Showing illustrative logos — real customer logos replace these as they convert.
        </p>
      </div>
    </section>
  )
}
