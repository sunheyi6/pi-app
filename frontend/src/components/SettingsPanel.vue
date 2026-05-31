<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usePiAgent } from '../composables/usePiAgent'
import { useThemeStore, type ThemeMode } from '../stores/themeStore'
import type { ModelInfo, ThinkingLevel } from '../types'
import ExtensionPackagesPanel from './ExtensionPackagesPanel.vue'

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
type SettingsTab = 'model' | 'apikey' | 'thinking' | 'theme' | 'extensions'
const activeTab = ref<SettingsTab>('model')

const tabs: { key: SettingsTab; label: string; icon: string }[] = [
  { key: 'model',    label: '模型选择', icon: '🤖' },
  { key: 'apikey',   label: 'API 密钥', icon: '🔑' },
  { key: 'thinking', label: '思考级别', icon: '🧠' },
  { key: 'theme',    label: '主题外观', icon: '🎨' },
  { key: 'extensions', label: '扩展包', icon: '＋' },
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
  await loadAuthKeys()
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

// ========== API Key 管理 ==========
const authKeys = ref<Record<string, string>>({})
const editingProvider = ref<string | null>(null)
const editingKey = ref('')
const savingProvider = ref<string | null>(null)
const saveError = ref('')

// provider → auth.json key 映射
const providerAuthMap: Record<string, string> = {
  anthropic: 'anthropic',
  openai: 'openai',
  deepseek: 'deepseek',
  google: 'google',
  groq: 'groq',
  openrouter: 'openrouter',
  xai: 'xai',
  mistral: 'mistral',
  cerebras: 'cerebras',
  together: 'together',
  'kimi-coding': 'kimi-coding',
  minimax: 'minimax',
  'vercel-ai-gateway': 'vercel-ai-gateway',
  'cloudflare-ai-gateway': 'cloudflare-ai-gateway',
}

async function loadAuthKeys() {
  authKeys.value = await piAgent.getAuthKeys()
}

function startEdit(provider: string) {
  editingProvider.value = provider
  editingKey.value = ''
  saveError.value = ''
}

function cancelEdit() {
  editingProvider.value = null
  editingKey.value = ''
  saveError.value = ''
}

async function saveKey(provider: string) {
  const authKey = providerAuthMap[provider]
  if (!authKey) return

  savingProvider.value = provider
  saveError.value = ''

  const ok = await piAgent.setApiKey(authKey, editingKey.value)
  if (ok) {
    editingProvider.value = null
    editingKey.value = ''
    await loadAuthKeys()
  } else {
    saveError.value = '保存失败，请重试'
  }
  savingProvider.value = null
}

async function deleteKey(provider: string) {
  const authKey = providerAuthMap[provider]
  if (!authKey) return
  await piAgent.setApiKey(authKey, '')
  await loadAuthKeys()
}
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

        <!-- API Key -->
        <div v-if="activeTab === 'apikey'">
          <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1">
            API 密钥
          </div>
          <p class="text-[12px] text-gray-400 dark:text-gray-500 mb-4">
            配置各模型的 API Key，保存至 <code class="text-[11px] bg-gray-100 dark:bg-white/[0.06] px-1 rounded">~/.pi/agent/auth.json</code>
          </p>
          <div v-if="saveError" class="text-[12px] text-red-400 mb-3">{{ saveError }}</div>
          <div v-if="models.length === 0" class="text-sm text-gray-400 dark:text-gray-500 py-4 text-center">
            尚未加载模型列表
          </div>
          <div v-else class="space-y-1">
            <div
              v-for="m in models"
              :key="m.id"
              class="flex items-center justify-between px-3 py-2.5 rounded-xl hover:bg-gray-50 dark:hover:bg-white/[0.04]"
            >
              <div class="flex-1 min-w-0">
                <div class="text-[13px] text-gray-900 dark:text-white">{{ m.name }}</div>
                <div class="text-[11px] text-gray-400 dark:text-gray-500">{{ m.provider }}</div>
              </div>

              <!-- 已配置 key -->
              <template v-if="editingProvider === m.provider">
                <div class="flex items-center gap-1.5">
                  <input
                    v-model="editingKey"
                    type="password"
                    placeholder="输入 API Key"
                    class="w-40 text-[12px] px-2.5 py-1.5 rounded-lg border border-gray-200 dark:border-white/[0.1]
                           bg-gray-50 dark:bg-white/[0.04] text-gray-900 dark:text-white
                           placeholder-gray-400 outline-none focus:border-blue-400"
                    @keydown.enter="m.provider && saveKey(m.provider)"
                    @keydown.escape="cancelEdit"
                  />
                  <button
                    @click="m.provider && saveKey(m.provider)"
                    :disabled="savingProvider === m.provider"
                    class="text-[11px] px-2.5 py-1.5 rounded-lg bg-blue-500 text-white hover:bg-blue-600
                           disabled:opacity-50 transition-colors"
                  >
                    {{ savingProvider === m.provider ? '保存中…' : '保存' }}
                  </button>
                  <button
                    @click="cancelEdit"
                    class="w-6 h-6 flex items-center justify-center rounded-full
                           text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </template>

              <!-- 已配置 -->
              <template v-else-if="authKeys[m.provider]">
                <div class="flex items-center gap-2">
                  <span class="text-[11px] text-green-500 dark:text-green-400 font-mono">{{ authKeys[m.provider] }}</span>
                  <button
                    @click="startEdit(m.provider)"
                    class="text-[11px] px-2 py-1 rounded-lg text-gray-400 hover:text-blue-400 hover:bg-blue-500/10 transition-colors"
                  >
                    修改
                  </button>
                  <button
                    @click="m.provider && deleteKey(m.provider)"
                    class="text-[11px] px-2 py-1 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                  >
                    删除
                  </button>
                </div>
              </template>

              <!-- 未配置 -->
              <button
                v-else
                @click="startEdit(m.provider)"
                class="text-[11px] px-2.5 py-1.5 rounded-lg text-gray-400 hover:text-blue-400
                       hover:bg-blue-500/10 border border-dashed border-gray-300 dark:border-white/[0.1]
                       transition-colors"
              >
                设置 Key
              </button>
            </div>
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

        <ExtensionPackagesPanel v-if="activeTab === 'extensions'" />
      </div>
    </div>
  </div>
</template>
