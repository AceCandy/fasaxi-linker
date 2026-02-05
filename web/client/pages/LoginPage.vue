<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="login-background">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card glass-card">
      <div class="login-header">
        <div class="logo-container">
          <img src="/logo.jpg" alt="Linker Logo" class="login-logo" />
        </div>
        <h1 class="login-title text-primary-glow">Linker</h1>
        <p class="login-subtitle">智能硬链接管理系统</p>
      </div>

      <v-form @submit.prevent="handleLogin" class="login-form">
        <v-text-field
          v-model="username"
          label="用户名"
          variant="outlined"
          prepend-inner-icon="mdi-account"
          class="glass-input-field"
          :error-messages="usernameError"
          @input="clearErrors"
          autofocus
        ></v-text-field>

        <v-text-field
          v-model="password"
          label="密码"
          variant="outlined"
          prepend-inner-icon="mdi-lock"
          :type="showPassword ? 'text' : 'password'"
          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showPassword = !showPassword"
          class="glass-input-field"
          :error-messages="passwordError"
          @input="clearErrors"
        ></v-text-field>

        <v-alert
          v-if="errorMessage"
          type="error"
          variant="tonal"
          density="compact"
          class="mb-4 error-alert"
        >
          {{ errorMessage }}
        </v-alert>

        <v-btn
          type="submit"
          block
          size="large"
          :loading="loading"
          class="login-btn btn-neon"
        >
          <v-icon start>mdi-login</v-icon>
          登录
        </v-btn>
      </v-form>

      <div class="login-footer">
        <p class="footer-text">© 2024 Linker - 硬链接管理工具</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const errorMessage = ref('')
const usernameError = ref('')
const passwordError = ref('')

function clearErrors() {
  errorMessage.value = ''
  usernameError.value = ''
  passwordError.value = ''
}

async function handleLogin() {
  clearErrors()

  // Validation
  if (!username.value.trim()) {
    usernameError.value = '请输入用户名'
    return
  }
  if (!password.value) {
    passwordError.value = '请输入密码'
    return
  }

  loading.value = true

  try {
    await authStore.login(username.value, password.value)
    router.push('/')
  } catch (error: unknown) {
    if (error instanceof Error) {
      errorMessage.value = error.message || '登录失败，请重试'
    } else {
      errorMessage.value = '登录失败，请重试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  position: relative;
  overflow: hidden;
  background: var(--app-bg-gradient);
}

.login-background {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: float 20s ease-in-out infinite;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, var(--color-primary) 0%, transparent 70%);
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, var(--color-secondary) 0%, transparent 70%);
  bottom: -150px;
  right: -150px;
  animation-delay: -7s;
}

.orb-3 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, var(--color-accent) 0%, transparent 70%);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -14s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  25% {
    transform: translate(30px, -30px) scale(1.05);
  }
  50% {
    transform: translate(-20px, 20px) scale(0.95);
  }
  75% {
    transform: translate(-30px, -20px) scale(1.02);
  }
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 40px;
  position: relative;
  z-index: 1;
  border-radius: 20px !important;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-container {
  width: 80px;
  height: 80px;
  margin: 0 auto 16px;
  background: rgba(15, 23, 42, 0.8);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: 0 0 20px rgba(0, 240, 255, 0.3);
  border: 1px solid rgba(0, 240, 255, 0.3);
  transition: transform 0.3s ease;
}

.logo-container:hover {
  transform: scale(1.05);
  box-shadow: 0 0 30px rgba(0, 240, 255, 0.5);
}

.login-logo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.login-title {
  font-family: 'Orbitron', sans-serif;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 2px;
  margin-bottom: 8px;
}

.login-subtitle {
  color: var(--color-text-muted);
  font-family: 'Space Mono', monospace;
  font-size: 0.875rem;
}

.login-form {
  margin-bottom: 24px;
}

.login-form .v-text-field {
  margin-bottom: 16px;
}

.error-alert {
  border-radius: 8px !important;
}

.login-btn {
  height: 48px !important;
  font-size: 1rem !important;
  font-weight: 600 !important;
  letter-spacing: 1px;
  margin-top: 8px;
}

.login-footer {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.footer-text {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  font-family: 'Space Mono', monospace;
}

/* Override Vuetify input styles for glass effect */
:deep(.v-field__outline) {
  --v-field-border-opacity: 0.3;
}

:deep(.v-field--focused .v-field__outline) {
  --v-field-border-opacity: 1;
  border-color: var(--color-primary) !important;
}

:deep(.v-field__input) {
  color: var(--color-text) !important;
}

:deep(.v-label) {
  color: var(--color-text-muted) !important;
}

:deep(.v-icon) {
  color: var(--color-text-muted) !important;
}

:deep(.v-field--focused .v-icon) {
  color: var(--color-primary) !important;
}
</style>

