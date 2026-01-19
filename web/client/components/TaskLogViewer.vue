<template>
  <v-dialog v-model="isOpen" max-width="960" scrollable class="glass-dialog">
    <v-card class="glass-content-card border-neon">
      <!-- Header -->
      <div class="dialog-header border-b border-neon">
        <div class="header-icon-box">
          <v-icon icon="mdi-text-box-search-outline" color="primary" size="24"></v-icon>
        </div>
        <div>
          <div class="text-h6 font-weight-bold text-primary-glow font-display">任务日志</div>
          <div class="text-caption text-slate-400 d-flex align-center font-mono">
            {{ taskName }}
            <span class="mx-2 text-slate-600">|</span>
            <span class="text-primary">{{ logLines.length }} 条记录</span>
          </div>
        </div>
        <v-spacer></v-spacer>
        <div class="d-flex align-center gap-2">
          <v-btn
            variant="text"
            color="grey"
            prepend-icon="mdi-refresh"
            size="small"
            @click="fetchLog"
            :loading="loading"
            class="text-none action-btn font-mono"
          >
            刷新
          </v-btn>
          <v-btn
            color="error"
            variant="text"
            prepend-icon="mdi-delete-outline"
            size="small"
            :disabled="!logLines.length"
            @click="handleClearLog"
            class="text-none action-btn font-mono"
          >
            清空
          </v-btn>
          <v-divider vertical class="mx-2 h-50 border-slate-600"></v-divider>
          <v-btn icon="mdi-close" variant="text" size="small" @click="isOpen = false" color="grey"></v-btn>
        </div>
      </div>
      
      <!-- Filters Toolbar -->
      <div class="filter-toolbar px-6 py-3 bg-black/20 border-b border-slate-800">
        <div class="d-flex align-center gap-4">
          <v-text-field
            v-model="searchText"
            prepend-inner-icon="mdi-magnify"
            placeholder="搜索关键词..."
            variant="outlined"
            density="compact"
            hide-details
            bg-color="rgba(15, 23, 42, 0.5)"
            class="search-input flex-grow-1 font-mono"
            style="max-width: 300px"
          ></v-text-field>
          
          <v-btn-toggle
            v-model="levelFilter"
            color="primary"
            variant="outlined"
            density="compact"
            mandatory
            class="filter-toggle"
            rounded="lg"
          >
            <v-btn value="all" class="text-none font-mono">全部</v-btn>
            <v-btn value="SUCCESS" class="text-none text-success font-mono">成功</v-btn>
            <v-btn value="ERROR" class="text-none text-error font-mono">错误</v-btn>
          </v-btn-toggle>
        </div>
      </div>
      
      <!-- Log Content -->
      <div class="log-container pa-0 position-relative">
        <div v-if="loading" class="loading-overlay">
          <v-progress-circular indeterminate color="primary" size="40"></v-progress-circular>
        </div>
        
        <div v-if="!loading && !filteredLines.length" class="empty-state d-flex flex-column align-center justify-center">
          <v-icon size="48" color="slate-700" class="mb-3 opacity-50">mdi-text-box-remove-outline</v-icon>
          <div class="text-slate-500 font-mono">{{ searchText ? '未找到相关日志' : '暂无日志记录' }}</div>
        </div>

        <div v-else class="log-content custom-scrollbar pa-4 text-body-2" ref="logContainer">
          <div
            v-for="(line, index) in filteredLines"
            :key="index"
            class="log-line mb-1 font-jetbrains"
            v-html="formatLine(line)"
          ></div>
        </div>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import fetch from '../kit/fetch'

