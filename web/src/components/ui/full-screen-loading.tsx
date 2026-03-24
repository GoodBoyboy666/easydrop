'use client'

import { LoaderCircleIcon } from 'lucide-react'

interface FullScreenLoadingProps {
  description?: string
  title?: string
}

export function FullScreenLoading({
  description = '正在完成身份校验，请稍候。',
  title = '正在进入 EasyDrop',
}: FullScreenLoadingProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
      <div className="flex max-w-sm flex-col items-center gap-4 text-center">
        <div className="flex size-14 items-center justify-center rounded-full border border-border bg-card shadow-sm">
          <LoaderCircleIcon className="size-6 animate-spin text-primary" />
        </div>
        <div className="space-y-1.5">
          <div className="font-heading text-lg font-semibold">{title}</div>
          <p className="text-sm leading-6 text-muted-foreground">
            {description}
          </p>
        </div>
      </div>
    </div>
  )
}
