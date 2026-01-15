import { ref } from 'vue'
import fetch from '../kit/fetch'
import type { TConfig } from '../../types/shim'
import type { TAllConfig } from '@hlink/core'

type ConfigDetail = Omit<TAllConfig, 'withoutConfirm' | 'reverse'>

// 全局状态，避免重复请求
const globalConfigState = {
    data: ref<TConfig[]>([]),
    error: ref<any>(null),
    loading: ref(false),
    initialized: false,
    lastFetch: 0
}

export function useList() {
    const mutate = async (force = false) => {
        // 避免重复请求（5秒内不重复请求）
        const now = Date.now()
        if (!force && globalConfigState.initialized && (now - globalConfigState.lastFetch) < 5000) {
            return
        }

        globalConfigState.loading.value = true
        try {
            const result = await fetch.get<TConfig[]>('/api/config/list')
            globalConfigState.data.value = result || []
            globalConfigState.error.value = null
            globalConfigState.initialized = true
            globalConfigState.lastFetch = now
            console.log('[ConfigList] 数据加载成功，配置数量:', result?.length || 0)
        } catch (e) {
            globalConfigState.error.value = e
            console.error('[ConfigList] 数据加载失败:', e)
        } finally {
            globalConfigState.loading.value = false
        }
    }

    // 只在第一次调用或强制刷新时加载数据
    if (!globalConfigState.initialized && !globalConfigState.loading.value) {
        console.log('[ConfigList] 初始化数据加载')
        mutate()
    }

    return { 
        data: globalConfigState.data, 
        error: globalConfigState.error, 
        loading: globalConfigState.loading, 
        mutate 
    }
}

export function useAddOrEdit(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const addOrUpdateConfig = async (newConfig: Partial<TConfig>, currentConfigId?: number) => {
        loading.value = true
        try {
            const url = '/api/config'
            const method = currentConfigId ? fetch.put : fetch.post
            const params = currentConfigId
                ? { id: currentConfigId, ...newConfig }
                : newConfig

            console.log('[useConfig] 请求方法:', currentConfigId ? 'PUT' : 'POST')
            console.log('[useConfig] 请求参数:', params)
            
            const res = await method<boolean>(url, params as any)
            console.log('[useConfig] 服务器响应:', res)
            options?.onSuccess?.(res)
            return res
        } catch (e) {
            console.error('[useConfig] 请求失败:', e)
            options?.onError?.(e)
            throw e
        } finally {
            loading.value = false
        }
    }

    return { addOrUpdateConfig, loading }
}

export function useGet(options?: { onSuccess?: (data: TConfig) => void; onError?: (e: any) => void }) {
    const data = ref<TConfig>()
    const loading = ref(false)

    const getItem = async (id?: number) => {
        if (!id) {
            data.value = undefined
            return
        }
        loading.value = true
        try {
            const res = await fetch.get<TConfig>('/api/config', { id })
            data.value = res
            options?.onSuccess?.(res)
        } catch (e) {
            options?.onError?.(e)
        } finally {
            loading.value = false
        }
    }

    return { data, getItem, loading }
}

export function useDelete(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const rmItem = async (id: number) => {
        loading.value = true
        try {
            const res = await fetch.delete<boolean>('/api/config', { id })
            options?.onSuccess?.(res)
            return res
        } catch (e) {
            options?.onError?.(e)
            throw e
        } finally {
            loading.value = false
        }
    }

    return { rmItem, loading }
}
