<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  content: string
  isStreaming?: boolean
}>()

const displayContent = computed(() => {
  if (!props.content) return ''
  return props.content
})
</script>

<template>
  <div class="my-1">
    <!-- 内容：左对齐，无图标 -->
    <div
      v-if="displayContent"
      class="text-[12px] text-gray-500 dark:text-gray-400 leading-relaxed whitespace-pre-wrap
             max-h-60 overflow-y-auto"
    >
      {{ displayContent }}
      <span
        v-if="isStreaming"
        class="inline-block w-1 h-3.5 bg-purple-400 ml-0.5 animate-pulse-dot align-middle rounded-sm"
      />
    </div>
  </div>
</template>
