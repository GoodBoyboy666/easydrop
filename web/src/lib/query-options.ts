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

function authScope(isAuthenticated?: boolean) {
  return isAuthenticated ? 'authenticated' : 'anonymous'
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
  currentUser: () => ['current-user'] as const,
  hitokoto: () => ['hitokoto'] as const,
  initStatus: () => ['init-status'] as const,
  latestCommentsPrefix: () => ['latest-comments'] as const,
  latestComments: (
    query?: Record<string, string | number | boolean | undefined>,
  ) =>
    ['latest-comments', normalizeQuery(query)] as const,
  myCommentsPrefix: () => ['my-comments'] as const,
  myComments: (
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['my-comments', normalizeQuery(query)] as const,
  postComments: (
    postId: number,
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['post-comments', postId, normalizeQuery(query)] as const,
  postCommentsPrefix: (postId: number) => ['post-comments', postId] as const,
  postPrefix: () => ['post'] as const,
  post: (postId: number, isAuthenticated?: boolean) =>
    ['post', authScope(isAuthenticated), postId] as const,
  posts: (
    isAuthenticated?: boolean,
    query?: Record<string, string | number | boolean | undefined>,
  ) => ['posts', authScope(isAuthenticated), normalizeQuery(query)] as const,
  postsPrefix: () => ['posts'] as const,
  publicSettings: () => ['public-settings'] as const,
  tagsPrefix: () => ['tags'] as const,
  tags: (query?: Record<string, string | number | boolean | undefined>) =>
    ['tags', normalizeQuery(query)] as const,
  adminUsersPrefix: () => ['admin-users'] as const,
  adminUsers: (query?: AdminUserListQuery) =>
    ['admin-users', normalizeQuery(query)] as const,
  adminPostsPrefix: () => ['admin-posts'] as const,
  adminPosts: (query?: AdminPostListQuery) =>
    ['admin-posts', normalizeQuery(query)] as const,
  adminCommentsPrefix: () => ['admin-comments'] as const,
  adminComments: (query?: AdminCommentListQuery) =>
    ['admin-comments', normalizeQuery(query)] as const,
  adminAttachmentsPrefix: () => ['admin-attachments'] as const,
  adminAttachments: (query?: AdminAttachmentListQuery) =>
    ['admin-attachments', normalizeQuery(query)] as const,
  adminSettingsPrefix: () => ['admin-settings'] as const,
  adminSettings: (query?: AdminSettingListQuery) =>
    ['admin-settings', normalizeQuery(query)] as const,
}

export function captchaConfigQueryOptions() {
  return queryOptions<CaptchaConfigResult>({
    queryKey: queryKeys.captchaConfig(),
    queryFn: () => api.getCaptchaConfig(),
    staleTime: 5 * 60 * 1000,
  })
}

export function currentUserQueryOptions() {
  return queryOptions<UserDTO>({
    queryKey: queryKeys.currentUser(),
    queryFn: () => api.getCurrentUser(),
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
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.myComments(query),
    queryFn: () => api.getMyComments(undefined, query),
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

export function postQueryOptions(postId: number, isAuthenticated?: boolean) {
  return queryOptions<PostDTO>({
    queryKey: queryKeys.post(postId, isAuthenticated),
    queryFn: () => api.getPost(postId),
  })
}

export function postsQueryOptions(
  isAuthenticated?: boolean,
  query?: Record<string, string | number | boolean | undefined>,
) {
  return queryOptions<PublicPostListResult>({
    queryKey: queryKeys.posts(isAuthenticated, query),
    queryFn: () => api.getPosts(query),
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
  query: AdminUserListQuery,
) {
  return queryOptions<PagedResult<UserDTO>>({
    queryKey: queryKeys.adminUsers(query),
    queryFn: () => api.getAdminUsers(query),
  })
}

export function adminPostsQueryOptions(
  query: AdminPostListQuery,
) {
  return queryOptions<PagedResult<PostDTO>>({
    queryKey: queryKeys.adminPosts(query),
    queryFn: () => api.getAdminPosts(query),
  })
}

export function adminCommentsQueryOptions(
  query: AdminCommentListQuery,
) {
  return queryOptions<PagedResult<CommentDTO>>({
    queryKey: queryKeys.adminComments(query),
    queryFn: () => api.getAdminComments(query),
  })
}

export function adminAttachmentsQueryOptions(
  query: AdminAttachmentListQuery,
) {
  return queryOptions<PagedResult<AttachmentDTO>>({
    queryKey: queryKeys.adminAttachments(query),
    queryFn: () => api.getAdminAttachments(query),
  })
}

export function adminSettingsQueryOptions(
  query: AdminSettingListQuery,
) {
  return queryOptions<PagedResult<SettingItem>>({
    queryKey: queryKeys.adminSettings(query),
    queryFn: () => api.getAdminSettings(query),
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
