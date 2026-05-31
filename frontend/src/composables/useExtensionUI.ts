import { ref } from 'vue'
import type { ExtensionNotification, ExtensionUIRequest } from '../types'

const activeRequest = ref<ExtensionUIRequest | null>(null)
const queue = ref<ExtensionUIRequest[]>([])
const notifications = ref<ExtensionNotification[]>([])

const blockingMethods = new Set(['select', 'confirm', 'input', 'editor'])

function app() {
  return window.go?.main?.App
}

async function sendResponse(id: string, value = '', confirmed = false, cancelled = false) {
  await app()?.RespondToExtensionUI(id, value, confirmed, cancelled)
}

function showNext() {
  activeRequest.value = queue.value.shift() || null
}

export function useExtensionUI() {
  async function handleRequest(request: ExtensionUIRequest) {
    if (request.method === 'notify') {
      const notice = {
        id: request.id,
        message: request.message || request.title || '',
        type: request.notifyType || 'info',
      }
      notifications.value.push(notice)
      window.setTimeout(() => removeNotification(notice.id), 4500)
      return
    }
    if (!blockingMethods.has(request.method)) {
      await sendResponse(request.id, '', false, true)
      return
    }
    if (!activeRequest.value) {
      activeRequest.value = request
    } else {
      queue.value.push(request)
    }
  }

  async function respond(value = '', confirmed = true, cancelled = false) {
    const current = activeRequest.value
    if (!current) return
    activeRequest.value = null
    await sendResponse(current.id, value, confirmed, cancelled)
    showNext()
  }

  async function cancel() {
    await respond('', false, true)
  }

  function removeNotification(id: string) {
    notifications.value = notifications.value.filter(item => item.id !== id)
  }

  function reset() {
    activeRequest.value = null
    queue.value = []
    notifications.value = []
  }

  async function clearPending() {
    const requests = [activeRequest.value, ...queue.value].filter(Boolean) as ExtensionUIRequest[]
    reset()
    await Promise.all(requests.map(request => sendResponse(request.id, '', false, true)))
  }

  return {
    activeRequest,
    notifications,
    handleRequest,
    respond,
    cancel,
    removeNotification,
    reset,
    clearPending,
  }
}
