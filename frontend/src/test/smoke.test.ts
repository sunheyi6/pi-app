import { describe, expect, it } from 'vitest'

describe('frontend test harness', () => {
  it('runs in jsdom', () => {
    expect(document.createElement('div').tagName).toBe('DIV')
  })
})
