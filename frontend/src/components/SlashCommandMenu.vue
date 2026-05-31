<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { CommandInfo } from '../types'

const props = defineProps<{
  commands: CommandInfo[]
  query: string
}>()

const emit = defineEmits<{
  select: [command: CommandInfo]
  close: []
}>()

const activeIndex = ref(0)
const filteredCommands = computed(() => {
  const query = props.query.toLowerCase()
  return props.commands.filter(command =>
    command.name.toLowerCase().includes(query) ||
    (command.description || '').toLowerCase().includes(query)
  )
})

watch(filteredCommands, () => {
  activeIndex.value = 0
})

function handleKeydown(event: KeyboardEvent): boolean {
  if (event.key === 'Escape') {
    emit('close')
    return true
  }
  if (filteredCommands.value.length === 0) return false
  if (event.key === 'ArrowDown') {
    activeIndex.value = (activeIndex.value + 1) % filteredCommands.value.length
    return true
  }
  if (event.key === 'ArrowUp') {
    activeIndex.value = (activeIndex.value - 1 + filteredCommands.value.length) % filteredCommands.value.length
    return true
  }
  if (event.key === 'Enter') {
    emit('select', filteredCommands.value[activeIndex.value])
    return true
  }
  return false
}

defineExpose({ handleKeydown })
</script>

<template>
  <div class="absolute left-0 right-0 bottom-full mb-2 rounded-xl border border-gray-200 dark:border-white/[0.1]
              bg-white dark:bg-[#1c1c1e] shadow-xl overflow-hidden z-20">
    <div class="px-3 py-2 text-[10px] uppercase tracking-wider text-gray-400 border-b border-gray-100 dark:border-white/[0.06]">
      命令
    </div>
    <div v-if="filteredCommands.length === 0" class="px-3 py-3 text-[12px] text-gray-400">没有匹配命令</div>
    <button
      v-for="(command, index) in filteredCommands"
      :key="`${command.source}:${command.name}`"
      @mousedown.prevent="$emit('select', command)"
      class="w-full flex items-center justify-between gap-3 px-3 py-2 text-left transition-colors"
      :class="index === activeIndex ? 'bg-blue-500/10' : 'hover:bg-gray-50 dark:hover:bg-white/[0.04]'"
    >
      <div class="min-w-0">
        <div class="text-[13px] text-gray-800 dark:text-gray-100">/{{ command.name }}</div>
        <div v-if="command.description" class="text-[11px] text-gray-400 truncate mt-0.5">{{ command.description }}</div>
      </div>
      <span class="text-[10px] text-gray-400 shrink-0">{{ command.source }}</span>
    </button>
  </div>
</template>
