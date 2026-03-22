import {
  ChevronDownIcon,
  Clock3Icon,
  MessageSquareTextIcon,
} from 'lucide-react'

import type { Comment, Post } from '#/lib/easydrop-api'
import { formatUserLabel } from '#/lib/easydrop-api'
import { formatAbsoluteTime, formatRelativeTime } from '#/lib/time'
import { cn } from '#/lib/utils'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
  CardDescription,
} from '#/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '#/components/ui/collapsible'
import { MarkdownContent } from './markdown-content'
import { CommentPanel } from './comment-panel'

type PostCardProps = {
  comments: Comment[]
  commentsError: string | null
  commentsLoading: boolean
  commentsOpen: boolean
  onRetryComments: () => void
  onToggleComments: (open: boolean) => void
  post: Post
}

export function PostCard({
  comments,
  commentsError,
  commentsLoading,
  commentsOpen,
  onRetryComments,
  onToggleComments,
  post,
}: PostCardProps) {
  const createdAt = formatRelativeTime(post.created_at)
  const absoluteCreatedAt = formatAbsoluteTime(post.created_at)
  const wasEdited = post.updated_at && post.updated_at !== post.created_at

  return (
    <Collapsible
      className="w-full"
      onOpenChange={onToggleComments}
      open={commentsOpen}
    >
      <Card className="overflow-hidden border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.07)] backdrop-blur-xl">
        <CardHeader className="gap-3 border-b border-border/60">
          <div className="flex items-start gap-3">
            <Avatar size="lg">
              <AvatarFallback>{String(post.user_id).slice(-2)}</AvatarFallback>
            </Avatar>
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <CardTitle>{formatUserLabel(post.user_id)}</CardTitle>
              <CardDescription className="flex flex-wrap items-center gap-2">
                <span>{createdAt}</span>
                <span aria-hidden="true">·</span>
                <time dateTime={post.created_at} title={absoluteCreatedAt}>
                  {absoluteCreatedAt}
                </time>
                {wasEdited ? (
                  <>
                    <span aria-hidden="true">·</span>
                    <span title={formatAbsoluteTime(post.updated_at)}>
                      已编辑
                    </span>
                  </>
                ) : null}
              </CardDescription>
            </div>
          </div>
          <CardAction className="hidden sm:block">
            <Badge variant={post.hide ? 'destructive' : 'secondary'}>
              {post.hide ? '隐藏' : '公开'}
            </Badge>
          </CardAction>
        </CardHeader>

        <CardContent className="flex flex-col gap-4 pt-4">
          <MarkdownContent content={post.content} />

          {(post.tags?.length ?? 0) > 0 ? (
            <div className="flex flex-wrap gap-2">
              {post.tags?.map((tag) => (
                <Badge key={tag.id} variant="outline">
                  #{tag.name}
                </Badge>
              ))}
            </div>
          ) : null}
        </CardContent>

        <CardFooter className="flex flex-col gap-3 border-t border-border/60 bg-muted/35">
          <div className="flex w-full flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Clock3Icon className="size-3.5" />
              <span>仅渲染 Markdown，已禁用脚本执行</span>
            </div>

            <CollapsibleTrigger asChild>
              <Button size="sm" variant="outline">
                <MessageSquareTextIcon data-icon="inline-start" />
                {commentsOpen ? '收起评论' : '查看评论'}
                <ChevronDownIcon
                  className={cn(
                    'transition-transform',
                    commentsOpen ? 'rotate-180' : 'rotate-0',
                  )}
                  data-icon="inline-end"
                />
              </Button>
            </CollapsibleTrigger>
          </div>

          <CollapsibleContent className="w-full">
            <CommentPanel
              comments={comments}
              error={commentsError}
              isLoading={commentsLoading}
              onRetry={onRetryComments}
              postAuthorId={post.user_id}
            />
          </CollapsibleContent>
        </CardFooter>
      </Card>
    </Collapsible>
  )
}
