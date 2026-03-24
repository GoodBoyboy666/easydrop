'use client'

import MDEditor from '@uiw/react-md-editor'
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
      <MDEditor.Markdown
        components={markdownComponents}
        source={content}
        rehypePlugins={markdownRehypePlugins}
        remarkPlugins={markdownRemarkPlugins}
      />
    </div>
  )
}
