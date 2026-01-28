<template>
  <v-dialog v-model="isOpen" max-width="960" scrollable class="glass-dialog">
    <v-card class="glass-content-card">
      <!-- Header -->
      <div class="glass-dialog-header">
        <div class="header-content d-flex align-center gap-2">
          <div class="header-icon-box">
            <v-icon icon="mdi-text-box-search-outline" color="primary" size="24"></v-icon>
          </div>
          <div>
            <div class="text-h6 font-weight-bold text-primary-glow font-display">任务日志</div>
            <div class="text-caption text-text-muted d-flex align-center font-mono">
              {{ taskName }}
              <span class="mx-2 opacity-50">|</span>
              <span class="text-primary">{{ total }} 条记录</span>
            </div>
          </div>
        </div>
        <v-spacer></v-spacer>
        <div class="d-flex align-center gap-2">
          <v-btn
            variant="text"
            color="grey"
            prepend-icon="mdi-refresh"
            size="small"
            @click="handleRefresh"
            :loading="loading"
            class="text-none action-btn font-mono"
            style="color: var(--color-text-muted);"
          >
            刷新
          </v-btn>
          <v-btn
            color="error"
            variant="text"
            prepend-icon="mdi-delete-outline"
            size="small"
            :disabled="logs.length === 0"
            @click="handleClearLog"
            class="text-none action-btn font-mono"
          >
            清空
          </v-btn>
          <v-divider vertical class="mx-2 h-50" style="border-color: var(--color-border);"></v-divider>
          <v-btn icon="mdi-close" variant="text" size="small" @click="isOpen = false" color="grey" style="color: var(--color-text-muted);"></v-btn>
        </div>
      </div>
      
      <!-- Filters Toolbar -->
      <div class="filter-toolbar px-6 py-3 border-b" style="background: rgba(var(--color-background-rgb), 0.3); border-color: var(--color-border); border-bottom-style: solid; border-bottom-width: 1px;">
        <div class="d-flex align-center gap-4 flex-wrap">
          <!-- File Selector -->
          <v-select
            v-model="selectedFile"
            :items="fileItems"
            item-title="label"
            item-value="name"
            variant="outlined"
            density="compact"
            hide-details
            bg-color="transparent"
            class="file-selector glass-input-field font-mono"
            style="max-width: 280px"
            @update:model-value="handleFileChange"
          >
            <template v-slot:prepend-inner>
              <v-icon size="18" class="mr-1">mdi-file-document-outline</v-icon>
            </template>
            <template v-slot:item="{ item, props }">
              <v-list-item v-bind="props">
                <template v-slot:prepend>
                  <v-icon :icon="getFileIcon(item.raw.type)" :color="getFileColor(item.raw.type)" size="18"></v-icon>
                </template>
              </v-list-item>
            </template>
          </v-select>
          
          <v-text-field
            v-model="searchText"
            prepend-inner-icon="mdi-magnify"
            placeholder="搜索关键词..."
            variant="outlined"
            density="compact"
            hide-details
            bg-color="transparent"
            class="search-input glass-input-field flex-grow-1 font-mono"
            style="max-width: 200px"
            @keydown.enter="handleSearch"
          ></v-text-field>
          
          <v-btn-toggle
            v-model="levelFilter"
            color="primary"
            variant="outlined"
            density="compact"
            mandatory
            class="filter-toggle"
            rounded="lg"
            @update:model-value="handleFilterChange"
          >
            <v-btn value="all" class="text-none font-mono">全部</v-btn>
            <v-btn value="SUCCESS" class="text-none text-success font-mono">成功</v-btn>
            <v-btn value="ERROR" class="text-none text-error font-mono">错误</v-btn>
          </v-btn-toggle>
        </div>
      </div>
      
      <!-- Log Content -->
      <div class="log-container pa-0 position-relative">
        <div
          ref="containerRef"
          class="log-content custom-scrollbar pa-4 text-body-2"
          @scroll="onScroll"
        >
          <div v-if="logs.length === 0 && !loading" class="empty-state d-flex flex-column align-center justify-center">
            <v-icon size="48" color="slate-700" class="mb-3 opacity-50">mdi-text-box-remove-outline</v-icon>
            <div class="text-slate-500 font-mono">{{ searchText ? '未找到相关日志' : '暂无日志记录' }}</div>
          </div>

          <div v-else>
            <div
              v-for="(entry, index) in logs"
              :key="index"
              class="log-line mb-1 font-jetbrains d-flex align-start"
            >
              <span class="log-time mr-3 text-no-wrap opacity-70">[{{ formatDate(entry.createdAt || '') }}]</span>
              <span :class="[getLevelClass(entry.level), 'mr-4', 'font-weight-bold', 'text-no-wrap', 'text-center']" style="min-width: 80px; display: inline-block;">
                [{{ entry.level }}]
              </span>
              <span class="log-message text-break" v-html="formatMessage(entry.message)"></span>
            </div>
          </div>

          <div v-if="loading" class="text-center py-2 text-gray-500">
             <v-progress-circular indeterminate size="20" width="2"></v-progress-circular>
             <span class="ml-2">加载中...</span>
          </div>
          <div v-else-if="!hasMore && logs.length > 0" class="text-center py-2 text-gray-600 text-xs">
             - 已加载全部日志 -
          </div>
        </div>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import fetch from '../kit/fetch'
