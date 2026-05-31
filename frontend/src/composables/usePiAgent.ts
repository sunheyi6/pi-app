import { onMounted, onUnmounted } from 'vue'
import { useChatStore } from '../stores/chatStore'
import type { RPCEvent, ModelInfo, SessionInfo, PackageActionResult, PackageInfo, PackageScope } from '../types'

// 模块级标志：防止事件监听器被注册多次
let eventListenerStarted = false
// 记录已自动命名的会话，防止重复命名
const autoNamedSessionIds = new Set<string>()

// 获取 Go 后端绑定的安全包装
function getApp() {
  if (window.go?.main?.App) {
    return window.go.main.App
  }
  console.warn('[usePiAgent] Wails Go 后端未就绪，运行在纯前端模式')
  return null
}

export function usePiAgent() {
  const store = useChatStore()
  let unsubscribe: (() => void) | null = null

  // ========== 初始化 ==========
  async function init(cwd: string, sessionPath?: string) {
    const app = getApp()
    if (!app) return

    try {
      await app.StartAgent(cwd, sessionPath || '')
      console.log('[usePiAgent] Agent 已启动, session=', sessionPath || '(new)')

      // 获取当前状态并更新 store
      await refreshState()
      // 如果有会话文件，加载历史消息
      if (sessionPath) {
        await loadMessages()
      }
      // 刷新侧边栏会话列表
      await loadSessions()
    } catch (err) {
      console.error('[usePiAgent] 启动 Agent 失败:', err)
      throw err
    }
  }

  // 刷新当前会话状态
  async function refreshState() {
    const app = getApp()
    if (!app) return
    try {
      const json = await app.GetState()
      const state = JSON.parse(json)
      store.updateAppState({
        sessionFile: state.sessionFile || null,
        sessionId: state.sessionId || '',
        sessionName: state.sessionName || '',
        isStreaming: state.isStreaming || false,
        isCompacting: state.isCompacting || false,
        model: state.model || null,
        thinkingLevel: state.thinkingLevel || 'medium',
        messageCount: state.messageCount || 0,
      })
    } catch (e) {
      // 状态获取失败不阻塞
    }
  }

  // 加载当前会话的历史消息
  async function loadMessages() {
    const app = getApp()
    if (!app) return
    try {
      const json = await app.GetMessages()
      const data = JSON.parse(json)
      const msgs = data.messages || []
      store.clearMessages()
      for (const m of msgs) {
        store.addMessage({
          id: m.id || `hist-${Date.now()}-${Math.random().toString(36).slice(2,6)}`,
          role: m.role,
          content: Array.isArray(m.content) ? m.content : [{ type: 'text', text: String(m.content || '') }],
          timestamp: m.timestamp || Date.now(),
          model: m.model,
          provider: m.provider,
          usage: m.usage,
          stopReason: m.stopReason,
          toolCallId: m.toolCallId,
          toolName: m.toolName,
          isError: m.isError,
          command: m.command,
          output: m.output,
          exitCode: m.exitCode,
        })
      }
    } catch (e) {
      // 消息加载失败不阻塞
    }
  }

  // 启动事件监听（模块级保证只注册一次）
  function startEventListener() {
    if (eventListenerStarted) return
    eventListenerStarted = true

    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn('pi-event', (eventStr: string) => {
        try {
          const event: RPCEvent = JSON.parse(eventStr)
          handlePiEvent(event)
        } catch (e) {
          console.error('[usePiAgent] 解析事件失败:', e)
        }
      })
    }
  }

  function stopEventListener() {
    if (window.runtime?.EventsOff) {
      window.runtime.EventsOff('pi-event')
    }
  }

  // ========== RPC 事件处理 ==========
  function handlePiEvent(event: any) {
    switch (event.type) {
      case 'agent_start':
        store.setAgentRunning(true)
        store.setStreaming(true)
        break

      case 'agent_end':
        store.setAgentRunning(false)
        store.setStreaming(false)
        store.finalizeStreamingMessage()
        loadSessions().then(() => autoNameCurrentSession())
        // 处理输入队列：自动发送下一条
        processInputQueue()
        break

      case 'message_start':
        // 新消息开始 - 可能是助手消息
        break

      case 'message_update':
        handleMessageUpdate(event)
        break

      case 'message_end':
        // 消息完成 - 用户消息已在 handleSend 中乐观添加，这里只处理助手消息
        if (event.message?.role === 'assistant') {
          store.finalizeStreamingMessage(
            event.message.model,
            event.message.provider,
            event.message.usage,
            event.message.stopReason
          )
          // pi agent 可能不发送 agent_end，在 assistant 消息完成时尝试自动命名和队列处理
          autoNameCurrentSession()
          processInputQueue()
        }
        break

      case 'turn_end':
        // turn 结束时也尝试自动命名和队列处理（兼容不同事件模型）
        autoNameCurrentSession()
        processInputQueue()
        break

      case 'tool_execution_start':
        store.startToolCall(
          event.toolCallId,
          event.toolName,
          event.args || {}
        )
        break

      case 'tool_execution_update':
        if (event.partialResult?.content) {
          const text = event.partialResult.content
            .map((c: any) => c.text || '')
            .join('\n')
          store.updateToolOutput(event.toolCallId, text)
        }
        break

      case 'tool_execution_end':
        store.endToolCall(event.toolCallId, event.isError || false)
        break

      case 'queue_update':
        store.setPendingSteering(event.steering || [])
        store.setPendingFollowUp(event.followUp || [])
        break

      case 'compaction_start':
        store.updateAppState({ isCompacting: true })
        break

      case 'compaction_end':
        store.updateAppState({ isCompacting: false })
        break

      case 'turn_start':
        break

      default:
        console.log('[usePiAgent] 未处理的事件:', event.type)
    }
  }

  function handleMessageUpdate(event: any) {
    const delta = event.assistantMessageEvent
    if (!delta) return

    switch (delta.type) {
      case 'text_delta':
        store.appendTextDelta(delta.delta || '', delta.contentIndex || 0)
        break

      case 'thinking_delta':
        store.appendThinkingDelta(delta.delta || '')
        break

      case 'thinking_start':
        // 可显示 "思考中..."
        break

      case 'thinking_end':
        store.finalizeThinkingBlock()
        break

      case 'toolcall_start':
        // 工具调用开始已在 tool_execution_start 处理
        break

      case 'toolcall_delta':
        // 工具参数增量
        break

      case 'toolcall_end':
        if (delta.toolCall) {
          store.startToolCall(
            delta.toolCall.id,
            delta.toolCall.name,
            delta.toolCall.arguments || {}
          )
        }
        break

      case 'done':
        // 消息完成
        break

      case 'error':
        console.error('[usePiAgent] 流式错误:', delta.reason)
        break
    }
  }

  // ========== 命令方法 ==========
  async function sendPrompt(message: string, images?: any[]): Promise<string> {
    // 立即将用户消息添加到 store（乐观更新，不需要等 pi 回传）
    store.addMessage({
      id: `user-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      role: 'user',
      content: [{ type: 'text', text: message }],
      timestamp: Date.now(),
    })

    const app = getApp()
    if (!app) {
      throw new Error('后端未就绪，消息未发送到 pi agent')
    }
    return app.SendPrompt(message, images || [])
  }

  async function sendSteer(message: string): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return app.SendSteer(message)
  }

  async function sendFollowUp(message: string): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return app.SendFollowUp(message)
  }

  // 处理输入队列：AI 回答结束后自动发送队列中下一条消息
  let processingQueue = false
  async function processInputQueue() {
    if (processingQueue) {
      console.log('[usePiAgent] 队列处理中，跳过重复调用')
      return
    }
    processingQueue = true
    try {
      const nextMsg = store.dequeueInput()
      if (!nextMsg) return
      console.log('[usePiAgent] 队列取出消息:', nextMsg.slice(0, 30))
      await sendPrompt(nextMsg)
      console.log('[usePiAgent] 队列消息已发送')
    } catch (e: any) {
      console.error('[usePiAgent] 队列消息发送失败:', e)
    } finally {
      processingQueue = false
    }
  }

  async function abort(): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return app.Abort()
  }

  async function setModel(provider: string, modelId: string): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return app.SetModel(provider, modelId)
  }

  async function setThinkingLevel(level: string): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return app.SetThinkingLevel(level)
  }

  async function newSession(): Promise<string> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    const resp = await app.NewSession()
    // 新会话创建成功，刷新状态和列表
    store.clearMessages()
    await refreshState()
    await loadSessions()
    return resp
  }

  async function loadSessions(): Promise<SessionInfo[]> {
    const app = getApp()
    if (!app) return []
    try {
      const json = await app.GetSessions()
      const sessions: SessionInfo[] = JSON.parse(json)
      store.setSessions(sessions)
      return sessions
    } catch (err) {
      console.error('[usePiAgent] 加载会话列表失败:', err)
      return []
    }
  }

  // agent 完成后自动为未命名的会话设置标题
  async function autoNameCurrentSession() {
    const app = getApp()
    if (!app) return

    // 去重：已命名过的会话不再重复命名
    const sessionId = store.appState.sessionId || store.currentSession?.sessionId
    if (sessionId && autoNamedSessionIds.has(sessionId)) {
      return
    }

    // 找到第一条用户消息作为标题素材
    const firstUserMsg = store.messages.find(m => m.role === 'user')
    if (!firstUserMsg) return

    // 提取第一条消息的文本内容
    const textContent = firstUserMsg.content.find(c => c.type === 'text')
    if (!textContent || !('text' in textContent)) return
    const firstMessage = textContent.text
    if (!firstMessage) return

    try {
      const newName = await app.EnsureSessionNamed(firstMessage)
      if (newName) {
        console.log('[usePiAgent] 自动命名会话:', newName)
        if (sessionId) autoNamedSessionIds.add(sessionId)
        await loadSessions() // 刷新列表以显示新标题
      }
    } catch (err) {
      console.error('[usePiAgent] 自动命名失败:', err)
    }
  }

  async function getAvailableModels(): Promise<ModelInfo[]> {
    const app = getApp()
    if (!app) return []
    try {
      const json = await app.GetAvailableModels()
      const data = JSON.parse(json)
      return data.models || []
    } catch (err) {
      console.error('[usePiAgent] 获取模型列表失败:', err)
      return []
    }
  }

  async function selectDirectory(): Promise<string> {
    const app = getApp()
    if (!app) return ''
    return app.SelectDirectory()
  }

  async function getAppInfo(): Promise<any> {
    const app = getApp()
    if (!app) return {}
    try {
      const json = await app.GetAppInfo()
      return JSON.parse(json)
    } catch {
      return {}
    }
  }

  // ========== API Key 管理 ==========
  async function getAuthKeys(): Promise<Record<string, string>> {
    const app = getApp()
    if (!app) return {}
    try {
      const json = await app.GetAuthKeys()
      return JSON.parse(json)
    } catch {
      return {}
    }
  }

  async function setApiKey(provider: string, key: string): Promise<boolean> {
    const app = getApp()
    if (!app) return false
    try {
      const json = await app.SetApiKey(provider, key)
      const result = JSON.parse(json)
      return result.success === true
    } catch {
      return false
    }
  }

  // ========== 扩展包管理 ==========
  async function listPackages(scope: PackageScope): Promise<PackageInfo[]> {
    const app = getApp()
    if (!app) return []
    const json = await app.ListPackages(scope)
    return JSON.parse(json).packages || []
  }

  async function installPackage(source: string, scope: PackageScope): Promise<PackageActionResult> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return JSON.parse(await app.InstallPackage(source, scope))
  }

  async function removePackage(source: string, scope: PackageScope): Promise<PackageActionResult> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return JSON.parse(await app.RemovePackage(source, scope))
  }

  async function updatePackage(source: string): Promise<PackageActionResult> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return JSON.parse(await app.UpdatePackage(source))
  }

  async function updateAllPackages(): Promise<PackageActionResult> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    return JSON.parse(await app.UpdateAllPackages())
  }

  async function retryAgentStartup(): Promise<void> {
    const app = getApp()
    if (!app) throw new Error('后端未就绪')
    await app.RetryAgentStartup()
  }

  // ========== 生命周期 ==========
  // 只有第一个调用者（App.vue）注册全局生命周期
  const isRoot = !eventListenerStarted

  onMounted(() => {
    startEventListener()
  })

  onUnmounted(() => {
    // 只有根实例才执行清理
    if (!isRoot) return
    stopEventListener()
    eventListenerStarted = false
    const app = getApp()
    if (app) {
      app.StopAgent()
    }
  })

  return {
    init,
    sendPrompt,
    sendSteer,
    sendFollowUp,
    abort,
    setModel,
    setThinkingLevel,
    newSession,
    loadSessions,
    getAvailableModels,
    selectDirectory,
    getAppInfo,
    getAuthKeys,
    setApiKey,
    listPackages,
    installPackage,
    removePackage,
    updatePackage,
    updateAllPackages,
    retryAgentStartup,
  }
}
