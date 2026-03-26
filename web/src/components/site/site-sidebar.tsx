'use client'

import { motion, useReducedMotion } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { HashIcon, MegaphoneIcon, MessageSquareTextIcon } from 'lucide-react'
import { useAuth } from '#/lib/auth'
import { formatRelativeTime, getInitials } from '#/lib/format'
import {
  latestCommentsQueryOptions,
  postsQueryOptions,
  tagsQueryOptions,
} from '#/lib/query-options'
import { useSiteSettings } from '#/lib/site-settings'
import type { CommentDTO } from '#/lib/types'
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '#/components/ui/empty'
import { Skeleton } from '#/components/ui/skeleton'

const LATEST_COMMENTS_PAGE_SIZE = 6
const LATEST_COMMENTS_FETCH_SIZE = 24
const SIDEBAR_MOTION_DELAY_SECONDS = 0.12

function isTopLevelComment(comment: CommentDTO) {
  return comment.root_id == null && comment.parent_id == null
}

export function SiteSidebar() {
  const auth = useAuth()
  const prefersReducedMotion = useReducedMotion()
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)
  const {
    error: settingsError,
    loading: settingsLoading,
    siteAnnouncement,
    siteOwner,
    siteOwnerDescription,
  } = useSiteSettings()
  const postsQuery = useQuery({
    ...postsQueryOptions(auth.status === 'authenticated', {
      limit: 1,
      offset: 0,
      order: 'created_at_desc',
    }),
  })
  const latestCommentsQuery = useQuery({
    ...latestCommentsQueryOptions({
      limit: LATEST_COMMENTS_FETCH_SIZE,
      offset: 0,
      order: 'created_at_desc',
    }),
    select: (result) => ({
      ...result,
      items: result.items
        .filter(isTopLevelComment)
        .slice(0, LATEST_COMMENTS_PAGE_SIZE),
    }),
  })
  const tagsQuery = useQuery(
    tagsQueryOptions({
      limit: 16,
      offset: 0,
      order: 'hot_desc',
    }),
  )

  const latestComments = latestCommentsQuery.data ?? {
    items: [],
    total: 0,
  }
  const tags = tagsQuery.data ?? {
    items: [],
    total: 0,
  }
  const postsTotal = postsQuery.data?.total ?? 0
  const latestCommentsError =
    latestCommentsQuery.error instanceof Error
      ? latestCommentsQuery.error.message
      : null
  const tagsError =
    tagsQuery.error instanceof Error ? tagsQuery.error.message : null
  const normalizedAnnouncement = siteAnnouncement.trim() || '暂无公告'
  const siteStats = useMemo(
    () => [
      { label: '日志', value: postsTotal },
      { label: '评论', value: latestComments.total },
      { label: '标签', value: tags.total },
    ],
    [latestComments.total, postsTotal, tags.total],
  )

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

  return (
    <motion.aside
      animate={motionReady ? { opacity: 1, x: 0 } : { opacity: 0, x: 10 }}
      className="flex min-w-0 flex-col gap-4"
      initial={false}
      transition={
        prefersReducedMotion
          ? { duration: 0 }
          : {
              type: 'spring',
              duration: 0.32,
              ease: 'easeOut',
              delay: SIDEBAR_MOTION_DELAY_SECONDS,
            }
      }
    >
      <Card className="bg-card/50 shadow-sm">
        <CardHeader>
          <CardTitle>{siteOwner}</CardTitle>
          <CardDescription>{siteOwnerDescription}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {settingsError ? (
            <Alert variant="destructive">
              <AlertTitle>站点信息读取失败</AlertTitle>
              <AlertDescription>{settingsError}</AlertDescription>
            </Alert>
          ) : null}

          <div className="grid grid-cols-3 gap-2">
            {siteStats.map((item) => (
              <div
                key={item.label}
                className="rounded-xl px-3 py-2 text-center"
              >
                <div className="text-lg font-semibold">{item.value}</div>
                <div className="text-xs text-muted-foreground">
                  {item.label}
                </div>
              </div>
            ))}
          </div>

          {settingsLoading ? (
            <div className="text-xs text-muted-foreground">
              正在同步站点配置…
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card className="bg-card/50 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MegaphoneIcon className="size-4 text-muted-foreground" />
            网站公告
          </CardTitle>
        </CardHeader>
        <CardContent className="text-sm leading-7 text-muted-foreground">
          {normalizedAnnouncement}
        </CardContent>
      </Card>

      <Card className="bg-card/50 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MessageSquareTextIcon className="size-4 text-muted-foreground" />
            最新评论
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {latestCommentsQuery.isPending
            ? Array.from({ length: 3 }).map((_, index) => (
                <div
                  key={index}
                  className={`px-2.5 pt-2.5 pb-1.5 ${
                    index > 0 ? 'border-t border-dashed border-border/60' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <Skeleton className="size-6 rounded-full" />
                    <div className="flex flex-1 flex-col gap-2">
                      <Skeleton className="h-3.5 w-20" />
                      <Skeleton className="h-3.5 w-full" />
                    </div>
                  </div>
                </div>
              ))
            : null}

          {latestCommentsError ? (
            <Alert variant="destructive">
              <AlertTitle>最新评论读取失败</AlertTitle>
              <AlertDescription>{latestCommentsError}</AlertDescription>
            </Alert>
          ) : null}

          {!latestCommentsQuery.isPending &&
          !latestCommentsError &&
          latestComments.items.length === 0 ? (
            <Empty className="border border-dashed border-border/80 bg-muted/20">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquareTextIcon />
                </EmptyMedia>
                <EmptyTitle>暂无最新评论</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : null}

          {!latestCommentsQuery.isPending && !latestCommentsError
            ? latestComments.items.map((comment, index) => (
                <article
                  key={comment.id}
                  className={`px-2.5 pt-2.5 pb-1.5 ${
                    index > 0 ? 'border-t border-dashed border-border/60' : ''
                  }`}
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
                        <Link
                          className="font-medium leading-none hover:underline"
                          params={{ id: String(comment.post_id) }}
                          to="/posts/$id"
                        >
                          {comment.author.nickname}
                        </Link>
                        <span className="text-[0.7rem] text-muted-foreground">
                          {formatRelativeTime(comment.created_at)}
                        </span>
                      </div>
                      <div className="mt-1 text-foreground/85">
                        <MarkdownContent compact content={comment.content} />
                      </div>
                      <div className="mt-1.5 text-[0.7rem] text-muted-foreground">
                        来自日志 #{comment.post_id}
                      </div>
                    </div>
                  </div>
                </article>
              ))
            : null}
        </CardContent>
      </Card>

      <Card className="bg-card/50 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <HashIcon className="size-4 text-muted-foreground" />
            全站 Tag
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {tagsError ? (
            <Alert variant="destructive">
              <AlertTitle>标签读取失败</AlertTitle>
              <AlertDescription>{tagsError}</AlertDescription>
            </Alert>
          ) : null}

          <div className="flex flex-wrap gap-2">
            {tagsQuery.isPending
              ? Array.from({ length: 8 }).map((_, index) => (
                  <Skeleton key={index} className="h-7 w-20 rounded-full" />
                ))
              : tags.items.map((tag) => (
                  <Badge key={tag.id} variant="secondary">
                    <HashIcon data-icon="inline-start" />
                    {tag.name}
                  </Badge>
                ))}
          </div>

          {!tagsQuery.isPending && !tagsError && tags.items.length === 0 ? (
            <Empty className="border border-dashed border-border/80 bg-muted/20">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <HashIcon />
                </EmptyMedia>
                <EmptyTitle>暂无标签</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : null}
        </CardContent>
      </Card>
    </motion.aside>
  )
}
