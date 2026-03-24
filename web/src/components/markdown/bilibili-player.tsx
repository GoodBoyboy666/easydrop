"use client"

const BILIBILI_BVID_PATTERN = /^BV[0-9A-Za-z]{10}$/

interface BilibiliPlayerProps {
  bvid?: string
}

function isValidBvid(value: string) {
  return BILIBILI_BVID_PATTERN.test(value)
}

export function BilibiliPlayer({ bvid }: BilibiliPlayerProps) {
  const normalizedBvid = bvid?.trim()

  if (!normalizedBvid || !isValidBvid(normalizedBvid)) {
    return null
  }

  return (
    <span className="my-4 block overflow-hidden rounded-xl border border-border/70 bg-card/60">
      <span className="aspect-video block w-full">
        <iframe
          allow="accelerometer; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
          allowFullScreen
          className="size-full border-0"
          referrerPolicy="strict-origin-when-cross-origin"
          src={`https://player.bilibili.com/player.html?bvid=${normalizedBvid}&page=1&autoplay=0`}
          title={`Bilibili 视频 ${normalizedBvid}`}
        />
      </span>
    </span>
  )
}
