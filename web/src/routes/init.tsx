import { createFileRoute, Link } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  AlertCircleIcon,
  CheckCircle2Icon,
  LoaderCircleIcon,
  ShieldCheckIcon,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { ApiError, api } from '#/lib/api'
import { initStatusQueryOptions } from '#/lib/query-options'
import { invalidateInitStatusQueries } from '#/lib/query-invalidation'
import type { InitInput } from '#/lib/types'
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
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Switch } from '#/components/ui/switch'
import { Textarea } from '#/components/ui/textarea'

export const Route = createFileRoute('/init')({
  component: InitPage,
})

const DEFAULT_SITE_NAME = 'EasyDrop'

function getDefaultSiteUrl() {
  if (typeof window !== 'undefined' && window.location.origin) {
    return window.location.origin
  }

  return ''
}

function createInitialFormState(): InitInput {
  return {
    allow_register: true,
    email: '',
    nickname: '',
    password: '',
    site_announcement: '',
    site_name: DEFAULT_SITE_NAME,
    site_url: getDefaultSiteUrl(),
    username: '',
  }
}

function InitPage() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<InitInput>(createInitialFormState)
  const initStatusQuery = useQuery(initStatusQueryOptions())
  const initializeMutation = useMutation({
    mutationFn: api.initializeSystem,
    onSuccess: async () => {
      await invalidateInitStatusQueries(queryClient)
    },
  })

  useEffect(() => {
    setForm((current) =>
      current.site_url.trim()
        ? current
        : {
            ...current,
            site_url: getDefaultSiteUrl(),
          },
    )
  }, [])

  useEffect(() => {
    if (initStatusQuery.error) {
      const message =
        initStatusQuery.error instanceof Error
          ? initStatusQuery.error.message
          : '初始化状态读取失败'
      toast.error(message)
    }
  }, [initStatusQuery.error])

  function updateField<TKey extends keyof InitInput>(
    key: TKey,
    value: InitInput[TKey],
  ) {
    setForm((current) => ({
      ...current,
      [key]: value,
    }))
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (
      !form.username.trim() ||
      !form.nickname.trim() ||
      !form.email.trim() ||
      !form.password.trim() ||
      !form.site_name.trim() ||
      !form.site_url.trim()
    ) {
      toast.error('请完整填写管理员信息与站点基础配置')
      return
    }

    try {
      await initializeMutation.mutateAsync({
        ...form,
        email: form.email.trim(),
        nickname: form.nickname.trim(),
        password: form.password,
        site_announcement: form.site_announcement.trim(),
        site_name: form.site_name.trim(),
        site_url: form.site_url.trim(),
        username: form.username.trim(),
      })

      queryClient.setQueryData(initStatusQueryOptions().queryKey, {
        initialized: true,
      })
      window.location.assign('/login')
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) {
        queryClient.setQueryData(initStatusQueryOptions().queryKey, {
          initialized: true,
        })
        toast.error('系统已经初始化，无需重复创建首个管理员')
        return
      }

      toast.error(
        error instanceof Error ? error.message : '初始化失败，请稍后重试',
      )
    }
  }

  if (initStatusQuery.isPending) {
    return (
      <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
        <Card className="w-full max-w-xl border border-border/70 bg-card/95 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <LoaderCircleIcon className="animate-spin" />
              正在检查初始化状态
            </CardTitle>
            <CardDescription>
              首次部署时需要先创建管理员账号并写入基础站点配置。
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  if (initStatusQuery.error) {
    return (
      <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
        <Card className="w-full max-w-xl border border-border/70 bg-card/95 shadow-sm">
          <CardHeader>
            <CardTitle>初始化状态读取失败</CardTitle>
            <CardDescription>
              无法确认系统是否已经初始化，请检查后端服务是否可用。
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <Alert variant="destructive">
              <AlertCircleIcon />
              <AlertTitle>读取失败</AlertTitle>
              <AlertDescription>
                {initStatusQuery.error instanceof Error
                  ? initStatusQuery.error.message
                  : '初始化状态读取失败'}
              </AlertDescription>
            </Alert>
            <div className="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <Button
                variant="outline"
                type="button"
                onClick={() => void initStatusQuery.refetch()}
              >
                重新检查
              </Button>
              <Button asChild>
                <Link search={{ content: undefined }} to="/">
                  返回首页
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (initStatusQuery.data.initialized) {
    return (
      <div className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8">
        <Card className="w-full max-w-2xl border border-border/70 bg-card/95 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle2Icon className="text-primary" />
              系统已初始化
            </CardTitle>
            <CardDescription>
              首个管理员与基础站点配置已经创建完成，这个页面仅在首次部署时使用。
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <Button asChild variant="outline">
                <Link search={{ content: undefined }} to="/">
                  返回首页
                </Link>
              </Button>
              <Button asChild>
                <Link search={{ redirect: '/' }} to="/login">
                  前往登录
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <Card className="border border-border/70 bg-card/95 shadow-sm px-3 py-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-2xl">
              <ShieldCheckIcon />
              初始化系统
            </CardTitle>
            <CardDescription>
              首次部署时，请先创建第一个管理员账号，并完成站点的基础配置。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit}>
              <FieldGroup>
                <FieldGroup>
                  <div className="text-sm font-medium text-foreground">
                    管理员信息
                  </div>

                  <Field>
                    <FieldLabel htmlFor="init-username">用户名</FieldLabel>
                    <Input
                      id="init-username"
                      onChange={(event) =>
                        updateField('username', event.target.value)
                      }
                      placeholder="用于登录后台的管理员用户名"
                      value={form.username}
                    />
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="init-nickname">昵称</FieldLabel>
                    <Input
                      id="init-nickname"
                      onChange={(event) =>
                        updateField('nickname', event.target.value)
                      }
                      placeholder="站内展示的管理员昵称"
                      value={form.nickname}
                    />
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="init-email">邮箱</FieldLabel>
                    <Input
                      id="init-email"
                      onChange={(event) =>
                        updateField('email', event.target.value)
                      }
                      placeholder="用于接收通知和找回密码"
                      type="email"
                      value={form.email}
                    />
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="init-password">密码</FieldLabel>
                    <Input
                      id="init-password"
                      onChange={(event) =>
                        updateField('password', event.target.value)
                      }
                      placeholder="请设置一个安全的管理员密码"
                      type="password"
                      value={form.password}
                    />
                  </Field>
                </FieldGroup>

                <FieldGroup>
                  <div className="text-sm font-medium text-foreground">
                    站点配置
                  </div>

                  <Field>
                    <FieldLabel htmlFor="init-site-name">站点名称</FieldLabel>
                    <Input
                      id="init-site-name"
                      onChange={(event) =>
                        updateField('site_name', event.target.value)
                      }
                      placeholder="例如 EasyDrop"
                      value={form.site_name}
                    />
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="init-site-url">站点地址</FieldLabel>
                    <Input
                      id="init-site-url"
                      onChange={(event) =>
                        updateField('site_url', event.target.value)
                      }
                      placeholder="例如 https://example.com"
                      value={form.site_url}
                    />
                    <FieldDescription>
                      请填写站点最终对外访问的完整地址，默认会使用当前访问来源。
                    </FieldDescription>
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="init-site-announcement">
                      站点公告
                    </FieldLabel>
                    <Textarea
                      id="init-site-announcement"
                      onChange={(event) =>
                        updateField('site_announcement', event.target.value)
                      }
                      placeholder="可选，首次进入站点时展示给所有访客的公告内容。"
                      rows={5}
                      value={form.site_announcement}
                    />
                  </Field>

                  <Field orientation="horizontal">
                    <Switch
                      checked={form.allow_register}
                      id="init-allow-register"
                      onCheckedChange={(checked) =>
                        updateField('allow_register', checked)
                      }
                    />
                    <FieldContent>
                      <FieldLabel htmlFor="init-allow-register">
                        <FieldTitle>开放公开注册</FieldTitle>
                      </FieldLabel>
                      <FieldDescription>
                        开启后，访客可以直接通过注册页创建普通账号；关闭后仅管理员可在后台创建用户。
                      </FieldDescription>
                    </FieldContent>
                  </Field>
                </FieldGroup>
              </FieldGroup>

              <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:justify-end">
                <Button asChild type="button" variant="outline">
                  <Link search={{ content: undefined }} to="/">
                    返回首页
                  </Link>
                </Button>
                <Button disabled={initializeMutation.isPending} type="submit">
                  {initializeMutation.isPending
                    ? '正在初始化…'
                    : '创建管理员并完成初始化'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <aside className="flex flex-col gap-4">
          <Card className="border border-border/70 bg-card/90 shadow-sm">
            <CardHeader>
              <CardTitle>初始化说明</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3 text-sm leading-7 text-muted-foreground">
              <p>这个页面只在系统首次部署时使用，用于创建第一个管理员账号。</p>
              <p>若需要修改站点配置，请在初始化完成后进入后台设置页面处理。</p>
              <p>成功后页面会跳转到登录页，请使用刚创建的管理员账号登录。</p>
            </CardContent>
          </Card>
        </aside>
      </div>
    </div>
  )
}
