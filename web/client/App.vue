<template>
  <v-app>
    <router-view />
    
    <!-- 全局消息提示 - Dark Tech Futuristic -->
    <Transition name="message-slide">
      <div v-if="messageStore.state.visible" class="message-toast glass-toast" :class="`message-toast--${messageStore.state.type}`">
        <div class="message-content">
          <div class="message-icon-box">
            <v-icon :icon="messageIcon" size="20" class="glow-icon"></v-icon>
          </div>
          <span class="message-text font-mono">{{ messageStore.state.text }}</span>
          <button class="message-close" @click="messageStore.hide()">
            <v-icon icon="mdi-close" size="16"></v-icon>
          </button>
        </div>
        <!-- 底部进度条/装饰线 -->
        <div class="toast-progress"></div>
      </div>
    </Transition>
  </v-app>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useMessageStore } from './stores/message'

const messageStore = useMessageStore()

const messageIcon = computed(() => {
  const iconMap: Record<string, string> = {
    success: 'mdi-check-circle-outline',
    error: 'mdi-alert-circle-outline',
    warning: 'mdi-alert-outline',
    info: 'mdi-information-outline'
  }
  return iconMap[messageStore.state.type] || 'mdi-information-outline'
})

// 自动隐藏消息
let hideTimer: number | null = null
watch(() => messageStore.state.visible, (visible) => {
  if (visible) {
    if (hideTimer) {
      clearTimeout(hideTimer)
    }
    hideTimer = window.setTimeout(() => {
      messageStore.hide()
    }, messageStore.state.timeout)
  }
})
</script>

<style scoped>
.message-toast {
  position: fixed;
  top: 80px; /* 避开导航栏 */
  right: 24px; /* 改为右上角显示，更符合桌面端习惯 */
  left: auto;
  transform: none;
  z-index: 9999;
  min-width: 300px;
  max-width: 400px;
  border-radius: 8px;
  backdrop-filter: blur(12px);
  overflow: hidden;
}

.glass-toast {
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.5);
}

.message-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  position: relative;
  z-index: 2;
}

.message-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
}

.message-text {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.5;
  color: #E0F2F7;
  letter-spacing: 0.5px;
}

.message-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: rgba(224, 242, 247, 0.5);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.message-close:hover {
  color: #fff;
  transform: rotate(90deg);
}

/* 类型样式变体 */
.message-toast--success {
  border-color: rgba(34, 197, 94, 0.5);
  box-shadow: 0 0 20px rgba(34, 197, 94, 0.15);
}
.message-toast--success .message-icon-box {
  color: #4ade80;
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.3);
}
.message-toast--success .toast-progress {
  background: #4ade80;
  box-shadow: 0 0 10px #4ade80;
}

.message-toast--error {
  border-color: rgba(239, 68, 68, 0.5);
  box-shadow: 0 0 20px rgba(239, 68, 68, 0.15);
}
.message-toast--error .message-icon-box {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
}
.message-toast--error .toast-progress {
  background: #f87171;
  box-shadow: 0 0 10px #f87171;
}

.message-toast--warning {
  border-color: rgba(249, 115, 22, 0.5);
  box-shadow: 0 0 20px rgba(249, 115, 22, 0.15);
}
.message-toast--warning .message-icon-box {
  color: #fb923c;
  background: rgba(249, 115, 22, 0.1);
  border-color: rgba(249, 115, 22, 0.3);
}
.message-toast--warning .toast-progress {
  background: #fb923c;
  box-shadow: 0 0 10px #fb923c;
}

.message-toast--info {
  border-color: rgba(0, 240, 255, 0.5);
  box-shadow: 0 0 20px rgba(0, 240, 255, 0.15);
}
.message-toast--info .message-icon-box {
  color: #00F0FF;
  background: rgba(0, 240, 255, 0.1);
  border-color: rgba(0, 240, 255, 0.3);
}
.message-toast--info .toast-progress {
  background: #00F0FF;
  box-shadow: 0 0 10px #00F0FF;
}

/* 底部装饰线 */
.toast-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 2px;
  width: 100%;
  animation: progress 3s linear forwards;
}

@keyframes progress {
  from { width: 100%; }
  to { width: 0%; }
}

/* 进出场动画 */
.message-slide-enter-active,
.message-slide-leave-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.message-slide-enter-from {
  opacity: 0;
  transform: translateX(50px) scale(0.95);
}

.message-slide-leave-to {
  opacity: 0;
  transform: translateX(50px) scale(0.95);
}

.font-mono {
  font-family: 'Space Mono', monospace;
}
</style>
