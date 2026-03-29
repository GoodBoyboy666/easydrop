const relativeTimeFormatter = new Intl.RelativeTimeFormat('zh-CN', {
  numeric: 'auto',
})

export function formatDateTime(
  input?: string,
  options?: { includeYear?: boolean },
) {
  if (!input) {
    return '未知时间'
  }

  const date = new Date(input)

  if (Number.isNaN(date.getTime())) {
    return input
  }

  return new Intl.DateTimeFormat('zh-CN', {
    ...(options?.includeYear ? { year: 'numeric' as const } : {}),
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
  if (!path) {
    return '/'
  }

  const normalizedPath = path.trim()
  if (!normalizedPath.startsWith('/')) {
    return '/'
  }

  if (normalizedPath.startsWith('//') || normalizedPath.startsWith('/\\')) {
    return '/'
  }

  try {
    const origin =
      typeof window !== 'undefined' && window.location.origin
        ? window.location.origin
        : 'http://localhost'
    const redirectUrl = new URL(normalizedPath, origin)

    if (redirectUrl.origin !== origin) {
      return '/'
    }

    return `${redirectUrl.pathname}${redirectUrl.search}${redirectUrl.hash}`
  } catch {
    return '/'
  }
}

export function safeHttpUrl(input?: string | null) {
  if (typeof input !== 'string') {
    return null
  }

  const normalizedInput = input.trim()
  if (!normalizedInput) {
    return null
  }

  const isAbsoluteHttpUrl = /^https?:\/\//i.test(normalizedInput)
  const isRelativePath = normalizedInput.startsWith('/')

  if (!isAbsoluteHttpUrl && !isRelativePath) {
    return null
  }

  if (normalizedInput.startsWith('//') || normalizedInput.startsWith('/\\')) {
    return null
  }

  try {
    const origin =
      typeof window !== 'undefined' && window.location.origin
        ? window.location.origin
        : 'http://localhost'
    const parsedUrl = new URL(normalizedInput, origin)

    if (parsedUrl.protocol !== 'http:' && parsedUrl.protocol !== 'https:') {
      return null
    }

    return parsedUrl.toString()
  } catch {
    return null
  }
}
