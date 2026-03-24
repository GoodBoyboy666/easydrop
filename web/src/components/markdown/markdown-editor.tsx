"use client"

import { useEffect, useState } from 'react'
import type { MDEditorProps } from '@uiw/react-md-editor'
import type { ICommand } from '@uiw/react-md-editor/commands'
import * as mdCommands from '@uiw/react-md-editor/commands'
import { ClapperboardIcon, Disc3 } from 'lucide-react'
import remarkGfm from 'remark-gfm'
import {
  markdownComponents,
  markdownRehypePlugins,
} from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

const DEFAULT_BILIBILI_BVID = 'BV1xx411c7mD'
const DEFAULT_NETEASE_SONG_ID = '347230'
const BILIBILI_PREFIX = '<bilibili bvid="'
const BILIBILI_SUFFIX = '"></bilibili>'
const NETEASE_PREFIX = '<netease songid="'
const NETEASE_SUFFIX = '"></netease>'

function normalizeNeteaseSongId(value: string) {
  const trimmedValue = value.trim()

  if (/^\d+$/.test(trimmedValue)) {
    return trimmedValue
  }

  const matchedSongId = trimmedValue.match(/(?:^|[?&])id=(\d+)(?:&|$)/)?.[1]

  return matchedSongId?.trim() || null
}

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

const neteaseCommand: ICommand = {
  buttonProps: {
    'aria-label': '插入网易云音乐',
    title: '插入网易云音乐',
  },
  icon: <Disc3 className="size-3.5" />,
  keyCommand: 'netease',
  name: 'netease',
  execute: (state, api) => {
    const selectedSongId = normalizeNeteaseSongId(state.selectedText)
    const songId = selectedSongId || DEFAULT_NETEASE_SONG_ID
    const template = `${NETEASE_PREFIX}${songId}${NETEASE_SUFFIX}`
    const nextState = api.replaceSelection(template)
    const songIdStart = state.selection.start + NETEASE_PREFIX.length

    api.setSelectionRange({
      start: songIdStart,
      end: songIdStart + songId.length,
    })

    return nextState
  },
}

const editorCommands: ICommand[] = [
  ...mdCommands.getCommands(),
  mdCommands.divider,
  bilibiliCommand,
  neteaseCommand,
]

interface MarkdownEditorProps
  extends Omit<MDEditorProps, 'onChange' | 'preview' | 'value'> {
  onChange: (value: string) => void
  placeholder?: string
  value: string
}

type MDEditorComponent = (typeof import('@uiw/react-md-editor'))['default']

export function MarkdownEditor({
  className,
  height = 240,
  onChange,
  placeholder,
  value,
  ...props
}: MarkdownEditorProps) {
  const { resolvedTheme } = useTheme()
  const [EditorComponent, setEditorComponent] = useState<MDEditorComponent | null>(null)

  useEffect(() => {
    let cancelled = false

    void import('@uiw/react-md-editor').then((module) => {
      if (!cancelled) {
        setEditorComponent(() => module.default)
      }
    })

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div
      data-color-mode={resolvedTheme}
      className={cn('markdown-editor-shell overflow-hidden rounded-xl', className)}
    >
      {EditorComponent ? (
        <EditorComponent
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
      ) : (
        <div
          className="flex min-h-[inherit] items-center justify-center px-4 py-8 text-sm text-muted-foreground"
          style={{ height }}
        >
          编辑器加载中…
        </div>
      )}
    </div>
  )
}
