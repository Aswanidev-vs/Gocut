/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0F0F0F',
        panel: '#1A1A1A',
        border: '#2A2A2A',
        accent: '#00D4FF',
        'accent-hover': '#00B8E0',
        'text-primary': '#E8E8E8',
        'text-secondary': '#888888',
      },
      fontFamily: {
        'dm-sans': ['DM Sans', 'sans-serif'],
        'jetbrains-mono': ['JetBrains Mono', 'monospace'],
      },
      borderRadius: {
        DEFAULT: '6px',
        sm: '3px',
      },
    },
  },
  plugins: [],
}
