<template>
  <v-dialog
    v-model="isOpen"
    max-width="900"
    persistent
    class="glass-dialog"
  >
    <v-card class="glass-content-card">
      <!-- 优化后的标题栏 -->
      <div class="glass-dialog-header">
        <div class="header-icon-box" :class="`bg-${statusIconColor}-lighten-5`">
          <v-icon :color="statusIconColor" size="24">{{ statusIcon }}</v-icon>
        </div>
        <div>
          <div class="text-h6 font-weight-bold text-grey-darken-3">{{ taskName }}</div>
          <div class="text-caption" :class="`text-${statusIconColor}`" style="opacity: 0.9;">{{ statusText }}</div>
        </div>
        <v-spacer></v-spacer>
        <v-btn 
          icon="mdi-close" 
          variant="text" 
          density="comfortable" 
          color="grey"
          @click="handleClose" 
          :disabled="loading"
        ></v-btn>
      </div>

      <v-divider class="border-opacity-50"></v-divider>

      <!-- 优化后的内容区域 -->
      <v-card-text class="pa-0">
        <RunPanel :data="logs" />
      </v-card-text>

      <v-divider class="border-opacity-50"></v-divider>

      <!-- 优化后的操作按钮区 -->
      <v-card-actions class="pa-5 bg-grey-lighten-5">
        <v-spacer></v-spacer>
        <v-btn
          v-if="showCancel"
          color="error"
          variant="text"
          :loading="cancelLoading"
          @click="handleCancel"
          prepend-icon="mdi-close-circle"
          class="action-btn"
        >
          取消
        </v-btn>
        <v-btn
          v-if="showConfirm"
          color="error"
          variant="flat"
          :loading="confirmLoading"
          @click="handleConfirm"
          prepend-icon="mdi-delete-clock"
          class="action-btn elevation-2"
        >
          确认删除
        </v-btn>
        <v-btn
          v-if="!loading && !showConfirm"
          color="primary"
          variant="flat"
          @click="handleClose"
          prepend-icon="mdi-check"
          class="action-btn elevation-2 px-6"
        >
          完成
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import RunPanel from './RunPanel.vue'
import { useCheckConfig, cancel, makeDeleteFile } from '../composables/useTask'
import { useMessageStore } from '../stores/message'
import type { TSendData, TTaskStatus, TTaskType } from '../../types/shim'

const props = defineProps<{
  modelValue: boolean
  taskId?: number
  taskName?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
}>()

const messageStore = useMessageStore()

const isOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const logs = ref<string[]>([])
const status = ref<TTaskStatus>('ongoing')
const type = ref<TTaskType>('main')
const needConfirm = ref(false)
const eventSource = ref<EventSource | null>(null)
const cancelLoading = ref(false)
const confirmLoading = ref(false)

const loading = computed(() => status.value === 'ongoing')

const statusText = computed(() => {
  const action = type.value === 'main' ? '执行' : '分析'
  const state = {
    succeed: '完成',
    failed: '出错',
    ongoing: '中...',
  }[status.value]
  return action + state
})

const statusIcon = computed(() => {
  return {
    succeed: 'mdi-check-circle',
    failed: 'mdi-alert-circle',
    ongoing: 'mdi-loading mdi-spin',
  }[status.value]
})

const statusIconColor = computed(() => {
  return {
    succeed: 'success',
    failed: 'error',
    ongoing: 'info',
  }[status.value]
})

const showCancel = computed(() => {
  if (loading.value) return true
  if (type.value === 'prune' && needConfirm.value) return true
  return false
})

const showConfirm = computed(() => {
  return type.value === 'prune' && !loading.value && needConfirm.value
})

const { check } = useCheckConfig()

