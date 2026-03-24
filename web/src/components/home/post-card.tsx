"use client"

import { useEffect, useState } from 'react'
import {
  ChevronDownIcon,
  MessageSquareMoreIcon,
  SquarePenIcon,
  SendHorizontalIcon,
  Trash2Icon,
  XIcon,
} from 'lucide-react'
import { api, ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatDateTime, formatRelativeTime, getInitials } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
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
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '#/components/ui/collapsible'
import {
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'
import { Skeleton } from '#/components/ui/skeleton'
import { Switch } from '#/components/ui/switch'

interface PostCardProps {
  onPostDeleted?: (postId: number) => void
  post: PostDTO
}

interface CommentState {
  items: CommentDTO[]
  loading: boolean
  loadingMore: boolean
  total: number
}

type PendingDelete =
  | { type: 'post' }
  | { type: 'comment'; comment: CommentDTO }
  | null

const INITIAL_COMMENT_PAGE_SIZE = 3
const LOAD_MORE_COMMENT_PAGE_SIZE = 5

export function PostCard({ onPostDeleted, post }: PostCardProps) {
  const auth = useAuth()
  const [postState, setPostState] = useState(post)
  const [commentState, setCommentState] = useState<CommentState>({
    items: [],
    loading: true,
    loadingMore: false,
    total: 0,
  })
  const [commentError, setCommentError] = useState<string | null>(null)
  const [commentSectionOpen, setCommentSectionOpen] = useState(false)
  const [composerOpen, setComposerOpen] = useState(false)
  const [commentDraft, setCommentDraft] = useState('')
  const [replyTarget, setReplyTarget] = useState<CommentDTO | null>(null)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [updatingCommentSetting, setUpdatingCommentSetting] = useState(false)
  const [deletingCommentId, setDeletingCommentId] = useState<number | null>(null)
  const [deletingPost, setDeletingPost] = useState(false)
  const [editingPost, setEditingPost] = useState(false)
  const [editDraft, setEditDraft] = useState(post.content)
  const [editHidden, setEditHidden] = useState(!!post.hide)
  const [editPinned, setEditPinned] = useState(post.pin != null)
  const [editPin, setEditPin] = useState(post.pin != null ? String(post.pin) : '')
  const [editError, setEditError] = useState<string | null>(null)
  const [savingPostEdit, setSavingPostEdit] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<PendingDelete>(null)

  async function refreshComments(limit = INITIAL_COMMENT_PAGE_SIZE) {
    setCommentError(null)
    setCommentState((current) => ({
      ...current,
      loading: true,
      loadingMore: false,
    }))

    try {
      const result = await api.getPostComments(postState.id, {
        limit,
        offset: 0,
        order: 'created_at_desc',
      })

      setCommentState({
        items: result.items,
        loading: false,
        loadingMore: false,
        total: result.total,
      })
    } catch (error) {
      setCommentError(
        error instanceof Error ? error.message : '加载评论时出现未知错误'
      )
      setCommentState({
        items: [],
        loading: false,
        loadingMore: false,
        total: 0,
      })
    }
  }

  useEffect(() => {
    setPostState(post)
    setEditDraft(post.content)
    setEditHidden(!!post.hide)
    setEditPinned(post.pin != null)
    setEditPin(post.pin != null ? String(post.pin) : '')
    setEditError(null)
    setEditingPost(false)
  }, [post])

  useEffect(() => {
    let cancelled = false

    async function loadComments() {
      try {
        const result = await api.getPostComments(postState.id, {
          limit: INITIAL_COMMENT_PAGE_SIZE,
          offset: 0,
          order: 'created_at_desc',
        })

        if (cancelled) {
          return
        }

        setCommentState({
          items: result.items,
          loading: false,
          loadingMore: false,
          total: result.total,
        })
      } catch (error) {
        if (cancelled) {
          return
        }

        setCommentError(
          error instanceof Error ? error.message : '加载评论时出现未知错误'
        )
        setCommentState({
          items: [],
          loading: false,
          loadingMore: false,
          total: 0,
        })
      }
    }

    void loadComments()

    return () => {
      cancelled = true
    }
  }, [postState.id])

  useEffect(() => {
    setReplyTarget(null)
    setCommentDraft('')
    setSubmitError(null)
    setCommentSectionOpen(false)
    setComposerOpen(false)
  }, [postState.id])

  useEffect(() => {
    if (!postState.disable_comment) {
      return
    }

    setReplyTarget(null)
    setCommentDraft('')
    setSubmitError(null)
    setComposerOpen(false)
  }, [postState.disable_comment])

  function startEditPost() {
    if (!auth.token || !auth.isAdmin || savingPostEdit) {
      return
    }

    setEditDraft(postState.content)
    setEditHidden(!!postState.hide)
    setEditPinned(postState.pin != null)
    setEditPin(postState.pin != null ? String(postState.pin) : '')
    setEditError(null)
    setEditingPost(true)
  }

  function cancelEditPost() {
    if (savingPostEdit) {
      return
    }

    setEditDraft(postState.content)
    setEditHidden(!!postState.hide)
    setEditPinned(postState.pin != null)
    setEditPin(postState.pin != null ? String(postState.pin) : '')
    setEditError(null)
    setEditingPost(false)
  }

  async function loadMoreComments() {
    setCommentError(null)
    setCommentState((current) => ({ ...current, loadingMore: true }))

    try {
      const result = await api.getPostComments(postState.id, {
        limit: LOAD_MORE_COMMENT_PAGE_SIZE,
        offset: commentState.items.length,
        order: 'created_at_desc',
      })

      setCommentState((current) => ({
        items: [...current.items, ...result.items],
        loading: false,
        loadingMore: false,
        total: result.total,
      }))
    } catch (error) {
      setCommentState((current) => ({
        ...current,
        loadingMore: false,
      }))
      setCommentError(error instanceof Error ? error.message : '加载更多评论失败')
    }
  }

  function redirectToLogin() {
    window.location.assign('/login?redirect=/')
  }

  async function toggleCommentAvailability() {
    if (!auth.token || !auth.isAdmin || updatingCommentSetting) {
      return
    }

    setUpdatingCommentSetting(true)
    setCommentError(null)

    try {
      const updatedPost = await api.updateAdminPost(
        postState.id,
        {
          disable_comment: !postState.disable_comment,
        },
        auth.token
      )
      setPostState(updatedPost)
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        auth.logout()
        redirectToLogin()
        return
      }

      setCommentError(
        error instanceof Error ? error.message : '更新评论区状态失败'
      )
    } finally {
      setUpdatingCommentSetting(false)
    }
  }

  function handleOpenComposer() {
    if (postState.disable_comment) {
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
    if (postState.disable_comment) {
      return
    }

    if (auth.status !== 'authenticated' || !auth.token) {
      redirectToLogin()
      return
    }

    setCommentSectionOpen(true)
    setReplyTarget(comment)
    setComposerOpen(true)
    setSubmitError(null)
  }

  function cancelReply() {
    setReplyTarget(null)
    setSubmitError(null)
  }

  function requestDeletePost() {
    if (!auth.token || !auth.isAdmin || deletingPost) {
      return
    }

    setPendingDelete({ type: 'post' })
  }

  function requestDeleteComment(comment: CommentDTO) {
    if (!auth.token || !auth.isAdmin || deletingCommentId) {
      return
    }

    setPendingDelete({ type: 'comment', comment })
  }

  async function handleDeletePost() {
    if (!auth.token || !auth.isAdmin || deletingPost) {
      return
    }

    setDeletingPost(true)
    setCommentError(null)

    try {
      await api.deleteAdminPost(postState.id, auth.token)
      onPostDeleted?.(postState.id)
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        auth.logout()
        redirectToLogin()
        return
      }

      setCommentError(error instanceof Error ? error.message : '删除日志失败')
    } finally {
      setDeletingPost(false)
      setPendingDelete(null)
    }
  }

  async function handleDeleteComment(comment: CommentDTO) {
    if (!auth.token || !auth.isAdmin || deletingCommentId) {
      return
    }

    setDeletingCommentId(comment.id)
    setCommentError(null)

    try {
      await api.deleteAdminComment(comment.id, auth.token)
      if (replyTarget?.id === comment.id) {
        setReplyTarget(null)
      }
      await refreshComments(
        Math.max(INITIAL_COMMENT_PAGE_SIZE, commentState.items.length)
      )
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        auth.logout()
        redirectToLogin()
        return
      }

      setCommentError(error instanceof Error ? error.message : '删除评论失败')
    } finally {
      setDeletingCommentId(null)
      setPendingDelete(null)
    }
  }

  async function handleConfirmDelete() {
    if (!pendingDelete) {
      return
    }

    if (pendingDelete.type === 'post') {
      await handleDeletePost()
      return
    }

    await handleDeleteComment(pendingDelete.comment)
  }

  async function handleEditPostSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!auth.token || !auth.isAdmin || savingPostEdit) {
      return
    }

    if (!hasMarkdownContent(editDraft)) {
      setEditError('日志内容不能为空')
      return
    }

    const normalizedPin = editPin.trim()
    let pin: number | undefined
    let clearPin = false

    if (editPinned) {
      if (!normalizedPin) {
        setEditError('启用置顶后请填写 Pin 值')
        return
      }

      const parsedPin = Number(normalizedPin)

      if (!Number.isInteger(parsedPin) || parsedPin <= 0) {
        setEditError('Pin 必须是大于 0 的整数')
        return
      }

      pin = parsedPin
    } else if (postState.pin != null) {
      clearPin = true
    }

    setSavingPostEdit(true)
    setEditError(null)

    try {
      const updatedPost = await api.updateAdminPost(
        postState.id,
        {
          clear_pin: clearPin || undefined,
          content: normalizeMarkdownContent(editDraft),
          hide: editHidden,
          pin,
        },
        auth.token
      )

      setPostState(updatedPost)
      setEditDraft(updatedPost.content)
      setEditingPost(false)
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        auth.logout()
        redirectToLogin()
        return
      }

      setEditError(error instanceof Error ? error.message : '更新日志失败')
    } finally {
      setSavingPostEdit(false)
    }
  }

  async function handleCommentSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (postState.disable_comment) {
      setSubmitError('该日志已关闭评论')
      return
    }

    if (!hasMarkdownContent(commentDraft)) {
      setSubmitError('评论内容不能为空')
      return
    }

    if (!auth.token) {
      redirectToLogin()
      return
    }

    setSubmitting(true)
    setSubmitError(null)

    try {
      await api.createPostComment(
        postState.id,
        {
          content: normalizeMarkdownContent(commentDraft),
          parent_id: replyTarget?.id,
        },
        auth.token
      )

      setCommentDraft('')
      setCommentSectionOpen(true)
      setComposerOpen(true)
      setReplyTarget(null)

      await refreshComments(
        Math.max(
          LOAD_MORE_COMMENT_PAGE_SIZE,
          commentState.items.length || INITIAL_COMMENT_PAGE_SIZE
        )
      )
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        auth.logout()
        redirectToLogin()
        return
      }

      setSubmitError(error instanceof Error ? error.message : '发表评论失败')
    } finally {
      setSubmitting(false)
    }
  }

  const hasMoreComments = commentState.items.length < commentState.total
  const commentPlaceholder = replyTarget
    ? `回复 ${replyTarget.author.nickname}，支持 Markdown 评论。`
    : '期待你的发言，支持 Markdown 评论。'
  const deleteDialogBusy = deletingPost || deletingCommentId !== null
  const deleteDialogTitle =
    pendingDelete?.type === 'post' ? '删除这条日志？' : '删除这条评论？'
  const deleteDialogDescription =
    pendingDelete?.type === 'post'
      ? '删除后这条日志会立即从当前列表移除，相关评论也会一并消失。此操作不可撤销。'
      : '删除后这条评论会立即从当前日志的评论列表移除。此操作不可撤销。'

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
              <Trash2Icon/>
            </AlertDialogMedia>
            <AlertDialogTitle>{deleteDialogTitle}</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteDialogDescription}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteDialogBusy}>取消</AlertDialogCancel>
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

      <Card className="bg-card/90 shadow-sm">
        <CardHeader className="gap-4 border-b border-border/60">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-3">
            <Avatar size="lg">
              <AvatarImage
                alt={postState.author.nickname}
                src={postState.author.avatar}
              />
              <AvatarFallback>{getInitials(postState.author.nickname)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <CardTitle className="flex flex-wrap items-center gap-2">
                <span className="truncate">{postState.author.nickname}</span>
                {postState.pin ? (
                  <Badge
                    className="h-4 px-1.5 text-[10px] leading-none"
                    variant="outline"
                  >
                    置顶
                  </Badge>
                ) : null}
                {postState.hide ? (
                  <Badge
                    className="h-4 px-1.5 text-[10px] leading-none"
                    variant="secondary"
                  >
                    私密
                  </Badge>
                ) : null}
                {postState.disable_comment ? (
                  <Badge
                    className="h-4 px-1.5 text-[10px] leading-none"
                    variant="outline"
                  >
                    已关闭评论
                  </Badge>
                ) : null}
              </CardTitle>
              <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span>{formatRelativeTime(postState.created_at)}</span>
                <span>发布于 {formatDateTime(postState.created_at)}</span>
              </div>
            </div>
          </div>
          {auth.isAdmin ? (
            <div className="flex shrink-0 items-center gap-1">
              <Button
                aria-label="编辑日志"
                disabled={savingPostEdit}
                onClick={startEditPost}
                size="icon-sm"
                type="button"
                variant={editingPost ? 'secondary' : 'ghost'}
              >
                <SquarePenIcon />
              </Button>
              <Button
                aria-label="删除日志"
                disabled={deletingPost || savingPostEdit}
                onClick={requestDeletePost}
                size="icon-sm"
                type="button"
                variant="ghost"
              >
                <Trash2Icon />
              </Button>
            </div>
          ) : null}
        </div>
      </CardHeader>

        <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-3">
          {editingPost ? (
            <form
              className="rounded-xl border border-border/70 bg-muted/40 p-3"
              onSubmit={handleEditPostSubmit}
            >
                <FieldGroup>
                  <Field data-invalid={!!editError}>
                    <MarkdownEditor
                    height={180}
                    onChange={setEditDraft}
                    placeholder="编辑当前日志内容，支持 Markdown。"
                    value={editDraft}
                    />
                    <FieldError>{editError}</FieldError>
                  </Field>

                  <Field orientation="horizontal">
                    <Switch
                      checked={editPinned}
                      id={`post-edit-pin-enabled-${postState.id}`}
                      onCheckedChange={(checked) => {
                        setEditPinned(checked)
                        if (!checked) {
                          setEditPin('')
                        }
                      }}
                      size="sm"
                    />
                    <FieldContent>
                      <FieldLabel htmlFor={`post-edit-pin-enabled-${postState.id}`}>
                        <FieldTitle>置顶</FieldTitle>
                      </FieldLabel>
                      {editPinned ? (
                        <div className="mt-2">
                          <Input
                            className="w-28"
                            id={`post-edit-pin-${postState.id}`}
                            inputMode="numeric"
                            min={1}
                            onChange={(event) => setEditPin(event.target.value)}
                            placeholder="权重"
                            step={1}
                            type="number"
                            value={editPin}
                          />
                        </div>
                      ) : null}
                    </FieldContent>
                  </Field>

                  <Field orientation="horizontal">
                    <Switch
                      checked={editHidden}
                      id={`post-edit-hidden-${postState.id}`}
                      onCheckedChange={setEditHidden}
                      size="sm"
                    />
                    <FieldContent>
                      <FieldLabel htmlFor={`post-edit-hidden-${postState.id}`}>
                        <FieldTitle>私密发布</FieldTitle>
                      </FieldLabel>
                    </FieldContent>
                  </Field>
                </FieldGroup>

              <div className="mt-3 flex items-center justify-end gap-2">
                <Button
                  disabled={savingPostEdit}
                  onClick={cancelEditPost}
                  type="button"
                  variant="outline"
                >
                  取消
                </Button>
                <Button disabled={savingPostEdit} type="submit">
                  {savingPostEdit ? '正在保存…' : '保存修改'}
                </Button>
              </div>
            </form>
          ) : (
            <MarkdownContent content={postState.content} />
          )}
          {postState.tags?.length ? (
            <div className="flex flex-wrap gap-2">
              {postState.tags.map((tag) => (
                <Badge key={tag.id} variant="secondary">
                  #{tag.name}
                </Badge>
              ))}
            </div>
          ) : null}
        </div>

        <Separator />

        <div className="flex items-center justify-between gap-3">
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
                评论 ({commentState.total})
              </Button>
            </CollapsibleTrigger>
          </Collapsible>

          <div className="flex shrink-0 items-center gap-2">
            {auth.isAdmin ? (
              <Button
                disabled={updatingCommentSetting}
                onClick={() => void toggleCommentAvailability()}
                size="sm"
                type="button"
                variant="outline"
              >
                {updatingCommentSetting
                  ? '正在更新…'
                  : postState.disable_comment
                    ? '开启评论区'
                    : '关闭评论区'}
              </Button>
            ) : null}
            <Button
              disabled={postState.disable_comment}
              size="sm"
              variant={composerOpen && !postState.disable_comment ? 'secondary' : 'outline'}
              type="button"
              onClick={handleOpenComposer}
            >
              <MessageSquareMoreIcon data-icon="inline-start" />
              {postState.disable_comment ? '评论已关闭' : '发评论'}
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
                  <Button disabled={submitting} type="submit">
                    <SendHorizontalIcon data-icon="inline-start" />
                    {submitting ? '正在提交…' : '发布评论'}
                  </Button>
                </div>
              </form>
            </CollapsibleContent>
          </Collapsible>
        ) : null}

        {commentSectionOpen ? (
          <Collapsible
            className="flex w-full flex-col"
            open={commentSectionOpen}
            onOpenChange={setCommentSectionOpen}
          >
            <CollapsibleContent className="w-full">
            {commentState.loading ? (
                <div className="flex flex-col gap-2">
                  {Array.from({ length: 2 }).map((_, index) => (
                    <div
                      key={index}
                      className="px-2.5 pt-2.5 pb-1.5"
                    >
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

            {!commentState.loading && commentError ? (
              <Alert variant="destructive">
                <AlertTitle>评论加载失败</AlertTitle>
                <AlertDescription>{commentError}</AlertDescription>
              </Alert>
            ) : null}

            {!commentState.loading &&
            !commentError &&
            commentState.items.length > 0 ? (
              <div className="flex flex-col gap-2">
                {commentState.items.map((comment) => (
                  <article
                    key={comment.id}
                    className="px-2.5 pt-2.5 pb-1.5"
                  >
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
                        <div className="flex flex-wrap items-center gap-2 text-[0.8rem]">
                          <span className="font-medium leading-none">{comment.author.nickname}</span>
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
                            onClick={() => startReply(comment)}
                            size="sm"
                            type="button"
                            variant="link"
                          >
                            回复
                          </Button>
                          {auth.isAdmin ? (
                            <Button
                              className="h-auto px-0 py-0 text-[0.7rem] text-destructive"
                              disabled={deletingCommentId === comment.id}
                              onClick={() => requestDeleteComment(comment)}
                              size="sm"
                              type="button"
                              variant="link"
                            >
                              删除
                            </Button>
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
                      </div>
                    </div>
                  </article>
                ))}

                {hasMoreComments ? (
                  <Button
                    disabled={commentState.loadingMore}
                    onClick={() => void loadMoreComments()}
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    <ChevronDownIcon data-icon="inline-start" />
                    {commentState.loadingMore ? '正在加载…' : '更多评论'}
                  </Button>
                ) : null}
              </div>
            ) : null}
            </CollapsibleContent>
          </Collapsible>
        ) : null}
        </CardContent>

        <CardFooter className="justify-between gap-3">
          <div className="text-xs text-muted-foreground">
            日志 #{postState.id} · 评论 {commentState.total}
          </div>
        </CardFooter>
      </Card>
    </>
  )
}
