import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2Icon, KeyRoundIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { Transition } from 'motion/react'
import { useState } from 'react'
import { api } from '#/lib/api'
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

export const Route = createFileRoute('/reset-password')({
  validateSearch: (search: Record<string, unknown>) => ({
    token: typeof search.token === 'string' ? search.token : '',
  }),
  component: ResetPasswordPage,
})

function ResetPasswordPage() {
  const prefersReducedMotion = useReducedMotion()
  const { token } = Route.useSearch()
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [resetError, setResetError] = useState<string | null>(null)
  const [resetSuccess, setResetSuccess] = useState<string | null>(null)
  const confirmMutation = useMutation({
    mutationFn: api.confirmPasswordReset,
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

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setResetError(null)
    setResetSuccess(null)

    if (!token) {
      setResetError('缺少重置 token，请重新打开邮件中的完整链接')
      return
    }
    if (!newPassword.trim()) {
      setResetError('请输入新密码')
      return
    }
    if (newPassword !== confirmPassword) {
      setResetError('两次输入的新密码不一致')
      return
    }

    try {
      await confirmMutation.mutateAsync({
        token,
        new_password: newPassword,
      })
      setNewPassword('')
      setConfirmPassword('')
      setResetSuccess('密码已重置，请使用新密码重新登录。')
    } catch (submitError) {
      setResetError(
        submitError instanceof Error ? submitError.message : '重置密码失败',
      )
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
        <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-2xl">
              <KeyRoundIcon className="size-6" />
              设置新密码
            </CardTitle>
            <CardDescription>
              请输入新的登录密码并提交。
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
                  <AlertTitle>缺少重置凭证</AlertTitle>
                  <AlertDescription>
                    当前链接没有包含有效 token，请重新打开邮件中的完整链接。
                  </AlertDescription>
                </Alert>
              </motion.div>
            ) : null}

            <motion.form
              animate={{ opacity: 1, y: 0 }}
              className="space-y-4"
              initial={prefersReducedMotion ? false : { opacity: 0, y: 12 }}
              onSubmit={handleSubmit}
              transition={sectionTransition(0.22)}
            >
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="reset-new-password">新密码</FieldLabel>
                  <Input
                    autoComplete="new-password"
                    id="reset-new-password"
                    onChange={(event) => setNewPassword(event.target.value)}
                    placeholder="请输入新密码"
                    type="password"
                    value={newPassword}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="reset-confirm-password">
                    确认新密码
                  </FieldLabel>
                  <Input
                    autoComplete="new-password"
                    id="reset-confirm-password"
                    onChange={(event) => setConfirmPassword(event.target.value)}
                    placeholder="请再次输入新密码"
                    type="password"
                    value={confirmPassword}
                  />
                </Field>
              </FieldGroup>

              {resetError ? <FieldError>{resetError}</FieldError> : null}
              {resetSuccess ? (
                <Alert>
                  <CheckCircle2Icon className="size-4" />
                  <AlertTitle>密码已重置</AlertTitle>
                  <AlertDescription>{resetSuccess}</AlertDescription>
                </Alert>
              ) : null}

              <motion.div
                animate={{ opacity: 1, y: 0 }}
                className="flex flex-col gap-3 sm:flex-row"
                initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
                transition={sectionTransition(0.28)}
              >
                <Button
                  className="sm:flex-1"
                  disabled={confirmMutation.isPending || !token}
                  type="submit"
                >
                  {confirmMutation.isPending ? '提交中…' : '保存新密码'}
                </Button>
                <Button asChild className="sm:flex-1" variant="outline">
                  <Link to="/login">返回登录</Link>
                </Button>
              </motion.div>
            </motion.form>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
