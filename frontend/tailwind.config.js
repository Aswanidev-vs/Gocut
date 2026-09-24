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

        // Control-room palette used by the Design (compositing) workspace.
        // Near-black void, hairline rules, and a single amber "signal" accent
        // reserved for numbers and animated state; cyan stays the selection
        // colour so the two never compete.
        'cr-void': '#08080A',
        'cr-panel': '#0E0E11',
        'cr-raise': '#15151A',
        'cr-line': '#232329',
        'cr-line-soft': '#1A1A1F',
        signal: '#FFB020',
        'signal-dim': '#5C4210',
        ink: '#EDEDF0',
        'ink-dim': '#8A8A94',
        'ink-faint': '#54545C',
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
