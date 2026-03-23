import { createFileRoute, Link } from '@tanstack/react-router'
import { MessageSquareTextIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatDateTime, formatRelativeTime } from '#/lib/format'
import type { CommentDTO } from '#/lib/types'
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
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

export const Route = createFileRoute('/me/comments')({
  component: MyCommentsPage,
})

function MyCommentsPage() {
  const auth = useAuth()
  const [comments, setComments] = useState<CommentDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = auth.token

    if (!token) {
      setLoading(false)
      return
    }

    void (async () => {
      try {
        const result = await api.getMyComments(token, {
          limit: 20,
          offset: 0,
          order: 'created_at_desc',
        })

        setComments(result.items)
        setError(null)
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : '评论加载失败')
      } finally {
        setLoading(false)
      }
    })()
  }, [auth.token])

  if (!auth.token) {
    return (
      <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <Alert>
          <AlertTitle>需要先登录</AlertTitle>
          <AlertDescription>登录后才可以查看自己发表过的评论。</AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-4">
        <h1 className="font-heading text-2xl font-semibold">我的评论</h1>
        <p className="text-sm text-muted-foreground">
          从 `/users/me/comments` 拉取当前登录用户的评论列表。
        </p>
      </div>

      {loading ? (
        <div className="flex flex-col gap-4">
          {Array.from({ length: 3 }).map((_, index) => (
            <Card key={index} className="border border-border/70 bg-card/90">
              <CardContent className="flex flex-col gap-3 pt-4">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-8/12" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : null}

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>我的评论读取失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {!loading && !error && comments.length === 0 ? (
        <Empty className="border border-dashed border-border/80 bg-card/80">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <MessageSquareTextIcon />
            </EmptyMedia>
            <EmptyTitle>你还没有评论</EmptyTitle>
            <EmptyDescription>
              回到首页，在任意日志下点击“发评论”即可参与讨论。
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : null}

      {!loading && !error ? (
        <div className="flex flex-col gap-4">
          {comments.map((comment) => (
            <Card key={comment.id} className="border border-border/70 bg-card/90 shadow-sm">
              <CardHeader className="px-4 py-4">
                <CardTitle className="text-sm">日志 #{comment.post_id}</CardTitle>
                <CardDescription>
                  {formatRelativeTime(comment.created_at)} · {formatDateTime(comment.created_at)}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3 px-4 pb-4">
                <MarkdownContent compact content={comment.content} />
                <div className="flex justify-end">
                  <Button asChild size="sm" variant="ghost">
                    <Link to="/">返回首页继续浏览</Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : null}
    </div>
  )
}
