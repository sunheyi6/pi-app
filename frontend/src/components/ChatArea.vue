<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import ChatMessage from './ChatMessage.vue'
import InputBox from './InputBox.vue'
import ThinkingBlock from './ThinkingBlock.vue'
import ToolCallCard from './ToolCallCard.vue'
import type { ChatMessage as ChatMsgType, AppState, ToolCallContent } from '../types'
import { useChatStore } from '../stores/chatStore'

const props = defineProps<{
  messages: ChatMsgType[]
  isStreaming: boolean
  isAgentRunning: boolean
  showWelcome: boolean
  appState: AppState
}>()

const emit = defineEmits<{
  send: [message: string]
  abort: []
}>()

const store = useChatStore()
const chatContainer = ref<HTMLElement | null>(null)
const showThinking = ref(false)
const showTools = ref(false)
const shouldAutoScroll = ref(true)
const activeAnchor = ref<string | null>(null)

// 欢迎页示例问题
const examples = [
  '这个项目的结构是怎样的？',
  '帮我分析 frontend/src/App.vue 的代码',
  '搜索一下 React 19 的新特性',
  '帮我重构 ChatArea 组件',
]

// 收集所有用户消息作为目录项
const userMessages = computed(() =>
  props.messages.filter(m => m.role === 'user')
)

// 提取用户消息的文本摘要
function messageSummary(msg: ChatMsgType): string {
  const textBlock = msg.content.find(c => c.type === 'text')
  if (!textBlock || !('text' in textBlock)) return '...'
  const text = textBlock.text
  return text.length > 25 ? text.slice(0, 25) + '…' : text
}

// 滚动到指定消息
function scrollToMessage(msgId: string) {
  const el = document.getElementById(`msg-${msgId}`)
  if (el) {
    shouldAutoScroll.value = false
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeAnchor.value = msgId
  }
}

// 根据滚动位置更新当前锚点
function updateActiveAnchor() {
  if (!chatContainer.value) return
  const container = chatContainer.value
  const containerTop = container.scrollTop + 100

  let closest: string | null = null
  let closestDist = Infinity

  for (const el of container.querySelectorAll('[data-anchor]')) {
    const anchorId = el.getAttribute('data-anchor')
    if (!anchorId) continue
    const rect = (el as HTMLElement).getBoundingClientRect()
    const containerRect = container.getBoundingClientRect()
    const dist = rect.top - containerRect.top
    if (dist >= -50 && dist < closestDist) {
      closest = anchorId
      closestDist = dist
    }
  }

  if (closest) activeAnchor.value = closest
}

function scrollToBottom(force = false) {
  if (!chatContainer.value) return
  if (!force && !shouldAutoScroll.value) return
  chatContainer.value.scrollTop = chatContainer.value.scrollHeight
}

// 检测用户手动上滑
function handleScroll() {
  if (!chatContainer.value) return
  const { scrollTop, scrollHeight, clientHeight } = chatContainer.value
  shouldAutoScroll.value = scrollHeight - scrollTop - clientHeight < 60
  updateActiveAnchor()
}

watch(
  () => [props.messages.length],
  () => nextTick(() => scrollToBottom(true)),
)

watch(
  () => {
    const msg = props.messages.find(m => m.id === store.streamingMessageId)
    if (msg) {
      const last = msg.content[msg.content.length - 1]
      return last?.type === 'text' ? (last as any).text?.length : 0
    }
    return store.currentThinkingBlock?.length || 0
  },
  () => nextTick(() => scrollToBottom()),
)

function handleSend(message: string) {
  shouldAutoScroll.value = true
  emit('send', message)
}

function handleAbort() {
  emit('abort')
}

const pendingInfo = computed(() => {
  const parts: string[] = []
  if (store.pendingSteering.length > 0) parts.push(`${store.pendingSteering.length} 条引导消息`)
  if (store.pendingFollowUp.length > 0) parts.push(`${store.pendingFollowUp.length} 条跟进消息`)
  return parts.join('，')
})
</script>

