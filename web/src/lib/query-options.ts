import { queryOptions } from '@tanstack/react-query'
import { api, toPublicSettingsMap } from '#/lib/api'
import type {
  AdminAttachmentListQuery,
  AdminCommentListQuery,
  AdminPostListQuery,
  AdminSettingListQuery,
  AdminUserListQuery,
  AttachmentDTO,
  CaptchaConfigResult,
  CommentDTO,
  InitStatusResult,
  PagedResult,
  PostDTO,
  PublicPostListResult,
  PublicSettingsMap,
  SettingItem,
  TagDTO,
  UserDTO,
} from '#/lib/types'

interface HitokotoResponse {
  from?: string
  from_who?: string | null
  hitokoto?: string
}

export interface HitokotoResult {
  source: string
  text: string
}

const FALLBACK_HITOKOTO: HitokotoResult = {
  text: '把正在做的事情做到位，本身就是一种答案。',
  source: 'EasyDrop',
}

function isQueryParamValue(value: unknown): value is string | number | boolean {
  return (
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  )
}

function normalizeQuery<TQuery extends object>(
  query?: TQuery,
) {
  if (!query) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(query)
      .filter(
        ([, value]) => value !== undefined && value !== '' && isQueryParamValue(value),
      )
      .sort(([left], [right]) => left.localeCompare(right)),
  )
}

function authScope(token?: string | null) {
  return token ?? 'anonymous'
}

function buildHitokotoSource(payload: HitokotoResponse) {
  const parts = [payload.from_who, payload.from].filter(Boolean)
  return parts.join(' · ') || FALLBACK_HITOKOTO.source
}

async function getHitokoto() {
  const response = await fetch(
    'https://v1.hitokoto.cn/?encode=json&max_length=56',
  )

  if (!response.ok) {
    throw new Error('Hitokoto request failed')
  }

  const payload = (await response.json()) as HitokotoResponse

  if (!payload.hitokoto?.trim()) {
    return FALLBACK_HITOKOTO
  }

  return {
    text: payload.hitokoto.trim(),
    source: buildHitokotoSource(payload),
  }
}

export const queryKeys = {
  captchaConfig: () => ['captcha-config'] as const,
  currentUser: (token: string) => ['current-user', token] as const,
  hitokoto: () => ['hitokoto'] as const,
  initStatus: () => ['init-status'] as const,
  latestCommentsPrefix: () => ['latest-comments'] as const,
  latestComments: (
    query?: Record<string, string | number | boolean | undefined>,
  ) =>
    ['latest-comments', normalizeQuery(query)] as const,
  myCommentsPrefix: () => ['my-comments'] as const,
  myComments: (
    token: string,
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['my-comments', token, normalizeQuery(query)] as const,
  postComments: (
    postId: number,
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['post-comments', postId, normalizeQuery(query)] as const,
  postCommentsPrefix: (postId: number) => ['post-comments', postId] as const,
  post: (postId: number, token?: string | null) =>
    ['post', authScope(token), postId] as const,
  posts: (
    token?: string | null,
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['posts', authScope(token), normalizeQuery(query)] as const,
  postsPrefix: (token?: string | null) => ['posts', authScope(token)] as const,
  publicSettings: () => ['public-settings'] as const,
  tagsPrefix: () => ['tags'] as const,
  tags: (query?: Record<string, string | number | boolean | undefined>) =>
    ['tags', normalizeQuery(query)] as const,
  adminUsersPrefix: (token?: string | null) =>
    ['admin-users', authScope(token)] as const,
  adminUsers: (token: string, query?: AdminUserListQuery) =>
    ['admin-users', authScope(token), normalizeQuery(query)] as const,
  adminPostsPrefix: (token?: string | null) =>
    ['admin-posts', authScope(token)] as const,
  adminPosts: (token: string, query?: AdminPostListQuery) =>
    ['admin-posts', authScope(token), normalizeQuery(query)] as const,
  adminCommentsPrefix: (token?: string | null) =>
    ['admin-comments', authScope(token)] as const,
  adminComments: (token: string, query?: AdminCommentListQuery) =>
    ['admin-comments', authScope(token), normalizeQuery(query)] as const,
  adminAttachmentsPrefix: (token?: string | null) =>
    ['admin-attachments', authScope(token)] as const,
  adminAttachments: (token: string, query?: AdminAttachmentListQuery) =>
    ['admin-attachments', authScope(token), normalizeQuery(query)] as const,
  adminSettingsPrefix: (token?: string | null) =>
    ['admin-settings', authScope(token)] as const,
  adminSettings: (token: string, query?: AdminSettingListQuery) =>
    ['admin-settings', authScope(token), normalizeQuery(query)] as const,
}

export function captchaConfigQueryOptions() {
  return queryOptions<CaptchaConfigResult>({
    queryKey: queryKeys.captchaConfig(),
    queryFn: () => api.getCaptchaConfig(),
    staleTime: 5 * 60 * 1000,
  })
}

export function currentUserQueryOptions(token: string) {
  return queryOptions<UserDTO>({
    queryKey: queryKeys.currentUser(token),
    queryFn: () => api.getCurrentUser(token),
  })
}

export function initStatusQueryOptions() {
  return queryOptions<InitStatusResult>({
    queryKey: queryKeys.initStatus(),
    queryFn: () => api.getInitStatus(),
  })
}

export function latestCommentsQueryOptions(
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.latestComments(query),
    queryFn: () => api.getLatestComments(query),
  })
}

export function myCommentsQueryOptions(
  token: string,
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.myComments(token, query),
    queryFn: () => api.getMyComments(token, query),
  })
}

