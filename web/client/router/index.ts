import { createRouter, createWebHistory } from 'vue-router'

import DefaultLayout from '../layouts/DefaultLayout.vue'
import TaskListPage from '../pages/TaskListPage.vue'
import ConfigListPage from '../pages/ConfigListPage.vue'

const routes = [
    {
        path: '/',
        component: DefaultLayout,
        children: [
            { path: '', component: TaskListPage },
            { path: 'config', component: ConfigListPage },
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router
