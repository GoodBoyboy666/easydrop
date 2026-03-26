import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ArrowRightIcon, UserPlusIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { safeRedirectPath } from '#/lib/format'
import { captchaConfigQueryOptions } from '#/lib/query-options'
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

export const Route = createFileRoute('/register')({
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: typeof search.redirect === 'string' ? search.redirect : '/',
  }),
  component: RegisterPage,
})

function RegisterPage() {
  const auth = useAuth()
  const { allowRegister, siteName } = useSiteSettings()
  const { redirect } = Route.useSearch()
  const [form, setForm] = useState({
    email: '',
    nickname: '',
    password: '',
    username: '',
  })
  const [error, setError] = useState<string | null>(null)
  const captchaConfigQuery = useQuery(captchaConfigQueryOptions())
  const registerMutation = useMutation({
    mutationFn: api.register,
  })
  const captchaNotice = captchaConfigQuery.data?.enabled
    ? `当前站点启用了 ${captchaConfigQuery.data.provider || '验证码'}，骨架版暂未接入可视化验证控件。`
    : null

  useEffect(() => {
    if (auth.status === 'authenticated') {
      window.location.assign(safeRedirectPath(redirect))
    }
  }, [auth.status, redirect])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (!allowRegister) {
      setError('当前站点未开放注册')
      return
    }

    if (
      !form.username.trim() ||
      !form.nickname.trim() ||
      !form.email.trim() ||
      !form.password.trim()
    ) {
      setError('请完整填写注册信息')
      return
    }

    try {
      await registerMutation.mutateAsync({
        email: form.email.trim(),
        nickname: form.nickname.trim(),
        password: form.password,
        username: form.username.trim(),
      })
      await auth.refreshUser()
      window.location.assign(safeRedirectPath(redirect))
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : '注册失败')
    }
  }

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-2xl border border-border/70 bg-card/95 shadow-sm px-4 py-8">
        <CardHeader>
          <CardTitle className="text-2xl">注册</CardTitle>
          <CardDescription>Sign up</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="username">用户名</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  disabled={!allowRegister}
                  id="username"
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      username: event.target.value,
                    }))
                  }
                  placeholder="用于登录的用户名"
                  value={form.username}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="nickname">昵称</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  disabled={!allowRegister}
                  id="nickname"
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      nickname: event.target.value,
                    }))
                  }
                  placeholder="站内展示昵称"
                  value={form.nickname}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="email">邮箱</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  disabled={!allowRegister}
                  id="email"
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      email: event.target.value,
                    }))
                  }
                  placeholder="用于接收通知的邮箱"
                  type="email"
                  value={form.email}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="register-password">密码</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  disabled={!allowRegister}
                  id="register-password"
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      password: event.target.value,
                    }))
                  }
                  placeholder="一个安全的密码"
                  type="password"
                  value={form.password}
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

            {!allowRegister ? (
              <Alert className="mt-4">
                <AlertTitle>当前未开放注册</AlertTitle>
                <AlertDescription>
                  {siteName} 当前关闭了公开注册，请联系管理员开通账号。
                </AlertDescription>
              </Alert>
            ) : null}

            <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <Button
                disabled={!allowRegister || registerMutation.isPending}
                type="submit"
              >
                <UserPlusIcon data-icon="inline-start" />
                {registerMutation.isPending ? '正在注册…' : '创建账号'}
              </Button>
              <Button asChild size="sm" variant="ghost">
                <Link to="/login" search={{ redirect }}>
                  已有账号？去登录
                  <ArrowRightIcon data-icon="inline-end" />
                </Link>
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
