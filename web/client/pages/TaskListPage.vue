<template>
  <TaskList @show-cache="handleShowCache" />
    
    <!-- 缓存管理弹窗 -->
    <v-dialog v-if="cacheVisible" v-model="cacheVisible" max-width="700">
      <v-card v-if="currentTask" class="rounded-lg">
        <v-card-title class="d-flex align-center py-4 px-5 bg-grey-lighten-4">
          <v-icon icon="mdi-database" class="mr-2" color="primary"></v-icon>
          <span>缓存管理 - {{ currentTask.name }}</span>
          <v-spacer></v-spacer>
          <v-btn icon="mdi-close" variant="text" size="small" @click="cacheVisible = false"></v-btn>
        </v-card-title>
        
        <v-divider></v-divider>
        
        <v-card-text class="pa-5">
          <!-- 缓存说明 -->
          <v-alert
            type="info"
            variant="tonal"
            density="compact"
            class="mb-4"
          >
            <div class="text-body-2">
              缓存记录了已创建硬链的源文件路径。清空缓存后，下次执行将重新处理这些文件。
            </div>
          </v-alert>

          <!-- 加载状态 -->
          <div v-if="cacheLoading" class="d-flex justify-center align-center py-8">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
            <span class="ml-3 text-grey">加载缓存中...</span>
          </div>

          <!-- 缓存文件列表 -->
          <div v-else>
            <div class="d-flex align-center justify-space-between mb-3">
              <div class="text-subtitle-2 font-weight-bold">
                已缓存文件
                <v-chip size="x-small" color="primary" class="ml-2">{{ cacheFiles.length }}</v-chip>
              </div>
              <v-text-field
                v-if="cacheFiles.length > 0"
                v-model="searchQuery"
                prepend-inner-icon="mdi-magnify"
                placeholder="搜索文件..."
                variant="outlined"
                density="compact"
                hide-details
                style="max-width: 200px"
                clearable
              ></v-text-field>
            </div>

            <div v-if="cacheFiles.length === 0" class="empty-state text-center py-8">
              <v-icon icon="mdi-database-off-outline" size="48" color="grey-lighten-1" class="mb-3"></v-icon>
              <div class="text-body-2 text-grey">暂无缓存记录</div>
              <div class="text-caption text-grey-lighten-1">执行任务后，已硬链的文件路径将显示在这里</div>
            </div>

            <div v-else class="cache-list" style="max-height: 400px; overflow-y: auto;">
              <div 
                v-for="(item, index) in displayedFiles" 
                :key="item" 
                class="cache-item d-flex align-center px-3"
              >
                <span class="text-grey-darken-1 mr-2" style="min-width: 32px; font-size: 12px;">{{ index + 1 }}.</span>
                <v-icon icon="mdi-file-link-outline" size="16" color="success" class="mr-2 flex-shrink-0"></v-icon>
                <span class="text-body-2 text-truncate flex-grow-1" :title="item">{{ item }}</span>
                <v-btn 
                  icon 
                  size="x-small" 
                  variant="text" 
                  color="error"
                  @click="handleDeleteSingle(item)"
                  class="ml-1 flex-shrink-0"
                >
                  <v-icon size="14">mdi-close</v-icon>
                  <v-tooltip activator="parent" location="top">删除此项</v-tooltip>
                </v-btn>
              </div>
              
              <div v-if="filteredFiles.length > maxDisplay" class="text-caption text-grey text-center py-2 border-t">
                显示前 {{ maxDisplay }} 条，共 {{ filteredFiles.length }} 条记录
              </div>
            </div>
          </div>
        </v-card-text>
        
        <v-divider></v-divider>
        
        <v-card-actions class="pa-4">
          <v-btn
            variant="text"
            color="info"
            prepend-icon="mdi-refresh"
            @click="loadCache"
            :loading="cacheLoading"
          >
            刷新
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="cacheVisible = false">关闭</v-btn>
          <v-btn 
            v-if="cacheFiles.length > 0"
            color="error" 
            variant="flat"
            prepend-icon="mdi-delete-sweep" 
            :loading="clearLoading"
            @click="handleClearAll"
          >
            全部清空 ({{ cacheFiles.length }})
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 删除单个缓存确认弹窗 -->
    <v-dialog v-if="deleteDialog.visible" v-model="deleteDialog.visible" max-width="400" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="d-flex align-center py-4 px-5 bg-error-lighten-5">
          <v-icon icon="mdi-alert-circle" color="error" class="mr-2"></v-icon>
          <span>确认删除</span>
        </v-card-title>
        <v-card-text class="pa-5">
          <p class="text-body-1 mb-2">确定要删除此缓存项吗？</p>
          <div class="text-body-2 text-grey bg-grey-lighten-4 pa-3 rounded" style="word-break: break-all;">
            {{ deleteDialog.filePath }}
          </div>
        </v-card-text>
        <v-card-actions class="pa-4 pt-0">
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="deleteDialog.visible = false">取消</v-btn>
          <v-btn color="error" variant="flat" @click="confirmDeleteSingle">删除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 清空全部确认弹窗 -->
    <v-dialog v-if="clearDialog.visible" v-model="clearDialog.visible" max-width="400" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="d-flex align-center py-4 px-5 bg-warning-lighten-5">
          <v-icon icon="mdi-delete-sweep" color="warning" class="mr-2"></v-icon>
          <span>确认清空</span>
        </v-card-title>
        <v-card-text class="pa-5">
          <p class="text-body-1">确定要清空所有 <strong class="text-error">{{ cacheFiles.length }}</strong> 条缓存吗？</p>
          <p class="text-body-2 text-grey mt-2">此操作不可撤销，清空后下次执行任务将重新处理这些文件。</p>
        </v-card-text>
        <v-card-actions class="pa-4 pt-0">
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="clearDialog.visible = false">取消</v-btn>
          <v-btn color="error" variant="flat" @click="confirmClearAll" :loading="clearLoading">清空</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 提示弹窗 -->
    <v-snackbar v-model="snackbar.visible" :color="snackbar.color" :timeout="2000" location="top">
      {{ snackbar.text }}
    </v-snackbar>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import TaskList from '../components/TaskList.vue'
