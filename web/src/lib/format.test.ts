// @vitest-environment jsdom

import { describe, expect, it } from 'vitest'
import { safeRedirectPath } from '#/lib/format'

describe('safeRedirectPath', () => {
  it('returns root path when redirect is empty or invalid', () => {
    expect(safeRedirectPath()).toBe('/')
    expect(safeRedirectPath('')).toBe('/')
    expect(safeRedirectPath('dashboard')).toBe('/')
    expect(safeRedirectPath('https://evil.example')).toBe('/')
  })

  it('blocks protocol-relative and backslash-prefixed redirects', () => {
    expect(safeRedirectPath('//evil.example/path')).toBe('/')
    expect(safeRedirectPath('/\\evil.example/path')).toBe('/')
  })

  it('keeps in-site redirect paths', () => {
    expect(safeRedirectPath('/me')).toBe('/me')
    expect(safeRedirectPath('/me/comments?tab=mine#latest')).toBe(
      '/me/comments?tab=mine#latest',
    )
  })
})
