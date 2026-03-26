'use client'

import type { ChangeEvent } from 'react'
import { useEffect, useRef, useState } from 'react'
import type MDEditor from '@uiw/react-md-editor'
import type { MDEditorProps } from '@uiw/react-md-editor'
import type { ICommand } from '@uiw/react-md-editor/commands'
import * as mdCommands from '@uiw/react-md-editor/commands'
import { useMutation } from '@tanstack/react-query'
import {
  ClapperboardIcon,
  Disc3,
  LoaderCircleIcon,
  PaperclipIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import type { AttachmentDTO } from '#/lib/types'
import {
  markdownComponents,
  markdownRehypePlugins,
  markdownRemarkPlugins,
} from '#/lib/markdown-sanitize'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'

const DEFAULT_BILIBILI_BVID = 'BV1xx411c7mD'
const DEFAULT_NETEASE_SONG_ID = '347230'
const BILIBILI_PREFIX = '<bilibili bvid="'
const BILIBILI_SUFFIX = '"></bilibili>'
const NETEASE_PREFIX = '<netease songid="'
const NETEASE_SUFFIX = '"></netease>'

interface PendingAttachmentInsertion {
  end: number
  start: number
}

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
  execute: (state, editorApi) => {
    const selectedBvid = state.selectedText.trim()
    const bvid = selectedBvid || DEFAULT_BILIBILI_BVID
    const template = `${BILIBILI_PREFIX}${bvid}${BILIBILI_SUFFIX}`
    const nextState = editorApi.replaceSelection(template)
    const bvidStart = state.selection.start + BILIBILI_PREFIX.length

    editorApi.setSelectionRange({
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
  execute: (state, editorApi) => {
    const selectedSongId = normalizeNeteaseSongId(state.selectedText)
    const songId = selectedSongId || DEFAULT_NETEASE_SONG_ID
    const template = `${NETEASE_PREFIX}${songId}${NETEASE_SUFFIX}`
    const nextState = editorApi.replaceSelection(template)
    const songIdStart = state.selection.start + NETEASE_PREFIX.length

    editorApi.setSelectionRange({
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

function escapeMarkdownText(value: string) {
  return value.replace(/[[\]\\]/g, '\\$&')
}

function escapeHtmlAttribute(value: string) {
  return value.replace(/&/g, '&amp;').replace(/"/g, '&quot;')
}

function buildAttachmentSnippet(attachment: AttachmentDTO, fileName: string) {
  const normalizedName = escapeMarkdownText(fileName.trim() || '附件')
  const normalizedUrl = `<${attachment.url}>`

  switch (attachment.biz_type) {
    case 1:
      return `![${normalizedName}](${normalizedUrl})`
    case 2:
      return `<video controls src="${escapeHtmlAttribute(attachment.url)}"></video>`
    case 3:
      return `<audio controls src="${escapeHtmlAttribute(attachment.url)}"></audio>`
    default:
      return `[${normalizedName}](${normalizedUrl})`
  }
}

interface MarkdownEditorProps extends Omit<
  MDEditorProps,
  'onChange' | 'preview' | 'value'
> {
  onChange: (value: string) => void
  placeholder?: string
  value: string
}

type MDEditorComponent = typeof MDEditor

export function MarkdownEditor({
  className,
  height = 240,
  onChange,
  placeholder,
  value,
  ...props
}: MarkdownEditorProps) {
  const auth = useAuth()
  const { resolvedTheme } = useTheme()
  const [EditorComponent, setEditorComponent] =
    useState<MDEditorComponent | null>(null)
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const pendingInsertionRef = useRef<PendingAttachmentInsertion | null>(null)
  const valueRef = useRef(value)

  valueRef.current = value

  const uploadAttachmentMutation = useMutation({
    mutationFn: ({ file, token }: { file: File; token: string }) =>
      api.uploadAttachment(file, token),
  })

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

  function insertAttachmentSnippet(snippet: string) {
    const insertion = pendingInsertionRef.current
    pendingInsertionRef.current = null

    if (!insertion) {
      onChange(`${valueRef.current}${valueRef.current ? '\n' : ''}${snippet}`)
      return
    }

    onChange(
      `${valueRef.current.slice(0, insertion.start)}${snippet}${valueRef.current.slice(insertion.end)}`,
    )
    pendingInsertionRef.current = null
  }

  async function handleAttachmentSelection(
    event: ChangeEvent<HTMLInputElement>,
  ) {
    const selectedFile = event.target.files?.[0]
    event.target.value = ''

    if (!selectedFile) {
      pendingInsertionRef.current = null
      return
    }

    if (!auth.token) {
      pendingInsertionRef.current = null
      toast.error('登录后才能上传附件')
      return
    }

    try {
      const attachment = await uploadAttachmentMutation.mutateAsync({
        file: selectedFile,
        token: auth.token,
      })
      insertAttachmentSnippet(
        buildAttachmentSnippet(attachment, selectedFile.name),
      )
      toast.success(`附件 ${selectedFile.name} 上传成功`)
    } catch (error) {
      pendingInsertionRef.current = null
      toast.error(error instanceof Error ? error.message : '上传附件失败')
    }
  }

  const uploadAttachmentCommand: ICommand | null =
    auth.status === 'authenticated'
      ? {
          buttonProps: {
            'aria-label': '上传附件',
            disabled: uploadAttachmentMutation.isPending,
            title: uploadAttachmentMutation.isPending
              ? '上传附件中…'
              : '上传附件',
          },
          icon: uploadAttachmentMutation.isPending ? (
            <LoaderCircleIcon className="size-3.5 animate-spin" />
          ) : (
            <PaperclipIcon className="size-3.5" />
          ),
          keyCommand: 'upload-attachment',
          name: 'upload-attachment',
          execute: (state) => {
            if (uploadAttachmentMutation.isPending) {
              return state
            }

            pendingInsertionRef.current = {
              end: state.selection.end,
              start: state.selection.start,
            }
            fileInputRef.current?.click()
            return state
          },
        }
      : null

  const commands = uploadAttachmentCommand
    ? [
        ...mdCommands.getCommands(),
        mdCommands.divider,
        uploadAttachmentCommand,
        bilibiliCommand,
        neteaseCommand,
      ]
    : [...editorCommands]

  return (
    <div
      data-color-mode={resolvedTheme}
      className={cn(
        'markdown-editor-shell overflow-hidden rounded-xl',
        className,
      )}
    >
      <input
        ref={fileInputRef}
        className="hidden"
        onChange={(event) => void handleAttachmentSelection(event)}
        type="file"
      />

      {EditorComponent ? (
        <EditorComponent
          commands={commands}
          value={value}
          onChange={(nextValue) => onChange(nextValue ?? '')}
          preview="edit"
          height={height}
          visibleDragbar={false}
          previewOptions={{
            components: markdownComponents,
            rehypePlugins: markdownRehypePlugins,
            remarkPlugins: markdownRemarkPlugins,
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
