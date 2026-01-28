import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'dark' | 'light'

export const useThemeStore = defineStore('theme', () => {
    // Initialize from localStorage or default to dark
    const storedTheme = localStorage.getItem('theme') as Theme | null
    const currentTheme = ref<Theme>(storedTheme || 'dark')

    const toggleTheme = () => {
        currentTheme.value = currentTheme.value === 'dark' ? 'light' : 'dark'
    }

    const setTheme = (theme: Theme) => {
        currentTheme.value = theme
    }

    // Watch for changes to update DOM and localStorage
    watch(currentTheme, (newTheme) => {
        localStorage.setItem('theme', newTheme)
        document.documentElement.setAttribute('data-theme', newTheme)

        // Update meta theme-color if needed
        const metaThemeColor = document.querySelector('meta[name="theme-color"]')
        if (metaThemeColor) {
            metaThemeColor.setAttribute('content', newTheme === 'dark' ? '#0F172A' : '#F9FAFB')
        }
    }, { immediate: true })

    return {
        currentTheme,
        toggleTheme,
        setTheme
    }
})
