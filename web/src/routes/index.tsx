import { useEffect, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import {
  LoaderCircleIcon,
  RefreshCcwIcon,
  TriangleAlertIcon,
} from 'lucide-react'

import type {
  Comment,
  CurrentUser,
  Post,
  PublicSettingsRecord,
  Tag,
} from '#/lib/easydrop-api'
import {
  ApiError,
  createAdminPost,
  getAccessTokenFromStorage,
  getCurrentUser,
  getPublicSettings,
  GLOBAL_COMMENTS_PAGE_SIZE,
  isRootComment,
  listComments,
  listPostComments,
  listPosts,
  listTags,
  POSTS_PAGE_SIZE,
  TAGS_PAGE_SIZE,
} from '#/lib/easydrop-api'
import { AdminComposer } from '#/components/memos/admin-composer'
import { PostCard } from '#/components/memos/post-card'
import { SiteSidebar } from '#/components/memos/site-sidebar'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Button } from '#/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '#/components/ui/empty'
import { Skeleton } from '#/components/ui/skeleton'

type CommentsState = {
  error: string | null
  items: Comment[]
  status: 'idle' | 'loading' | 'success' | 'error'
}

type ModuleState<T> = {
  error: string | null
  items: T[]
  status: 'loading' | 'success' | 'error'
}

type AuthStatus = 'loading' | 'anonymous' | 'user' | 'admin'

export const Route = createFileRoute('/')({
  head: () => ({
    meta: [
      {
        title: 'EasyDrop',
      },
    ],
  }),
  component: HomePage,
})

function TimelineSkeleton() {
  return (
    <div className="flex flex-col gap-4">
      {[0, 1, 2].map((item) => (
        <div
          className="rounded-3xl border border-border/70 bg-card/80 p-5 shadow-[0_24px_64px_rgba(15,23,42,0.05)]"
          key={item}
        >
          <div className="flex items-start gap-3">
            <Skeleton className="size-10 rounded-full" />
            <div className="flex flex-1 flex-col gap-2">
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-3 w-52" />
            </div>
          </div>
          <div className="mt-5 flex flex-col gap-3">
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-11/12" />
            <Skeleton className="h-3 w-4/5" />
          </div>
        </div>
      ))}
    </div>
  )
}