import type { LogEntry, LogFile } from '../composables/useTask'
import { useLog, useLogFiles, clearLog } from '../composables/useTask'

const props = defineProps<{
  modelValue: boolean
  taskId?: number
  taskName?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const searchText = ref('')
const levelFilter = ref('all')
const selectedFile = ref('')
const page = ref(1)
const pageSize = 200

// Use composables
const { data: logs, execute: fetchLogs, loading, hasMore, total } = useLog(props.taskId)
const { data: logFiles, execute: fetchLogFiles } = useLogFiles(props.taskId)

const containerRef = ref<HTMLElement | null>(null)

// File items for selector
const fileItems = computed(() => {
  return logFiles.value.map(f => ({
    name: f.name,
    type: f.type,
    label: getFileLabel(f)
  }))
})

const getFileLabel = (file: LogFile) => {
  const typeLabels: Record<string, string> = {
    watch: '🔄 监控',
    run: '▶️ 单次',
    cron: '⏰ 定时'
  }
  const typePrefix = typeLabels[file.type] || '📄'
  // Extract datetime from filename
  const datePart = file.name.replace(/^(watch|run|cron)_/, '').replace('.jsonl', '')
  return `${typePrefix} ${datePart}`
}

const getFileIcon = (type: string) => {
  switch (type) {
    case 'watch': return 'mdi-sync'
    case 'run': return 'mdi-play-circle-outline'
    case 'cron': return 'mdi-clock-outline'
    default: return 'mdi-file-document-outline'
  }
}

const getFileColor = (type: string) => {
  switch (type) {
    case 'watch': return 'primary'
    case 'run': return 'success'
    case 'cron': return 'warning'
    default: return 'grey'
  }
}

// Reload when visible or taskId changes
watch(() => [props.modelValue, props.taskId], async ([visible, taskId]) => {
  if (visible && taskId) {
    page.value = 1
    searchText.value = ''
    levelFilter.value = 'all'
    
    // Load file list first
    await fetchLogFiles()
    
    // Select first file if available
    if (logFiles.value.length > 0) {
      selectedFile.value = logFiles.value[0].name
    } else {
      selectedFile.value = ''
    }
    
    // Load logs
    fetchLogs(1, pageSize, true, selectedFile.value)
  }
}, { immediate: true })

const handleRefresh = async () => {
  await fetchLogFiles()
  page.value = 1
  fetchLogs(1, pageSize, true, selectedFile.value, levelFilter.value, searchText.value)
}

const handleFileChange = () => {
  page.value = 1
  fetchLogs(1, pageSize, true, selectedFile.value, levelFilter.value, searchText.value)
}

const handleFilterChange = () => {
  page.value = 1
  fetchLogs(1, pageSize, true, selectedFile.value, levelFilter.value, searchText.value)
}

const handleSearch = () => {
  page.value = 1
  fetchLogs(1, pageSize, true, selectedFile.value, levelFilter.value, searchText.value)
}

const onScroll = (e: Event) => {
  const target = e.target as HTMLElement
  
  // Infinite scroll: load more when near bottom
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 50) {
     if (!loading.value && hasMore.value) {
       page.value++
       fetchLogs(page.value, pageSize, false, selectedFile.value, levelFilter.value, searchText.value)
     }
  }
}

