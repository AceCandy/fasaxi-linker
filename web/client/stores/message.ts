import { ref } from 'vue'
import { defineStore } from 'pinia'

export type MessageType = 'success' | 'error' | 'warning' | 'info'

interface MessageState {
  visible: boolean
  text: string
  type: MessageType
  timeout: number
}

export const useMessageStore = defineStore('message', () => {
  const state = ref<MessageState>({
    visible: false,
    text: '',
    type: 'info',
    timeout: 3000
  })

  const show = (text: string, type: MessageType = 'info', timeout = 3000) => {
    state.value = {
      visible: true,
      text,
      type,
      timeout
    }
  }

  const success = (text: string, timeout = 3000) => show(text, 'success', timeout)
  const error = (text: string, timeout = 4000) => show(text, 'error', timeout)
  const warning = (text: string, timeout = 3000) => show(text, 'warning', timeout)
  const info = (text: string, timeout = 3000) => show(text, 'info', timeout)

  const hide = () => {
    state.value.visible = false
  }

  return {
    state,
    show,
    success,
    error,
    warning,
    info,
    hide
  }
})
