import { Link, Outlet, useLocation } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { CartesianGrid, Line, LineChart, XAxis } from 'recharts'
import type { ChartConfig } from '#/components/ui/chart'
import { useAuth } from '#/lib/auth'
import { adminOverviewQueryOptions } from '#/lib/query-options'
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

function formatTrendDayLabel(date: string) {
  const matched = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date)
  if (!matched) {
    return date
  }

  return `${Number(matched[2])}/${Number(matched[3])}`
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

  const overviewQuery = useQuery({
    ...adminOverviewQueryOptions(),
    enabled,
  })

  const overviewChartData = (overviewQuery.data?.recent_activity ?? []).map(
    (item) => ({
      comments: item.comments,
      day: formatTrendDayLabel(item.date),
      posts: item.posts,
    }),
  )

  return (
    <div className="space-y-6">
      <AdminPageHeader title="概览" />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard
          title="用户"
          value={overviewQuery.data?.totals.users ?? '--'}
        />
        <AdminStatCard
          title="日志"
          value={overviewQuery.data?.totals.posts ?? '--'}
        />
        <AdminStatCard
          title="评论"
          value={overviewQuery.data?.totals.comments ?? '--'}
        />
        <AdminStatCard
          title="附件"
          value={overviewQuery.data?.totals.attachments ?? '--'}
        />
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