export function postCommentsQueryOptions(
  postId: number,
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.postComments(postId, query),
    queryFn: () => api.getPostComments(postId, query),
  })
}

export function postQueryOptions(postId: number, token?: string | null) {
  return queryOptions<PostDTO>({
    queryKey: queryKeys.post(postId, token),
    queryFn: () => api.getPost(postId, token),
  })
}

export function postsQueryOptions(
  token?: string | null,
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PublicPostListResult>({
    queryKey: queryKeys.posts(token, query),
    queryFn: () => api.getPosts(query, token),
  })
}

export function publicSettingsMapQueryOptions() {
  return queryOptions<PublicSettingsMap>({
    queryKey: queryKeys.publicSettings(),
    queryFn: () => api.getPublicSettings().then(toPublicSettingsMap),
    staleTime: 5 * 60 * 1000,
  })
}

export function tagsQueryOptions(
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PagedResult<TagDTO>>({
    queryKey: queryKeys.tags(query),
    queryFn: async () => {
      try {
        return await api.getTags(query)
      } catch (error) {
        if (query?.order !== 'hot_desc') {
          throw error
        }

        return api.getTags({
          ...query,
          order: 'created_at_desc',
        })
      }
    },
  })
}

export function adminUsersQueryOptions(
  token: string,
  query: AdminUserListQuery,
) {
  return queryOptions<PagedResult<UserDTO>>({
    queryKey: queryKeys.adminUsers(token, query),
    queryFn: () => api.getAdminUsers(query, token),
  })
}

export function adminPostsQueryOptions(
  token: string,
  query: AdminPostListQuery,
) {
  return queryOptions<PagedResult<PostDTO>>({
    queryKey: queryKeys.adminPosts(token, query),
    queryFn: () => api.getAdminPosts(query, token),
  })
}

export function adminCommentsQueryOptions(
  token: string,
  query: AdminCommentListQuery,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.adminComments(token, query),
    queryFn: () => api.getAdminComments(query, token),
  })
}

export function adminAttachmentsQueryOptions(
  token: string,
  query: AdminAttachmentListQuery,
) {
  return queryOptions<PagedResult<AttachmentDTO>>({
    queryKey: queryKeys.adminAttachments(token, query),
    queryFn: () => api.getAdminAttachments(query, token),
  })
}

export function adminSettingsQueryOptions(
  token: string,
  query: AdminSettingListQuery,
) {
  return queryOptions<PagedResult<SettingItem>>({
    queryKey: queryKeys.adminSettings(token, query),
    queryFn: () => api.getAdminSettings(query, token),
  })
}

export function hitokotoQueryOptions() {
  return queryOptions<HitokotoResult>({
    queryKey: queryKeys.hitokoto(),
    queryFn: getHitokoto,
    staleTime: 10 * 60 * 1000,
    retry: 0,
  })
}
