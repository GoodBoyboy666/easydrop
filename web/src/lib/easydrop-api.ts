export const API_BASE = '/api/v1'
export const POSTS_PAGE_SIZE = 10
export const COMMENTS_PAGE_SIZE = 20
export const GLOBAL_COMMENTS_PAGE_SIZE = 24
export const TAGS_PAGE_SIZE = 12

export type Tag = {
  id: number
  name: string
  created_at: string
}

export type Post = {
  id: number
  content: string
  created_at: string
  updated_at: string
  hide: boolean
  user_id: number
  tags?: Tag[]
}

export type Comment = {
  id: number
  content: string
  created_at: string
  updated_at: string
  parent_id?: number | null
  post_id: number
  reply_to_user_id?: number | null
  root_id?: number | null
  user_id: number
}

export type PublicSettingItem = {
  key: string
  value: string
}

export type CurrentUser = {
  admin?: boolean
  avatar?: string
  created_at?: string
  email?: string
  email_verified?: boolean
  id?: number
  nickname?: string
  status?: number
  storage_quota?: number
  storage_used?: number
  updated_at?: string
  username?: string
}

export type CreateAdminPostInput = {
  content: string
  hide: boolean
}

export type PostListResult = {
  items?: Post[]
  total?: number
}

export type CommentListResult = {
  items?: Comment[]
  total?: number
}

export type PublicSettingResult = {
  items?: PublicSettingItem[]
}

export type TagListResult = {
  items?: Tag[]
  total?: number
}

export type PublicSettingsRecord = Partial<Record<string, string>>

type QueryValue = string | number | boolean | undefined

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type RequestJsonOptions = {
  body?: unknown
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  token?: string
}

function buildQuery(params: Record<string, QueryValue>) {
  const searchParams = new URLSearchParams()

  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') {
      continue
    }

    searchParams.set(key, String(value))
  }

  const query = searchParams.toString()
  return query ? `?${query}` : ''
}

function buildAuthHeaders(token?: string) {
  if (!token) {
    return {}
  }

  return {
    Authorization: `Bearer ${token}`,
  }
}

async function requestJson<T>(path: string, options: RequestJsonOptions = {}) {
  const { body, method = 'GET', token } = options
  const response = await fetch(`${API_BASE}${path}`, {
    body: body ? JSON.stringify(body) : undefined,
    headers: {
      Accept: 'application/json',
      ...(body ? { 'Content-Type': 'application/json' } : {}),
      ...buildAuthHeaders(token),
    },
    method,
  })

  const isJson = response.headers
    .get('content-type')
    ?.includes('application/json')

  const payload = isJson
    ? ((await response.json()) as T | { message?: string })
    : null

  if (!response.ok) {
    const message =
      payload && 'message' in payload && typeof payload.message === 'string'
        ? payload.message
        : '请求失败，请稍后重试。'

    throw new ApiError(message, response.status)
  }

  return payload as T
}

export function toPublicSettingsRecord(items: PublicSettingItem[]) {
  return items.reduce<PublicSettingsRecord>((record, item) => {
    record[item.key] = item.value
    return record
  }, {})
}

export function getBooleanSetting(settings: PublicSettingsRecord, key: string) {
  const value = settings[key]?.trim().toLowerCase()

  if (!value) {
    return false
  }

  return value === 'true' || value === '1' || value === 'yes' || value === 'on'
}

export function formatUserLabel(userId: number | null | undefined) {
  if (!userId) {
    return '匿名访客'
  }

  return `用户 #${userId}`
}

export function getAccessTokenFromStorage() {
  if (typeof window === 'undefined') {
    return null
  }

  const token = window.localStorage.getItem('access_token')?.trim()
  return token ? token : null
}

export function isRootComment(comment: Comment) {
  const parentId = comment.parent_id ?? 0
  const rootId = comment.root_id ?? 0

  return parentId === 0 && rootId === 0
}

export async function listPosts(offset = 0, limit = POSTS_PAGE_SIZE) {
  return requestJson<PostListResult>(
    `/posts${buildQuery({
      limit,
      offset,
      order: 'created_at_desc',
    })}`,
  )
}

export async function listPostComments(
  postId: number,
  offset = 0,
  limit = COMMENTS_PAGE_SIZE,
) {
  return requestJson<CommentListResult>(
    `/posts/${postId}/comments${buildQuery({
      limit,
      offset,
      order: 'created_at_asc',
    })}`,
  )
}

export async function getPublicSettings() {
  const result = await requestJson<PublicSettingResult>('/settings/public')
  return toPublicSettingsRecord(result.items ?? [])
}

export async function listComments(
  offset = 0,
  limit = GLOBAL_COMMENTS_PAGE_SIZE,
  order = 'created_at_desc',
) {
  return requestJson<CommentListResult>(
    `/comments${buildQuery({
      limit,
      offset,
      order,
    })}`,
  )
}

export async function listTags(
  offset = 0,
  limit = TAGS_PAGE_SIZE,
  order = 'hot_desc',
  keyword?: string,
) {
  return requestJson<TagListResult>(
    `/tags${buildQuery({
      keyword,
      limit,
      offset,
      order,
    })}`,
  )
}

export async function getCurrentUser(token: string) {
  return requestJson<CurrentUser>('/users/me', { token })
}

export async function createAdminPost(
  token: string,
  input: CreateAdminPostInput,
) {
  return requestJson<Post>('/admin/posts', {
    body: input,
    method: 'POST',
    token,
  })
}
