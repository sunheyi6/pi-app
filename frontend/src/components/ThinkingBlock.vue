<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  content: string
  isStreaming?: boolean
}>()

const expanded = ref(false)

// 流式思考期间自动展开，结束后保持折叠
watch(() => props.isStreaming, (streaming) => {
  if (streaming) expanded.value = true
})

const displayContent = computed(() => {
  if (!props.content) return ''
  if (!expanded.value && props.content.length > 400) {
    return props.content.slice(0, 400) + '…'
  }
  return props.content
})
</script>

<template>
  <div class="group my-1">
    <!-- 折叠头 -->
    <button
      @click="expanded = !expanded"
      class="flex items-center gap-2 w-full text-left py-1.5 px-1 -mx-1 rounded-lg
             hover:bg-gray-100 dark:hover:bg-white/[0.03] transition-colors select-none"
    >
      <svg
        class="w-3 h-3 text-gray-500 transition-transform duration-200 shrink-0"
        :class="{ 'rotate-90': expanded }"
        fill="none" stroke="currentColor" viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="flex items-center gap-1.5">
        <svg v-if="isStreaming" class="w-3.5 h-3.5 text-purple-400 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="16" stroke-linecap="round" />
        </svg>
        <svg v-else class="w-3.5 h-3.5 text-purple-400/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
        <span class="text-[11px] font-medium text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-400">
          {{ isStreaming ? '思考中…' : '思考过程' }}
        </span>
      </span>
    </button>

    <!-- 内容 -->
    <div
      v-show="expanded && displayContent"
      class="mt-1 ml-5 pl-3 border-l-2 border-purple-500/20
             text-[12px] text-gray-500 dark:text-gray-400 leading-relaxed whitespace-pre-wrap
             max-h-60 overflow-y-auto animate-slide-down"
    >
      {{ displayContent }}
      <span
        v-if="isStreaming"
        class="inline-block w-1 h-3.5 bg-purple-400 ml-0.5 animate-pulse-dot align-middle rounded-sm"
      />
    </div>
  </div>
</template>
