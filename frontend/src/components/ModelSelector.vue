<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usePiAgent } from '../composables/usePiAgent'
import { useThemeStore, type ThemeMode } from '../stores/themeStore'
import type { ModelInfo, ThinkingLevel } from '../types'

const props = defineProps<{
  model: ModelInfo | null
  thinkingLevel: ThinkingLevel
}>()

const piAgent = usePiAgent()
const themeStore = useThemeStore()
const models = ref<ModelInfo[]>([])
const showPanel = ref(false)
const loading = ref(false)

const themeOptions: { value: ThemeMode; label: string; icon: string }[] = [
  { value: 'system', label: '跟随系统', icon: '🖥️' },
  { value: 'light',  label: '浅色',     icon: '☀️' },
  { value: 'dark',   label: '深色',     icon: '🌙' },
]

const thinkingLevels: { value: ThinkingLevel; label: string; desc: string }[] = [
  { value: 'off',     label: '关闭', desc: '直接回答，不展示思考' },
  { value: 'minimal', label: '极简', desc: '最少思考输出' },
  { value: 'low',     label: '低',   desc: '简要思考过程' },
  { value: 'medium',  label: '中',   desc: '适中思考深度' },
  { value: 'high',    label: '高',   desc: '深度推理思考' },
  { value: 'xhigh',   label: '超高', desc: '最大推理深度' },
]

onMounted(async () => {
  loading.value = true
  models.value = await piAgent.getAvailableModels()
  loading.value = false
})

async function selectModel(model: ModelInfo) {
  await piAgent.setModel(model.provider, model.id)
  showPanel.value = false
}

async function selectThinkingLevel(level: ThinkingLevel) {
  await piAgent.setThinkingLevel(level)
}

// 点击遮罩关闭
function close() {
  showPanel.value = false
}
</script>

<template>
  <div class="relative">
    <!-- 触发按钮 -->
    <button
      @click="showPanel = !showPanel"
      class="flex items-center gap-1.5 text-[12px] px-3 py-1.5 rounded-full
             bg-gray-200 dark:bg-white/[0.06] hover:bg-gray-300 dark:hover:bg-white/[0.1] text-gray-700 dark:text-gray-300 transition-all
             border border-gray-300 dark:border-white/[0.06] hover:border-gray-400 dark:hover:border-white/[0.1]"
    >
      <span class="font-medium">{{ model?.name || '选择模型' }}</span>
      <svg class="w-3 h-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- 遮罩 -->
    <div
      v-if="showPanel"
      @click="close"
      class="fixed inset-0 z-40 bg-black/30 backdrop-blur-sm"
    />

    <!-- iOS 风格底部面板 -->
    <Transition name="slide-up">
      <div
        v-if="showPanel"
        class="fixed bottom-0 left-0 right-0 z-50 max-h-[70vh] overflow-y-auto
               bg-white/95 dark:bg-[#1c1c1e]/95 backdrop-blur-2xl rounded-t-[20px]
               border border-gray-200 dark:border-white/[0.08] shadow-2xl"
      >
        <!-- 拖拽条 -->
        <div class="flex justify-center pt-3 pb-1">
          <div class="w-9 h-1 rounded-full bg-gray-300 dark:bg-white/20" />
        </div>

        <div class="px-4 pb-6">
          <!-- 标题 -->
          <div class="text-center mb-4">
            <h3 class="text-[17px] font-semibold text-gray-900 dark:text-white">设置</h3>
          </div>

          <!-- 主题切换 -->
          <div class="mb-6">
            <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider px-2 mb-2">主题</div>
            <div class="grid grid-cols-3 gap-1.5">
              <button
                v-for="opt in themeOptions"
                :key="opt.value"
                @click="themeStore.setTheme(opt.value)"
                class="px-2 py-2.5 rounded-xl text-center transition-all"
                :class="themeStore.mode === opt.value
                  ? 'bg-blue-500/20 text-blue-500 dark:text-blue-400 ring-1 ring-blue-500/30'
                  : 'bg-gray-100 dark:bg-white/[0.04] text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/[0.08]'"
              >
                <div class="text-[18px] mb-0.5">{{ opt.icon }}</div>
                <div class="text-[11px] font-medium">{{ opt.label }}</div>
              </button>
            </div>
          </div>

          <!-- 模型列表 -->
          <div class="space-y-1 mb-6">
            <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider px-2 mb-2">可用模型</div>
            <div v-if="loading" class="px-2 py-4 text-center text-sm text-gray-400 dark:text-gray-500">加载中…</div>
            <button
              v-for="m in models"
              :key="m.id"
              @click="selectModel(m)"
              class="w-full text-left px-3 py-3 rounded-xl transition-colors flex items-center
                     justify-between hover:bg-gray-100 dark:hover:bg-white/[0.06]"
              :class="{ 'bg-gray-100 dark:bg-white/[0.08] ring-1 ring-blue-500/30': props.model?.id === m.id }"
            >
              <div>
                <div class="text-[15px] text-gray-900 dark:text-white font-medium">{{ m.name }}</div>
                <div class="text-[12px] text-gray-400 dark:text-gray-500 mt-0.5">
                  {{ m.provider }}
                  <span v-if="m.contextWindow">· {{ Math.round(m.contextWindow / 1000) }}k 上下文</span>
                </div>
              </div>
              <svg v-if="props.model?.id === m.id" class="w-5 h-5 text-blue-500 dark:text-blue-400 shrink-0" fill="currentColor" viewBox="0 0 24 24">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
              </svg>
            </button>
          </div>

          <!-- 思考级别 -->
          <div>
            <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider px-2 mb-2">思考级别</div>
            <div class="grid grid-cols-3 gap-1.5">
              <button
                v-for="tl in thinkingLevels"
                :key="tl.value"
                @click="selectThinkingLevel(tl.value)"
                class="px-2 py-2.5 rounded-xl text-center transition-all"
                :class="props.thinkingLevel === tl.value
                  ? 'bg-blue-500/20 text-blue-500 dark:text-blue-400 ring-1 ring-blue-500/30'
                  : 'bg-gray-100 dark:bg-white/[0.04] text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/[0.08]'"
              >
                <div class="text-[13px] font-medium">{{ tl.label }}</div>
                <div class="text-[9px] mt-0.5 opacity-60">{{ tl.desc }}</div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.slide-up-enter-active {
  transition: all 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-up-leave-active {
  transition: all 0.2s ease-in;
}
.slide-up-enter-from {
  transform: translateY(100%);
  opacity: 0;
}
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
