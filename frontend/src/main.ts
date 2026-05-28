import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'

// highlight.js 语法高亮（避免使用 top-level await）
async function initHighlight(): Promise<void> {
  try {
    const hljs = await import('highlight.js')
    const languages: Record<string, () => Promise<any>> = {
      javascript: () => import('highlight.js/lib/languages/javascript'),
      typescript: () => import('highlight.js/lib/languages/typescript'),
      python: () => import('highlight.js/lib/languages/python'),
      bash: () => import('highlight.js/lib/languages/bash'),
      json: () => import('highlight.js/lib/languages/json'),
      xml: () => import('highlight.js/lib/languages/xml'),
      css: () => import('highlight.js/lib/languages/css'),
      sql: () => import('highlight.js/lib/languages/sql'),
      go: () => import('highlight.js/lib/languages/go'),
      rust: () => import('highlight.js/lib/languages/rust'),
      yaml: () => import('highlight.js/lib/languages/yaml'),
      markdown: () => import('highlight.js/lib/languages/markdown'),
      diff: () => import('highlight.js/lib/languages/diff'),
    }

    for (const [name, loader] of Object.entries(languages)) {
      try {
        const mod = await loader()
        hljs.default.registerLanguage(name, mod.default)
      } catch {
        // 某个语言注册失败不影响整体
      }
    }
  } catch {
    console.warn('[main] highlight.js 加载失败，代码高亮将不可用')
  }
}

void initHighlight()

const app = createApp(App)
app.use(createPinia())
app.mount('#app')

console.log('[main] Vue 应用已挂载')
