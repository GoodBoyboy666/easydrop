"use client"

import MDEditor from '@uiw/react-md-editor'
import type { MDEditorProps } from '@uiw/react-md-editor'
import remarkGfm from 'remark-gfm'
import { markdownRehypePlugins } from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

interface MarkdownEditorProps
  extends Omit<MDEditorProps, 'onChange' | 'preview' | 'value'> {
  onChange: (value: string) => void
  placeholder?: string
  value: string
}

export function MarkdownEditor({
  className,
  height = 240,
  onChange,
  placeholder,
  value,
  ...props
}: MarkdownEditorProps) {
  const { resolvedTheme } = useTheme()

  return (
    <div
      data-color-mode={resolvedTheme}
      className={cn('markdown-editor-shell overflow-hidden rounded-xl', className)}
    >
      <MDEditor
        value={value}
        onChange={(nextValue) => onChange(nextValue ?? '')}
        preview="edit"
        height={height}
        visibleDragbar={false}
        previewOptions={{
          rehypePlugins: markdownRehypePlugins,
          remarkPlugins: [remarkGfm],
        }}
        textareaProps={{
          placeholder,
        }}
        {...props}
      />
    </div>
  )
}
