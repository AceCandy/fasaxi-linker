<template>
  <div>
    <TaskList @show-cache="handleShowCache" />
    
    <!-- 缓存管理弹窗 -->
    <v-dialog v-if="cacheVisible" v-model="cacheVisible" max-width="800" class="glass-dialog" scrollable>
      <v-card class="glass-content-card border-neon d-flex flex-column" style="height: 80vh; max-height: 800px;">
        <!-- 头部 -->
        <div class="dialog-header border-b border-neon flex-shrink-0">
          <div class="header-icon-box">
            <v-icon icon="mdi-database" color="primary" size="24"></v-icon>
          </div>
          <div>
            <div class="text-h6 font-weight-bold text-primary font-display">缓存管理</div>
            <div class="text-caption text-slate-400 font-mono">{{ currentTask?.name }}</div>
          </div>
          <v-spacer></v-spacer>
          <v-btn icon="mdi-close" variant="text" density="comfortable" @click="cacheVisible = false" color="grey"></v-btn>
        </div>
        
        <!-- 内容区域：自适应高度 -->
        <v-card-text class="pa-6 flex-grow-1 overflow-hidden d-flex flex-column">
          <!-- 缓存说明 (固定在顶部) -->
          <v-alert
            color="primary"
            variant="tonal"
            density="compact"
            class="mb-6 rounded-lg border border-primary/20 bg-primary/5 flex-shrink-0"
          >
            <template v-slot:prepend>
              <v-icon icon="mdi-information-outline" color="primary"></v-icon>
            </template>
            <div class="text-body-2 text-slate-300 font-mono">
              缓存记录了已创建硬链的源文件路径。清空缓存后，下次执行将重新处理这些文件。
            </div>
          </v-alert>

          <!-- 加载状态 -->
          <div v-if="cacheLoading" class="d-flex justify-center align-center flex-grow-1">
            <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
            <span class="ml-3 text-slate-400 font-weight-medium font-mono">加载中...</span>
          </div>

          <!-- 缓存文件列表容器 -->
          <div v-else class="d-flex flex-column flex-grow-1 overflow-hidden">
            <!-- 工具栏 -->
            <div class="d-flex align-center justify-space-between mb-4 flex-shrink-0">
              <div class="text-subtitle-2 font-weight-bold text-slate-300 d-flex align-center font-display">
                已缓存文件
                <span class="px-2 py-0.5 ml-2 bg-primary/10 text-caption rounded-pill text-primary font-mono">{{ total }}</span>
              </div>
              <v-text-field
                v-if="total > 0 || searchQuery"
                v-model="searchQuery"
                @keydown.enter="handleSearchEnter"
                prepend-inner-icon="mdi-magnify"
                placeholder="搜索 path... (回车或等待)"
                variant="outlined"
                density="compact"
                hide-details
                bg-color="rgba(15, 23, 42, 0.5)"
                class="search-input font-mono"
                style="max-width: 260px"
              ></v-text-field>
            </div>

            <!-- 空状态 -->
            <div v-if="cacheFiles.length === 0" class="empty-state text-center d-flex flex-column align-center justify-center flex-grow-1 rounded-xl border-dashed border-slate-700">
              <div class="icon-circle mb-3 mx-auto">
                <v-icon icon="mdi-database-off-outline" size="32" color="slate-500"></v-icon>
              </div>
              <div class="text-body-2 text-slate-400 font-weight-medium font-mono">{{ searchQuery ? '未找到匹配的文件' : '暂无缓存记录' }}</div>
            </div>

            <!-- 列表 + 分页 -->
            <div v-else class="d-flex flex-column flex-grow-1 overflow-hidden">
              <div class="cache-list custom-scrollbar bg-slate-900/50 border border-slate-700 flex-grow-1" style="overflow-y: auto;">
                <div 
                  v-for="(item, index) in cacheFiles" 
                  :key="item.filePath" 
                  class="cache-item border-b border-slate-800 hover:bg-white/5"
                >
                  <div class="d-flex align-center w-100">
                    <span class="index-badge mr-3 text-slate-500 font-mono" style="min-width: 30px; text-align: right;">{{ (page - 1) * pageSize + index + 1 }}</span>
                    <v-icon icon="mdi-file-clock-outline" size="16" color="primary" class="mr-3 opacity-60"></v-icon>
                    
                    <div class="d-flex flex-column flex-grow-1 overflow-hidden" style="min-width: 0;">
                      <span class="text-body-2 text-truncate text-slate-300 font-mono" :title="item.filePath">{{ item.filePath }}</span>
                      <span class="text-caption text-slate-500 font-mono mt-0.5">
                        <v-icon size="12" class="mr-1">mdi-clock-outline</v-icon>
                        {{ formatDate(item.createdAt) }}
                      </span>
                    </div>

                    <v-btn 
                      icon 
                      size="x-small" 
                      variant="text" 
                      color="grey"
                      class="delete-btn ml-2 flex-shrink-0"
                      @click="handleDeleteSingle(item)"
                    >
                      <v-icon size="20" color="error">mdi-close-circle-outline</v-icon>
                      <v-tooltip activator="parent" location="top">移除</v-tooltip>
                    </v-btn>
                  </div>
                </div>
              </div>
              
              <!-- Pagination -->
              <div class="pt-4 d-flex justify-center flex-shrink-0">
                <v-pagination
                  v-model="page"
                  :length="Math.ceil(total / pageSize)"
                  total-visible="5"
                  density="compact"
                  variant="text"
                  rounded="circle"
                  active-color="primary"
                ></v-pagination>
              </div>
            </div>
          </div>
        </v-card-text>
        
        <v-divider class="border-neon opacity-20"></v-divider>
        
        <!-- 底部按钮 (固定) -->
        <v-card-actions class="pa-5 bg-black/20 flex-shrink-0">
          <v-btn
            variant="text"
            color="grey"
            prepend-icon="mdi-refresh"
            @click="loadCache"
            :loading="cacheLoading"
            class="action-btn font-mono"
          >
            刷新
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            color="grey"
            class="action-btn mr-2 font-mono"
            @click="cacheVisible = false"
          >
            关闭
          </v-btn>
          <v-btn 
            v-if="total > 0"
            color="error" 
            variant="text"
            prepend-icon="mdi-delete-sweep-outline" 
            :loading="clearLoading"
            @click="handleClearAll"
            class="action-btn font-mono"
          >
            全部清空 ({{ total }})
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
      <div class="d-flex justify-center align-center font-mono">
        <v-icon :icon="snackbar.color === 'success' ? 'mdi-check-circle' : 'mdi-alert-circle'" class="mr-2"></v-icon>
        {{ snackbar.text }}
      </div>
    </v-snackbar>
  </div>
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