// 清空日志
const handleClearLog = async () => {
  try {
    if (!props.taskId) return
    await clearLog(props.taskId, selectedFile.value)
    // Reload
    await fetchLogFiles()
    if (logFiles.value.length > 0) {
      selectedFile.value = logFiles.value[0].name
    } else {
      selectedFile.value = ''
    }
    page.value = 1
    fetchLogs(1, pageSize, true, selectedFile.value)
  } catch (e) {
    console.error('清空日志失败:', e)
  }
}

const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}/${pad(d.getMonth()+1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const getLevelClass = (level: string) => {
    const l = level.toUpperCase()
    if (l === 'SUCCESS' || l === 'SUCCEED') return 'log-primary'
    if (l === 'ERROR') return 'log-error'
    if (l === 'WARN' || l === 'WARNING') return 'log-warn text-warning'
    if (l === 'INFO') return 'log-info'
    return 'text-slate-400'
}

const formatMessage = (msg: string) => {
  if (!msg) return ''
  let cleanLine = msg
    .replace(/\u001b\[[0-9;]*m/g, '')
    .replace(/\r/g, '')
    .replace(/\t/g, '    ')

  return cleanLine
    .replace(/(→|->)/g, '<span class="log-arrow font-weight-bold mx-1">→</span>')
    .replace(/→\s*([^\s]+)/g, '→ <span class="log-path text-decoration-underline text-primary-lighten-2">$1</span>')
}
</script>

<style scoped>
.header-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(var(--color-primary-rgb), 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(var(--color-primary-rgb), 0.2);
  box-shadow: 0 0 10px rgba(var(--color-primary-rgb), 0.1);
}

.log-container {
  height: 500px;
  background: rgba(var(--color-background-rgb), 0.5);
  position: relative;
  display: flex;
  flex-direction: column;
}

.log-content {
  flex: 1;
  overflow-y: auto;
  font-family: 'JetBrains Mono', Consolas, Monaco, monospace;
  font-size: 13px;
  line-height: 1.6;
  color: var(--color-text-muted);
}

.log-line {
  word-break: break-all;
  white-space: pre-wrap;
  padding-left: 8px;
  border-left: 2px solid transparent;
  transition: all 0.1s;
}

.log-line:hover {
  background: rgba(var(--color-text-rgb), 0.05);
  border-left-color: var(--color-primary);
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(var(--color-background-rgb), 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.empty-state {
  height: 100%;
  color: var(--color-text-muted);
  background: transparent;
}

/* File selector */
.file-selector :deep(.v-field__input) {
  color: var(--color-text) !important;
  font-size: 13px;
}

/* Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
  background-color: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 4px;
  border: 2px solid transparent;
  background-clip: content-box;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--color-primary);
}

/* Log Coloring */
:deep(.log-time) { color: var(--color-text-muted); font-family: 'Space Mono', monospace; font-size: 12px; opacity: 0.7; }
:deep(.log-primary) { color: var(--color-success); font-weight: bold; }
:deep(.log-error) { color: var(--color-error); font-weight: bold; }
:deep(.log-info) { color: var(--color-secondary); font-weight: bold; }
:deep(.log-warn) { color: var(--color-warning); font-weight: bold; }
:deep(.log-arrow) { color: var(--color-text-muted); font-weight: bold; margin: 0 4px; opacity: 0.5; }
:deep(.log-path) { color: var(--color-text); text-decoration: none; border-bottom: 1px dashed var(--color-border); }

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
</style>
