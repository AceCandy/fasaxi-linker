<template>
  <v-card class="config-card glass-card-interactive mb-4" :class="{ 'editing': isDirty }">
    <!-- 头部 -->
    <div class="card-header pt-4 px-4 pb-2 d-flex align-center justify-space-between" style="border-bottom: 1px dashed var(--color-border);">
      <div class="d-flex align-center flex-grow-1 mr-2" style="min-width: 0;">
         <div class="header-icon-wrapper mr-3">
            <v-icon icon="mdi-cog" color="primary" size="18"></v-icon>
         </div>
        <div class="flex-grow-1" style="min-width: 0;">
          <v-text-field
            v-model="localData.name"
            variant="plain"
            density="compact"
            hide-details
            class="title-input font-weight-bold ml-n1 text-text-muted font-mono"
            placeholder="配置名称"
            @update:model-value="checkDirty"
          ></v-text-field>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions flex-shrink-0 d-flex align-center">
        <v-btn
          v-if="isDirty"
          color="primary"
          size="x-small"
          variant="flat"
          class="mr-2 fade-in round-btn elevation-2 font-mono"
          :loading="saving"
          @click="handleSave"
        >
          保存
        </v-btn>
        <v-btn
          icon
          size="x-small"
          variant="text"
          color="grey"
          class="delete-btn"
          @click="$emit('delete', data)"
        >
          <v-icon size="20">mdi-delete-outline</v-icon>
          <v-tooltip activator="parent" location="top">删除配置</v-tooltip>
        </v-btn>
      </div>
    </div>

    <!-- 表单区域 -->
    <v-card-text class="pt-4 px-4 pb-4">
      <ConfigForm
        v-model="visualData"
        class="flat-form"
        compact
      />
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import ConfigForm from './ConfigForm.vue'
import defaultConfig from '../kit/defaultConfig'
import type { TConfig } from '../../types/shim'

const props = defineProps<{
  data: TConfig
}>()

const emit = defineEmits<{
  (e: 'save', data: TConfig): void
  (e: 'delete', data: TConfig): void
}>()

const saving = ref(false)
const isDirty = ref(false)
const localData = ref({
  name: ''
})

// Visual data for ConfigForm
const visualData = ref<any>({})

// Initial state for dirty check
let initialSnapshot = ''

const takeSnapshot = () => {
  return JSON.stringify({
    name: localData.value.name,
    visual: visualData.value
  })
}

const checkDirty = () => {
  const current = takeSnapshot()
  isDirty.value = current !== initialSnapshot
}

// Parse detail string/object to visual object
const parseDetail = (detail: any) => {
  try {
    if (!detail) return defaultConfig.get()

    // If object, use directly
    if (typeof detail === 'object') {
      return {
        include: detail.include || [],
        exclude: detail.exclude || [],
        keepDirStruct: detail.keepDirStruct ?? true,
        openCache: detail.openCache ?? true,
        mkdirIfSingle: detail.mkdirIfSingle ?? false,
        deleteDir: detail.deleteDir ?? false,
      }
    }

    // If string (legacy JS code)
    if (typeof detail === 'string') {
       // eslint-disable-next-line no-eval
       const parsed = eval(`(${detail.replace(/(export|default)/g, '')})`)
       return {
        include: parsed.include || [],
        exclude: parsed.exclude || [],
        keepDirStruct: parsed.keepDirStruct ?? true,
        openCache: parsed.openCache ?? true,
        mkdirIfSingle: parsed.mkdirIfSingle ?? false,
        deleteDir: parsed.deleteDir ?? false,
      }
    }
  } catch (e) {
    console.error('Parse config error:', e)
  }
  return {
    include: [],
    exclude: [],
    keepDirStruct: true,
    openCache: true,
    mkdirIfSingle: false,
    deleteDir: false,
  }
}

watch(() => props.data, (newVal) => {
  if (newVal) {
    localData.value.name = newVal.name
    visualData.value = parseDetail(newVal.detail || defaultConfig.get())

    // Wait for next tick to set snapshot to avoid immediate dirty
    setTimeout(() => {
      initialSnapshot = takeSnapshot()
      isDirty.value = false
    }, 0)
  }
}, { immediate: true, deep: true })

// Watch visualData changes to update dirty state
watch(visualData, () => {
  checkDirty()
}, { deep: true })

const handleSave = async () => {
  saving.value = true
  try {
    // Construct payload
    const payload: TConfig = {
      ...props.data,
      name: localData.value.name,
      detail: visualData.value // Send object
    }

    await emit('save', payload)

    // Update snapshot after successful save (parent should update props, but just in case)
    initialSnapshot = takeSnapshot()
    isDirty.value = false
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.config-card {
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  border-radius: 12px !important;
  overflow: hidden;
  box-shadow: var(--shadow-glass);
  color: var(--color-text);
}

.config-card.editing {
  background: var(--glass-bg-strong);
  border-color: var(--color-primary) !important;
  box-shadow: var(--shadow-neon);
}

.title-input :deep(input) {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text) !important;
  letter-spacing: 0.01em;
  padding-top: 0;
  padding-bottom: 0;
  min-height: 24px;
}

.round-btn {
  border-radius: 20px !important;
}

.header-icon-wrapper {
  width: 32px;
  height: 32px;
  background: rgba(var(--color-primary-rgb), 0.05);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  box-shadow: 0 0 10px rgba(var(--color-primary-rgb), 0.05);
  border: 1px solid var(--color-border);
}

.glass-card-interactive:hover .header-icon-wrapper {
  background: rgba(var(--color-primary-rgb), 0.15);
  box-shadow: 0 0 15px rgba(var(--color-primary-rgb), 0.2);
  border-color: rgba(var(--color-primary-rgb), 0.3);
}

.delete-btn {
  transition: all 0.2s ease;
  opacity: 0.6;
}

.delete-btn:hover {
  color: #ef4444 !important;
  opacity: 1;
  background: rgba(239, 68, 68, 0.1); /* Red tint on hover */
}

.flat-form :deep(.v-card) {
  box-shadow: none !important;
  border: none;
  background: transparent;
}

.fade-in {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateX(10px); }
  to { opacity: 1; transform: translateX(0); }
}

.font-mono {
  font-family: 'Space Mono', monospace;
}
</style>
