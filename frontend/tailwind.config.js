/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        apple: {
          base: 'var(--apple-bg-base)',
          elevated: 'var(--apple-bg-elevated)',
          raised: 'var(--apple-bg-raised)',
          hover: 'var(--apple-bg-hover)',
          accent: 'var(--apple-accent)',
          green: 'var(--apple-green)',
          red: 'var(--apple-red)',
          orange: 'var(--apple-orange)',
          purple: 'var(--apple-purple)',
          separator: 'var(--apple-separator)',
        },
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"SF Pro Display"', '"SF Pro Text"', '"Helvetica Neue"', 'sans-serif'],
        mono: ['"SF Mono"', '"JetBrains Mono"', '"Fira Code"', '"Cascadia Code"', 'Consolas', 'monospace'],
      },
      borderRadius: {
        'apple': '12px',
      },
    },
  },
  plugins: [],
}
