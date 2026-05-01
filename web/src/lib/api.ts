import axios from 'axios'
import type {
  AdminAttachmentListQuery,
  AdminCommentListQuery,
  AdminOverviewResult,
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
  EmailChangeConfirmInput,
  EmailVerifyConfirmInput,
  InitInput,
  InitStatusResult,
  LoginInput,
  PagedResult,
  PasswordResetConfirmInput,
  PasswordResetRequestInput,
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
const CSRF_COOKIE_NAME = 'easydrop_csrf_token'
const CSRF_HEADER_NAME = 'X-CSRF-Token'

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

export function isApiErrorWithStatus(
  error: unknown,
  status: number,
): error is ApiError {
  return error instanceof ApiError && error.status === status
}

export function isUnauthorizedApiError(error: unknown): error is ApiError {
  return isApiErrorWithStatus(error, 401)
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

function normalizeAdminOverviewResult(payload: {
  recent_activity?: Array<{
    comments?: number | null
    date?: string | null
    posts?: number | null
  }> | null
  totals?: {
    attachments?: number | null
    comments?: number | null
    posts?: number | null
    users?: number | null
  } | null
}) {
  return {
    recent_activity: asArray(payload.recent_activity).map((item) => ({
      comments: typeof item.comments === 'number' ? item.comments : 0,
      date: typeof item.date === 'string' ? item.date : '',
      posts: typeof item.posts === 'number' ? item.posts : 0,
    })),
    totals: {
      attachments:
        typeof payload.totals?.attachments === 'number'
          ? payload.totals.attachments
          : 0,
      comments:
        typeof payload.totals?.comments === 'number'
          ? payload.totals.comments
          : 0,
      posts:
        typeof payload.totals?.posts === 'number' ? payload.totals.posts : 0,
      users:
        typeof payload.totals?.users === 'number' ? payload.totals.users : 0,
    },
  }
}

function normalizeBaseUrl(baseUrl: string) {
  return baseUrl.replace(/\/+$/, '')
}

function readCookieValue(name: string) {
  if (typeof document === 'undefined') {
    return ''
  }

  const cookieName = `${name}=`
  const parts = document.cookie.split(';')
  for (const part of parts) {
    const trimmed = part.trim()
    if (!trimmed.startsWith(cookieName)) {
      continue
    }
    return decodeURIComponent(trimmed.slice(cookieName.length))
  }

  return ''
}

export const axiosInstance = axios.create({
  adapter: 'fetch',
  baseURL: API_BASE_URL,
  withCredentials: true,
})

axiosInstance.interceptors.request.use((config) => {
  const method = (config.method ?? 'get').toUpperCase()
  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
    const csrfToken = readCookieValue(CSRF_COOKIE_NAME)
    if (csrfToken) {
      config.headers[CSRF_HEADER_NAME] = csrfToken
    }
  }
  return config
})

async function request<T>(
  path: string,
  options?: {
    method?: string
    data?: unknown
    query?: object
    token?: string | null
  },
) {
  const { query, token, method = 'GET', data } = options ?? {}

  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  try {
    const response = await axiosInstance.request<T>({
      url: path,
      method,
      params: query,
      data,
      headers,
    })
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      const message =
        error.response.data?.message ||
        error.response.statusText ||
        '请求失败'
      throw new ApiError(message, error.response.status)
    }
    throw error
  }
}

