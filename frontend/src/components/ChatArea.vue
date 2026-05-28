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
const showThinking = ref(true)
const showTools = ref(true)
const shouldAutoScroll = ref(true)

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

    <!-- 消息列表 -->
    <div
      ref="chatContainer"
      @scroll="handleScroll"
      class="flex-1 min-h-0 overflow-y-auto px-5 py-6"
    >
      <!-- 欢迎页 -->
      <div v-if="showWelcome && messages.length === 0" class="flex items-center justify-center h-full">
        <div class="text-center max-w-sm">
          <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-500
                      flex items-center justify-center mx-auto mb-6
                      shadow-2xl shadow-blue-500/20">
            <span class="text-white text-2xl font-bold">π</span>
          </div>
          <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-2 tracking-tight">Pi Desktop</h2>
          <p class="text-gray-400 dark:text-gray-500 text-sm leading-relaxed mb-8">
            基于 pi coding agent 的桌面 AI 编程助手
          </p>
          <div class="flex flex-wrap justify-center gap-1.5">
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">📖 代码阅读</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">✏️ 精准编辑</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">⚡ 命令执行</span>
            <span class="text-[11px] text-gray-500 dark:text-gray-600 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-white/[0.04]">🔍 搜索项目</span>
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
          <ChatMessage
            v-if="msg.role === 'user' || msg.role === 'assistant'"
            :message="msg"
            :is-streaming="msg.id === store.streamingMessageId && isStreaming"
          />
        </template>

        <!-- 流式思考 -->
        <ThinkingBlock
          v-if="showThinking && store.currentThinkingBlock && isStreaming"
          :content="store.currentThinkingBlock"
          :is-streaming="true"
        />
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
