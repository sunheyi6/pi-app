import { defineStore } from "pinia";
import { ref, computed } from "vue";
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
} from "../types";

export const useChatStore = defineStore("chat", () => {
	// ========== 状态 ==========
	const messages = ref<ChatMessage[]>([]);
	const sessions = ref<SessionInfo[]>([]);
	const currentSession = ref<SessionInfo | null>(null);
	const isStreaming = ref(false);
	const isAgentRunning = ref(false);
	const streamingMessageId = ref<string | null>(null);
	const currentThinkingBlock = ref<string>("");
	const currentToolCalls = ref<
		Map<
			string,
			{
				name: string;
				args: Record<string, any>;
				output: string;
				isError: boolean;
			}
		>
	>(new Map());
	const pendingSteering = ref<string[]>([]);
	const pendingFollowUp = ref<string[]>([]);
	// 消息输入队列：AI 回答期间用户输入的消息会排队，回答结束后自动发送
	const inputQueue = ref<string[]>([]);
	// 输入框聚焦触发器：递增时 InputBox 自动聚焦
	const focusInputCounter = ref(0);
	// 折叠状态：回答完成后自动折叠工具调用
	const collapsedMessageIds = ref<Record<string, boolean>>({});
	// 会话消息缓存：按 sessionPath 缓存最近的消息，切换时即时显示
	const sessionMessageCache = ref<Map<string, ChatMessage[]>>(new Map());
	const MAX_CACHED_SESSIONS = 10;
	const appState = ref<AppState>({
		isStreaming: false,
		isCompacting: false,
		steeringMode: "one-at-a-time",
		followUpMode: "one-at-a-time",
		autoCompactionEnabled: true,
		model: null,
		thinkingLevel: "medium",
		sessionFile: null,
		sessionId: "",
		messageCount: 0,
		pendingMessageCount: 0,
	});

	// ========== 计算属性 ==========
	const lastAssistantMessage = computed(() => {
		for (let i = messages.value.length - 1; i >= 0; i--) {
			if (messages.value[i].role === "assistant") {
				return messages.value[i];
			}
		}
		return null;
	});

	// ========== 消息管理 ==========
	function addMessage(message: ChatMessage) {
		messages.value.push(message);
	}

	function updateMessage(
		messageId: string,
		updater: (msg: ChatMessage) => void,
	) {
		const idx = messages.value.findIndex((m) => m.id === messageId);
		if (idx !== -1) {
			updater(messages.value[idx]);
		}
	}

	// 获取或创建当前的流式助手消息
	function getOrCreateStreamingMessage(): ChatMessage {
		if (streamingMessageId.value) {
			const msg = messages.value.find((m) => m.id === streamingMessageId.value);
			if (msg) return msg;
		}

		const id = `streaming-${Date.now()}`;
		const msg: ChatMessage = {
			id,
			role: "assistant",
			content: [],
			timestamp: Date.now(),
		};
		messages.value.push(msg);
		streamingMessageId.value = id;
		return msg;
	}

	// 添加文本 delta
	function appendTextDelta(delta: string, contentIndex: number) {
		const msg = getOrCreateStreamingMessage();
		// 确保有足够的 content 槽位
		while (msg.content.length <= contentIndex) {
			msg.content.push({ type: "text", text: "" });
		}
		const textContent = msg.content[contentIndex] as TextContent;
		textContent.text += delta;
	}

	// 开始思考块
	function appendThinkingDelta(delta: string) {
		currentThinkingBlock.value += delta;
	}

	// 结束思考块 - 将思考内容添加到消息
	function finalizeThinkingBlock() {
		if (currentThinkingBlock.value) {
			const msg = getOrCreateStreamingMessage();
			msg.content.push({
				type: "thinking",
				thinking: currentThinkingBlock.value,
			});
			currentThinkingBlock.value = "";
		}
	}

	// 工具调用开始
	function startToolCall(
		toolCallId: string,
		toolName: string,
		args: Record<string, any>,
	) {
		currentToolCalls.value.set(toolCallId, {
			name: toolName,
			args,
			output: "",
			isError: false,
		});

		const msg = getOrCreateStreamingMessage();
		msg.content.push({
			type: "toolCall",
			id: toolCallId,
			name: toolName,
			arguments: args,
		});
	}

	// 工具调用输出更新
	function updateToolOutput(toolCallId: string, output: string) {
		const tc = currentToolCalls.value.get(toolCallId);
		if (tc) {
			tc.output = output;
		}
	}

	// 工具调用结束
	function endToolCall(toolCallId: string, isError: boolean) {
		const tc = currentToolCalls.value.get(toolCallId);
		if (tc) {
			tc.isError = isError;
		}

		// 将工具结果作为独立消息添加
		if (tc) {
			messages.value.push({
				id: `toolresult-${toolCallId}`,
				role: "toolResult",
				content: [{ type: "text", text: tc.output }],
				timestamp: Date.now(),
				toolCallId,
				toolName: tc.name,
				isError,
			});
		}
	}

	// 完成流式消息
	function finalizeStreamingMessage(
		model?: string,
		provider?: string,
		usage?: any,
		stopReason?: string,
	) {
		if (streamingMessageId.value) {
			const msgId = streamingMessageId.value;
			updateMessage(msgId, (msg) => {
				if (model) msg.model = model;
				if (provider) msg.provider = provider;
				if (usage) msg.usage = usage;
				if (stopReason) msg.stopReason = stopReason;
			});
			// 完成时自动折叠工具调用（思考过程保持展开）
			const finishedMsg = messages.value.find((m) => m.id === msgId);
			if (finishedMsg && finishedMsg.content.some((c) => c.type === "toolCall")) {
				collapsedMessageIds.value[msgId] = true;
			}
			streamingMessageId.value = null;
		}
		currentThinkingBlock.value = "";
		currentToolCalls.value.clear();
	}

	// ========== 会话管理 ==========
	function setSessions(list: SessionInfo[]) {
		sessions.value = list;
	}

	function setCurrentSession(session: SessionInfo | null) {
		currentSession.value = session;
	}

	function clearMessages() {
		messages.value = [];
		streamingMessageId.value = null;
		currentThinkingBlock.value = "";
		currentToolCalls.value.clear();
	}

	// ========== 状态 ==========
	function setStreaming(val: boolean) {
		isStreaming.value = val;
		appState.value.isStreaming = val;
	}

	function setAgentRunning(val: boolean) {
		isAgentRunning.value = val;
	}

	function updateAppState(state: Partial<AppState>) {
		Object.assign(appState.value, state);
	}

	function setPendingSteering(list: string[]) {
		pendingSteering.value = list;
	}

	function setPendingFollowUp(list: string[]) {
		pendingFollowUp.value = list;
	}

	// ========== 输入队列 ==========
	function enqueueInput(text: string) {
		inputQueue.value.push(text);
	}

	function dequeueInput(): string | undefined {
		return inputQueue.value.shift();
	}

	function clearInputQueue() {
		inputQueue.value = [];
	}

	function removeFromQueue(index: number) {
		inputQueue.value.splice(index, 1);
	}

	// ========== 折叠中间过程 ==========
	function expandAssistantMessage(messageId: string) {
		delete collapsedMessageIds.value[messageId];
	}

	function isMessageCollapsed(messageId: string): boolean {
		return !!collapsedMessageIds.value[messageId];
	}

	// 对所有已加载的助手消息自动折叠工具调用（用于从会话恢复时）
	function collapseLoadedMessages() {
		const newCollapsed: Record<string, boolean> = {};
		for (const msg of messages.value) {
			if (
				msg.role === "assistant" &&
				msg.content.some((c) => c.type === "toolCall")
			) {
				// 已手动展开的消息保留展开状态
				if (collapsedMessageIds.value[msg.id] === false) continue;
				newCollapsed[msg.id] = true;
			}
		}
		collapsedMessageIds.value = newCollapsed;
	}

	function getIntermediateCount(messageId: string): number {
		const msg = messages.value.find((m) => m.id === messageId);
		if (!msg) return 0;
		// 统计非文本内容 + 关联的工具结果消息
		let count = msg.content.filter((c) => c.type !== "text").length;
		const toolCallIds = msg.content
			.filter((c) => c.type === "toolCall")
			.map((c) => (c as any).id);
		count += messages.value.filter(
			(m) =>
				m.role === "toolResult" && toolCallIds.includes(m.toolCallId || ""),
		).length;
		return count;
	}

	// ========== 会话消息缓存 ==========
	// 保存当前消息到缓存（切换会话前调用）
	function cacheCurrentSession(sessionPath: string) {
		if (!sessionPath || messages.value.length === 0) return;
		const cache = sessionMessageCache.value;
		// 超过上限时删除最旧的条目
		if (cache.size >= MAX_CACHED_SESSIONS) {
			const firstKey = cache.keys().next().value;
			if (firstKey) cache.delete(firstKey);
		}
		// 重新设置以更新 Map 的插入顺序（最近使用的保留）
		cache.delete(sessionPath);
		cache.set(sessionPath, [...messages.value]);
	}

	// 从缓存加载消息（返回 true 表示命中缓存）
	function loadCachedSession(sessionPath: string): boolean {
		const cached = sessionMessageCache.value.get(sessionPath);
		if (!cached || cached.length === 0) return false;
		messages.value = [...cached];
		return true;
	}

	// ========== 输入框聚焦 ==========
	function requestFocusInput() {
		focusInputCounter.value++;
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

		// 折叠
		collapsedMessageIds,
		expandAssistantMessage,
		isMessageCollapsed,
		getIntermediateCount,
		collapseLoadedMessages,

		// 会话消息缓存
		sessionMessageCache,
		cacheCurrentSession,
		loadCachedSession,

		// 聚焦
		focusInputCounter,
		requestFocusInput,
	};
});
