import {
  createFileRoute,
  Outlet,
  useLocation,
} from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { MailIcon, ShieldCheckIcon, UserCircleIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { HTMLMotionProps, Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { api, ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { formatDateTime, getInitials } from '#/lib/format'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import { Field, FieldError, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'

export const Route = createFileRoute('/me')({
  component: MePage,
})

const MOTION_DELAY_SECONDS = 0.1
const gpuTransformTemplate: NonNullable<
  HTMLMotionProps<'div'>['transformTemplate']
> = (_, generatedTransform) =>
  generatedTransform ? `${generatedTransform} translateZ(0)` : 'translateZ(0)'
const GPU_ACCELERATED_MOTION_PROPS = {
  style: { willChange: 'transform, opacity' },
  transformTemplate: gpuTransformTemplate,
} as const
const PAGE_ENTER_INITIAL = { opacity: 0, y: 10 }
const SECTION_ENTER_INITIAL = { opacity: 0, y: 12 }

function MePage() {
  const auth = useAuth()
  const location = useLocation()
  const prefersReducedMotion = useReducedMotion()
  const [nickname, setNickname] = useState('')
  const [nicknameError, setNicknameError] = useState<string | null>(null)
  const [nicknameSuccess, setNicknameSuccess] = useState<string | null>(null)

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null)

  const [newEmail, setNewEmail] = useState('')
  const [emailPassword, setEmailPassword] = useState('')
  const [emailError, setEmailError] = useState<string | null>(null)
  const [emailSuccess, setEmailSuccess] = useState<string | null>(null)
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)

  const updateProfileMutation = useMutation({
    mutationFn: (nextNickname: string) =>
      api.updateMyProfile({ nickname: nextNickname }),
  })
  const changePasswordMutation = useMutation({
    mutationFn: (input: { old_password: string; new_password: string }) =>
      api.changeMyPassword(input),
  })
  const changeEmailMutation = useMutation({
    mutationFn: (input: { current_password: string; new_email: string }) =>
      api.requestMyEmailChange(input),
  })

  useEffect(() => {
    if (!auth.user) {
      return
    }

    setNickname(auth.user.nickname)
  }, [auth.user])

  useEffect(() => {
    if (prefersReducedMotion) {
      setMotionReady(true)
      return
    }

    setMotionReady(false)
    let raf1 = 0
    let raf2 = 0

    raf1 = window.requestAnimationFrame(() => {
      raf2 = window.requestAnimationFrame(() => {
        setMotionReady(true)
      })
    })

    return () => {
      window.cancelAnimationFrame(raf1)
      window.cancelAnimationFrame(raf2)
    }
  }, [prefersReducedMotion])

  const getEntranceTransition = (delay = 0): Transition =>
    prefersReducedMotion
      ? { duration: 0 }
      : {
          type: 'spring',
          duration: 0.32,
          ease: 'easeOut' as const,
          delay: delay + MOTION_DELAY_SECONDS,
        }

  function redirectToLogin() {
    window.location.assign('/login?redirect=/me')
  }

  function handleUnauthorized(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      auth.logout()
      redirectToLogin()
      return true
    }

    return false
  }

  async function handleProfileSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || !auth.user || updateProfileMutation.isPending) {
      return
    }

    setNicknameError(null)
    setNicknameSuccess(null)

    const normalizedNickname = nickname.trim()
    if (normalizedNickname === auth.user.nickname) {
      setNicknameSuccess('昵称没有变化')
      return
    }

    try {
      await updateProfileMutation.mutateAsync(normalizedNickname)
      await auth.refreshUser()
      setNicknameSuccess('昵称已更新')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setNicknameError(error instanceof Error ? error.message : '更新昵称失败')
    }
  }

  async function handlePasswordSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || changePasswordMutation.isPending) {
      return
    }

    setPasswordError(null)
    setPasswordSuccess(null)

    if (newPassword.trim() === '' || oldPassword.trim() === '') {
      setPasswordError('请填写旧密码和新密码')
      return
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('两次输入的新密码不一致')
      return
    }

    try {
      await changePasswordMutation.mutateAsync({
        old_password: oldPassword,
        new_password: newPassword,
      })
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setPasswordSuccess('密码已修改')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setPasswordError(error instanceof Error ? error.message : '修改密码失败')
    }
  }

  async function handleEmailSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || changeEmailMutation.isPending) {
      return
    }

    setEmailError(null)
    setEmailSuccess(null)

    const normalizedEmail = newEmail.trim()
    if (normalizedEmail === '') {
      setEmailError('请输入新邮箱')
      return
    }

    if (emailPassword.trim() === '') {
      setEmailError('请输入当前密码')
      return
    }

    try {
      await changeEmailMutation.mutateAsync({
        current_password: emailPassword,
        new_email: normalizedEmail,
      })
      setEmailPassword('')
      setEmailSuccess('验证邮件已发送到新邮箱，请按邮件指引完成确认')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      setEmailError(error instanceof Error ? error.message : '提交邮箱修改失败')
    }
  }

  if (location.pathname !== '/me') {
    return <Outlet />
  }

  if (auth.status !== 'authenticated' || !auth.user) {
    return (
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8"
        initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
        transition={getEntranceTransition()}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <Alert>
          <AlertTitle>需要先登录</AlertTitle>
          <AlertDescription>
            当前页面用于展示个人信息，请先登录后再查看。
          </AlertDescription>
        </Alert>
      </motion.div>
    )
  }

  return (
    <motion.div
      animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
      className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8"
      initial={prefersReducedMotion ? false : PAGE_ENTER_INITIAL}
      transition={getEntranceTransition()}
      {...GPU_ACCELERATED_MOTION_PROPS}
    >
      <motion.div
        animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
        initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
        transition={getEntranceTransition(0.04)}
        {...GPU_ACCELERATED_MOTION_PROPS}
      >
        <Card className="border border-border/70 bg-card/50 shadow-sm">
          <CardHeader>
            <CardTitle>个人信息</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              className="flex flex-col gap-4 rounded-2xl border border-border/70 bg-muted/30 p-4 sm:flex-row sm:items-center"
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.08)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Avatar size="lg">
                <AvatarImage alt={auth.user.nickname} src={auth.user.avatar} />
                <AvatarFallback>{getInitials(auth.user.nickname)}</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="text-lg font-semibold">{auth.user.nickname}</div>
                <div className="text-sm text-muted-foreground">
                  @{auth.user.username}
                </div>
              </div>
              <div className="flex flex-wrap gap-2 sm:ml-auto">
                {auth.user.admin ? (
                  <Badge variant="secondary">
                    <ShieldCheckIcon data-icon="inline-start" />
                    管理员
                  </Badge>
                ) : (
                  <Badge variant="outline">
                    <UserCircleIcon data-icon="inline-start" />
                    普通用户
                  </Badge>
                )}
              </div>
            </motion.div>

            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              className="grid gap-4 sm:grid-cols-2"
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.1)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <div className="rounded-2xl bg-transparent p-4">
                <div className="text-sm text-muted-foreground">邮箱</div>
                <div className="mt-2 flex items-center gap-2 text-sm">
                  <MailIcon />
                  <span>{auth.user.email || '未设置邮箱'}</span>
                </div>
              </div>
              <div className="rounded-2xl bg-transparent p-4">
                <div className="text-sm text-muted-foreground">注册时间</div>
                <div className="mt-2 text-sm">
                  {formatDateTime(auth.user.created_at)}
                </div>
              </div>
            </motion.div>

            <Separator />

            <div className="grid gap-4 lg:grid-cols-2">
              <motion.div
                animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
                initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
                transition={getEntranceTransition(0.12)}
                {...GPU_ACCELERATED_MOTION_PROPS}
              >
                <Card className='bg-card/50'>
                  <CardHeader>
                    <CardTitle className="text-base">修改昵称</CardTitle>
                    <CardDescription>
                      你的昵称将用于站内展示，留空时将使用用户名。
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <form className="space-y-4" onSubmit={handleProfileSubmit}>
                      <FieldGroup>
                        <Field>
                          <FieldLabel htmlFor="me-nickname">昵称</FieldLabel>
                          <Input
                            id="me-nickname"
                            onChange={(event) => setNickname(event.target.value)}
                            placeholder="请输入新的昵称"
                            value={nickname}
                          />
                        </Field>
                      </FieldGroup>
                      {nicknameError ? <FieldError>{nicknameError}</FieldError> : null}
                      {nicknameSuccess ? (
                        <Alert>
                          <AlertDescription>{nicknameSuccess}</AlertDescription>
                        </Alert>
                      ) : null}
                      <Button disabled={updateProfileMutation.isPending} type="submit">
                        {updateProfileMutation.isPending ? '保存中…' : '保存昵称'}
                      </Button>
                    </form>
                  </CardContent>
                </Card>
              </motion.div>

              <motion.div
                animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
                initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
                transition={getEntranceTransition(0.16)}
                {...GPU_ACCELERATED_MOTION_PROPS}
              >
                <Card className='bg-card/50'>
                  <CardHeader>
                    <CardTitle className="text-base">修改密码</CardTitle>
                    <CardDescription>
                      请输入当前密码并设置新密码，密码需要满足系统安全规则。
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <form className="space-y-4" onSubmit={handlePasswordSubmit}>
                      <FieldGroup>
                        <Field>
                          <FieldLabel htmlFor="me-old-password">当前密码</FieldLabel>
                          <Input
                            autoComplete="current-password"
                            id="me-old-password"
                            onChange={(event) => setOldPassword(event.target.value)}
                            placeholder="请输入当前密码"
                            type="password"
                            value={oldPassword}
                          />
                        </Field>
                        <Field>
                          <FieldLabel htmlFor="me-new-password">新密码</FieldLabel>
                          <Input
                            autoComplete="new-password"
                            id="me-new-password"
                            onChange={(event) => setNewPassword(event.target.value)}
                            placeholder="请输入新密码"
                            type="password"
                            value={newPassword}
                          />
                        </Field>
                        <Field>
                          <FieldLabel htmlFor="me-confirm-password">
                            确认新密码
                          </FieldLabel>
                          <Input
                            autoComplete="new-password"
                            id="me-confirm-password"
                            onChange={(event) =>
                              setConfirmPassword(event.target.value)
                            }
                            placeholder="请再次输入新密码"
                            type="password"
                            value={confirmPassword}
                          />
                        </Field>
                      </FieldGroup>
                      {passwordError ? <FieldError>{passwordError}</FieldError> : null}
                      {passwordSuccess ? (
                        <Alert>
                          <AlertDescription>{passwordSuccess}</AlertDescription>
                        </Alert>
                      ) : null}
                      <Button disabled={changePasswordMutation.isPending} type="submit">
                        {changePasswordMutation.isPending ? '提交中…' : '修改密码'}
                      </Button>
                    </form>
                  </CardContent>
                </Card>
              </motion.div>
            </div>

            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.2)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className='bg-card/50'>
                <CardHeader>
                  <CardTitle className="text-base">修改邮箱</CardTitle>
                  <CardDescription>
                    提交后会向新邮箱发送确认邮件，完成验证后将更新邮箱。
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <form className="space-y-4" onSubmit={handleEmailSubmit}>
                    <FieldGroup>
                      <Field>
                        <FieldLabel htmlFor="me-new-email">新邮箱</FieldLabel>
                        <Input
                          autoComplete="email"
                          id="me-new-email"
                          onChange={(event) => setNewEmail(event.target.value)}
                          placeholder={auth.user.email || '请输入新邮箱'}
                          type="email"
                          value={newEmail}
                        />
                      </Field>
                      <Field>
                        <FieldLabel htmlFor="me-email-password">当前密码</FieldLabel>
                        <Input
                          autoComplete="current-password"
                          id="me-email-password"
                          onChange={(event) => setEmailPassword(event.target.value)}
                          placeholder="用于确认身份"
                          type="password"
                          value={emailPassword}
                        />
                      </Field>
                    </FieldGroup>
                    {emailError ? <FieldError>{emailError}</FieldError> : null}
                    {emailSuccess ? (
                      <Alert>
                        <AlertDescription>{emailSuccess}</AlertDescription>
                      </Alert>
                    ) : null}
                    <Button disabled={changeEmailMutation.isPending} type="submit">
                      {changeEmailMutation.isPending ? '提交中…' : '发送验证邮件'}
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
