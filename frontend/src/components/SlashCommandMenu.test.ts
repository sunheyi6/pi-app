import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SlashCommandMenu from './SlashCommandMenu.vue'
import type { CommandInfo } from '../types'

const commands: CommandInfo[] = [
  { name: 'review', description: 'Review changes', source: 'extension' },
  { name: 'refactor', description: 'Refactor code', source: 'prompt' },
  { name: 'skill:search', description: 'Search docs', source: 'skill' },
]

describe('SlashCommandMenu', () => {
  it('filters visible commands from slash text', () => {
    const wrapper = mount(SlashCommandMenu, { props: { commands, query: 'ref' } })
    expect(wrapper.text()).toContain('refactor')
    expect(wrapper.text()).not.toContain('review')
  })

  it('moves selection with arrow keys and selects with enter', async () => {
    const wrapper = mount(SlashCommandMenu, { props: { commands, query: '' } })
    ;(wrapper.vm as any).handleKeydown(new KeyboardEvent('keydown', { key: 'ArrowDown' }))
    ;(wrapper.vm as any).handleKeydown(new KeyboardEvent('keydown', { key: 'Enter' }))
    expect(wrapper.emitted('select')?.[0]?.[0]).toEqual(commands[1])
  })

  it('closes with escape', () => {
    const wrapper = mount(SlashCommandMenu, { props: { commands, query: '' } })
    ;(wrapper.vm as any).handleKeydown(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
})
