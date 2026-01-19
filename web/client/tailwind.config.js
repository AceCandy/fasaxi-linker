/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: '#00F0FF', // Cyan-Neon
                secondary: '#4D4DFF', // Electric Blue
                background: '#0F172A', // Slate 950
                surface: '#1E293B', // Slate 800 (for lighter backgrounds)
                text: '#E0F2F7', // Sky 100
                accent: '#6D28D9', // Violet 700
                success: '#10B981', // Green 500
                warning: '#FBBF24', // Amber 400
                error: '#EF4444', // Red 500
            },
            fontFamily: {
                sans: ['Space Mono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
                display: ['Orbitron', 'sans-serif'],
            },
            boxShadow: {
                'neon': '0 0 5px theme("colors.primary"), 0 0 20px theme("colors.primary")',
                'neon-strong': '0 0 10px theme("colors.primary"), 0 0 40px theme("colors.primary")',
            }
        },
    },
    plugins: [],
}