import type { TTask } from '../../types/shim'

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
const deleteDialog = ref({ visible: false, filePath: '' })
const clearDialog = ref({ visible: false })
const snackbar = ref({ visible: false, text: '', color: 'success' })

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

    // 直接通过 taskName 查询该任务的缓存
    const response = await fetch(`/api/cache?taskName=${encodeURIComponent(taskName)}`)
    console.log('[缓存管理] 响应状态:', response.status, response.statusText)

    // 尝试直接解析 JSON
    let data
    try {
      const text = await response.text()
      console.log('[缓存管理] 原始响应内容(前100字符):', text.slice(0, 100))

      if (!text || !text.trim()) {
        cacheFiles.value = []
        return
      }

      try {
        data = JSON.parse(text)
      } catch (e) {
        console.warn('[缓存管理] JSON.parse 失败', e)
        cacheFiles.value = []
        return
      }
    } catch (e) {
      console.error('[缓存管理] 读取响应失败:', e)
      cacheFiles.value = []
      return
    }

    // 处理多种可能的返回格式
    let finalFiles = []

    if (Array.isArray(data)) {
      // 情况1: 直接返回数组
      finalFiles = data
    } else if (data && typeof data === 'object') {
      // 情况2: 返回对象，可能包含 success/data 字段
      if (data.data) {
        if (Array.isArray(data.data)) {
           // data.data 直接是数组
           finalFiles = data.data
        } else if (typeof data.data === 'string') {
          // data.data 是字符串，尝试二次解析
          try {
            const nested = JSON.parse(data.data)
            if (Array.isArray(nested)) {
              finalFiles = nested
              console.log('[缓存管理] 从 data 字段二次解析成功')
            } else {
              console.warn('[缓存管理] data 字段解析后不是数组:', nested)
            }
          } catch (e) {
            console.warn('[缓存管理] data 字段解析失败:', e)
          }
        }
      } else {
         console.warn('[缓存管理] 响应对象中缺少 data 字段:', data)
      }
    } else if (typeof data === 'string') {
        // 情况3: 顶层就是字符串
         try {
            const nested = JSON.parse(data)
            if (Array.isArray(nested)) {
                finalFiles = nested
            }
        } catch {
             // ignore
        }
    }

    if (Array.isArray(finalFiles)) {
       // 后端已经按 taskName 过滤，直接使用
       cacheFiles.value = finalFiles
       console.log('[缓存管理] 该任务缓存数量:', cacheFiles.value.length)
    } else {
       console.warn('[缓存管理] 无法解析出有效的数组数据')
       cacheFiles.value = []
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
  deleteDialog.value = { visible: true, filePath }
}

// 确认删除单个缓存项
const confirmDeleteSingle = async () => {
  const filePath = deleteDialog.value.filePath
  deleteDialog.value.visible = false
  
  try {
    // 从全局缓存中移除
    const newAllFiles = allCacheFiles.value.filter(f => f !== filePath)
    
    // 更新后端缓存
    await fetch('/api/cache', {
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
  clearDialog.value.visible = true
}

// 确认清空全部（只清空当前任务的缓存）
const confirmClearAll = async () => {
  clearLoading.value = true
  try {
    // 保留其他任务的缓存，只删除当前任务的
    const task = currentTask.value
    let newAllFiles = allCacheFiles.value
    
    if (task && task.pathsMapping && task.pathsMapping.length > 0) {
      const sourcePaths = task.pathsMapping.map((m: any) => m.source).filter(Boolean)
      // 保留不属于该任务的缓存
      newAllFiles = allCacheFiles.value.filter((filePath: string) => 
        !sourcePaths.some((src: string) => filePath.startsWith(src))
      )
    } else {
      // 没有 pathsMapping，清空所有
      newAllFiles = []
    }
    
    await fetch('/api/cache', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: JSON.stringify(newAllFiles) })
    })
    
    allCacheFiles.value = newAllFiles
    cacheFiles.value = []
    searchQuery.value = ''
    clearDialog.value.visible = false
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
.cache-list {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  max-height: 350px;
  overflow-y: auto;
}

.cache-item {
  min-height: 40px;
  padding-top: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f1f5f9;
}

.cache-item:last-child {
  border-bottom: none;
}

.cache-item:hover {
  background: #f1f5f9;
}

.empty-state {
  background: #f8fafc;
  border: 1px dashed #e2e8f0;
  border-radius: 8px;
}
</style>
