import { LoaderCircleIcon, SendHorizonalIcon, ShieldIcon } from 'lucide-react'

import type { CurrentUser } from '#/lib/easydrop-api'
import { formatAbsoluteTime } from '#/lib/time'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import { Label } from '#/components/ui/label'
import { Switch } from '#/components/ui/switch'
import { Textarea } from '#/components/ui/textarea'

type AdminComposerProps = {
  content: string
  error: string | null
  hide: boolean
  isSubmitting: boolean
  onContentChange: (content: string) => void
  onHideChange: (hide: boolean) => void
  onSubmit: () => void
  successMessage: string | null
  user: CurrentUser
}

function getDisplayName(user: CurrentUser) {
  return (
    user.nickname ||
    user.username ||
    (user.id ? `管理员 #${user.id}` : '管理员')
  )
}

export function AdminComposer({
  content,
  error,
  hide,
  isSubmitting,
  onContentChange,
  onHideChange,
  onSubmit,
  successMessage,
  user,
}: AdminComposerProps) {
  const canSubmit = content.trim().length > 0 && !isSubmitting

  return (
    <Card className="overflow-hidden border border-border/70 bg-card/90 shadow-[0_24px_64px_rgba(15,23,42,0.07)] backdrop-blur-xl">
      <CardHeader className="gap-3 border-b border-border/60">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">
            <ShieldIcon data-icon="inline-start" />
            管理员快捷发布
          </Badge>
          <Badge variant="outline">仅管理员可见</Badge>
        </div>
        <CardTitle className="text-2xl">前台直接发一条说说</CardTitle>
        <CardDescription className="flex flex-wrap items-center gap-2">
          <span>当前身份：{getDisplayName(user)}</span>
          {user.updated_at ? (
            <>
              <span aria-hidden="true">·</span>
              <span title={formatAbsoluteTime(user.updated_at)}>
                最近资料更新时间 {formatAbsoluteTime(user.updated_at)}
              </span>
            </>
          ) : null}
        </CardDescription>
      </CardHeader>

      <CardContent className="flex flex-col gap-5 pt-5">
        <div className="flex flex-col gap-2">
          <Label htmlFor="post-content">说说内容</Label>
          <Textarea
            id="post-content"
            onChange={(event) => onContentChange(event.target.value)}
            placeholder="支持 Markdown。写点今天想留下的话。"
            value={content}
          />
          <p className="text-sm text-muted-foreground">
            说说内容会按 Markdown 渲染，脚本和危险链接不会执行。
          </p>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border bg-background/60 px-4 py-3">
          <div className="flex flex-col gap-1">
            <Label htmlFor="post-hide">隐藏说说</Label>
            <p className="text-sm text-muted-foreground">
              打开后仍会按接口字段提交 `hide=true`。
            </p>
          </div>
          <Switch
            checked={hide}
            id="post-hide"
            onCheckedChange={onHideChange}
          />
        </div>

        {error ? (
          <Alert variant="destructive">
            <AlertTitle>发布失败</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        {successMessage ? (
          <Alert>
            <AlertTitle>发布成功</AlertTitle>
            <AlertDescription>{successMessage}</AlertDescription>
          </Alert>
        ) : null}
      </CardContent>

      <CardFooter className="flex items-center justify-between gap-3 border-t border-border/60 bg-muted/35">
        <p className="text-sm text-muted-foreground">
          发布成功后会自动刷新首页时间流。
        </p>
        <Button disabled={!canSubmit} onClick={onSubmit} size="lg">
          {isSubmitting ? (
            <LoaderCircleIcon
              className="animate-spin"
              data-icon="inline-start"
            />
          ) : (
            <SendHorizonalIcon data-icon="inline-start" />
          )}
          {isSubmitting ? '发布中...' : '立即发布'}
        </Button>
      </CardFooter>
    </Card>
  )
}
