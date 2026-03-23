const relativeTimeFormatter = new Intl.RelativeTimeFormat('zh-CN', {
  numeric: 'auto',
})

export function formatDateTime(input?: string) {
  if (!input) {
    return '未知时间'
  }

  const date = new Date(input)

  if (Number.isNaN(date.getTime())) {
    return input
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export function formatRelativeTime(input?: string) {
  if (!input) {
    return '刚刚'
  }

  const date = new Date(input)

  if (Number.isNaN(date.getTime())) {
    return input
  }

  const diffMs = date.getTime() - Date.now()
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour

  if (Math.abs(diffMs) < hour) {
    return relativeTimeFormatter.format(Math.round(diffMs / minute), 'minute')
  }

  if (Math.abs(diffMs) < day) {
    return relativeTimeFormatter.format(Math.round(diffMs / hour), 'hour')
  }

  if (Math.abs(diffMs) < 7 * day) {
    return relativeTimeFormatter.format(Math.round(diffMs / day), 'day')
  }

  return formatDateTime(input)
}

export function getInitials(name?: string) {
  if (!name) {
    return '访'
  }

  return name.trim().slice(0, 2).toUpperCase()
}

export function safeRedirectPath(path?: string) {
  if (!path || !path.startsWith('/')) {
    return '/'
  }

  return path
}
