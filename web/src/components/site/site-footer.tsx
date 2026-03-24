'use client'

import { useQuery } from '@tanstack/react-query'
import { hitokotoQueryOptions } from '#/lib/query-options'

export function SiteFooter() {
  const hitokotoQuery = useQuery(hitokotoQueryOptions())
  const hitokoto = hitokotoQuery.data ?? {
    text: '把正在做的事情做到位，本身就是一种答案。',
    source: 'EasyDrop',
  }

  return (
    <footer className="border-t border-border/70 bg-background/90">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-2 px-4 py-6 text-sm text-muted-foreground sm:px-6 lg:px-8 md:flex-row md:items-center md:justify-between">
        <div className="min-w-0 max-w-3xl">
          <span className="text-foreground/85">“{hitokoto.text}”</span>
          <span className="ml-2 text-xs text-muted-foreground">
            - {hitokoto.source}
          </span>
        </div>
        <div>© {new Date().getFullYear()} EasyDrop</div>
      </div>
    </footer>
  )
}
