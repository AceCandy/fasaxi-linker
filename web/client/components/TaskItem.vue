<template>
  <div class="task-card-wrapper">
    <!-- Ribbon 类型标签 - 在卡片外层,可以延伸出去 -->
    <div class="ribbon-wrapper">
      <div class="ribbon" :class="data.type === 'main' ? 'ribbon--hardlink' : 'ribbon--sync'">
        <span class="ribbon-text">{{ data.type === 'main' ? '硬链' : '同步' }}</span>
      </div>
      <div class="ribbon-corner" :class="data.type === 'main' ? 'ribbon-corner--hardlink' : 'ribbon-corner--sync'"></div>
    </div>

    <div class="task-card" :class="{ 'task-card--watching': data.isWatching, 'task-card--error': data.watchError }">
      <!-- 头部：任务名称 + 配置 + 状态 -->
      <div class="card-header">
        <div class="task-name">{{ data.name }}</div>
        <div class="header-right">
          <span class="config-badge" @click="$emit('show-config', data.config)">
            <v-icon icon="mdi-cog-outline" size="14"></v-icon>
            {{ data.config }}
          </span>
          <div class="status-dot" :class="data.isWatching ? 'status-active' : 'status-inactive'"></div>
        </div>
      </div>

      <!-- 定时任务标签（如果有） -->
      <div v-if="data.scheduleType" class="schedule-info">
        <v-icon icon="mdi-clock-outline" size="14" color="#3b82f6"></v-icon>
        <span>{{ data.scheduleType === 'cron' ? data.scheduleValue : `每 ${data.scheduleValue}s` }}</span>
      </div>

      <!-- 路径映射列表 -->
      <div v-if="data.pathsMapping?.length" class="path-list">
        <!-- 监听失败提示 -->
        <div v-if="data.watchError" class="watch-error-info">
          <v-icon icon="mdi-alert-circle-outline" size="14" color="error"></v-icon>
          <span class="error-text">{{ data.watchError }}</span>
        </div>
        <div v-for="(mapping, idx) in data.pathsMapping" :key="idx" class="path-group">
          <div class="path-item">
            <v-icon icon="mdi-folder-outline" size="16" color="primary"></v-icon>
            <span class="path-text">{{ mapping.source }}</span>
          </div>
          <div class="path-item path-dest">
            <v-icon icon="mdi-arrow-right" size="14" color="grey"></v-icon>
            <v-icon icon="mdi-link-variant" size="16" color="success"></v-icon>
            <span class="path-text text-success">{{ mapping.dest }}</span>
          </div>
        </div>
      </div>

      <!-- 功能按钮区域 -->
      <div class="action-buttons">
        <button 
          class="action-btn" 
          :class="{ 'action-btn--active': data.isWatching }"
          @click="toggleWatch"
          :disabled="watchLoading"
        >
          <v-icon :icon="data.isWatching ? 'mdi-eye' : 'mdi-eye-outline'" size="18"></v-icon>
          <span>{{ data.isWatching ? '监控中' : '实时监控' }}</span>
          <v-progress-circular v-if="watchLoading" indeterminate size="14" width="2" class="ml-1"></v-progress-circular>
        </button>
        <button 
          class="action-btn" 
          :class="{ 'action-btn--scheduled': data.scheduleType }"
          @click="data.scheduleType ? confirmCancelSchedule() : $emit('set-schedule', data.name)"
        >
          <v-icon icon="mdi-clock-outline" size="18"></v-icon>
          <span>{{ data.scheduleType ? '取消定时' : '定时同步' }}</span>
        </button>
      </div>

      <!-- 底部操作图标栏 -->
      <div class="icon-bar">
        <button class="icon-btn" @click="$emit('edit', data.name)">
          <v-icon icon="mdi-pencil-outline" size="20" color="primary"></v-icon>
          <v-tooltip activator="parent" location="top">编辑</v-tooltip>
        </button>
        <button class="icon-btn" @click="$emit('show-log', data.name)">
          <v-icon icon="mdi-text-box-search-outline" size="20" color="info"></v-icon>
          <v-tooltip activator="parent" location="top">日志</v-tooltip>
        </button>
        <button class="icon-btn icon-btn--play" @click="$emit('play', data.name)">
          <v-icon icon="mdi-play" size="20" color="success"></v-icon>
          <v-tooltip activator="parent" location="top">单次执行</v-tooltip>
        </button>
        <button class="icon-btn" @click="$emit('show-cache', data)">
          <v-icon icon="mdi-database-outline" size="20" color="grey"></v-icon>
          <v-tooltip activator="parent" location="top">缓存管理</v-tooltip>
        </button>
        <button class="icon-btn" @click="confirmDelete">
          <v-icon icon="mdi-delete-outline" size="20" color="error"></v-icon>
          <v-tooltip activator="parent" location="top">删除</v-tooltip>
        </button>
      </div>
    </div>

    <!-- 确认对话框 -->
    <v-dialog v-if="deleteDialog" v-model="deleteDialog" max-width="320">
      <v-card class="rounded-lg">
        <v-card-title class="text-subtitle-1">确认删除</v-card-title>
        <v-card-text class="text-body-2">确认删除任务 "{{ data.name }}"?</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" size="small" @click="deleteDialog = false">取消</v-btn>
          <v-btn color="error" variant="flat" size="small" @click="handleDelete">删除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-if="cancelScheduleDialog" v-model="cancelScheduleDialog" max-width="320">
      <v-card class="rounded-lg">
        <v-card-title class="text-subtitle-1">取消定时</v-card-title>
        <v-card-text class="text-body-2">确认取消定时任务?</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" size="small" @click="cancelScheduleDialog = false">取消</v-btn>
          <v-btn color="primary" variant="flat" size="small" @click="handleCancelSchedule">确认</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { TTask } from '../../types/shim'
