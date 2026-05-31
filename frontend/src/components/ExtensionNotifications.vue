<script setup lang="ts">
import type { ExtensionNotification } from '../types'

defineProps<{ notifications: ExtensionNotification[] }>()
const emit = defineEmits<{ dismiss: [id: string] }>()
</script>

<template>
  <div class="fixed top-4 right-4 z-[80] w-80 space-y-2">
    <button
      v-for="notice in notifications"
      :key="notice.id"
      @click="emit('dismiss', notice.id)"
      class="w-full text-left rounded-xl border border-gray-200 dark:border-white/[0.1]
             bg-white/95 dark:bg-[#1c1c1e]/95 shadow-lg px-3 py-2.5 text-[12px]
             text-gray-700 dark:text-gray-200 hover:border-blue-400/50 transition-colors"
    >
      {{ notice.message }}
    </button>
  </div>
</template>
