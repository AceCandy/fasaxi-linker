<template>
  <v-dialog 
    v-model="visible" 
    max-width="420"
    persistent
    class="glass-dialog"
  >
    <v-card class="dialog-card">
      <div class="dialog-background-glow"></div>
      
      <v-card-text class="pt-8 pb-6 px-6 text-center position-relative">
        <!-- Icon Avatar -->
        <div class="icon-wrapper mb-5">
          <div class="icon-bg" :class="`icon-bg--${type}`"></div>
          <v-icon :icon="icon" size="32" color="white" class="position-relative z-10"></v-icon>
        </div>

        <!-- Title -->
        <h3 class="text-h6 font-weight-bold mb-3 text-grey-darken-4 dialog-title">
          {{ title }}
        </h3>

        <!-- Content -->
        <div 
          class="text-body-2 text-grey-darken-1 mb-2 dialog-content"
          v-html="content"
        ></div>
      </v-card-text>

      <v-card-actions class="px-6 pb-6 pt-2 justify-center gap-3">
        <v-btn
          variant="text"
          class="flex-1-0 rounded-pill text-none px-6 font-weight-medium btn-cancel"
          height="42"
          @click="visible = false"
          :disabled="loading"
        >
          取消
        </v-btn>
        
        <v-btn
          variant="flat"
          class="flex-1-0 rounded-pill text-none px-6 font-weight-bold btn-confirm elevation-4"
          :class="`btn-confirm--${type}`"
          height="42"
          @click="$emit('confirm')"
          :loading="loading"
        >
          {{ confirmText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: boolean
  title?: string
  content?: string
  loading?: boolean
  type?: 'info' | 'warning' | 'error'
  confirmText?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const icon = computed(() => {
  switch (props.type) {
    case 'error': return 'mdi-alert-circle'
    case 'info': return 'mdi-information'
    case 'warning': 
    default: return 'mdi-alert'
  }
})
</script>

<style scoped>
.dialog-card {
  border-radius: 24px !important;
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  box-shadow: 
    0 20px 40px -10px rgba(0, 0, 0, 0.15),
    0 0 0 1px rgba(255, 255, 255, 0.5) inset !important;
  overflow: hidden;
}

.dialog-background-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 120px;
  background: linear-gradient(180deg, rgba(var(--v-theme-surface-variant), 0.05) 0%, transparent 100%);
  pointer-events: none;
}

/* Icon Styles */
.icon-wrapper {
  position: relative;
  width: 64px;
  height: 64px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-bg {
  position: absolute;
  inset: 0;
  border-radius: 20px;
  transform: rotate(45deg);
  transition: all 0.3s ease;
  box-shadow: 0 8px 16px -4px currentColor;
  opacity: 0.9;
}

.icon-bg--warning {
  background: linear-gradient(135deg, #FF9F43 0%, #FF6B6B 100%);
  color: rgba(255, 159, 67, 0.4);
}

.icon-bg--error {
  background: linear-gradient(135deg, #FF6B6B 0%, #EE5253 100%);
  color: rgba(255, 107, 107, 0.4);
}

.icon-bg--info {
  background: linear-gradient(135deg, #54A0FF 0%, #2E86DE 100%);
  color: rgba(84, 160, 255, 0.4);
}

/* Typography */
.dialog-title {
  letter-spacing: -0.01em;
}

.dialog-content {
  line-height: 1.6;
  opacity: 0.85;
}

/* Buttons */
.gap-3 {
  gap: 12px;
}

.flex-1-0 {
  flex: 1 0 auto;
}

.btn-cancel {
  background: rgba(0, 0, 0, 0.04) !important;
  color: #555 !important;
  transition: all 0.2s ease;
}

.btn-cancel:hover {
  background: rgba(0, 0, 0, 0.08) !important;
}

.btn-confirm {
  color: white !important;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.btn-confirm:hover {
  transform: translateY(-2px);
  filter: brightness(1.1);
}

.btn-confirm:active {
  transform: translateY(0);
}

.btn-confirm--warning {
  background: linear-gradient(135deg, #FF9F43 0%, #FF6B6B 100%) !important;
  box-shadow: 0 8px 16px -4px rgba(255, 107, 107, 0.5);
}

.btn-confirm--error {
  background: linear-gradient(135deg, #FF6B6B 0%, #EE5253 100%) !important;
  box-shadow: 0 8px 16px -4px rgba(238, 82, 83, 0.5);
}

.btn-confirm--info {
  background: linear-gradient(135deg, #54A0FF 0%, #2E86DE 100%) !important;
  box-shadow: 0 8px 16px -4px rgba(46, 134, 222, 0.5);
}
</style>
