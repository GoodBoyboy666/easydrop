import { createFileRoute, Link } from '@tanstack/react-router'
import { ShieldCheckIcon } from 'lucide-react'
import { useAuth } from '#/lib/auth'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'

export const Route = createFileRoute('/admin')({
  component: AdminPlaceholderPage,
})

function AdminPlaceholderPage() {
  const auth = useAuth()

  if (auth.status !== 'authenticated' || !auth.isAdmin) {
    return (
      <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <Alert variant="destructive">
          <AlertTitle>无权访问后台管理</AlertTitle>
          <AlertDescription>当前账号不是管理员，因此不能进入后台管理入口。</AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      <Card className="border border-border/70 bg-card/95 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheckIcon />
            后台管理
          </CardTitle>
          <CardDescription>
            这是管理员后台的站内占位入口，当前骨架阶段先提供进入点和权限校验。
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <p className="text-sm leading-7 text-muted-foreground">
            后续可以在这里继续接入日志管理、评论审核、标签管理和站点设置等完整后台功能。
          </p>
          <div className="flex justify-end">
            <Button asChild variant="outline">
              <Link to="/">返回首页</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
