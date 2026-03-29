import { Link } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PencilLineIcon, Trash2Icon } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import {
  ADMIN_PAGE_SIZE,
  ADMIN_SELECT_CLASSNAME,
  parseOptionalInteger,
} from '#/lib/admin'
import { api } from '#/lib/api'
import { formatDateTime } from '#/lib/format'
import { hasMarkdownContent, normalizeMarkdownContent } from '#/lib/markdown'
import { adminCommentsQueryOptions } from '#/lib/query-options'
import { invalidateAdminCommentQueries } from '#/lib/query-invalidation'
import type { CommentDTO } from '#/lib/types'
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
import { MarkdownContent } from '#/components/markdown/markdown-content'
import { MarkdownEditor } from '#/components/markdown/markdown-editor'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'

interface CommentFilterState {
  order: string
  postId: string
  userId: string
}

const EMPTY_FILTERS: CommentFilterState = {
  order: 'created_at_desc',
  postId: '',
  userId: '',
}

export function AdminCommentsPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/comments')
  const [draftFilters, setDraftFilters] =
    useState<CommentFilterState>(EMPTY_FILTERS)
  const [filters, setFilters] = useState<CommentFilterState>(EMPTY_FILTERS)
  const [page, setPage] = useState(1)
  const [editingComment, setEditingComment] = useState<CommentDTO | null>(null)
  const [draftContent, setDraftContent] = useState('')
  const [pendingDelete, setPendingDelete] = useState<CommentDTO | null>(null)

  const commentsQuery = useQuery({
    ...adminCommentsQueryOptions({
      page,
      order: filters.order,
      size: ADMIN_PAGE_SIZE,
      post_id: parseOptionalInteger(filters.postId),
      user_id: parseOptionalInteger(filters.userId),
    }),
    enabled: auth.status === 'authenticated',
  })

  const updateMutation = useMutation({
    mutationFn: () =>
      api.updateAdminComment(
        editingComment!.id,
        {
          content: normalizeMarkdownContent(draftContent),
        },
      ),
  })
  const deleteMutation = useMutation({
    mutationFn: (comment: CommentDTO) => api.deleteAdminComment(comment.id),
  })

  const comments = commentsQuery.data?.items ?? []

  function startEdit(comment: CommentDTO) {
    setEditingComment(comment)
    setDraftContent(comment.content)
  }

  function resetEditor() {
    setEditingComment(null)
    setDraftContent('')
  }

  async function invalidateCommentQueries(comment: CommentDTO) {
    await invalidateAdminCommentQueries(queryClient, comment.post_id)
  }

  async function handleSaveComment(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (auth.status !== 'authenticated' || !editingComment) {
      return
    }

    if (!hasMarkdownContent(draftContent)) {
      toast.error('评论内容不能为空')
      return
    }

    try {
      const updatedComment = await updateMutation.mutateAsync()
      toast.success(`评论 #${updatedComment.id} 已更新`)
      await invalidateCommentQueries(updatedComment)
      resetEditor()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '保存评论失败')
    }
  }

  async function handleDeleteComment() {
    if (auth.status !== 'authenticated' || !pendingDelete) {
      return
    }

    try {
      await deleteMutation.mutateAsync(pendingDelete)
      toast.success(`评论 #${pendingDelete.id} 已删除`)
      await invalidateCommentQueries(pendingDelete)
      if (editingComment?.id === pendingDelete.id) {
        resetEditor()
      }
      setPendingDelete(null)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '删除评论失败')
      setPendingDelete(null)
    }
  }

  return (
    <div className="space-y-6">
      <AdminDangerDialog
        busy={deleteMutation.isPending}
        confirmLabel="确认删除"
        description="删除后这条评论会从前台评论区和相关列表中移除。"
        onConfirm={() => void handleDeleteComment()}
        onOpenChange={(open) => {
          if (!open && !deleteMutation.isPending) {
            setPendingDelete(null)
          }
        }}
        open={pendingDelete !== null}
        title="删除这条评论？"
      />

      <AdminPageHeader title="评论管理" />

      <AdminSection title="筛选评论">
        <form
          className="grid gap-4 lg:grid-cols-4"
          onSubmit={(event) => {
            event.preventDefault()
            setFilters(draftFilters)
            setPage(1)
          }}
        >
          <Field>
            <FieldLabel htmlFor="admin-comments-post">日志 ID</FieldLabel>
            <Input
              id="admin-comments-post"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  postId: event.target.value,
                }))
              }
              placeholder="例如 12"
              type="number"
              value={draftFilters.postId}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-comments-user">用户 ID</FieldLabel>
            <Input
              id="admin-comments-user"
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
            <FieldLabel htmlFor="admin-comments-order">排序</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-comments-order"
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

          <div className="flex flex-wrap gap-2 lg:col-span-4">
            <Button type="submit">应用筛选</Button>
            <Button
              onClick={() => {
                setDraftFilters(EMPTY_FILTERS)
                setFilters(EMPTY_FILTERS)
                setPage(1)
              }}
              type="button"
              variant="outline"
            >
              重置
            </Button>
          </div>
        </form>
      </AdminSection>

      {editingComment ? (
        <AdminSection title={`编辑评论 #${editingComment.id}`}>
          <form className="space-y-4" onSubmit={handleSaveComment}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="admin-comment-editor">评论内容</FieldLabel>
                <MarkdownEditor
                  height={220}
                  onChange={setDraftContent}
                  placeholder="编辑评论，支持 Markdown。"
                  value={draftContent}
                />
              </Field>
            </FieldGroup>

            <div className="flex flex-wrap gap-2">
              <Button disabled={updateMutation.isPending} type="submit">
                {updateMutation.isPending ? '保存中…' : '保存评论'}
              </Button>
              <Button onClick={resetEditor} type="button" variant="outline">
                取消编辑
              </Button>
            </div>
          </form>
        </AdminSection>
      ) : null}

      <AdminSection title="评论列表">
        {commentsQuery.isPending ? <AdminListSkeleton /> : null}
        {commentsQuery.error instanceof Error ? (
          <AdminErrorAlert
            description={commentsQuery.error.message}
            title="评论列表读取失败"
          />
        ) : null}
        {!commentsQuery.isPending &&
        !commentsQuery.error &&
        comments.length === 0 ? (
          <AdminEmptyState
            description="可以调整筛选条件，或等待用户产生新的评论。"
            title="没有找到符合条件的评论"
          />
        ) : null}

        {!commentsQuery.isPending &&
        !commentsQuery.error &&
        comments.length > 0 ? (
          <>
            <div className="overflow-hidden bg-transparent">
              {comments.map((comment, index) => (
                <div key={comment.id}>
                  <AdminMotionItem className="p-4" delay={index * 0.03}>
                    <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <div className="font-medium">
                            {comment.author.nickname}
                          </div>
                          {comment.author.admin ? <Badge>管理员</Badge> : null}
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          评论ID: {comment.id} · 用户ID: {comment.author.id} ·
                          日志 #{comment.post_id} · 创建于{' '}
                          {formatDateTime(comment.created_at)}
                          {comment.updated_at &&
                          comment.updated_at !== comment.created_at
                            ? ` · 更新于 ${formatDateTime(comment.updated_at)}`
                            : ''}
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <Button
                          asChild
                          size="sm"
                          type="button"
                          variant="outline"
                        >
                          <Link
                            params={{ id: String(comment.post_id) }}
                            to="/posts/$id"
                          >
                            查看日志
                          </Link>
                        </Button>
                        <Button
                          onClick={() => startEdit(comment)}
                          size="sm"
                          type="button"
                          variant="outline"
                        >
                          <PencilLineIcon data-icon="inline-start" />
                          编辑
                        </Button>
                        <Button
                          onClick={() => setPendingDelete(comment)}
                          size="sm"
                          type="button"
                          variant="destructive"
                        >
                          <Trash2Icon data-icon="inline-start" />
                          删除
                        </Button>
                      </div>
                    </div>

                    <div className="mt-4">
                      {comment.reply_to_user ? (
                        <div className="flex items-baseline gap-2">
                          <span className="shrink-0 text-sm text-muted-foreground">
                            回复 @{comment.reply_to_user.nickname}
                          </span>
                          <div className="min-w-0 flex-1">
                            <MarkdownContent
                              className="[&_.markdown-body]:inline [&_.markdown-body>p:first-child]:inline [&_.markdown-body>p:first-child]:m-0"
                              content={comment.content}
                            />
                          </div>
                        </div>
                      ) : (
                        <MarkdownContent content={comment.content} />
                      )}
                    </div>
                  </AdminMotionItem>

                  {index < comments.length - 1 ? (
                    <Separator className="bg-border/80 data-horizontal:h-0.5" />
                  ) : null}
                </div>
              ))}
            </div>

            <AdminPagination
              onPageChange={setPage}
              page={page}
              pageSize={ADMIN_PAGE_SIZE}
              total={commentsQuery.data.total}
            />
          </>
        ) : null}
      </AdminSection>
    </div>
  )
}
