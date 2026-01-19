<template>
  <v-dialog v-model="isOpen" max-width="750" persistent class="glass-dialog">
    <v-card class="editor-card glass-content-card border-neon">
      <!-- 头部 -->
      <div class="editor-header border-b border-neon">
        <div class="header-content">
          <div class="header-icon">
            <v-icon icon="mdi-cog-outline" size="18" class="text-primary"></v-icon>
          </div>
          <div>
            <div class="header-title font-display text-primary-glow">{{ data ? `编辑配置` : '创建新配置' }}</div>
            <div v-if="data" class="header-subtitle font-mono text-slate-400">{{ data.name }}</div>
          </div>
        </div>
        <button class="close-btn" @click="isOpen = false">
          <v-icon icon="mdi-close" size="18" class="text-slate-400 hover:text-white"></v-icon>
        </button>
      </div>

      <!-- 内容区域 -->
      <v-card-text class="editor-content">
        <v-form ref="form" v-model="valid" @submit.prevent="handleSubmit">
          <!-- 基本信息区域 -->
          <div class="form-section border border-slate-700 bg-slate-900/50">
            <div class="section-title mb-2 text-primary font-display">
              基本信息
            </div>
            
            <!-- 名称字段 -->
            <v-text-field
              v-if="!data"
              v-model="formData.name"
              label="配置名称"
              placeholder="请输入配置名称"
              variant="outlined"
              :rules="nameRules"
              class="mb-2"
              density="compact"
              prepend-inner-icon="mdi-tag-outline"
              hide-details="auto"
              bg-color="transparent"
            ></v-text-field>


          </div>

          <!-- 配置规则区域 -->
          <div class="form-section border border-slate-700 bg-slate-900/50">
            <div class="section-title mb-2 d-flex align-center justify-space-between text-primary font-display">
              <span>配置规则</span>
              <v-btn 
                v-if="data"
                variant="text" 
                size="x-small" 
                @click="resetToDefault"
                prepend-icon="mdi-restore"
                color="secondary"
                class="opacity-70 hover:opacity-100"
              >
                还原
              </v-btn>
            </div>
            
            <ConfigForm
              v-model="visualData"
              class="visual-form-wrapper"
            />
          </div>
        </v-form>
      </v-card-text>

      <!-- 底部操作区域 -->
      <div class="editor-footer border-t border-neon bg-slate-900/80">
        <v-btn variant="text" size="default" color="grey" @click="isOpen = false" class="font-mono">取消</v-btn>
        <v-btn 
          class="submit-btn btn-neon ml-2"
          size="default"
          @click="handleSubmit" 
          :disabled="!valid"
          variant="tonal"
          color="primary"
        >
          <v-icon icon="mdi-check" size="18" class="mr-1"></v-icon>
          {{ data ? '保存' : '创建' }}
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import ConfigForm from './ConfigForm.vue'
import defaultConfig from '../kit/defaultConfig'
import type { TConfig } from '../../types/shim'

