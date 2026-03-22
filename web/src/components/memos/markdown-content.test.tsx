/** @vitest-environment jsdom */

import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { MarkdownContent, sanitizeUrl } from './markdown-content'

describe('sanitizeUrl', () => {
  it('allows safe protocols and relative paths', () => {
    expect(sanitizeUrl('https://example.com')).toBe('https://example.com')
    expect(sanitizeUrl('/posts/1')).toBe('/posts/1')
  })

  it('blocks javascript protocol', () => {
    expect(sanitizeUrl('javascript:alert(1)')).toBeNull()
  })
})

describe('MarkdownContent', () => {
  it('does not render dangerous links as anchors', () => {
    render(<MarkdownContent content="[危险链接](javascript:alert(1))" />)

    expect(screen.queryByRole('link', { name: '危险链接' })).toBeNull()
    expect(screen.getByText('危险链接')).not.toBeNull()
  })

  it('skips raw html', () => {
    const { container } = render(
      <MarkdownContent content={'安全文本\n\n<script>alert(1)</script>'} />,
    )

    expect(screen.getByText('安全文本')).not.toBeNull()
    expect(container.querySelector('script')).toBeNull()
  })
})
