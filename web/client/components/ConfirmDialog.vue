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
          class="text-body-2 text-slate-300 font-mono mb-2 dialog-content"
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
  background: rgba(15, 23, 42, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  border: 1px solid rgba(0, 240, 255, 0.3) !important;
  box-shadow: 0 0 40px rgba(0, 0, 0, 0.5) !important;
  border-radius: 20px !important;
  overflow: hidden;
}

.dialog-background-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 120px;
  background: radial-gradient(circle at 50% 0%, rgba(0, 240, 255, 0.15) 0%, transparent 70%);
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
  opacity: 0.2;
  border: 1px solid currentColor;
  box-shadow: 0 0 15px currentColor;
}

.icon-bg--warning {
  color: #fb923c;
  background: rgba(251, 146, 60, 0.1);
}

.icon-bg--error {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
}

.icon-bg--info {
  color: #00F0FF;
  background: rgba(0, 240, 255, 0.1);
}

.glow-icon {
  filter: drop-shadow(0 0 8px currentColor);
}

/* Typography */
.dialog-title {
  letter-spacing: 1px;
}

.dialog-content {
  line-height: 1.6;
  opacity: 0.9;
}

/* Buttons */
.gap-3 {
  gap: 12px;
}

.flex-1-0 {
  flex: 1 0 auto;
}

.btn-cancel {
  background: rgba(255, 255, 255, 0.05) !important;
  color: #94a3b8 !important; /* slate-400 */
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.2s ease;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  color: white !important;
}

.btn-confirm {
  color: #0f172a !important; /* Dark text for contrast on bright buttons */
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
  font-weight: 800 !important;
}

.btn-confirm:hover {
  transform: translateY(-2px);
  filter: brightness(1.2);
  box-shadow: 0 0 20px currentColor;
}

.btn-confirm:active {
  transform: translateY(0);
}

.btn-confirm--warning {
  background: linear-gradient(135deg, #fb923c 0%, #ea580c 100%) !important;
  color: white !important;
  box-shadow: 0 4px 12px rgba(234, 88, 12, 0.4);
}

.btn-confirm--error {
  background: linear-gradient(135deg, #f87171 0%, #dc2626 100%) !important;
  color: white !important;
  box-shadow: 0 4px 12px rgba(220, 38, 38, 0.4);
}

.btn-confirm--info {
  background: linear-gradient(135deg, #00F0FF 0%, #00a0aa 100%) !important;
  color: #0f172a !important;
  box-shadow: 0 4px 12px rgba(0, 240, 255, 0.4);
}

.font-display {
    font-family: 'Orbitron', sans-serif;
}
.font-mono {
    font-family: 'Space Mono', monospace;
}
.text-primary-glow {
    color: #E0F2F7;
    text-shadow: 0 0 10px rgba(224, 242, 247, 0.3);
}
</style>
