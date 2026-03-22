import rehypeSanitize from 'rehype-sanitize'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import { cn } from '#/lib/utils'

type MarkdownContentProps = {
  content: string
  className?: string
}

const allowedProtocols = new Set(['http:', 'https:', 'mailto:', 'tel:'])

export function sanitizeUrl(url: string | null | undefined) {
  if (!url) {
    return null
  }

  const value = url.trim()

  if (
    value.startsWith('/') ||
    value.startsWith('./') ||
    value.startsWith('../') ||
    value.startsWith('#')
  ) {
    return value
  }

  try {
    const parsed = new URL(value)
    return allowedProtocols.has(parsed.protocol) ? value : null
  } catch {
    return null
  }
}

export function MarkdownContent({ content, className }: MarkdownContentProps) {
  return (
    <div className={cn('markdown-content', className)}>
      <ReactMarkdown
        rehypePlugins={[rehypeSanitize]}
        remarkPlugins={[remarkGfm]}
        skipHtml
        components={{
          a({ children, href }) {
            const safeHref = sanitizeUrl(href)

            if (!safeHref) {
              return <span>{children}</span>
            }

            const external = /^https?:\/\//i.test(safeHref)

            return (
              <a
                href={safeHref}
                rel={external ? 'noreferrer noopener' : undefined}
                target={external ? '_blank' : undefined}
              >
                {children}
              </a>
            )
          },
          img({ alt, src }) {
            const safeSrc = sanitizeUrl(src)

            if (!safeSrc) {
              return null
            }

            return <img alt={alt ?? ''} loading="lazy" src={safeSrc} />
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
