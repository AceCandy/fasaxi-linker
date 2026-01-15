<template>
  <TaskList @show-cache="handleShowCache" />
    
    <!-- 缓存管理弹窗 -->
    <v-dialog v-if="cacheVisible" v-model="cacheVisible" max-width="640" class="glass-dialog">
      <v-card class="glass-content-card">
        <!-- 头部 -->
        <div class="dialog-header">
          <div class="header-icon-box">
            <v-icon icon="mdi-database" color="primary" size="24"></v-icon>
          </div>
          <div>
            <div class="text-h6 font-weight-bold text-grey-darken-3">缓存管理</div>
            <div class="text-caption text-grey">{{ currentTask?.name }}</div>
          </div>
          <v-spacer></v-spacer>
          <v-btn icon="mdi-close" variant="text" density="comfortable" @click="cacheVisible = false" color="grey"></v-btn>
        </div>
        
        <v-divider class="border-opacity-50"></v-divider>
        
        <v-card-text class="pa-6">
          <!-- 缓存说明 -->
          <v-alert
            color="primary"
            variant="tonal"
            density="compact"
            class="mb-6 rounded-lg border-0 bg-blue-grey-lighten-5"
          >
            <template v-slot:prepend>
              <v-icon icon="mdi-information-outline" color="primary"></v-icon>
            </template>
            <div class="text-body-2 text-grey-darken-2">
              缓存记录了已创建硬链的源文件路径。清空缓存后，下次执行将重新处理这些文件。
            </div>
          </v-alert>

          <!-- 加载状态 -->
          <div v-if="cacheLoading" class="d-flex justify-center align-center py-12">
            <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
            <span class="ml-3 text-grey font-weight-medium">加载中...</span>
          </div>

          <!-- 缓存文件列表 -->
          <div v-else>
            <div class="d-flex align-center justify-space-between mb-4">
              <div class="text-subtitle-2 font-weight-bold text-grey-darken-3 d-flex align-center">
                已缓存文件
                <span class="px-2 py-0.5 ml-2 bg-grey-lighten-3 text-caption rounded-pill text-grey-darken-1">{{ cacheFiles.length }}</span>
              </div>
              <v-text-field
                v-if="cacheFiles.length > 0"
                v-model="searchQuery"
                prepend-inner-icon="mdi-magnify"
                placeholder="搜索 path..."
                variant="outlined"
                density="compact"
                hide-details
                bg-color="white"
                class="search-input"
                style="max-width: 220px"
              ></v-text-field>
            </div>

            <div v-if="cacheFiles.length === 0" class="empty-state text-center py-12 rounded-xl border-dashed">
              <div class="icon-circle mb-3 mx-auto">
                <v-icon icon="mdi-database-off-outline" size="32" color="grey-lighten-1"></v-icon>
              </div>
              <div class="text-body-2 text-grey-darken-1 font-weight-medium">暂无缓存记录</div>
              <div class="text-caption text-grey-lighten-1 mt-1">执行任务后，已硬链的文件路径将显示在这里</div>
            </div>

            <div v-else class="cache-list custom-scrollbar">
              <div 
                v-for="(item, index) in displayedFiles" 
                :key="item" 
                class="cache-item"
              >
                <div class="d-flex align-center w-100">
                  <span class="index-badge mr-3">{{ index + 1 }}</span>
                  <v-icon icon="mdi-file-link-outline" size="16" color="primary" class="mr-3 opacity-60"></v-icon>
                  <span class="text-body-2 text-truncate flex-grow-1 text-grey-darken-3 font-mono" :title="item">{{ item }}</span>
                  <v-btn 
                    icon 
                    size="x-small" 
                    variant="text" 
                    color="grey"
                    class="delete-btn ml-2"
                    @click="handleDeleteSingle(item)"
                  >
                    <v-icon size="16">mdi-close</v-icon>
                    <v-tooltip activator="parent" location="top">移除</v-tooltip>
                  </v-btn>
                </div>
              </div>
              
              <div v-if="filteredFiles.length > maxDisplay" class="text-caption text-grey text-center py-3 border-t">
                显示前 {{ maxDisplay }} 条，共 {{ filteredFiles.length }} 条记录
              </div>
            </div>
          </div>
        </v-card-text>
        
        <v-divider class="border-opacity-50"></v-divider>
        
        <v-card-actions class="pa-5 bg-grey-lighten-5">
          <v-btn
            variant="text"
            color="grey-darken-1"
            prepend-icon="mdi-refresh"
            @click="loadCache"
            :loading="cacheLoading"
            class="action-btn"
          >
            刷新
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            color="grey-darken-1"
            class="action-btn mr-2"
            @click="cacheVisible = false"
          >
            关闭
          </v-btn>
          <v-btn 
            v-if="cacheFiles.length > 0"
            color="error" 
            variant="flat"
            prepend-icon="mdi-delete-sweep-outline" 
            :loading="clearLoading"
            @click="handleClearAll"
            class="action-btn elevation-2"
          >
            全部清空
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 统一确认弹窗 -->
    <ConfirmDialog
      v-model="dialogState.visible"
      :title="dialogState.title"
      :content="dialogState.content"
      :type="dialogState.type"
      :loading="clearLoading"
      :confirm-text="dialogState.confirmText"
      @confirm="handleDialogConfirm"
    />
    
    <!-- 提示弹窗 (Snackbar) -->
    <v-snackbar 
      v-model="snackbar.visible" 
      :color="snackbar.color" 
      :timeout="2000" 
      location="top"
      rounded="pill"
      elevation="4"
    >
      <div class="d-flex justify-center align-center">
        <v-icon :icon="snackbar.color === 'success' ? 'mdi-check-circle' : 'mdi-alert-circle'" class="mr-2"></v-icon>
        {{ snackbar.text }}
      </div>
    </v-snackbar>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, reactive } from 'vue'
