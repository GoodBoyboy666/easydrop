import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2Icon, MailCheckIcon, RotateCwIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { Transition } from 'motion/react'
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
  const prefersReducedMotion = useReducedMotion()
  const auth = useAuth()
  const { token } = Route.useSearch()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [submittedToken, setSubmittedToken] = useState<string | null>(null)
  const verifyMutation = useMutation({
    mutationFn: api.confirmVerifyEmail,
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
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
                transition={sectionTransition(0.16)}
              >
                <Alert variant="destructive">
                  <AlertTitle>缺少验证凭证</AlertTitle>
                  <AlertDescription>
                    当前链接没有包含有效 token，请重新打开邮件中的完整链接。
                  </AlertDescription>
                </Alert>
              </motion.div>
            ) : null}

            {token && verifyMutation.isPending ? (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
                transition={sectionTransition(0.2)}
              >
                <Alert>
                  <RotateCwIcon className="size-4 animate-spin" />
                  <AlertTitle>正在验证</AlertTitle>
                  <AlertDescription>
                    正在确认你的邮箱，请稍候。
                  </AlertDescription>
                </Alert>
              </motion.div>
            ) : null}

            {success ? (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
                transition={sectionTransition(0.24)}
              >
                <Alert>
                  <CheckCircle2Icon className="size-4" />
                  <AlertTitle>验证成功</AlertTitle>
                  <AlertDescription>{success}</AlertDescription>
                </Alert>
              </motion.div>
            ) : null}

            {error ? (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
                transition={sectionTransition(0.24)}
              >
                <Alert variant="destructive">
                  <AlertTitle>验证失败</AlertTitle>
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              </motion.div>
            ) : null}

            <motion.div
              animate={{ opacity: 1, y: 0 }}
              className="flex flex-col gap-3 sm:flex-row"
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              transition={sectionTransition(0.3)}
            >
              <Button asChild className="sm:flex-1">
                <Link to="/">返回首页</Link>
              </Button>
              <Button asChild className="sm:flex-1" variant="outline">
                <Link to="/login">前往登录</Link>
              </Button>
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