const startTask = () => {
  if (!props.taskId) return
  
  logs.value = []
  status.value = 'ongoing'
  
  // 直接开始执行任务，不检查配置
  const taskId = props.taskId
  console.log('[RunDetail] 创建 EventSource 连接:', `/api/task/run?taskId=${encodeURIComponent(String(taskId))}`)
  
  if (window.EventSource) {
    const es = new window.EventSource(`/api/task/run?taskId=${encodeURIComponent(String(taskId))}`)
    eventSource.value = es
    
    es.onmessage = (event) => {
      const result = JSON.parse(event.data) as TSendData
      const { output, status: newStatus, type: newType, confirm } = result
      
      console.log('[RunDetail] 收到消息:', result)
      
      if (output) {
        logs.value.push(output)
      }
      
      status.value = newStatus
      type.value = newType
      if (confirm !== undefined) {
        needConfirm.value = confirm
      }
      
      if (newStatus !== 'ongoing') {
        console.log('[RunDetail] 任务完成，状态:', newStatus)
        es.close()
        eventSource.value = null
      }
    }
    
    es.onerror = (e) => {
      console.error('[RunDetail] EventSource error', e)
      es.close()
      eventSource.value = null
      status.value = 'failed'
      logs.value.push('连接错误，任务执行失败')
    }
    
    es.onopen = () => {
      console.log('[RunDetail] EventSource 连接已建立')
    }
  } else {
    console.error("Sorry, server logs aren't supported on this browser :(")
    messageStore.warning("您的浏览器不支持服务器日志功能 (EventSource)")
    status.value = 'failed'
  }
}

const handleClose = () => {
  if (eventSource.value) {
    eventSource.value.close()
    eventSource.value = null
  }
  isOpen.value = false
  emit('close')
}

const handleCancel = async () => {
  if (!props.taskId) return
  cancelLoading.value = true
  
  try {
    if (type.value === 'prune' && !loading.value) {
      // Cancel delete confirmation
      await makeDeleteFile(props.taskId, true)
      handleClose()
    } else {
      // Cancel ongoing task
      await cancel(props.taskId)
    }
  } catch (e) {
    console.error(e)
  } finally {
    cancelLoading.value = false
  }
}

const handleConfirm = async () => {
  if (!props.taskId) return
  confirmLoading.value = true
  try {
    await makeDeleteFile(props.taskId)
    handleClose()
  } catch (e) {
    console.error(e)
  } finally {
    confirmLoading.value = false
  }
}

watch(() => props.modelValue, (val) => {
  console.log('[RunDetail] modelValue 变化:', val, 'taskId:', props.taskId)
  if (val && props.taskId) {
    console.log('[RunDetail] 开始执行任务:', props.taskName)
    startTask()
  } else {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
  }
}, { immediate: true })

// 也监听 props.taskId 的变化
watch(() => props.taskId, (newTaskId) => {
  console.log('[RunDetail] taskId 变化:', newTaskId, 'modelValue:', props.modelValue)
  if (props.modelValue && newTaskId) {
    console.log('[RunDetail] taskId 变化后开始执行任务:', props.taskName)
    startTask()
  }
})

onUnmounted(() => {
  if (eventSource.value) {
    eventSource.value.close()
  }
})
</script>

<style scoped>
.header-icon-box {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--color-primary-rgb), 0.1);
  box-shadow: 0 0 15px rgba(var(--color-primary-rgb), 0.1);
  border: 1px solid rgba(var(--color-primary-rgb), 0.2);
}

.bg-success-lighten-5 { background-color: rgba(16, 185, 129, 0.1) !important; border-color: rgba(16, 185, 129, 0.3) !important; box-shadow: 0 0 10px rgba(16, 185, 129, 0.2) !important; }
.bg-error-lighten-5 { background-color: rgba(239, 68, 68, 0.1) !important; border-color: rgba(239, 68, 68, 0.3) !important; box-shadow: 0 0 10px rgba(239, 68, 68, 0.2) !important; }
.bg-info-lighten-5 { background-color: rgba(0, 240, 255, 0.1) !important; border-color: rgba(0, 240, 255, 0.3) !important; }

.action-btn {
  font-weight: 600;
  letter-spacing: 0.5px;
  font-family: 'Space Mono', monospace;
}

:deep(.v-card-actions) {
  background: rgba(var(--color-surface-rgb), 0.8) !important;
  border-top: 1px solid var(--color-border);
}

.text-grey-darken-3 {
  color: var(--color-text) !important;
  font-family: 'Orbitron', sans-serif;
  letter-spacing: 1px;
}
</style>
