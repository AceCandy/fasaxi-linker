/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: 'rgb(var(--color-primary-rgb) / <alpha-value>)',
                secondary: 'rgb(var(--color-secondary-rgb) / <alpha-value>)',
                background: 'rgb(var(--color-background-rgb) / <alpha-value>)',
                't-surface': 'rgb(var(--color-surface-rgb) / <alpha-value>)',
                text: 'rgb(var(--color-text-rgb) / <alpha-value>)',
                accent: 'rgb(var(--color-accent-rgb) / <alpha-value>)',
                success: 'var(--color-success)',
                warning: 'var(--color-warning)',
                error: 'var(--color-error)',
                border: 'rgb(var(--color-border-rgb) / <alpha-value>)',
            },
            fontFamily: {
                sans: ['Space Mono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
                display: ['Orbitron', 'sans-serif'],
            },
            boxShadow: {
                'neon': 'var(--shadow-neon)',
                'neon-strong': 'var(--shadow-neon-strong)',
            }
        },
    },
    plugins: [],
}
