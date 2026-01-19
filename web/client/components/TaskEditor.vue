<template>
  <v-dialog v-model="isOpen" max-width="750" scrollable persistent class="glass-dialog">
    <v-card class="editor-card glass-content-card border-neon">
      <!-- 头部 -->
      <div class="editor-header border-b border-neon">
        <div class="header-content">
          <div class="header-icon">
            <v-icon icon="mdi-file-document-edit" size="18" class="text-primary"></v-icon>
          </div>
          <span class="header-title font-display text-primary-glow">{{ isEdit ? `编辑${typeText}任务` : `创建${typeText}任务` }}</span>
        </div>
        <button class="close-btn" @click="isOpen = false">
          <v-icon icon="mdi-close" size="18" class="text-slate-400 hover:text-white"></v-icon>
        </button>
      </div>

      <!-- 内容区域 -->
      <v-card-text class="editor-content" style="max-height: 500px;">
        <v-form ref="form" v-model="valid" @submit.prevent="handleSubmit">
          <!-- 基本信息区域 -->
          <div class="form-section border border-slate-700 bg-slate-900/50">
            <div class="section-title mb-2 text-primary font-display">基本信息</div>
            
            <v-text-field
              v-model="formData.name"
              label="任务名称"
              placeholder="请输入任务名称"
              variant="outlined"
              :rules="nameRules"
              :disabled="isEdit"
              class="mb-2"
              density="compact"
              hide-details="auto"
              bg-color="transparent"
            ></v-text-field>

            <v-row class="mb-0" dense>
              <v-col cols="12" sm="6">
                <v-select
                  v-model="formData.type"
                  label="任务类型"
                  :items="typeOptions"
                  item-title="text"
                  item-value="value"
                  variant="outlined"
                  :rules="[v => !!v || '必须选择任务类型']"
                  density="compact"
                  hide-details="auto"
                  bg-color="transparent"
                ></v-select>
              </v-col>
              <v-col cols="12" sm="6">
                <v-select
                  v-model="formData.configId"
                  label="配置文件"
                  :items="configStore.configs"
                  item-title="name"
                  item-value="id"
                  variant="outlined"
                  :rules="[v => !!v || '必须选择配置文件']"
                  :loading="configStore.loading"
                  density="compact"
                  hide-details="auto"
                  bg-color="transparent"
                >
                  <template v-slot:item="{ props, item }">
                    <v-list-item v-bind="props" class="hover:bg-primary/20"></v-list-item>
                  </template>
                </v-select>
              </v-col>
            </v-row>
          </div>

          <!-- 路径映射区域 -->
          <div class="form-section border border-slate-700 bg-slate-900/50">
            <div class="section-title d-flex align-center justify-space-between mb-2 text-primary font-display">
              <span>路径映射</span>
              <v-chip 
                size="x-small" 
                color="primary" 
                variant="tonal"
                class="font-mono bg-primary/10"
              >
                {{ formData.pathsMapping?.length || 0 }}
              </v-chip>
            </div>
            
            <div class="mappings-container custom-scrollbar">
              <div 
                v-for="(mapping, index) in formData.pathsMapping" 
                :key="index" 
                class="mapping-item d-flex align-start gap-2 mb-2 pa-2 border rounded-lg border-slate-700 bg-slate-800/50"
                :class="{ 'last-item': index === ((formData.pathsMapping?.length || 0) - 1) }"
              >
                <div class="flex-grow-1 mapping-fields">
                  <div class="d-flex gap-2">
                    <v-text-field
                      v-model="mapping.source"
                      label="源路径"
                      variant="outlined"
                      density="compact"
                      :rules="[v => !!v || '请输入源路径']"
                      hide-details="auto"
                      class="flex-grow-1"
                      bg-color="transparent"
                    >
                      <template v-slot:prepend-inner>
                        <v-icon size="16" color="primary" class="opacity-70">mdi-folder-open</v-icon>
                      </template>
                    </v-text-field>
                    
                    <v-text-field
                      v-model="mapping.dest"
                      label="目标路径"
                      variant="outlined"
                      density="compact"
                      :rules="[v => !!v || '请输入目标路径']"
                      hide-details="auto"
                      class="flex-grow-1"
                      bg-color="transparent"
                    >
                      <template v-slot:prepend-inner>
                        <v-icon size="16" color="primary" class="opacity-70">mdi-folder-move</v-icon>
                      </template>
                    </v-text-field>
                  </div>
                </div>

                <v-btn
                  icon="mdi-minus-circle-outline"
                  variant="text"
                  color="error"
                  size="small"
                  class="remove-btn"
                  @click="removeMapping(index)"
                  :disabled="formData.pathsMapping?.length === 1"
                ></v-btn>
              </div>
            </div>

            <v-btn
              block
              variant="outlined"
              color="primary"
              prepend-icon="mdi-plus"
              size="small"
              class="mt-1 btn-neon"
              @click="addMapping"
            >
              添加映射
          </v-btn>
          </div>

          <!-- 高级选项 -->
          <div v-if="formData.type === 'prune'" class="form-section border border-slate-700 bg-slate-900/50">
            <div class="section-title mb-2 text-primary font-display">高级选项</div>
            <v-switch
              v-model="formData.reverse"
              label="反向检测"
              color="primary"
              hide-details
              density="compact"
              class="text-slate-300"
            >
              <template v-slot:append>
                <v-tooltip location="top">
                  <template v-slot:activator="{ props }">
                    <v-icon v-bind="props" size="16" color="grey">mdi-help-circle-outline</v-icon>
                  </template>
                  <span>开启后，将从目标路径向源路径检测</span>
                </v-tooltip>
              </template>
            </v-switch>
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
          :loading="submitting" 
          :disabled="!valid"
          variant="tonal"
          color="primary"
        >
          <v-icon icon="mdi-check" size="18" class="mr-1"></v-icon>
          {{ isEdit ? '保存' : '创建' }}
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { TTask } from '../../types/shim'
import { useConfigStore } from '../stores/config'

