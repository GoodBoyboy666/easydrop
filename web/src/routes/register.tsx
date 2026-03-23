import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRightIcon, UserPlusIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { safeRedirectPath } from '#/lib/format'
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
  FieldDescription,
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
  const { redirect } = Route.useSearch()
  const [form, setForm] = useState({
    email: '',
    nickname: '',
    password: '',
    username: '',
  })
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

    if (
      !form.username.trim() ||
      !form.nickname.trim() ||
      !form.email.trim() ||
      !form.password.trim()
    ) {
      setError('请完整填写注册信息')
      return
    }

    setSubmitting(true)

    try {
      await auth.register({
        email: form.email.trim(),
        nickname: form.nickname.trim(),
        password: form.password,
        username: form.username.trim(),
      })
      window.location.assign(safeRedirectPath(redirect))
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : '注册失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-2xl border border-border/70 bg-card/95 shadow-sm">
        <CardHeader>
          <CardTitle className="text-2xl">注册</CardTitle>
          <CardDescription>
            注册完成后会自动登录并返回首页，你就可以开始参与评论。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="username">用户名</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  id="username"
                  onChange={(event) =>
                    setForm((current) => ({ ...current, username: event.target.value }))
                  }
                  placeholder="用于登录的用户名"
                  value={form.username}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="nickname">昵称</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  id="nickname"
                  onChange={(event) =>
                    setForm((current) => ({ ...current, nickname: event.target.value }))
                  }
                  placeholder="站内展示昵称"
                  value={form.nickname}
                />
              </Field>

              <Field data-invalid={!!error}>
                <FieldLabel htmlFor="email">邮箱</FieldLabel>
                <Input
                  aria-invalid={!!error}
                  id="email"
                  onChange={(event) =>
                    setForm((current) => ({ ...current, email: event.target.value }))
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
                  id="register-password"
                  onChange={(event) =>
                    setForm((current) => ({ ...current, password: event.target.value }))
                  }
                  placeholder="至少准备一个安全密码"
                  type="password"
                  value={form.password}
                />
                <FieldDescription>
                  注册接口会直接返回登录态，提交成功后会自动进入已登录状态。
                </FieldDescription>
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
                <UserPlusIcon data-icon="inline-start" />
                {submitting ? '正在注册…' : '创建账号'}
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
