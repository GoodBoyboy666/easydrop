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
    <div className="my-4 overflow-hidden rounded-xl border border-border/70 bg-card/60">
      <div className="aspect-video w-full">
        <iframe
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
          allowFullScreen
          className="size-full border-0"
          referrerPolicy="strict-origin-when-cross-origin"
          src={`https://player.bilibili.com/player.html?bvid=${normalizedBvid}&page=1`}
          title={`Bilibili 视频 ${normalizedBvid}`}
        />
      </div>
    </div>
  )
}