// 缓存条目类型
interface CacheEntry {
  filePath: string
  createdAt: string
}

const cacheVisible = ref(false)
const currentTask = ref<TTask | null>(null)
const clearLoading = ref(false)
const cacheLoading = ref(false)
const cacheFiles = ref<CacheEntry[]>([]) 
const searchQuery = ref('')
const total = ref(0)
const page = ref(1)
const pageSize = 10

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
    const filePath = (payload as CacheEntry).filePath || payload
    dialogState.content = `确定要删除此缓存项吗？<br><div class="text-body-2 text-slate-300 bg-slate-800 pa-2 mt-2 rounded border border-slate-700 font-mono" style="word-break: break-all;">${filePath}</div>`
    dialogState.type = 'error'
    dialogState.confirmText = '删除'
  } else if (type === 'clear_all') {
    dialogState.title = '确认清空'
    dialogState.content = `确定要清空所有 <strong class="text-error">${total.value}</strong> 条缓存吗？<br><span class="text-caption text-slate-400">此操作不可撤销，清空后下次执行任务将重新处理这些文件。</span>`
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

const handleShowCache = (task: TTask) => {
  currentTask.value = task
  cacheVisible.value = true
  page.value = 1
  searchQuery.value = ''
  loadCache()
}

// Watchers for search and page
let searchTimeout: any
watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    loadCache()
  }, 2000)
})

const handleSearchEnter = () => {
    clearTimeout(searchTimeout) // Cancel pending auto-search
    page.value = 1
    loadCache()
}

watch(page, () => {
  loadCache()
})

// Format date helper - 使用 UTC 方法避免时区转换（DB 存的已经是本地时间）
const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getUTCFullYear()}/${pad(d.getUTCMonth()+1)}/${pad(d.getUTCDate())} ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}`
}

// 从后端加载缓存文件列表
const loadCache = async () => {
  cacheLoading.value = true
  try {
    const taskName = currentTask.value?.name
    if (!taskName) return

    const res = await fetch.get<{
       list: CacheEntry[],
       total: number
    }>(`/api/cache`, { 
       taskName,
       page: page.value,
       pageSize,
       search: searchQuery.value
    })
    
    // Process response
    if (res && Array.isArray(res.list)) {
        cacheFiles.value = res.list
        total.value = res.total || 0
    } else {
        cacheFiles.value = []
        total.value = 0
    }

  } catch (e) {
    console.error('[缓存管理] 加载失败:', e)
    cacheFiles.value = []
    total.value = 0
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
  const payload = dialogState.payload
  const filePath = (payload as CacheEntry).filePath || payload
  dialogState.visible = false
  
  try {
    const taskName = currentTask.value?.name
    if (!taskName) return

    await fetch.delete('/api/cache', {
        taskName,
        files: [filePath]
    })
    
    snackbar.value = { visible: true, text: '已删除', color: 'success' }
    loadCache() // Reload current page
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
    const taskName = currentTask.value?.name
    if (!taskName) return

    await fetch.delete('/api/task/cache', { taskName })
    
    loadCache()
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
  background: rgba(15, 23, 42, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 0 0 40px rgba(0, 240, 255, 0.1) !important;
  overflow: hidden;
  color: #E0F2F7;
}

.dialog-header {
  display: flex;
  align-items: center;
  padding: 20px 24px;
  gap: 12px;
  background: linear-gradient(to right, rgba(15, 23, 42, 0.95), rgba(0, 0, 0, 0.8));
}

.header-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(0, 240, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
}

.cache-item {
  padding: 10px 16px;
  transition: all 0.2s;
  cursor: default;
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.cache-item:hover .delete-btn {
  opacity: 1;
}

.icon-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(30, 41, 59, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.2);
}

/* Custom Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 240, 255, 0.4);
}

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}

.border-neon {
    border-color: rgba(0, 240, 255, 0.3) !important;
}

:deep(.v-field__input) {
    color: #E0F2F7 !important;
}
</style>
