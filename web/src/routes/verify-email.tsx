import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2Icon, MailCheckIcon, RotateCwIcon } from 'lucide-react'
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

export const Route = createFileRoute('/verify-email')({
  validateSearch: (search: Record<string, unknown>) => ({
    token: typeof search.token === 'string' ? search.token : '',
  }),
  component: VerifyEmailPage,
})

function VerifyEmailPage() {
  const auth = useAuth()
  const { token } = Route.useSearch()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [submittedToken, setSubmittedToken] = useState<string | null>(null)
  const verifyMutation = useMutation({
    mutationFn: api.confirmVerifyEmail,
  })

  useEffect(() => {
    if (!token || submittedToken === token || verifyMutation.isPending) {
      return
    }

    let cancelled = false
    setSubmittedToken(token)
    setError(null)
    setSuccess(null)

    verifyMutation
      .mutateAsync({ token })
      .then(async () => {
        if (cancelled) {
          return
        }
        setSuccess('邮箱验证已完成，你现在可以正常使用需要验证邮箱的功能。')
        if (auth.status === 'authenticated') {
          try {
            await auth.refreshUser()
          } catch {
            // 页面提示以当前操作结果为准，刷新用户信息失败不影响确认结果。
          }
        }
      })
      .catch((submitError) => {
        if (cancelled) {
          return
        }
        setError(
          submitError instanceof Error ? submitError.message : '邮箱验证失败',
        )
      })

    return () => {
      cancelled = true
    }
  }, [auth, submittedToken, token, verifyMutation])

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <MailCheckIcon className="size-6" />
            验证邮箱
          </CardTitle>
          <CardDescription>
            正在处理邮件中的验证链接。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!token ? (
            <Alert variant="destructive">
              <AlertTitle>缺少验证凭证</AlertTitle>
              <AlertDescription>
                当前链接没有包含有效 token，请重新打开邮件中的完整链接。
              </AlertDescription>
            </Alert>
          ) : null}

          {token && verifyMutation.isPending ? (
            <Alert>
              <RotateCwIcon className="size-4 animate-spin" />
              <AlertTitle>正在验证</AlertTitle>
              <AlertDescription>
                正在确认你的邮箱，请稍候。
              </AlertDescription>
            </Alert>
          ) : null}

          {success ? (
            <Alert>
              <CheckCircle2Icon className="size-4" />
              <AlertTitle>验证成功</AlertTitle>
              <AlertDescription>{success}</AlertDescription>
            </Alert>
          ) : null}

          {error ? (
            <Alert variant="destructive">
              <AlertTitle>验证失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}

          <div className="flex flex-col gap-3 sm:flex-row">
            <Button asChild className="sm:flex-1">
              <Link to="/">返回首页</Link>
            </Button>
            <Button asChild className="sm:flex-1" variant="outline">
              <Link to="/login">前往登录</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
