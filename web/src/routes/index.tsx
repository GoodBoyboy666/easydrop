import { createFileRoute } from '@tanstack/react-router'
import {
  AlertCircleIcon,
  CornerRightUpIcon,
  HashIcon,
  LogsIcon,
  MegaphoneIcon,
  MessageSquareTextIcon,
  RefreshCwIcon,
} from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatRelativeTime, getInitials } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { useSiteSettings } from '#/lib/site-settings'
import type { CommentDTO, PagedResult, PostDTO, TagDTO } from '#/lib/types'
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { MarkdownEditor } from '#/components/markdown/markdown-editor'
import { PostCard } from '#/components/home/post-card'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
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
import {
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Skeleton } from '#/components/ui/skeleton'
import { Switch } from '#/components/ui/switch'

export const Route = createFileRoute('/')({
  component: HomePage,
})

const FEED_PAGE_SIZE = 8
const LATEST_COMMENTS_PAGE_SIZE = 6
const LATEST_COMMENTS_FETCH_SIZE = 24

interface FeedState {
  items: PostDTO[]
  pinnedItems: PostDTO[]
  loading: boolean
  loadingMore: boolean
  error: string | null
  total: number
}

function mergePostsById(current: PostDTO[], incoming: PostDTO[]) {
  const seen = new Set<number>()
  const merged: PostDTO[] = []

  for (const post of [...current, ...incoming]) {
    if (seen.has(post.id)) {
      continue
    }
    seen.add(post.id)
    merged.push(post)
  }

  return merged
}

function isTopLevelComment(comment: CommentDTO) {
  return comment.root_id == null && comment.parent_id == null
}

