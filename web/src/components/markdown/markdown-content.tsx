"use client"

import { useEffect, useState } from 'react'
import remarkGfm from 'remark-gfm'
import {
  markdownComponents,
  markdownRehypePlugins,
} from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

interface MarkdownContentProps {
  className?: string
  compact?: boolean
  content: string
}

type MarkdownPreviewComponent = (typeof import('@uiw/react-md-editor'))['default']['Markdown']

export function MarkdownContent({
  className,
  compact = false,
  content,
}: MarkdownContentProps) {
  const { resolvedTheme } = useTheme()
  const [PreviewComponent, setPreviewComponent] = useState<MarkdownPreviewComponent | null>(null)

  useEffect(() => {
    let cancelled = false

    void import('@uiw/react-md-editor').then((module) => {
      if (!cancelled) {
        setPreviewComponent(() => module.default.Markdown)
      }
    })

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div
      data-color-mode={resolvedTheme}
      className={cn(
        'markdown-content rounded-xl',
        compact ? 'markdown-content-compact' : null,
        className
      )}
    >
      {PreviewComponent ? (
        <PreviewComponent
          components={markdownComponents}
          source={content}
          rehypePlugins={markdownRehypePlugins}
          remarkPlugins={[remarkGfm]}
        />
      ) : null}
    </div>
  )
}
