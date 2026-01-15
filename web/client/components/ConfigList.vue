<template>
  <div>
    <v-card class="glass-card config-list-card fade-in" elevation="0">
      <!-- 卡片头部 -->
      <v-card-title class="d-flex align-center py-6 px-8 dialog-header">
        <div class="d-flex align-center">
          <div class="header-icon-box mr-4">
            <v-icon icon="mdi-cog-outline" size="28" class="gradient-text"></v-icon>
          </div>
          <div>
            <span class="text-h5 font-weight-bold text-grey-darken-3">配置管理</span>
            <div class="text-subtitle-2 text-grey pt-1">管理您的硬链配置规则</div>
          </div>
        </div>
        <v-spacer></v-spacer>
        <v-btn
          v-if="configStore.configs?.length"
          class="gradient-btn rounded-pill px-6"
          prepend-icon="mdi-plus"
          height="44"
          @click="handleCreate"
        >
          创建配置
        </v-btn>
      </v-card-title>
      
      <v-divider class="border-opacity-50"></v-divider>

      <v-card-text class="pa-8">
        <!-- 空状态 -->
        <div v-if="!configStore.configs?.length" class="empty-state d-flex flex-column align-center justify-center pa-16 text-center rounded-xl border-dashed">
          <div class="empty-icon-container mb-6">
            <v-icon size="64" color="grey-lighten-2">mdi-cog-off-outline</v-icon>
          </div>
          <div class="text-h6 text-grey-darken-1 mb-2">暂无配置</div>
          <div class="text-body-1 text-grey mb-8">创建一个配置来定义硬链规则</div>
          <v-btn class="gradient-btn rounded-pill px-8" prepend-icon="mdi-plus" height="48" @click="handleCreate">
            立即创建
          </v-btn>
        </div>

        <!-- 配置列表 -->
        <v-row v-else>
          <v-col
            v-for="item in configStore.configs"
            :key="item.id"
            cols="12"
            sm="6"
            md="4"
            lg="3"
            xl="3"
          >
            <ConfigCard 
              :data="item"
              @save="handleInlineSave"
              @delete="confirmDelete"
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Create Dialog only -->
    <ConfigEditor
      v-if="editorVisible"
      v-model="editorVisible"
      @submit="handleEditorSubmit"
    />

    <!-- 统一确认弹窗 -->
    <ConfirmDialog
      v-model="dialogState.visible"
      :title="dialogState.title"
      :content="dialogState.content"
      :type="dialogState.type"
      :loading="dialogState.loading"
      :confirm-text="dialogState.confirmText"
      @confirm="handleDialogConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, defineAsyncComponent, onMounted, reactive } from 'vue'
import { useGet, useAddOrEdit, useDelete } from '../composables/useConfig'
import { useConfigStore } from '../stores/config'
import { useTaskStore } from '../stores/task'
import { useMessageStore } from '../stores/message'
import ConfigCard from './ConfigCard.vue'
import ConfirmDialog from './ConfirmDialog.vue'
const ConfigEditor = defineAsyncComponent(() => import('./ConfigEditor.vue'))
import type { TConfig } from '../../types/shim'
import fetch from '../kit/fetch'

const configStore = useConfigStore()
const taskStore = useTaskStore()
const messageStore = useMessageStore()

// 每次组件创建时都获取最新数据
onMounted(() => {
  console.log('[ConfigList] 组件挂载，获取最新配置列表')
  configStore.fetchConfigs(true) // 强制刷新
})

const { addOrUpdateConfig } = useAddOrEdit({ onSuccess: () => configStore.refreshConfigs() })
const { rmItem, loading: deleteLoading } = useDelete({ onSuccess: () => {
  configStore.refreshConfigs()
  closeDialog()
}})

const editorVisible = ref(false)

// Dialog State Management
type DialogType = 'delete' | 'save_warn'
const dialogState = reactive({
  visible: false,
  title: '',
  content: '',
  type: 'warning' as 'warning' | 'error' | 'info',
  loading: false,
  confirmText: '确认',
  actionType: null as DialogType | null,
  payload: null as any
})

