import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ExtensionPackagesPanel from './ExtensionPackagesPanel.vue'

function installApp(overrides: Record<string, any> = {}) {
  const app = {
    ListPackages: vi.fn().mockResolvedValue(JSON.stringify({
      packages: [{ source: 'npm:pi-web-access', scope: 'project', type: 'npm' }],
    })),
    InstallPackage: vi.fn().mockResolvedValue(JSON.stringify({ restarted: true })),
    RemovePackage: vi.fn().mockResolvedValue(JSON.stringify({ restarted: true })),
    UpdatePackage: vi.fn().mockResolvedValue(JSON.stringify({ restarted: true })),
    UpdateAllPackages: vi.fn().mockResolvedValue(JSON.stringify({ restarted: true })),
    RetryAgentStartup: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
  window.go = { main: { App: app as any } }
  return app
}

function mountPanel() {
  return mount(ExtensionPackagesPanel, {
    global: {
      plugins: [createPinia()],
    },
  })
}

describe('ExtensionPackagesPanel', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    installApp()
  })

  it('loads project packages by default and switches to global scope', async () => {
    const app = installApp()
    const wrapper = mountPanel()
    await flushPromises()

    expect(app.ListPackages).toHaveBeenCalledWith('project')

    await wrapper.get('[data-testid="scope-global"]').trigger('click')
    await flushPromises()
    expect(app.ListPackages).toHaveBeenLastCalledWith('global')
  })

  it('requires a security confirmation before installing', async () => {
    const app = installApp()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('[data-testid="package-source"]').setValue('npm:demo')
    await wrapper.get('[data-testid="install-package"]').trigger('click')

    expect(window.confirm).toHaveBeenCalled()
    expect(app.InstallPackage).not.toHaveBeenCalled()
  })

  it('disables controls while an install is running', async () => {
    let resolveInstall!: (value: string) => void
    const app = installApp({
      InstallPackage: vi.fn().mockImplementation(() => new Promise(resolve => {
        resolveInstall = resolve
      })),
    })
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('[data-testid="package-source"]').setValue('npm:demo')
    await wrapper.get('[data-testid="install-package"]').trigger('click')
    await flushPromises()

    expect(app.InstallPackage).toHaveBeenCalledWith('npm:demo', 'project')
    expect(wrapper.get('[data-testid="install-package"]').attributes('disabled')).toBeDefined()

    resolveInstall(JSON.stringify({ restarted: true }))
    await flushPromises()
  })

  it('shows readable operation errors', async () => {
    installApp({ UpdateAllPackages: vi.fn().mockRejectedValue(new Error('network unavailable')) })
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('[data-testid="update-all"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('network unavailable')
  })

  it('offers update and remove actions for each package row', async () => {
    const app = installApp()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('[data-testid="update-package"]').trigger('click')
    await flushPromises()
    expect(app.UpdatePackage).toHaveBeenCalledWith('npm:pi-web-access')

    await wrapper.get('[data-testid="remove-package"]').trigger('click')
    await flushPromises()
    expect(app.RemovePackage).toHaveBeenCalledWith('npm:pi-web-access', 'project')
  })
})
