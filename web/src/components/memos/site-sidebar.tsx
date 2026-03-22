import {
  ArrowUpRightIcon,
  MessageSquareQuoteIcon,
  ShapesIcon,
  TriangleAlertIcon,
} from 'lucide-react'

import type { Comment, PublicSettingsRecord, Tag } from '#/lib/easydrop-api'
import { formatAbsoluteTime, formatRelativeTime } from '#/lib/time'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Badge } from '#/components/ui/badge'
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
import { Skeleton } from '#/components/ui/skeleton'

type ModuleStatus = 'loading' | 'success' | 'error'

type SiteSidebarProps = {
  latestRootComments: Comment[]
  latestRootCommentsError: string | null
  latestRootCommentsStatus: ModuleStatus
  settings: PublicSettingsRecord
  settingsError: string | null
  tags: Tag[]
  tagsError: string | null
  tagsStatus: ModuleStatus
}

function toCommentSnippet(content: string) {
  return (
    content.replace(/\s+/g, ' ').trim().slice(0, 110) ||
    '这条评论没有可展示的文本摘要。'
  )
}

function ModuleSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-4 w-28" />
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-4/5" />
      <Skeleton className="h-3 w-3/5" />
    </div>
  )
}

function ModuleError({
  description,
  title,
}: {
  description: string
  title: string
}) {
  return (
    <Alert variant="destructive">
      <TriangleAlertIcon />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
  )
}

export function SiteSidebar({
  latestRootComments,
  latestRootCommentsError,
  latestRootCommentsStatus,
  settings,
  settingsError,
  tags,
  tagsError,
  tagsStatus,
}: SiteSidebarProps) {
  const siteName = settings.site_name?.trim() || 'EasyDrop'
  const siteAnnouncement = settings.site_announcement?.trim()
  const siteUrl = settings.site_url?.trim()

  return (
    <div className="flex flex-col gap-4 lg:sticky lg:top-24">
      <Card className="border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.06)] backdrop-blur-xl">
        <CardHeader className="gap-3 border-b border-border/60">
          <Badge className="w-fit" variant="secondary">
            站点公告
          </Badge>
          <CardTitle className="text-xl">{siteName}</CardTitle>
          {siteAnnouncement ? (
            <CardDescription>{siteAnnouncement}</CardDescription>
          ) : null}
        </CardHeader>
        <CardContent className="flex flex-col gap-4 pt-4">
          {settingsError ? (
            <ModuleError description={settingsError} title="公告配置加载失败" />
          ) : null}

          {siteUrl ? (
            <Button
              asChild
              className="w-full justify-between"
              variant="outline"
            >
              <a href={siteUrl} rel="noreferrer noopener" target="_blank">
                访问站点链接
                <ArrowUpRightIcon data-icon="inline-end" />
              </a>
            </Button>
          ) : null}
        </CardContent>
      </Card>

      <Card className="border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.05)]">
        <CardHeader className="gap-2 border-b border-border/60">
          <div className="flex items-center gap-2">
            <MessageSquareQuoteIcon className="size-4 text-foreground" />
            <CardTitle className="text-base">最新评论</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="pt-4">
          {latestRootCommentsStatus === 'loading' ? <ModuleSkeleton /> : null}

          {latestRootCommentsStatus === 'error' && latestRootCommentsError ? (
            <ModuleError
              description={latestRootCommentsError}
              title="最新评论加载失败"
            />
          ) : null}

          {latestRootCommentsStatus === 'success' &&
          latestRootComments.length === 0 ? (
            <Empty className="rounded-2xl border bg-background/60">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquareQuoteIcon />
                </EmptyMedia>
                <EmptyTitle>暂时没有根评论</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : null}

          {latestRootCommentsStatus === 'success' &&
          latestRootComments.length > 0 ? (
            <div className="flex flex-col gap-3">
              {latestRootComments.map((comment) => (
                <article
                  className="rounded-2xl border bg-background/65 p-4"
                  key={comment.id}
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="outline">评论 #{comment.id}</Badge>
                    <Badge variant="secondary">说说 #{comment.post_id}</Badge>
                  </div>
                  <p className="mt-3 line-clamp-3 text-sm leading-6 text-foreground">
                    {toCommentSnippet(comment.content)}
                  </p>
                  <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <span>{formatRelativeTime(comment.created_at)}</span>
                    <span aria-hidden="true">·</span>
                    <span title={formatAbsoluteTime(comment.created_at)}>
                      {formatAbsoluteTime(comment.created_at)}
                    </span>
                  </div>
                </article>
              ))}
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card className="border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.05)]">
        <CardHeader className="gap-2 border-b border-border/60">
          <div className="flex items-center gap-2">
            <ShapesIcon className="size-4 text-foreground" />
            <CardTitle className="text-base">标签</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="pt-4">
          {tagsStatus === 'loading' ? <ModuleSkeleton /> : null}

          {tagsStatus === 'error' && tagsError ? (
            <ModuleError description={tagsError} title="标签加载失败" />
          ) : null}

          {tagsStatus === 'success' && tags.length === 0 ? (
            <Empty className="rounded-2xl border bg-background/60">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <ShapesIcon />
                </EmptyMedia>
                <EmptyTitle>还没有标签</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : null}

          {tagsStatus === 'success' && tags.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => (
                <Badge key={tag.id} variant="outline">
                  #{tag.name}
                </Badge>
              ))}
            </div>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}
