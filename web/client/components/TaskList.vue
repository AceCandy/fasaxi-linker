<template>
  <div class="task-list-wrapper">
    <v-card class="glass-card task-list-card fade-in" elevation="0">
      <!-- 卡片头部 -->
      <v-card-title class="d-flex align-center py-6 px-8 dialog-header border-b border-neon">
        <div class="d-flex align-center">
          <div class="header-icon-box mr-4">
            <v-icon icon="mdi-clipboard-check-outline" size="28" class="text-primary"></v-icon>
          </div>
          <div>
            <span class="text-h5 font-weight-bold text-primary-glow font-display">任务列表</span>
            <div class="text-subtitle-2 text-slate-400 font-mono mt-1">管理您的硬链和同步任务</div>
          </div>
        </div>
        <v-spacer></v-spacer>
        <v-btn
          v-if="taskStore.tasks.length"
          class="btn-neon px-6"
          prepend-icon="mdi-plus"
          height="44"
          @click="handleCreate"
        >
          创建任务
        </v-btn>
      </v-card-title>
      
      <v-divider class="border-neon opacity-20"></v-divider>
      
      <!-- 内容区域 - 允许页面滚动 -->
      <v-card-text class="pa-8 bg-transparent">
        <!-- 加载状态 -->
        <div v-if="taskStore.loading" class="d-flex justify-center align-center pa-16">
          <v-progress-circular indeterminate color="primary" size="48" width="4"></v-progress-circular>
          <span class="ml-4 text-slate-400 font-mono">加载中...</span>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!taskStore.tasks.length" class="empty-state d-flex flex-column align-center justify-center pa-16 text-center rounded-xl border-dashed border-slate-700">
          <div class="empty-icon-container mb-6">
            <v-icon size="72" color="primary" class="opacity-50 float-animation">mdi-clipboard-text-off-outline</v-icon>
          </div>
          <div class="text-h6 text-slate-300 mb-2 font-display">暂无任务</div>
          <div class="text-body-1 text-slate-500 mb-8 font-mono">创建您的第一个硬链任务开始使用</div>
          <v-btn class="btn-neon px-8" prepend-icon="mdi-plus" height="48" @click="handleCreate">
            立即创建
          </v-btn>
        </div>

        <!-- 任务列表 -->
        <v-row v-else>
          <v-col
            cols="12"
            sm="6"
            lg="4"
            v-for="(item, index) in taskStore.tasks"
            :key="item.name"
          >
            <TaskItem
              :data="item"
              :index="index"
              @edit="handleEdit"
              @delete="handleDelete"
              @play="handlePlay"
              @set-schedule="handleSetSchedule"
              @cancel-schedule="handleCancelSchedule"
              @show-config="handleShowConfig"
              @show-log="handleShowLog"
              @show-cache="handleShowCache"
              @watch-change="() => {}"
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Modals -->
    <TaskLogViewer
      v-if="logVisible"
      v-model="logVisible"
      :task-name="currentTaskName"
    />

    <RunDetail
      v-if="runVisible"
      v-model="runVisible"
      :name="currentTaskName"
    />

    <TaskEditor
      ref="taskEditorRef"
      v-if="editVisible"
      v-model="editVisible"
      :edit="currentTaskData"
      @submit="handleTaskSubmit"
    />

    <ConfigEditor
      v-if="configEditorVisible"
      v-model="configEditorVisible"
      :data="currentConfigData"
      @submit="handleConfigSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, defineAsyncComponent, onMounted } from 'vue'
import { useDelete, useAddOrEdit, useSchedule, useCancelSchedule, useGet } from '../composables/useTask'
import { useGet as useGetConfig, useAddOrEdit as useAddOrEditConfig } from '../composables/useConfig'
import { useTaskStore } from '../stores/task'
import { useConfigStore } from '../stores/config'
import { useMessageStore } from '../stores/message'
import TaskItem from './TaskItem.vue'
const TaskLogViewer = defineAsyncComponent(() => import('./TaskLogViewer.vue'))
const RunDetail = defineAsyncComponent(() => import('./RunDetail.vue'))
const TaskEditor = defineAsyncComponent(() => import('./TaskEditor.vue'))
const ConfigEditor = defineAsyncComponent(() => import('./ConfigEditor.vue'))
import type { TTask, TConfig } from '../../types/shim'

const emit = defineEmits<{
  (e: 'show-cache', task: TTask): void
}>()

const taskStore = useTaskStore()
const configStore = useConfigStore()
const messageStore = useMessageStore()

// 每次组件创建时都获取最新数据（因为路由有key，每次切换都会重新创建组件）
onMounted(() => {
  console.log('[TaskList] 组件挂载，获取最新任务列表')
  taskStore.fetchTasks(true) // 强制刷新
})

