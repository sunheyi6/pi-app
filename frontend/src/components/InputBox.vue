<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useChatStore } from '../stores/chatStore'

const props = defineProps<{
  isStreaming: boolean
  isAgentRunning: boolean
  pendingInfo?: string
}>()

const emit = defineEmits<{
  send: [message: string]
  abort: []
}>()

const store = useChatStore()

const inputText = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const isSending = ref(false)
let lastSendTime = 0
let sendingTimer: ReturnType<typeof setTimeout> | null = null

const canSend = computed(() => inputText.value.trim().length > 0 && !isSending.value)

function handleSend() {
  const now = Date.now()
  if (now - lastSendTime < 500) return

  const text = inputText.value.trim()
  if (!text || isSending.value) return

  lastSendTime = now

  // AI 正在回答 → 消息入队列
  // 兜底保护：如果没有实际流式消息 ID，说明状态异常，直接发送
  if ((props.isStreaming || props.isAgentRunning) && store.streamingMessageId) {
    store.enqueueInput(text)
    inputText.value = ''
    nextTick(() => {
      autoResize()
      focusInput()
    })
    return
  }

  isSending.value = true

  // 安全超时：防止极端情况下 isSending 卡住
  if (sendingTimer) clearTimeout(sendingTimer)
  sendingTimer = setTimeout(() => {
    isSending.value = false
    sendingTimer = null
    focusInput()
  }, 3000)

  try {
    emit('send', text)
    inputText.value = ''
  } finally {
    nextTick(() => {
      if (sendingTimer) clearTimeout(sendingTimer)
      sendingTimer = null
      isSending.value = false
      focusInput()
    })
  }
}

function handleAbort() {
  emit('abort')
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
  if (e.key === 'Escape') {
    if (props.isStreaming) {
      e.preventDefault()
      handleAbort()
    }
  }
}

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 180) + 'px'
}

function focusInput() {
  // WebView2 中需要延迟聚焦以确保焦点正确转移
  nextTick(() => {
    textareaRef.value?.focus()
  })
}

// 点击输入区域时强制聚焦 textarea（解决 WebView2 焦点问题）
function handleContainerClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  // 不拦截按钮点击
  if (target.tagName === 'BUTTON' || target.closest('button')) return
  // 如果点击的不是 textarea 本身，则聚焦 textarea
  if (e.target !== textareaRef.value) {
    e.preventDefault()
    textareaRef.value?.focus()
  }
}

onMounted(() => {
  // WebView2 渲染较慢，双重延迟确保 DOM 就绪
  nextTick(() => {
    setTimeout(() => focusInput(), 100)
  })
  window.addEventListener('focus', focusInput)
})

onUnmounted(() => {
  window.removeEventListener('focus', focusInput)
})

// 监听聚焦请求（新建会话、启动完成等场景）
watch(() => store.focusInputCounter, () => {
  nextTick(() => {
    setTimeout(() => focusInput(), 100)
  })
})
</script>

<template>
  <div class="border-t border-gray-200 dark:border-white/5 bg-white dark:bg-[#0a0a0a] px-4 py-3 shrink-0">
    <div class="max-w-3xl mx-auto">
      <!-- 输入队列提示 -->
      <div
        v-if="store.inputQueue.length > 0"
        class="mb-2 flex flex-col gap-1"
      >
        <div
          v-for="(item, idx) in store.inputQueue"
          :key="idx"
          class="text-[15px] text-amber-500 dark:text-amber-400 flex items-center gap-2 animate-fade-in
                 bg-amber-500/5 rounded-lg px-3 py-1.5 border border-amber-500/10"
        >
          <svg class="w-3 h-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="flex-1 truncate">{{ item }}</span>
          <span class="text-[10px] text-amber-500/50 shrink-0">#{{ idx + 1 }}</span>
          <button
            @click="store.removeFromQueue(idx)"
            class="w-5 h-5 flex items-center justify-center rounded-full
                   text-amber-500/40 hover:text-red-400 hover:bg-red-500/10
                   transition-all duration-150 shrink-0"
            title="移除此条"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- 待处理引导/跟进消息提示 -->
      <div
        v-if="pendingInfo"
        class="mb-2 text-[11px] text-yellow-400/80 flex items-center gap-1.5 animate-fade-in"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ pendingInfo }}
      </div>

      <!-- 输入区域 — Codex style pill -->
      <div
        @mousedown="handleContainerClick"
        class="relative flex items-end gap-2 p-2 bg-gray-100 dark:bg-[#1c1c1e] rounded-2xl
                border border-gray-300 dark:border-white/10 focus-within:border-gray-400 dark:focus-within:border-white/20
                focus-within:shadow-lg focus-within:shadow-blue-500/5
                transition-all duration-200 cursor-text">
        <textarea
          ref="textareaRef"
          v-model="inputText"
          @keydown="handleKeydown"
          @input="autoResize"
          :disabled="isSending"
          :placeholder="isAgentRunning ? 'Agent 工作中...' : '输入消息...'"
          rows="1"
          class="flex-1 bg-transparent text-[15px] text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500
                 px-1 py-1 resize-none outline-none leading-relaxed
                 disabled:opacity-40 disabled:cursor-not-allowed
                 min-h-[28px] max-h-[180px]"
        />

        <!-- 中止按钮 -->
        <button
          v-if="isStreaming"
          @click="handleAbort"
          class="w-8 h-8 flex items-center justify-center rounded-full
                 bg-red-500/20 hover:bg-red-500/30 text-red-400
                 transition-all duration-150 shrink-0"
          title="中止 (Esc)"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <rect x="6" y="6" width="12" height="12" rx="2" />
          </svg>
        </button>

        <!-- 发送按钮（始终显示，AI 回答时入队列） -->
        <button
          @click="handleSend"
          :disabled="!canSend"
          class="w-8 h-8 flex items-center justify-center rounded-full
                 transition-all duration-200 shrink-0"
          :class="canSend
            ? 'bg-gray-900 hover:bg-gray-700 dark:bg-white dark:hover:bg-gray-200 text-white dark:text-black active:scale-90'
            : 'bg-gray-300 dark:bg-white/10 text-gray-400 dark:text-gray-600 cursor-not-allowed'"
          :title="isStreaming ? '排队发送 (Enter)' : '发送 (Enter)'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <!-- 底部提示 -->
      <div class="flex items-center justify-between mt-2 px-1 text-[10px] text-gray-400 dark:text-gray-600">
        <div class="flex items-center gap-2.5">
          <span>↵ 发送</span>
          <span>⇧↵ 换行</span>
          <span>Esc 中止</span>
        </div>
        <div v-if="isAgentRunning" class="flex items-center gap-1.5 text-blue-500 dark:text-blue-400">
          <span class="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
          运行中
        </div>
      </div>
    </div>
  </div>
</template>
