import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRightIcon, LogInIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { safeRedirectPath } from '#/lib/format'
import { useSiteSettings } from '#/lib/site-settings'
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
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'

export const Route = createFileRoute('/login')({
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: typeof search.redirect === 'string' ? search.redirect : '/',
  }),
  component: LoginPage,
})

function LoginPage() {
  const auth = useAuth()
  const { allowRegister } = useSiteSettings()
  const { redirect } = Route.useSearch()
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')
  const [captchaNotice, setCaptchaNotice] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    void (async () => {
      try {
        const config = await api.getCaptchaConfig()

        if (config.enabled) {
          setCaptchaNotice(
            `当前站点启用了 ${config.provider || '验证码'}，骨架版暂未接入可视化验证控件。`
          )
        }
      } catch {
        setCaptchaNotice(null)
      }
    })()
  }, [])

  useEffect(() => {
    if (auth.status === 'authenticated') {
      window.location.assign(safeRedirectPath(redirect))
    }
  }, [auth.status, redirect])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (!account.trim() || !password.trim()) {
      setError('请填写账号和密码')
      return
    }

    setSubmitting(true)

    try {
      await auth.login({
        account: account.trim(),
        password,
      })
      window.location.assign(safeRedirectPath(redirect))
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : '登录失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-xl border border-border/70 bg-card/95 shadow-sm px-4 py-8">
        <CardHeader>
          <CardTitle className="text-2xl">欢迎回来</CardTitle>
          <CardDescription>
            Welcome back
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="account">账号或邮箱</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  id="account"
                  onChange={(event) => setAccount(event.target.value)}
                  placeholder="请输入用户名或邮箱"
                  value={account}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="password">密码</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  id="password"
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="请输入密码"
                  type="password"
                  value={password}
                />
                <FieldError>{error}</FieldError>
              </Field>
            </FieldGroup>

            {captchaNotice ? (
              <Alert className="mt-4">
                <AlertTitle>验证码提示</AlertTitle>
                <AlertDescription>{captchaNotice}</AlertDescription>
              </Alert>
            ) : null}

            <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <Button disabled={submitting} type="submit">
                <LogInIcon data-icon="inline-start" />
                {submitting ? '正在登录…' : '立即登录'}
              </Button>
              {allowRegister ? (
                <Button asChild size="sm" variant="ghost">
                  <Link to="/register" search={{ redirect }}>
                    没有账号？去注册
                    <ArrowRightIcon data-icon="inline-end" />
                  </Link>
                </Button>
              ) : (
                <div className="text-sm text-muted-foreground">当前站点未开放注册</div>
              )}
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
