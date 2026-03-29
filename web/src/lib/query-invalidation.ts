import type { QueryClient } from '@tanstack/react-query'
import { queryKeys } from '#/lib/query-options'

type QueryKey = readonly unknown[]

async function invalidateQueryKeyBatch(
  queryClient: QueryClient,
  queryKeyList: QueryKey[],
) {
  await Promise.all(
    queryKeyList.map((queryKey) =>
      queryClient.invalidateQueries({
        queryKey,
      }),
    ),
  )
}

export async function invalidatePublicFeedQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.postsPrefix(),
    queryKeys.latestCommentsPrefix(),
    queryKeys.tagsPrefix(),
  ])
}

export async function invalidatePostCardDeleteQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.postPrefix(),
    queryKeys.postsPrefix(),
    queryKeys.latestCommentsPrefix(),
    queryKeys.tagsPrefix(),
  ])
}

export async function invalidatePostCardUpdateQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.postPrefix(),
    queryKeys.postsPrefix(),
    queryKeys.tagsPrefix(),
  ])
}

export async function invalidatePostCommentQueries(
  queryClient: QueryClient,
  postId: number,
) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.postPrefix(),
    queryKeys.postsPrefix(),
    queryKeys.latestCommentsPrefix(),
    queryKeys.postCommentsPrefix(postId),
    queryKeys.myCommentsPrefix(),
  ])
}

export async function invalidateMyCommentQueries(
  queryClient: QueryClient,
  postId: number,
) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.latestCommentsPrefix(),
    queryKeys.myCommentsPrefix(),
    queryKeys.postCommentsPrefix(postId),
  ])
}

export async function invalidateAdminCommentQueries(
  queryClient: QueryClient,
  postId: number,
) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.adminCommentsPrefix(),
    queryKeys.latestCommentsPrefix(),
    queryKeys.postCommentsPrefix(postId),
    queryKeys.myCommentsPrefix(),
  ])
}

export async function invalidateAdminPostQueries(
  queryClient: QueryClient,
  options?: { postId?: number },
) {
  const queryKeyList: QueryKey[] = [
    queryKeys.adminPostsPrefix(),
    queryKeys.postsPrefix(),
    queryKeys.latestCommentsPrefix(),
    queryKeys.adminCommentsPrefix(),
  ]

  if (typeof options?.postId === 'number') {
    queryKeyList.push(queryKeys.postPrefix())
    queryKeyList.push(queryKeys.postCommentsPrefix(options.postId))
  }

  await invalidateQueryKeyBatch(queryClient, queryKeyList)
}

export async function invalidateAdminAttachmentQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.adminAttachmentsPrefix(),
    queryKeys.adminUsersPrefix(),
  ])
}

export async function invalidateAdminUserQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.adminUsersPrefix(),
    queryKeys.currentUser(),
  ])
}

export async function invalidateAdminSettingQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [
    queryKeys.adminSettingsPrefix(),
    queryKeys.publicSettings(),
  ])
}

export async function invalidateInitStatusQueries(queryClient: QueryClient) {
  await invalidateQueryKeyBatch(queryClient, [queryKeys.initStatus()])
}
