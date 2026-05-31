<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { usePiAgent } from '../composables/usePiAgent'
import type { PackageInfo, PackageScope } from '../types'

const piAgent = usePiAgent()
const scope = ref<PackageScope>('project')
const source = ref('')
const packages = ref<PackageInfo[]>([])
const loading = ref(false)
const error = ref('')
const status = ref('')
const restartError = ref('')

async function loadPackages() {
  loading.value = true
  error.value = ''
  try {
    packages.value = await piAgent.listPackages(scope.value)
  } catch (err: any) {
    error.value = err?.message || String(err)
  } finally {
    loading.value = false
  }
}

async function runOperation(operation: () => Promise<{ restartError?: string }>, successText: string) {
  loading.value = true
  error.value = ''
  status.value = ''
  restartError.value = ''
  try {
    const result = await operation()
    restartError.value = result.restartError || ''
    status.value = restartError.value ? '扩展包已更新，Agent 重新启动失败' : successText
    await loadPackages()
  } catch (err: any) {
    error.value = err?.message || String(err)
  } finally {
    loading.value = false
  }
}

async function installPackage() {
  const value = source.value.trim()
  if (!value) return
  const accepted = window.confirm('第三方扩展可能读取本机文件并执行命令。确认安装此扩展包吗？')
  if (!accepted) return
  await runOperation(() => piAgent.installPackage(value, scope.value), '扩展包已安装，Agent 已重新加载')
  if (!error.value) source.value = ''
}

async function removePackage(item: PackageInfo) {
  if (!window.confirm(`确认卸载 ${item.source} 吗？`)) return
  await runOperation(() => piAgent.removePackage(item.source, item.scope), '扩展包已卸载，Agent 已重新加载')
}

async function retryStartup() {
  loading.value = true
  error.value = ''
  try {
    await piAgent.retryAgentStartup()
    restartError.value = ''
    status.value = 'Agent 已恢复运行'
  } catch (err: any) {
    error.value = err?.message || String(err)
  } finally {
    loading.value = false
  }
}

function browseCatalog() {
  window.open('https://pi.dev/packages/', '_blank', 'noopener,noreferrer')
}

watch(scope, loadPackages)
onMounted(loadPackages)
</script>

<template>
  <div>
    <div class="flex items-start justify-between gap-3 mb-4">
      <div>
        <div class="text-[11px] text-gray-400 dark:text-gray-500 uppercase tracking-wider">扩展包</div>
        <p class="text-[12px] text-gray-400 dark:text-gray-500 mt-1 leading-relaxed">
          安装 Pi 扩展、技能和提示词模板。默认仅对当前项目生效。
        </p>
      </div>
      <button
        data-testid="update-all"
        :disabled="loading"
        @click="runOperation(() => piAgent.updateAllPackages(), '扩展包已更新，Agent 已重新加载')"
        class="shrink-0 text-[11px] px-2.5 py-1.5 rounded-lg text-gray-500 dark:text-gray-400
               hover:text-blue-500 hover:bg-blue-500/10 disabled:opacity-40 transition-colors"
      >
        全部更新
      </button>
    </div>

    <div class="inline-flex p-0.5 rounded-lg bg-gray-100 dark:bg-white/[0.05] mb-3">
      <button
        data-testid="scope-project"
        :disabled="loading"
        @click="scope = 'project'"
        class="px-3 py-1.5 rounded-md text-[12px] transition-colors"
        :class="scope === 'project' ? 'bg-white dark:bg-white/[0.1] text-gray-900 dark:text-white shadow-sm' : 'text-gray-400'"
      >当前项目</button>
      <button
        data-testid="scope-global"
        :disabled="loading"
        @click="scope = 'global'"
        class="px-3 py-1.5 rounded-md text-[12px] transition-colors"
        :class="scope === 'global' ? 'bg-white dark:bg-white/[0.1] text-gray-900 dark:text-white shadow-sm' : 'text-gray-400'"
      >全局</button>
    </div>

    <div class="flex gap-2">
      <input
        v-model="source"
        data-testid="package-source"
        :disabled="loading"
        placeholder="npm:package、Git URL 或本地路径"
        class="min-w-0 flex-1 text-[12px] px-3 py-2 rounded-lg border border-gray-200 dark:border-white/[0.1]
               bg-gray-50 dark:bg-white/[0.04] text-gray-900 dark:text-white outline-none focus:border-blue-400"
        @keydown.enter="installPackage"
      />
      <button
        data-testid="install-package"
        :disabled="loading || !source.trim()"
        @click="installPackage"
        class="px-3 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white text-[12px]
               disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      >安装</button>
    </div>

    <button @click="browseCatalog" class="mt-2 text-[11px] text-blue-500 hover:text-blue-600 transition-colors">
      浏览扩展目录 ↗
    </button>

    <div v-if="error" class="mt-3 text-[12px] text-red-400">{{ error }}</div>
    <div v-if="status" class="mt-3 text-[12px] text-green-500 dark:text-green-400">{{ status }}</div>
    <div v-if="restartError" class="mt-2 p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
      <div class="text-[12px] text-amber-500">{{ restartError }}</div>
      <button
        :disabled="loading"
        @click="retryStartup"
        class="mt-2 text-[11px] text-amber-600 dark:text-amber-400 hover:underline disabled:opacity-40"
      >重试启动 Agent</button>
    </div>

    <div class="mt-5 border-t border-gray-100 dark:border-white/[0.06]">
      <div v-if="loading && packages.length === 0" class="text-[12px] text-gray-400 py-5 text-center">加载中...</div>
      <div v-else-if="packages.length === 0" class="text-[12px] text-gray-400 py-5 text-center">当前范围暂无扩展包</div>
      <div
        v-for="item in packages"
        :key="`${item.scope}:${item.source}`"
        class="flex items-center justify-between gap-3 py-3 border-b border-gray-100 dark:border-white/[0.06]"
      >
        <div class="min-w-0">
          <div class="text-[13px] text-gray-800 dark:text-gray-200 truncate">{{ item.source }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ item.type }} · {{ item.scope === 'project' ? '当前项目' : '全局' }}</div>
        </div>
        <div class="flex items-center gap-1 shrink-0">
          <button
            data-testid="update-package"
            :disabled="loading"
            @click="runOperation(() => piAgent.updatePackage(item.source), '扩展包已更新，Agent 已重新加载')"
            class="text-[11px] px-2 py-1 rounded-md text-gray-400 hover:text-blue-500 hover:bg-blue-500/10 disabled:opacity-40"
          >更新</button>
          <button
            data-testid="remove-package"
            :disabled="loading"
            @click="removePackage(item)"
            class="text-[11px] px-2 py-1 rounded-md text-gray-400 hover:text-red-400 hover:bg-red-500/10 disabled:opacity-40"
          >卸载</button>
        </div>
      </div>
    </div>
  </div>
</template>
