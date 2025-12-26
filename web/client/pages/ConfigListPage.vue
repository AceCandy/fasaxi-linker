<template>
  <ConfigList />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import ConfigList from '../components/ConfigList.vue'

// 页面挂载时输出日志
onMounted(() => {
  console.log('[ConfigListPage] 页面组件挂载')
})
</script>
