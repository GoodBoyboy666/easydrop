'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { SquarePenIcon, Trash2Icon } from 'lucide-react'
import { api, ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatDateTime, formatRelativeTime, getInitials } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { queryKeys } from '#/lib/query-options'
import type { PostDTO } from '#/lib/types'
import { PostCommentsSection } from '#/components/post/post-comments-section'
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { MarkdownEditor } from '#/components/markdown/markdown-editor'
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
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'
import { Switch } from '#/components/ui/switch'

interface PostCardProps {
  onPostDeleted?: (postId: number) => void
  onPostUpdated?: (post: PostDTO) => void
  post: PostDTO
  showComments?: boolean
}

export function PostCard({
  onPostDeleted,
  onPostUpdated,
  post,
  showComments = true,
}: PostCardProps) {
  const auth = useAuth()
  const queryClient = useQueryClient()
  const [postState, setPostState] = useState(post)
  const [commentTotal, setCommentTotal] = useState(0)
  const [editingPost, setEditingPost] = useState(false)
  const [editDraft, setEditDraft] = useState(post.content)
  const [editHidden, setEditHidden] = useState(!!post.hide)
  const [editPinned, setEditPinned] = useState(post.pin != null)
  const [editPin, setEditPin] = useState(
    post.pin != null ? String(post.pin) : '',
  )
  const [editError, setEditError] = useState<string | null>(null)
  const [pendingDeletePost, setPendingDeletePost] = useState(false)

  const deletePostMutation = useMutation({
    mutationFn: () => api.deleteAdminPost(postState.id),
  })
  const editPostMutation = useMutation({
    mutationFn: (input: {
      clear_pin?: boolean
      content: string
      hide: boolean
      pin?: number
    }) => api.updateAdminPost(postState.id, input),
  })

  useEffect(() => {
    setPostState(post)
    setCommentTotal(0)
    setEditDraft(post.content)
    setEditHidden(!!post.hide)
    setEditPinned(post.pin != null)
    setEditPin(post.pin != null ? String(post.pin) : '')
    setEditError(null)
    setEditingPost(false)
    setPendingDeletePost(false)
  }, [post])

  function redirectToLogin() {
    window.location.assign('/login?redirect=/')
  }

  function handleApiError(error: unknown, fallbackMessage: string) {
    if (error instanceof ApiError && error.status === 401) {
      void auth.logout()
      redirectToLogin()
      return true
    }

    setEditError(error instanceof Error ? error.message : fallbackMessage)
    return false
  }

  function handlePostUpdated(updatedPost: PostDTO) {
    setPostState(updatedPost)
    onPostUpdated?.(updatedPost)
  }

  function startEditPost() {
    if (auth.status !== 'authenticated' || !auth.isAdmin || editPostMutation.isPending) {
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
    if (editPostMutation.isPending) {
      return
    }

    setEditDraft(postState.content)
    setEditHidden(!!postState.hide)
    setEditPinned(postState.pin != null)
    setEditPin(postState.pin != null ? String(postState.pin) : '')
    setEditError(null)
    setEditingPost(false)
  }

  function requestDeletePost() {
    if (auth.status !== 'authenticated' || !auth.isAdmin || deletePostMutation.isPending) {
      return
    }

    setPendingDeletePost(true)
  }

  async function handleDeletePost() {
    if (auth.status !== 'authenticated' || !auth.isAdmin || deletePostMutation.isPending) {
      return
    }

    try {
      await deletePostMutation.mutateAsync()
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: queryKeys.postPrefix(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.postsPrefix(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.latestCommentsPrefix(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.tagsPrefix(),
        }),
      ])
      onPostDeleted?.(postState.id)
    } catch (error) {
      handleApiError(error, '删除日志失败')
    } finally {
      setPendingDeletePost(false)
    }
  }

  async function handleEditPostSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || !auth.isAdmin || editPostMutation.isPending) {
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

    setEditError(null)

    try {
      const updatedPost = await editPostMutation.mutateAsync({
        clear_pin: clearPin || undefined,
        content: normalizeMarkdownContent(editDraft),
        hide: editHidden,
        pin,
      })

      handlePostUpdated(updatedPost)
      setEditDraft(updatedPost.content)
      setEditingPost(false)
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: queryKeys.postPrefix(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.postsPrefix(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.tagsPrefix(),
        }),
      ])
    } catch (error) {
      handleApiError(error, '更新日志失败')
    }
  }

  return (
    <>
      <AlertDialog
        open={pendingDeletePost}
        onOpenChange={(open) => {
          if (!open && !deletePostMutation.isPending) {
            setPendingDeletePost(false)
          }
        }}
      >
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogMedia className="bg-background">
              <Trash2Icon />
            </AlertDialogMedia>
            <AlertDialogTitle>删除这条日志？</AlertDialogTitle>
            <AlertDialogDescription>
              删除后这条日志会立即移除，相关评论也会一并消失。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deletePostMutation.isPending}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deletePostMutation.isPending}
              onClick={() => void handleDeletePost()}
              variant="destructive"
            >
              {deletePostMutation.isPending ? '正在删除…' : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card className="bg-card/50 shadow-sm">
        <CardHeader className="gap-4 border-b border-border/60">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 items-start gap-3">
              <Avatar size="lg">
                <AvatarImage
                  alt={postState.author.nickname}
                  src={postState.author.avatar}
                />
                <AvatarFallback>
                  {getInitials(postState.author.nickname)}
                </AvatarFallback>
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
                  <span>
                    发布于{' '}
                    {formatDateTime(postState.created_at, {
                      includeYear: true,
                    })}
                  </span>
                </div>
              </div>
            </div>
            {auth.isAdmin ? (
              <div className="flex shrink-0 items-center gap-1">
                <Button
                  aria-label="编辑日志"
                  disabled={editPostMutation.isPending}
                  onClick={startEditPost}
                  size="icon-sm"
                  type="button"
                  variant={editingPost ? 'secondary' : 'ghost'}
                >
                  <SquarePenIcon />
                </Button>
                <Button
                  aria-label="删除日志"
                  disabled={
                    deletePostMutation.isPending || editPostMutation.isPending
                  }
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
                      <FieldLabel
                        htmlFor={`post-edit-pin-enabled-${postState.id}`}
                      >
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
                    disabled={editPostMutation.isPending}
                    onClick={cancelEditPost}
                    type="button"
                    variant="outline"
                  >
                    取消
                  </Button>
                  <Button disabled={editPostMutation.isPending} type="submit">
                    {editPostMutation.isPending ? '正在保存…' : '保存修改'}
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

          {showComments ? (
            <>
              <Separator />
              <PostCommentsSection
                loginRedirectPath="/"
                onCommentTotalChange={setCommentTotal}
                onPostUpdated={handlePostUpdated}
                post={postState}
              />
            </>
          ) : null}
        </CardContent>

        <CardFooter className="justify-between gap-3">
          <div className="text-xs text-muted-foreground">
            {showComments
              ? `日志 #${postState.id} · 评论 ${commentTotal}`
              : `日志 #${postState.id}`}
          </div>
        </CardFooter>
      </Card>
    </>
  )
}
