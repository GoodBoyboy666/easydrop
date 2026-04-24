const FAVICON_SELECTOR = 'link[rel="icon"]'
const DESCRIPTION_SELECTOR = 'meta[name="description"]'

export function syncSiteMetadata(input: {
  siteDescription: string
  siteFaviconUrl: string
  siteName: string
}) {
  if (typeof document === 'undefined') {
    return
  }

  const siteName = input.siteName.trim()
  const siteDescription = input.siteDescription.trim()

  document.title = siteName

  let descriptionMeta = document.head.querySelector<HTMLMetaElement>(
    DESCRIPTION_SELECTOR,
  )
  if (!descriptionMeta) {
    descriptionMeta = document.createElement('meta')
    descriptionMeta.name = 'description'
    document.head.appendChild(descriptionMeta)
  }
  descriptionMeta.content = siteDescription

  let iconLink = document.head.querySelector<HTMLLinkElement>(FAVICON_SELECTOR)
  if (!iconLink) {
    iconLink = document.createElement('link')
    iconLink.rel = 'icon'
    document.head.appendChild(iconLink)
  }
  iconLink.href = input.siteFaviconUrl
}
