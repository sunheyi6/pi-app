<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usePiAgent } from '../composables/usePiAgent'
import { useThemeStore, type ThemeMode } from '../stores/themeStore'
import type { ModelInfo, ThinkingLevel } from '../types'

const props = defineProps<{
  model: ModelInfo | null
  thinkingLevel: ThinkingLevel
}>()

const emit = defineEmits<{
  close: []
}>()

const piAgent = usePiAgent()
const themeStore = useThemeStore()
const models = ref<ModelInfo[]>([])
const loading = ref(false)

// 左侧导航菜单
type SettingsTab = 'model' | 'theme' | 'thinking'
const activeTab = ref<SettingsTab>('model')

const tabs: { key: SettingsTab; label: string; icon: string }[] = [
  { key: 'model',    label: '模型选择', icon: '🤖' },
  { key: 'thinking', label: '思考级别', icon: '🧠' },
  { key: 'theme',    label: '主题外观', icon: '🎨' },
]

const themeOptions: { value: ThemeMode; label: string; icon: string }[] = [
  { value: 'system', label: '跟随系统', icon: '🖥️' },
  { value: 'light',  label: '浅色',     icon: '☀️' },
  { value: 'dark',   label: '深色',     icon: '🌙' },
]

const thinkingLevels: { value: ThinkingLevel; label: string; desc: string }[] = [
  { value: 'off',     label: '关闭', desc: '直接回答，不展示思考过程' },
  { value: 'minimal', label: '极简', desc: '最少思考输出' },
  { value: 'low',     label: '低',   desc: '简要思考过程' },
  { value: 'medium',  label: '中',   desc: '适中思考深度' },
  { value: 'high',    label: '高',   desc: '深度推理思考' },
  { value: 'xhigh',   label: '超高', desc: '最大推理深度（仅部分模型支持）' },
]

onMounted(async () => {
  loading.value = true
  models.value = await piAgent.getAvailableModels()
  loading.value = false
})

async function selectModel(model: ModelInfo) {
  await piAgent.setModel(model.provider, model.id)
}

async function selectThinkingLevel(level: ThinkingLevel) {
  await piAgent.setThinkingLevel(level)
}

// 当前选中的模型和思考级别（通过外部事件监听来更新，这里简单处理）
// 实际上可以从 store 读取，但为了让 Panel 自包含，通过 props 或 store 获取
</script>

<template>
  <!-- 遮罩层 -->
  <div
    class="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm"
    @click="$emit('close')"
  />

  <!-- 设置面板 - 从右侧滑入 -->
  <div class="fixed top-0 right-0 bottom-0 z-50 w-[480px] max-w-[90vw]
              bg-white dark:bg-[#1a1a1a] border-l border-gray-200 dark:border-white/[0.08]
              shadow-2xl flex flex-col">
    <!-- 头部 -->
    <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-white/[0.05] shrink-0">
      <h2 class="text-[17px] font-semibold text-gray-900 dark:text-white">设置</h2>
      <button
        @click="$emit('close')"
        class="w-8 h-8 flex items-center justify-center rounded-full
               text-gray-400 hover:text-gray-600 dark:hover:text-gray-300
               hover:bg-gray-100 dark:hover:bg-white/[0.06] transition-colors"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- 内容区域：左侧导航 + 右侧内容 -->
    <div class="flex-1 flex min-h-0">
      <!-- 左侧导航 -->
      <div class="w-36 shrink-0 border-r border-gray-100 dark:border-white/[0.05] py-3 px-2">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="w-full flex items-center gap-2.5 px-3 py-2.5 rounded-xl text-left transition-all mb-0.5"
          :class="activeTab === tab.key
            ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 font-medium'
            : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-white/[0.04]'"
        >
          <span class="text-base">{{ tab.icon }}</span>
          <span class="text-[13px]">{{ tab.label }}</span>
        </button>
      </div>

      <!-- 右侧内容 -->
      <div class="flex-1 overflow-y-auto px-5 py-4">
        <!-- 模型选择 -->
        <div v-if="activeTab === 'model'">
          <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">
            可用模型
          </div>
          <div v-if="loading" class="text-sm text-gray-400 dark:text-gray-500 py-8 text-center">
            加载中…
          </div>
          <div v-else class="space-y-1">
            <button
              v-for="m in models"
              :key="m.id"
              @click="selectModel(m)"
              class="w-full text-left px-3 py-3 rounded-xl transition-colors flex items-center
                     justify-between hover:bg-gray-50 dark:hover:bg-white/[0.04]"
              :class="{ 'bg-blue-50 dark:bg-blue-500/[0.08] ring-1 ring-blue-500/30': props.model?.id === m.id }"
            >
              <div>
                <div class="text-[14px] text-gray-900 dark:text-white font-medium">{{ m.name }}</div>
                <div class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
                  {{ m.provider }}
                  <span v-if="m.contextWindow"> · {{ Math.round(m.contextWindow / 1000) }}k 上下文</span>
                  <span v-if="m.reasoning" class="ml-1 text-purple-400">· 支持推理</span>
                </div>
              </div>
              <svg
                v-if="props.model?.id === m.id"
                class="w-5 h-5 text-blue-500 dark:text-blue-400 shrink-0"
                fill="currentColor" viewBox="0 0 24 24"
              >
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 思考级别 -->
        <div v-if="activeTab === 'thinking'">
          <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">
            推理深度
          </div>
          <p class="text-[12px] text-gray-400 dark:text-gray-500 mb-4">
            控制模型在回答前进行推理的深度。更高的级别需要更多 Token 和时间。
          </p>
          <div class="space-y-2">
            <button
              v-for="tl in thinkingLevels"
              :key="tl.value"
              @click="selectThinkingLevel(tl.value)"
              class="w-full text-left px-4 py-3 rounded-xl transition-all flex items-center justify-between"
              :class="props.thinkingLevel === tl.value
                ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 ring-1 ring-blue-500/30'
                : 'bg-gray-50 dark:bg-white/[0.03] text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/[0.06]'"
            >
              <div>
                <div class="text-[14px] font-medium">{{ tl.label }}</div>
                <div class="text-[11px] mt-0.5 opacity-70">{{ tl.desc }}</div>
              </div>
              <svg
                v-if="props.thinkingLevel === tl.value"
                class="w-5 h-5 shrink-0"
                fill="currentColor" viewBox="0 0 24 24"
              >
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 主题外观 -->
        <div v-if="activeTab === 'theme'">
          <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">
            外观模式
          </div>
          <p class="text-[12px] text-gray-400 dark:text-gray-500 mb-4">
            选择应用的颜色主题。"跟随系统"会根据系统设置自动切换浅色/深色。
          </p>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="opt in themeOptions"
              :key="opt.value"
              @click="themeStore.setTheme(opt.value)"
              class="px-3 py-4 rounded-xl text-center transition-all"
              :class="themeStore.mode === opt.value
                ? 'bg-blue-500/15 text-blue-600 dark:text-blue-400 ring-1 ring-blue-500/30'
                : 'bg-gray-50 dark:bg-white/[0.03] text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/[0.06]'"
            >
              <div class="text-[24px] mb-1.5">{{ opt.icon }}</div>
              <div class="text-[12px] font-medium">{{ opt.label }}</div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
