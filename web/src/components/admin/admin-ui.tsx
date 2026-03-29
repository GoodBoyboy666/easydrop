import { Link } from '@tanstack/react-router'
import {
  AlertTriangleIcon,
  ArrowLeftIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { AdminNavItem } from '#/lib/admin'
import { adminNavItems } from '#/lib/admin'
import { cn } from '#/lib/utils'
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
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '#/components/ui/empty'
import { Separator } from '#/components/ui/separator'
import { Skeleton } from '#/components/ui/skeleton'

function getAdminMotionProps(prefersReducedMotion: boolean, delay = 0) {
  if (prefersReducedMotion) {
    return {
      animate: { opacity: 1, y: 0 },
      initial: false as const,
      transition: { duration: 0 },
    }
  }

  return {
    animate: { opacity: 1, y: 0 },
    initial: { opacity: 0, y: 14 },
    transition: {
      delay,
      duration: 0.28,
      ease: [0.22, 1, 0.36, 1] as const,
    },
  }
}

export function AdminMotionItem({
  children,
  className,
  delay = 0,
}: {
  children: React.ReactNode
  className?: string
  delay?: number
}) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <motion.div
      className={className}
      {...getAdminMotionProps(prefersReducedMotion, delay)}
    >
      {children}
    </motion.div>
  )
}

export function AdminAccessNotice({
  actions,
  description,
  title,
}: {
  actions?: React.ReactNode
  description: React.ReactNode
  title: string
}) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      <motion.div {...getAdminMotionProps(prefersReducedMotion)}>
        <Card className="border border-border/70 bg-transparent shadow-sm backdrop-blur-sm">
          <CardHeader>
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </CardHeader>
          {actions ? <CardContent className="flex gap-2">{actions}</CardContent> : null}
        </Card>
      </motion.div>
    </div>
  )
}

export function AdminLayout({
  activePath,
  children,
}: {
  activePath: string
  children: React.ReactNode
}) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="grid gap-6 lg:grid-cols-[280px_minmax(0,1fr)]">
        <aside className="space-y-4">
          <motion.div {...getAdminMotionProps(prefersReducedMotion)}>
            <Card className="overflow-hidden border border-border/70 bg-transparent shadow-sm ring-0 backdrop-blur-sm">
              <CardHeader>
                <CardTitle>后台管理</CardTitle>
              </CardHeader>
              <CardContent className="flex gap-2">
                <Button asChild size="sm" variant="outline">
                  <Link search={{ content: undefined }} to="/">
                    <ArrowLeftIcon data-icon="inline-start" />
                    返回站点
                  </Link>
                </Button>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div {...getAdminMotionProps(prefersReducedMotion, 0.05)}>
            <Card className="border border-border/70 bg-transparent shadow-sm backdrop-blur-sm">
              <CardHeader>
                <CardTitle>菜单</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-col gap-2">
                {adminNavItems.map((item) => (
                  <AdminNavButton
                    key={item.to}
                    active={activePath === item.to}
                    item={item}
                  />
                ))}
              </CardContent>
            </Card>
          </motion.div>
        </aside>

        <div className="min-w-0">{children}</div>
      </div>
    </div>
  )
}

function AdminNavButton({
  active,
  item,
}: {
  active: boolean
  item: AdminNavItem
}) {
  const Icon = item.icon

  return (
    <Link
      to={item.to}
      className={cn(
        'rounded-xl border px-3 py-3 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring',
        active
          ? 'border-primary/40 bg-primary/6 text-foreground'
          : 'border-border/70 bg-transparent hover:bg-muted/25'
      )}
    >
      <div className="flex items-center gap-3">
        <div
          className={cn(
            'flex size-10 items-center justify-center rounded-xl',
            active
              ? 'bg-primary/12 text-primary'
              : 'bg-transparent text-muted-foreground'
          )}
        >
          <Icon />
        </div>
        <div className="min-w-0">
          <div className="font-medium">{item.label}</div>
        </div>
      </div>
    </Link>
  )
}

