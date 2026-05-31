<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import ChatArea from './components/ChatArea.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import { usePiAgent } from './composables/usePiAgent'
import { useChatStore } from './stores/chatStore'
import { useThemeStore } from './stores/themeStore'

const store = useChatStore()
const piAgent = usePiAgent()
const themeStore = useThemeStore()
const initialized = ref(false)
const showWelcome = ref(true)
const sidebarCollapsed = ref(false)
const showSettings = ref(false)
const startupError = ref('')
const startupLog = ref<string[]>([])

function log(msg: string) {
  startupLog.value.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
  console.log(msg)
}

onMounted(async () => {
  log('App 挂载成功，开始初始化...')

  // 初始化主题（默认跟随系统）
  themeStore.applyTheme()
  themeStore.startSystemThemeListener()

  // 先标记初始化完成，让 UI 渲染（不等后端）
  initialized.value = true

  // 尝试获取应用信息（Go 后端可能未就绪）
  try {
    const appInfo = await piAgent.getAppInfo()
    log(`获取应用信息成功: cwd=${appInfo.homeDir || '未知'}`)
  } catch (e: any) {
    log(`获取应用信息失败（后端未就绪）: ${e.message || e}`)
  }

  // 尝试初始化 Agent（不阻塞 UI）
  try {
    const appInfo = await piAgent.getAppInfo().catch(() => ({ homeDir: '' }))
    await piAgent.init(appInfo.homeDir || '')
    log('Agent 初始化完成')
    store.requestFocusInput()
    setTimeout(() => window.dispatchEvent(new Event('focus')), 200)
    // 等待会话列表加载完成后再检查
    await piAgent.loadSessions()
    // 已有历史会话 → 自动恢复最近一次对话
    if (store.sessions.length > 0) {
      const latest = store.sessions[0]
      store.setCurrentSession(latest)
      showWelcome.value = false
      // 加载最近会话的消息
      if (!store.loadCachedSession(latest.filePath)) {
        await piAgent.loadMessagesFromFile(latest.filePath)
      }
      // 后台切换 pi 到该会话
      piAgent.switchSessionInBackground(appInfo.homeDir || '', latest.filePath)
      log(`恢复会话: ${latest.displayName}`)
    }
  } catch (e: any) {
    startupError.value = `Agent 启动失败: ${e.message || e}`
    log(startupError.value)
    log('提示：确保已安装 pi (npm install -g @earendil-works/pi-coding-agent)')
  }
})

async function handleNewChat() {
  try {
    // 缓存当前会话消息
    const curPath = store.currentSession?.filePath
    if (curPath) {
      store.cacheCurrentSession(curPath)
    }
    await piAgent.newSession()
    store.clearMessages()
    showWelcome.value = true
    log('新建会话成功')
    store.requestFocusInput()
    setTimeout(() => window.dispatchEvent(new Event('focus')), 150)
    if (store.sessions.length > 0) {
      store.setCurrentSession(store.sessions[0])
    }
  } catch (e: any) {
    log(`新建会话失败: ${e.message || e}`)
  }
}

async function handleSelectSession(session: any) {
  try {
    showWelcome.value = false

    // 缓存当前会话消息，以便切换回来时即时恢复
    const curPath = store.currentSession?.filePath
    if (curPath) {
      store.cacheCurrentSession(curPath)
    }

    store.setCurrentSession(session)

    // 从缓存加载消息（即时显示）
    const cacheHit = store.loadCachedSession(session.filePath)
    if (!cacheHit) {
      // 缓存未命中，从文件加载
      await piAgent.loadMessagesFromFile(session.filePath).catch(() => {})
    }

    // 确保 pi 子进程已切换到当前会话上下文（后台执行，不阻塞）
    const appInfo = await piAgent.getAppInfo().catch(() => ({ homeDir: '' }))
    piAgent.switchSessionInBackground(appInfo.homeDir || '', session.filePath)

    log(`切换到会话: ${session.displayName}`)
    store.requestFocusInput()
    setTimeout(() => window.dispatchEvent(new Event('focus')), 150)
  } catch (e: any) {
    log(`切换会话失败: ${e.message || e}`)
  }
}

async function handleChangeDirectory() {
  try {
    const dir = await piAgent.selectDirectory().catch(() => '')
    if (dir) {
      store.clearMessages()
      await piAgent.init(dir)
      showWelcome.value = false
      log(`切换目录: ${dir}`)
    }
  } catch (e: any) {
    log(`切换目录失败: ${e.message || e}`)
  }
}

onUnmounted(() => {
  themeStore.stopSystemThemeListener()
})
</script>

<template>
  <!-- 如果完全黑屏超过 2 秒，至少展示这个 -->
  <div class="flex h-screen bg-white dark:bg-[#000] text-gray-900 dark:text-white antialiased">
    <!-- 左侧边栏 -->
    <Sidebar
      v-if="!sidebarCollapsed"
      :sessions="store.sessions"
      :current-session="store.currentSession"
      :model="store.appState.model"
      :thinking-level="store.appState.thinkingLevel"
      @new-chat="handleNewChat"
      @select-session="handleSelectSession"
      @change-directory="handleChangeDirectory"
      @collapse="sidebarCollapsed = true"
      @open-settings="showSettings = true"
    />

    <!-- 折叠按钮 -->
    <button
      v-if="sidebarCollapsed"
      @click="sidebarCollapsed = false"
      class="w-12 flex flex-col items-center pt-4 bg-gray-50 dark:bg-[#0d0d0d] border-r border-gray-200 dark:border-white/5 shrink-0
             hover:bg-gray-200 dark:hover:bg-[#1c1c1e] transition-colors"
    >
      <svg class="w-5 h-5 text-gray-500 hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </button>

    <!-- 右侧聊天区域 -->
    <div class="flex-1 flex flex-col min-w-0 min-h-0">
      <!-- 启动错误提示 -->
      <div v-if="startupError" class="px-4 py-3 bg-red-500/10 border-b border-red-500/20
                    text-red-400 text-sm flex items-center gap-2">
        <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ startupError }}
      </div>

      <ChatArea
        :messages="store.messages"
        :is-streaming="store.isStreaming"
        :is-agent-running="store.isAgentRunning"
        :show-welcome="showWelcome"
        :app-state="store.appState"
        @send="(msg: string) => piAgent.sendPrompt(msg).catch(e => log(`发送失败: ${e}`))"
        @abort="piAgent.abort().catch(e => log(`中止失败: ${e}`)); store.clearInputQueue()"
      />
    </div>

    <!-- 设置面板 -->
    <SettingsPanel
      v-if="showSettings"
      :model="store.appState.model"
      :thinking-level="store.appState.thinkingLevel"
      @close="showSettings = false"
    />
  </div>
</template>