import TaskList from '../components/TaskList.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import type { TTask } from '../../types/shim'
import fetch from '../kit/fetch'

// 页面挂载时输出日志
onMounted(() => {
  console.log('[TaskListPage] 页面组件挂载')
})

const cacheVisible = ref(false)
const currentTask = ref<TTask | null>(null)
const clearLoading = ref(false)
const cacheLoading = ref(false)
const cacheFiles = ref<string[]>([]) // 当前任务的缓存
const allCacheFiles = ref<string[]>([]) // 全局所有缓存（用于更新后端）
const searchQuery = ref('')
const maxDisplay = 500 // 最多显示500条

// 对话框状态
const snackbar = ref({ visible: false, text: '', color: 'success' })

// Unified Dialog State
type DialogType = 'delete_single' | 'clear_all'
const dialogState = reactive({
  visible: false,
  title: '',
  content: '',
  type: 'warning' as 'warning' | 'error' | 'info',
  confirmText: '确认',
  actionType: null as DialogType | null,
  payload: null as any
})

const showDialog = (type: DialogType, payload?: any) => {
  dialogState.actionType = type
  dialogState.payload = payload
  dialogState.visible = true
  
  if (type === 'delete_single') {
    dialogState.title = '确认删除'
    dialogState.content = `确定要删除此缓存项吗？<br><div class="text-body-2 text-grey bg-grey-lighten-5 pa-2 mt-2 rounded" style="word-break: break-all;">${payload}</div>`
    dialogState.type = 'error'
    dialogState.confirmText = '删除'
  } else if (type === 'clear_all') {
    dialogState.title = '确认清空'
    dialogState.content = `确定要清空所有 <strong class="text-error">${cacheFiles.value.length}</strong> 条缓存吗？<br><span class="text-caption text-grey">此操作不可撤销，清空后下次执行任务将重新处理这些文件。</span>`
    dialogState.type = 'error'
    dialogState.confirmText = '清空'
  }
}

const handleDialogConfirm = async () => {
  if (dialogState.actionType === 'delete_single') {
    await confirmDeleteSingle()
  } else if (dialogState.actionType === 'clear_all') {
    await confirmClearAll()
  }
}

// 过滤后的文件列表
const filteredFiles = computed(() => {
  if (!Array.isArray(cacheFiles.value)) return []
  if (!searchQuery.value) return cacheFiles.value
  const query = searchQuery.value.toLowerCase()
  return cacheFiles.value.filter(f => f && f.toLowerCase().includes(query))
})

// 限制显示数量
const displayedFiles = computed(() => {
  return filteredFiles.value.slice(0, maxDisplay)
})

const handleShowCache = (task: TTask) => {
  currentTask.value = task
  cacheVisible.value = true
  loadCache()
}

