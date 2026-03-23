import { createFileRoute, Link } from '@tanstack/react-router'
import { MailIcon, ShieldCheckIcon, UserCircleIcon } from 'lucide-react'
import { useAuth } from '#/lib/auth'
import { formatDateTime, getInitials } from '#/lib/format'
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

export const Route = createFileRoute('/me')({
  component: MePage,
})

function MePage() {
  const auth = useAuth()

  if (auth.status !== 'authenticated' || !auth.user) {
    return (
      <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <Alert>
          <AlertTitle>需要先登录</AlertTitle>
          <AlertDescription>
            当前页面用于展示个人信息，请先登录后再查看。
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      <Card className="border border-border/70 bg-card/95 shadow-sm">
        <CardHeader>
          <CardTitle>个人信息</CardTitle>
          <CardDescription>首版展示当前登录用户的基础资料。</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          <div className="flex flex-col gap-4 rounded-2xl border border-border/70 bg-muted/30 p-4 sm:flex-row sm:items-center">
            <Avatar size="lg">
              <AvatarImage alt={auth.user.nickname} src={auth.user.avatar} />
              <AvatarFallback>{getInitials(auth.user.nickname)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <div className="text-lg font-semibold">{auth.user.nickname}</div>
              <div className="text-sm text-muted-foreground">@{auth.user.username}</div>
            </div>
            <div className="flex flex-wrap gap-2 sm:ml-auto">
              {auth.user.admin ? (
                <Badge variant="secondary">
                  <ShieldCheckIcon data-icon="inline-start" />
                  管理员
                </Badge>
              ) : (
                <Badge variant="outline">
                  <UserCircleIcon data-icon="inline-start" />
                  普通用户
                </Badge>
              )}
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
              <div className="text-sm text-muted-foreground">邮箱</div>
              <div className="mt-2 flex items-center gap-2 text-sm">
                <MailIcon />
                <span>{auth.user.email || '未设置邮箱'}</span>
              </div>
            </div>
            <div className="rounded-2xl border border-border/70 bg-background/80 p-4">
              <div className="text-sm text-muted-foreground">注册时间</div>
              <div className="mt-2 text-sm">
                {formatDateTime(auth.user.created_at)}
              </div>
            </div>
          </div>

          <div className="flex justify-end">
            <Button asChild variant="outline">
              <Link to="/me/comments">查看我的评论</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
