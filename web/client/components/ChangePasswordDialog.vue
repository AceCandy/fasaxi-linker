<template>
  <v-dialog v-model="dialog" max-width="500px" persistent>
    <v-card class="glass-card">
      <v-card-title class="text-h5 text-primary-glow">
        修改密码
      </v-card-title>

      <v-card-text>
        <v-form ref="formRef" @submit.prevent="handleSubmit">
          <v-text-field
            v-model="oldPassword"
            label="旧密码"
            type="password"
            variant="outlined"
            prepend-inner-icon="mdi-lock-outline"
            :error-messages="oldPasswordError"
            @input="clearErrors"
            class="mb-2"
          ></v-text-field>

          <v-text-field
            v-model="newPassword"
            label="新密码"
            type="password"
            variant="outlined"
            prepend-inner-icon="mdi-lock"
            :error-messages="newPasswordError"
            @input="clearErrors"
            class="mb-2"
          ></v-text-field>

          <v-text-field
            v-model="confirmPassword"
            label="确认新密码"
            type="password"
            variant="outlined"
            prepend-inner-icon="mdi-lock-check"
            :error-messages="confirmPasswordError"
            @input="clearErrors"
          ></v-text-field>

          <v-alert
            v-if="errorMessage"
            type="error"
            variant="tonal"
            density="compact"
            class="mt-4"
          >
            {{ errorMessage }}
          </v-alert>

          <v-alert
            v-if="successMessage"
            type="success"
            variant="tonal"
            density="compact"
            class="mt-4"
          >
            {{ successMessage }}
          </v-alert>
        </v-form>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          variant="text"
          @click="handleClose"
          :disabled="loading"
        >
          取消
        </v-btn>
        <v-btn
          color="primary"
          variant="elevated"
          @click="handleSubmit"
          :loading="loading"
          class="btn-neon"
        >
          确认修改
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const authStore = useAuthStore()

const dialog = ref(props.modelValue)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const oldPasswordError = ref('')
const newPasswordError = ref('')
const confirmPasswordError = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const loading = ref(false)

watch(() => props.modelValue, (val) => {
  dialog.value = val
  if (val) {
    resetForm()
  }
})

watch(dialog, (val) => {
  emit('update:modelValue', val)
})

function clearErrors() {
  oldPasswordError.value = ''
  newPasswordError.value = ''
  confirmPasswordError.value = ''
  errorMessage.value = ''
  successMessage.value = ''
}

function resetForm() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  clearErrors()
}

function validateForm(): boolean {
  clearErrors()
  let isValid = true

  if (!oldPassword.value) {
    oldPasswordError.value = '请输入旧密码'
    isValid = false
  }

  if (!newPassword.value) {
    newPasswordError.value = '请输入新密码'
    isValid = false
  } else if (newPassword.value.length < 6) {
    newPasswordError.value = '新密码至少需要6个字符'
    isValid = false
  }

  if (!confirmPassword.value) {
    confirmPasswordError.value = '请确认新密码'
    isValid = false
  } else if (newPassword.value !== confirmPassword.value) {
    confirmPasswordError.value = '两次输入的密码不一致'
    isValid = false
  }

  return isValid
}

async function handleSubmit() {
  if (!validateForm()) {
    return
  }

  loading.value = true
  try {
    await authStore.changePassword(oldPassword.value, newPassword.value)
    successMessage.value = '密码修改成功！'
    
    // 2秒后关闭对话框
    setTimeout(() => {
      handleClose()
    }, 2000)
  } catch (error: unknown) {
    if (error instanceof Error) {
      errorMessage.value = error.message || '修改密码失败，请重试'
    } else {
      errorMessage.value = '修改密码失败，请重试'
    }
  } finally {
    loading.value = false
  }
}

function handleClose() {
  dialog.value = false
}
</script>

<style scoped>
:deep(.v-field__outline) {
  --v-field-border-opacity: 0.3;
}

:deep(.v-field--focused .v-field__outline) {
  --v-field-border-opacity: 1;
  border-color: var(--color-primary) !important;
}
</style>