const closeDialog = () => {
  dialogState.visible = false
  dialogState.loading = false
  setTimeout(() => {
    dialogState.actionType = null
    dialogState.payload = null
  }, 300)
}

const showDialog = (type: DialogType, payload: any) => {
  dialogState.actionType = type
  dialogState.payload = payload
  dialogState.loading = false
  dialogState.visible = true
  
  if (type === 'delete') {
    dialogState.title = '确认删除'
    dialogState.content = `确认删除配置 <b>${payload.name}</b> 吗？<br><span class="text-caption text-error">此操作不可恢复</span>`
    dialogState.type = 'error'
    dialogState.confirmText = '删除'
  } else if (type === 'save_warn') {
    dialogState.title = '确认修改'
    dialogState.content = `该配置正在被以下任务使用：<br><b>${payload.taskList}</b><br><br>修改配置可能影响这些任务的运行，确认保存吗？`
    dialogState.type = 'warning'
    dialogState.confirmText = '确认修改'
  }
}

const handleDialogConfirm = async () => {
  if (!dialogState.actionType) return
  
  if (dialogState.actionType === 'delete') {
    dialogState.loading = true
    await rmItem(dialogState.payload.id)
    // rmItem success callback closes dialog
  } else if (dialogState.actionType === 'save_warn') {
    // Proceed with save
    dialogState.visible = false
    await performSave(dialogState.payload.config)
  }
}

// Create always opens clean dialog
const handleCreate = () => {
  editorVisible.value = true
}

const performSave = async (config: TConfig) => {
   try {
    await addOrUpdateConfig(config, config.id)
    messageStore.success('配置已保存')
  } catch (e) {
    console.error('[ConfigList] 保存失败:', e)
    messageStore.error((e as Error).message || '保存失败')
  }
}

// Inline save from ConfigCard
const handleInlineSave = async (config: TConfig) => {
  try {
    console.log('[ConfigList] 保存配置:', config)

    // Check for related tasks before saving
    if (config.id) {
      const relatedTasks = await fetch.get<string[]>('/api/config/related-tasks', { id: config.id })
      if (relatedTasks && relatedTasks.length > 0) {
        const taskList = relatedTasks.join('、')
        // Show Dialog
        showDialog('save_warn', { taskList, config })
        return
      }
    }

    // Direct save if no conflicts
    await performSave(config)
  } catch (e) {
    console.error('[ConfigList] Pre-save check failed:', e)
    messageStore.error('保存前检查失败')
  }
}

// Dialog submit (Create only)
const handleEditorSubmit = async (config: TConfig) => {
  try {
    await addOrUpdateConfig(config)
    messageStore.success('配置创建成功')
    editorVisible.value = false
  } catch (e) {
    messageStore.error((e as Error).message || '创建失败')
  }
}

const confirmDelete = async (config: TConfig) => {
  if (!config?.id) {
    messageStore.warning('缺少配置ID，无法删除')
    return
  }

  // Check for related tasks using backend API
  try {
    const relatedTasks = await fetch.get<string[]>('/api/config/related-tasks', { id: config.id })
    if (relatedTasks && relatedTasks.length > 0) {
      const taskList = relatedTasks.join('、')
      messageStore.warning(`无法删除：该配置被以下任务使用：${taskList}`)
      return
    }
  } catch (e) {
    console.error('[ConfigList] 检查关联任务失败:', e)
    messageStore.error('检查关联任务失败')
    return
  }
  
  showDialog('delete', config)
}
</script>

<style scoped>
.config-list-card {
  min-height: calc(100vh - 100px);
}

.header-icon-box {
  width: 56px;
  height: 56px;
  background: white;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
  border: 1px solid rgba(0,0,0,0.05);
}

.empty-state {
  background: rgba(255,255,255,0.5);
  border-color: rgba(0,0,0,0.05);
}

.empty-icon-container {
  width: 100px;
  height: 100px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
}

.fade-in {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
