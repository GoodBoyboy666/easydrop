"use client"

const NETEASE_SONG_ID_PATTERN = /^\d+$/
const NETEASE_SONG_URL_PATTERN = /(?:^|[?&])id=(\d+)(?:&|$)/
const NETEASE_SANITIZED_ID_PREFIX = 'user-content-'

interface NeteasePlayerProps {
  id?: string
  songid?: string
}

function normalizeSongId(value: string) {
  const trimmedValue = value.trim()
  const normalizedValue = trimmedValue.startsWith(NETEASE_SANITIZED_ID_PREFIX)
    ? trimmedValue.slice(NETEASE_SANITIZED_ID_PREFIX.length)
    : trimmedValue

  if (NETEASE_SONG_ID_PATTERN.test(normalizedValue)) {
    return normalizedValue
  }

  const matchedSongId = normalizedValue.match(NETEASE_SONG_URL_PATTERN)?.[1]

  if (matchedSongId && NETEASE_SONG_ID_PATTERN.test(matchedSongId)) {
    return matchedSongId
  }

  return null
}

export function NeteasePlayer({ id, songid }: NeteasePlayerProps) {
  const normalizedSongId = songid
    ? normalizeSongId(songid)
    : id
      ? normalizeSongId(id)
      : null

  if (!normalizedSongId) {
    return null
  }

  return (
    <span className="my-4 block overflow-hidden rounded-xl border border-border/70 bg-card/60">
      <span className="block h-[86px] w-full">
        <iframe
          className="size-full border-0"
          referrerPolicy="strict-origin-when-cross-origin"
          src={`https://music.163.com/outchain/player?type=2&id=${normalizedSongId}&auto=0&height=66`}
          title={`网易云音乐 ${normalizedSongId}`}
        />
      </span>
    </span>
  )
}
