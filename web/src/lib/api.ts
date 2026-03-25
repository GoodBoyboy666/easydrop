import type {
  AdminAttachmentListQuery,
  AdminCommentListQuery,
  AdminPostListQuery,
  AdminSettingListQuery,
  AdminUserListQuery,
  AttachmentBatchDeleteResult,
  AttachmentDTO,
  AuthResult,
  CaptchaConfigResult,
  CommentDTO,
  CreateUserInput,
  CreateCommentInput,
  CreatePostInput,
  ChangeMyEmailInput,
  ChangeMyPasswordInput,
  InitInput,
  InitStatusResult,
  LoginInput,
  PagedResult,
  PostDTO,
  PublicPostListResult,
  PublicSettingsMap,
  RegisterInput,
  SettingItem,
  SettingPublicResult,
  TagDTO,
  UpdateSettingInput,
  UpdateCommentInput,
  UpdateMyProfileInput,
  UpdatePostInput,
  UpdateUserInput,
  UserDTO,
} from '#/lib/types'

const DEFAULT_API_BASE_URL = '/api/v1'

export const API_BASE_URL = normalizeBaseUrl(
  import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL,
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

function isQueryParamValue(value: unknown): value is string | number | boolean {
  return (
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  )
}

function buildUrl(
  path: string,
  query?: object,
) {
  const pathname = path.startsWith('/') ? path : `/${path}`
  const url = new URL(`${API_BASE_URL}${pathname}`, 'http://localhost')

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== '' && isQueryParamValue(value)) {
        url.searchParams.set(key, String(value))
      }
    }
  }

  return `${url.pathname}${url.search}`
}

async function parseResponse<T>(response: Response) {
  const contentType = response.headers.get('content-type') ?? ''
  const isJson = contentType.includes('application/json')
  const payload = isJson
    ? ((await response.json()) as T | { message?: string })
    : null

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
    query?: object
    token?: string | null
  },
) {
  const { query, token, headers, ...rest } = init ?? {}
  const isFormData =
    typeof FormData !== 'undefined' && rest.body instanceof FormData
  const response = await fetch(buildUrl(path, query), {
    ...rest,
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
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
  updateMyProfile(input: UpdateMyProfileInput, token: string) {
    return request<UserDTO>('/users/me/profile', {
      method: 'PATCH',
      body: JSON.stringify(input),
      token,
    })
  },
  changeMyPassword(input: ChangeMyPasswordInput, token: string) {
    return request<{ message?: string }>('/users/me/password', {
      method: 'PATCH',
      body: JSON.stringify(input),
      token,
    })
  },
  requestMyEmailChange(input: ChangeMyEmailInput, token: string) {
    return request<{ message?: string }>('/users/me/email-change', {
      method: 'POST',
      body: JSON.stringify(input),
      token,
    })
  },
  getPosts(
    query?: Record<string, string | number | boolean | undefined>,
    token?: string | null,
  ) {
    return request<PublicPostListResult & { pinned_items?: PostDTO[] | null }>(
      '/posts',
      {
        query,
        token,
      },
    ).then(normalizePublicPostListResult)
  },
  getPost(postId: number, token?: string | null) {
    return request<PostDTO>(`/posts/${postId}`, {
      token,
    })
  },
  getPublicSettings() {
    return request<SettingPublicResult>('/settings/public').then(
      normalizeSettingPublicResult,
    )
  },
  getTags(query?: Record<string, string | number | boolean | undefined>) {
    return request<PagedResult<TagDTO>>('/tags', { query }).then(
      normalizePagedResult,
    )
  },
  getLatestComments(
    query?: Record<string, string | number | boolean | undefined>,
  ) {
    return request<PagedResult<CommentDTO>>('/comments', { query }).then(
      normalizePagedResult,
    )
  },
  getPostComments(
    postId: number,
    query?: Record<string, string | number | boolean | undefined>,
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
  updateAdminPost(postId: number, input: UpdatePostInput, token: string) {
    return request<PostDTO>(`/admin/posts/${postId}`, {
      method: 'PATCH',
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
  deleteMyComment(commentId: number, token: string) {
    return request<{ message?: string }>(`/users/me/comments/${commentId}`, {
      method: 'DELETE',
      token,
    })
  },
  updateAdminComment(
    commentId: number,
    input: UpdateCommentInput,
    token: string,
  ) {
    return request<CommentDTO>(`/admin/comments/${commentId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
      token,
    })
  },
  updateMyComment(commentId: number, input: UpdateCommentInput, token: string) {
    return request<CommentDTO>(`/users/me/comments/${commentId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
      token,
    })
  },
  getMyComments(
    token: string,
    query?: Record<string, string | number | boolean | undefined>,
  ) {
    return request<PagedResult<CommentDTO>>('/users/me/comments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminUsers(query: AdminUserListQuery, token: string) {
    return request<PagedResult<UserDTO>>('/admin/users', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  createAdminUser(input: CreateUserInput, token: string) {
    return request<UserDTO>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(input),
      token,
    })
  },
  updateAdminUser(userId: number, input: UpdateUserInput, token: string) {
    return request<UserDTO>(`/admin/users/${userId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
      token,
    })
  },
  deleteAdminUser(userId: number, token: string) {
    return request<{ message?: string }>(`/admin/users/${userId}`, {
      method: 'DELETE',
      token,
    })
  },
  uploadAdminUserAvatar(userId: number, file: File, token: string) {
    const body = new FormData()
    body.set('avatar', file)
    return request<UserDTO>(`/admin/users/${userId}/avatar`, {
      method: 'POST',
      body,
      token,
    })
  },
  deleteAdminUserAvatar(userId: number, token: string) {
    return request<{ message?: string }>(`/admin/users/${userId}/avatar`, {
      method: 'DELETE',
      token,
    })
  },
  getAdminPosts(query: AdminPostListQuery, token: string) {
    return request<PagedResult<PostDTO>>('/admin/posts', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminPost(postId: number, token: string) {
    return request<PostDTO>(`/admin/posts/${postId}`, { token })
  },
  getAdminComments(query: AdminCommentListQuery, token: string) {
    return request<PagedResult<CommentDTO>>('/admin/comments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminComment(commentId: number, token: string) {
    return request<CommentDTO>(`/admin/comments/${commentId}`, {
      token,
    })
  },
  getAdminAttachments(query: AdminAttachmentListQuery, token: string) {
    return request<PagedResult<AttachmentDTO>>('/admin/attachments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  deleteAdminAttachment(attachmentId: number, token: string) {
    return request<{ message?: string }>(`/admin/attachments/${attachmentId}`, {
      method: 'DELETE',
      token,
    })
  },
  batchDeleteAdminAttachments(ids: number[], token: string) {
    return request<AttachmentBatchDeleteResult>('/admin/attachments/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      token,
    })
  },
  getAdminSettings(query: AdminSettingListQuery, token: string) {
    return request<PagedResult<SettingItem>>('/admin/settings', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  updateAdminSetting(
    key: string,
    input: UpdateSettingInput,
    token: string,
  ) {
    return request<{ message?: string }>(
      `/admin/settings/${encodeURIComponent(key)}`,
      {
        method: 'PATCH',
        body: JSON.stringify(input),
        token,
      },
    )
  },
}

export function toPublicSettingsMap(
  result: SettingPublicResult,
): PublicSettingsMap {
  return asArray(result.items).reduce<PublicSettingsMap>((acc, item) => {
    acc[item.key] = item.value
    return acc
  }, {})
}
