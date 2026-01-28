export type TConfig = {
  id?: number
  name: string
  detail?: any
  configPath?: string
}

export type TTask = {
  id?: number
  name: string
  type: TTaskType
  config?: string
  configId?: number
  reverse?: boolean
  scheduleType?: 'cron' | 'loop'
  scheduleValue?: string
  pathsMapping?: { source: string, dest: string }[]
  isWatching?: boolean
  watchError?: string
  isRunning?: boolean
}

export type TSchedule = {
  taskId: number
  scheduleType: 'cron' | 'loop'
  scheduleValue: string
}

export type TTaskStatus = 'succeed' | 'failed' | 'ongoing'
export type TTaskType = 'main' | 'prune'

export type TSendData = {
  status: TTaskStatus
  type: TTaskType
  output?: string
  confirm?: boolean
}

export interface SSELog {
  send?: (data: TSendData) => void
  sendEnd?: () => void
}

declare module 'koa' {
  interface BaseContext extends SSELog {
    withSSE: true
  }
}
