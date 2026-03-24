import type {
  AuthResult,
  CaptchaConfigResult,
  CommentDTO,
  CreateCommentInput,
  CreatePostInput,
  InitInput,
  InitStatusResult,
  LoginInput,
  PagedResult,
  PostDTO,
  PublicPostListResult,
  PublicSettingsMap,
  RegisterInput,
  SettingPublicResult,
  TagDTO,
  UserDTO,
} from '#/lib/types'

const DEFAULT_API_BASE_URL = '/api/v1'

export const API_BASE_URL = normalizeBaseUrl(
  import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL
)

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

function asArray<T>(value: T[] | null | undefined) {
  return Array.isArray(value) ? value : []
}

function normalizePagedResult<T>(payload: {
  items?: T[] | null
  total?: number | null
}) {
  return {
    items: asArray(payload.items),
    total: typeof payload.total === 'number' ? payload.total : 0,
  }
}

function normalizePublicPostListResult(payload: {
  items?: PostDTO[] | null
  pinned_items?: PostDTO[] | null
  total?: number | null
}) {
  return {
    items: asArray(payload.items),
    pinnedItems: asArray(payload.pinned_items),
    total: typeof payload.total === 'number' ? payload.total : 0,
  }
}

function normalizeSettingPublicResult(payload: {
  items?: Array<{ key: string; value: string }> | null
}) {
  return {
    items: asArray(payload.items),
  }
}

function normalizeBaseUrl(baseUrl: string) {
  return baseUrl.replace(/\/+$/, '')
}

function buildUrl(path: string, query?: Record<string, string | number | undefined>) {
  const pathname = path.startsWith('/') ? path : `/${path}`
  const url = new URL(`${API_BASE_URL}${pathname}`, 'http://localhost')

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== '') {
        url.searchParams.set(key, String(value))
      }
    }
  }

  return `${url.pathname}${url.search}`
}

async function parseResponse<T>(response: Response) {
  const contentType = response.headers.get('content-type') ?? ''
  const isJson = contentType.includes('application/json')
  const payload = isJson ? ((await response.json()) as T | { message?: string }) : null

  if (!response.ok) {
    const message =
      typeof payload === 'object' && payload && 'message' in payload
        ? payload.message || '请求失败'
        : response.statusText || '请求失败'
    throw new ApiError(message, response.status)
  }

  return payload as T
}

async function request<T>(
  path: string,
  init?: RequestInit & {
    query?: Record<string, string | number | undefined>
    token?: string | null
  }
) {
  const { query, token, headers, ...rest } = init ?? {}
  const response = await fetch(buildUrl(path, query), {
    ...rest,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headers,
    },
  })

  return parseResponse<T>(response)
}

export const api = {
  login(input: LoginInput) {
    return request<AuthResult>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  getInitStatus() {
    return request<InitStatusResult>('/init/status')
  },
  initializeSystem(input: InitInput) {
    return request<{ message?: string }>('/init', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  register(input: RegisterInput) {
    return request<AuthResult>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  getCaptchaConfig() {
    return request<CaptchaConfigResult>('/captcha/config')
  },
  getCurrentUser(token: string) {
    return request<UserDTO>('/users/me', { token })
  },
  getPosts(
    query?: Record<string, string | number | undefined>,
    token?: string | null
  ) {
    return request<PublicPostListResult & { pinned_items?: PostDTO[] | null }>('/posts', {
      query,
      token,
    }).then(normalizePublicPostListResult)
  },
  getPublicSettings() {
    return request<SettingPublicResult>('/settings/public').then(
      normalizeSettingPublicResult
    )
  },
  getTags(query?: Record<string, string | number | undefined>) {
    return request<PagedResult<TagDTO>>('/tags', { query }).then(normalizePagedResult)
  },
  getLatestComments(query?: Record<string, string | number | undefined>) {
    return request<PagedResult<CommentDTO>>('/comments', { query }).then(
      normalizePagedResult
    )
  },
  getPostComments(
    postId: number,
    query?: Record<string, string | number | undefined>
  ) {
    return request<PagedResult<CommentDTO>>(`/posts/${postId}/comments`, {
      query,
    }).then(normalizePagedResult)
  },
  createPostComment(postId: number, input: CreateCommentInput, token: string) {
    return request<CommentDTO>(`/posts/${postId}/comments`, {
      method: 'POST',
      body: JSON.stringify(input),
      token,
    })
  },
  createAdminPost(input: CreatePostInput, token: string) {
    return request<PostDTO>('/admin/posts', {
      method: 'POST',
      body: JSON.stringify(input),
      token,
    })
  },
  deleteAdminPost(postId: number, token: string) {
    return request<{ message?: string }>(`/admin/posts/${postId}`, {
      method: 'DELETE',
      token,
    })
  },
  deleteAdminComment(commentId: number, token: string) {
    return request<{ message?: string }>(`/admin/comments/${commentId}`, {
      method: 'DELETE',
      token,
    })
  },
  getMyComments(token: string, query?: Record<string, string | number | undefined>) {
    return request<PagedResult<CommentDTO>>('/users/me/comments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
}

export function toPublicSettingsMap(result: SettingPublicResult): PublicSettingsMap {
  return asArray(result.items).reduce<PublicSettingsMap>((acc, item) => {
    acc[item.key] = item.value
    return acc
  }, {})
}
