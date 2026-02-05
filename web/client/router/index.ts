import { createRouter, createWebHistory } from 'vue-router'

import DefaultLayout from '../layouts/DefaultLayout.vue'
import TaskListPage from '../pages/TaskListPage.vue'
import ConfigListPage from '../pages/ConfigListPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
    {
        path: '/login',
        name: 'login',
        component: LoginPage,
        meta: { requiresAuth: false }
    },
    {
        path: '/',
        component: DefaultLayout,
        meta: { requiresAuth: true },
        children: [
            { path: '', name: 'home', component: TaskListPage },
            { path: 'config', name: 'config', component: ConfigListPage },
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// Navigation guard for authentication
router.beforeEach((to, _from, next) => {
    const authStore = useAuthStore()

    // Check if the route requires authentication
    const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

    if (requiresAuth && !authStore.isAuthenticated) {
        // Not authenticated, redirect to login
        next({ name: 'login' })
    } else if (to.name === 'login' && authStore.isAuthenticated) {
        // Already authenticated, redirect to home
        next({ name: 'home' })
    } else {
        next()
    }
})

export default router