function HomePage() {
  const [posts, setPosts] = useState<Post[]>([])
  const [postsTotal, setPostsTotal] = useState(0)
  const [postsStatus, setPostsStatus] = useState<
    'loading' | 'success' | 'error'
  >('loading')
  const [postsError, setPostsError] = useState<string | null>(null)
  const [isLoadingMore, setIsLoadingMore] = useState(false)

  const [settings, setSettings] = useState<PublicSettingsRecord>({})
  const [settingsError, setSettingsError] = useState<string | null>(null)

  const [authStatus, setAuthStatus] = useState<AuthStatus>('loading')
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null)

  const [composerContent, setComposerContent] = useState('')
  const [composerHide, setComposerHide] = useState(false)
  const [composerError, setComposerError] = useState<string | null>(null)
  const [composerSuccess, setComposerSuccess] = useState<string | null>(null)
  const [isComposerSubmitting, setIsComposerSubmitting] = useState(false)

  const [latestRootCommentsState, setLatestRootCommentsState] = useState<
    ModuleState<Comment>
  >({
    error: null,
    items: [],
    status: 'loading',
  })
  const [tagsState, setTagsState] = useState<ModuleState<Tag>>({
    error: null,
    items: [],
    status: 'loading',
  })

  const [openPostIds, setOpenPostIds] = useState<
    Partial<Record<number, boolean>>
  >({})
  const [commentsByPost, setCommentsByPost] = useState<
    Partial<Record<number, CommentsState>>
  >({})

  useEffect(() => {
    void loadInitialPosts()
    void loadSettings()
    void loadLatestRootComments()
    void loadTagsModule()
    void loadAuth()
  }, [])

  async function loadInitialPosts() {
    setPostsStatus('loading')
    setPostsError(null)

    try {
      const postResult = await listPosts(0, POSTS_PAGE_SIZE)
      setPosts(postResult.items ?? [])
      setPostsTotal(postResult.total ?? 0)
      setPostsStatus('success')
    } catch (error) {
      const message =
        error instanceof Error ? error.message : '首页加载失败，请稍后重试。'
      setPostsStatus('error')
      setPostsError(message)
    }
  }

  async function loadSettings() {
    setSettingsError(null)

    try {
      const publicSettings = await getPublicSettings()
      setSettings(publicSettings)
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : '公开配置加载失败，请稍后重试。'
      setSettingsError(message)
    }
  }

  async function loadLatestRootComments() {
    setLatestRootCommentsState({
      error: null,
      items: [],
      status: 'loading',
    })

    try {
      const result = await listComments(0, GLOBAL_COMMENTS_PAGE_SIZE)
      const rootComments = (result.items ?? [])
        .filter(isRootComment)
        .slice(0, 5)

      setLatestRootCommentsState({
        error: null,
        items: rootComments,
        status: 'success',
      })
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : '最新评论加载失败，请稍后重试。'

      setLatestRootCommentsState({
        error: message,
        items: [],
        status: 'error',
      })
    }
  }

  async function loadTagsModule() {
    setTagsState({
      error: null,
      items: [],
      status: 'loading',
    })

    try {
      const preferred = await listTags(0, TAGS_PAGE_SIZE, 'hot_desc')
      setTagsState({
        error: null,
        items: preferred.items ?? [],
        status: 'success',
      })
    } catch (error) {
      try {
        const fallback = await listTags(0, TAGS_PAGE_SIZE, 'created_at_desc')
        setTagsState({
          error: null,
          items: fallback.items ?? [],
          status: 'success',
        })
      } catch (fallbackError) {
        const message =
          fallbackError instanceof Error
            ? fallbackError.message
            : error instanceof Error
              ? error.message
              : '标签加载失败，请稍后重试。'

        setTagsState({
          error: message,
          items: [],
          status: 'error',
        })
      }
    }
  }

  async function loadAuth() {
    const token = getAccessTokenFromStorage()

    if (!token) {
      setAuthStatus('anonymous')
      setCurrentUser(null)
      return
    }

    setAuthStatus('loading')

    try {
      const user = await getCurrentUser(token)
      setCurrentUser(user)
      setAuthStatus(user.admin ? 'admin' : 'user')
    } catch (error) {
      setCurrentUser(null)

      if (error instanceof ApiError && error.status === 401) {
        setAuthStatus('anonymous')
        return
      }

      setAuthStatus('anonymous')
    }
  }

  async function reloadTimeline() {
    setPostsStatus('loading')
    setPostsError(null)

    try {
      const postResult = await listPosts(0, POSTS_PAGE_SIZE)
      setPosts(postResult.items ?? [])
      setPostsTotal(postResult.total ?? 0)
      setPostsStatus('success')
      setOpenPostIds({})
      setCommentsByPost({})
    } catch (error) {
      const message =
        error instanceof Error ? error.message : '刷新失败，请稍后重试。'
      setPostsStatus('error')
      setPostsError(message)
    }
  }

  async function loadMorePosts() {
    setIsLoadingMore(true)
    setPostsError(null)

    try {
      const result = await listPosts(posts.length, POSTS_PAGE_SIZE)
      setPosts((current) => [...current, ...(result.items ?? [])])
      setPostsTotal(result.total ?? postsTotal)
    } catch (error) {
      const message =
        error instanceof Error ? error.message : '加载更多失败，请稍后重试。'
      setPostsError(message)
    } finally {
      setIsLoadingMore(false)
    }
  }

  async function loadComments(postId: number) {
    setCommentsByPost((current) => ({
      ...current,
      [postId]: {
        error: null,
        items: current[postId]?.items ?? [],
        status: 'loading',
      },
    }))

    try {
      const result = await listPostComments(postId)
      setCommentsByPost((current) => ({
        ...current,
        [postId]: {
          error: null,
          items: result.items ?? [],
          status: 'success',
        },
      }))
    } catch (error) {
      const message =
        error instanceof Error ? error.message : '评论加载失败，请稍后重试。'
      setCommentsByPost((current) => ({
        ...current,
        [postId]: {
          error: message,
          items: current[postId]?.items ?? [],
          status: 'error',
        },
      }))
    }
  }

  function handleCommentsToggle(postId: number, open: boolean) {
    setOpenPostIds((current) => ({
      ...current,
      [postId]: open,
    }))

    if (!open) {
      return
    }

    const state = commentsByPost[postId]

    if (state?.status === 'loading' || state?.status === 'success') {
      return
    }

    void loadComments(postId)
  }

  async function handleComposerSubmit() {
    const token = getAccessTokenFromStorage()
    const normalizedContent = composerContent.trim()

    if (!token) {
      setAuthStatus('anonymous')
      setCurrentUser(null)
      return
    }

    if (!normalizedContent) {
      setComposerError('请先填写说说内容。')
      setComposerSuccess(null)
      return
    }

    setIsComposerSubmitting(true)
    setComposerError(null)
    setComposerSuccess(null)

    try {
      await createAdminPost(token, {
        content: normalizedContent,
        hide: composerHide,
      })

      setComposerContent('')
      setComposerHide(false)
      setComposerSuccess('新说说已发布，首页时间流已刷新。')
      await reloadTimeline()
    } catch (error) {
      const message =
        error instanceof Error ? error.message : '发布失败，请稍后重试。'
      setComposerError(message)

      if (error instanceof ApiError) {
        if (error.status === 401) {
          setAuthStatus('anonymous')
          setCurrentUser(null)
        }

        if (error.status === 403) {
          setAuthStatus('user')
          setCurrentUser((current) =>
            current ? { ...current, admin: false } : current,
          )
        }
      }
    } finally {
      setIsComposerSubmitting(false)
    }
  }

  const hasMorePosts = posts.length < postsTotal
  const showComposer = authStatus === 'admin' && currentUser

  return (
    <main className="page-wrap px-4 pb-16 pt-8 sm:pt-10">
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] xl:grid-cols-[minmax(0,1fr)_340px]">
        <section className="flex min-w-0 flex-col gap-5">
          {showComposer ? (
            <AdminComposer
              content={composerContent}
              error={composerError}
              hide={composerHide}
              isSubmitting={isComposerSubmitting}
              onContentChange={(content) => {
                setComposerContent(content)
                setComposerError(null)
                setComposerSuccess(null)
              }}
              onHideChange={setComposerHide}
              onSubmit={() => void handleComposerSubmit()}
              successMessage={composerSuccess}
              user={currentUser}
            />
          ) : null}

          {postsStatus === 'error' ? (
            <Alert
              className="border-destructive/30 bg-destructive/5"
              variant="destructive"
            >
              <TriangleAlertIcon />
              <AlertTitle>时间流加载失败</AlertTitle>
              <AlertDescription>
                {postsError ?? '暂时无法获取说说内容。'}
              </AlertDescription>
              <Button
                className="mt-2 w-fit"
                onClick={() => void reloadTimeline()}
                size="sm"
                variant="outline"
              >
                <RefreshCcwIcon data-icon="inline-start" />
                重新加载
              </Button>
            </Alert>
          ) : null}

          {postsStatus === 'loading' ? <TimelineSkeleton /> : null}

          {postsStatus === 'success' && posts.length === 0 ? (
            <Empty className="rounded-[2rem] border border-border/70 bg-card/85 py-12 shadow-[0_24px_64px_rgba(15,23,42,0.05)]">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <LoaderCircleIcon />
                </EmptyMedia>
                <EmptyTitle>这里还没有公开说说</EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : null}

          {posts.length > 0 ? (
            <div className="flex flex-col gap-4">
              {posts.map((post) => {
                const commentsState = commentsByPost[post.id] ?? {
                  error: null,
                  items: [],
                  status: 'idle' as const,
                }

                return (
                  <PostCard
                    comments={commentsState.items}
                    commentsError={commentsState.error}
                    commentsLoading={commentsState.status === 'loading'}
                    commentsOpen={Boolean(openPostIds[post.id])}
                    key={post.id}
                    onRetryComments={() => void loadComments(post.id)}
                    onToggleComments={(open) =>
                      handleCommentsToggle(post.id, open)
                    }
                    post={post}
                  />
                )
              })}
            </div>
          ) : null}

          {posts.length > 0 ? (
            <div className="flex justify-center pt-2">
              {hasMorePosts ? (
                <Button
                  disabled={isLoadingMore}
                  onClick={() => void loadMorePosts()}
                  size="lg"
                  variant="outline"
                >
                  {isLoadingMore ? (
                    <LoaderCircleIcon
                      className="animate-spin"
                      data-icon="inline-start"
                    />
                  ) : null}
                  {isLoadingMore ? '加载中...' : '加载更多'}
                </Button>
              ) : (
                <p className="text-sm text-muted-foreground">
                  已经读到当前时间流底部了。
                </p>
              )}
            </div>
          ) : null}
        </section>

        <aside className="min-w-0">
          <SiteSidebar
            latestRootComments={latestRootCommentsState.items}
            latestRootCommentsError={latestRootCommentsState.error}
            latestRootCommentsStatus={latestRootCommentsState.status}
            settings={settings}
            settingsError={settingsError}
            tags={tagsState.items}
            tagsError={tagsState.error}
            tagsStatus={tagsState.status}
          />
        </aside>
      </div>
    </main>
  )
}
