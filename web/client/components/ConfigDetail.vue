<template>
  <v-navigation-drawer
    v-model="isOpen"
    location="right"
    width="800"
    temporary
  >
    <div class="d-flex flex-column h-100">
      <!-- 优化后的标题栏 -->
      <div class="pa-4 d-flex align-center justify-space-between elevation-1" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
        <div class="d-flex align-center gap-2">
          <v-icon color="white" size="20">mdi-file-code</v-icon>
          <div class="text-h6 font-weight-medium text-white">{{ name }}</div>
        </div>
        <v-btn 
          icon="mdi-close" 
          variant="text" 
          density="compact" 
          color="white"
          @click="handleClose"
        ></v-btn>
      </div>

      <!-- 优化后的内容区域 -->
      <div class="flex-grow-1 overflow-hidden pa-2">
        <div v-if="loading" class="d-flex flex-column justify-center align-center h-100 gap-3">
          <v-progress-circular 
            indeterminate 
            color="primary" 
            size="48"
          ></v-progress-circular>
          <div class="text-body-2 text-grey">加载配置中...</div>
        </div>
        <div v-else-if="!configDetail?.detail" class="d-flex flex-column justify-center align-center h-100 gap-3">
          <v-icon size="64" color="grey-lighten-1">mdi-file-document-outline</v-icon>
          <div class="text-body-1 text-grey">暂无配置内容</div>
        </div>
        <Editor
          v-else
          :model-value="configDetail?.detail"
          language="javascript"
          read-only
          height="100%"
        />
      </div>
    </div>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Editor from './Editor.vue'
import { useGet } from '../composables/useConfig'

const props = defineProps<{
  id?: number
  name?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const isOpen = ref(false)
const { data: configDetail, getItem, loading } = useGet()

watch(() => props.id, async (newId) => {
  if (newId) {
    await getItem(newId)
    console.log('[ConfigDetail] 加载配置:', newId, '数据:', configDetail.value)
    isOpen.value = true
  } else {
    isOpen.value = false
  }
})

const handleClose = () => {
  isOpen.value = false
  emit('close')
}
</script>
