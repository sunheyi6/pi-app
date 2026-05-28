<script setup lang="ts">
import { computed } from 'vue'
import ModelSelector from './ModelSelector.vue'
import type { SessionInfo, ModelInfo, ThinkingLevel } from '../types'

const props = defineProps<{
  sessions: SessionInfo[]
  currentSession: SessionInfo | null
  model: ModelInfo | null
  thinkingLevel: ThinkingLevel
}>()

const emit = defineEmits<{
  'new-chat': []
  'select-session': [session: SessionInfo]
  'change-directory': []
  'collapse': []
}>()

const recentSessions = computed(() => props.sessions.slice(0, 50))

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days} 天前`
  if (days < 30) return `${Math.floor(days / 7)} 周前`
  return dateStr.slice(0, 10)
}

function truncate(str: string, max: number) {
  if (str.length <= max) return str
  return str.slice(0, max - 1) + '…'
}
</script>

<template>
  <div class="w-64 bg-gray-50 dark:bg-[#0d0d0d] border-r border-gray-200 dark:border-white/5 flex flex-col shrink-0 select-none">
    <!-- 头部 -->
    <div class="px-4 pt-5 pb-3 border-b border-gray-200 dark:border-white/5">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center
                      shadow-lg shadow-blue-500/20">
            <span class="text-white text-sm font-bold">π</span>
          </div>
          <span class="text-[15px] font-semibold text-gray-900 dark:text-white tracking-tight">Pi Desktop</span>
        </div>
        <button
          @click="$emit('collapse')"
          class="w-6 h-6 flex items-center justify-center rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-200 dark:hover:bg-white/5 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          </svg>
        </button>
      </div>
      <button
        @click="$emit('new-chat')"
        class="w-full py-2.5 px-4 bg-gray-900 hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-100 text-white dark:text-black rounded-xl
               text-[13px] font-medium transition-all duration-200 flex items-center justify-center gap-2
               active:scale-[0.97]"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新对话
      </button>
    </div>

    <!-- 会话列表 -->
    <div class="flex-1 overflow-y-auto py-1 px-2">
      <div class="text-[11px] font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider px-2 py-2">最近对话</div>

      <div
        v-for="session in recentSessions"
        :key="session.sessionId"
        @click="$emit('select-session', session)"
        class="group px-3 py-2.5 rounded-xl cursor-pointer transition-all duration-150 mb-0.5
               hover:bg-gray-200 dark:hover:bg-white/[0.06]"
        :class="{ 'bg-gray-200 dark:bg-white/[0.08]': currentSession?.sessionId === session.sessionId }"
      >
        <div class="text-[13px] font-medium text-gray-800 dark:text-gray-200 truncate leading-tight">
          {{ truncate(session.displayName, 28) }}
        </div>
        <div class="flex items-center justify-between mt-1.5">
          <span class="text-[11px] text-gray-400 dark:text-gray-500">
            {{ session.messageCount }} 条
          </span>
          <span class="text-[10px] text-gray-400 dark:text-gray-600">
            {{ formatDate(session.lastModified) }}
          </span>
        </div>
      </div>

      <div
        v-if="recentSessions.length === 0"
        class="px-3 py-8 text-center text-[12px] text-gray-400 dark:text-gray-600 leading-relaxed"
      >
        暂无历史对话<br>
        <span class="text-gray-400 dark:text-gray-700">点击上方按钮开始</span>
      </div>
    </div>

    <!-- 底部 -->
    <div class="px-2 py-2 border-t border-gray-200 dark:border-white/5 space-y-2">
      <div class="px-2">
        <div class="text-[11px] font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1.5">设置</div>
        <ModelSelector
          :model="props.model"
          :thinking-level="props.thinkingLevel"
        />
      </div>

      <button
        @click="$emit('change-directory')"
        class="w-full py-2 px-3 text-[12px] text-gray-500 hover:text-gray-700 dark:hover:text-gray-300
               hover:bg-gray-200 dark:hover:bg-white/[0.04] rounded-xl transition-colors flex items-center gap-2.5"
      >
        <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        切换工作目录
      </button>
    </div>
  </div>
</template>