export const api = {
  login(input: LoginInput) {
    return request<AuthResult>('/auth/login', {
      method: 'POST',
      data: input,
    })
  },
  getInitStatus() {
    return request<InitStatusResult>('/init/status')
  },
  initializeSystem(input: InitInput) {
    return request<{ message?: string }>('/init', {
      method: 'POST',
      data: input,
    })
  },
  register(input: RegisterInput) {
    return request<{ message?: string }>('/auth/register', {
      method: 'POST',
      data: input,
    })
  },
  requestPasswordReset(input: PasswordResetRequestInput) {
    return request<{ message?: string }>('/auth/password-reset/request', {
      method: 'POST',
      data: input,
    })
  },
  confirmPasswordReset(input: PasswordResetConfirmInput) {
    return request<{ message?: string }>('/auth/password-reset/confirm', {
      method: 'POST',
      data: input,
    })
  },
  confirmVerifyEmail(input: EmailVerifyConfirmInput) {
    return request<{ message?: string }>('/auth/verify-email/confirm', {
      method: 'POST',
      data: input,
    })
  },
  confirmEmailChange(input: EmailChangeConfirmInput) {
    return request<UserDTO>('/auth/email-change/confirm', {
      method: 'POST',
      data: input,
    })
  },
  getCaptchaConfig() {
    return request<CaptchaConfigResult>('/captcha/config')
  },
  logout() {
    return request<{ message?: string }>('/auth/logout', {
      method: 'POST',
    })
  },
  getCurrentUser(token?: string | null) {
    return request<UserDTO>('/users/me', { token })
  },
  updateMyProfile(input: UpdateMyProfileInput, token?: string | null) {
    return request<UserDTO>('/users/me/profile', {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  changeMyPassword(input: ChangeMyPasswordInput, token?: string | null) {
    return request<{ message?: string }>('/users/me/password', {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  requestMyEmailChange(input: ChangeMyEmailInput, token?: string | null) {
    return request<{ message?: string }>('/users/me/email-change', {
      method: 'POST',
      data: input,
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
  createPostComment(
    postId: number,
    input: CreateCommentInput,
    token?: string | null,
  ) {
    return request<CommentDTO>(`/posts/${postId}/comments`, {
      method: 'POST',
      data: input,
      token,
    })
  },
  createAdminPost(input: CreatePostInput, token?: string | null) {
    return request<PostDTO>('/admin/posts', {
      method: 'POST',
      data: input,
      token,
    })
  },
  updateAdminPost(
    postId: number,
    input: UpdatePostInput,
    token?: string | null,
  ) {
    return request<PostDTO>(`/admin/posts/${postId}`, {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  deleteAdminPost(postId: number, token?: string | null) {
    return request<{ message?: string }>(`/admin/posts/${postId}`, {
      method: 'DELETE',
      token,
    })
  },
  deleteAdminComment(commentId: number, token?: string | null) {
    return request<{ message?: string }>(`/admin/comments/${commentId}`, {
      method: 'DELETE',
      token,
    })
  },
  deleteMyComment(commentId: number, token?: string | null) {
    return request<{ message?: string }>(`/users/me/comments/${commentId}`, {
      method: 'DELETE',
      token,
    })
  },
  uploadAttachment(file: File, token?: string | null) {
    const formData = new FormData()
    formData.set('file', file)
    return request<AttachmentDTO>('/attachments', {
      method: 'POST',
      data: formData,
      token,
    })
  },
  updateAdminComment(
    commentId: number,
    input: UpdateCommentInput,
    token?: string | null,
  ) {
    return request<CommentDTO>(`/admin/comments/${commentId}`, {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  updateMyComment(
    commentId: number,
    input: UpdateCommentInput,
    token?: string | null,
  ) {
    return request<CommentDTO>(`/users/me/comments/${commentId}`, {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  getMyComments(
    token?: string | null,
    query?: Record<string, string | number | boolean | undefined>,
  ) {
    return request<PagedResult<CommentDTO>>('/users/me/comments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminUsers(query: AdminUserListQuery, token?: string | null) {
    return request<PagedResult<UserDTO>>('/admin/users', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminOverview(token?: string | null) {
    return request<AdminOverviewResult>('/admin/overview', {
      token,
    }).then(normalizeAdminOverviewResult)
  },
  createAdminUser(input: CreateUserInput, token?: string | null) {
    return request<UserDTO>('/admin/users', {
      method: 'POST',
      data: input,
      token,
    })
  },
  updateAdminUser(
    userId: number,
    input: UpdateUserInput,
    token?: string | null,
  ) {
    return request<UserDTO>(`/admin/users/${userId}`, {
      method: 'PATCH',
      data: input,
      token,
    })
  },
  deleteAdminUser(userId: number, token?: string | null) {
    return request<{ message?: string }>(`/admin/users/${userId}`, {
      method: 'DELETE',
      token,
    })
  },
  uploadAdminUserAvatar(userId: number, file: File, token?: string | null) {
    const formData = new FormData()
    formData.set('avatar', file)
    return request<UserDTO>(`/admin/users/${userId}/avatar`, {
      method: 'POST',
      data: formData,
      token,
    })
  },
  deleteAdminUserAvatar(userId: number, token?: string | null) {
    return request<{ message?: string }>(`/admin/users/${userId}/avatar`, {
      method: 'DELETE',
      token,
    })
  },
  getAdminPosts(query: AdminPostListQuery, token?: string | null) {
    return request<PagedResult<PostDTO>>('/admin/posts', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminPost(postId: number, token?: string | null) {
    return request<PostDTO>(`/admin/posts/${postId}`, { token })
  },
  getAdminComments(query: AdminCommentListQuery, token?: string | null) {
    return request<PagedResult<CommentDTO>>('/admin/comments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  getAdminComment(commentId: number, token?: string | null) {
    return request<CommentDTO>(`/admin/comments/${commentId}`, {
      token,
    })
  },
  getAdminAttachments(query: AdminAttachmentListQuery, token?: string | null) {
    return request<PagedResult<AttachmentDTO>>('/admin/attachments', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  deleteAdminAttachment(attachmentId: number, token?: string | null) {
    return request<{ message?: string }>(`/admin/attachments/${attachmentId}`, {
      method: 'DELETE',
      token,
    })
  },
  batchDeleteAdminAttachments(ids: number[], token?: string | null) {
    return request<AttachmentBatchDeleteResult>(
      '/admin/attachments/batch-delete',
      {
        method: 'POST',
        data: { ids },
        token,
      },
    )
  },
  getAdminSettings(query: AdminSettingListQuery, token?: string | null) {
    return request<PagedResult<SettingItem>>('/admin/settings', {
      query,
      token,
    }).then(normalizePagedResult)
  },
  updateAdminSetting(
    key: string,
    input: UpdateSettingInput,
    token?: string | null,
  ) {
    return request<{ message?: string }>(
      `/admin/settings/${encodeURIComponent(key)}`,
      {
        method: 'PATCH',
        data: input,
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
