<script setup lang="ts">
import { ref } from 'vue'
import type { ExtensionUIRequest } from '../types'

const props = defineProps<{ request: ExtensionUIRequest }>()
const emit = defineEmits<{
  respond: [value: string, confirmed: boolean]
  cancel: []
}>()

const text = ref(props.request.prefill || '')
</script>

<template>
  <div class="fixed inset-0 z-[70] bg-black/45 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="w-full max-w-md rounded-2xl border border-gray-200 dark:border-white/[0.1]
                bg-white dark:bg-[#1c1c1e] shadow-2xl p-5">
      <h3 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ request.title || '扩展请求' }}</h3>
      <p v-if="request.message" class="mt-2 text-[13px] text-gray-500 dark:text-gray-400">{{ request.message }}</p>

      <div v-if="request.method === 'select'" class="mt-4 space-y-1">
        <button
          v-for="option in request.options || []"
          :key="option"
          :data-testid="`select-option-${option}`"
          @click="$emit('respond', option, true)"
          class="w-full text-left px-3 py-2.5 rounded-lg text-[13px] text-gray-700 dark:text-gray-200
                 bg-gray-50 dark:bg-white/[0.04] hover:bg-blue-500/10 transition-colors"
        >{{ option }}</button>
      </div>

      <div v-else-if="request.method === 'confirm'" class="mt-5 flex justify-end gap-2">
        <button data-testid="confirm-no" @click="$emit('respond', '', false)"
                class="px-3 py-2 rounded-lg text-[12px] text-gray-500 hover:bg-gray-100 dark:hover:bg-white/[0.06]">取消</button>
        <button data-testid="confirm-yes" @click="$emit('respond', '', true)"
                class="px-3 py-2 rounded-lg text-[12px] bg-blue-500 hover:bg-blue-600 text-white">确认</button>
      </div>

      <form v-else class="mt-4" @submit.prevent="$emit('respond', text, true)">
        <textarea
          v-if="request.method === 'editor'"
          v-model="text"
          rows="7"
          :placeholder="request.placeholder"
          class="w-full rounded-lg border border-gray-200 dark:border-white/[0.1] bg-gray-50 dark:bg-white/[0.04]
                 px-3 py-2 text-[13px] text-gray-900 dark:text-white outline-none focus:border-blue-400 resize-y"
        />
        <input
          v-else
          v-model="text"
          :placeholder="request.placeholder"
          class="w-full rounded-lg border border-gray-200 dark:border-white/[0.1] bg-gray-50 dark:bg-white/[0.04]
                 px-3 py-2 text-[13px] text-gray-900 dark:text-white outline-none focus:border-blue-400"
        />
        <div class="mt-4 flex justify-end gap-2">
          <button type="button" @click="$emit('cancel')"
                  class="px-3 py-2 rounded-lg text-[12px] text-gray-500 hover:bg-gray-100 dark:hover:bg-white/[0.06]">取消</button>
          <button type="submit" class="px-3 py-2 rounded-lg text-[12px] bg-blue-500 hover:bg-blue-600 text-white">提交</button>
        </div>
      </form>
    </div>
  </div>
</template>
