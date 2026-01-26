import { ref } from 'vue'
import fetch from '../kit/fetch'
import type { TTask, TSchedule } from '../../types/shim'

export interface LogEntry {
    createdAt: string
    level: string
    message: string
}

// 全局状态，避免重复请求
const globalTaskState = {
    data: ref<TTask[]>([]),
    error: ref<any>(null),
    loading: ref(false),
    initialized: false,
    lastFetch: 0
}

export function useList() {
    const mutate = async (force = false) => {
        // 避免重复请求（5秒内不重复请求）
        const now = Date.now()
        if (!force && globalTaskState.initialized && (now - globalTaskState.lastFetch) < 5000) {
            return
        }

        globalTaskState.loading.value = true
        try {
            const result = await fetch.get<TTask[]>('/api/task/list')
            globalTaskState.data.value = result || []
            globalTaskState.error.value = null
            globalTaskState.initialized = true
            globalTaskState.lastFetch = now
            console.log('[TaskList] 数据加载成功，任务数量:', result?.length || 0)
        } catch (e) {
            globalTaskState.error.value = e
            console.error('[TaskList] 数据加载失败:', e)
        } finally {
            globalTaskState.loading.value = false
        }
    }

    // 只在第一次调用或强制刷新时加载数据
    if (!globalTaskState.initialized && !globalTaskState.loading.value) {
        console.log('[TaskList] 初始化数据加载')
        mutate()
    }

    return {
        data: globalTaskState.data,
        error: globalTaskState.error,
        loading: globalTaskState.loading,
        mutate
    }
}

export function useAddOrEdit(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const addOrUpdateTask = async (newTask: TTask, currentTask?: string) => {
        loading.value = true
        try {
            const url = '/api/task'
            const method = currentTask ? fetch.put : fetch.post
            const params = currentTask
                ? { preName: currentTask, ...newTask }
                : newTask

            const res = await method<boolean>(url, params as any)
            options?.onSuccess?.(res)
            return res
        } catch (e) {
            options?.onError?.(e)
            throw e
        } finally {
            loading.value = false
        }
    }

    return { addOrUpdateTask, loading }
}

export function useDelete(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const rmItem = async (name: string) => {
        loading.value = true
        try {
            const res = await fetch.delete<boolean>('/api/task', { name })
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

export function useGet(options?: { onSuccess?: (data: TTask) => void; onError?: (e: any) => void }) {
    const data = ref<TTask>()
    const loading = ref(false)

    const getItem = async (name?: string) => {
        if (!name) {
            data.value = undefined
            return
        }
        loading.value = true
        try {
            const res = await fetch.get<TTask>('/api/task', { name })
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

export function startWatch(name: string) {
    return fetch.post<boolean>('/api/task/watch/start', { name })
}

export function useLog(name: string | undefined) {
    const data = ref<LogEntry[]>([])
    const error = ref<any>(null)
    const loading = ref(false)
    const hasMore = ref(true)

    const execute = async (page: number = 1, pageSize: number = 200, reset: boolean = false) => {
        if (!name) return
        loading.value = true
        try {
            const res = await fetch.get<LogEntry[]>(`/api/task/log`, { name, page, pageSize })
            if (reset) {
                data.value = res
            } else {
                data.value = [...data.value, ...res]
            }
            hasMore.value = res.length === pageSize
            error.value = null
        } catch (e) {
            error.value = e
        } finally {
            loading.value = false
        }
    }

    // Initial load
    if (name) {
        execute(1, 200, true)
    }

    return { data, error, loading, hasMore, execute }
}

export function clearLog(name: string) {
    return fetch.delete<boolean>('/api/task/log', { name })
}

export function stopWatch(name: string) {
    return fetch.post<boolean>('/api/task/watch/stop', { name })
}

export function useCheckConfig() {
    const loading = ref(false)
    const check = async (name: string) => {
        loading.value = true
        try {
            await fetch.get('/api/task/check_config', { name })
        } finally {
            loading.value = false
        }
    }
    return { check, loading }
}

export function cancel(name: string) {
    return fetch.get<boolean>('/api/task/cancel', { name })
}

export function makeDeleteFile(name: string, cancel?: boolean) {
    return fetch.delete<boolean>('/api/task/files', { name, cancel })
}

export function useSchedule(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const addScheduleTask = async (scheduleTask: TSchedule) => {
        loading.value = true
        try {
            const res = await fetch.post<boolean>('/api/task/schedule', scheduleTask)
            options?.onSuccess?.(res)
            return res
        } catch (e) {
            options?.onError?.(e)
            throw e
        } finally {
            loading.value = false
        }
    }

    return { addScheduleTask, loading }
}

export function useCancelSchedule(options?: { onSuccess?: (data: boolean) => void; onError?: (e: any) => void }) {
    const loading = ref(false)

    const cancelSchedule = async (name: string) => {
        loading.value = true
        try {
            const res = await fetch.delete<boolean>('/api/task/schedule', { name })
            options?.onSuccess?.(res)
            return res
        } catch (e) {
            options?.onError?.(e)
            throw e
        } finally {
            loading.value = false
        }
    }

    return { cancelSchedule, loading }
}
