import { useQuery } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion, useReducedMotion } from 'motion/react'
import type { HTMLMotionProps, Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { postQueryOptions } from '#/lib/query-options'
import type { PostDTO } from '#/lib/types'
import { PostCard } from '#/components/home/post-card'
import { PostCommentsSection } from '#/components/post/post-comments-section'
import { SiteSidebar } from '#/components/site/site-sidebar'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Card, CardContent } from '#/components/ui/card'
import { Skeleton } from '#/components/ui/skeleton'

export const Route = createFileRoute('/posts/$id')({
  component: PostDetailPage,
})

const MOTION_DELAY_SECONDS = 0.1
const gpuTransformTemplate: NonNullable<
  HTMLMotionProps<'div'>['transformTemplate']
> = (_, generatedTransform) =>
  generatedTransform ? `${generatedTransform} translateZ(0)` : 'translateZ(0)'
const GPU_ACCELERATED_MOTION_PROPS = {
  style: { willChange: 'transform, opacity' },
  transformTemplate: gpuTransformTemplate,
} as const
const PAGE_ENTER_INITIAL = { opacity: 0, y: 10 }
const SECTION_ENTER_INITIAL = { opacity: 0, y: 12 }
const SIDEBAR_ENTER_INITIAL = { opacity: 0, x: 12 }

function PostDetailPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const prefersReducedMotion = useReducedMotion()
  const { id } = Route.useParams()
  const postId = Number(id)
  const [postState, setPostState] = useState<PostDTO | null>(null)
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)
  const postQuery = useQuery({
    ...postQueryOptions(postId, auth.token),
    enabled: Number.isInteger(postId) && postId > 0,
  })

  useEffect(() => {
    if (postQuery.data) {
      setPostState(postQuery.data)
    }
  }, [postQuery.data])

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

  const errorMessage =
    postQuery.error instanceof Error ? postQuery.error.message : null

  if (!Number.isInteger(postId) || postId <= 0) {
    return (
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8"
        initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
        key={`invalid-${postId}`}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <Alert variant="destructive">
          <AlertTitle>页面地址无效</AlertTitle>
          <AlertDescription>说说 ID 不合法。</AlertDescription>
        </Alert>
      </motion.div>
    )
  }

  if (postQuery.isPending && !postState) {
    return (
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8"
        initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
        key={`loading-${postId}`}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
          <div className="flex min-w-0 flex-col gap-4">
            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.05)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className="border border-border/70 bg-card/90">
                <CardContent className="flex flex-col gap-4 pt-6">
                  <Skeleton className="h-6 w-40" />
                  <Skeleton className="h-4 w-52" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-10/12" />
                </CardContent>
              </Card>
            </motion.div>
            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.1)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className="border border-border/70 bg-card/90">
                <CardContent className="flex flex-col gap-4 pt-6">
                  <Skeleton className="h-5 w-28" />
                  <Skeleton className="h-10 w-24" />
                  <Skeleton className="h-20 w-full" />
                </CardContent>
              </Card>
            </motion.div>
          </div>

          <motion.div
            animate={motionReady ? { opacity: 1, x: 0 } : { opacity: 0, x: 12 }}
            initial={prefersReducedMotion ? false : SIDEBAR_ENTER_INITIAL}
            transition={getEntranceTransition(0.12)}
            {...GPU_ACCELERATED_MOTION_PROPS}
          >
            <SiteSidebar />
          </motion.div>
        </div>
      </motion.div>
    )
  }

  if (postQuery.error && !postState) {
    return (
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8"
        initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
        key={`error-${postId}`}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
          <motion.div
            animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
            className="min-w-0"
            initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
            transition={getEntranceTransition(0.05)}
            {...GPU_ACCELERATED_MOTION_PROPS}
          >
            <Alert variant="destructive">
              <AlertTitle>
                {postQuery.error instanceof ApiError &&
                postQuery.error.status === 404
                  ? '说说不存在'
                  : '说说加载失败'}
              </AlertTitle>
              <AlertDescription>
                {errorMessage ?? '无法读取这条说说。'}
              </AlertDescription>
            </Alert>
          </motion.div>

          <motion.div
            animate={motionReady ? { opacity: 1, x: 0 } : { opacity: 0, x: 12 }}
            initial={prefersReducedMotion ? false : SIDEBAR_ENTER_INITIAL}
            transition={getEntranceTransition(0.1)}
            {...GPU_ACCELERATED_MOTION_PROPS}
          >
            <SiteSidebar />
          </motion.div>
        </div>
      </motion.div>
    )
  }

  if (!postState) {
    return null
  }

  return (
    <motion.div
      animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
      className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8"
      initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
      key={`post-${postId}`}
      transition={getEntranceTransition()}
      {...GPU_ACCELERATED_MOTION_PROPS}
    >
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div className="flex min-w-0 flex-col gap-4">
          <motion.div
            animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
            initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
            transition={getEntranceTransition(0.05)}
            {...GPU_ACCELERATED_MOTION_PROPS}
          >
            <PostCard
              onPostDeleted={() =>
                void navigate({ to: '/', search: { content: undefined } })
              }
              onPostUpdated={setPostState}
              post={postState}
              showComments={false}
            />
          </motion.div>

          <motion.div
            animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
            initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
            transition={getEntranceTransition(0.1)}
            {...GPU_ACCELERATED_MOTION_PROPS}
          >
            <Card className="bg-card/90 shadow-sm">
              <CardContent className="pt-2">
                <PostCommentsSection
                  alwaysExpanded
                  loginRedirectPath={`/posts/${postState.id}`}
                  onPostUpdated={setPostState}
                  post={postState}
                />
              </CardContent>
            </Card>
          </motion.div>
        </div>

        <motion.div
          animate={motionReady ? { opacity: 1, x: 0 } : { opacity: 0, x: 12 }}
          initial={prefersReducedMotion ? false : SIDEBAR_ENTER_INITIAL}
          transition={getEntranceTransition(0.12)}
          {...GPU_ACCELERATED_MOTION_PROPS}
        >
          <SiteSidebar />
        </motion.div>
      </div>
    </motion.div>
  )
}
