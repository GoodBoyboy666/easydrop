import { useQuery } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
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

function PostDetailPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const { id } = Route.useParams()
  const postId = Number(id)
  const [postState, setPostState] = useState<PostDTO | null>(null)
  const postQuery = useQuery({
    ...postQueryOptions(postId, auth.token),
    enabled: Number.isInteger(postId) && postId > 0,
  })

  useEffect(() => {
    if (postQuery.data) {
      setPostState(postQuery.data)
    }
  }, [postQuery.data])

  const errorMessage =
    postQuery.error instanceof Error ? postQuery.error.message : null

  if (!Number.isInteger(postId) || postId <= 0) {
    return (
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <Alert variant="destructive">
          <AlertTitle>页面地址无效</AlertTitle>
          <AlertDescription>说说 ID 不合法。</AlertDescription>
        </Alert>
      </div>
    )
  }

  if (postQuery.isPending && !postState) {
    return (
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
          <div className="flex min-w-0 flex-col gap-4">
            <Card className="border border-border/70 bg-card/90">
              <CardContent className="flex flex-col gap-4 pt-6">
                <Skeleton className="h-6 w-40" />
                <Skeleton className="h-4 w-52" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-10/12" />
              </CardContent>
            </Card>
            <Card className="border border-border/70 bg-card/90">
              <CardContent className="flex flex-col gap-4 pt-6">
                <Skeleton className="h-5 w-28" />
                <Skeleton className="h-10 w-24" />
                <Skeleton className="h-20 w-full" />
              </CardContent>
            </Card>
          </div>

          <SiteSidebar />
        </div>
      </div>
    )
  }

  if (postQuery.error && !postState) {
    return (
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
          <div className="min-w-0">
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
          </div>

          <SiteSidebar />
        </div>
      </div>
    )
  }

  if (!postState) {
    return null
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div className="flex min-w-0 flex-col gap-4">
          <PostCard
            onPostDeleted={() => void navigate({ to: '/' })}
            onPostUpdated={setPostState}
            post={postState}
            showComments={false}
          />

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
        </div>

        <SiteSidebar />
      </div>
    </div>
  )
}
