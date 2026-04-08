import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      animation: {
        'pulse-slow': 'pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'gradient-shift': 'gradientShift 4s ease infinite',
        'fade-up': 'fadeUp 0.7s ease-out forwards',
      },
      boxShadow: {
        'glow-cyan': '0 0 40px rgba(0, 212, 255, 0.15)',
      },
    },
  },
  plugins: [],
}

export default config
