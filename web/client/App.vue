<template>
  <v-app>
    <router-view />
    
    <!-- 全局消息提示 - 现代化设计 -->
    <Transition name="message-slide">
      <div v-if="messageStore.state.visible" class="message-toast" :class="`message-toast--${messageStore.state.type}`">
        <div class="message-content">
          <div class="message-icon">
            <v-icon :icon="messageIcon" size="20"></v-icon>
          </div>
          <span class="message-text">{{ messageStore.state.text }}</span>
          <button class="message-close" @click="messageStore.hide()">
            <v-icon icon="mdi-close" size="18"></v-icon>
          </button>
        </div>
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
    success: 'mdi-check-circle',
    error: 'mdi-alert-circle',
    warning: 'mdi-alert',
    info: 'mdi-information'
  }
  return iconMap[messageStore.state.type] || 'mdi-information'
})

// 自动隐藏消息
let hideTimer: number | null = null
watch(() => messageStore.state.visible, (visible) => {
  if (visible) {
    // 清除之前的定时器
    if (hideTimer) {
      clearTimeout(hideTimer)
    }
    // 设置新的定时器
    hideTimer = window.setTimeout(() => {
      messageStore.hide()
    }, messageStore.state.timeout)
  }
})
</script>

<style scoped>
.message-toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  min-width: 320px;
  max-width: 500px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12), 0 2px 8px rgba(0, 0, 0, 0.08);
  backdrop-filter: blur(10px);
  animation: message-appear 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
}

.message-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
}

.message-text {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.5;
}

.message-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.message-close:hover {
  background: rgba(0, 0, 0, 0.08);
}

/* Success 样式 */
.message-toast--success {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.95) 0%, rgba(22, 163, 74, 0.95) 100%);
  color: white;
}

.message-toast--success .message-icon {
  background: rgba(255, 255, 255, 0.2);
}

/* Error 样式 */
.message-toast--error {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.95) 0%, rgba(220, 38, 38, 0.95) 100%);
  color: white;
}

.message-toast--error .message-icon {
  background: rgba(255, 255, 255, 0.2);
}

/* Warning 样式 */
.message-toast--warning {
  background: linear-gradient(135deg, rgba(251, 146, 60, 0.95) 0%, rgba(249, 115, 22, 0.95) 100%);
  color: white;
}

.message-toast--warning .message-icon {
  background: rgba(255, 255, 255, 0.2);
}

/* Info 样式 */
.message-toast--info {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.95) 0%, rgba(37, 99, 235, 0.95) 100%);
  color: white;
}

.message-toast--info .message-icon {
  background: rgba(255, 255, 255, 0.2);
}

/* 动画 */
.message-slide-enter-active,
.message-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-slide-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}

.message-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}

@keyframes message-appear {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}
</style>
