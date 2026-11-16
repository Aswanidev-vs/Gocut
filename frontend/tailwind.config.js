/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Legacy aliases mapped to new tokens (compatibility)
        bg: 'var(--surface-app)',
        panel: 'var(--surface-panel)',
        border: 'var(--line)',
        accent: 'var(--accent)',
        'accent-hover': 'var(--accent)',
        'text-primary': 'var(--text-primary)',
        'text-secondary': 'var(--text-secondary)',

        // New semantic tokens
        'surface-well': 'var(--surface-well)',
        'surface-app': 'var(--surface-app)',
        'surface-panel': 'var(--surface-panel)',
        'surface-raised': 'var(--surface-raised)',
        'line': 'var(--line)',
        'line-strong': 'var(--line-strong)',
        'text-muted': 'var(--text-muted)',
        'signal': 'var(--signal)',
        'danger': 'var(--danger)',
        'success': 'var(--success)',

        // Control-room compatibility (retired palette absorbed)
        'cr-void': 'var(--surface-well)',
        'cr-panel': 'var(--surface-panel)',
        'cr-raise': 'var(--surface-raised)',
        'cr-line': 'var(--line)',
        'cr-line-soft': 'var(--line)',
        'ink': 'var(--text-primary)',
        'ink-dim': 'var(--text-secondary)',
        'ink-faint': 'var(--text-muted)',
        'signal-dim': 'var(--signal)',
      },
      fontFamily: {
        'dm-sans': ['IBM Plex Sans', 'sans-serif'],
        'jetbrains-mono': ['IBM Plex Mono', 'monospace'],
      },
      borderRadius: {
        DEFAULT: '4px',
        sm: '3px',
        md: '8px',
      },
    },
  },
  plugins: [],
}