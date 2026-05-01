'use client'

import { useEffect, useRef, useState } from 'react'
import { useSiteSettings } from '#/lib/site-settings'
import { useTheme } from '#/lib/theme'

interface MetingPlayerProps {
  server?: string
  type?: string
  id?: string
}

const DEFAULT_METING_API = 'https://api.i-meto.com/meting/api'

const VALID_SERVERS = new Set([
  'baidu',
  'kugou',
  'kuwo',
  'migu',
  'netease',
  'tencent',
])

const VALID_TYPES = new Set(['album', 'artist', 'playlist', 'song'])

interface MetingSong {
  title?: string
  name?: string
  author?: string
  artist?: string
  url?: string
  pic?: string
  cover?: string
  lrc?: string
}

function resolveThemeColor() {
  if (typeof document === 'undefined') return ''
  return (
    getComputedStyle(document.documentElement)
      .getPropertyValue('--primary')
      .trim() || ''
  )
}

async function fetchMetingSongs(
  apiUrl: string,
  server: string,
  type: string,
  id: string,
  signal?: AbortSignal,
) {
  const url = `${apiUrl}?server=${encodeURIComponent(server)}&type=${encodeURIComponent(type)}&id=${encodeURIComponent(id)}`
  const response = await fetch(url, { signal })
  if (!response.ok) {
    throw new Error(`Meting API ${response.status}`)
  }
  const data = (await response.json()) as MetingSong[] | MetingSong
  return Array.isArray(data) ? data : [data]
}

export function MetingPlayer({ server, type, id }: MetingPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const apRef = useRef<import('aplayer').default | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { settingsMap } = useSiteSettings()
  const { isDark } = useTheme()

  const normalizedId = id?.trim()
  const normalizedServer = server?.trim() ?? ''
  const normalizedType = type?.trim() ?? ''

  useEffect(() => {
    if (!normalizedId || !containerRef.current) return
    if (!normalizedServer || !VALID_SERVERS.has(normalizedServer)) return
    if (!normalizedType || !VALID_TYPES.has(normalizedType)) return

    const abortController = new AbortController()

    async function init() {
      try {
        const APlayer = (await import('aplayer')).default
        if (abortController.signal.aborted) return

        const apiUrl =
          settingsMap['site.meting_api_url']?.trim() || DEFAULT_METING_API

        const songs = await fetchMetingSongs(
          apiUrl,
          normalizedServer!,
          normalizedType!,
          normalizedId!,
          abortController.signal,
        )
        if (abortController.signal.aborted) return

        const validSongs = songs.filter((s) => s.url)
        if (validSongs.length === 0) {
          if (abortController.signal.aborted) return
          setError('未找到音乐')
          return
        }

        if (apRef.current) {
          apRef.current.destroy()
          apRef.current = null
        }

        if (abortController.signal.aborted) return

        const themeColor = resolveThemeColor()

        apRef.current = new APlayer({
          container: containerRef.current!,
          audio: validSongs.map((s) => ({
            name: s.title || s.name || '',
            artist: s.author || s.artist || '',
            url: s.url || '',
            cover: s.pic || s.cover || '',
            lrc: s.lrc || '',
          })),
          autoplay: false,
          mutex: true,
          listFolded: validSongs.length > 1,
          listMaxHeight: '200px',
          lrcType: 1,
          theme: themeColor,
        })

        setError(null)
      } catch (err) {
        if (abortController.signal.aborted) return
        setError(err instanceof Error ? err.message : '加载失败')
      }
    }

    init()

    return () => {
      abortController.abort()
      if (apRef.current) {
        apRef.current.destroy()
        apRef.current = null
      }
    }
  }, [normalizedServer, normalizedType, normalizedId, settingsMap])

  useEffect(() => {
    if (!apRef.current) return
    const color = resolveThemeColor()
    if (color) {
      apRef.current.theme(color, apRef.current.list.index)
    }
  }, [isDark])

  if (error) {
    return (
      <span className="my-4 block rounded-xl border border-border/70 bg-card/60 p-4 text-sm text-muted-foreground">
        {error}
      </span>
    )
  }

  if (!normalizedId) return null

  return (
    <span className="my-4 block overflow-hidden rounded-xl border border-border/70 bg-card/60">
      <div ref={containerRef} />
    </span>
  )
}
