<template>
  <v-dialog v-model="isOpen" max-width="750" scrollable persistent>
    <v-card class="editor-card">
      <!-- 头部 -->
      <div class="editor-header">
        <div class="header-content">
          <div class="header-icon">
            <v-icon icon="mdi-file-document-edit" size="18" color="white"></v-icon>
          </div>
          <span class="header-title">{{ isEdit ? `编辑${typeText}任务` : `创建${typeText}任务` }}</span>
        </div>
        <button class="close-btn" @click="isOpen = false">
          <v-icon icon="mdi-close" size="18"></v-icon>
        </button>
      </div>

      <!-- 内容区域 -->
      <v-card-text class="editor-content" style="max-height: 500px;">
        <v-form ref="form" v-model="valid" @submit.prevent="handleSubmit">
          <!-- 基本信息区域 -->
          <div class="form-section">
            <div class="section-title mb-2">基本信息</div>
            
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
                ></v-select>
              </v-col>
              <v-col cols="12" sm="6">
                <v-select
                  v-model="formData.config"
                  label="配置文件"
                  :items="configStore.configs"
                  item-title="name"
                  item-value="name"
                  variant="outlined"
                  :rules="[v => !!v || '必须选择配置文件']"
                  :loading="configStore.loading"
                  density="compact"
                  hide-details="auto"
                >
                  <template v-slot:item="{ props, item }">
                    <v-list-item v-bind="props" :subtitle="item.raw.description"></v-list-item>
                  </template>
                </v-select>
              </v-col>
            </v-row>
          </div>

          <!-- 路径映射区域 -->
          <div class="form-section">
            <div class="section-title d-flex align-center justify-space-between mb-2">
              <span>路径映射</span>
              <v-chip 
                size="x-small" 
                color="grey-lighten-1" 
                variant="tonal"
              >
                {{ formData.pathsMapping?.length || 0 }}
              </v-chip>
            </div>
            
            <div class="mappings-container">
              <div 
                v-for="(mapping, index) in formData.pathsMapping" 
                :key="index" 
                class="mapping-item d-flex align-start gap-2 mb-2 pa-2 border rounded-lg"
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
                    >
                      <template v-slot:prepend-inner>
                        <v-icon size="16" color="grey">mdi-folder-open</v-icon>
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
                    >
                      <template v-slot:prepend-inner>
                        <v-icon size="16" color="grey">mdi-folder-move</v-icon>
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
              variant="tonal"
              color="primary"
              prepend-icon="mdi-plus"
              size="small"
              class="mt-1"
              @click="addMapping"
            >
              添加映射
          </v-btn>
          </div>

          <!-- 高级选项 -->
          <div v-if="formData.type === 'prune'" class="form-section">
            <div class="section-title mb-2">高级选项</div>
            <v-switch
              v-model="formData.reverse"
              label="反向检测"
              color="primary"
              hide-details
              density="compact"
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
      <div class="editor-footer">
        <v-btn variant="text" size="default" @click="isOpen = false">取消</v-btn>
        <v-btn 
          class="submit-btn"
          size="default"
          @click="handleSubmit" 
          :loading="submitting" 
          :disabled="!valid"
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
  config: '',
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
        formData.value.config = configStore.configs[0].name
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
.editor-card {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
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
  background: rgba(255, 255, 255, 0.2);
  border-radius: 6px;
}

.header-title {
  font-size: 15px;
  font-weight: 600;
}

.close-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.editor-content {
  padding: 16px !important;
}

.editor-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 16px;
  background: #fafafa;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.submit-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  color: white !important;
  font-weight: 600 !important;
  border-radius: 8px !important;
  text-transform: none !important;
  letter-spacing: 0 !important;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3) !important;
}

.submit-btn:hover {
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4) !important;
}

.form-section {
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  background-color: #fafafa;
  padding: 12px;
  margin-bottom: 12px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  color: #667eea;
  text-transform: uppercase;
  letter-spacing: 0.5px;
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
  background-color: white;
  border-color: rgba(0, 0, 0, 0.06) !important;
  border-radius: 8px !important;
  transition: all 0.2s ease;
}

.mapping-item:hover {
  background-color: #fafafa;
  border-color: rgba(102, 126, 234, 0.3) !important;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
}

.mapping-item.last-item {
  margin-bottom: 0 !important;
}

.mapping-fields {
  min-width: 0;
}

.remove-btn {
  margin-top: 4px;
  flex-shrink: 0;
}

.close-btn:hover {
  background-color: rgba(0, 0, 0, 0.08);
}

/* 响应式优化 */
@media (max-width: 600px) {
  .form-section {
    padding: 16px;
  }
  
  .mapping-fields .d-flex {
    flex-direction: column;
    gap: 12px !important;
  }
}

/* 滚动条样式 */
.mappings-container::-webkit-scrollbar {
  width: 6px;
}

.mappings-container::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.mappings-container::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.mappings-container::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 优化表单字段样式 */
:deep(.v-field) {
  margin-bottom: 0;
}

:deep(.v-field__prepend-inner) {
  padding-left: 8px;
}

:deep(.v-field__prepend-inner .v-icon) {
  opacity: 0.7;
}

/* 优化按钮图标间距 */
:deep(.v-btn--prepend-icon .v-btn__content) {
  gap: 6px;
}

/* Chip 样式优化 */
:deep(.v-chip) {
  font-size: 12px;
}

/* Switch 样式优化 */
:deep(.v-switch .v-switch__track) {
  opacity: 0.8;
}

:deep(.v-switch--inset .v-switch__track) {
  background-color: rgba(0, 0, 0, 0.08);
}

:deep(.v-switch--inset.v-switch--model-value .v-switch__track) {
  background-color: #1976d2;
}
</style>
