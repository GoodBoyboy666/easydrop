// @vitest-environment jsdom

import { describe, expect, it } from 'vitest'
import { safeHttpUrl, safeRedirectPath } from '#/lib/format'

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

describe('safeHttpUrl', () => {
  it('returns null for empty or malformed input', () => {
    expect(safeHttpUrl()).toBeNull()
    expect(safeHttpUrl('')).toBeNull()
    expect(safeHttpUrl('   ')).toBeNull()
    expect(safeHttpUrl('not-a-url')).toBeNull()
  })

  it('only allows http and https protocols', () => {
    expect(safeHttpUrl('https://example.com/file.png')).toBe(
      'https://example.com/file.png',
    )
    expect(safeHttpUrl('http://example.com/file.png')).toBe(
      'http://example.com/file.png',
    )
    expect(safeHttpUrl('javascript:alert(1)')).toBeNull()
    expect(safeHttpUrl('data:text/plain;base64,SGVsbG8=')).toBeNull()
    expect(safeHttpUrl('ftp://example.com/file.png')).toBeNull()
    expect(safeHttpUrl('//example.com/file.png')).toBeNull()
  })

  it('accepts same-site relative url after normalization', () => {
    expect(safeHttpUrl('/uploads/file/test.png')).toBe(
      `${window.location.origin}/uploads/file/test.png`,
    )
  })
})
