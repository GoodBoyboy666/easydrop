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

export const Route = createFileRoute('/forgot-password')({
  component: ForgotPasswordPage,
})

function ForgotPasswordPage() {
  const [captcha, setCaptcha] = useState(createEmptyCaptchaInput)
  const [captchaResetSignal, setCaptchaResetSignal] = useState(0)
  const [email, setEmail] = useState('')
  const [requestError, setRequestError] = useState<string | null>(null)
  const [requestSuccess, setRequestSuccess] = useState<string | null>(null)
  const captchaConfigQuery = useQuery(captchaConfigQueryOptions())
  const requestMutation = useMutation({
    mutationFn: api.requestPasswordReset,
  })

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
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

  return (
    <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
      <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <KeyRoundIcon className="size-6" />
            找回密码
          </CardTitle>
          <CardDescription>
            输入注册邮箱，系统会向该邮箱发送密码重置邮件。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
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

          <form className="space-y-4" onSubmit={handleSubmit}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="forgot-password-email">注册邮箱</FieldLabel>
                <Input
                  autoComplete="email"
                  id="forgot-password-email"
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
        </CardContent>
      </Card>
    </div>
  )
}
