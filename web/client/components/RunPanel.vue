<template>
  <div
    ref="containerRef"
    class="run-panel bg-grey-darken-4 text-grey-lighten-2 pa-4 rounded-lg overflow-y-auto font-mono text-body-2 elevation-2"
    style="min-height: 300px; max-height: 50vh; white-space: pre-wrap; line-height: 1.6;"
  >
    <div v-if="!formattedContent" class="d-flex flex-column justify-center align-center h-100 text-grey-lighten-1">
      <div class="text-body-2">等待任务输出...</div>
    </div>
    <div v-else v-html="formattedContent"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
// @ts-ignore
import ansiHtml from 'ansi-html'

// Configure ansi-html colors to match the dark theme
ansiHtml.setColors({
  red: 'ca372d',
  green: '4c7b3a',
  yellow: 'c6c964',
  blue: '4387cf',
  magenta: 'b86cb4',
  cyan: '71d2c4',
  white: 'c3cac1',
  gray: '9a9b99',
})

const props = defineProps<{
  data: string[]
}>()

const containerRef = ref<HTMLElement | null>(null)

const formattedContent = computed(() => {
  // Deduplicate and join
  const uniqueData = Array.from(new Set(props.data))
  return uniqueData.map(line => ansiHtml(line)).join('\n')
})

// Auto-scroll to bottom when data changes
watch(() => props.data, () => {
  nextTick(() => {
    if (containerRef.value) {
      containerRef.value.scrollTop = containerRef.value.scrollHeight
    }
  })
}, { deep: true })
</script>

<style scoped>
.font-mono {
  font-family: Consolas, Monaco, monospace;
}
</style>
