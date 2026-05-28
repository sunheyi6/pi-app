<script setup lang="ts">
import { computed, ref } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import type { ChatMessage, TextContent } from '../types'

const props = defineProps<{
  message: ChatMessage
  isStreaming?: boolean
}>()

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const md: MarkdownIt = new MarkdownIt({
  html: false,
  breaks: true,
  linkify: true,
  highlight(str: string, lang: string): string {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre data-lang="${lang}"><code class="hljs language-${lang}">${hljs.highlight(str, { language: lang, ignoreIllegals: true }).value}</code></pre>`
      } catch {}
    }
    return `<pre><code class="hljs">${escapeHtml(str)}</code></pre>`
  },
})

const isUser = computed(() => props.message.role === 'user')
const contentEl = ref<HTMLElement | null>(null)

const textContents = computed(() => {
  return props.message.content
    .filter((c): c is TextContent => c.type === 'text')
    .map(c => c.text)
    .join('\n\n')
})

// 检测是否是 diff 内容
const isDiff = computed(() => {
  const txt = textContents.value
  return txt.includes('@@') && (txt.includes('\n+') || txt.includes('\n-'))
})

// 解析 diff 行
const diffLines = computed(() => {
  if (!isDiff.value) return []
  return textContents.value.split('\n').map(line => {
    if (line.startsWith('+') && !line.startsWith('+++')) return { type: 'add', text: line }
    if (line.startsWith('-') && !line.startsWith('---')) return { type: 'remove', text: line }
    if (line.startsWith('@@')) return { type: 'header', text: line }
    return { type: 'context', text: line }
  })
})

const hasUsage = computed(() => {
  return props.message.usage && (
    (props.message.usage.input || 0) > 0 ||
    (props.message.usage.output || 0) > 0
  )
})

function formatTokens(n: number | undefined) {
  if (!n) return '0'
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

function formatCost(n: number | undefined) {
  if (!n || n === 0) return '$0'
  if (n < 0.01) return '<$0.01'
  return `$${n.toFixed(2)}`
}
</script>

<template>
  <div
    class="animate-fade-in"
    :class="isUser ? 'flex justify-end' : 'flex justify-start'"
  >
    <!-- 用户消息气泡 -->
    <div v-if="isUser" class="max-w-[80%] px-4 py-2.5 rounded-2xl rounded-br-md
                          bg-[#0a84ff] text-white shadow-sm shadow-blue-500/10">
      <div class="whitespace-pre-wrap break-words text-[15px] leading-relaxed">
        {{ textContents }}
      </div>
    </div>

    <!-- 助手消息 -->
    <div v-else class="max-w-[90%]">
      <!-- Diff 内容 -->
      <div v-if="isDiff" class="diff-block">
        <div class="diff-header">变更内容</div>
        <div
          v-for="(line, i) in diffLines"
          :key="i"
          class="diff-line"
          :class="{
            'diff-add': line.type === 'add',
            'diff-remove': line.type === 'remove',
            'diff-context': line.type === 'context',
            'diff-header': line.type === 'header',
          }"
        >
          <span v-if="line.type === 'header'" class="text-blue-400 font-medium">{{ line.text }}</span>
          <span v-else>{{ line.text }}</span>
        </div>
      </div>

      <!-- 普通 Markdown -->
      <div
        v-else-if="textContents"
        ref="contentEl"
        class="markdown-body text-[15px] leading-relaxed"
        v-html="md.render(textContents)"
      />

      <!-- 无文本内容（可能只有工具调用） -->
      <div v-else class="text-gray-500 text-sm italic">...</div>

      <!-- 流式光标 -->
      <span
        v-if="isStreaming"
        class="inline-block w-[6px] h-[17px] bg-blue-400 ml-0.5 rounded-sm animate-pulse-dot align-text-bottom"
      />

      <!-- Token 用量条 -->
      <div
        v-if="hasUsage && !isStreaming"
        class="mt-2 flex items-center gap-3 text-[10px] text-gray-500 select-none"
      >
        <span class="flex items-center gap-1">
          <span class="w-1 h-1 rounded-full bg-gray-600" /> 输入 {{ formatTokens(message.usage?.input) }}
        </span>
        <span class="flex items-center gap-1">
          <span class="w-1 h-1 rounded-full bg-blue-600" /> 输出 {{ formatTokens(message.usage?.output) }}
        </span>
        <span v-if="message.usage?.cacheRead" class="flex items-center gap-1">
          <span class="w-1 h-1 rounded-full bg-green-600" /> 缓存 {{ formatTokens(message.usage.cacheRead) }}
        </span>
        <span class="text-gray-400 dark:text-gray-600">
          {{ formatCost(message.usage?.cost?.total) }}
        </span>
      </div>
    </div>
  </div>
</template>