const props = defineProps<{
  modelValue: boolean
  taskName: string
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
const logContent = ref('')
const loading = ref(false)

// 获取日志
const fetchLog = async () => {
  if (!props.taskName) return
  loading.value = true
  try {
    logContent.value = await fetch.get<string>('/api/task/log', { name: props.taskName })
  } catch (e) {
    console.error('获取日志失败:', e)
    logContent.value = ''
  } finally {
    loading.value = false
  }
}

// 清空日志
const handleClearLog = async () => {
  try {
    await fetch.delete('/api/task/log', { name: props.taskName })
    logContent.value = ''
    searchText.value = ''
  } catch (e) {
    console.error('清空日志失败:', e)
  }
}

// 当打开抽屉或任务名变化时重新获取日志
watch(() => [props.taskName, props.modelValue], ([, newOpen]) => {
  if (newOpen && props.taskName) {
    fetchLog()
  }
}, { immediate: true })

// 日志行数组
const logLines = computed(() => {
  if (!logContent.value) return []
  return logContent.value
    .split('\n')
    .filter(line => line.trim())
    .reverse()
})

// 过滤后的日志
const filteredLines = computed(() => {
  let lines = logLines.value

  if (levelFilter.value !== 'all') {
    const searchPattern = levelFilter.value === 'SUCCESS' ? /(SUCCESS|SUCCEED)/i : new RegExp(levelFilter.value, 'i')
    lines = lines.filter(line => searchPattern.test(line))
  }

  if (searchText.value.trim()) {
    const searchLower = searchText.value.toLowerCase()
    lines = lines.filter(line => line.toLowerCase().includes(searchLower))
  }

  return lines
})

const formatLine = (line: string) => {
  let cleanLine = line
    .replace(/\u001b\[[0-9;]*m/g, '')
    .replace(/\r/g, '')
    .replace(/\t/g, '    ')
    .replace(/^\[([^\]]+)\]\s*\[([^\]]+)\]/, '[$1]')

  return cleanLine
    .replace(/\[([\/\d\s:]+)\]/g, '<span class="log-time">[$1]</span>')
    .replace(/(SUCCESS|SUCCEED)/g, '<span class="log-primary">$1</span>')
    .replace(/ERROR/g, '<span class="log-error">ERROR</span>')
    .replace(/INFO/g, '<span class="log-info">INFO</span>')
    .replace(/WARN/g, '<span class="log-warn">WARN</span>')
    .replace(/(→|->)/g, '<span class="log-arrow">→</span>')
    .replace(/→\s*([^\s]+)/g, '→ <span class="log-path">$1</span>')
}
</script>

<style scoped>
.glass-content-card {
  background: rgba(15, 23, 42, 0.98) !important;
  backdrop-filter: blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 0 20px 40px rgba(0,0,0,0.5) !important;
  overflow: hidden;
  border: 1px solid rgba(0, 240, 255, 0.1);
  color: #E0F2F7;
}

.dialog-header {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  gap: 12px;
  background: rgba(15, 23, 42, 0.9);
}

.header-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(0, 240, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 240, 255, 0.2);
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.1);
}

.log-container {
  height: 500px;
  background: #0f172a; /* Darkest slate */
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
  color: #94a3b8; /* slate-400 */
}

.log-line {
  word-break: break-all;
  white-space: pre-wrap;
  padding-left: 8px;
  border-left: 2px solid transparent;
  transition: all 0.1s;
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.05);
  border-left-color: #00F0FF;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.empty-state {
  height: 100%;
  color: #555;
  background: #0f172a;
}

/* Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
  background-color: #0f172a;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 4px;
  border: 2px solid #0f172a;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* Log Coloring */
:deep(.log-time) { color: #64748b; font-family: 'Space Mono', monospace; font-size: 12px; }
:deep(.log-primary) { color: #00F0FF; font-weight: bold; text-shadow: 0 0 5px rgba(0, 240, 255, 0.3); }
:deep(.log-error) { color: #f87171; font-weight: bold; text-shadow: 0 0 5px rgba(248, 113, 113, 0.3); }
:deep(.log-info) { color: #38bdf8; font-weight: bold; }
:deep(.log-warn) { color: #fb923c; font-weight: bold; }
:deep(.log-arrow) { color: #818cf8; font-weight: bold; margin: 0 4px; }
:deep(.log-path) { color: #e2e8f0; text-decoration: none; border-bottom: 1px dashed #475569; }

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
.border-neon {
    border-color: rgba(0, 240, 255, 0.3) !important;
}

.search-input :deep(.v-field__outline__start),
.search-input :deep(.v-field__outline__end),
.search-input :deep(.v-field__outline__notch) {
  border-color: rgba(51, 65, 85, 0.5) !important;
}

.search-input :deep(.v-field__input) {
  color: #E0F2F7 !important;
}
</style>
