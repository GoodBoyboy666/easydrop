import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ArrowRightIcon, LogInIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import {
  CaptchaPanel,
  createEmptyCaptchaInput,
  isCaptchaComplete,
} from '#/components/site/captcha-panel'
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

export const Route = createFileRoute('/login')({
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: typeof search.redirect === 'string' ? search.redirect : '/',
  }),
  component: LoginPage,
})

function LoginPage() {
  const auth = useAuth()
  const prefersReducedMotion = useReducedMotion()
  const { allowRegister } = useSiteSettings()
  const { redirect } = Route.useSearch()
  const [account, setAccount] = useState('')
  const [captcha, setCaptcha] = useState(createEmptyCaptchaInput)
  const [captchaResetSignal, setCaptchaResetSignal] = useState(0)
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const captchaConfigQuery = useQuery(captchaConfigQueryOptions())
  const loginMutation = useMutation({
    mutationFn: api.login,
  })
  const pageTransition: Transition = prefersReducedMotion
    ? { duration: 0 }
    : { duration: 0.34, ease: 'easeOut' }

  const sectionTransition = (delay: number): Transition =>
    prefersReducedMotion
      ? { duration: 0 }
      : {
          duration: 0.3,
          ease: 'easeOut',
          delay,
        }

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

    if (!isCaptchaComplete(captchaConfigQuery.data, captcha)) {
      setError('请先完成验证码')
      return
    }

    try {
      await loginMutation.mutateAsync({
        account: account.trim(),
        captcha: captchaConfigQuery.data?.enabled ? captcha : undefined,
        password,
      })
      await auth.refreshUser()
      window.location.assign(safeRedirectPath(redirect))
    } catch (submitError) {
      if (captchaConfigQuery.data?.enabled) {
        setCaptcha(createEmptyCaptchaInput())
        setCaptchaResetSignal((current) => current + 1)
      }
      setError(submitError instanceof Error ? submitError.message : '登录失败')
    }
  }

  return (
    <motion.div
      animate={{ opacity: 1, y: 0 }}
      className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8"
      initial={prefersReducedMotion ? false : { opacity: 0, y: 20 }}
      transition={pageTransition}
    >
      <motion.div
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="flex w-full justify-center"
        initial={prefersReducedMotion ? false : { opacity: 0, y: 16, scale: 0.98 }}
        transition={sectionTransition(0.08)}
      >
        <Card className="w-full max-w-xl border border-border/70 bg-card/95 shadow-sm px-4 py-8">
          <CardHeader>
            <CardTitle className="text-2xl">欢迎回来</CardTitle>
            <CardDescription>Welcome back</CardDescription>
          </CardHeader>
          <CardContent>
            <motion.form
              animate={{ opacity: 1, y: 0 }}
              initial={prefersReducedMotion ? false : { opacity: 0, y: 12 }}
              onSubmit={handleSubmit}
              transition={sectionTransition(0.14)}
            >
            <motion.div
              animate={{ opacity: 1, y: 0 }}
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              transition={sectionTransition(0.2)}
            >
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
            </motion.div>

            <motion.div
              animate={{ opacity: 1, y: 0 }}
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              transition={sectionTransition(0.26)}
            >
              <CaptchaPanel
                config={captchaConfigQuery.data}
                errorMessage={
                  captchaConfigQuery.error instanceof Error
                    ? captchaConfigQuery.error.message
                    : null
                }
                isLoading={captchaConfigQuery.isLoading}
                onChange={setCaptcha}
                resetSignal={captchaResetSignal}
                value={captcha}
              />
            </motion.div>

            <motion.div
              animate={{ opacity: 1, y: 0 }}
              className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              transition={sectionTransition(0.32)}
            >
              <Button disabled={loginMutation.isPending} type="submit">
                <LogInIcon data-icon="inline-start" />
                {loginMutation.isPending ? '正在登录…' : '立即登录'}
              </Button>
              <div className="flex flex-col items-start gap-2 sm:items-end">
                <Button asChild size="sm" variant="ghost">
                  <Link to="/forgot-password">忘记密码？</Link>
                </Button>
                {allowRegister ? (
                  <Button asChild size="sm" variant="ghost">
                    <Link to="/register" search={{ redirect }}>
                      没有账号？去注册
                      <ArrowRightIcon data-icon="inline-end" />
                    </Link>
                  </Button>
                ) : (
                  <div className="text-sm text-muted-foreground">
                    当前站点未开放注册
                  </div>
                )}
              </div>
            </motion.div>
            </motion.form>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