const props = defineProps<{
  modelValue: boolean
  data?: TConfig
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', value: TConfig): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const valid = ref(false)
const form = ref<any>(null)

const formData = ref({
  name: '',
  detail: defaultConfig.get()

})

const visualData = ref<any>({})

const nameRules = [
  (v: string) => !!v || '必须填写名称',
  (v: string) => /^[\u4e00-\u9fa5\w-]+$/.test(v) || '名称只能包含中文/数字/字母/下划线/短横线'
]

// Sync Code -> Visual
const syncCodeToVisual = () => {
  try {
    const detail = formData.value.detail || defaultConfig.get()
    // eslint-disable-next-line no-eval
    const parsed = eval(`(${detail?.replace(/(export|default)/g, '')})`)
    
    // 新格式：include 和 exclude 直接是数组
    visualData.value = {
      include: parsed.include || [],
      exclude: parsed.exclude || [],
      keepDirStruct: parsed.keepDirStruct ?? true,
      openCache: parsed.openCache ?? true,
      mkdirIfSingle: parsed.mkdirIfSingle ?? false,
      deleteDir: parsed.deleteDir ?? false,
    }
  } catch (e) {
    console.error('[ConfigEditor] 解析配置失败:', e)
  }
}

// Sync Visual -> Code
const syncVisualToCode = () => {
  const {
    include,
    exclude,
    keepDirStruct,
    openCache,
    mkdirIfSingle,
    deleteDir,
  } = visualData.value

  // 生成 include 数组字符串
  const includeStr = include?.length 
    ? `[${include.map((s: string) => `"${s}"`).join(', ')}]`
    : '[]'

  // 生成 exclude 数组字符串
  const excludeStr = exclude?.length 
    ? `[${exclude.map((s: string) => `"${s}"`).join(', ')}]`
    : '[]'

  const configContent = `export default {
  include: ${includeStr},
  exclude: ${excludeStr},
  keepDirStruct: ${keepDirStruct},
  openCache: ${openCache},
  mkdirIfSingle: ${mkdirIfSingle},
  deleteDir: ${deleteDir},
}`
  formData.value.detail = configContent
}

const resetToDefault = () => {
  formData.value.detail = defaultConfig.get()
  syncCodeToVisual()
}

const handleSubmit = async () => {
  console.log('[ConfigEditor] 开始提交，表单验证中...')
  const { valid: isValid } = await form.value.validate()
  console.log('[ConfigEditor] 表单验证结果:', isValid)
  
  if (isValid) {
    console.log('[ConfigEditor] 可视化数据:', visualData.value)
    
    // 直接使用可视化数据作为 detail（新格式：对象格式）
    const detailObject = {
      include: visualData.value.include || [],
      exclude: visualData.value.exclude || [],
      keepDirStruct: visualData.value.keepDirStruct ?? true,
      openCache: visualData.value.openCache ?? true,
      mkdirIfSingle: visualData.value.mkdirIfSingle ?? false,
      deleteDir: visualData.value.deleteDir ?? false,
    }
    
    console.log('[ConfigEditor] 生成的配置对象（新格式）:', detailObject)

    const payload = props.data 
      ? { id: props.data.id, name: props.data.name, detail: detailObject } 
      : { name: formData.value.name, detail: detailObject }
    
    console.log('[ConfigEditor] 准备提交的数据:', payload)
    emit('submit', payload as unknown as TConfig)
    isOpen.value = false
  } else {
    console.warn('[ConfigEditor] 表单验证失败')
  }
}



// 将 detail 转换为字符串格式（用于代码编辑器）
const convertDetailToString = (detail: any): string => {
  if (!detail) {
    return defaultConfig.get()
  }
  
  // 如果已经是字符串，直接返回
  if (typeof detail === 'string') {
    return detail
  }
  
  // 如果是对象，转换为代码字符串（新格式）
  if (typeof detail === 'object') {
    console.log('[ConfigEditor] detail 是对象，转换为字符串:', detail)
    
    // 生成 include 数组字符串
    const includeStr = detail.include?.length 
      ? `[${detail.include.map((s: string) => `"${s}"`).join(', ')}]`
      : '[]'
    
    // 生成 exclude 数组字符串
    const excludeStr = detail.exclude?.length 
      ? `[${detail.exclude.map((s: string) => `"${s}"`).join(', ')}]`
      : '[]'
    
    return `export default {
  include: ${includeStr},
  exclude: ${excludeStr},
  keepDirStruct: ${detail.keepDirStruct ?? true},
  openCache: ${detail.openCache ?? true},
  mkdirIfSingle: ${detail.mkdirIfSingle ?? false},
  deleteDir: ${detail.deleteDir ?? false},
}`
  }
  
  return defaultConfig.get()
}

watch(() => props.modelValue, (val) => {
  if (val) {
    console.log('[ConfigEditor] 打开编辑器, props.data:', props.data)
    if (props.data) {
      const detailString = convertDetailToString(props.data.detail)
      formData.value = {
        name: props.data.name,

        detail: detailString
      }
      console.log('[ConfigEditor] 初始化表单数据:', formData.value)
    } else {
      formData.value = {
        name: '',

        detail: defaultConfig.get()
      }
    }
    // Initialize visual data
    syncCodeToVisual()
  }
}, { immediate: true })

// 也监听 props.data 的变化
watch(() => props.data, (newData) => {
  if (props.modelValue && newData) {
    console.log('[ConfigEditor] props.data 变化:', newData)
    const detailString = convertDetailToString(newData.detail)
    formData.value = {
      name: newData.name,

      detail: detailString
    }
    syncCodeToVisual()
  }
})
</script>

<style scoped>
.glass-content-card {
  background: rgba(15, 23, 42, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  border-radius: 20px !important;
  box-shadow: 0 0 50px rgba(0, 240, 255, 0.15) !important;
  overflow: hidden;
  color: #E0F2F7;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: linear-gradient(to right, rgba(15, 23, 42, 0.8), rgba(0, 0, 0, 0.6));
}

.header-content {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 240, 255, 0.1);
  border-radius: 8px;
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.header-subtitle {
  font-size: 11px;
  opacity: 0.7;
  margin-top: 2px;
}

.close-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.editor-content {
  padding: 20px !important;
}

/* 顶部/底部边框 */
.border-neon {
  border-color: rgba(0, 240, 255, 0.3) !important;
}
.border-slate-700 {
    border-color: rgba(51, 65, 85, 0.5) !important;
}

.form-section {
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 400px;
}

.code-editor-wrapper {
  flex: 1;
  border-color: rgba(51, 65, 85, 0.5);
  background-color: rgba(15, 23, 42, 0.5);
  overflow: hidden;
  min-height: 400px;
}

.visual-form-wrapper {
  flex: 1;
}

/* 字体工具 */
.font-display {
  font-family: 'Orbitron', sans-serif;
}
.font-mono {
  font-family: 'Space Mono', monospace;
}

/* Input Styles Override */
:deep(.v-field__outline__start),
:deep(.v-field__outline__end),
:deep(.v-field__outline__notch) {
  border-color: rgba(51, 65, 85, 0.5) !important;
}

:deep(.v-field:hover .v-field__outline__start),
:deep(.v-field:hover .v-field__outline__end),
:deep(.v-field:hover .v-field__outline__notch) {
  border-color: rgba(0, 240, 255, 0.5) !important;
}

:deep(.v-field--focused .v-field__outline__start),
:deep(.v-field--focused .v-field__outline__end),
:deep(.v-field--focused .v-field__outline__notch) {
  border-color: #00F0FF !important;
  box-shadow: 0 0 5px rgba(0, 240, 255, 0.2);
}

:deep(.v-field__input) {
  color: #E0F2F7 !important;
  font-family: 'Space Mono', monospace;
}

:deep(.v-label) {
  color: rgba(148, 163, 184, 0.8) !important;
}

/* 优化按钮图标间距 */
:deep(.v-btn--prepend-icon .v-btn__content) {
  gap: 4px;
}

/* 优化切换按钮样式 */
:deep(.v-btn-toggle) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.v-btn-toggle .v-btn) {
  border-radius: 0;
  opacity: 0.8;
}

:deep(.v-btn-toggle .v-btn--active) {
  opacity: 1;
}
</style>
