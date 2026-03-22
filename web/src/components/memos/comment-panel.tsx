import {
  MessageSquareTextIcon,
  RefreshCcwIcon,
  TriangleAlertIcon,
} from 'lucide-react'

import type { Comment } from '#/lib/easydrop-api'
import { formatUserLabel } from '#/lib/easydrop-api'
import { formatAbsoluteTime, formatRelativeTime } from '#/lib/time'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '#/components/ui/empty'
import { Separator } from '#/components/ui/separator'
import { Skeleton } from '#/components/ui/skeleton'
import { MarkdownContent } from './markdown-content'

type CommentPanelProps = {
  comments: Comment[]
  error: string | null
  isLoading: boolean
  onRetry: () => void
  postAuthorId: number
}

function CommentSkeleton() {
  return (
    <div className="flex flex-col gap-3 rounded-xl border bg-background/70 p-4">
      <div className="flex items-center gap-3">
        <Skeleton className="size-8 rounded-full" />
        <div className="flex flex-col gap-2">
          <Skeleton className="h-3 w-20" />
          <Skeleton className="h-3 w-28" />
        </div>
      </div>
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-4/5" />
    </div>
  )
}

export function CommentPanel({
  comments,
  error,
  isLoading,
  onRetry,
  postAuthorId,
}: CommentPanelProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-3 px-4 pb-4">
        <CommentSkeleton />
        <CommentSkeleton />
      </div>
    )
  }

  if (error) {
    return (
      <div className="px-4 pb-4">
        <Alert variant="destructive">
          <TriangleAlertIcon />
          <AlertTitle>评论加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
          <Button
            className="mt-2 w-fit"
            onClick={onRetry}
            size="sm"
            variant="outline"
          >
            <RefreshCcwIcon data-icon="inline-start" />
            重试
          </Button>
        </Alert>
      </div>
    )
  }

  if (comments.length === 0) {
    return (
      <div className="px-4 pb-4">
        <Empty className="rounded-2xl border bg-background/60">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <MessageSquareTextIcon />
            </EmptyMedia>
            <EmptyTitle>还没有评论</EmptyTitle>
            <EmptyDescription>
              这条说说暂时没有收到回复，时间线保持清爽。
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3 px-4 pb-4">
      {comments.map((comment, index) => {
        const isAuthor = comment.user_id === postAuthorId

        return (
          <div className="flex flex-col gap-3" key={comment.id}>
            {index > 0 ? <Separator /> : null}
            <article className="flex flex-col gap-3 rounded-2xl border bg-background/70 p-4">
              <header className="flex items-start gap-3">
                <Avatar size="sm">
                  <AvatarFallback>
                    {String(comment.user_id).slice(-2)}
                  </AvatarFallback>
                </Avatar>
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium text-foreground">
                      {formatUserLabel(comment.user_id)}
                    </span>
                    {isAuthor ? <Badge variant="secondary">作者</Badge> : null}
                    {comment.reply_to_user_id ? (
                      <Badge variant="outline">
                        回复 {formatUserLabel(comment.reply_to_user_id)}
                      </Badge>
                    ) : null}
                  </div>
                  <time
                    className="text-xs text-muted-foreground"
                    dateTime={comment.created_at}
                    title={formatAbsoluteTime(comment.created_at)}
                  >
                    {formatRelativeTime(comment.created_at)}
                  </time>
                </div>
              </header>
              <MarkdownContent content={comment.content} />
            </article>
          </div>
        )
      })}
    </div>
  )
}
