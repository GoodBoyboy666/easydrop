import { Link } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PinIcon, Trash2Icon } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import {
  ADMIN_PAGE_SIZE,
  ADMIN_SELECT_CLASSNAME,
  parseOptionalBoolean,
  parseOptionalInteger,
} from '#/lib/admin'
import { api } from '#/lib/api'
import { formatDateTime } from '#/lib/format'
import { adminPostsQueryOptions, queryKeys } from '#/lib/query-options'
import type { PostDTO } from '#/lib/types'
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
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { Field, FieldLabel } from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'

interface PostFilterState {
  content: string
  hide: string
  order: string
  tagId: string
  userId: string
}

const EMPTY_FILTERS: PostFilterState = {
  content: '',
  hide: 'all',
  order: 'created_at_desc',
  tagId: '',
  userId: '',
}

export function AdminPostsPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/posts')
  const [draftFilters, setDraftFilters] =
    useState<PostFilterState>(EMPTY_FILTERS)
  const [filters, setFilters] = useState<PostFilterState>(EMPTY_FILTERS)
  const [page, setPage] = useState(0)
  const [pendingDelete, setPendingDelete] = useState<PostDTO | null>(null)

  const postsQuery = useQuery({
    ...adminPostsQueryOptions({
      content: filters.content.trim() || undefined,
      hide: parseOptionalBoolean(filters.hide),
      limit: ADMIN_PAGE_SIZE,
      offset: page * ADMIN_PAGE_SIZE,
      order: filters.order,
      tag_id: parseOptionalInteger(filters.tagId),
      user_id: parseOptionalInteger(filters.userId),
    }),
    enabled: auth.status === 'authenticated',
  })

  const deleteMutation = useMutation({
    mutationFn: (post: PostDTO) => api.deleteAdminPost(post.id),
  })

  const posts = postsQuery.data?.items ?? []

  async function invalidatePostQueries(postId?: number) {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: queryKeys.adminPostsPrefix(),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.postsPrefix(),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.latestCommentsPrefix(),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.adminCommentsPrefix(),
      }),
      ...(postId
        ? [
            queryClient.invalidateQueries({
              queryKey: queryKeys.postPrefix(),
            }),
            queryClient.invalidateQueries({
              queryKey: queryKeys.postCommentsPrefix(postId),
            }),
          ]
        : []),
    ])
  }

  async function handleDeletePost() {
    if (auth.status !== 'authenticated' || !pendingDelete) {
      return
    }

    try {
      await deleteMutation.mutateAsync(pendingDelete)
      toast.success(`日志 #${pendingDelete.id} 已删除`)
      await invalidatePostQueries(pendingDelete.id)
      setPendingDelete(null)
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '删除日志失败')
      setPendingDelete(null)
    }
  }

  return (
    <div className="space-y-6">
      <AdminDangerDialog
        busy={deleteMutation.isPending}
        confirmLabel="确认删除"
        description="删除日志后，对应前台内容会立即移除，相关评论也可能受影响。"
        onConfirm={() => void handleDeletePost()}
        onOpenChange={(open) => {
          if (!open && !deleteMutation.isPending) {
            setPendingDelete(null)
          }
        }}
        open={pendingDelete !== null}
        title="删除这篇日志？"
      />

      <AdminPageHeader title="日志管理" />

      <AdminSection title="筛选日志">
        <form
          className="grid gap-4 lg:grid-cols-5"
          onSubmit={(event) => {
            event.preventDefault()
            setFilters(draftFilters)
            setPage(0)
          }}
        >
          <Field>
            <FieldLabel htmlFor="admin-posts-content">内容关键词</FieldLabel>
            <Input
              id="admin-posts-content"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  content: event.target.value,
                }))
              }
              placeholder="搜索内容片段"
              value={draftFilters.content}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-posts-user">作者 ID</FieldLabel>
            <Input
              id="admin-posts-user"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  userId: event.target.value,
                }))
              }
              placeholder="例如 1"
              type="number"
              value={draftFilters.userId}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-posts-tag">标签 ID</FieldLabel>
            <Input
              id="admin-posts-tag"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  tagId: event.target.value,
                }))
              }
              placeholder="例如 3"
              type="number"
              value={draftFilters.tagId}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-posts-hide">隐藏状态</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-posts-hide"
              onChange={(event) =>
                setDraftFilters((current) => ({
                  ...current,
                  hide: event.target.value,
                }))
              }
              value={draftFilters.hide}
            >
              <option value="all">全部</option>
              <option value="false">仅公开</option>
              <option value="true">仅隐藏</option>
            </select>
          </Field>

          <Field>
            <FieldLabel htmlFor="admin-posts-order">排序</FieldLabel>
            <select
              className={ADMIN_SELECT_CLASSNAME}
              id="admin-posts-order"
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

          <div className="flex flex-wrap gap-2 lg:col-span-5">
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

      <AdminSection title="日志列表">
        {postsQuery.isPending ? <AdminListSkeleton /> : null}
        {postsQuery.error instanceof Error ? (
          <AdminErrorAlert
            description={postsQuery.error.message}
            title="日志列表读取失败"
          />
        ) : null}
        {!postsQuery.isPending && !postsQuery.error && posts.length === 0 ? (
          <AdminEmptyState
            description="可以调整筛选条件后重试。"
            title="没有找到匹配的日志"
          />
        ) : null}

        {!postsQuery.isPending && !postsQuery.error && posts.length > 0 ? (
          <>
            <div className="overflow-hidden bg-transparent">
              {posts.map((post, index) => (
                <div key={post.id}>
                  <AdminMotionItem className="p-4" delay={index * 0.03}>
                    <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <div className="font-medium">日志 #{post.id}</div>
                          {post.hide ? (
                            <Badge variant="outline">已隐藏</Badge>
                          ) : null}
                          {post.disable_comment ? (
                            <Badge variant="outline">已禁评</Badge>
                          ) : null}
                          {typeof post.pin === 'number' ? (
                            <Badge>
                              <PinIcon data-icon="inline-start" />
                              置顶 {post.pin}
                            </Badge>
                          ) : null}
                        </div>
                        <div className="mt-1 text-sm text-muted-foreground">
                          {post.author.nickname}（ID: {post.author.id}） ·
                          创建于 {formatDateTime(post.created_at)}
                          {post.updated_at &&
                          post.updated_at !== post.created_at ? (
                            <> · 更新于 {formatDateTime(post.updated_at)}</>
                          ) : null}
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
                            params={{ id: String(post.id) }}
                            to="/posts/$id"
                          >
                            预览
                          </Link>
                        </Button>
                        <Button
                          onClick={() => setPendingDelete(post)}
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
                      <MarkdownContent content={post.content} />
                    </div>

                    {post.tags && post.tags.length > 0 ? (
                      <div className="mt-4 flex flex-wrap gap-2">
                        {post.tags.map((tag) => (
                          <Badge key={tag.id} variant="outline">
                            #{tag.name}
                          </Badge>
                        ))}
                      </div>
                    ) : null}
                  </AdminMotionItem>

                  {index < posts.length - 1 ? (
                    <Separator className="bg-border/80 data-horizontal:h-0.5" />
                  ) : null}
                </div>
              ))}
            </div>

            <AdminPagination
              onPageChange={setPage}
              page={page}
              pageSize={ADMIN_PAGE_SIZE}
              total={postsQuery.data.total}
            />
          </>
        ) : null}
      </AdminSection>
    </div>
  )
}
