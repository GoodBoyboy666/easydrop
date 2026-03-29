'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import {
  ChevronDownIcon,
  MessageSquareMoreIcon,
  SendHorizontalIcon,
  Trash2Icon,
  XIcon,
} from 'lucide-react'
import { api, ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatRelativeTime, getInitials } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { postCommentsQueryOptions, queryKeys } from '#/lib/query-options'
import { invalidatePostCommentQueries } from '#/lib/query-invalidation'
import type { CommentDTO, PostDTO } from '#/lib/types'
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { MarkdownEditor } from '#/components/markdown/markdown-editor'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from '#/components/ui/alert-dialog'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '#/components/ui/collapsible'
import { Field, FieldError, FieldGroup } from '#/components/ui/field'
import { Skeleton } from '#/components/ui/skeleton'

interface PostCommentsSectionProps {
  alwaysExpanded?: boolean
  loginRedirectPath?: string
  onCommentTotalChange?: (total: number) => void
  onPostUpdated?: (post: PostDTO) => void
  post: PostDTO
}

type PendingDeleteComment = CommentDTO | null

const INITIAL_COMMENT_PAGE_SIZE = 3
const LOAD_MORE_COMMENT_PAGE_SIZE = 5

export function PostCommentsSection({
  alwaysExpanded = false,
  loginRedirectPath = '/',
  onCommentTotalChange,
  onPostUpdated,
  post,
}: PostCommentsSectionProps) {
  const auth = useAuth()
  const queryClient = useQueryClient()
  const [commentLimit, setCommentLimit] = useState(INITIAL_COMMENT_PAGE_SIZE)
  const [commentError, setCommentError] = useState<string | null>(null)
  const [commentSectionOpen, setCommentSectionOpen] = useState(false)
  const [composerOpen, setComposerOpen] = useState(false)
  const [commentDraft, setCommentDraft] = useState('')
  const [replyTarget, setReplyTarget] = useState<CommentDTO | null>(null)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null)
  const [editingCommentDraft, setEditingCommentDraft] = useState('')
  const [editingCommentError, setEditingCommentError] = useState<string | null>(
    null,
  )
  const [pendingDelete, setPendingDelete] = useState<PendingDeleteComment>(null)

  const commentsQuery = useQuery({
    ...postCommentsQueryOptions(post.id, {
      limit: commentLimit,
      offset: 0,
      order: 'created_at_desc',
    }),
    placeholderData: (previousData) => previousData,
  })
  const toggleCommentAvailabilityMutation = useMutation({
    mutationFn: (input: { disable_comment: boolean }) =>
      api.updateAdminPost(post.id, input),
  })
  const deleteCommentMutation = useMutation({
    mutationFn: (comment: CommentDTO) =>
      auth.isAdmin
        ? api.deleteAdminComment(comment.id)
        : api.deleteMyComment(comment.id),
  })
  const editCommentMutation = useMutation({
    mutationFn: (comment: CommentDTO) =>
      auth.isAdmin
        ? api.updateAdminComment(
            comment.id,
            {
              content: normalizeMarkdownContent(editingCommentDraft),
            },
          )
        : api.updateMyComment(
            comment.id,
            {
              content: normalizeMarkdownContent(editingCommentDraft),
            },
          ),
  })
  const createCommentMutation = useMutation({
    mutationFn: () =>
      api.createPostComment(
        post.id,
        {
          content: normalizeMarkdownContent(commentDraft),
          parent_id: replyTarget?.id,
        },
      ),
  })

  const comments = commentsQuery.data?.items ?? []
  const commentsTotal = commentsQuery.data?.total ?? 0
  const commentsLoadError =
    commentsQuery.error instanceof Error ? commentsQuery.error.message : null

  useEffect(() => {
    onCommentTotalChange?.(commentsTotal)
  }, [commentsTotal, onCommentTotalChange])

  useEffect(() => {
    setCommentLimit(INITIAL_COMMENT_PAGE_SIZE)
    setCommentError(null)
    setReplyTarget(null)
    setCommentDraft('')
    setSubmitError(null)
    setEditingCommentId(null)
    setEditingCommentDraft('')
    setEditingCommentError(null)
    setCommentSectionOpen(alwaysExpanded)
    setComposerOpen(false)
  }, [alwaysExpanded, post.id])

  useEffect(() => {
    if (alwaysExpanded) {
      setCommentSectionOpen(true)
    }
  }, [alwaysExpanded])

  useEffect(() => {
    if (!post.disable_comment) {
      return
    }

    setReplyTarget(null)
    setCommentDraft('')
    setSubmitError(null)
    setComposerOpen(false)
  }, [post.disable_comment])

  function redirectToLogin() {
    window.location.assign(`/login?redirect=${loginRedirectPath}`)
  }

  function handleUnauthorized(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      void auth.logout()
      redirectToLogin()
      return true
    }

    return false
  }

  function handleApiError(error: unknown, fallbackMessage: string) {
    if (handleUnauthorized(error)) {
      return true
    }

    setCommentError(error instanceof Error ? error.message : fallbackMessage)
    return false
  }

  async function invalidatePostQueries() {
    await invalidatePostCommentQueries(queryClient, post.id)
  }

  function isCommentOwner(comment: CommentDTO) {
    return auth.user?.id === comment.author.id
  }

  function canManageComment(comment: CommentDTO) {
    return auth.isAdmin || isCommentOwner(comment)
  }

  function loadMoreComments() {
    setCommentError(null)
    setCommentLimit((current) => current + LOAD_MORE_COMMENT_PAGE_SIZE)
  }

  async function toggleCommentAvailability() {
    if (
      auth.status !== 'authenticated' ||
      !auth.isAdmin ||
      toggleCommentAvailabilityMutation.isPending
    ) {
      return
    }

    setCommentError(null)

    try {
      const updatedPost = await toggleCommentAvailabilityMutation.mutateAsync({
        disable_comment: !post.disable_comment,
      })
      onPostUpdated?.(updatedPost)
      await invalidatePostQueries()
    } catch (error) {
      handleApiError(error, '更新评论区状态失败')
    }
  }

  function handleOpenComposer() {
    if (post.disable_comment) {
      return
    }

    if (auth.status !== 'authenticated') {
      redirectToLogin()
      return
    }

    setSubmitError(null)

    if (replyTarget) {
      setReplyTarget(null)
      setComposerOpen(true)
      return
    }

    setComposerOpen((open) => !open)
  }

  function startReply(comment: CommentDTO) {
    if (post.disable_comment) {
      return
    }

    if (auth.status !== 'authenticated') {
      redirectToLogin()
      return
    }

    setCommentSectionOpen(true)
    setReplyTarget(comment)
    setComposerOpen(true)
    setSubmitError(null)
  }

  function startEditComment(comment: CommentDTO) {
    if (
      auth.status !== 'authenticated' ||
      !canManageComment(comment) ||
      editCommentMutation.isPending ||
      editingCommentId !== null
    ) {
      return
    }

    setCommentSectionOpen(true)
    setEditingCommentId(comment.id)
    setEditingCommentDraft(comment.content)
    setEditingCommentError(null)
  }

  function cancelEditComment() {
    if (editCommentMutation.isPending) {
      return
    }

    setEditingCommentId(null)
    setEditingCommentDraft('')
    setEditingCommentError(null)
  }

  function cancelReply() {
    setReplyTarget(null)
    setSubmitError(null)
  }

  function requestDeleteComment(comment: CommentDTO) {
    if (
      auth.status !== 'authenticated' ||
      !canManageComment(comment) ||
      deleteCommentMutation.isPending ||
      editCommentMutation.isPending
    ) {
      return
    }

    setPendingDelete(comment)
  }

  async function handleDeleteComment(comment: CommentDTO) {
    if (
      auth.status !== 'authenticated' ||
      !canManageComment(comment) ||
      deleteCommentMutation.isPending ||
      editCommentMutation.isPending
    ) {
      return
    }

    setCommentError(null)

    try {
      await deleteCommentMutation.mutateAsync(comment)
      if (replyTarget?.id === comment.id) {
        setReplyTarget(null)
      }
      if (editingCommentId === comment.id) {
        setEditingCommentId(null)
        setEditingCommentDraft('')
        setEditingCommentError(null)
      }
      await invalidatePostQueries()
    } catch (error) {
      handleApiError(error, '删除评论失败')
    } finally {
      setPendingDelete(null)
    }
  }

  async function handleConfirmDelete() {
    if (!pendingDelete) {
      return
    }

    await handleDeleteComment(pendingDelete)
  }

  async function handleEditCommentSubmit(
    event: React.FormEvent<HTMLFormElement>,
    comment: CommentDTO,
  ) {
    event.preventDefault()

    if (
      auth.status !== 'authenticated' ||
      !canManageComment(comment) ||
      editCommentMutation.isPending ||
      editingCommentId !== comment.id
    ) {
      return
    }

    if (!hasMarkdownContent(editingCommentDraft)) {
      setEditingCommentError('评论内容不能为空')
      return
    }

    setEditingCommentError(null)

    try {
      const updatedComment = await editCommentMutation.mutateAsync(comment)

      queryClient.setQueryData<{ items: CommentDTO[]; total: number }>(
        queryKeys.postComments(post.id, {
          limit: commentLimit,
          offset: 0,
          order: 'created_at_desc',
        }),
        (previousData) => {
          if (!previousData) {
            return previousData
          }

          return {
            ...previousData,
            items: previousData.items.map((item) =>
              item.id === updatedComment.id ? updatedComment : item,
            ),
          }
        },
      )

      setEditingCommentId(null)
      setEditingCommentDraft('')
      setEditingCommentError(null)
      await invalidatePostQueries()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setEditingCommentError(
        error instanceof Error ? error.message : '更新评论失败',
      )
    }
  }

  async function handleCommentSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (post.disable_comment) {
      setSubmitError('该日志已关闭评论')
      return
    }

    if (!hasMarkdownContent(commentDraft)) {
      setSubmitError('评论内容不能为空')
      return
    }

    if (auth.status !== 'authenticated') {
      redirectToLogin()
      return
    }

    setSubmitError(null)
    setCommentError(null)

    try {
      await createCommentMutation.mutateAsync()
      setCommentDraft('')
      setCommentSectionOpen(true)
      setComposerOpen(true)
      setReplyTarget(null)
      await invalidatePostQueries()
    } catch (error) {
      if (handleApiError(error, '发表评论失败')) {
        return
      }

      setSubmitError(error instanceof Error ? error.message : '发表评论失败')
    }
  }

  const hasMoreComments = comments.length < commentsTotal
  const commentPlaceholder = replyTarget
    ? `回复 ${replyTarget.author.nickname}，支持 Markdown 评论。`
    : '期待你的发言，支持 Markdown 评论。'
  const deleteDialogBusy = deleteCommentMutation.isPending
  const currentCommentError = commentError ?? commentsLoadError
  const commentsVisible = alwaysExpanded || commentSectionOpen

  return (
    <>
      <AlertDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open && !deleteDialogBusy) {
            setPendingDelete(null)
          }
        }}
      >
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogMedia className="bg-background">
              <Trash2Icon />
            </AlertDialogMedia>
            <AlertDialogTitle>删除这条评论？</AlertDialogTitle>
            <AlertDialogDescription>
              删除后这条评论会立即从当前日志的评论列表移除。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteDialogBusy}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deleteDialogBusy}
              onClick={() => void handleConfirmDelete()}
              variant="destructive"
            >
              {deleteDialogBusy ? '正在删除…' : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between gap-3">
          {alwaysExpanded ? (
            <div className="text-sm font-medium text-foreground">
              评论 ({commentsTotal})
            </div>
          ) : (
            <Collapsible
              open={commentSectionOpen}
              onOpenChange={setCommentSectionOpen}
            >
              <CollapsibleTrigger asChild>
                <Button
                  className="h-auto px-0 py-0 text-sm font-medium text-foreground hover:bg-transparent"
                  size="sm"
                  type="button"
                  variant="ghost"
                >
                  <ChevronDownIcon
                    className={`transition-transform ${
                      commentSectionOpen ? 'rotate-180' : ''
                    }`}
                    data-icon="inline-start"
                  />
                  评论 ({commentsTotal})
                </Button>
              </CollapsibleTrigger>
            </Collapsible>
          )}

          <div className="flex shrink-0 items-center gap-2">
            {auth.isAdmin ? (
              <Button
                disabled={toggleCommentAvailabilityMutation.isPending}
                onClick={() => void toggleCommentAvailability()}
                size="sm"
                type="button"
                variant="outline"
              >
                {toggleCommentAvailabilityMutation.isPending
                  ? '正在更新…'
                  : post.disable_comment
                    ? '开启评论区'
                    : '关闭评论区'}
              </Button>
            ) : null}
            <Button
              disabled={post.disable_comment}
              size="sm"
              variant={
                composerOpen && !post.disable_comment ? 'secondary' : 'outline'
              }
              type="button"
              onClick={handleOpenComposer}
            >
              <MessageSquareMoreIcon data-icon="inline-start" />
              {post.disable_comment ? '评论已关闭' : '发评论'}
            </Button>
          </div>
        </div>

        {composerOpen ? (
          <Collapsible
            className="flex w-full flex-col"
            open={composerOpen}
            onOpenChange={setComposerOpen}
          >
            <CollapsibleContent className="w-full">
              <form
                className="rounded-xl border border-border/70 bg-muted/40 p-3"
                onSubmit={handleCommentSubmit}
              >
                {replyTarget ? (
                  <div className="mb-3 flex items-center justify-between gap-3 rounded-lg border border-border/70 bg-background/80 px-3 py-2 text-sm">
                    <div className="min-w-0 text-muted-foreground">
                      正在回复
                      <span className="ml-1 font-medium text-foreground">
                        {replyTarget.author.nickname}
                      </span>
                    </div>
                    <Button
                      onClick={cancelReply}
                      size="sm"
                      type="button"
                      variant="ghost"
                    >
                      <XIcon data-icon="inline-start" />
                      取消回复
                    </Button>
                  </div>
                ) : null}

                <FieldGroup>
                  <Field data-invalid={!!submitError}>
                    <MarkdownEditor
                      height={120}
                      onChange={setCommentDraft}
                      placeholder={commentPlaceholder}
                      value={commentDraft}
                    />
                    <FieldError>{submitError}</FieldError>
                  </Field>
                </FieldGroup>

                <div className="mt-3 flex items-center justify-end">
                  <Button
                    disabled={createCommentMutation.isPending}
                    type="submit"
                  >
                    <SendHorizontalIcon data-icon="inline-start" />
                    {createCommentMutation.isPending ? '正在提交…' : '发布评论'}
                  </Button>
                </div>
              </form>
            </CollapsibleContent>
          </Collapsible>
        ) : null}

        {commentsVisible ? (
          <Collapsible
            className="flex w-full flex-col"
            open={commentsVisible}
            onOpenChange={alwaysExpanded ? undefined : setCommentSectionOpen}
          >
            <CollapsibleContent className="w-full">
              {commentsQuery.isPending ? (
                <div className="flex flex-col gap-2">
                  {Array.from({ length: 2 }).map((_, index) => (
                    <div key={index} className="px-2.5 pt-2.5 pb-1.5">
                      <div className="flex items-center gap-3">
                        <Skeleton className="size-6 rounded-full" />
                        <div className="flex flex-1 flex-col gap-2">
                          <Skeleton className="h-3.5 w-24" />
                          <Skeleton className="h-3.5 w-full" />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}

              {!commentsQuery.isPending && currentCommentError ? (
                <Alert variant="destructive">
                  <AlertTitle>评论加载失败</AlertTitle>
                  <AlertDescription>{currentCommentError}</AlertDescription>
                </Alert>
              ) : null}

              {!commentsQuery.isPending &&
              !currentCommentError &&
              comments.length > 0 ? (
                <div className="flex flex-col gap-2">
                  {comments.map((comment) => (
                    <article key={comment.id} className="px-2.5 pt-2.5 pb-1.5">
                      <div className="flex items-start gap-3">
                        <Avatar size="sm">
                          <AvatarImage
                            alt={comment.author.nickname}
                            src={comment.author.avatar}
                          />
                          <AvatarFallback>
                            {getInitials(comment.author.nickname)}
                          </AvatarFallback>
                        </Avatar>
                        <div className="min-w-0 flex-1">
                          {editingCommentId === comment.id ? (
                            <form
                              className="rounded-xl border border-border/70 bg-muted/40 p-3"
                              onSubmit={(event) =>
                                void handleEditCommentSubmit(event, comment)
                              }
                            >
                              <FieldGroup>
                                <Field data-invalid={!!editingCommentError}>
                                  <MarkdownEditor
                                    height={120}
                                    onChange={setEditingCommentDraft}
                                    placeholder="编辑这条评论，支持 Markdown。"
                                    value={editingCommentDraft}
                                  />
                                  <FieldError>{editingCommentError}</FieldError>
                                </Field>
                              </FieldGroup>

                              <div className="mt-3 flex items-center justify-end gap-2">
                                <Button
                                  disabled={editCommentMutation.isPending}
                                  onClick={cancelEditComment}
                                  type="button"
                                  variant="outline"
                                >
                                  取消
                                </Button>
                                <Button
                                  disabled={editCommentMutation.isPending}
                                  type="submit"
                                >
                                  {editCommentMutation.isPending
                                    ? '正在保存…'
                                    : '保存修改'}
                                </Button>
                              </div>
                            </form>
                          ) : (
                            <>
                              <div className="flex flex-wrap items-center gap-2 text-[0.8rem]">
                                <span className="font-medium leading-none">
                                  {comment.author.nickname}
                                </span>
                                {comment.author.admin ? (
                                  <Badge className="h-4 px-1.5 text-[10px] leading-none">
                                    管理员
                                  </Badge>
                                ) : null}
                                <span className="text-[0.7rem] text-muted-foreground">
                                  {formatRelativeTime(comment.created_at)}
                                </span>
                                <Button
                                  className="h-auto px-0 py-0 text-[0.7rem] text-muted-foreground"
                                  disabled={editingCommentId !== null}
                                  onClick={() => startReply(comment)}
                                  size="sm"
                                  type="button"
                                  variant="link"
                                >
                                  回复
                                </Button>
                                {canManageComment(comment) ? (
                                  <>
                                    <Button
                                      className="h-auto px-0 py-0 text-[0.7rem] text-muted-foreground"
                                      disabled={
                                        editingCommentId !== null ||
                                        editCommentMutation.isPending
                                      }
                                      onClick={() => startEditComment(comment)}
                                      size="sm"
                                      type="button"
                                      variant="link"
                                    >
                                      编辑
                                    </Button>
                                    <Button
                                      className="h-auto px-0 py-0 text-[0.7rem] text-destructive"
                                      disabled={
                                        deleteCommentMutation.isPending ||
                                        editCommentMutation.isPending ||
                                        editingCommentId !== null
                                      }
                                      onClick={() =>
                                        requestDeleteComment(comment)
                                      }
                                      size="sm"
                                      type="button"
                                      variant="link"
                                    >
                                      删除
                                    </Button>
                                  </>
                                ) : null}
                              </div>
                              {comment.reply_to_user ? (
                                <div className="mt-1 text-[0.7rem] text-muted-foreground">
                                  回复 {comment.reply_to_user.nickname}
                                </div>
                              ) : null}
                              <MarkdownContent
                                compact
                                className="mt-1"
                                content={comment.content}
                              />
                            </>
                          )}
                        </div>
                      </div>
                    </article>
                  ))}

                  {hasMoreComments ? (
                    <Button
                      disabled={commentsQuery.isFetching}
                      onClick={loadMoreComments}
                      size="sm"
                      type="button"
                      variant="outline"
                    >
                      <ChevronDownIcon data-icon="inline-start" />
                      {commentsQuery.isFetching ? '正在加载…' : '更多评论'}
                    </Button>
                  ) : null}
                </div>
              ) : null}
            </CollapsibleContent>
          </Collapsible>
        ) : null}
      </div>
    </>
  )
}
