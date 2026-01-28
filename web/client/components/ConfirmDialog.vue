<template>
  <v-dialog 
    v-model="visible" 
    max-width="420"
    persistent
    class="glass-dialog"
  >
    <v-card class="dialog-card glass-content-card">
      <div class="dialog-background-glow"></div>
      
      <v-card-text class="pt-8 pb-6 px-6 text-center position-relative">
        <!-- Icon Avatar -->
        <div class="icon-wrapper mb-5">
          <div class="icon-bg" :class="`icon-bg--${type}`"></div>
          <v-icon :icon="icon" size="32" class="position-relative z-10 glow-icon"></v-icon>
        </div>

        <!-- Title -->
        <h3 class="text-h6 font-weight-bold mb-3 font-display text-primary-glow dialog-title">
          {{ title }}
        </h3>

        <!-- Content -->
        <div 
          class="text-body-2 text-text-muted font-mono mb-2 dialog-content"
          v-html="content"
        ></div>
      </v-card-text>

      <v-card-actions class="px-6 pb-6 pt-2 justify-center gap-3">
        <v-btn
          variant="text"
          class="flex-1-0 rounded-pill text-none px-6 font-weight-medium btn-cancel font-mono"
          height="42"
          @click="visible = false"
          :disabled="loading"
        >
          取消
        </v-btn>
        
        <v-btn
          variant="flat"
          class="flex-1-0 rounded-pill text-none px-6 font-weight-bold btn-confirm elevation-4 font-mono"
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
.glass-content-card {
  /* Using global style */
}

.dialog-background-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 140px;
  background: radial-gradient(circle at 50% 0%, rgba(var(--color-primary-rgb), 0.1) 0%, transparent 80%);
  pointer-events: none;
}

/* Icon Styles */
.icon-wrapper {
  position: relative;
  width: 72px;
  height: 72px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-bg {
  position: absolute;
  inset: 0;
  border-radius: 24px;
  transform: rotate(45deg);
  transition: all 0.3s ease;
  opacity: 0.2;
  border: 1px solid currentColor;
  box-shadow: 0 0 20px currentColor;
}

.icon-bg--warning {
  color: var(--color-warning);
  background: rgba(var(--color-warning), 0.15);
}

.icon-bg--error {
  color: var(--color-error);
  background: rgba(var(--color-error), 0.15);
}

.icon-bg--info {
  color: var(--color-secondary);
  background: rgba(var(--color-secondary), 0.15);
}

.glow-icon {
  filter: drop-shadow(0 0 10px currentColor);
}

/* Typography */
.dialog-title {
  letter-spacing: 1px;
  font-size: 1.25rem;
  color: var(--color-text);
}

.dialog-content {
  line-height: 1.6;
  opacity: 0.9;
  font-size: 0.95rem;
}

/* Buttons */
.gap-3 {
  gap: 16px;
}

.flex-1-0 {
  flex: 1 0 auto;
}

.btn-cancel {
  background: rgba(var(--color-surface-rgb), 0.5) !important;
  color: var(--color-text-muted) !important;
  border: 1px solid var(--color-border);
  transition: all 0.2s ease;
}

.btn-cancel:hover {
  background: rgba(var(--color-surface-rgb), 0.8) !important;
  color: var(--color-text) !important;
  border-color: var(--color-primary);
}

.btn-confirm {
  color: #fff !important; /* Always white text for these gradient buttons */
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
  font-weight: 800 !important;
  border: none;
}

.btn-confirm:hover {
  transform: translateY(-2px);
  filter: brightness(1.1);
  box-shadow: 0 0 25px currentColor;
}

.btn-confirm:active {
  transform: translateY(0);
}

.btn-confirm--warning {
  background: linear-gradient(135deg, #fb923c 0%, #ea580c 100%) !important;
  color: white !important;
  box-shadow: 0 4px 15px rgba(234, 88, 12, 0.4);
}

.btn-confirm--error {
  background: linear-gradient(135deg, #f87171 0%, #dc2626 100%) !important;
  color: white !important;
  box-shadow: 0 4px 15px rgba(220, 38, 38, 0.4);
}

.btn-confirm--info {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%) !important;
  color: white !important;
  box-shadow: 0 4px 15px rgba(0, 240, 255, 0.4);
}

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
.text-primary-glow {
    color: var(--color-text);
    text-shadow: 0 0 15px rgba(var(--color-primary-rgb), 0.2);
}
</style>
