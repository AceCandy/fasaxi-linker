<template>
  <v-dialog v-model="isOpen" max-width="960" scrollable class="glass-dialog">
    <v-card class="glass-content-card">
      <!-- Header -->
      <div class="dialog-header">
        <div class="header-icon-box">
          <v-icon icon="mdi-text-box-search-outline" color="info" size="24"></v-icon>
        </div>
        <div>
          <div class="text-h6 font-weight-bold text-grey-darken-3">任务日志</div>
          <div class="text-caption text-grey-darken-1 d-flex align-center">
            {{ taskName }}
            <span class="mx-2 text-grey-lighten-2">|</span>
            <span class="text-info">{{ logLines.length }} 条记录</span>
          </div>
        </div>
        <v-spacer></v-spacer>
        <div class="d-flex align-center gap-2">
          <v-btn
            variant="text"
            color="grey-darken-1"
            prepend-icon="mdi-refresh"
            size="small"
            @click="fetchLog"
            :loading="loading"
            class="text-none action-btn"
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
            class="text-none action-btn"
          >
            清空
          </v-btn>
          <v-divider vertical class="mx-2 h-50"></v-divider>
          <v-btn icon="mdi-close" variant="text" size="small" @click="isOpen = false" color="grey"></v-btn>
        </div>
      </div>
      
      <v-divider class="border-opacity-50"></v-divider>

      <!-- Filters Toolbar -->
      <div class="filter-toolbar px-6 py-3 bg-grey-lighten-5">
        <div class="d-flex align-center gap-4">
          <v-text-field
            v-model="searchText"
            prepend-inner-icon="mdi-magnify"
            placeholder="搜索关键词..."
            variant="outlined"
            density="compact"
            hide-details
            bg-color="white"
            class="search-input flex-grow-1"
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
            <v-btn value="all" class="text-none">全部</v-btn>
            <v-btn value="SUCCESS" class="text-none text-success">成功</v-btn>
            <v-btn value="ERROR" class="text-none text-error">错误</v-btn>
          </v-btn-toggle>
        </div>
      </div>
      
      <v-divider class="border-opacity-50"></v-divider>

      <!-- Log Content -->
      <div class="log-container pa-0 position-relative">
        <div v-if="loading" class="loading-overlay">
          <v-progress-circular indeterminate color="primary" size="40"></v-progress-circular>
        </div>
        
        <div v-if="!loading && !filteredLines.length" class="empty-state d-flex flex-column align-center justify-center">
          <v-icon size="48" color="grey-darken-3" class="mb-3 opacity-20">mdi-text-box-remove-outline</v-icon>
          <div class="text-grey-darken-1">{{ searchText ? '未找到相关日志' : '暂无日志记录' }}</div>
        </div>

        <div v-else class="log-content custom-scrollbar pa-4 text-body-2" ref="logContainer">
          <div
            v-for="(line, index) in filteredLines"
            :key="index"
            class="log-line mb-1"
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
    .replace(/(SUCCESS|SUCCEED)/g, '<span class="log-success">$1</span>')
    .replace(/ERROR/g, '<span class="log-error">ERROR</span>')
    .replace(/INFO/g, '<span class="log-info">INFO</span>')
    .replace(/WARN/g, '<span class="log-warn">WARN</span>')
    .replace(/(→|->)/g, '<span class="log-arrow">→</span>')
    .replace(/→\s*([^\s]+)/g, '→ <span class="log-path">$1</span>')
}
</script>

<style scoped>
.glass-content-card {
  background: rgba(255, 255, 255, 0.98) !important;
  backdrop-filter: blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1) !important;
  overflow: hidden;
  border: 1px solid rgba(0,0,0,0.05);
}

.dialog-header {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  gap: 12px;
  background: white;
}

.header-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(59, 130, 246, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.log-container {
  height: 500px;
  background: #1e1e1e; /* VS Code style dark bg */
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
  color: #d4d4d4;
}

.log-line {
  word-break: break-all;
  white-space: pre-wrap;
  padding-left: 8px;
  border-left: 2px solid transparent;
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.05);
  border-left-color: #667eea;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(30, 30, 30, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.empty-state {
  height: 100%;
  color: #555;
  background: #1e1e1e;
}

/* Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 10px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: #1e1e1e;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 5px;
  border: 2px solid #1e1e1e;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #555;
}

/* Log Coloring using global styles or deep selectors */
:deep(.log-time) { color: #569cd6; opacity: 0.8; }
:deep(.log-success) { color: #4ec9b0; font-weight: bold; }
:deep(.log-error) { color: #f44747; font-weight: bold; }
:deep(.log-info) { color: #9cdcfe; font-weight: bold; }
:deep(.log-warn) { color: #ce9178; font-weight: bold; }
:deep(.log-arrow) { color: #c586c0; font-weight: bold; }
:deep(.log-path) { color: #ce9178; text-decoration: underline; text-underline-offset: 2px; }

.action-btn {
  font-weight: 500;
  letter-spacing: 0.5px;
}
</style>
