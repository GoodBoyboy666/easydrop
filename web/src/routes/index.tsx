import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import {
  AlertCircleIcon,
  CornerRightUpIcon,
  LogsIcon,
  RefreshCwIcon,
} from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import type { HTMLMotionProps, Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { postsQueryOptions } from '#/lib/query-options'
import { invalidatePublicFeedQueries } from '#/lib/query-invalidation'
import { MarkdownEditor } from '#/components/markdown/markdown-editor'
import { PostCard } from '#/components/home/post-card'
import { SiteSidebar } from '#/components/site/site-sidebar'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
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
  validateSearch: (search: Record<string, unknown>) => {
    const content =
      typeof search.content === 'string' ? search.content.trim() : ''

    return {
      content: content || undefined,
    }
  },
  component: HomePage,
})

const FEED_PAGE_SIZE = 8
const MOTION_DELAY_SECONDS = 0.1
const gpuTransformTemplate: NonNullable<
  HTMLMotionProps<'div'>['transformTemplate']
> = (_, generatedTransform) =>
  generatedTransform ? `${generatedTransform} translateZ(0)` : 'translateZ(0)'
const GPU_ACCELERATED_MOTION_PROPS = {
  style: { willChange: 'transform, opacity' },
  transformTemplate: gpuTransformTemplate,
} as const

function HomePage() {
  const auth = useAuth()
  const { content } = Route.useSearch()
  const queryClient = useQueryClient()
  const prefersReducedMotion = useReducedMotion()
  const searchContent = content ?? ''
  const isSearchMode = searchContent.length > 0
  const [feedLimit, setFeedLimit] = useState(FEED_PAGE_SIZE)
  const [feedLimitSearchKey, setFeedLimitSearchKey] = useState(searchContent)
  const [publishDraft, setPublishDraft] = useState('')
  const [publishHidden, setPublishHidden] = useState(false)
  const [publishPinned, setPublishPinned] = useState(false)
  const [publishPin, setPublishPin] = useState('')
  const [publishError, setPublishError] = useState<string | null>(null)
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)
  const effectiveFeedLimit =
    feedLimitSearchKey === searchContent ? feedLimit : FEED_PAGE_SIZE

  const feedQuery = useQuery({
    ...postsQueryOptions(auth.status === 'authenticated', {
      content: searchContent || undefined,
      limit: effectiveFeedLimit,
      offset: 0,
      order: 'created_at_desc',
    }),
    placeholderData:
      feedLimitSearchKey === searchContent
        ? (previousData) => previousData
        : undefined,
  })
  const publishMutation = useMutation({
    mutationFn: () =>
      api.createAdminPost(
        {
          content: normalizeMarkdownContent(publishDraft),
          hide: publishHidden,
          pin: publishPinned ? Number(publishPin.trim()) : undefined,
        },
      ),
  })

  const feedData = feedQuery.data ?? {
    items: [],
    pinnedItems: [],
    total: 0,
  }
  const feedError =
    feedQuery.error instanceof Error ? feedQuery.error.message : null
  const loadedPostCount = feedData.items.length + feedData.pinnedItems.length
  const canLoadMorePosts = loadedPostCount < feedData.total

  useEffect(() => {
    if (feedLimitSearchKey !== searchContent) {
      setFeedLimit(FEED_PAGE_SIZE)
      setFeedLimitSearchKey(searchContent)
    }
  }, [feedLimitSearchKey, searchContent])

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

  async function refreshFeed() {
    await feedQuery.refetch()
  }

  async function handlePublish(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!hasMarkdownContent(publishDraft)) {
      setPublishError('发布内容不能为空')
      return
    }

    if (auth.status !== 'authenticated' || !auth.isAdmin) {
      setPublishError('只有管理员可以发布日志')
      return
    }

    const normalizedPin = publishPin.trim()

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
    }

    setPublishError(null)

    try {
      await publishMutation.mutateAsync()
      setPublishDraft('')
      setPublishHidden(false)
      setPublishPinned(false)
      setPublishPin('')
      await invalidatePublicFeedQueries(queryClient)
    } catch (error) {
      setPublishError(error instanceof Error ? error.message : '发布失败')
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <motion.section
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]"
        initial={false}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <div className="flex min-w-0 flex-col gap-6">
          {auth.status === 'authenticated' && auth.isAdmin ? (
            <motion.div
              animate={
                motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }
              }
              initial={false}
              transition={getEntranceTransition(0.06)}
              {...GPU_ACCELERATED_MOTION_PROPS}
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
                                onChange={(event) =>
                                  setPublishPin(event.target.value)
                                }
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
                      <Button
                        disabled={publishMutation.isPending}
                        type="submit"
                      >
                        <CornerRightUpIcon data-icon="inline-start" />
                        {publishMutation.isPending ? '正在发布…' : '立即发布'}
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
                {...GPU_ACCELERATED_MOTION_PROPS}
              >
                <Button
                  onClick={() => void refreshFeed()}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  <RefreshCwIcon data-icon="inline-start" />
                  刷新
                </Button>
              </motion.div>
            </div>

            {feedQuery.isPending ? (
              <div className="flex flex-col gap-4">
                {Array.from({ length: 3 }).map((_, index) => (
                  <Card
                    key={index}
                    className="border border-border/70 bg-card/90"
                  >
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

            {!feedQuery.isPending && feedError ? (
              <Alert variant="destructive">
                <AlertCircleIcon />
                <AlertTitle>日志流加载失败</AlertTitle>
                <AlertDescription>{feedError}</AlertDescription>
              </Alert>
            ) : null}

            {!feedQuery.isPending &&
            !feedError &&
            feedData.pinnedItems.length === 0 &&
            feedData.items.length === 0 ? (
              <Empty className="border border-dashed border-border/80 bg-card/60">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <AlertCircleIcon />
                  </EmptyMedia>
                  <EmptyTitle>
                    {isSearchMode ? '没有找到匹配的日志' : '还没有任何日志'}
                  </EmptyTitle>
                  <EmptyDescription>
                    {isSearchMode
                      ? `没有找到包含“${searchContent}”的日志内容。`
                      : '发布第一条日志后，这里会成为站点的主时间线。'}
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : null}

            {!feedQuery.isPending && !feedError ? (
              <div className="flex flex-col gap-4">
                <AnimatePresence initial={false}>
                  {[...feedData.pinnedItems, ...feedData.items].map(
                    (post, index) => (
                      <motion.div
                        key={post.id}
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
                        transition={getEntranceTransition(
                          Math.min(index * 0.03, 0.18),
                        )}
                        {...GPU_ACCELERATED_MOTION_PROPS}
                      >
                        <PostCard
                          onPostDeleted={() => {
                            void invalidatePublicFeedQueries(queryClient)
                          }}
                          post={post}
                        />
                      </motion.div>
                    ),
                  )}
                </AnimatePresence>
              </div>
            ) : null}

            {canLoadMorePosts ? (
              <motion.div
                whileHover={prefersReducedMotion ? undefined : { y: -1 }}
                whileTap={prefersReducedMotion ? undefined : { scale: 0.99 }}
                {...GPU_ACCELERATED_MOTION_PROPS}
              >
                <Button
                  disabled={feedQuery.isFetching}
                  onClick={() =>
                    setFeedLimit((current) => current + FEED_PAGE_SIZE)
                  }
                  type="button"
                  variant="outline"
                >
                  {feedQuery.isFetching ? '正在加载…' : '加载更多日志'}
                </Button>
              </motion.div>
            ) : null}
          </section>
        </div>

        <SiteSidebar />
      </motion.section>
    </div>
  )
}
