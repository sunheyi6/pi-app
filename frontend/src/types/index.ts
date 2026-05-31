// ============ 前端类型定义 ============

// 消息角色
export type MessageRole = 'user' | 'assistant' | 'toolResult' | 'bashExecution'

// 消息内容块
export interface TextContent {
  type: 'text'
  text: string
}

export interface ThinkingContent {
  type: 'thinking'
  thinking: string
}

export interface ToolCallContent {
  type: 'toolCall'
  id: string
  name: string
  arguments: Record<string, any>
}

export type MessageContent = TextContent | ThinkingContent | ToolCallContent

// 消息
export interface ChatMessage {
  id: string
  role: MessageRole
  content: MessageContent[]
  timestamp: number
  // 助手消息特有
  model?: string
  provider?: string
  usage?: TokenUsage
  stopReason?: string
  // 工具结果特有
  toolCallId?: string
  toolName?: string
  isError?: boolean
  // bash 执行特有
  command?: string
  output?: string
  exitCode?: number
}

// Token 用量
export interface TokenUsage {
  input: number
  output: number
  cacheRead: number
  cacheWrite: number
  cost?: CostBreakdown
}

export interface CostBreakdown {
  input: number
  output: number
  cacheRead: number
  cacheWrite: number
  total: number
}

// 流式 delta 事件
export interface AssistantMessageEvent {
  type: DeltaType
  contentIndex?: number
  delta?: string
  content?: string | MessageContent[]
  partial?: any
  toolCall?: ToolCall
  reason?: string
}

export type DeltaType =
  | 'start'
  | 'text_start'
  | 'text_delta'
  | 'text_end'
  | 'thinking_start'
  | 'thinking_delta'
  | 'thinking_end'
  | 'toolcall_start'
  | 'toolcall_delta'
  | 'toolcall_end'
  | 'done'
  | 'error'

export interface ToolCall {
  id: string
  name: string
  arguments: Record<string, any>
}

// 工具执行
export interface ToolExecution {
  toolCallId: string
  toolName: string
  args: Record<string, any>
  content?: ToolResultContent[]
  isError?: boolean
}

export interface ToolResultContent {
  type: string
  text: string
}

// 会话信息
export interface SessionInfo {
  filePath: string
  sessionId: string
  displayName: string
  messageCount: number
  lastModified: string
  size?: string
}

// 模型信息
export interface ModelInfo {
  id: string
  name: string
  api: string
  provider: string
  baseUrl?: string
  reasoning?: boolean
  input?: string[]
  contextWindow?: number
  maxTokens?: number
  cost?: {
    input: number
    output: number
    cacheRead?: number
    cacheWrite?: number
  }
}

// 命令信息
export interface CommandInfo {
  name: string
  description?: string
  source: 'extension' | 'prompt' | 'skill'
  location?: string
  path?: string
}

export type PackageScope = 'project' | 'global'

export interface PackageInfo {
  source: string
  scope: PackageScope
  type: string
  raw?: string
}

export interface PackageActionResult {
  output?: string
  packages?: PackageInfo[]
  restarted?: boolean
  restartError?: string
}

// 应用状态
export interface AppState {
  isStreaming: boolean
  isCompacting: boolean
  steeringMode: string
  followUpMode: string
  autoCompactionEnabled: boolean
  model: ModelInfo | null
  thinkingLevel: ThinkingLevel
  sessionFile: string | null
  sessionId: string
  sessionName?: string
  messageCount: number
  pendingMessageCount: number
}

export type ThinkingLevel = 'off' | 'minimal' | 'low' | 'medium' | 'high' | 'xhigh'

// RPC 事件类型
export type RpcEventType =
  | 'agent_start'
  | 'agent_end'
  | 'turn_start'
  | 'turn_end'
  | 'message_start'
  | 'message_update'
  | 'message_end'
  | 'tool_execution_start'
  | 'tool_execution_update'
  | 'tool_execution_end'
  | 'queue_update'
  | 'compaction_start'
  | 'compaction_end'
  | 'auto_retry_start'
  | 'auto_retry_end'
  | 'extension_error'
  | 'extension_ui_request'

export interface RPCEvent {
  type: RpcEventType
  [key: string]: any
}

export interface ExtensionUIRequest {
  type: 'extension_ui_request'
  id: string
  method: string
  title?: string
  message?: string
  options?: string[]
  timeout?: number
  placeholder?: string
  prefill?: string
  notifyType?: string
}

export interface ExtensionNotification {
  id: string
  message: string
  type: string
}

// 图片数据
export interface ImageData {
  type: string
  data: string
  mimeType: string
}
