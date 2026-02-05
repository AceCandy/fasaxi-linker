import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const TOKEN_KEY = 'linker_auth_token'
const USERNAME_KEY = 'linker_username'

interface LoginResponse {
    token: string
    username: string
}

interface User {
    id: number
    username: string
}

export const useAuthStore = defineStore('auth', () => {
    // State
    const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
    const username = ref<string | null>(localStorage.getItem(USERNAME_KEY))
    const user = ref<User | null>(null)

    // Getters
    const isAuthenticated = computed(() => !!token.value)

    // Actions
    async function login(usernameInput: string, password: string): Promise<boolean> {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username: usernameInput, password }),
            })

            const data = await response.json()

            if (data.success && data.data) {
                const loginData = data.data as LoginResponse
                token.value = loginData.token
                username.value = loginData.username
                localStorage.setItem(TOKEN_KEY, loginData.token)
                localStorage.setItem(USERNAME_KEY, loginData.username)
                return true
            } else {
                throw new Error(data.errorMessage || '登录失败')
            }
        } catch (error) {
            console.error('Login failed:', error)
            throw error
        }
    }

    function logout() {
        token.value = null
        username.value = null
        user.value = null
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USERNAME_KEY)
    }

    async function fetchCurrentUser(): Promise<User | null> {
        if (!token.value) return null

        try {
            const response = await fetch('/api/auth/user', {
                headers: {
                    Authorization: `Bearer ${token.value}`,
                },
            })

            const data = await response.json()

            if (data.success && data.data) {
                user.value = data.data as User
                return user.value
            } else if (response.status === 401) {
                // Token expired or invalid
                logout()
                return null
            }
            return null
        } catch (error) {
            console.error('Failed to fetch user:', error)
            return null
        }
    }

    function getAuthHeaders(): Record<string, string> {
        if (!token.value) return {}
        return {
            Authorization: `Bearer ${token.value}`,
        }
    }

    // Check if token is valid on init
    async function checkAuth(): Promise<boolean> {
        if (!token.value) return false
        const user = await fetchCurrentUser()
        return !!user
    }

    async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
        if (!token.value) {
            throw new Error('未登录')
        }

        try {
            const response = await fetch('/api/auth/change-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token.value}`,
                },
                body: JSON.stringify({ oldPassword, newPassword }),
            })

            const data = await response.json()

            if (!data.success) {
                throw new Error(data.errorMessage || '修改密码失败')
            }
        } catch (error) {
            console.error('Change password failed:', error)
            throw error
        }
    }

    return {
        token,
        username,
        user,
        isAuthenticated,
        login,
        logout,
        fetchCurrentUser,
        getAuthHeaders,
        checkAuth,
        changePassword,
    }
})

