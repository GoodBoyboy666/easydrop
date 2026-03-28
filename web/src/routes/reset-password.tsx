import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation, useQuery } from '@tanstack/react-query'
import { CheckCircle2Icon, KeyRoundIcon } from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import {
  CaptchaPanel,
  createEmptyCaptchaInput,
  isCaptchaComplete,
} from '#/components/site/captcha-panel'
import { captchaConfigQueryOptions } from '#/lib/query-options'
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
  const { token } = Route.useSearch()
  const [captcha, setCaptcha] = useState(createEmptyCaptchaInput)
  const [captchaResetSignal, setCaptchaResetSignal] = useState(0)
  const [email, setEmail] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [requestError, setRequestError] = useState<string | null>(null)
  const [requestSuccess, setRequestSuccess] = useState<string | null>(null)
  const [resetError, setResetError] = useState<string | null>(null)
  const [resetSuccess, setResetSuccess] = useState<string | null>(null)
  const captchaConfigQuery = useQuery(captchaConfigQueryOptions())
  const requestMutation = useMutation({
    mutationFn: api.requestPasswordReset,
  })
  const confirmMutation = useMutation({
    mutationFn: api.confirmPasswordReset,
  })

  async function handleRequestSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setRequestError(null)
    setRequestSuccess(null)

    if (!email.trim()) {
      setRequestError('请输入注册邮箱')
      return
    }

    if (!isCaptchaComplete(captchaConfigQuery.data, captcha)) {
      setRequestError('请先完成验证码')
      return
    }

    try {
      await requestMutation.mutateAsync({
        captcha: captchaConfigQuery.data?.enabled ? captcha : undefined,
        email: email.trim(),
      })
      setRequestSuccess(
        '如果该邮箱存在可重置账号，系统会向该邮箱发送重置密码邮件。',
      )
      if (captchaConfigQuery.data?.enabled) {
        setCaptcha(createEmptyCaptchaInput())
        setCaptchaResetSignal((current) => current + 1)
      }
    } catch (submitError) {
      if (captchaConfigQuery.data?.enabled) {
        setCaptcha(createEmptyCaptchaInput())
        setCaptchaResetSignal((current) => current + 1)
      }
      setRequestError(
        submitError instanceof Error ? submitError.message : '发送重置邮件失败',
      )
    }
  }

  async function handleConfirmSubmit(event: React.FormEvent<HTMLFormElement>) {
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
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <KeyRoundIcon className="size-6" />
            {token ? '设置新密码' : '找回密码'}
          </CardTitle>
          <CardDescription>
            {token
              ? '请输入新的登录密码并提交。'
              : '输入注册邮箱，系统会发送密码重置邮件。'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!token ? (
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
          ) : null}

          {token ? (
            <form className="space-y-4" onSubmit={handleConfirmSubmit}>
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

              <div className="flex flex-col gap-3 sm:flex-row">
                <Button
                  className="sm:flex-1"
                  disabled={confirmMutation.isPending}
                  type="submit"
                >
                  {confirmMutation.isPending ? '提交中…' : '保存新密码'}
                </Button>
                <Button asChild className="sm:flex-1" variant="outline">
                  <Link to="/login">返回登录</Link>
                </Button>
              </div>
            </form>
          ) : (
            <form className="space-y-4" onSubmit={handleRequestSubmit}>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="reset-email">注册邮箱</FieldLabel>
                  <Input
                    autoComplete="email"
                    id="reset-email"
                    onChange={(event) => setEmail(event.target.value)}
                    placeholder="请输入注册邮箱"
                    type="email"
                    value={email}
                  />
                </Field>
              </FieldGroup>

              {requestError ? <FieldError>{requestError}</FieldError> : null}
              {requestSuccess ? (
                <Alert>
                  <CheckCircle2Icon className="size-4" />
                  <AlertTitle>请求已提交</AlertTitle>
                  <AlertDescription>{requestSuccess}</AlertDescription>
                </Alert>
              ) : null}

              <div className="flex flex-col gap-3 sm:flex-row">
                <Button
                  className="sm:flex-1"
                  disabled={requestMutation.isPending}
                  type="submit"
                >
                  {requestMutation.isPending ? '发送中…' : '发送重置邮件'}
                </Button>
                <Button asChild className="sm:flex-1" variant="outline">
                  <Link to="/login">返回登录</Link>
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