import { startWatch, stopWatch } from '../composables/useTask'
import copy from 'copy-to-clipboard'

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
  (e: 'show-config', config: string): void
  (e: 'show-log', name: string): void
  (e: 'show-cache', task: TTask): void
  (e: 'watch-change'): void
}>()

const deleteDialog = ref(false)
const cancelScheduleDialog = ref(false)
const watchLoading = ref(false)

// 截断路径显示
const truncatePath = (path: string) => {
  if (!path) return ''
  if (path.length <= 25) return path
  const parts = path.split('/')
  if (parts.length <= 2) return path
  return `/${parts[1]}/...${parts.slice(-1)[0]}`
}

const confirmDelete = () => {
  deleteDialog.value = true
}

const handleDelete = () => {
  emit('delete', props.data.name)
  deleteDialog.value = false
}

const confirmCancelSchedule = () => {
  cancelScheduleDialog.value = true
}

const handleCancelSchedule = () => {
  emit('cancel-schedule', props.data.name)
  cancelScheduleDialog.value = false
}

const copyScript = () => {
  const url = `${location.origin}/api/task/run?name=${encodeURIComponent(props.data.name)}&alive=0`
  if (copy(`curl ${url}`)) {
    // Use a toast or snackbar here if available, or just console
    console.log('Copied')
  }
}

const toggleWatch = async () => {
  watchLoading.value = true
  try {
    if (props.data.isWatching) {
      await stopWatch(props.data.name)
      // 本地更新状态，不刷新整个列表
      props.data.isWatching = false
      props.data.watchError = '' // 清除错误
    } else {
      await startWatch(props.data.name)
      // 本地更新状态，不刷新整个列表
      props.data.isWatching = true
      props.data.watchError = '' // 清除错误
    }
    // 通知父组件状态已变更（可选，用于其他需要）
    emit('watch-change')
  } catch (e: any) {
    // 失败时恢复状态并显示错误
    props.data.isWatching = false
    props.data.watchError = e.response?.data?.message || e.message || '未知错误'
    console.error('Watch toggle failed:', e)
  } finally {
    watchLoading.value = false
  }
}


</script>

<style scoped>
.task-card-wrapper {
  height: 100%;
  position: relative;
  padding-left: 6px;
  padding-top: 4px;
}

.task-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  position: relative;
}

.task-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
  border-color: rgba(102, 126, 234, 0.3);
}

.task-card--error {
  border: 1.5px solid #ff5252;
  box-shadow: 0 0 12px rgba(255, 82, 82, 0.2);
}

.task-card--error:hover {
  border-color: #ff5252;
  box-shadow: 0 4px 16px rgba(255, 82, 82, 0.3);
  transform: translateY(-4px);
}

