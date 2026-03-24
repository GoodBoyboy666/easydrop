"use client"

import { useEffect, useState } from 'react'
import {
  ChevronDownIcon,
  MessageSquareMoreIcon,
  SendHorizontalIcon,
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
  FieldDescription,
  FieldError,
  FieldGroup,
} from '#/components/ui/field'
import { Separator } from '#/components/ui/separator'
import { Skeleton } from '#/components/ui/skeleton'

interface PostCardProps {
  onPostRefreshed?: () => Promise<void> | void
  post: PostDTO
}

interface CommentState {
  items: CommentDTO[]
  loading: boolean
  loadingMore: boolean
  total: number
}

const INITIAL_COMMENT_PAGE_SIZE = 3
const LOAD_MORE_COMMENT_PAGE_SIZE = 5

export function PostCard({ onPostRefreshed, post }: PostCardProps) {
  const auth = useAuth()
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

  useEffect(() => {
    let cancelled = false

    async function loadComments() {
      setCommentError(null)
      setCommentState((current) => ({ ...current, loading: true }))

      try {
        const result = await api.getPostComments(post.id, {
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
  }, [post.id])

  useEffect(() => {
    setReplyTarget(null)
    setCommentDraft('')
    setSubmitError(null)
    setCommentSectionOpen(false)
    setComposerOpen(false)
  }, [post.id])

  async function loadMoreComments() {
    setCommentError(null)
    setCommentState((current) => ({ ...current, loadingMore: true }))

    try {
      const result = await api.getPostComments(post.id, {
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

  function startReply(comment: CommentDTO) {
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

  async function handleCommentSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

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
        post.id,
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

      const result = await api.getPostComments(post.id, {
        limit: Math.max(
          LOAD_MORE_COMMENT_PAGE_SIZE,
          commentState.items.length || INITIAL_COMMENT_PAGE_SIZE
        ),
        offset: 0,
        order: 'created_at_desc',
      })

      setCommentState({
        items: result.items,
        loading: false,
        loadingMore: false,
        total: result.total,
      })

      await onPostRefreshed?.()
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
    : '说点什么，支持 Markdown 评论。'
  const commentDescription = replyTarget
    ? `当前将回复 ${replyTarget.author.nickname}，发布后会以扁平评论形式展示。`
    : '支持直接评论和回复评论，原始 HTML 默认禁用。'

  return (
    <Card className="border border-border/70 bg-card/90 shadow-sm">
      <CardHeader className="gap-4 border-b border-border/60">
        <div className="flex items-start gap-3">
          <Avatar size="lg">
            <AvatarImage alt={post.author.nickname} src={post.author.avatar} />
            <AvatarFallback>{getInitials(post.author.nickname)}</AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <CardTitle className="truncate">{post.author.nickname}</CardTitle>
            <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span>{formatRelativeTime(post.created_at)}</span>
              <span>发布于 {formatDateTime(post.created_at)}</span>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-3">
          <MarkdownContent content={post.content} />
          {post.tags?.length ? (
            <div className="flex flex-wrap gap-2">
              {post.tags.map((tag) => (
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

          <div className="flex shrink-0 items-center">
            <Button
              size="sm"
              variant={composerOpen ? 'secondary' : 'outline'}
              type="button"
              onClick={() => {
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
              }}
            >
              <MessageSquareMoreIcon data-icon="inline-start" />
              发评论
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
                    <FieldDescription>{commentDescription}</FieldDescription>
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
          日志 #{post.id} · 评论 {commentState.total}
        </div>
      </CardFooter>
    </Card>
  )
}
