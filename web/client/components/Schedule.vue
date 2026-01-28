<template>
  <v-navigation-drawer
    v-model="isOpen"
    location="right"
    width="450"
    temporary
  >
    <div class="d-flex flex-column h-100">
      <div class="pa-4 border-b bg-grey-lighten-4 d-flex align-center justify-space-between">
        <div class="text-h6 font-weight-bold">{{ taskName }} 定时任务设置</div>
        <v-btn icon="mdi-close" variant="text" density="compact" @click="isOpen = false"></v-btn>
      </div>

      <div class="pa-4 flex-grow-1">
        <v-form ref="form" v-model="valid" @submit.prevent="handleSubmit">
          <v-select
            v-model="formData.scheduleType"
            label="定时任务类型"
            :items="typeOptions"
            item-title="text"
            item-value="value"
            variant="outlined"
            :rules="[v => !!v || '必须选择类型']"
            class="mb-4"
          ></v-select>

          <v-text-field
            v-if="formData.scheduleType === 'cron'"
            v-model="formData.cronValue"
            label="cron规则"
            placeholder="请输入cron规则"
            variant="outlined"
            :rules="[v => !!v || '填入cron规则']"
            hint="需要帮助? 点击查找 crontab规则"
            persistent-hint
          >
            <template v-slot:append-inner>
              <v-btn
                icon="mdi-help-circle-outline"
                variant="text"
                density="compact"
                href="https://tooltt.com/crontab/c/56.html"
                target="_blank"
              ></v-btn>
            </template>
          </v-text-field>

          <v-text-field
            v-if="formData.scheduleType === 'loop'"
            v-model="formData.loopValue"
            label="执行周期"
            placeholder="多少"
            variant="outlined"
            prefix="每"
            suffix="秒执行一次"
            type="number"
            :rules="[v => !!v || '填入执行周期']"
          ></v-text-field>
        </v-form>
      </div>

      <div class="pa-4 border-t bg-grey-lighten-5 d-flex justify-end gap-2">
        <v-btn variant="text" @click="isOpen = false">关闭</v-btn>
        <v-btn color="primary" @click="handleSubmit" :disabled="!valid">确定</v-btn>
      </div>
    </div>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { TSchedule } from '../../types/shim'

const props = defineProps<{
  modelValue: boolean
  taskId: number
  taskName?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', value: TSchedule): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const valid = ref(false)
const form = ref<any>(null)

const formData = ref({
  scheduleType: 'loop' as TSchedule['scheduleType'],
  loopValue: '',
  cronValue: ''
})

const typeOptions = [
  { text: '定时循环(新手推荐)', value: 'loop' },
  { text: '计划任务(cron)', value: 'cron' }
]

const handleSubmit = async () => {
  const { valid: isValid } = await form.value.validate()
  if (isValid) {
    emit('submit', {
      taskId: props.taskId,
      scheduleType: formData.value.scheduleType,
      scheduleValue: formData.value.scheduleType === 'cron' 
        ? formData.value.cronValue 
        : formData.value.loopValue
    })
    isOpen.value = false
  }
}

// Reset form when opening
watch(() => props.modelValue, (val) => {
  if (val) {
    formData.value = {
      scheduleType: 'loop',
      loopValue: '',
      cronValue: ''
    }
  }
})
</script>