<template>
  <div class="flex-1 flex flex-col min-w-0 min-h-0 bg-apple-base">
    <!-- 顶栏 -->
    <div class="h-11 border-b border-gray-200 dark:border-white/5 flex items-center justify-between px-5 bg-white dark:bg-[#0a0a0a]
                shrink-0 z-10">
      <div class="flex items-center gap-1">
        <!-- 显示选项 -->
        <button
          @click="showThinking = !showThinking"
          class="text-[11px] px-2.5 py-1.5 rounded-full transition-all duration-200"
          :class="showThinking
            ? 'bg-purple-500/15 text-purple-500 dark:text-purple-400'
            : 'text-gray-400 dark:text-gray-600 hover:text-gray-600 dark:hover:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/[0.04]'"
        >思考</button>
        <button
          @click="showTools = !showTools"
          class="text-[11px] px-2.5 py-1.5 rounded-full transition-all duration-200"
          :class="showTools
            ? 'bg-blue-500/15 text-blue-500 dark:text-blue-400'
            : 'text-gray-400 dark:text-gray-600 hover:text-gray-600 dark:hover:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/[0.04]'"
        >工具</button>

        <span v-if="appState.messageCount > 0" class="text-[10px] text-gray-400 dark:text-gray-600 ml-2">
          {{ appState.messageCount }} 条
        </span>
      </div>
    </div>

    <!-- 消息列表 + 右侧目录 -->
    <div class="flex-1 min-h-0 flex relative">
      <div
        ref="chatContainer"
        @scroll="handleScroll"
        class="flex-1 min-h-0 overflow-y-auto px-5 py-6"
      >
      <!-- 欢迎页 -->
      <div v-if="showWelcome && messages.length === 0" class="flex flex-col items-center justify-center h-full px-4">
        <div class="text-center max-w-lg w-full">
          <!-- Logo -->
          <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-500
                      flex items-center justify-center mx-auto mb-6
                      shadow-2xl shadow-blue-500/20">
            <span class="text-white text-2xl font-bold">π</span>
          </div>
          <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-1 tracking-tight">Pi Desktop</h2>
          <p class="text-gray-400 dark:text-gray-500 text-sm mb-8">
            基于 pi coding agent 的桌面 AI 编程助手
          </p>

          <!-- 输入引导提示 -->
          <div class="bg-gray-50 dark:bg-white/[0.03] rounded-2xl border border-gray-200 dark:border-white/[0.06] p-4 mb-6">
            <div class="flex items-center gap-2 mb-3">
              <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span class="text-[13px] font-medium text-gray-600 dark:text-gray-300">在下方输入你的问题</span>
            </div>
            <p class="text-[12px] text-gray-400 dark:text-gray-500 leading-relaxed">
              你可以问我代码问题、让我修改文件、执行命令，或者搜索互联网上的信息。
            </p>
          </div>

          <!-- 示例问题 -->
          <div class="text-left space-y-2">
            <p class="text-[11px] font-medium text-gray-400 dark:text-gray-600 uppercase tracking-wider px-1">试试这样问</p>
            <button
              v-for="example in examples"
              :key="example"
              @click="handleSend(example)"
              class="w-full text-left text-[13px] text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200
                     bg-gray-50 dark:bg-white/[0.02] hover:bg-gray-100 dark:hover:bg-white/[0.06]
                     rounded-xl px-4 py-2.5 transition-all duration-150 border border-transparent hover:border-gray-200 dark:hover:border-white/[0.08]"
            >
              {{ example }}
            </button>
          </div>

          <!-- 能力标签 -->
          <div class="flex flex-wrap justify-center gap-1.5 mt-8">
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">📖 代码阅读</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">✏️ 精准编辑</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">⚡ 命令执行</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">🔍 搜索项目</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">🌐 联网搜索</span>
          </div>
        </div>
      </div>

      <!-- 消息列表 -->
      <div class="max-w-3xl mx-auto space-y-4">
        <template v-for="msg in messages" :key="msg.id">
          <!-- 思考块 -->
          <template v-if="showThinking && msg.role === 'assistant'">
            <template v-for="(block, idx) in msg.content" :key="`think-${msg.id}-${idx}`">
              <ThinkingBlock
                v-if="block.type === 'thinking'"
                :content="(block as any).thinking"
              />
            </template>
          </template>

          <!-- 工具调用 -->
          <template v-if="showTools && msg.role === 'assistant'">
            <template v-for="(block, idx) in msg.content" :key="`tool-${msg.id}-${idx}`">
              <ToolCallCard
                v-if="block.type === 'toolCall'"
                :tool-call="block as ToolCallContent"
                :result="messages.find(m => m.role === 'toolResult' && m.toolCallId === (block as ToolCallContent).id)"
              />
            </template>
          </template>

          <!-- 聊天消息 -->
          <div
            v-if="msg.role === 'user' || msg.role === 'assistant'"
            :id="`msg-${msg.id}`"
            :data-anchor="msg.role === 'user' ? msg.id : undefined"
          >
            <ChatMessage
              :message="msg"
              :is-streaming="msg.id === store.streamingMessageId && isStreaming"
            />
          </div>
        </template>

        <!-- 流式思考 -->
        <ThinkingBlock
          v-if="showThinking && store.currentThinkingBlock && isStreaming"
          :content="store.currentThinkingBlock"
          :is-streaming="true"
        />
      </div>
    </div>

      <!-- 右侧悬浮对话目录 -->
      <div
        v-if="userMessages.length > 1"
        class="w-44 shrink-0 overflow-y-auto border-l border-gray-100 dark:border-white/[0.04]
               bg-gray-50/50 dark:bg-[#0a0a0a]/50 py-3 px-2"
      >
        <div class="text-[10px] font-medium text-gray-400 dark:text-gray-600 uppercase tracking-wider px-2 mb-2">
          对话目录
        </div>
        <button
          v-for="msg in userMessages"
          :key="msg.id"
          @click="scrollToMessage(msg.id)"
          class="w-full text-left px-2.5 py-2 rounded-lg transition-all duration-150 mb-0.5
                 text-[12px] leading-tight"
          :class="activeAnchor === msg.id
            ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 font-medium'
            : 'text-gray-500 dark:text-gray-500 hover:bg-gray-100 dark:hover:bg-white/[0.04] hover:text-gray-700 dark:hover:text-gray-300'"
        >
          {{ messageSummary(msg) }}
        </button>
      </div>
    </div>

    <!-- 输入框 -->
    <InputBox
      :is-streaming="isStreaming"
      :is-agent-running="isAgentRunning"
      :pending-info="pendingInfo"
      @send="handleSend"
      @abort="handleAbort"
    />
  </div>
</template>
