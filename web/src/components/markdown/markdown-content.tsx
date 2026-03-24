'use client'

import ReactMarkdown from 'react-markdown'
import {
  markdownComponents,
  markdownRehypePlugins,
  markdownRemarkPlugins,
} from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

interface MarkdownContentProps {
  className?: string
  compact?: boolean
  content: string
}

export function MarkdownContent({
  className,
  compact = false,
  content,
}: MarkdownContentProps) {
  const { resolvedTheme } = useTheme()

  return (
    <div
      data-color-mode={resolvedTheme}
      className={cn(
        'markdown-content rounded-xl',
        compact ? 'markdown-content-compact' : null,
        className,
      )}
    >
      <div className="markdown-body">
        <ReactMarkdown
          components={markdownComponents}
          rehypePlugins={markdownRehypePlugins}
          remarkPlugins={markdownRemarkPlugins}
        >
          {content}
        </ReactMarkdown>
      </div>
    </div>
  )
}
