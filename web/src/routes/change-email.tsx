import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2Icon, MailPlusIcon, RotateCwIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
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

export const Route = createFileRoute('/change-email')({
  validateSearch: (search: Record<string, unknown>) => ({
    token: typeof search.token === 'string' ? search.token : '',
  }),
  component: ChangeEmailPage,
})

function ChangeEmailPage() {
  const auth = useAuth()
  const { token } = Route.useSearch()
  const [error, setError] = useState<string | null>(null)
  const [successEmail, setSuccessEmail] = useState<string | null>(null)
  const [submittedToken, setSubmittedToken] = useState<string | null>(null)
  const confirmMutation = useMutation({
    mutationFn: api.confirmEmailChange,
  })

  useEffect(() => {
    if (!token || submittedToken === token || confirmMutation.isPending) {
      return
    }

    let cancelled = false
    setSubmittedToken(token)
    setError(null)
    setSuccessEmail(null)

    confirmMutation
      .mutateAsync({ token })
      .then(async (user) => {
        if (cancelled) {
          return
        }
        setSuccessEmail(user.email ?? '新邮箱')
        if (auth.status === 'authenticated') {
          try {
            await auth.refreshUser()
          } catch {
            // 邮箱变更已完成，刷新资料失败不影响当前确认结果。
          }
        }
      })
      .catch((submitError) => {
        if (cancelled) {
          return
        }
        setError(
          submitError instanceof Error
            ? submitError.message
            : '确认邮箱修改失败',
        )
      })

    return () => {
      cancelled = true
    }
  }, [auth, confirmMutation, submittedToken, token])

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <MailPlusIcon className="size-6" />
            确认修改邮箱
          </CardTitle>
          <CardDescription>
            正在确认这次邮箱变更请求。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!token ? (
            <Alert variant="destructive">
              <AlertTitle>缺少确认凭证</AlertTitle>
              <AlertDescription>
                当前链接没有包含有效 token，请重新打开新邮箱中的完整链接。
              </AlertDescription>
            </Alert>
          ) : null}

          {token && confirmMutation.isPending ? (
            <Alert>
              <RotateCwIcon className="size-4 animate-spin" />
              <AlertTitle>正在确认</AlertTitle>
              <AlertDescription>
                邮箱变更处理中，请稍候。
              </AlertDescription>
            </Alert>
          ) : null}

          {successEmail ? (
            <Alert>
              <CheckCircle2Icon className="size-4" />
              <AlertTitle>邮箱已更新</AlertTitle>
              <AlertDescription>
                当前账号邮箱已切换为 {successEmail}。
              </AlertDescription>
            </Alert>
          ) : null}

          {error ? (
            <Alert variant="destructive">
              <AlertTitle>确认失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}

          <div className="flex flex-col gap-3 sm:flex-row">
            <Button asChild className="sm:flex-1">
              <Link to="/">返回首页</Link>
            </Button>
            <Button asChild className="sm:flex-1" variant="outline">
              <Link to="/me">查看个人资料</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