export function AdminPageHeader({
  actions,
  description,
  title,
}: {
  actions?: React.ReactNode
  description?: React.ReactNode
  title: string
}) {
  return (
    <div className="mb-6 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <h1 className="font-heading text-2xl font-semibold tracking-tight">{title}</h1>
        {description ? (
          <p className="mt-1 text-sm text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {actions ? <div className="flex flex-wrap gap-2">{actions}</div> : null}
    </div>
  )
}

export function AdminStatCard({
  title,
  value,
}: {
  title: string
  value: React.ReactNode
}) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <motion.div {...getAdminMotionProps(prefersReducedMotion)}>
      <Card className="border border-border/70 bg-transparent shadow-sm backdrop-blur-sm">
        <CardHeader>
          <CardDescription>{title}</CardDescription>
          <CardTitle className="text-2xl">{value}</CardTitle>
        </CardHeader>
      </Card>
    </motion.div>
  )
}

export function AdminSection({
  children,
  description,
  title,
}: {
  children: React.ReactNode
  description?: React.ReactNode
  title: string
}) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <motion.div {...getAdminMotionProps(prefersReducedMotion, 0.04)}>
      <Card className="border border-border/70 bg-transparent shadow-sm backdrop-blur-sm">
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          {description ? (
            <CardDescription>{description}</CardDescription>
          ) : null}
        </CardHeader>
        <CardContent>{children}</CardContent>
      </Card>
    </motion.div>
  )
}

export function AdminListSkeleton({ rows = 3 }: { rows?: number }) {
  const prefersReducedMotion = useReducedMotion() ?? false

  return (
    <div className="flex flex-col gap-3">
      {Array.from({ length: rows }).map((_, index) => (
        <motion.div
          key={index}
          {...getAdminMotionProps(prefersReducedMotion, index * 0.04)}
        >
          <Card className="border border-border/70 bg-transparent shadow-sm backdrop-blur-sm">
            <CardContent className="flex flex-col gap-3 pt-4">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-8/12" />
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </div>
  )
}

export function AdminEmptyState({
  description,
  title,
}: {
  description: string
  title: string
}) {
  return (
    <Empty className="border border-dashed border-border/80 bg-transparent backdrop-blur-sm">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <AlertTriangleIcon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function AdminErrorAlert({
  description,
  title,
}: {
  description: React.ReactNode
  title: string
}) {
  return (
    <Alert variant="destructive">
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
  )
}

export function AdminPagination({
  onPageChange,
  page,
  pageSize,
  total,
}: {
  onPageChange: (page: number) => void
  page: number
  pageSize: number
  total: number
}) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="mt-4 flex flex-col gap-3 bg-transparent px-3 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="text-sm text-muted-foreground">
        第 {page} / {totalPages} 页，共 {total} 条记录
      </div>
      <div className="flex gap-2">
        <Button
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          size="sm"
          type="button"
          variant="outline"
        >
          <ChevronLeftIcon data-icon="inline-start" />
          上一页
        </Button>
        <Button
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
          size="sm"
          type="button"
          variant="outline"
        >
          下一页
          <ChevronRightIcon data-icon="inline-end" />
        </Button>
      </div>
    </div>
  )
}

export function AdminDangerDialog({
  busy,
  confirmLabel = '确认',
  description,
  onConfirm,
  onOpenChange,
  open,
  title,
}: {
  busy?: boolean
  confirmLabel?: string
  description: string
  onConfirm: () => void
  onOpenChange: (open: boolean) => void
  open: boolean
  title: string
}) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent size="sm">
        <AlertDialogHeader>
          <AlertDialogMedia className="bg-background">
            <AlertTriangleIcon />
          </AlertDialogMedia>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>取消</AlertDialogCancel>
          <AlertDialogAction
            disabled={busy}
            onClick={onConfirm}
            variant="destructive"
          >
            {busy ? '处理中…' : confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export function AdminItemMeta({
  items,
}: {
  items: Array<{ label: string; value: React.ReactNode }>
}) {
  return (
    <div className="grid gap-3 sm:grid-cols-2">
      {items.map((item) => (
        <div
          key={item.label}
          className="rounded-xl bg-transparent px-3 py-3 ring-1 ring-border/60"
        >
          <div className="text-xs text-muted-foreground">{item.label}</div>
          <div className="mt-1 text-xs font-medium">{item.value}</div>
        </div>
      ))}
    </div>
  )
}

export function AdminDividerLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="my-4">
      <Separator />
      <div className="-mt-2 text-center text-xs text-muted-foreground">{children}</div>
    </div>
  )
}
