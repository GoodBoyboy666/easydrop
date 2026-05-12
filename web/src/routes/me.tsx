import {
  createFileRoute,
  Outlet,
  useLocation,
} from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FingerprintIcon, Link2Icon, MailIcon, PencilIcon, ShieldCheckIcon, Trash2Icon, UserCircleIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { HTMLMotionProps, Transition } from 'motion/react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import {
  requireAuthenticatedRoute,
  useUnauthorizedHandler,
} from '#/lib/auth-guards'
import { formatDateTime, getInitials } from '#/lib/format'
import { myPasskeysQueryOptions, oAuthBindingsQueryOptions, oAuthProvidersQueryOptions, queryKeys } from '#/lib/query-options'
import { ProviderIcon } from '#/components/site/provider-icons'
import { setOAuthIntent } from './oauth.$provider'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from '#/components/ui/alert-dialog'
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
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'

export const Route = createFileRoute('/me')({
  beforeLoad: async () => {
    await requireAuthenticatedRoute()
  },
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
const FLAT_PROFILE_FORM_CARD_CLASSNAME =
  'rounded-none bg-transparent py-0 ring-0 shadow-none'

function MePage() {
  const { auth, handleUnauthorized } = useUnauthorizedHandler('/me')
  const location = useLocation()
  const prefersReducedMotion = useReducedMotion()
  const [nickname, setNickname] = useState('')

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  const [newEmail, setNewEmail] = useState('')
  const [emailPassword, setEmailPassword] = useState('')
  const [motionReady, setMotionReady] = useState(prefersReducedMotion)

  const queryClient = useQueryClient()
  const passkeyListQuery = useQuery({
    ...myPasskeysQueryOptions(),
    enabled: auth.status === 'authenticated',
  })
  const registerPasskeyMutation = useMutation({
    mutationFn: async () => {
      if (!isWebAuthnSupported()) {
        throw new Error('当前浏览器不支持通行密钥')
      }
      const { options, session_id } = await api.beginPasskeyRegistration()
      const credential = await navigator.credentials.create({
        publicKey: PublicKeyCredential.parseCreationOptionsFromJSON(options),
      })
      if (!(credential instanceof PublicKeyCredential)) {
        throw new Error('浏览器不支持或取消通行密钥注册')
      }
      return api.finishPasskeyRegistration(credential, session_id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.myPasskeys() })
      toast.success('通行密钥已添加')
    },
    onError: (error) => {
      if (handleUnauthorized(error)) {
        return
      }
      toast.error(error instanceof Error ? error.message : '添加通行密钥失败')
    },
  })
  const renamePasskeyMutation = useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) =>
      api.renamePasskey(id, { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.myPasskeys() })
      toast.success('已重命名')
    },
    onError: (error) => {
      if (handleUnauthorized(error)) {
        return
      }
      toast.error(error instanceof Error ? error.message : '重命名失败')
    },
  })
  const deletePasskeyMutation = useMutation({
    mutationFn: (id: number) => api.deletePasskey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.myPasskeys() })
      toast.success('已删除')
    },
    onError: (error) => {
      if (handleUnauthorized(error)) {
        return
      }
      toast.error(error instanceof Error ? error.message : '删除失败')
    },
  })

  const oAuthBindingsQuery = useQuery({
    ...oAuthBindingsQueryOptions(auth.status === 'authenticated'),
    enabled: auth.status === 'authenticated',
  })
  const oAuthProvidersQuery = useQuery({
    ...oAuthProvidersQueryOptions(),
  })
  const unbindOAuthMutation = useMutation({
    mutationFn: (bindId: number) => api.unbindOAuth(bindId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oAuthBindingsPrefix() })
      toast.success('已解绑社交账号')
    },
    onError: (error) => {
      if (handleUnauthorized(error)) {
        return
      }
      toast.error(error instanceof Error ? error.message : '解绑失败')
    },
  })

  const [oauthDeleteTarget, setOauthDeleteTarget] = useState<{ id: number; label: string } | null>(null)
  const [editingPasskeyID, setEditingPasskeyID] = useState<number | null>(null)
  const [editingPasskeyName, setEditingPasskeyName] = useState('')

  function startRename(id: number, currentName: string) {
    setEditingPasskeyID(id)
    setEditingPasskeyName(currentName)
  }

  function cancelRename() {
    setEditingPasskeyID(null)
    setEditingPasskeyName('')
  }

  function submitRename(id: number) {
    const name = editingPasskeyName.trim()
    if (!name || name.length > 15) {
      toast.error('名称长度应为 1-15 个字符')
      return
    }
    renamePasskeyMutation.mutate({ id, name })
    cancelRename()
  }

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

  async function handleProfileSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || !auth.user || updateProfileMutation.isPending) {
      return
    }

    const normalizedNickname = nickname.trim()
    if (normalizedNickname === auth.user.nickname) {
      toast.message('昵称没有变化')
      return
    }

    try {
      await updateProfileMutation.mutateAsync(normalizedNickname)
      await auth.refreshUser()
      toast.success('昵称已更新')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '更新昵称失败')
    }
  }

  async function handlePasswordSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || changePasswordMutation.isPending) {
      return
    }

    if (newPassword.trim() === '' || oldPassword.trim() === '') {
      toast.error('请填写旧密码和新密码')
      return
    }

    if (newPassword !== confirmPassword) {
      toast.error('两次输入的新密码不一致')
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
      toast.success('密码已修改')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '修改密码失败')
    }
  }

  async function handleEmailSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || changeEmailMutation.isPending) {
      return
    }

    const normalizedEmail = newEmail.trim()
    if (normalizedEmail === '') {
      toast.error('请输入新邮箱')
      return
    }

    if (emailPassword.trim() === '') {
      toast.error('请输入当前密码')
      return
    }

    try {
      await changeEmailMutation.mutateAsync({
        current_password: emailPassword,
        new_email: normalizedEmail,
      })
      setEmailPassword('')
      toast.success('验证邮件已发送到新邮箱，请按邮件指引完成确认')
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '提交邮箱修改失败')
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
        <Card>
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
              <div className="rounded-2xl bg-transparent px-2">
                <div className="text-sm text-muted-foreground">邮箱</div>
                <div className="mt-2 flex items-center gap-2 text-sm">
                  <MailIcon />
                  <span>{auth.user.email || '未设置邮箱'}</span>
                </div>
              </div>
              <div className="rounded-2xl bg-transparent px-2">
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
                <Card className={FLAT_PROFILE_FORM_CARD_CLASSNAME}>
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
                <Card className={FLAT_PROFILE_FORM_CARD_CLASSNAME}>
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
              <Card className={FLAT_PROFILE_FORM_CARD_CLASSNAME}>
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
                    <Button disabled={changeEmailMutation.isPending} type="submit">
                      {changeEmailMutation.isPending ? '提交中…' : '发送验证邮件'}
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </motion.div>

            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.24)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className={FLAT_PROFILE_FORM_CARD_CLASSNAME}>
                <CardHeader>
                  <CardTitle className="text-base">通行密钥</CardTitle>
                  <CardDescription>
                    使用指纹、面容或安全密钥快速登录，无需输入密码。
                    最多可添加 10 个通行密钥。
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {passkeyListQuery.data && passkeyListQuery.data.length > 0 ? (
                    <div className="space-y-2">
                      {passkeyListQuery.data.map((pk) => (
                        <div
                          className="flex items-center gap-3 rounded-lg border border-border/70 bg-muted/20 px-3 py-2"
                          key={pk.id}
                        >
                          <FingerprintIcon className="size-4 text-muted-foreground" />
                          <div className="min-w-0 flex-1">
                            {editingPasskeyID === pk.id ? (
                              <form
                                className="flex items-center gap-2"
                                onSubmit={(e) => {
                                  e.preventDefault()
                                  submitRename(pk.id)
                                }}
                              >
                                <Input
                                  autoFocus
                                  className="h-7 text-sm"
                                  maxLength={15}
                                  onChange={(e) =>
                                    setEditingPasskeyName(e.target.value)
                                  }
                                  value={editingPasskeyName}
                                />
                                <Button
                                  disabled={renamePasskeyMutation.isPending}
                                  size="sm"
                                  type="submit"
                                  variant="ghost"
                                >
                                  保存
                                </Button>
                                <Button
                                  onClick={cancelRename}
                                  size="sm"
                                  type="button"
                                  variant="ghost"
                                >
                                  取消
                                </Button>
                              </form>
                            ) : (
                              <>
                                <div className="text-sm font-medium">
                                  {pk.name}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                  {formatDateTime(pk.created_at)}
                                </div>
                              </>
                            )}
                          </div>
                          {editingPasskeyID !== pk.id && (
                            <div className="flex items-center gap-1">
                              <Button
                                aria-label="重命名"
                                onClick={() => startRename(pk.id, pk.name)}
                                size="icon-sm"
                                variant="ghost"
                              >
                                <PencilIcon />
                              </Button>
                              <Button
                                aria-label="删除"
                                disabled={deletePasskeyMutation.isPending}
                                onClick={() => {
                                  if (
                                    window.confirm(
                                      `确定要删除通行密钥「${pk.name}」吗？`,
                                    )
                                  ) {
                                    deletePasskeyMutation.mutate(pk.id)
                                  }
                                }}
                                size="icon-sm"
                                variant="ghost"
                              >
                                <Trash2Icon />
                              </Button>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-muted-foreground">
                      尚未添加通行密钥
                    </div>
                  )}
                  <div className="mt-4">
                    {isWebAuthnSupported() ? (
                      <Button
                        disabled={
                          registerPasskeyMutation.isPending ||
                          (passkeyListQuery.data &&
                            passkeyListQuery.data.length >= 10)
                        }
                        onClick={() => registerPasskeyMutation.mutate()}
                        size="sm"
                        variant="outline"
                      >
                      <FingerprintIcon data-icon="inline-start" />
                      {registerPasskeyMutation.isPending
                        ? '正在添加…'
                        : '添加通行密钥'}
                    </Button>
                    ) : (
                      <div className="text-sm text-muted-foreground">
                        当前浏览器不支持通行密钥
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            <motion.div
              animate={motionReady ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
              initial={prefersReducedMotion ? false : SECTION_ENTER_INITIAL}
              transition={getEntranceTransition(0.28)}
              {...GPU_ACCELERATED_MOTION_PROPS}
            >
              <Card className={FLAT_PROFILE_FORM_CARD_CLASSNAME}>
                <CardHeader>
                  <CardTitle className="text-base">社交账号绑定</CardTitle>
                  <CardDescription>
                    绑定社交平台账号后可通过对应平台快捷登录。
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {oAuthBindingsQuery.data && oAuthBindingsQuery.data.length > 0 ? (
                    <div className="space-y-2">
                      {oAuthBindingsQuery.data.map((bind) => (
                        <div
                          className="flex items-center gap-3 rounded-lg border border-border/70 bg-muted/20 px-3 py-2"
                          key={bind.id}
                        >
                          <Link2Icon className="size-4 text-muted-foreground" />
                          <div className="min-w-0 flex-1">
                            <div className="text-sm font-medium">
                              {PROVIDER_LABELS[bind.provider] ?? bind.provider}
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {bind.provider_email}
                            </div>
                          </div>
                          <Button
                            aria-label="解绑"
                            disabled={unbindOAuthMutation.isPending}
                            onClick={() =>
                              setOauthDeleteTarget({
                                id: bind.id,
                                label: PROVIDER_LABELS[bind.provider] ?? bind.provider,
                              })
                            }
                            size="icon-sm"
                            variant="ghost"
                          >
                            <Trash2Icon />
                          </Button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-muted-foreground">
                      尚未绑定社交账号
                    </div>
                  )}
                  {oAuthProvidersQuery.data?.providers && oAuthProvidersQuery.data.providers.length > 0 ? (
                    <div className="mt-4 flex flex-wrap gap-2">
                      {oAuthProvidersQuery.data.providers
                        .filter(
                          (p) =>
                            !oAuthBindingsQuery.data?.some(
                              (b) => b.provider === p.provider,
                            ),
                        )
                        .map((p) => (
                          <Button
                            key={p.provider}
                            onClick={() => {
                              setOAuthIntent('bind')
                              window.location.href = `/api/v1/auth/oauth/${p.provider}`
                            }}
                            size="sm"
                            variant="outline"
                          >
                            <ProviderIcon className="size-4" provider={p.provider} />
                            绑定 {PROVIDER_LABELS[p.provider] ?? p.provider}
                          </Button>
                        ))}
                    </div>
                  ) : null}
                </CardContent>
              </Card>
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>

      <AlertDialog open={oauthDeleteTarget !== null} onOpenChange={(open) => { if (!open) setOauthDeleteTarget(null) }}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogMedia className="bg-background">
              <Trash2Icon />
            </AlertDialogMedia>
            <AlertDialogTitle>解绑社交账号</AlertDialogTitle>
            <AlertDialogDescription>
              确定要解绑{oauthDeleteTarget?.label}账号吗？解绑后你将无法通过该账号快捷登录。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={unbindOAuthMutation.isPending}>取消</AlertDialogCancel>
            <AlertDialogAction
              disabled={unbindOAuthMutation.isPending}
              onClick={() => {
                if (oauthDeleteTarget) {
                  unbindOAuthMutation.mutate(oauthDeleteTarget.id)
                }
                setOauthDeleteTarget(null)
              }}
              variant="destructive"
            >
              {unbindOAuthMutation.isPending ? '解绑中…' : '确定解绑'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </motion.div>
  )
}

function isWebAuthnSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.PublicKeyCredential === 'function' &&
    typeof PublicKeyCredential.parseCreationOptionsFromJSON === 'function' &&
    typeof PublicKeyCredential.parseRequestOptionsFromJSON === 'function'
  )
}

const PROVIDER_LABELS: Record<string, string> = {
  google: 'Google',
  github: 'GitHub',
  twitter: 'X',
  microsoft: 'Microsoft',
  apple: 'Apple',
}

