<template>
  <div class="task-card-wrapper">
    <div class="task-card" :class="{ 'task-card--watching': data.isWatching, 'task-card--error': data.watchError }">
      <!-- 极简标签 (右上角) -->
      <div class="card-badge" :class="data.type === 'main' ? 'card-badge--hardlink' : 'card-badge--sync'">
        {{ data.type === 'main' ? '硬链' : '同步' }}
      </div>

      <!-- 头部：任务名称 -->
      <div class="card-header">
        <div class="task-name" :title="data.name">{{ data.name }}</div>
        
        <!-- 状态指示灯 -->
        <div class="status-indicator" :class="{ 'status-indicator--active': data.isWatching }">
          <div class="status-dot"></div>
          <span class="status-text">{{ data.isWatching ? 'Running' : 'Stopped' }}</span>
        </div>
      </div>

      <!-- 路径列表区域 -->
      <div class="path-section">
        <!-- 监听错误提示 -->
        <div v-if="data.watchError" class="error-banner">
          <v-icon icon="mdi-alert-circle" size="16" class="mr-2"></v-icon>
          <span class="text-truncate">{{ data.watchError }}</span>
        </div>

        <div v-if="data.pathsMapping?.length" class="path-list custom-scrollbar">
          <div v-for="(mapping, idx) in data.pathsMapping" :key="idx" class="path-row">
            <div class="path-source">
              <v-icon icon="mdi-folder-outline" size="14" class="path-icon"></v-icon>
              <span class="path-text">{{ mapping.source }}</span>
            </div>
            <div class="path-arrow">
              <v-icon icon="mdi-arrow-right-thin" size="16" color="grey-lighten-1"></v-icon>
            </div>
            <div class="path-dest">
              <v-icon icon="mdi-link-variant" size="14" class="path-icon"></v-icon>
              <span class="path-text">{{ mapping.dest }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部控制栏 -->
      <div class="card-footer">
        <!-- 主操作按钮 -->
        <button 
          class="param-chip"
          @click="$emit('show-config', { id: data.configId, name: data.config })"
        >
          <v-icon icon="mdi-cog-outline" size="14"></v-icon>
          <span>{{ data.config || '未配置' }}</span>
        </button>

        <button 
          v-if="data.scheduleType"
          class="param-chip param-chip--schedule"
          @click="confirmCancelSchedule"
        >
          <v-icon icon="mdi-clock-outline" size="14"></v-icon>
          <span>{{ data.scheduleType === 'cron' ? data.scheduleValue : `${data.scheduleValue}s` }}</span>
        </button>

        <v-spacer></v-spacer>

        <!-- 悬浮操作栏 -->
        <div class="action-group">
          <div class="divider-vertical"></div>
          
          <v-btn icon size="small" variant="text" color="grey-darken-1" @click="$emit('show-log', data.name)">
            <v-icon>mdi-text-box-search-outline</v-icon>
            <v-tooltip activator="parent" location="top">日志</v-tooltip>
          </v-btn>

          <v-btn icon size="small" variant="text" color="grey-darken-1" @click="$emit('edit', data.name)">
            <v-icon>mdi-pencil-outline</v-icon>
            <v-tooltip activator="parent" location="top">编辑</v-tooltip>
          </v-btn>

          <v-menu location="top end">
            <template v-slot:activator="{ props }">
              <v-btn icon size="small" variant="text" color="grey-darken-1" v-bind="props">
                <v-icon>mdi-dots-horizontal</v-icon>
              </v-btn>
            </template>
            <v-list class="menu-popup" density="compact">
              <v-list-item @click="$emit('play', data.name)" prepend-icon="mdi-play" class="text-success">
                <v-list-item-title>立即执行</v-list-item-title>
              </v-list-item>
              <v-list-item @click="toggleWatch" :prepend-icon="data.isWatching ? 'mdi-eye-off' : 'mdi-eye'">
                <v-list-item-title>{{ data.isWatching ? '停止监控' : '开启监控' }}</v-list-item-title>
              </v-list-item>
              <v-list-item @click="data.scheduleType ? confirmCancelSchedule() : $emit('set-schedule', data.name)" prepend-icon="mdi-clock-outline">
                <v-list-item-title>{{ data.scheduleType ? '取消定时' : '定时执行' }}</v-list-item-title>
              </v-list-item>
              <v-divider class="my-1"></v-divider>
              <v-list-item @click="$emit('show-cache', data)" prepend-icon="mdi-database-outline">
                <v-list-item-title>缓存管理</v-list-item-title>
              </v-list-item>
              <v-list-item @click="confirmDelete" prepend-icon="mdi-delete-outline" class="text-error">
                <v-list-item-title>删除任务</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
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
  (e: 'edit', name: string): void
  (e: 'delete', name: string): void
  (e: 'play', name: string): void
  (e: 'set-schedule', name: string): void
  (e: 'cancel-schedule', name: string): void
  (e: 'show-config', config: { id?: number; name?: string }): void
  (e: 'show-log', name: string): void
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
    dialogState.title = '确认删除'
    dialogState.content = `确认删除任务 <b>${props.data.name}</b> 吗？`
    dialogState.type = 'error'
    dialogState.confirmText = '删除'
  } else if (type === 'cancel_schedule') {
    dialogState.title = '取消定时'
    dialogState.content = `确认取消任务 <b>${props.data.name}</b> 的定时执行吗？`
    dialogState.type = 'warning'
    dialogState.confirmText = '取消定时'
  }
}

