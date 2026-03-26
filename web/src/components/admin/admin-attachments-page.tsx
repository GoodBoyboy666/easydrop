import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ExternalLinkIcon, Trash2Icon } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'
import {
  ADMIN_PAGE_SIZE,
  ADMIN_SELECT_CLASSNAME,
  formatAttachmentBizType,
  formatBytes,
  parseOptionalInteger,
} from '#/lib/admin'
import { api } from '#/lib/api'
import { formatDateTime } from '#/lib/format'
import { adminAttachmentsQueryOptions, queryKeys } from '#/lib/query-options'
import type { AttachmentDTO } from '#/lib/types'
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
import { Button } from '#/components/ui/button'
import { Field, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'

interface AttachmentFilterState {
  attachmentId: string
  bizType: string
  createdFrom: string
  createdTo: string
  order: string
  userId: string
}

const EMPTY_FILTERS: AttachmentFilterState = {
  attachmentId: '',
  bizType: 'all',
  createdFrom: '',
  createdTo: '',
  order: 'created_at_desc',
  userId: '',
}

function toUnixSeconds(value: string) {
  if (!value) {
    return undefined
  }

  const timestamp = new Date(value).getTime()
  if (Number.isNaN(timestamp)) {
    return undefined
  }

  return Math.floor(timestamp / 1000)
}

export function AdminAttachmentsPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/attachments')
  const [draftFilters, setDraftFilters] =
    useState<AttachmentFilterState>(EMPTY_FILTERS)
  const [filters, setFilters] = useState<AttachmentFilterState>(EMPTY_FILTERS)
  const [page, setPage] = useState(0)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [pendingDelete, setPendingDelete] = useState<AttachmentDTO | null>(null)
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false)

  const attachmentsQuery = useQuery({
    ...adminAttachmentsQueryOptions(auth.token ?? '', {
      biz_type: filters.bizType === 'all' ? undefined : Number(filters.bizType),
      created_from: toUnixSeconds(filters.createdFrom),
      created_to: toUnixSeconds(filters.createdTo),
      id: parseOptionalInteger(filters.attachmentId),
      limit: ADMIN_PAGE_SIZE,
      offset: page * ADMIN_PAGE_SIZE,
      order: filters.order,
      user_id: parseOptionalInteger(filters.userId),
    }),
    enabled: !!auth.token,
  })

  const deleteMutation = useMutation({
    mutationFn: (attachment: AttachmentDTO) =>
      api.deleteAdminAttachment(attachment.id, auth.token!),
  })
  const batchDeleteMutation = useMutation({
    mutationFn: (ids: number[]) =>
      api.batchDeleteAdminAttachments(ids, auth.token!),
  })

  const attachments = attachmentsQuery.data?.items ?? []
  const selectedCount = selectedIds.length
  const allVisibleSelected =
    attachments.length > 0 &&
    attachments.every((attachment) => selectedIds.includes(attachment.id))

  async function invalidateAttachmentQueries() {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: queryKeys.adminAttachmentsPrefix(auth.token),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.adminUsersPrefix(auth.token),
      }),
    ])
  }

  function toggleSelected(id: number) {
    setSelectedIds((current) =>
      current.includes(id)
        ? current.filter((item) => item !== id)
        : [...current, id],
    )
  }

  function toggleSelectAll() {
    if (allVisibleSelected) {
      setSelectedIds((current) =>
        current.filter((id) => !attachments.some((item) => item.id === id)),
      )
      return
    }

    setSelectedIds((current) => [
      ...new Set([...current, ...attachments.map((item) => item.id)]),
    ])
  }

  async function handleDeleteSingle() {
    if (!auth.token || !pendingDelete) {
      return
    }

    try {
      await deleteMutation.mutateAsync(pendingDelete)
      toast.success(`附件 #${pendingDelete.id} 已删除`)
      setSelectedIds((current) =>
        current.filter((id) => id !== pendingDelete.id),
      )
      setPendingDelete(null)
      await invalidateAttachmentQueries()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '删除附件失败')
      setPendingDelete(null)
    }
  }

  async function handleBatchDelete() {
    if (!auth.token || selectedIds.length === 0) {
      return
    }

    try {
      const result = await batchDeleteMutation.mutateAsync(selectedIds)
      if (result.success_ids.length > 0) {
        toast.success(`已删除 ${result.success_ids.length} 个附件`)
      }
      if (result.failed.length > 0) {
        toast.error(`有 ${result.failed.length} 个附件删除失败`)
      }
      setSelectedIds(result.failed.map((item) => item.id))
      setBatchDeleteOpen(false)
      await invalidateAttachmentQueries()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '批量删除失败')
      setBatchDeleteOpen(false)
    }
  }

  const summaryText = useMemo(() => {
    if (selectedCount === 0) {
      return '当前没有选中的附件。'
    }

    return `当前已选中 ${selectedCount} 个附件，可以直接执行批量删除。`
  }, [selectedCount])

  return (
    <div className="space-y-6">
      <AdminDangerDialog
        busy={deleteMutation.isPending}
        confirmLabel="确认删除"
        description="删除后附件文件和数据库记录都会被清理，此操作不可撤销。"
        onConfirm={() => void handleDeleteSingle()}
        onOpenChange={(open) => {
          if (!open && !deleteMutation.isPending) {
            setPendingDelete(null)
          }
        }}
        open={pendingDelete !== null}
        title="删除这个附件？"
      />

      <AdminDangerDialog
        busy={batchDeleteMutation.isPending}
        confirmLabel="批量删除"
        description={`即将删除 ${selectedCount} 个附件。批量操作会逐条处理，失败项会保留在选中列表中。`}
        onConfirm={() => void handleBatchDelete()}
        onOpenChange={(open) => {
          if (!open && !batchDeleteMutation.isPending) {
            setBatchDeleteOpen(false)
          }
        }}
        open={batchDeleteOpen}
        title="批量删除选中的附件？"
      />

      <AdminPageHeader
        title="附件管理"
        description="按用户、类型和时间范围排查附件，支持单条删除和批量清理。"
        actions={
          <Button
            disabled={selectedCount === 0}
            onClick={() => setBatchDeleteOpen(true)}
            type="button"
            variant="destructive"
          >
            <Trash2Icon data-icon="inline-start" />
            批量删除
          </Button>
        }
      />

      <AdminSection
        title="筛选附件"
        description="时间筛选使用本地时间，提交时会转换为 Unix 秒时间戳。"
      >
        <form
          className="grid gap-4 lg:grid-cols-6"
          onSubmit={(event) => {
            event.preventDefault()
            setFilters(draftFilters)
            setPage(0)
          }}
        >
          <Field>
            <FieldLabel htmlFor="admin-attachments-id">附件 ID</FieldLabel>
            <Input
              id="admin-attachments-id"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  attachmentId: event.target.value,
                }))
              }
              placeholder="例如 12"
              type="number"
              value={draftFilters.attachmentId}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-attachments-user">用户 ID</FieldLabel>
            <Input
              id="admin-attachments-user"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  userId: event.target.value,
                }))
              }
              placeholder="例如 3"
              type="number"
              value={draftFilters.userId}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-attachments-type">附件类型</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-attachments-type"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  bizType: event.target.value,
                }))
              }
              value={draftFilters.bizType}
            >
              <option value="all">全部类型</option>
              <option value="1">图片</option>
              <option value="2">视频</option>
              <option value="3">音频</option>
              <option value="4">文件</option>
            </select>
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-attachments-created-from">
              起始时间
            </FieldLabel>
            <Input
              id="admin-attachments-created-from"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  createdFrom: event.target.value,
                }))
              }
              type="datetime-local"
              value={draftFilters.createdFrom}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-attachments-created-to">
              结束时间
            </FieldLabel>
            <Input
              id="admin-attachments-created-to"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  createdTo: event.target.value,
                }))
              }
              type="datetime-local"
              value={draftFilters.createdTo}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-attachments-order">排序</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-attachments-order"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  order: event.target.value,
                }))
              }
              value={draftFilters.order}
            >
              <option value="created_at_desc">最新优先</option>
              <option value="created_at_asc">最早优先</option>
            </select>
          </Field>

          <div className="flex flex-wrap gap-2 lg:col-span-6">
            <Button type="submit">应用筛选</Button>
            <Button
              onClick={() => {
                setDraftFilters(EMPTY_FILTERS)
                setFilters(EMPTY_FILTERS)
                setPage(0)
                setSelectedIds([])
              }}
              type="button"
              variant="outline"
            >
              重置
            </Button>
          </div>
        </form>
      </AdminSection>

      <AdminSection title="附件列表" description={summaryText}>
        <div className="mb-4 flex flex-wrap gap-2">
          <Button
            onClick={toggleSelectAll}
            size="sm"
            type="button"
            variant="outline"
          >
            {allVisibleSelected ? '取消当前页全选' : '选中当前页'}
          </Button>
          <Button
            disabled={selectedCount === 0}
            onClick={() => setSelectedIds([])}
            size="sm"
            type="button"
            variant="outline"
          >
            清空选择
          </Button>
        </div>

        {attachmentsQuery.isPending ? <AdminListSkeleton /> : null}
        {attachmentsQuery.error instanceof Error ? (
          <AdminErrorAlert
            description={attachmentsQuery.error.message}
            title="附件列表读取失败"
          />
        ) : null}
        {!attachmentsQuery.isPending &&
        !attachmentsQuery.error &&
        attachments.length === 0 ? (
          <AdminEmptyState
            description="可以调整筛选条件，或者等待新的附件上传。"
            title="没有找到符合条件的附件"
          />
        ) : null}

        {!attachmentsQuery.isPending &&
        !attachmentsQuery.error &&
        attachments.length > 0 ? (
          <>
            <div className="overflow-hidden bg-transparent">
              {attachments.map((attachment, index) => (
                <div key={attachment.id}>
                  <AdminMotionItem className="p-4" delay={index * 0.03}>
                    <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                      <div className="flex min-w-0 items-start gap-3">
                        <input
                          checked={selectedIds.includes(attachment.id)}
                          className="mt-1 size-4 rounded border-border"
                          onChange={() => toggleSelected(attachment.id)}
                          type="checkbox"
                        />
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <div className="font-medium">
                              附件ID: {attachment.id}
                            </div>
                          </div>
                          <div className="text-sm text-muted-foreground">
                            
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            用户ID: {attachment.user_id} ·{' '}
                            {formatAttachmentBizType(attachment.biz_type)} ·
                            文件大小: {formatBytes(attachment.file_size)} ·
                            存储类型: {attachment.storage_type} · 上传时间:{' '}
                            {formatDateTime(attachment.created_at)}
                          </div>
                          <div className="mt-2 text-xs text-muted-foreground break-all">
                            对象键: {attachment.file_key}
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground break-all">
                            访问地址: {attachment.url}
                          </div>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <Button
                          asChild
                          size="sm"
                          type="button"
                          variant="outline"
                        >
                          <a
                            href={attachment.url}
                            rel="noreferrer"
                            target="_blank"
                          >
                            <ExternalLinkIcon data-icon="inline-start" />
                            打开文件
                          </a>
                        </Button>
                        <Button
                          onClick={() => setPendingDelete(attachment)}
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

                  {index < attachments.length - 1 ? (
                    <Separator className="bg-border/80 data-horizontal:h-0.5" />
                  ) : null}
                </div>
              ))}
            </div>

            <AdminPagination
              onPageChange={setPage}
              page={page}
              pageSize={ADMIN_PAGE_SIZE}
              total={attachmentsQuery.data.total}
            />
          </>
        ) : null}
      </AdminSection>
    </div>
  )
}
