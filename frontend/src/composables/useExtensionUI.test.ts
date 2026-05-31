import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useExtensionUI } from './useExtensionUI'

describe('useExtensionUI', () => {
  beforeEach(() => {
    const app = { RespondToExtensionUI: vi.fn().mockResolvedValue(undefined) }
    window.go = { main: { App: app as any } }
    useExtensionUI().reset()
  })

  it('queues blocking dialogs in FIFO order', async () => {
    const ui = useExtensionUI()
    ui.handleRequest({ type: 'extension_ui_request', id: '1', method: 'input', title: 'First' })
    ui.handleRequest({ type: 'extension_ui_request', id: '2', method: 'confirm', title: 'Second' })

    expect(ui.activeRequest.value?.id).toBe('1')
    await ui.respond('done', true)
    expect(ui.activeRequest.value?.id).toBe('2')
  })

  it('turns notify requests into non-blocking notices', () => {
    const ui = useExtensionUI()
    ui.handleRequest({ type: 'extension_ui_request', id: 'n1', method: 'notify', message: 'Saved', notifyType: 'info' })
    expect(ui.notifications.value[0]?.message).toBe('Saved')
    expect(ui.activeRequest.value).toBeNull()
  })

  it('cancels unsupported blocking methods', async () => {
    const ui = useExtensionUI()
    await ui.handleRequest({ type: 'extension_ui_request', id: 'x1', method: 'widget', title: 'Unsupported' })
    expect(window.go?.main?.App?.RespondToExtensionUI).toHaveBeenCalledWith('x1', '', false, true)
  })
})
