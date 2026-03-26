import { Link, Outlet, useLocation } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { CartesianGrid, Line, LineChart, XAxis } from 'recharts'
import type { ChartConfig } from '#/components/ui/chart'
import type { CommentDTO, PostDTO } from '#/lib/types'
import { useAuth } from '#/lib/auth'
import {
  adminAttachmentsQueryOptions,
  adminCommentsQueryOptions,
  adminPostsQueryOptions,
  adminUsersQueryOptions,
} from '#/lib/query-options'
import {
  AdminAccessNotice,
  AdminLayout,
  AdminPageHeader,
  AdminSection,
  AdminStatCard,
} from '#/components/admin/admin-ui'
import { Button } from '#/components/ui/button'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '#/components/ui/chart'

const overviewChartConfig = {
  posts: {
    label: '日志发布',
    color: 'var(--color-chart-2)',
  },
  comments: {
    label: '评论发布',
    color: 'var(--color-chart-4)',
  },
} satisfies ChartConfig

function formatDayKey(date: Date) {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatDayLabel(date: Date) {
  return `${date.getMonth() + 1}/${date.getDate()}`
}

function countItemsByDay(items: Array<{ created_at: string }>) {
  const counts = new Map<string, number>()

  for (const item of items) {
    const createdAt = new Date(item.created_at)
    if (Number.isNaN(createdAt.getTime())) {
      continue
    }

    createdAt.setHours(0, 0, 0, 0)
    const key = formatDayKey(createdAt)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }

  return counts
}

function buildRecentTrendChartData(posts: PostDTO[], comments: CommentDTO[]) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const dates = Array.from({ length: 7 }, (_, index) => {
    const current = new Date(today)
    current.setDate(today.getDate() - (6 - index))
    return current
  })

  const postCounts = countItemsByDay(posts)
  const commentCounts = countItemsByDay(comments)

  return dates.map((date) => {
    const dayKey = formatDayKey(date)

    return {
      comments: commentCounts.get(dayKey) ?? 0,
      day: formatDayLabel(date),
      posts: postCounts.get(dayKey) ?? 0,
    }
  })
}

export function AdminRoutePage() {
  const auth = useAuth()
  const location = useLocation()

  if (auth.status !== 'authenticated') {
    return (
      <AdminAccessNotice
        title="需要管理员登录"
        description="后台管理仅对已登录管理员开放。请先登录。"
        actions={
          <>
            <Button asChild>
              <Link search={{ redirect: location.pathname }} to="/login">
                去登录
              </Link>
            </Button>
            <Button asChild variant="outline">
              <Link search={{ content: undefined }} to="/">
                返回首页
              </Link>
            </Button>
          </>
        }
      />
    )
  }

  if (!auth.isAdmin) {
    return (
      <AdminAccessNotice
        title="无权访问后台管理"
        description="当前账号不是管理员，不能进入后台管理入口。"
        actions={
          <Button asChild variant="outline">
            <Link search={{ content: undefined }} to="/">
              返回首页
            </Link>
          </Button>
        }
      />
    )
  }

  return (
    <AdminLayout activePath={location.pathname}>
      {location.pathname === '/admin' ? <AdminOverviewPage /> : <Outlet />}
    </AdminLayout>
  )
}

function AdminOverviewPage() {
  const auth = useAuth()
  const enabled = auth.status === 'authenticated' && auth.isAdmin

  const usersQuery = useQuery({
    ...adminUsersQueryOptions({
      limit: 1,
      offset: 0,
      order: 'created_at_desc',
    }),
    enabled,
  })
  const postsQuery = useQuery({
    ...adminPostsQueryOptions({
      limit: 500,
      offset: 0,
      order: 'created_at_desc',
    }),
    enabled,
  })
  const commentsQuery = useQuery({
    ...adminCommentsQueryOptions({
      limit: 500,
      offset: 0,
      order: 'created_at_desc',
    }),
    enabled,
  })
  const attachmentsQuery = useQuery({
    ...adminAttachmentsQueryOptions({
      limit: 1,
      offset: 0,
      order: 'created_at_desc',
    }),
    enabled,
  })

  const overviewChartData = buildRecentTrendChartData(
    postsQuery.data?.items ?? [],
    commentsQuery.data?.items ?? [],
  )

  return (
    <div className="space-y-6">
      <AdminPageHeader title="概览" />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard title="用户" value={usersQuery.data?.total ?? '--'} />
        <AdminStatCard title="日志" value={postsQuery.data?.total ?? '--'} />
        <AdminStatCard title="评论" value={commentsQuery.data?.total ?? '--'} />
        <AdminStatCard title="附件" value={attachmentsQuery.data?.total ?? '--'} />
      </div>

      <AdminSection title="数据图表">
        <ChartContainer
          className="h-[320px] w-full"
          config={overviewChartConfig}
        >
          <LineChart accessibilityLayer data={overviewChartData}>
            <CartesianGrid vertical={false} />
            <XAxis
              axisLine={false}
              dataKey="day"
              tickLine={false}
              tickMargin={10}
            />
            <ChartLegend content={<ChartLegendContent />} />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  formatter={(value, name) => `${value} ${name}`}
                  labelFormatter={(label) => `日期：${label}`}
                />
              }
              cursor={false}
            />
            <Line
              dataKey="posts"
              dot={{
                fill: 'var(--color-posts)',
                r: 4,
                stroke: 'var(--background)',
                strokeWidth: 2,
              }}
              stroke="var(--color-posts)"
              strokeWidth={2.5}
              type="monotone"
            />
            <Line
              dataKey="comments"
              dot={{
                fill: 'var(--color-comments)',
                r: 4,
                stroke: 'var(--background)',
                strokeWidth: 2,
              }}
              stroke="var(--color-comments)"
              strokeWidth={2.5}
              type="monotone"
            />
          </LineChart>
        </ChartContainer>
      </AdminSection>
    </div>
  )
}
