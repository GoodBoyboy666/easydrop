import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { MessageSquareTextIcon, SquarePenIcon, Trash2Icon } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import type { HTMLMotionProps, Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { api, ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatDateTime, formatRelativeTime } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { myCommentsQueryOptions, queryKeys } from '#/lib/query-options'
import type { CommentDTO } from '#/lib/types'
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
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '#/components/ui/empty'
import { Field, FieldError, FieldGroup } from '#/components/ui/field'
import { Separator } from '#/components/ui/separator'
import { Skeleton } from '#/components/ui/skeleton'

export const Route = createFileRoute('/me/comments')({
  component: MyCommentsPage,
})

const MOTION_DELAY_SECONDS = 0.1
const gpuTransformTemplate: NonNullable<
  HTMLMotionProps<'div'>['transformTemplate']
> = (_, generatedTransform) =>
  generatedTransform ? `${generatedTransform} translateZ(0)` : 'translateZ(0)'
const GPU_ACCELERATED_MOTION_PROPS = {
  style: { willChange: 'transform, opacity' },
  transformTemplate: gpuTransformTemplate,
} as const

function MyCommentsPage() {
  const auth = useAuth()
  const queryClient = useQueryClient()
  const prefersReducedMotion = useReducedMotion()
  const [actionError, setActionError] = useState<string | null>(null)
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null)
  const [editingDraft, setEditingDraft] = useState('')
  const [editingError, setEditingError] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<CommentDTO | null>(null)
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)
  const commentsQuery = useQuery({
    ...myCommentsQueryOptions(auth.token ?? '', {
      limit: 20,
      offset: 0,
      order: 'created_at_desc',
    }),
    enabled: !!auth.token,
  })
  const updateCommentMutation = useMutation({
    mutationFn: (comment: CommentDTO) =>
      api.updateMyComment(
        comment.id,
        {
          content: normalizeMarkdownContent(editingDraft),
        },
        auth.token!,
      ),
  })
  const deleteCommentMutation = useMutation({
    mutationFn: (comment: CommentDTO) =>
      api.deleteMyComment(comment.id, auth.token!),
  })
  const comments = commentsQuery.data?.items ?? []
  const loadError =
    commentsQuery.error instanceof Error ? commentsQuery.error.message : null
  const loading = commentsQuery.isPending
  const deleteDialogBusy = deleteCommentMutation.isPending

  useEffect(() => {
    if (prefersReducedMotion) {
      setMotionReady(true)
      return
    }

    setMotionReady(false)
    let raf1 = 0
    let raf2 = 0

    raf1 = window.requestAnimationFrame(() => {
      raf2 = window.requestAnimationFrame(() => {
        setMotionReady(true)
      })
    })

    return () => {
      window.cancelAnimationFrame(raf1)
      window.cancelAnimationFrame(raf2)
    }
  }, [prefersReducedMotion])

  const getEntranceTransition = (delay = 0): Transition =>
    prefersReducedMotion
      ? { duration: 0 }
      : {
          type: 'spring',
          duration: 0.32,
          ease: 'easeOut' as const,
          delay: delay + MOTION_DELAY_SECONDS,
        }

  function redirectToLogin() {
    window.location.assign('/login?redirect=/me/comments')
  }

  function handleUnauthorized(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      auth.logout()
      redirectToLogin()
      return true
    }

    return false
  }

  async function invalidateCommentQueries(comment: CommentDTO) {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: queryKeys.latestCommentsPrefix(),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.myCommentsPrefix(),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.postCommentsPrefix(comment.post_id),
      }),
    ])
  }

  function startEditComment(comment: CommentDTO) {
    if (
      !auth.token ||
      updateCommentMutation.isPending ||
      deleteCommentMutation.isPending
    ) {
      return
    }

    setActionError(null)
    setEditingCommentId(comment.id)
    setEditingDraft(comment.content)
    setEditingError(null)
  }

  function cancelEditComment() {
    if (updateCommentMutation.isPending) {
      return
    }

    setEditingCommentId(null)
    setEditingDraft('')
    setEditingError(null)
  }

  async function handleEditCommentSubmit(
    event: React.FormEvent<HTMLFormElement>,
    comment: CommentDTO,
  ) {
    event.preventDefault()

    if (
      !auth.token ||
      updateCommentMutation.isPending ||
      editingCommentId !== comment.id
    ) {
      return
    }

    if (!hasMarkdownContent(editingDraft)) {
      setEditingError('评论内容不能为空')
      return
    }

    setEditingError(null)
    setActionError(null)

    try {
      await updateCommentMutation.mutateAsync(comment)
      setEditingCommentId(null)
      setEditingDraft('')
      await invalidateCommentQueries(comment)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setEditingError(error instanceof Error ? error.message : '更新评论失败')
    }
  }

  function requestDeleteComment(comment: CommentDTO) {
    if (
      !auth.token ||
      updateCommentMutation.isPending ||
      deleteCommentMutation.isPending
    ) {
      return
    }

    setActionError(null)
    setPendingDelete(comment)
  }

  async function handleDeleteComment() {
    if (!auth.token || !pendingDelete || deleteCommentMutation.isPending) {
      return
    }

    setActionError(null)

    try {
      await deleteCommentMutation.mutateAsync(pendingDelete)
      if (editingCommentId === pendingDelete.id) {
        setEditingCommentId(null)
        setEditingDraft('')
        setEditingError(null)
      }
      await invalidateCommentQueries(pendingDelete)
      setPendingDelete(null)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setActionError(error instanceof Error ? error.message : '删除评论失败')
      setPendingDelete(null)
    }
  }

  if (!auth.token) {
    return (
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8"
        initial={false}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <Alert>
          <AlertTitle>需要先登录</AlertTitle>
          <AlertDescription>
            登录后才可以查看自己发表过的评论。
          </AlertDescription>
        </Alert>
      </motion.div>
    )
  }

  return (
    <motion.div
      animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
      className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8"
      initial={false}
      transition={getEntranceTransition()}
      {...GPU_ACCELERATED_MOTION_PROPS}
    >
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
              删除后这条评论会从“我的评论”和对应日志评论区中移除，此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteDialogBusy}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deleteDialogBusy}
              onClick={() => void handleDeleteComment()}
              variant="destructive"
            >
              {deleteDialogBusy ? '正在删除…' : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
        className="mb-4"
        initial={false}
        transition={getEntranceTransition(0.04)}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <h1 className="font-heading text-2xl font-semibold">我的评论</h1>
      </motion.div>

      {actionError ? (
        <motion.div
          animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          className="mb-4"
          initial={false}
          transition={getEntranceTransition(0.08)}
          {...GPU_ACCELERATED_MOTION_PROPS}
        >
          <Alert variant="destructive">
            <AlertTitle>评论操作失败</AlertTitle>
            <AlertDescription>{actionError}</AlertDescription>
          </Alert>
        </motion.div>
      ) : null}

      {loading ? (
        <motion.div
          animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          className="flex flex-col gap-4"
          initial={false}
          transition={getEntranceTransition(0.1)}
          {...GPU_ACCELERATED_MOTION_PROPS}
        >
          {Array.from({ length: 3 }).map((_, index) => (
            <motion.div
              key={index}
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={false}
              transition={getEntranceTransition(0.12 + index * 0.03)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className="border border-border/70 bg-card/90">
                <CardContent className="flex flex-col gap-3 pt-4">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-8/12" />
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </motion.div>
      ) : null}

      {loadError ? (
        <motion.div
          animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          initial={false}
          transition={getEntranceTransition(0.1)}
          {...GPU_ACCELERATED_MOTION_PROPS}
        >
          <Alert variant="destructive">
            <AlertTitle>我的评论读取失败</AlertTitle>
            <AlertDescription>{loadError}</AlertDescription>
          </Alert>
        </motion.div>
      ) : null}

      {!loading && !loadError && comments.length === 0 ? (
        <motion.div
          animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          initial={false}
          transition={getEntranceTransition(0.12)}
          {...GPU_ACCELERATED_MOTION_PROPS}
        >
          <Empty className="border border-dashed border-border/80 bg-card/80">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <MessageSquareTextIcon />
              </EmptyMedia>
              <EmptyTitle>暂时还没有评论</EmptyTitle>
              <EmptyDescription>
                回到首页，在任意日志下点击“发评论”即可参与讨论。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </motion.div>
      ) : null}

      {!loading && !loadError ? (
        <div className="flex flex-col gap-4">
          <AnimatePresence initial={false}>
            {comments.map((comment, index) => (
              <motion.div
                key={comment.id}
                animate={
                  motionReady
                    ? { opacity: 1, scale: 1, y: 0 }
                    : { opacity: 0, scale: 0.98, y: 12 }
                }
                exit={
                  prefersReducedMotion
                    ? undefined
                    : { opacity: 0, scale: 0.98, y: -8 }
                }
                initial={false}
                layout
                transition={getEntranceTransition(Math.min(index * 0.03, 0.18))}
                {...GPU_ACCELERATED_MOTION_PROPS}
              >
                <Card className="border border-border/70 bg-card/90 shadow-sm">
                  <CardHeader className="px-4">
                    <CardTitle className="text-sm">
                      <Link
                        className="rounded-sm outline-none transition-colors hover:text-primary focus-visible:text-primary focus-visible:ring-2 focus-visible:ring-ring"
                        to="/posts/$id"
                        params={{ id: String(comment.post_id) }}
                      >
                        日志 #{comment.post_id}
                      </Link>
                    </CardTitle>
                    <CardDescription>
                      {formatRelativeTime(comment.created_at)} ·{' '}
                      {formatDateTime(comment.created_at)}
                    </CardDescription>
                  </CardHeader>
                  <Separator />
                  <CardContent className="flex flex-col gap-2 px-4">
                    {editingCommentId === comment.id ? (
                      <form
                        onSubmit={(event) =>
                          void handleEditCommentSubmit(event, comment)
                        }
                      >
                        <FieldGroup>
                          <Field data-invalid={!!editingError}>
                            <MarkdownEditor
                              height={180}
                              onChange={setEditingDraft}
                              placeholder="编辑这条评论，支持 Markdown。"
                              value={editingDraft}
                            />
                            <FieldError>{editingError}</FieldError>
                          </Field>
                        </FieldGroup>

                        <div className="mt-3 flex justify-end gap-2">
                          <Button
                            disabled={updateCommentMutation.isPending}
                            onClick={cancelEditComment}
                            size="sm"
                            type="button"
                            variant="outline"
                          >
                            取消
                          </Button>
                          <Button
                            disabled={updateCommentMutation.isPending}
                            size="sm"
                            type="submit"
                          >
                            {updateCommentMutation.isPending
                              ? '正在保存…'
                              : '保存修改'}
                          </Button>
                        </div>
                      </form>
                    ) : (
                      <>
                        {comment.reply_to_user ? (
                          <div className="text-sm text-muted-foreground">
                            回复 {comment.reply_to_user.nickname}
                          </div>
                        ) : null}
                        <MarkdownContent content={comment.content} />
                        <div className="flex justify-end gap-2">
                          <Button
                            disabled={
                              updateCommentMutation.isPending ||
                              deleteCommentMutation.isPending
                            }
                            onClick={() => startEditComment(comment)}
                            size="sm"
                            type="button"
                            variant="outline"
                          >
                            <SquarePenIcon data-icon="inline-start" />
                            编辑
                          </Button>
                          <Button
                            disabled={
                              updateCommentMutation.isPending ||
                              deleteCommentMutation.isPending
                            }
                            onClick={() => requestDeleteComment(comment)}
                            size="sm"
                            type="button"
                            variant="destructive"
                          >
                            <Trash2Icon data-icon="inline-start" />
                            删除
                          </Button>
                        </div>
                      </>
                    )}
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      ) : null}
    </motion.div>
  )
}