const props = defineProps<{
  modelValue: boolean
  edit?: TTask
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', value: TTask): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const configStore = useConfigStore()

// 确保配置数据已加载
if (!configStore.initialized && !configStore.loading) {
  configStore.fetchConfigs()
}

const valid = ref(false)
const form = ref<any>(null)
const submitting = ref(false)

const defaultData = {
  name: '',
  type: 'main' as const,
  configId: undefined as number | undefined,
  reverse: false,
  pathsMapping: [{ source: '', dest: '' }]
}

const formData = ref<TTask>({ ...defaultData })

const isEdit = computed(() => !!props.edit)
const typeText = computed(() => formData.value.type === 'prune' ? '同步' : '硬链')

const typeOptions = [
  { text: '硬链(hlink)', value: 'main' },
  { text: '同步(hlink prune)', value: 'prune' }
]

const nameRules = [
  (v: string) => !!v || '必须填写名称',
  //(v: string) => /^[\u4e00-\u9fa5\w]+$/.test(v) || '文件名只能包含中文/数字/字母/下划线'
  (v: string) => /^[\u4e00-\u9fa5\w-]+$/.test(v) || '文件名只能包含中文/数字/字母/下划线/短横线'
]

const addMapping = () => {
  formData.value.pathsMapping?.push({ source: '', dest: '' })
}

const removeMapping = (index: number) => {
  formData.value.pathsMapping?.splice(index, 1)
}

const handleSubmit = async () => {
  const { valid: isValid } = await form.value.validate()
  if (isValid) {
    submitting.value = true
    emit('submit', { ...formData.value })
    // 不在这里关闭对话框，由父组件在成功后关闭
  }
}

// 暴露方法给父组件
const stopSubmitting = () => {
  submitting.value = false
}

const close = () => {
  isOpen.value = false
  submitting.value = false
}

defineExpose({ stopSubmitting, close })

watch(() => props.modelValue, (val) => {
  if (val) {
    console.log('[TaskEditor] 打开编辑器, props.edit:', props.edit)
    if (props.edit) {
      formData.value = JSON.parse(JSON.stringify(props.edit))
      console.log('[TaskEditor] 初始化表单数据:', formData.value)
    } else {
      formData.value = JSON.parse(JSON.stringify(defaultData))
      // Set default config if available
      if (configStore.configs?.length) {
        formData.value.configId = configStore.configs[0].id
      }
    }
  }
}, { immediate: true })

// 也监听 props.edit 的变化
watch(() => props.edit, (newEdit) => {
  if (props.modelValue && newEdit) {
    console.log('[TaskEditor] props.edit 变化:', newEdit)
    formData.value = JSON.parse(JSON.stringify(newEdit))
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

.mappings-container {
  max-height: 300px;
  overflow-y: auto;
  padding-right: 4px;
}

.mapping-item {
  transition: all 0.2s ease;
}

.mapping-item:hover {
  background-color: rgba(30, 41, 59, 0.8) !important;
  border-color: rgba(0, 240, 255, 0.3) !important;
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.05);
}

.remove-btn {
  margin-top: 4px;
  flex-shrink: 0;
  opacity: 0.6;
}

.remove-btn:hover {
  opacity: 1;
}

/* 滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(0,0,0,0.1);
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.1);
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 240, 255, 0.3);
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
</style>
