<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ChatMessage } from '../types'

const props = defineProps<{
  toolCall: { id: string; name: string; arguments: Record<string, any> }
  result?: ChatMessage
}>()

const expanded = ref(false)

const toolInfo = computed(() => {
  const defs: Record<string, { label: string; icon: string; color: string }> = {
    read:     { label: '读取文件', icon: '📖', color: '#0a84ff' },
    bash:     { label: '执行命令', icon: '⚡', color: '#ff9f0a' },
    edit:     { label: '编辑文件', icon: '✏️', color: '#bf5af2' },
    write:    { label: '写入文件', icon: '📝', color: '#30d158' },
    grep:     { label: '搜索代码', icon: '🔍', color: '#5ac8fa' },
    find:     { label: '查找文件', icon: '📁', color: '#5ac8fa' },
    ls:       { label: '列出目录', icon: '📂', color: '#5ac8fa' },
  }
  return defs[props.toolCall.name] || { label: props.toolCall.name, icon: '🔧', color: '#8e8e93' }
})

const argsPreview = computed(() => {
  const args = props.toolCall.arguments
  if (args.command) return (args.command as string).slice(0, 80)
  if (args.filePath || args.path) return (args.filePath || args.path) as string
  if (args.pattern) return (args.pattern as string).slice(0, 80)
  return ''
})

const resultText = computed(() => {
  if (!props.result) return ''
  return props.result.content
    .filter(c => c.type === 'text')
    .map(c => (c as any).text)
    .join('\n')
})

const isDiffResult = computed(() => {
  const txt = resultText.value
  return txt.includes('@@') && (txt.includes('\n+') || txt.includes('\n-'))
})

const diffLines = computed(() => {
  if (!isDiffResult.value) return []
  return resultText.value.split('\n').map(line => {
    if (line.startsWith('+') && !line.startsWith('+++')) return { type: 'add', text: line }
    if (line.startsWith('-') && !line.startsWith('---')) return { type: 'remove', text: line }
    if (line.startsWith('@@')) return { type: 'header', text: line }
    return { type: 'context', text: line }
  })
})

const truncatedOutput = computed(() => {
  const text = resultText.value
  if (!text) return ''
  const lines = text.split('\n')
  if (lines.length > 25) {
    return lines.slice(0, 25).join('\n') + `\n… 共 ${lines.length} 行`
  }
  return text
})
</script>

<template>
  <div class="my-1.5 rounded-xl border border-gray-200 dark:border-white/[0.06] overflow-hidden
              bg-gray-50 dark:bg-white/[0.02] hover:border-gray-300 dark:hover:border-white/[0.1] transition-colors">
    <!-- 头部：点击命令文本展开详情 -->
    <div class="flex items-center gap-2.5 px-3 py-2 select-none">
      <span class="text-sm">{{ toolInfo.icon }}</span>
      <span class="text-[12px] font-medium text-gray-700 dark:text-gray-300 shrink-0">{{ toolInfo.label }}</span>
      <span
        @click="expanded = !expanded"
        v-if="argsPreview"
        class="text-[11px] text-gray-400 dark:text-gray-500 truncate flex-1 font-mono
               cursor-pointer hover:text-gray-600 dark:hover:text-gray-300 transition-colors
               underline decoration-dotted underline-offset-4 decoration-gray-300 dark:decoration-gray-600"
      >{{ argsPreview }}</span>
      <span v-else class="flex-1" />
      <!-- 状态指示 -->
      <span v-if="result" class="flex items-center gap-1 shrink-0">
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="result.isError ? 'bg-red-400' : 'bg-green-400'"
        />
        <span class="text-[10px]"
              :class="result.isError ? 'text-red-400' : 'text-green-400'">
          {{ result.isError ? '失败' : '完成' }}
        </span>
      </span>
      <span v-else class="flex items-center gap-1 shrink-0">
        <span class="w-1.5 h-1.5 rounded-full bg-yellow-400 animate-pulse" />
        <span class="text-[10px] text-yellow-400">执行中</span>
      </span>
    </div>

    <!-- 详情面板 -->
    <div v-show="expanded" class="border-t border-gray-200 dark:border-white/[0.04] animate-slide-down">
      <!-- 参数 -->
      <div class="px-3 py-2">
        <div class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1">参数</div>
        <pre class="text-[11px] text-gray-600 dark:text-gray-400 font-mono bg-gray-100 dark:bg-black/30 rounded-lg p-2
                    overflow-x-auto leading-relaxed">{{ JSON.stringify(toolCall.arguments, null, 2) }}</pre>
      </div>

      <!-- 结果 -->
      <div v-if="result" class="px-3 pb-2.5 border-t border-gray-200 dark:border-white/[0.04] pt-2">
        <div class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1">输出</div>

        <!-- Diff 结果 -->
        <div v-if="isDiffResult" class="diff-block text-[11px]">
          <div
            v-for="(line, i) in diffLines"
            :key="i"
            class="diff-line text-[11px]"
            :class="{
              'diff-add': line.type === 'add',
              'diff-remove': line.type === 'remove',
              'diff-context': line.type === 'context',
            }"
          >
            <span v-if="line.type === 'header'" class="text-blue-400 font-medium text-[10px]">{{ line.text }}</span>
            <span v-else>{{ line.text }}</span>
          </div>
        </div>

        <!-- 普通输出 -->
        <pre
          v-else
          class="text-[11px] font-mono bg-gray-100 dark:bg-black/30 rounded-lg p-2 overflow-x-auto leading-relaxed max-h-60 overflow-y-auto"
          :class="result.isError ? 'text-red-400' : 'text-gray-600 dark:text-gray-400'"
        >{{ truncatedOutput || '(无输出)' }}</pre>
      </div>
    </div>
  </div>
</template>