function HomePage() {
  const auth = useAuth()
  const prefersReducedMotion = useReducedMotion()
  const {
    error: settingsError,
    loading: settingsLoading,
    siteAnnouncement,
    siteOwner,
    siteOwnerDescription,
  } = useSiteSettings()
  const [feedState, setFeedState] = useState<FeedState>({
    items: [],
    pinnedItems: [],
    loading: true,
    loadingMore: false,
    error: null,
    total: 0,
  })
  const [latestComments, setLatestComments] = useState<PagedResult<CommentDTO>>({
    items: [],
    total: 0,
  })
  const [latestCommentsLoading, setLatestCommentsLoading] = useState(true)
  const [latestCommentsError, setLatestCommentsError] = useState<string | null>(null)
  const [tags, setTags] = useState<PagedResult<TagDTO>>({
    items: [],
    total: 0,
  })
  const [tagsLoading, setTagsLoading] = useState(true)
  const [tagsError, setTagsError] = useState<string | null>(null)
  const [publishDraft, setPublishDraft] = useState('')
  const [publishHidden, setPublishHidden] = useState(false)
  const [publishPinned, setPublishPinned] = useState(false)
  const [publishPin, setPublishPin] = useState('')
  const [publishError, setPublishError] = useState<string | null>(null)
  const [publishing, setPublishing] = useState(false)

  async function loadFeed(mode: 'initial' | 'more' = 'initial') {
    const offset =
      mode === 'more'
        ? feedState.items.length + feedState.pinnedItems.length
        : 0

    setFeedState((current) => ({
      ...current,
      error: null,
      loading: mode === 'initial',
      loadingMore: mode === 'more',
    }))

    try {
      const result = await api.getPosts({
        limit: FEED_PAGE_SIZE,
        offset,
        order: 'created_at_desc',
      }, auth.token)

      setFeedState((current) => ({
        items:
          mode === 'more'
            ? mergePostsById(current.items, result.items)
            : result.items,
        pinnedItems:
          mode === 'more'
            ? mergePostsById(current.pinnedItems, result.pinnedItems)
            : result.pinnedItems,
        loading: false,
        loadingMore: false,
        error: null,
        total: result.total,
      }))
    } catch (error) {
      setFeedState((current) => ({
        ...current,
        loading: false,
        loadingMore: false,
        error: error instanceof Error ? error.message : '加载日志流失败',
      }))
    }
  }

  useEffect(() => {
    void loadFeed()

    // 登录态切换后重新拉取日志，确保已登录用户能看到私密内容。
  }, [auth.token])

  useEffect(() => {

    void (async () => {
      try {
        const result = await api.getLatestComments({
          limit: LATEST_COMMENTS_FETCH_SIZE,
          offset: 0,
          order: 'created_at_desc',
        })
        setLatestComments({
          items: result.items
            .filter(isTopLevelComment)
            .slice(0, LATEST_COMMENTS_PAGE_SIZE),
          total: result.total,
        })
        setLatestCommentsError(null)
      } catch (error) {
        setLatestComments({
          items: [],
          total: 0,
        })
        setLatestCommentsError(
          error instanceof Error ? error.message : '加载最新评论失败'
        )
      } finally {
        setLatestCommentsLoading(false)
      }
    })()

    void (async () => {
      try {
        const result = await api.getTags({
          limit: 16,
          offset: 0,
          order: 'hot_desc',
        })
        setTags(result)
        setTagsError(null)
      } catch {
        try {
          const fallback = await api.getTags({
            limit: 16,
            offset: 0,
            order: 'created_at_desc',
          })
          setTags(fallback)
          setTagsError(null)
        } catch (error) {
          setTags({
            items: [],
            total: 0,
          })
          setTagsError(error instanceof Error ? error.message : '加载标签失败')
        }
      } finally {
        setTagsLoading(false)
      }
    })()
  }, [])

  const loadedPostCount = feedState.items.length + feedState.pinnedItems.length
  const canLoadMorePosts = loadedPostCount < feedState.total
  const normalizedAnnouncement = siteAnnouncement.trim() || '暂无公告'
  const siteStats = useMemo(
    () => [
      { label: '日志', value: feedState.total },
      { label: '评论', value: latestComments.total },
      { label: '标签', value: tags.total },
    ],
    [feedState.total, latestComments.total, tags.total]
  )
  const getEntranceTransition = (delay = 0) =>
    prefersReducedMotion
      ? { duration: 0 }
      : { duration: 0.32, ease: 'easeOut' as const, delay }

  async function handlePublish(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!hasMarkdownContent(publishDraft)) {
      setPublishError('发布内容不能为空')
      return
    }

    if (!auth.token || !auth.isAdmin) {
      setPublishError('只有管理员可以发布日志')
      return
    }

    const normalizedPin = publishPin.trim()
    let pin: number | undefined

    if (publishPinned) {
      if (!normalizedPin) {
        setPublishError('启用置顶后请填写 Pin 值')
        return
      }

      const parsedPin = Number(normalizedPin)

      if (!Number.isInteger(parsedPin) || parsedPin <= 0) {
        setPublishError('Pin 必须是大于 0 的整数')
        return
      }

      pin = parsedPin
    }

    setPublishing(true)
    setPublishError(null)

    try {
      await api.createAdminPost(
        {
          content: normalizeMarkdownContent(publishDraft),
          hide: publishHidden,
          pin,
        },
        auth.token
      )
      setPublishDraft('')
      setPublishHidden(false)
      setPublishPinned(false)
      setPublishPin('')
      await loadFeed()
    } catch (error) {
      setPublishError(error instanceof Error ? error.message : '发布失败')
    } finally {
      setPublishing(false)
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <motion.section
        animate={{ opacity: 1, y: 0 }}
        className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]"
        initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
        transition={getEntranceTransition()}
      >
        <div className="flex min-w-0 flex-col gap-6">
          {auth.status === 'authenticated' && auth.isAdmin ? (
            <motion.div
              animate={{ opacity: 1, y: 0 }}
              initial={prefersReducedMotion ? false : { opacity: 0, y: 12 }}
              transition={getEntranceTransition(0.06)}
            >
              <Card className="bg-primary/5 shadow-sm">
                <CardHeader>
                  <CardTitle>快捷发布</CardTitle>
                </CardHeader>
                <CardContent>
                  <form onSubmit={handlePublish}>
                  <FieldGroup>
                    <Field data-invalid={!!publishError}>
                      <MarkdownEditor
                        height={150}
                        onChange={setPublishDraft}
                        placeholder="快写下你的想法吧，支持 Markdown。"
                        value={publishDraft}
                      />
                      <FieldError>{publishError}</FieldError>
                    </Field>

                    <Field orientation="horizontal">
                      <Switch
                        checked={publishPinned}
                        id="quick-publish-pin-enabled"
                        onCheckedChange={(checked) => {
                          setPublishPinned(checked)
                          if (!checked) {
                            setPublishPin('')
                          }
                        }}
                        size="sm"
                      />
                      <FieldContent>
                        <FieldLabel htmlFor="quick-publish-pin-enabled">
                          <FieldTitle>置顶</FieldTitle>
                        </FieldLabel>
                        {publishPinned ? (
                          <div className="mt-2">
                            <Input
                              className="w-28"
                              id="quick-publish-pin"
                              inputMode="numeric"
                              min={1}
                              onChange={(event) => setPublishPin(event.target.value)}
                              placeholder="权重"
                              step={1}
                              type="number"
                              value={publishPin}
                            />
                          </div>
                        ) : null}
                      </FieldContent>
                    </Field>

                    <Field orientation="horizontal">
                      <Switch
                        checked={publishHidden}
                        id="quick-publish-hidden"
                        onCheckedChange={setPublishHidden}
                        size="sm"
                      />
                      <FieldContent>
                        <FieldLabel htmlFor="quick-publish-hidden">
                          <FieldTitle>私密发布</FieldTitle>
                        </FieldLabel>
                      </FieldContent>
                    </Field>
                  </FieldGroup>

                  <div className="mt-3 flex justify-end">
                    <Button disabled={publishing} type="submit">
                      <CornerRightUpIcon data-icon="inline-start" />
                      {publishing ? '正在发布…' : '立即发布'}
                    </Button>
                  </div>
                  </form>
                </CardContent>
              </Card>
            </motion.div>
          ) : null}

          <section className="flex flex-col gap-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="font-heading flex items-center gap-2 text-lg font-semibold">
                  <LogsIcon className="size-4 text-muted-foreground" />
                  日志
                </h2>
              </div>
              <motion.div
                whileHover={prefersReducedMotion ? undefined : { scale: 1.02 }}
                whileTap={prefersReducedMotion ? undefined : { scale: 0.98 }}
              >
                <Button
                  onClick={() => void loadFeed()}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  <RefreshCwIcon data-icon="inline-start" />
                  刷新
                </Button>
              </motion.div>
            </div>

            {feedState.loading ? (
              <div className="flex flex-col gap-4">
                {Array.from({ length: 3 }).map((_, index) => (
                  <Card key={index} className="border border-border/70 bg-card/90">
                    <CardHeader>
                      <div className="flex items-center gap-3">
                        <Skeleton className="size-10 rounded-full" />
                        <div className="flex flex-1 flex-col gap-2">
                          <Skeleton className="h-4 w-36" />
                          <Skeleton className="h-4 w-28" />
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-3 pt-2">
                      <Skeleton className="h-4 w-full" />
                      <Skeleton className="h-4 w-10/12" />
                      <Skeleton className="h-4 w-8/12" />
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : null}

            {!feedState.loading && feedState.error ? (
              <Alert variant="destructive">
                <AlertCircleIcon />
                <AlertTitle>日志流加载失败</AlertTitle>
                <AlertDescription>{feedState.error}</AlertDescription>
              </Alert>
            ) : null}

            {!feedState.loading &&
            !feedState.error &&
            feedState.pinnedItems.length === 0 &&
            feedState.items.length === 0 ? (
              <Empty className="border border-dashed border-border/80 bg-card/60">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <AlertCircleIcon />
                  </EmptyMedia>
                  <EmptyTitle>还没有任何日志</EmptyTitle>
                  <EmptyDescription>
                    发布第一条日志后，这里会成为站点的主时间线。
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : null}

            {!feedState.loading && !feedState.error ? (
              <div className="flex flex-col gap-4">
                <AnimatePresence initial={false}>
                  {[...feedState.pinnedItems, ...feedState.items].map((post, index) => (
                    <motion.div
                      key={post.id}
                      animate={{ opacity: 1, scale: 1, y: 0 }}
                      exit={
                        prefersReducedMotion
                          ? undefined
                          : { opacity: 0, scale: 0.98, y: -8 }
                      }
                      initial={
                        prefersReducedMotion
                          ? false
                          : { opacity: 0, scale: 0.98, y: 12 }
                      }
                      layout
                      transition={getEntranceTransition(Math.min(index * 0.03, 0.18))}
                    >
                      <PostCard
                        onPostDeleted={(postId) => {
                          setFeedState((current) => ({
                            ...current,
                            pinnedItems: current.pinnedItems.filter(
                              (item) => item.id !== postId
                            ),
                            items: current.items.filter((item) => item.id !== postId),
                            total: Math.max(0, current.total - 1),
                          }))
                          setLatestComments((current) => ({
                            ...current,
                            items: current.items.filter((comment) => comment.post_id !== postId),
                          }))
                        }}
                        post={post}
                      />
                    </motion.div>
                  ))}
                </AnimatePresence>
              </div>
            ) : null}

            {canLoadMorePosts ? (
              <motion.div
                whileHover={prefersReducedMotion ? undefined : { y: -1 }}
                whileTap={prefersReducedMotion ? undefined : { scale: 0.99 }}
              >
                <Button
                  disabled={feedState.loadingMore}
                  onClick={() => void loadFeed('more')}
                  type="button"
                  variant="outline"
                >
                  {feedState.loadingMore ? '正在加载…' : '加载更多日志'}
                </Button>
              </motion.div>
            ) : null}
          </section>
        </div>

        <motion.aside
          animate={{ opacity: 1, x: 0 }}
          className="flex min-w-0 flex-col gap-4"
          initial={prefersReducedMotion ? false : { opacity: 0, x: 10 }}
          transition={getEntranceTransition(0.12)}
        >
          <Card className="bg-card/90 shadow-sm">
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
                    <div className="text-xs text-muted-foreground">{item.label}</div>
                  </div>
                ))}
              </div>

              {settingsLoading ? (
                <div className="text-xs text-muted-foreground">正在同步站点配置…</div>
              ) : null}
            </CardContent>
          </Card>

          <Card className="bg-card/90 shadow-sm">
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

          <Card className="bg-card/90 shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MessageSquareTextIcon className="size-4 text-muted-foreground" />
                最新评论
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {latestCommentsLoading
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

              {!latestCommentsLoading &&
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

              {!latestCommentsLoading && !latestCommentsError ? (
                <AnimatePresence initial={false}>
                  {latestComments.items.map((comment, index) => (
                    <motion.article
                      key={comment.id}
                      animate={{ opacity: 1, x: 0 }}
                      className={`px-2.5 pt-2.5 pb-1.5 ${
                        index > 0 ? 'border-t border-dashed border-border/60' : ''
                      }`}
                      exit={prefersReducedMotion ? undefined : { opacity: 0, x: -8 }}
                      initial={prefersReducedMotion ? false : { opacity: 0, x: 8 }}
                      transition={getEntranceTransition(Math.min(index * 0.03, 0.15))}
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
                    </motion.article>
                  ))}
                </AnimatePresence>
              ) : null}
            </CardContent>
          </Card>

          <Card className="bg-card/90 shadow-sm">
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
                {tagsLoading
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

              {!tagsLoading && !tagsError && tags.items.length === 0 ? (
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
      </motion.section>
    </div>
  )
}
