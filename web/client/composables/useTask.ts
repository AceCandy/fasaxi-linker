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

    const addOrUpdateTask = async (newTask: TTask, currentTaskId?: number) => {
        loading.value = true
        try {
            const url = '/api/task'
            const method = currentTaskId ? fetch.put : fetch.post
            const params = currentTaskId
                ? { taskId: currentTaskId, ...newTask }
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

    const rmItem = async (taskId: number) => {
        loading.value = true
        try {
            const res = await fetch.delete<boolean>('/api/task', { taskId })
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

    const getItem = async (taskId?: number) => {
        if (!taskId) {
            data.value = undefined
            return
        }
        loading.value = true
        try {
            const res = await fetch.get<TTask>('/api/task', { taskId })
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

export function startWatch(taskId: number) {
    return fetch.post<boolean>('/api/task/watch/start', { taskId })
}

export interface LogFile {
    name: string
    type: 'watch' | 'run' | 'cron' | 'unknown'
    size: number
    modTime: string
}

export interface LogResponse {
    list: LogEntry[]
    total: number
    file: string
}

export function useLogFiles(taskId: number | undefined) {
    const data = ref<LogFile[]>([])
    const loading = ref(false)

    const execute = async () => {
        if (!taskId) return
        loading.value = true
        try {
            const res = await fetch.get<LogFile[]>('/api/task/log/files', { taskId })
            data.value = res || []
        } catch (e) {
            console.error('Failed to load log files:', e)
            data.value = []
        } finally {
            loading.value = false
        }
    }

    return { data, loading, execute }
}

export function useLog(taskId: number | undefined) {
    const data = ref<LogEntry[]>([])
    const error = ref<any>(null)
    const loading = ref(false)
    const hasMore = ref(true)
    const total = ref(0)
    const currentFile = ref('')

    const execute = async (page: number = 1, pageSize: number = 200, reset: boolean = false, file?: string, level?: string, search?: string) => {
        if (!taskId) return
        loading.value = true
        try {
            const params: Record<string, any> = { taskId, page, pageSize }
            if (file) params.file = file
            if (level && level !== 'all') params.level = level
            if (search) params.search = search

            const res = await fetch.get<LogResponse>('/api/task/log', params)
            if (reset) {
                data.value = res.list || []
            } else {
                data.value = [...data.value, ...(res.list || [])]
            }
            total.value = res.total || 0
            currentFile.value = res.file || ''
            hasMore.value = (res.list?.length || 0) === pageSize
            error.value = null
        } catch (e) {
            error.value = e
        } finally {
            loading.value = false
        }
    }

    // Initial load
    if (taskId) {
        execute(1, 200, true)
    }

    return { data, error, loading, hasMore, total, currentFile, execute }
}

export function clearLog(taskId: number, file?: string) {
    const params: Record<string, any> = { taskId }
    if (file) params.file = file
    return fetch.delete<boolean>('/api/task/log', params)
}

export function stopWatch(taskId: number) {
    return fetch.post<boolean>('/api/task/watch/stop', { taskId })
}

export function useCheckConfig() {
    const loading = ref(false)
    const check = async (taskId: number) => {
        loading.value = true
        try {
            await fetch.get('/api/task/check_config', { taskId })
        } finally {
            loading.value = false
        }
    }
    return { check, loading }
}

export function cancel(taskId: number) {
    return fetch.get<boolean>('/api/task/cancel', { taskId })
}

export function makeDeleteFile(taskId: number, cancel?: boolean) {
    return fetch.delete<boolean>('/api/task/files', { taskId, cancel })
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

    const cancelSchedule = async (taskId: number) => {
        loading.value = true
        try {
            const res = await fetch.delete<boolean>('/api/task/schedule', { taskId })
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
