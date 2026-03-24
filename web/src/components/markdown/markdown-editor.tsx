"use client"

import MDEditor from '@uiw/react-md-editor'
import type { MDEditorProps } from '@uiw/react-md-editor'
import type { ICommand } from '@uiw/react-md-editor/commands'
import * as mdCommands from '@uiw/react-md-editor/commands'
import { ClapperboardIcon } from 'lucide-react'
import remarkGfm from 'remark-gfm'
import {
  markdownComponents,
  markdownRehypePlugins,
} from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

const DEFAULT_BILIBILI_BVID = 'BV1xx411c7mD'
const BILIBILI_PREFIX = '<bilibili bvid="'
const BILIBILI_SUFFIX = '"></bilibili>'

const bilibiliCommand: ICommand = {
  buttonProps: {
    'aria-label': '插入 Bilibili 视频',
    title: '插入 Bilibili 视频',
  },
  icon: <ClapperboardIcon className="size-3.5" />,
  keyCommand: 'bilibili',
  name: 'bilibili',
  execute: (state, api) => {
    const selectedBvid = state.selectedText.trim()
    const bvid = selectedBvid || DEFAULT_BILIBILI_BVID
    const template = `${BILIBILI_PREFIX}${bvid}${BILIBILI_SUFFIX}`
    const nextState = api.replaceSelection(template)
    const bvidStart = state.selection.start + BILIBILI_PREFIX.length

    api.setSelectionRange({
      start: bvidStart,
      end: bvidStart + bvid.length,
    })

    return nextState
  },
}

const editorCommands: ICommand[] = [
  ...mdCommands.getCommands(),
  mdCommands.divider,
  bilibiliCommand,
]

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
        commands={editorCommands}
        value={value}
        onChange={(nextValue) => onChange(nextValue ?? '')}
        preview="edit"
        height={height}
        visibleDragbar={false}
        previewOptions={{
          components: markdownComponents,
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
