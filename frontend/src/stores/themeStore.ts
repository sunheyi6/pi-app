import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type ThemeMode = 'system' | 'light' | 'dark'

const STORAGE_KEY = 'pi-app-theme'

function getStoredTheme(): ThemeMode {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored
    }
  } catch {}
  return 'system'
}

function getSystemTheme(): 'light' | 'dark' {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(getStoredTheme())

  const resolvedTheme = computed<'light' | 'dark'>(() => {
    if (mode.value === 'system') return getSystemTheme()
    return mode.value
  })

  function setTheme(newMode: ThemeMode) {
    mode.value = newMode
    try {
      localStorage.setItem(STORAGE_KEY, newMode)
    } catch {}
    applyTheme()
  }

  function applyTheme() {
    const root = document.documentElement
    if (resolvedTheme.value === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }

  // 监听系统主题变化（仅在 mode 为 system 时生效）
  let mediaQuery: MediaQueryList | null = null

  function startSystemThemeListener() {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      if (mode.value === 'system') {
        applyTheme()
      }
    })
  }

  function stopSystemThemeListener() {
    if (mediaQuery) {
      mediaQuery.removeEventListener('change', applyTheme)
      mediaQuery = null
    }
  }

  return {
    mode,
    resolvedTheme,
    setTheme,
    applyTheme,
    startSystemThemeListener,
    stopSystemThemeListener,
  }
})
