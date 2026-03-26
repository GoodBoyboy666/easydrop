import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ImageUpIcon, PencilLineIcon, PlusIcon, Trash2Icon } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import {
  ADMIN_PAGE_SIZE,
  ADMIN_SELECT_CLASSNAME,
  formatBytes,
  formatUserStatus,
} from '#/lib/admin'
import { api } from '#/lib/api'
import { formatDateTime, getInitials } from '#/lib/format'
import { adminUsersQueryOptions, queryKeys } from '#/lib/query-options'
import type { CreateUserInput, UpdateUserInput, UserDTO } from '#/lib/types'
import {
  AdminDangerDialog,
  AdminEmptyState,
  AdminErrorAlert,
  AdminListSkeleton,
  AdminMotionItem,
  AdminPageHeader,
  AdminPagination,
  AdminSection,
} from '#/components/admin/admin-ui'
import { useAdminSession } from '#/components/admin/use-admin-session'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'
import { Switch } from '#/components/ui/switch'

const USER_ORDER_OPTIONS = [
  { label: '最新注册', value: 'created_at_desc' },
  { label: '最早注册', value: 'created_at_asc' },
  { label: '用户名 A-Z', value: 'username_asc' },
  { label: '用户名 Z-A', value: 'username_desc' },
  { label: '状态升序', value: 'status_asc' },
  { label: '状态降序', value: 'status_desc' },
] as const

interface UserFilterState {
  email: string
  order: string
  status: string
  username: string
}

interface UserFormState {
  admin: boolean
  email: string
  emailVerified: boolean
  nickname: string
  password: string
  status: string
  storageQuota: string
  username: string
}

type UserEditorMode = 'create' | 'edit'

const EMPTY_FILTERS: UserFilterState = {
  email: '',
  order: 'created_at_desc',
  status: 'all',
  username: '',
}

const EMPTY_FORM: UserFormState = {
  admin: false,
  email: '',
  emailVerified: false,
  nickname: '',
  password: '',
  status: '1',
  storageQuota: '',
  username: '',
}

function toUserFormState(user: UserDTO): UserFormState {
  return {
    admin: !!user.admin,
    email: user.email ?? '',
    emailVerified: !!user.email_verified,
    nickname: user.nickname,
    password: '',
    status: String(user.status ?? 1),
    storageQuota:
      typeof user.storage_quota === 'number' ? String(user.storage_quota) : '',
    username: user.username,
  }
}

function toUserPayloadBase(formState: UserFormState) {
  return {
    admin: formState.admin,
    email: formState.email.trim(),
    email_verified: formState.emailVerified,
    nickname: formState.nickname.trim() || undefined,
    status: Number(formState.status),
    storage_quota:
      formState.storageQuota.trim() === ''
        ? null
        : Number(formState.storageQuota),
    username: formState.username.trim(),
  }
}

function formatUserQuota(quota?: number | null) {
  if (quota === null) {
    return '系统默认配额'
  }

  return formatBytes(quota)
}