.task-card:active {
  transform: translateY(-2px);
  transition: all 0.1s cubic-bezier(0.4, 0, 0.2, 1);
}

.task-card--watching {
  border-color: rgba(34, 197, 94, 0.5);
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.15);
}

/* Ribbon 容器 - 绝对定位在卡片左上角外侧 */
.ribbon-wrapper {
  position: absolute;
  top: 16px;
  left: 0;
  z-index: 10;
}

/* Ribbon 标签 - 3D丝带效果 */
.ribbon {
  position: relative;
  padding: 6px 14px 6px 10px;
  font-size: 12px;
  font-weight: 600;
  color: white;
  border-radius: 0 4px 4px 0;
  box-shadow: 
    2px 2px 6px rgba(0, 0, 0, 0.2),
    0 1px 3px rgba(0, 0, 0, 0.1);
}

.ribbon-text {
  position: relative;
  z-index: 1;
}

/* Ribbon 左下角折角 - 向内折叠的3D效果 */
.ribbon-corner {
  position: absolute;
  top: 100%;
  left: 0;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 6px 0 0 6px;
}

.ribbon--hardlink {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.ribbon-corner--hardlink {
  border-color: #1e40af transparent transparent transparent;
}

.ribbon--sync {
  background: linear-gradient(135deg, #ec4899 0%, #db2777 100%);
}

.ribbon-corner--sync {
  border-color: #be185d transparent transparent transparent;
}

/* 头部 */
.card-header {
  display: flex;
  align-items: center;
  padding: 16px 16px 12px 60px;
  border-bottom: 1px solid #f0f0f0;
  gap: 12px;
}

.task-name {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.config-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  color: #667eea;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.2s ease;
}

.config-badge:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.2) 0%, rgba(118, 75, 162, 0.2) 100%);
  border-color: rgba(102, 126, 234, 0.4);
  transform: translateY(-1px);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-active {
  background: #22c55e;
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
  animation: pulse 2s infinite;
}

.status-inactive {
  background: #d1d5db;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 定时任务信息 */
.schedule-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(59, 130, 246, 0.05);
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
  font-size: 12px;
  color: #3b82f6;
  font-weight: 500;
}

/* 路径列表 */
.path-list {
  padding: 0 16px 12px;
  flex: 1;
  overflow-y: auto;
  max-height: 150px;
  position: relative;
}

.path-group {
  padding: 8px 12px;
  background: #f8fafc;
  border-radius: 8px;
  margin-bottom: 8px;
}

.path-group:last-child {
  margin-bottom: 0;
}

.path-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.path-dest {
  margin-top: 4px;
  padding-left: 4px;
}

.path-text {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.watch-error-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  background: white; /* Ensure background to cover content behind when sticky */
  border-bottom: 1px solid rgba(255, 82, 82, 0.2);
  margin-bottom: 8px;
  position: sticky;
  top: 0;
  z-index: 5;
}

.watch-error-info::before {
  content: "";
  position: absolute;
  inset: 0;
  background: rgba(255, 82, 82, 0.08); /* Apply the red tint via pseudo-element */
  z-index: -1;
  border-radius: 4px;
}

.error-text {
  font-size: 11px;
  color: #ff5252;
  font-weight: 500;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 功能按钮区域 */
.action-buttons {
  display: flex;
  gap: 10px;
  padding: 12px 16px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 24px;
  border: 1px solid #e5e7eb;
  background: white;
  color: #667eea;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.action-btn:hover {
  background: rgba(102, 126, 234, 0.08);
  border-color: rgba(102, 126, 234, 0.4);
  transform: translateY(-1px);
}

.action-btn:active {
  transform: scale(0.96);
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.action-btn--active {
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.4);
  color: #22c55e;
}

.action-btn--active:hover {
  background: rgba(34, 197, 94, 0.2);
}

.action-btn--scheduled {
  background: rgba(251, 146, 60, 0.1);
  border-color: rgba(251, 146, 60, 0.4);
  color: #f97316;
}

/* 底部图标栏 */
.icon-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  padding: 10px 12px;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
}

.icon-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  background: rgba(0, 0, 0, 0.05);
}

.icon-btn--play {
  width: 40px;
  height: 40px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.icon-btn--play:hover {
  background: rgba(34, 197, 94, 0.2);
}
</style>