// 从后端加载缓存文件列表
const loadCache = async () => {
  cacheLoading.value = true
  try {
    const taskName = currentTask.value?.name
    if (!taskName) {
      console.warn('[缓存管理] 当前任务名称为空')
      cacheFiles.value = []
      return
    }

    // Using native fetch for raw control like before
    const response = await window.fetch(`/api/cache?taskName=${encodeURIComponent(taskName)}`)
    
    // 尝试直接解析 JSON
    let data
    try {
      const text = await response.text()
      if (!text || !text.trim()) {
        cacheFiles.value = []
        return
      }
      data = JSON.parse(text)
    } catch (e) {
      console.warn('[缓存管理] JSON.parse 失败', e)
      cacheFiles.value = []
      return
    }

    // 处理多种可能的返回格式
    let finalFiles = []
    if (Array.isArray(data)) {
      finalFiles = data
    } else if (data && typeof data === 'object') {
       // Support { data: ... } wrapper
       if (data.data) {
           if (Array.isArray(data.data)) finalFiles = data.data
           else if (typeof data.data === 'string') {
               try {
                   const nested = JSON.parse(data.data)
                   if (Array.isArray(nested)) finalFiles = nested
               } catch (e) {}
           }
       }
    } else if (typeof data === 'string') {
        try {
            const nested = JSON.parse(data)
            if (Array.isArray(nested)) finalFiles = nested
        } catch {}
    }

    if (Array.isArray(finalFiles)) {
       cacheFiles.value = finalFiles
       allCacheFiles.value = finalFiles
    } else {
       cacheFiles.value = []
       allCacheFiles.value = []
    }
  } catch (e) {
    console.error('[缓存管理] 加载失败:', e)
    cacheFiles.value = []
  } finally {
    cacheLoading.value = false
  }
}

// 删除单个缓存项 - 显示确认弹窗
const handleDeleteSingle = (filePath: string) => {
  showDialog('delete_single', filePath)
}

// 确认删除单个缓存项
const confirmDeleteSingle = async () => {
  const filePath = dialogState.payload
  dialogState.visible = false
  
  try {
    // 从全局缓存中移除
    const newAllFiles = allCacheFiles.value.filter(f => f !== filePath)
    
    // 更新后端缓存
    // 这里使用了 window.fetch 因为 kit/fetch 封装可能比较简单
    await window.fetch('/api/cache', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: JSON.stringify(newAllFiles) })
    })
    
    // 更新本地状态
    allCacheFiles.value = newAllFiles
    cacheFiles.value = cacheFiles.value.filter(f => f !== filePath)
    snackbar.value = { visible: true, text: '已删除', color: 'success' }
  } catch (e) {
    console.error('删除缓存项失败:', e)
    snackbar.value = { visible: true, text: '删除失败', color: 'error' }
  }
}

// 全部清空 - 显示确认弹窗
const handleClearAll = () => {
  showDialog('clear_all')
}

// 确认清空全部（只清空当前任务的缓存）
const confirmClearAll = async () => {
  clearLoading.value = true
  try {
    const newAllFiles: string[] = [] 
    
    await window.fetch('/api/cache', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: JSON.stringify(newAllFiles) })
    })
    
    allCacheFiles.value = newAllFiles
    cacheFiles.value = []
    searchQuery.value = ''
    dialogState.visible = false
    snackbar.value = { visible: true, text: '该任务缓存已清空', color: 'success' }
  } catch (e) {
    console.error('清空缓存失败:', e)
    snackbar.value = { visible: true, text: '清空失败', color: 'error' }
  } finally {
    clearLoading.value = false
  }
}

// 关闭弹窗时重置状态
watch(cacheVisible, (val) => {
  if (!val) {
    searchQuery.value = ''
  }
})
</script>

<style scoped>
.glass-content-card {
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1) !important;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  padding: 20px 24px;
  gap: 12px;
  background: linear-gradient(to right, rgba(255,255,255,0.95), rgba(255,255,255,0.8));
}

.header-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(102, 126, 234, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.cache-list {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  max-height: 400px;
  overflow-y: auto;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
}

.cache-item {
  padding: 10px 16px;
  border-bottom: 1px solid #f1f5f9;
  transition: all 0.2s;
  cursor: default;
}

.cache-item:last-child {
  border-bottom: none;
}

.cache-item:hover {
  background: #f8fafc;
}

.index-badge {
  color: #94a3b8;
  font-size: 11px;
  font-weight: 600;
  min-width: 24px;
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.cache-item:hover .delete-btn {
  opacity: 1;
}

.empty-state {
  background: #f8fafc;
  border: 1px dashed #e2e8f0;
}

.icon-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.04);
}

/* Custom Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