export function AdminUsersPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/users')
  const [draftFilters, setDraftFilters] =
    useState<UserFilterState>(EMPTY_FILTERS)
  const [filters, setFilters] = useState<UserFilterState>(EMPTY_FILTERS)
  const [page, setPage] = useState(0)
  const [editorMode, setEditorMode] = useState<UserEditorMode | null>(null)
  const [editingUser, setEditingUser] = useState<UserDTO | null>(null)
  const [formState, setFormState] = useState<UserFormState>(EMPTY_FORM)
  const [avatarFile, setAvatarFile] = useState<File | null>(null)
  const [pendingDelete, setPendingDelete] = useState<UserDTO | null>(null)

  const usersQuery = useQuery({
    ...adminUsersQueryOptions(auth.token ?? '', {
      email: filters.email.trim() || undefined,
      limit: ADMIN_PAGE_SIZE,
      offset: page * ADMIN_PAGE_SIZE,
      order: filters.order,
      status: filters.status === 'all' ? undefined : Number(filters.status),
      username: filters.username.trim() || undefined,
    }),
    enabled: !!auth.token,
  })

  const createMutation = useMutation({
    mutationFn: (input: CreateUserInput) =>
      api.createAdminUser(input, auth.token!),
  })
  const updateMutation = useMutation({
    mutationFn: (input: UpdateUserInput) =>
      api.updateAdminUser(editingUser!.id, input, auth.token!),
  })
  const deleteMutation = useMutation({
    mutationFn: (user: UserDTO) => api.deleteAdminUser(user.id, auth.token!),
  })
  const uploadAvatarMutation = useMutation({
    mutationFn: (file: File) =>
      api.uploadAdminUserAvatar(editingUser!.id, file, auth.token!),
  })
  const deleteAvatarMutation = useMutation({
    mutationFn: () => api.deleteAdminUserAvatar(editingUser!.id, auth.token!),
  })

  const users = usersQuery.data?.items ?? []
  const isCreating = editorMode === 'create'
  const editorBusy = createMutation.isPending || updateMutation.isPending

  function resetEditor() {
    setEditorMode(null)
    setEditingUser(null)
    setFormState(EMPTY_FORM)
    setAvatarFile(null)
  }

  function startEdit(user: UserDTO) {
    setEditorMode('edit')
    setEditingUser(user)
    setFormState(toUserFormState(user))
    setAvatarFile(null)
  }

  function startCreate() {
    setEditorMode('create')
    setEditingUser(null)
    setFormState(EMPTY_FORM)
    setAvatarFile(null)
  }

  async function refreshUserQueries(targetUserId?: number) {
    await queryClient.invalidateQueries({
      queryKey: queryKeys.adminUsersPrefix(auth.token),
    })

    if (auth.token) {
      await queryClient.invalidateQueries({
        queryKey: queryKeys.currentUser(auth.token),
      })

      if (targetUserId && targetUserId === auth.user?.id) {
        try {
          await auth.refreshUser()
        } catch (error) {
          handleUnauthorized(error)
        }
      }
    }
  }

  async function handleCreateSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!auth.token) {
      return
    }

    const payloadBase = toUserPayloadBase(formState)

    try {
      if (formState.password.trim() === '') {
        toast.error('创建用户时必须填写密码')
        return
      }

      const createdUser = await createMutation.mutateAsync({
        ...payloadBase,
        password: formState.password,
      })
      toast.success(`用户 @${createdUser.username} 已创建`)
      setEditorMode('edit')
      setEditingUser(createdUser)
      setFormState(toUserFormState(createdUser))
      await refreshUserQueries(createdUser.id)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '创建用户失败')
    }
  }

  async function handleEditorSubmit(event: React.FormEvent<HTMLFormElement>) {
    if (isCreating) {
      await handleCreateSubmit(event)
      return
    }

    await handleUpdateSubmit(event)
  }

  async function handleUpdateSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!auth.token || !editingUser) {
      return
    }

    const payloadBase = toUserPayloadBase(formState)

    try {
      const updatePayload: UpdateUserInput = {
        ...payloadBase,
      }

      if (formState.storageQuota.trim() === '') {
        updatePayload.use_default_storage_quota = true
      }

      if (formState.password.trim() !== '') {
        updatePayload.password = formState.password
      }

      const updatedUser = await updateMutation.mutateAsync(updatePayload)
      toast.success(`用户 @${updatedUser.username} 已更新`)
      await refreshUserQueries(updatedUser.id)
      resetEditor()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '保存用户失败')
    }
  }

  async function handleUploadAvatar() {
    if (!auth.token || !editingUser || !avatarFile) {
      return
    }

    try {
      const updatedUser = await uploadAvatarMutation.mutateAsync(avatarFile)
      toast.success('头像已上传')
      setEditingUser(updatedUser)
      setFormState(toUserFormState(updatedUser))
      setAvatarFile(null)
      await refreshUserQueries(updatedUser.id)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '上传头像失败')
    }
  }

  async function handleDeleteAvatar() {
    if (!auth.token || !editingUser) {
      return
    }

    try {
      await deleteAvatarMutation.mutateAsync()
      toast.success('头像已删除')
      const nextUser = {
        ...editingUser,
        avatar: undefined,
      }
      setEditingUser(nextUser)
      setFormState(toUserFormState(nextUser))
      await refreshUserQueries(editingUser.id)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '删除头像失败')
    }
  }

  async function handleDeleteUser() {
    if (!auth.token || !pendingDelete) {
      return
    }

    try {
      await deleteMutation.mutateAsync(pendingDelete)
      toast.success(`用户 @${pendingDelete.username} 已删除`)
      if (editingUser?.id === pendingDelete.id) {
        resetEditor()
      }
      const deletedSelf = pendingDelete.id === auth.user?.id
      await refreshUserQueries(pendingDelete.id)
      setPendingDelete(null)

      if (deletedSelf) {
        auth.logout()
        window.location.assign('/')
      }
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '删除用户失败')
      setPendingDelete(null)
    }
  }

  return (
    <div className="space-y-6">
      <AdminDangerDialog
        busy={deleteMutation.isPending}
        confirmLabel="确认删除"
        description="删除后该用户将无法继续正常使用站点，此操作不可撤销。"
        onConfirm={() => void handleDeleteUser()}
        onOpenChange={(open) => {
          if (!open && !deleteMutation.isPending) {
            setPendingDelete(null)
          }
        }}
        open={pendingDelete !== null}
        title="删除这个用户？"
      />

      <AdminPageHeader
        title="用户管理"
        actions={
          <Button onClick={startCreate} type="button" variant="default">
            <PlusIcon data-icon="inline-start" />
            新建用户
          </Button>
        }
      />

      <AdminSection title="筛选用户">
        <form
          className="grid gap-4 lg:grid-cols-4"
          onSubmit={(event) => {
            event.preventDefault()
            setFilters(draftFilters)
            setPage(0)
          }}
        >
          <Field>
            <FieldLabel htmlFor="admin-users-username">用户名</FieldLabel>
            <Input
              id="admin-users-username"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  username: event.target.value,
                }))
              }
              placeholder="模糊搜索用户名"
              value={draftFilters.username}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-users-email">邮箱</FieldLabel>
            <Input
              id="admin-users-email"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  email: event.target.value,
                }))
              }
              placeholder="模糊搜索邮箱"
              value={draftFilters.email}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-users-status">状态</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-users-status"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  status: event.target.value,
                }))
              }
              value={draftFilters.status}
            >
              <option value="all">全部状态</option>
              <option value="1">正常</option>
              <option value="2">封禁</option>
              <option value="3">停用</option>
            </select>
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-users-order">排序</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-users-order"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  order: event.target.value,
                }))
              }
              value={draftFilters.order}
            >
              {USER_ORDER_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </Field>

          <div className="flex flex-wrap gap-2 lg:col-span-4">
            <Button type="submit">应用筛选</Button>
            <Button
              onClick={() => {
                setDraftFilters(EMPTY_FILTERS)
                setFilters(EMPTY_FILTERS)
                setPage(0)
              }}
              type="button"
              variant="outline"
            >
              重置
            </Button>
          </div>
        </form>
      </AdminSection>

      {editorMode ? (
        <AdminSection title={isCreating ? '新建用户' : `编辑用户`}>
          <form
            className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]"
            onSubmit={handleEditorSubmit}
          >
            <div className="space-y-4">
              <FieldGroup className="grid gap-4 md:grid-cols-2">
                <Field>
                  <FieldLabel htmlFor="admin-user-form-username">
                    用户名
                  </FieldLabel>
                  <Input
                    id="admin-user-form-username"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        username: event.target.value,
                      }))
                    }
                    placeholder="登录用户名"
                    value={formState.username}
                  />
                </Field>

                <Field>
                  <FieldLabel htmlFor="admin-user-form-nickname">
                    昵称
                  </FieldLabel>
                  <Input
                    id="admin-user-form-nickname"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        nickname: event.target.value,
                      }))
                    }
                    placeholder="站内显示昵称"
                    value={formState.nickname}
                  />
                </Field>

                <Field>
                  <FieldLabel htmlFor="admin-user-form-email">邮箱</FieldLabel>
                  <Input
                    id="admin-user-form-email"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        email: event.target.value,
                      }))
                    }
                    placeholder="example@example.com"
                    type="email"
                    value={formState.email}
                  />
                </Field>

                <Field>
                  <FieldLabel htmlFor="admin-user-form-password">
                    {isCreating ? '密码' : '新密码（留空不改）'}
                  </FieldLabel>
                  <Input
                    id="admin-user-form-password"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        password: event.target.value,
                      }))
                    }
                    placeholder={
                      isCreating ? '设置登录密码' : '如需重置密码再填写'
                    }
                    type="password"
                    value={formState.password}
                  />
                </Field>

                <Field>
                  <FieldLabel htmlFor="admin-user-form-status">状态</FieldLabel>
                  <select
                    className={ADMIN_SELECT_CLASSNAME}
                    id="admin-user-form-status"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        status: event.target.value,
                      }))
                    }
                    value={formState.status}
                  >
                    <option value="1">正常</option>
                    <option value="2">封禁</option>
                    <option value="3">停用</option>
                  </select>
                </Field>

                <Field>
                  <FieldLabel htmlFor="admin-user-form-storage">
                    存储配额（字节）
                  </FieldLabel>
                  <Input
                    id="admin-user-form-storage"
                    onChange={(event) =>
                      setFormState((current) => ({
                        ...current,
                        storageQuota: event.target.value,
                      }))
                    }
                    placeholder="留空表示使用默认配额"
                    type="number"
                    value={formState.storageQuota}
                  />
                </Field>
              </FieldGroup>

              <div className="grid gap-4 sm:grid-cols-2">
                <label className="flex items-center justify-between gap-3 py-1">
                  <div>
                    <div className="font-medium">管理员</div>
                  </div>
                  <Switch
                    checked={formState.admin}
                    onCheckedChange={(checked) =>
                      setFormState((current) => ({
                        ...current,
                        admin: checked,
                      }))
                    }
                  />
                </label>

                <label className="flex items-center justify-between gap-3 py-1">
                  <div>
                    <div className="font-medium">邮箱已验证</div>
                  </div>
                  <Switch
                    checked={formState.emailVerified}
                    onCheckedChange={(checked) =>
                      setFormState((current) => ({
                        ...current,
                        emailVerified: checked,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="flex flex-wrap gap-2">
                <Button disabled={editorBusy} type="submit">
                  {isCreating
                    ? createMutation.isPending
                      ? '创建中…'
                      : '创建用户'
                    : updateMutation.isPending
                      ? '保存中…'
                      : '保存用户'}
                </Button>
                <Button onClick={resetEditor} type="button" variant="outline">
                  {isCreating ? '取消创建' : '结束编辑'}
                </Button>
              </div>
            </div>

            {editingUser ? (
              <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                <div className="mb-3 flex items-center gap-3">
                  <Avatar size="lg">
                    <AvatarImage
                      alt={formState.nickname || formState.username}
                      src={editingUser.avatar}
                    />
                    <AvatarFallback>
                      {getInitials(
                        formState.nickname || formState.username || '用户',
                      )}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">
                      {formState.nickname || formState.username || '未命名用户'}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      用户 #{editingUser.id}
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <Field>
                    <FieldLabel htmlFor="admin-user-avatar">
                      上传头像
                    </FieldLabel>
                    <Input
                      id="admin-user-avatar"
                      onChange={(event) =>
                        setAvatarFile(event.target.files?.[0] ?? null)
                      }
                      type="file"
                    />
                  </Field>
                  <div className="flex flex-wrap gap-2">
                    <Button
                      disabled={!avatarFile || uploadAvatarMutation.isPending}
                      onClick={() => void handleUploadAvatar()}
                      type="button"
                      variant="outline"
                    >
                      <ImageUpIcon data-icon="inline-start" />
                      {uploadAvatarMutation.isPending ? '上传中…' : '上传头像'}
                    </Button>
                    <Button
                      disabled={
                        !editingUser.avatar || deleteAvatarMutation.isPending
                      }
                      onClick={() => void handleDeleteAvatar()}
                      type="button"
                      variant="destructive"
                    >
                      <Trash2Icon data-icon="inline-start" />
                      {deleteAvatarMutation.isPending ? '删除中…' : '删除头像'}
                    </Button>
                  </div>
                </div>
              </div>
            ) : null}
          </form>
        </AdminSection>
      ) : null}

      <AdminSection title="用户列表">
        {usersQuery.isPending ? <AdminListSkeleton /> : null}
        {usersQuery.error instanceof Error ? (
          <AdminErrorAlert
            description={usersQuery.error.message}
            title="用户列表读取失败"
          />
        ) : null}
        {!usersQuery.isPending && !usersQuery.error && users.length === 0 ? (
          <AdminEmptyState
            description="可以调整筛选条件，或者直接在上方创建一个新用户。"
            title="没有找到符合条件的用户"
          />
        ) : null}

        {!usersQuery.isPending && !usersQuery.error && users.length > 0 ? (
          <>
            <div className="overflow-hidden bg-transparent">
              {users.map((user, index) => (
                <div key={user.id}>
                  <AdminMotionItem className="p-4" delay={index * 0.03}>
                    <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                      <div className="flex min-w-0 items-center gap-3">
                        <Avatar size="lg">
                          <AvatarImage alt={user.nickname} src={user.avatar} />
                          <AvatarFallback>
                            {getInitials(user.nickname)}
                          </AvatarFallback>
                        </Avatar>
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <div className="font-medium">{user.nickname}</div>
                            {user.admin ? <Badge>管理员</Badge> : null}
                            <Badge variant="outline">
                              {formatUserStatus(user.status)}
                            </Badge>
                          </div>
                          <div className="text-sm text-muted-foreground">
                            @{user.username} · {user.email}
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            用户 ID: {user.id} · 邮箱:{' '}
                            {user.email_verified ? '已验证' : '未验证'} · 存储:{' '}
                            {formatBytes(user.storage_used)} /{' '}
                            {formatUserQuota(user.storage_quota)} · 注册时间:{' '}
                            {formatDateTime(user.created_at)}
                          </div>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <Button
                          onClick={() => startEdit(user)}
                          size="sm"
                          type="button"
                          variant="outline"
                        >
                          <PencilLineIcon data-icon="inline-start" />
                          编辑
                        </Button>
                        <Button
                          onClick={() => setPendingDelete(user)}
                          size="sm"
                          type="button"
                          variant="destructive"
                        >
                          <Trash2Icon data-icon="inline-start" />
                          删除
                        </Button>
                      </div>
                    </div>
                  </AdminMotionItem>

                  {index < users.length - 1 ? (
                    <Separator className="bg-border/80 data-horizontal:h-0.5" />
                  ) : null}
                </div>
              ))}
            </div>

            <AdminPagination
              onPageChange={setPage}
              page={page}
              pageSize={ADMIN_PAGE_SIZE}
              total={usersQuery.data.total}
            />
          </>
        ) : null}
      </AdminSection>
    </div>
  )
}
