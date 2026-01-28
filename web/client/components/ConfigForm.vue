<template>
  <div class="config-form" :class="{ 'config-form-compact': compact }">
    <!-- Include Rules -->
    <div :class="compact ? 'mb-4' : 'mb-4 wrapper-card pa-4'">
      <div v-if="!compact" class="text-subtitle-2 font-weight-bold mb-1 d-flex align-center text-text-muted font-display">
        <v-icon icon="mdi-check-circle-outline" size="16" color="success" class="mr-1"></v-icon>
        匹配规则
      </div>
      <div v-else class="text-caption font-weight-bold mb-1 d-flex align-center text-text-muted font-display">
        <v-icon icon="mdi-check-circle-outline" size="14" color="success" class="mr-1"></v-icon>
        匹配规则 (Include)
      </div>
      
      <v-combobox
        v-model="formData.include"
        :placeholder="compact ? '输入匹配模式...' : '输入文件模式并回车添加 (如: *.mp4, *.mp3, mp*)'"
        multiple
        chips
        closable-chips
        variant="outlined"
        :density="compact ? 'compact' : 'comfortable'"
        hide-details="auto"
        :menu-props="{ maxHeight: 200, contentClass: 'glass-menu' }"
        class="custom-input"
        bg-color="transparent"
      >
        <template v-slot:chip="{ props, item }">
          <v-chip
            v-bind="props"
            :text="item.raw"
            closable
            size="x-small"
            color="success"
            variant="tonal"
            label
            class="ma-1 font-mono"
          ></v-chip>
        </template>
      </v-combobox>
      <div v-if="compact" class="text-caption text-slate-500 mt-1 ml-1" style="font-size: 10px !important;">
        支持通配符，如 *.mp4 匹配所有 mp4 文件
      </div>
      <div v-else class="text-caption text-slate-500 mt-1">
        支持通配符，如 *.mp4 匹配所有 mp4 文件
      </div>
    </div>

    <!-- Exclude Rules -->
    <div :class="compact ? 'mb-4' : 'mb-4 wrapper-card pa-4'">
      <div v-if="!compact" class="text-subtitle-2 font-weight-bold mb-1 d-flex align-center text-text-muted font-display">
        <v-icon icon="mdi-close-circle-outline" size="16" color="error" class="mr-1"></v-icon>
        排除规则
      </div>
      <div v-else class="text-caption font-weight-bold mb-1 d-flex align-center text-text-muted font-display">
        <v-icon icon="mdi-close-circle-outline" size="14" color="error" class="mr-1"></v-icon>
        排除规则 (Exclude)
      </div>

      <v-combobox
        v-model="formData.exclude"
        :placeholder="compact ? '输入排除模式...' : '输入要排除的文件模式并回车添加 (如: *.tmp, *swp*)'"
        multiple
        chips
        closable-chips
        variant="outlined"
        :density="compact ? 'compact' : 'comfortable'"
        hide-details="auto"
        :menu-props="{ maxHeight: 200, contentClass: 'glass-menu' }"
        class="custom-input"
        bg-color="transparent"
      >
        <template v-slot:chip="{ props, item }">
          <v-chip
            v-bind="props"
            :text="item.raw"
            closable
            size="x-small"
            color="error"
            variant="tonal"
            label
            class="ma-1 font-mono"
          ></v-chip>
        </template>
      </v-combobox>
      <div v-if="compact" class="text-caption text-slate-500 mt-1 ml-1" style="font-size: 10px !important;">
        支持通配符，如 *.tmp 排除所有临时文件
      </div>
      <div v-else class="text-caption text-slate-500 mt-1">
        支持通配符，如 *.tmp 排除所有临时文件
      </div>
    </div>

    <!-- Advanced Options -->
    <div :class="compact ? 'pt-2 border-t border-slate-700' : 'mb-4 wrapper-card pa-4'">
      <div v-if="!compact" class="text-subtitle-1 font-weight-bold mb-2 text-primary font-display">高级选项</div>
      <!-- In compact mode, title is removed -->
      
      <v-row :no-gutters="compact" :class="{ 'dense-row': compact }">
        <v-col cols="6">
          <v-switch
            v-model="formData.keepDirStruct"
            label="保持目录结构"
            color="primary"
            class="compact-switch text-text-muted"
            hide-details
            density="compact"
            inset
          ></v-switch>
        </v-col>
        <v-col cols="6">
          <v-switch
            v-model="formData.openCache"
            label="开启缓存(推荐)"
            color="primary"
            class="compact-switch text-text-muted"
            hide-details
            density="compact"
            inset
          ></v-switch>
        </v-col>
        <v-col cols="6">
          <v-switch
            v-model="formData.mkdirIfSingle"
            label="单文件创建目录"
            color="primary"
            class="compact-switch text-text-muted"
            hide-details
            density="compact"
            inset
          ></v-switch>
        </v-col>
        <v-col cols="6">
          <v-switch
            v-model="formData.deleteDir"
            label="清理空目录"
            color="primary"
            class="compact-switch text-text-muted"
            hide-details
            density="compact"
            inset
          ></v-switch>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: any
  compact?: boolean
}>(), {
  compact: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()

const formData = ref({
  include: [] as string[],
  exclude: [] as string[],
  keepDirStruct: true,
  openCache: true,
  mkdirIfSingle: false,
  deleteDir: false,
  ...props.modelValue
})

watch(formData, (val) => {
  emit('update:modelValue', val)
}, { deep: true })

watch(() => props.modelValue, (val) => {
  if (val && JSON.stringify(val) !== JSON.stringify(formData.value)) {
    formData.value = { ...formData.value, ...val }
  }
}, { deep: true })
</script>

<style scoped>
.wrapper-card {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background-color: transparent;
}

.border-t {
  border-top: 1px dashed var(--color-border);
}

.compact-switch :deep(.v-label) {
  font-size: 12px !important; /* Slightly smaller font for compact feel */
  opacity: 0.9;
  color: var(--color-text-muted) !important;
}

.compact-switch :deep(.v-switch__track) {
  height: 18px; /* Thinner track */
  width: 32px;
  min-width: 32px;
  background-color: var(--color-border);
}

.compact-switch :deep(.v-switch__thumb) {
  height: 14px;
  width: 14px;
}

/* Custom Input Style for Reference Look */
:deep(.custom-input .v-field__outline__start) {
  border-radius: 6px 0 0 6px !important;
  border-color: var(--color-border) !important;
}
:deep(.custom-input .v-field__outline__end) {
  border-radius: 0 6px 6px 0 !important;
  border-color: var(--color-border) !important;
}

:deep(.custom-input .v-field) {
  border-radius: 6px;
}

:deep(.custom-input .v-field--focused .v-field__outline__start),
:deep(.custom-input .v-field--focused .v-field__outline__end),
:deep(.custom-input .v-field--focused .v-field__outline__notch) {
    border-color: #00F0FF !important;
}

:deep(.v-chip) {
    font-family: 'Space Mono', monospace;
}

.dense-row {
  margin-top: 0;
}

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
</style>
