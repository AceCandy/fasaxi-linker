<template>
  <v-dialog v-model="isOpen" max-width="1000" scrollable>
    <v-card class="rounded-lg">
      <!-- Header -->
      <v-card-title class="d-flex align-center py-4 px-5 bg-grey-lighten-4">
        <v-icon icon="mdi-text-box" class="mr-2" color="primary"></v-icon>
        <div>
          <div class="text-h6">任务日志: {{ taskName }}</div>
          <div class="text-caption text-grey">共 {{ logLines.length }} 条记录</div>
        </div>
        <v-spacer></v-spacer>
        <div class="d-flex align-center gap-2">
          <v-btn
            color="error"
            variant="text"
            prepend-icon="mdi-delete"
            size="small"
            :disabled="!logLines.length"
            @click="handleClearLog"
          >
            清空日志
          </v-btn>
          <v-btn icon="mdi-refresh" variant="text" size="small" @click="fetchLog" :loading="loading" title="刷新日志"></v-btn>
          <v-btn icon="mdi-close" variant="text" size="small" @click="isOpen = false"></v-btn>
        </div>
      </v-card-title>
      
      <v-divider></v-divider>

      <!-- Filters -->
      <v-card-text class="pa-4 bg-grey-lighten-5">
        <v-text-field
          v-model="searchText"
          prepend-inner-icon="mdi-magnify"
          placeholder="搜索日志内容..."
          variant="outlined"
          density="compact"
          hide-details
          class="mb-3"
          clearable
        ></v-text-field>
        
        <v-btn-toggle
          v-model="levelFilter"
          color="primary"
          variant="outlined"
          density="compact"
          mandatory
        >
          <v-btn value="all">全部</v-btn>
          <v-btn value="SUCCESS">成功</v-btn>
          <v-btn value="ERROR">错误</v-btn>
          <v-btn value="INFO">信息</v-btn>
          <v-btn value="WARN">警告</v-btn>
        </v-btn-toggle>
      </v-card-text>
      
      <v-divider></v-divider>

      <!-- Log Content -->
      <v-card-text class="pa-4 bg-grey-darken-4 font-mono text-body-2" style="max-height: 500px; min-height: 400px;">
        <div v-if="loading" class="d-flex justify-center pa-4">
          <v-progress-circular indeterminate color="primary"></v-progress-circular>
        </div>
        
        <div v-else-if="!filteredLines.length" class="d-flex flex-column align-center justify-center h-100 text-grey">
          <v-icon size="48" class="mb-2">mdi-text-box-remove-outline</v-icon>
          <div>{{ searchText ? '未找到匹配的日志' : '该任务暂无执行日志' }}</div>
        </div>

        <div v-else>
          <div
            v-for="(line, index) in filteredLines"
            :key="index"
            class="mb-1 break-all whitespace-pre-wrap"
            v-html="formatLine(line)"
          ></div>
        </div>
      </v-card-text>
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
    .replace(/\[([\/\d\s:]+)\]/g, '<span class="text-teal-lighten-2">[$1]</span>')
    .replace(/(SUCCESS|SUCCEED)/g, '<span class="text-teal-lighten-2 font-weight-bold">$1</span>')
    .replace(/ERROR/g, '<span class="text-red-lighten-2 font-weight-bold">ERROR</span>')
    .replace(/INFO/g, '<span class="text-blue-lighten-2 font-weight-bold">INFO</span>')
    .replace(/WARN/g, '<span class="text-yellow-lighten-2 font-weight-bold">WARN</span>')
    .replace(/(→|->)/g, '<span class="text-purple-lighten-2 font-weight-bold">→</span>')
    .replace(/→\s*([^\s]+)/g, '→ <span class="text-orange-lighten-2">$1</span>')
}
</script>

<style scoped>
.font-mono {
  font-family: Consolas, Monaco, monospace;
}
.break-all {
  word-break: break-all;
}
.whitespace-pre-wrap {
  white-space: pre-wrap;
}
</style>
