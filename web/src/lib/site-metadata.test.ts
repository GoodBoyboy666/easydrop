// @vitest-environment jsdom

import { afterEach, describe, expect, it } from 'vitest'
import { syncSiteMetadata } from '#/lib/site-metadata'

describe('syncSiteMetadata', () => {
  afterEach(() => {
    document.head.innerHTML = ''
    document.title = ''
  })

  it('updates the document title, description meta, and favicon', () => {
    syncSiteMetadata({
      siteDescription: '自定义站点描述',
      siteFaviconUrl: '/custom-favicon.ico',
      siteName: '自定义站点',
    })

    expect(document.title).toBe('自定义站点')

    const descriptionMeta = document.head.querySelector(
      'meta[name="description"]',
    )
    expect(descriptionMeta?.getAttribute('content')).toBe('自定义站点描述')

    const iconLink = document.head.querySelector('link[rel="icon"]')
    expect(iconLink?.getAttribute('href')).toBe('/custom-favicon.ico')
  })

  it('reuses existing head nodes and trims text values', () => {
    const descriptionMeta = document.createElement('meta')
    descriptionMeta.name = 'description'
    descriptionMeta.content = '旧描述'
    document.head.appendChild(descriptionMeta)

    const iconLink = document.createElement('link')
    iconLink.rel = 'icon'
    iconLink.href = '/old.ico'
    document.head.appendChild(iconLink)

    syncSiteMetadata({
      siteDescription: ' 新描述 ',
      siteFaviconUrl: '/new.ico',
      siteName: ' 新站点 ',
    })

    expect(document.title).toBe('新站点')
    expect(
      document.head.querySelectorAll('meta[name="description"]').length,
    ).toBe(1)
    expect(
      document.head.querySelectorAll('link[rel="icon"]').length,
    ).toBe(1)
    expect(descriptionMeta.content).toBe('新描述')
    expect(iconLink.getAttribute('href')).toBe('/new.ico')
  })
})