const { rmItem } = useDelete({ 
  onSuccess: () => {
    // 本地删除，不刷新整个列表
    taskStore.removeTaskLocally(currentTaskName.value)
  } 
})
const { addOrUpdateTask } = useAddOrEdit({ 
  onSuccess: async () => { 
    // 编辑后获取最新的任务数据并本地更新
    if (currentTaskName.value) {
      // 编辑模式：获取更新后的任务数据
      await getItem(currentTaskName.value)
      if (taskData.value) {
        taskStore.upsertTaskLocally(taskData.value)
      }
    } else {
      // 新建模式：刷新列表获取新任务
      await taskStore.refreshTasks()
    }
    // 编辑器关闭由 handleTaskSubmit 控制
  } 
})
const { addScheduleTask } = useSchedule({ 
  onSuccess: () => {
    // 定时任务变更后刷新
    taskStore.refreshTasks()
  } 
})
const { cancelSchedule } = useCancelSchedule({ 
  onSuccess: () => {
    // 取消定时后刷新
    taskStore.refreshTasks()
  } 
})
const { getItem, data: taskData } = useGet()
const { getItem: getConfigItem, data: configData } = useGetConfig()
const { addOrUpdateConfig } = useAddOrEditConfig({ 
  onSuccess: () => {
    configStore.refreshConfigs()
    configEditorVisible.value = false
  } 
})

const editVisible = ref(false)
const logVisible = ref(false)
const runVisible = ref(false)
const configEditorVisible = ref(false)
const currentTaskName = ref('')
const currentTaskData = ref<TTask | undefined>(undefined)
const currentConfigData = ref<TConfig | undefined>(undefined)
const taskEditorRef = ref<{ stopSubmitting: () => void; close: () => void } | null>(null)

const handleCreate = async () => {
  // 确保配置数据已加载
  if (!configStore.initialized) {
    console.log('[TaskList] 创建任务前加载配置列表')
    await configStore.fetchConfigs()
  }
  
  if (configStore.configs?.length) {
    currentTaskName.value = ''
    currentTaskData.value = undefined
    editVisible.value = true
  } else {
    messageStore.warning('请先创建配置, 如果已有配置请刷新页面重试')
  }
}

const handleEdit = async (name: string) => {
  currentTaskName.value = name
  await getItem(name)
  console.log('[TaskList] 编辑任务:', name, '数据:', taskData.value)
  currentTaskData.value = taskData.value
  editVisible.value = true
}

const handleTaskSubmit = async (task: TTask) => {
  try {
    await addOrUpdateTask(task, currentTaskName.value || undefined)
    messageStore.success(currentTaskName.value ? '任务更新成功' : '任务创建成功')
    taskEditorRef.value?.close()
  } catch (e) {
    console.error('[TaskList] 任务保存失败:', e)
    messageStore.error((e as Error).message || '保存失败')
    taskEditorRef.value?.stopSubmitting()
  }
}

const handleDelete = async (name: string) => {
  currentTaskName.value = name
  await rmItem(name)
}

const handlePlay = (name: string) => {
  currentTaskName.value = name
  runVisible.value = true
}

const handleSetSchedule = (name: string) => {
  // 定时任务功能已移除
  messageStore.info('定时任务功能已移除')
}

const handleScheduleSubmit = async () => {
  // 定时任务功能已移除
}

const handleCancelSchedule = async (name: string) => {
  await cancelSchedule(name)
}

const handleShowLog = (name: string) => {
  currentTaskName.value = name
  logVisible.value = true
}

const handleShowCache = (task: TTask) => {
  emit('show-cache', task)
}

const handleShowConfig = async (configInfo: { id?: number; name?: string }) => {
  if (!configInfo?.id) {
    messageStore.warning('未找到配置 ID，无法查看')
    return
  }
  await getConfigItem(configInfo.id)
  console.log('[TaskList] 打开配置编辑器:', configInfo, '数据:', configData.value)
  currentConfigData.value = configData.value
  configEditorVisible.value = true
}

const handleConfigSubmit = async (config: TConfig) => {
  console.log('[TaskList] 收到配置提交:', config)
  try {
    await addOrUpdateConfig(config, currentConfigData.value?.id)
    console.log('[TaskList] 配置保存成功')
    messageStore.success('配置保存成功')
  } catch (e) {
    console.error('[TaskList] 配置保存失败:', e)
    messageStore.error((e as Error).message || '保存失败')
  }
}
</script>

<style scoped>
.task-list-card {
  min-height: calc(100vh - 100px);
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

/* Copied from ConfigList.vue for consistency */

.header-icon-box {
  width: 56px;
  height: 56px;
  background: rgba(0, 240, 255, 0.1);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.1);
  border: 1px solid rgba(0, 240, 255, 0.2);
}

.empty-state {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(0, 240, 255, 0.1);
}

.empty-icon-container {
  width: 100px;
  height: 100px;
  background: radial-gradient(circle, rgba(0, 240, 255, 0.1) 0%, transparent 70%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fade-in {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.float-animation {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
.border-neon {
    border-color: rgba(0, 240, 255, 0.2) !important;
}
.border-b {
    border-bottom-width: 1px;
    border-style: solid;
}
</style>
