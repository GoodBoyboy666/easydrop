import { createFileRoute } from '@tanstack/react-router'

import { Badge } from '#/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'

export const Route = createFileRoute('/about')({
  component: About,
})

function About() {
  return (
    <main className="page-wrap px-4 py-12">
      <Card className="rounded-[2rem] border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.06)]">
        <CardHeader className="gap-3 border-b border-border/60">
          <Badge className="w-fit" variant="secondary">
            About
          </Badge>
          <CardTitle className="display-title text-4xl sm:text-5xl">
            一个公开阅读优先的说说首页
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 pt-6 md:grid-cols-3">
          <div className="rounded-2xl border bg-background/70 p-5">
            <h2 className="m-0 text-base font-semibold text-foreground">
              公开接口驱动
            </h2>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">
              使用 `/posts`、`/posts/{'{'}id{'}'}/comments`、`/settings/public`
              构建首页，不依赖管理端接口。
            </p>
          </div>
          <div className="rounded-2xl border bg-background/70 p-5">
            <h2 className="m-0 text-base font-semibold text-foreground">
              扁平评论结构
            </h2>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">
              评论按时间顺序直接展示，仅补充“回复某人”的关系提示，不做树形嵌套。
            </p>
          </div>
          <div className="rounded-2xl border bg-background/70 p-5">
            <h2 className="m-0 text-base font-semibold text-foreground">
              Markdown 安全渲染
            </h2>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">
              内容只渲染 Markdown，跳过原始
              HTML，并拦截危险链接协议，避免执行脚本。
            </p>
          </div>
        </CardContent>
      </Card>
    </main>
  )
}
