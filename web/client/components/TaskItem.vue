<template>
  <div class="h-full p-2">
    <div 
      class="glass-card glass-card-interactive h-full flex flex-col group"
      :class="{ 
        'border-primary/50 shadow-neon': data.isWatching,
        'border-error/50 shadow-red-500/20': data.watchError 
      }"
    >
      <!-- 极简标签 (右上角) -->
      <div 
        class="absolute top-0 right-0 px-2 py-0.5 text-[10px] bg-t-surface backdrop-blur-sm border-b border-l border-border uppercase font-bold tracking-wider z-10"
        :class="data.type === 'main' ? 'text-primary' : 'text-accent'"
        style="border-bottom-left-radius: 6px;"
      >
        {{ data.type === 'main' ? '硬链接' : '同步' }}
      </div>

      <!-- 头部：任务名称 -->
      <div class="p-4 pb-2 relative overflow-hidden">
        <!-- 背景装饰线 -->
        <div class="absolute top-0 right-0 w-20 h-[1px] bg-gradient-to-l from-primary/50 to-transparent"></div>
        
        <div class="flex items-center justify-between mb-2">
          <div class="font-display font-bold text-lg text-primary truncate pr-14" :title="data.name">
            {{ data.name }}
          </div>
        </div>

        <!-- 状态指示灯 & 核心状态 -->
        <div class="flex items-center gap-4">
          <!-- Active/Idle Status -->
          <div class="flex items-center gap-2">
            <div 
              class="w-2 h-2 rounded-none transition-all duration-300"
              :class="data.isWatching ? 'bg-success shadow-[0_0_8px_#10B981]' : 'bg-text-muted'"
            ></div>
            <span class="text-xs font-mono font-bold" :class="data.isWatching ? 'text-success' : 'text-text-muted'">
              {{ data.isWatching ? '监听中' : '空闲' }}
            </span>
          </div>

          <!-- Schedule Status (Prominently Displayed) -->
          <div v-if="data.scheduleType" class="flex items-center gap-1.5 text-xs text-warning font-mono">
            <v-icon icon="mdi-clock-outline" size="12"></v-icon>
            <span>{{ data.scheduleType === 'cron' ? data.scheduleValue : `${data.scheduleValue}s` }}</span>
          </div>
        </div>
      </div>

      <!-- 路径列表区域 -->
      <div class="flex-1 px-4 py-2 flex flex-col min-h-[80px]">
        <!-- 监听错误提示 -->
        <div v-if="data.watchError" class="bg-error/10 border border-error/20 text-error px-2 py-1 mb-2 text-xs font-mono flex items-center">
          <v-icon icon="mdi-alert-circle" size="14" class="mr-1"></v-icon>
          <span class="truncate">{{ data.watchError }}</span>
        </div>

        <div v-if="data.pathsMapping?.length" class="overflow-y-auto max-h-[120px] custom-scrollbar pr-1">
          <div v-for="(mapping, idx) in data.pathsMapping" :key="idx" class="flex items-center gap-2 py-1.5 border-b border-border last:border-0">
            <div class="flex-1 min-w-0 flex items-center">
              <span class="text-xs text-text-muted truncate font-mono" :title="mapping.source">{{ mapping.source }}</span>
            </div>
            <v-icon icon="mdi-arrow-right-thin" size="14" class="text-primary/50 shrink-0"></v-icon>
            <div class="flex-1 min-w-0 flex items-center justify-end">
              <span class="text-xs text-text truncate font-mono" :title="mapping.dest">{{ mapping.dest }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部控制栏 - 优化布局，突出核心功能 -->
      <div class="p-3 border-t border-border space-y-3">
        
        <!-- 配置信息 -->
        <div class="flex items-center gap-2 mb-1">
          <button 
            class="flex items-center gap-1.5 px-2 py-1 bg-t-surface hover:bg-t-surface/80 border border-border hover:border-primary/50 text-[10px] text-text-muted transition-colors rounded-[2px]"
            @click="$emit('show-config', { id: data.configId, name: data.config })"
          >
            <v-icon icon="mdi-cog-outline" size="12" class="text-primary"></v-icon>
            <span class="truncate max-w-[80px]">{{ data.config || '无配置' }}</span>
          </button>
        </div>

        <!-- 显眼的操作按钮区域 -->
        <div class="flex flex-col gap-2">
          <!-- 第一行：主要功能区 (监控 + 定时) -->
          <div class="grid grid-cols-2 gap-2 h-[50px]">
            <!-- 1. 实时监听 (Watch) - Toggle -->
            <button 
              class="btn-action h-full group/btn" 
              :class="data.isWatching 
                ? 'text-primary bg-primary/10 border-primary/50 shadow-[0_0_15px_rgba(0,240,255,0.2)] hover:bg-primary/20 hover:shadow-[0_0_25px_rgba(0,240,255,0.4)] hover:border-primary' 
                : 'text-text-muted bg-t-surface border-border hover:border-primary hover:text-primary hover:bg-primary/10 hover:shadow-[0_0_15px_rgba(0,240,255,0.15)]'"
              @click="toggleWatch" 
              :title="data.isWatching ? '停止监控' : '开启监控'"
            >
              <v-icon :icon="data.isWatching ? 'mdi-eye' : 'mdi-eye-off'" size="24" class="mr-2 transition-transform duration-300 group-hover/btn:scale-110"></v-icon>
              <div class="flex flex-col items-start">
                <span class="text-xs font-bold leading-none">监控</span>
                <span class="text-[10px] opacity-60 font-mono scale-90 origin-left mt-0.5">{{ data.isWatching ? 'RUNNING' : 'STOPPED' }}</span>
              </div>
            </button>

            <!-- 2. 定时执行 (Schedule) - Toggle -->
             <button 
              class="btn-action h-full group/btn"
              :class="data.scheduleType 
                ? 'text-warning bg-warning/10 border-warning/50 shadow-[0_0_15px_rgba(251,146,60,0.2)] hover:bg-warning/20 hover:shadow-[0_0_25px_rgba(251,146,60,0.4)] hover:border-warning' 
                : 'text-text-muted bg-t-surface border-border hover:border-warning hover:text-warning hover:bg-warning/10 hover:shadow-[0_0_15px_rgba(251,146,60,0.15)]'"
              @click="data.scheduleType ? confirmCancelSchedule() : $emit('set-schedule', data)"
              :title="data.scheduleType ? '取消定时' : '设置定时'"
            >
              <v-icon icon="mdi-clock-outline" size="24" class="mr-2 transition-transform duration-300 group-hover/btn:scale-110"></v-icon>
              <div class="flex flex-col items-start">
                <span class="text-xs font-bold leading-none">定时</span>
                <span class="text-[10px] opacity-60 font-mono scale-90 origin-left mt-0.5">{{ data.scheduleType ? 'ENABLED' : 'DISABLED' }}</span>
              </div>
            </button>
          </div>

          <!-- 第二行：小工具栏 (执行、日志、编辑、缓存、删除) -->
          <div class="flex gap-1 h-[32px]">
             <!-- 执行/停止 -->
             <button 
               class="btn-micro" 
               :class="{ '!border-warning/30 !bg-warning/10': data.isRunning }"
               @click="data.isRunning ? $emit('stop', data) : $emit('play', data)" 
               :title="data.isRunning ? '停止执行' : '立即执行'"
             >
                <v-icon :icon="data.isRunning ? 'mdi-stop' : 'mdi-play'" size="16" :class="data.isRunning ? 'text-warning mr-1' : 'text-success mr-1'"></v-icon>
                <span class="text-[10px]">{{ data.isRunning ? '停止' : '执行' }}</span>
             </button>
             <!-- 日志 -->
             <button class="btn-micro" @click="$emit('show-log', data)" title="查看日志">
                <v-icon icon="mdi-text-box-search-outline" size="16" class="mr-1"></v-icon>
                <span class="text-[10px]">日志</span>
             </button>
             <!-- 编辑 -->
             <button class="btn-micro" @click="$emit('edit', data)" title="编辑任务">
                <v-icon icon="mdi-pencil-outline" size="16" class="mr-1"></v-icon>
                <span class="text-[10px]">编辑</span>
             </button>
             <!-- 缓存 -->
             <button class="btn-micro" @click="$emit('show-cache', data)" title="缓存管理">
                <v-icon icon="mdi-database-outline" size="16" class="mr-1"></v-icon>
                <span class="text-[10px]">缓存</span>
             </button>
             <!-- 删除 (红色) -->
             <button class="btn-micro !border-error/20 hover:!bg-error/20 flex-initial w-auto px-2" @click="confirmDelete" title="删除任务">
                <v-icon icon="mdi-delete-outline" size="16" class="text-error"></v-icon>
             </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 统一确认弹窗 -->
    <ConfirmDialog
      v-model="dialogState.visible"
      :title="dialogState.title"
      :content="dialogState.content"
      :type="dialogState.type"
      :loading="false"
      :confirm-text="dialogState.confirmText"
      @confirm="handleDialogConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import type { TTask } from '../../types/shim'
import { startWatch, stopWatch } from '../composables/useTask'
import ConfirmDialog from './ConfirmDialog.vue'

const props = defineProps<{
  data: TTask
  index: number
}>()

const emit = defineEmits<{
  (e: 'edit', task: TTask): void
  (e: 'delete', task: TTask): void
  (e: 'play', task: TTask): void
  (e: 'stop', task: TTask): void
  (e: 'set-schedule', task: TTask): void
  (e: 'cancel-schedule', task: TTask): void
  (e: 'show-config', config: { id?: number; name?: string }): void
  (e: 'show-log', task: TTask): void
  (e: 'show-cache', task: TTask): void
  (e: 'watch-change'): void
}>()

const watchLoading = ref(false)

// Dialog State
type DialogType = 'delete' | 'cancel_schedule'
const dialogState = reactive({
  visible: false,
  title: '',
  content: '',
  type: 'warning' as 'warning' | 'error' | 'info',
  confirmText: '确认',
  actionType: null as DialogType | null
})

const showDialog = (type: DialogType) => {
  dialogState.actionType = type
  dialogState.visible = true
  
  if (type === 'delete') {
    dialogState.title = '删除任务'
    dialogState.content = `确认删除任务 <b class="text-primary">${props.data.name}</b> 吗？`
    dialogState.type = 'error'
    dialogState.confirmText = '确认删除'
  } else if (type === 'cancel_schedule') {
    dialogState.title = '取消定时'
    dialogState.content = `确认取消任务 <b class="text-primary">${props.data.name}</b> 的定时执行吗？`
    dialogState.type = 'warning'
    dialogState.confirmText = '确认取消'
  }
}

const handleDialogConfirm = () => {
  if (dialogState.actionType === 'delete') {
    emit('delete', props.data)
  } else if (dialogState.actionType === 'cancel_schedule') {
    emit('cancel-schedule', props.data)
  }
  dialogState.visible = false
}

const confirmDelete = () => showDialog('delete')
const confirmCancelSchedule = () => showDialog('cancel_schedule')

const toggleWatch = async () => {
  watchLoading.value = true
  try {
    const taskId = props.data.id
    if (!taskId) {
      props.data.watchError = '缺少任务ID'
      return
    }
    if (props.data.isWatching) {
      await stopWatch(taskId)
      props.data.isWatching = false
      props.data.watchError = ''
    } else {
      await startWatch(taskId)
      props.data.isWatching = true
      props.data.watchError = ''
    }
    emit('watch-change')
  } catch (e: any) {
    props.data.isWatching = false
    props.data.watchError = e.response?.data?.message || e.message || '未知错误'
  } finally {
    watchLoading.value = false
  }
}
</script>

<style scoped>
.btn-action {
  @apply w-full flex items-center justify-center rounded-lg border transition-all duration-300 px-4;
  backdrop-filter: blur(4px);
}

.btn-micro {
  @apply flex-1 w-full flex items-center justify-center rounded-[4px] border transition-all duration-200;
  background: rgba(var(--color-surface-rgb), 0.5);
  border-color: var(--color-border);
  color: var(--color-text-muted);
  min-height: 24px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.btn-micro:hover {
  background: rgba(var(--color-surface-rgb), 0.8);
  color: var(--color-primary);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-neon);
}

.glass-dropdown {
  background: var(--glass-bg-strong) !important;
  backdrop-filter: blur(12px) !important;
  border: 1px solid var(--color-border) !important;
  border-radius: 4px !important;
  box-shadow: var(--shadow-glass) !important;
}

:deep(.v-list-item) {
  color: var(--color-text) !important;
}

:deep(.v-list-item:hover) {
  background: rgba(var(--color-primary-rgb), 0.1) !important;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.1);
  border-radius: 2px;
}
.custom-scrollbar:hover::-webkit-scrollbar-thumb {
  background: rgba(0,240,255,0.3);
}
</style>
