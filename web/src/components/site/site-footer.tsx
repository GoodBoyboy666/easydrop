"use client"

import { useEffect, useState } from 'react'

interface HitokotoResponse {
  from?: string
  from_who?: string | null
  hitokoto?: string
}

const FALLBACK_HITOKOTO = {
  text: '把正在做的事情做到位，本身就是一种答案。',
  source: 'EasyDrop',
}

function buildHitokotoSource(payload: HitokotoResponse) {
  const parts = [payload.from_who, payload.from].filter(Boolean)
  return parts.join(' · ') || FALLBACK_HITOKOTO.source
}

export function SiteFooter() {
  const [hitokoto, setHitokoto] = useState(FALLBACK_HITOKOTO)

  useEffect(() => {
    const controller = new AbortController()

    async function loadHitokoto() {
      try {
        const response = await fetch('https://v1.hitokoto.cn/?encode=json&max_length=56', {
          signal: controller.signal,
        })

        if (!response.ok) {
          throw new Error('Hitokoto request failed')
        }

        const payload = (await response.json()) as HitokotoResponse

        if (!payload.hitokoto?.trim()) {
          return
        }

        setHitokoto({
          text: payload.hitokoto.trim(),
          source: buildHitokotoSource(payload),
        })
      } catch {
        // 保留本地兜底文案，避免页脚出现空白。
      }
    }

    void loadHitokoto()

    return () => {
      controller.abort()
    }
  }, [])

  return (
    <footer className="border-t border-border/70 bg-background/90">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-2 px-4 py-6 text-sm text-muted-foreground sm:px-6 lg:px-8 md:flex-row md:items-center md:justify-between">
        <div className="min-w-0 max-w-3xl">
          <span className="text-foreground/85">“{hitokoto.text}”</span>
          <span className="ml-2 text-xs text-muted-foreground">- {hitokoto.source}</span>
        </div>
        <div>© {new Date().getFullYear()} EasyDrop</div>
      </div>
    </footer>
  )
}
