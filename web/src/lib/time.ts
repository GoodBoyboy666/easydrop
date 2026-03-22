const relativeTimeFormatter = new Intl.RelativeTimeFormat('zh-CN', {
  numeric: 'auto',
})

const absoluteTimeFormatter = new Intl.DateTimeFormat('zh-CN', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function toDate(input: string) {
  const date = new Date(input)

  if (Number.isNaN(date.getTime())) {
    return null
  }

  return date
}

export function formatAbsoluteTime(input: string) {
  const date = toDate(input)
  return date ? absoluteTimeFormatter.format(date) : input
}

export function formatRelativeTime(input: string, now = Date.now()) {
  const date = toDate(input)

  if (!date) {
    return input
  }

  const diffInSeconds = Math.round((date.getTime() - now) / 1000)
  const absSeconds = Math.abs(diffInSeconds)

  if (absSeconds < 45) {
    return '刚刚'
  }

  if (absSeconds < 60 * 45) {
    return relativeTimeFormatter.format(
      Math.round(diffInSeconds / 60),
      'minute',
    )
  }

  if (absSeconds < 60 * 60 * 20) {
    return relativeTimeFormatter.format(
      Math.round(diffInSeconds / 3600),
      'hour',
    )
  }

  if (absSeconds < 60 * 60 * 24 * 6) {
    return relativeTimeFormatter.format(
      Math.round(diffInSeconds / (60 * 60 * 24)),
      'day',
    )
  }

  if (absSeconds < 60 * 60 * 24 * 28) {
    return relativeTimeFormatter.format(
      Math.round(diffInSeconds / (60 * 60 * 24 * 7)),
      'week',
    )
  }

  if (absSeconds < 60 * 60 * 24 * 365) {
    return relativeTimeFormatter.format(
      Math.round(diffInSeconds / (60 * 60 * 24 * 30)),
      'month',
    )
  }

  return relativeTimeFormatter.format(
    Math.round(diffInSeconds / (60 * 60 * 24 * 365)),
    'year',
  )
}
