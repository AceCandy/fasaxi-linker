<template>
  <div class="config-form">
    <v-card variant="outlined" class="mb-4 pa-4">
      <div class="text-subtitle-1 font-weight-bold mb-3 d-flex align-center">
        <v-icon icon="mdi-check-circle-outline" size="20" color="success" class="mr-2"></v-icon>
        匹配规则 (Include)
      </div>
      <v-combobox
        v-model="formData.include"
        label="匹配模式"
        multiple
        chips
        closable-chips
        placeholder="输入文件模式并回车添加 (如: *.mp4, *.mp3, mp*)"
        variant="outlined"
        density="comfortable"
        hide-details="auto"
        hint="支持通配符，如 *.mp4 匹配所有 mp4 文件"
        persistent-hint
      >
        <template v-slot:chip="{ props, item }">
          <v-chip
            v-bind="props"
            :text="item.raw"
            closable
            color="success"
            variant="tonal"
          ></v-chip>
        </template>
      </v-combobox>
    </v-card>

    <v-card variant="outlined" class="mb-4 pa-4">
      <div class="text-subtitle-1 font-weight-bold mb-3 d-flex align-center">
        <v-icon icon="mdi-close-circle-outline" size="20" color="error" class="mr-2"></v-icon>
        排除规则 (Exclude)
      </div>
      <v-combobox
        v-model="formData.exclude"
        label="排除模式"
        multiple
        chips
        closable-chips
        placeholder="输入要排除的文件模式并回车添加 (如: *.tmp, *swp*)"
        variant="outlined"
        density="comfortable"
        hide-details="auto"
        hint="支持通配符，如 *.tmp 排除所有临时文件"
        persistent-hint
      >
        <template v-slot:chip="{ props, item }">
          <v-chip
            v-bind="props"
            :text="item.raw"
            closable
            color="error"
            variant="tonal"
          ></v-chip>
        </template>
      </v-combobox>
    </v-card>

    <v-card variant="outlined" class="mb-4 pa-4">
      <div class="text-subtitle-1 font-weight-bold mb-2">高级选项</div>
      <v-switch
        v-model="formData.keepDirStruct"
        label="保持目录结构"
        color="primary"
        hide-details
      ></v-switch>
      <v-switch
        v-model="formData.openCache"
        label="开启缓存 (推荐)"
        color="primary"
        hide-details
      ></v-switch>
      <v-switch
        v-model="formData.mkdirIfSingle"
        label="单文件是否创建目录"
        color="primary"
        hide-details
      ></v-switch>
      <v-switch
        v-model="formData.deleteDir"
        label="删除源文件后是否删除空目录"
        color="primary"
        hide-details
      ></v-switch>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, toRefs } from 'vue'

const props = defineProps<{
  modelValue: any
}>()

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
