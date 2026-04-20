import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: 'rgb(var(--background) / <alpha-value>)',
        foreground: 'rgb(var(--foreground) / <alpha-value>)',
        muted: 'rgb(var(--muted) / <alpha-value>)',
        surface: 'rgb(var(--surface) / <alpha-value>)',
        panel: 'rgb(var(--panel) / <alpha-value>)',
        border: 'rgb(var(--border) / <alpha-value>)',
        primary: 'rgb(var(--primary) / <alpha-value>)',
        secondary: 'rgb(var(--secondary) / <alpha-value>)',
        accent: 'rgb(var(--accent) / <alpha-value>)',
      },
      fontFamily: {
        sans: ['var(--font-body)'],
        display: ['var(--font-display)'],
      },
      boxShadow: {
        premium: '0 20px 50px rgba(0, 0, 0, 0.04)',
        card: '0 14px 30px rgba(15, 23, 42, 0.08)',
        glow: '0 24px 80px rgba(2, 132, 199, 0.15)',
      },
      borderRadius: {
        '2.5rem': '2.5rem',
      },
      backgroundImage: {
        'hero-radial':
          'radial-gradient(circle at top left, rgba(245, 158, 11, 0.25), transparent 36%), radial-gradient(circle at top right, rgba(34, 197, 94, 0.16), transparent 28%), linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(3, 7, 18, 1))',
      },
    },
  },
  plugins: [],
} satisfies Config;