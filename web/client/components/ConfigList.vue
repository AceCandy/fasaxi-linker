<template>
  <div>
    <v-card class="config-list-card rounded-xl fade-in">
      <!-- 卡片头部 -->
      <v-card-title class="d-flex align-center py-5 px-6 card-header">
        <div class="d-flex align-center">
          <div class="header-icon mr-3">
            <v-icon icon="mdi-cog" size="24" color="white"></v-icon>
          </div>
          <div>
            <span class="text-h6 font-weight-bold">配置管理</span>
            <div class="text-caption text-grey">管理您的硬链配置规则</div>
          </div>
        </div>
        <v-spacer></v-spacer>
        <v-btn
          v-if="configStore.configs?.length"
          class="create-btn"
          prepend-icon="mdi-plus"
          @click="handleCreate"
        >
          创建配置
        </v-btn>
      </v-card-title>
      
      <v-divider></v-divider>

      <v-card-text class="pa-6">
        <!-- 空状态 -->
        <div v-if="!configStore.configs?.length" class="empty-state d-flex flex-column align-center justify-center pa-12 text-center">
          <div class="empty-icon-container mb-6">
            <v-icon size="72" color="grey-lighten-1" class="float-animation">mdi-cog-off-outline</v-icon>
          </div>
          <div class="text-h6 text-grey-darken-1 mb-2">暂无配置</div>
          <div class="text-body-2 text-grey mb-6">创建一个配置来定义硬链规则</div>
          <v-btn class="create-btn" prepend-icon="mdi-plus" size="large" @click="handleCreate">
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
          lg="4"
        >
          <div class="config-card" @click="handleEdit(item.id)">
            <!-- 顶部彩色条 -->
            <div class="config-accent"></div>
              
              <div class="config-content">
                <div class="config-header">
                  <div class="config-icon">
                    <v-icon icon="mdi-file-cog-outline" size="20" color="white"></v-icon>
                  </div>
                  <div class="config-info">
                    <div class="config-name">{{ item.name }}</div>
                    <div class="config-desc">{{ item.description || '无描述' }}</div>
                  </div>
                </div>
              </div>
              
              <div class="config-actions">
                <v-btn icon size="x-small" variant="text" color="primary" @click.stop="handleEdit(item.id)">
                  <v-icon size="16">mdi-pencil-outline</v-icon>
                  <v-tooltip activator="parent" location="top">编辑</v-tooltip>
                </v-btn>
                <v-spacer></v-spacer>
                <v-btn icon size="x-small" variant="text" color="error" @click.stop="confirmDelete(item)">
                  <v-icon size="16">mdi-delete-outline</v-icon>
                  <v-tooltip activator="parent" location="top">删除</v-tooltip>
                </v-btn>
              </div>
            </div>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Modals -->
    <ConfigEditor
      v-if="editorVisible"
      v-model="editorVisible"
      :data="currentConfig"
      @submit="handleEditorSubmit"
    />

    <v-dialog v-if="deleteDialog.visible" v-model="deleteDialog.visible" max-width="320">
      <v-card class="rounded-lg">
        <v-card-title class="text-subtitle-1">确认删除</v-card-title>
        <v-card-text class="text-body-2">确认删除配置 "{{ deleteDialog.name }}" 吗？</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" size="small" @click="deleteDialog.visible = false">取消</v-btn>
          <v-btn color="error" variant="flat" size="small" @click="handleDelete" :loading="deleteLoading">删除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, defineAsyncComponent, onMounted } from 'vue'
import { useGet, useAddOrEdit, useDelete } from '../composables/useConfig'
import { useConfigStore } from '../stores/config'
import { useTaskStore } from '../stores/task'
import { useMessageStore } from '../stores/message'
const ConfigEditor = defineAsyncComponent(() => import('./ConfigEditor.vue'))
import type { TConfig } from '../../types/shim'

const configStore = useConfigStore()
const taskStore = useTaskStore()
const messageStore = useMessageStore()

// 每次组件创建时都获取最新数据（因为路由有key，每次切换都会重新创建组件）
onMounted(() => {
  console.log('[ConfigList] 组件挂载，获取最新配置列表')
  configStore.fetchConfigs(true) // 强制刷新
})

const { getItem, data: configData } = useGet()
const { addOrUpdateConfig } = useAddOrEdit({ onSuccess: () => configStore.refreshConfigs() })
const { rmItem, loading: deleteLoading } = useDelete({ onSuccess: () => {
  configStore.refreshConfigs()
  deleteDialog.value.visible = false
}})

const editorVisible = ref(false)
const currentConfig = ref<TConfig | undefined>(undefined)
const deleteDialog = ref({ visible: false, id: 0, name: '' })

const handleCreate = () => {
  currentConfig.value = undefined
  editorVisible.value = true
}

const handleEdit = async (id: number) => {
  await getItem(id)
  console.log('[ConfigList] 编辑配置:', id, '数据:', configData.value)
  currentConfig.value = configData.value
  editorVisible.value = true
}

const handleEditorSubmit = async (config: TConfig) => {
  console.log('[ConfigList] 收到提交的配置:', config)
  console.log('[ConfigList] 当前配置名称:', currentConfig.value?.name)
  try {
    await addOrUpdateConfig(config, currentConfig.value?.id)
    console.log('[ConfigList] 配置保存成功')
    messageStore.success(currentConfig.value?.name ? '配置更新成功' : '配置创建成功')
    editorVisible.value = false
  } catch (e) {
    console.error('[ConfigList] 配置保存失败:', e)
    messageStore.error((e as Error).message || '保存失败')
  }
}

const confirmDelete = async (config: TConfig) => {
  if (!config?.id) {
    messageStore.warning('缺少配置ID，无法删除')
    return
  }
  // 确保任务数据已加载，以检查配置是否被使用
  if (!taskStore.initialized) {
    console.log('[ConfigList] 删除配置前加载任务列表')
    await taskStore.fetchTasks()
  }
  
  const usedBy = taskStore.tasks?.find(t => (config.id && t.configId === config.id) || t.config === config.name)
  if (usedBy) {
    messageStore.warning(`无法删除：该配置被任务 "${usedBy.name}" 使用中`)
    return
  }
  deleteDialog.value = { visible: true, id: config.id, name: config.name }
}

const handleDelete = async () => {
  await rmItem(deleteDialog.value.id)
}
</script>

<style scoped>
.config-list-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
}

/* 页面渐入动画 */
.fade-in {
  animation: fadeInUp 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
}

.header-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

.create-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  color: white !important;
  border: none !important;
  border-radius: 10px !important;
  font-weight: 600 !important;
  padding: 0 24px !important;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3) !important;
  transition: all 0.3s ease !important;
}

.create-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4) !important;
}

.empty-state {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.03) 0%, rgba(118, 75, 162, 0.03) 100%);
  border-radius: 16px;
  border: 2px dashed rgba(102, 126, 234, 0.2);
}

.empty-icon-container {
  width: 120px;
  height: 120px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.float-animation {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.config-card {
  background: white;
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.config-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.12);
}

.config-card:active {
  transform: translateY(-2px);
  transition: all 0.1s cubic-bezier(0.4, 0, 0.2, 1);
}

.config-accent {
  height: 4px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.config-content {
  padding: 16px;
}

.config-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.config-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.config-info {
  flex: 1;
  min-width: 0;
}

.config-name {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.config-desc {
  font-size: 12px;
  color: #888;
  margin-top: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.config-actions {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
}
</style>
