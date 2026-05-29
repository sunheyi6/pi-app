import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  ChatMessage,
  SessionInfo,
  ModelInfo,
  AppState,
  ThinkingLevel,
  MessageContent,
  TextContent,
  ThinkingContent,
  ToolCallContent,
} from '../types'

export const useChatStore = defineStore('chat', () => {
  // ========== 状态 ==========
  const messages = ref<ChatMessage[]>([])
  const sessions = ref<SessionInfo[]>([])
  const currentSession = ref<SessionInfo | null>(null)
  const isStreaming = ref(false)
  const isAgentRunning = ref(false)
  const streamingMessageId = ref<string | null>(null)
  const currentThinkingBlock = ref<string>('')
  const currentToolCalls = ref<Map<string, { name: string; args: Record<string, any>; output: string; isError: boolean }>>(new Map())
  const pendingSteering = ref<string[]>([])
  const pendingFollowUp = ref<string[]>([])
  // 消息输入队列：AI 回答期间用户输入的消息会排队，回答结束后自动发送
  const inputQueue = ref<string[]>([])
  // 输入框聚焦触发器：递增时 InputBox 自动聚焦
  const focusInputCounter = ref(0)
  const appState = ref<AppState>({
    isStreaming: false,
    isCompacting: false,
    steeringMode: 'one-at-a-time',
    followUpMode: 'one-at-a-time',
    autoCompactionEnabled: true,
    model: null,
    thinkingLevel: 'medium',
    sessionFile: null,
    sessionId: '',
    messageCount: 0,
    pendingMessageCount: 0,
  })

  // ========== 计算属性 ==========
  const lastAssistantMessage = computed(() => {
    for (let i = messages.value.length - 1; i >= 0; i--) {
      if (messages.value[i].role === 'assistant') {
        return messages.value[i]
      }
    }
    return null
  })

  // ========== 消息管理 ==========
  function addMessage(message: ChatMessage) {
    messages.value.push(message)
  }

  function updateMessage(messageId: string, updater: (msg: ChatMessage) => void) {
    const idx = messages.value.findIndex(m => m.id === messageId)
    if (idx !== -1) {
      updater(messages.value[idx])
    }
  }

  // 获取或创建当前的流式助手消息
  function getOrCreateStreamingMessage(): ChatMessage {
    if (streamingMessageId.value) {
      const msg = messages.value.find(m => m.id === streamingMessageId.value)
      if (msg) return msg
    }

    const id = `streaming-${Date.now()}`
    const msg: ChatMessage = {
      id,
      role: 'assistant',
      content: [],
      timestamp: Date.now(),
    }
    messages.value.push(msg)
    streamingMessageId.value = id
    return msg
  }

  // 添加文本 delta
  function appendTextDelta(delta: string, contentIndex: number) {
    const msg = getOrCreateStreamingMessage()
    // 确保有足够的 content 槽位
    while (msg.content.length <= contentIndex) {
      msg.content.push({ type: 'text', text: '' })
    }
    const textContent = msg.content[contentIndex] as TextContent
    textContent.text += delta
  }

  // 开始思考块
  function appendThinkingDelta(delta: string) {
    currentThinkingBlock.value += delta
  }

  // 结束思考块 - 将思考内容添加到消息
  function finalizeThinkingBlock() {
    if (currentThinkingBlock.value) {
      const msg = getOrCreateStreamingMessage()
      msg.content.push({
        type: 'thinking',
        thinking: currentThinkingBlock.value,
      })
      currentThinkingBlock.value = ''
    }
  }

  // 工具调用开始
  function startToolCall(toolCallId: string, toolName: string, args: Record<string, any>) {
    currentToolCalls.value.set(toolCallId, {
      name: toolName,
      args,
      output: '',
      isError: false,
    })

    const msg = getOrCreateStreamingMessage()
    msg.content.push({
      type: 'toolCall',
      id: toolCallId,
      name: toolName,
      arguments: args,
    })
  }

  // 工具调用输出更新
  function updateToolOutput(toolCallId: string, output: string) {
    const tc = currentToolCalls.value.get(toolCallId)
    if (tc) {
      tc.output = output
    }
  }

  // 工具调用结束
  function endToolCall(toolCallId: string, isError: boolean) {
    const tc = currentToolCalls.value.get(toolCallId)
    if (tc) {
      tc.isError = isError
    }

    // 将工具结果作为独立消息添加
    if (tc) {
      messages.value.push({
        id: `toolresult-${toolCallId}`,
        role: 'toolResult',
        content: [{ type: 'text', text: tc.output }],
        timestamp: Date.now(),
        toolCallId,
        toolName: tc.name,
        isError,
      })
    }
  }

  // 完成流式消息
  function finalizeStreamingMessage(model?: string, provider?: string, usage?: any, stopReason?: string) {
    if (streamingMessageId.value) {
      updateMessage(streamingMessageId.value, (msg) => {
        if (model) msg.model = model
        if (provider) msg.provider = provider
        if (usage) msg.usage = usage
        if (stopReason) msg.stopReason = stopReason
      })
      streamingMessageId.value = null
    }
    currentThinkingBlock.value = ''
    currentToolCalls.value.clear()
  }

  // ========== 会话管理 ==========
  function setSessions(list: SessionInfo[]) {
    sessions.value = list
  }

  function setCurrentSession(session: SessionInfo | null) {
    currentSession.value = session
  }

  function clearMessages() {
    messages.value = []
    streamingMessageId.value = null
    currentThinkingBlock.value = ''
    currentToolCalls.value.clear()
  }

  // ========== 状态 ==========
  function setStreaming(val: boolean) {
    isStreaming.value = val
    appState.value.isStreaming = val
  }

  function setAgentRunning(val: boolean) {
    isAgentRunning.value = val
  }

  function updateAppState(state: Partial<AppState>) {
    Object.assign(appState.value, state)
  }

  function setPendingSteering(list: string[]) {
    pendingSteering.value = list
  }

  function setPendingFollowUp(list: string[]) {
    pendingFollowUp.value = list
  }

  // ========== 输入队列 ==========
  function enqueueInput(text: string) {
    inputQueue.value.push(text)
  }

  function dequeueInput(): string | undefined {
    return inputQueue.value.shift()
  }

  function clearInputQueue() {
    inputQueue.value = []
  }

  function removeFromQueue(index: number) {
    inputQueue.value.splice(index, 1)
  }

  // ========== 输入框聚焦 ==========
  function requestFocusInput() {
    focusInputCounter.value++
  }

  return {
    // 状态
    messages,
    sessions,
    currentSession,
    isStreaming,
    isAgentRunning,
    streamingMessageId,
    currentThinkingBlock,
    currentToolCalls,
    pendingSteering,
    pendingFollowUp,
    appState,

    // 计算属性
    lastAssistantMessage,

    // 消息方法
    addMessage,
    updateMessage,
    appendTextDelta,
    appendThinkingDelta,
    finalizeThinkingBlock,
    startToolCall,
    updateToolOutput,
    endToolCall,
    finalizeStreamingMessage,
    clearMessages,

    // 会话方法
    setSessions,
    setCurrentSession,

    // 状态方法
    setStreaming,
    setAgentRunning,
    updateAppState,
    setPendingSteering,
    setPendingFollowUp,

    // 输入队列
    inputQueue,
    enqueueInput,
    dequeueInput,
    clearInputQueue,
    removeFromQueue,

    // 聚焦
    focusInputCounter,
    requestFocusInput,
  }
})
