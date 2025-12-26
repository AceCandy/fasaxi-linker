import { defineStore } from 'pinia'
import fetch from '../kit/fetch'
import type { TTask } from '../../types/shim'

export const useTaskStore = defineStore('task', {
  state: () => ({
    tasks: [] as TTask[],
    loading: false,
    error: null as any,
    initialized: false,
    lastFetch: 0
  }),
  
  actions: {
    async fetchTasks(force = false) {
      // 如果正在加载，避免重复请求
      if (this.loading) {
        console.log('[TaskStore] 正在加载中，跳过重复请求')
        return
      }

      // 如果不是强制刷新且已有数据，直接使用缓存
      if (!force && this.initialized && this.tasks.length > 0) {
        console.log('[TaskStore] 使用缓存数据，任务数:', this.tasks.length)
        return
      }

      this.loading = true
      try {
        console.log('[TaskStore] 开始加载任务列表')
        const result = await fetch.get<TTask[]>('/api/task/list')
        this.tasks = result || []
        this.error = null
        this.initialized = true
        this.lastFetch = Date.now()
        console.log('[TaskStore] 任务列表加载成功，数量:', this.tasks.length)
      } catch (e) {
        this.error = e
        console.error('[TaskStore] 任务列表加载失败:', e)
      } finally {
        this.loading = false
      }
    },
    
    refreshTasks() {
      return this.fetchTasks(true)
    },
    
    // 本地删除任务，不重新加载整个列表
    removeTaskLocally(name: string) {
      const index = this.tasks.findIndex(task => task.name === name)
      if (index !== -1) {
        this.tasks.splice(index, 1)
        console.log('[TaskStore] 本地删除任务:', name, '剩余任务数:', this.tasks.length)
      }
    },
    
    // 本地更新任务状态，不重新加载整个列表
    updateTaskLocally(name: string, updates: Partial<TTask>) {
      const task = this.tasks.find(t => t.name === name)
      if (task) {
        Object.assign(task, updates)
        console.log('[TaskStore] 本地更新任务:', name, updates)
      }
    },
    
    // 本地添加或更新任务
    upsertTaskLocally(task: TTask) {
      const index = this.tasks.findIndex(t => t.name === task.name)
      if (index !== -1) {
        this.tasks[index] = task
        console.log('[TaskStore] 本地更新任务:', task.name)
      } else {
        this.tasks.push(task)
        console.log('[TaskStore] 本地添加任务:', task.name)
      }
    }
  }
})