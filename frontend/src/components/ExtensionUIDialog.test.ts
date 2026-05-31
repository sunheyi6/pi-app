import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ExtensionUIDialog from './ExtensionUIDialog.vue'

describe('ExtensionUIDialog', () => {
  it('returns a selected option', async () => {
    const wrapper = mount(ExtensionUIDialog, {
      props: {
        request: { type: 'extension_ui_request', id: '1', method: 'select', title: 'Choose', options: ['A', 'B'] },
      },
    })
    await wrapper.get('[data-testid="select-option-B"]').trigger('click')
    expect(wrapper.emitted('respond')?.[0]).toEqual(['B', true])
  })

  it('returns confirm false when declined', async () => {
    const wrapper = mount(ExtensionUIDialog, {
      props: {
        request: { type: 'extension_ui_request', id: '2', method: 'confirm', title: 'Continue?' },
      },
    })
    await wrapper.get('[data-testid="confirm-no"]').trigger('click')
    expect(wrapper.emitted('respond')?.[0]).toEqual(['', false])
  })

  it('submits editor text', async () => {
    const wrapper = mount(ExtensionUIDialog, {
      props: {
        request: { type: 'extension_ui_request', id: '3', method: 'editor', title: 'Edit', prefill: 'draft' },
      },
    })
    await wrapper.get('textarea').setValue('final text')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('respond')?.[0]).toEqual(['final text', true])
  })
})