const handleDialogConfirm = () => {
  if (dialogState.actionType === 'delete') {
    emit('delete', props.data.name)
  } else if (dialogState.actionType === 'cancel_schedule') {
    emit('cancel-schedule', props.data.name)
  }
  dialogState.visible = false
}

const confirmDelete = () => showDialog('delete')
const confirmCancelSchedule = () => showDialog('cancel_schedule')

const toggleWatch = async () => {
  watchLoading.value = true
  try {
    if (props.data.isWatching) {
      await stopWatch(props.data.name)
      props.data.isWatching = false
      props.data.watchError = ''
    } else {
      await startWatch(props.data.name)
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
.task-card-wrapper {
  height: 100%;
  padding: 6px;
}

.task-card {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    0 4px 6px -1px rgba(0, 0, 0, 0.02),
    0 10px 15px -3px rgba(0, 0, 0, 0.04);
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.task-card:hover {
  transform: translateY(-4px);
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 
    0 20px 25px -5px rgba(0, 0, 0, 0.05),
    0 8px 10px -6px rgba(0, 0, 0, 0.01);
  border-color: rgba(255, 255, 255, 0.9);
}

.task-card--watching {
  border: 1px solid rgba(84, 160, 255, 0.3);
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.9) 0%, rgba(240, 248, 255, 0.6) 100%);
}

.task-card--error {
  border: 1px solid rgba(255, 107, 107, 0.3);
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 245, 245, 0.6) 100%);
}

/* Badge */
.card-badge {
  position: absolute;
  top: 16px;
  right: 16px;
  font-size: 10px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 20px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-badge--hardlink {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  color: #0050b3;
}

.card-badge--sync {
  background: linear-gradient(135deg, #fff0f6 0%, #ffadd2 100%);
  color: #c41d7f;
}

/* Header */
.card-header {
  padding: 24px 24px 16px;
}

.task-name {
  font-size: 16px;
  font-weight: 700;
  color: #2d3436;
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-right: 50px;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.04);
  border-radius: 6px;
  transition: all 0.3s ease;
}

.status-indicator--active {
  background: rgba(34, 197, 94, 0.1);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #b2bec3;
  transition: all 0.3s ease;
}

.status-indicator--active .status-dot {
  background: #22c55e;
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.6);
}

.status-text {
  font-size: 11px;
  font-weight: 600;
  color: #636e72;
}

.status-indicator--active .status-text {
  color: #166534;
}

/* Path Section */
.path-section {
  flex: 1;
  padding: 0 24px;
  min-height: 100px;
  display: flex;
  flex-direction: column;
}

.path-list {
  flex: 1;
  overflow-y: auto;
  max-height: 160px;
}

.path-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px dashed rgba(0,0,0,0.05);
}

.path-row:last-child {
  border-bottom: none;
}

.path-icon {
  opacity: 0.5;
  margin-right: 4px;
}

.path-source, .path-dest {
  flex: 1;
  display: flex;
  align-items: center;
  min-width: 0;
}

.path-source .path-text { color: #57606f; }
.path-dest .path-text { color: #2f3542; font-weight: 500; }

.path-text {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.error-banner {
  background: #ffecec;
  color: #d63031;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 500;
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

/* Footer & Actions */
.card-footer {
  padding: 16px 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-top: 1px solid rgba(0,0,0,0.03);
}

.param-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  background: rgba(0,0,0,0.03);
  border-radius: 8px;
  font-size: 11px;
  font-weight: 500;
  color: #636e72;
  transition: all 0.2s ease;
}

.param-chip:hover {
  background: rgba(108, 92, 231, 0.08);
  color: #6c5ce7;
}

.param-chip--schedule {
  background: rgba(253, 203, 110, 0.15);
  color: #e17055;
}

.action-group {
  display: flex;
  align-items: center;
}

.divider-vertical {
  width: 1px;
  height: 16px;
  background: rgba(0,0,0,0.1);
  margin: 0 4px;
}

/* Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.1);
  border-radius: 2px;
}
.custom-scrollbar:hover::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.2);
}

.menu-popup {
  border-radius: 12px !important;
  box-shadow: 0 10px 30px -5px rgba(0,0,0,0.15) !important;
  border: 1px solid rgba(0,0,0,0.05);
  overflow: hidden;
}
</style>
