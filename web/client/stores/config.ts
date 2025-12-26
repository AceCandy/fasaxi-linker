import { defineStore } from 'pinia'
import fetch from '../kit/fetch'
import type { TConfig } from '../../types/shim'

export const useConfigStore = defineStore('config', {
  state: () => ({
    configs: [] as TConfig[],
    loading: false,
    error: null as any,
    initialized: false,
    lastFetch: 0
  }),
  
  actions: {
    async fetchConfigs(force = false) {
      // 如果正在加载，避免重复请求
      if (this.loading) {
        console.log('[ConfigStore] 正在加载中，跳过重复请求')
        return
      }

      // 如果不是强制刷新且已有数据，直接使用缓存
      if (!force && this.initialized && this.configs.length > 0) {
        console.log('[ConfigStore] 使用缓存数据，配置数:', this.configs.length)
        return
      }

      this.loading = true
      try {
        console.log('[ConfigStore] 开始加载配置列表')
        const result = await fetch.get<TConfig[]>('/api/config/list')
        this.configs = result || []
        this.error = null
        this.initialized = true
        this.lastFetch = Date.now()
        console.log('[ConfigStore] 配置列表加载成功，数量:', this.configs.length)
      } catch (e) {
        this.error = e
        console.error('[ConfigStore] 配置列表加载失败:', e)
      } finally {
        this.loading = false
      }
    },
    
    refreshConfigs() {
      return this.fetchConfigs(true)
    }
  }
})