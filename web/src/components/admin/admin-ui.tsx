import { Link } from '@tanstack/react-router'
import {
  AlertTriangleIcon,
  ArrowLeftIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ShieldCheckIcon,
} from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { AdminNavItem } from '#/lib/admin'
import { adminNavItems } from '#/lib/admin'
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
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from '#/components/ui/sidebar'
import { Skeleton } from '#/components/ui/skeleton'
import { TooltipProvider } from '#/components/ui/tooltip'

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
        <Card className="border border-border/70 bg-card shadow-sm backdrop-blur-sm">
          <CardHeader>
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </CardHeader>
          {actions ? (
            <CardContent className="flex gap-2">{actions}</CardContent>
          ) : null}
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
    <TooltipProvider delayDuration={120}>
      <SidebarProvider>
        <Sidebar collapsible="icon" variant="inset">
          <SidebarHeader className="gap-3 border-b border-sidebar-border p-3">
            <motion.div {...getAdminMotionProps(prefersReducedMotion)}>
              <div className="flex items-center gap-2 rounded-lg group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-2 group-data-[collapsible=icon]:py-2">
                <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
                  <ShieldCheckIcon className="size-4" />
                </div>
                <div className="min-w-0 overflow-hidden group-data-[collapsible=icon]:hidden">
                  <div className="truncate whitespace-nowrap text-sm font-semibold tracking-tight">
                    后台管理
                  </div>
                  <div className="mt-0.5 truncate whitespace-nowrap text-xs text-sidebar-foreground/70">
                    管理内容与站点设置
                  </div>
                </div>
              </div>
            </motion.div>
          </SidebarHeader>

          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupLabel>菜单</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu className='gap-1'>
                  {adminNavItems.map((item) => (
                    <AdminNavButton
                      key={item.to}
                      active={isAdminNavActive(activePath, item.to)}
                      item={item}
                    />
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>

          <SidebarFooter className="border-t border-sidebar-border p-2">
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild tooltip="返回站点">
                  <Link search={{ content: undefined }} to="/">
                    <ArrowLeftIcon />
                    <span>返回站点</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
          <SidebarRail />
        </Sidebar>

        <SidebarInset className="bg-white">
          <div className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
            <motion.div
              className="min-w-0"
              {...getAdminMotionProps(prefersReducedMotion, 0.05)}
            >
              {children}
            </motion.div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}

function isAdminNavActive(activePath: string, navPath: AdminNavItem['to']) {
  if (activePath === navPath) {
    return true
  }

  if (navPath === '/admin') {
    return false
  }

  return activePath.startsWith(`${navPath}/`)
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
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={active} tooltip={item.label}>
        <Link to={item.to}>
          <Icon />
          <span>{item.label}</span>
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
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
      <div className="flex items-start gap-2">
        <SidebarTrigger className="mt-0.5" />
        <div>
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            {title}
          </h1>
          {description ? (
            <p className="mt-1 text-sm text-muted-foreground">{description}</p>
          ) : null}
        </div>
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
      <Card className="border border-border/70 bg-card shadow-sm backdrop-blur-sm">
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
      <Card className="border border-border/70 bg-card shadow-sm backdrop-blur-sm">
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
          <Card className="border border-border/70 bg-card shadow-sm backdrop-blur-sm">
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
    <div className="border-t border-border/70 mt-4 flex flex-col gap-3 bg-transparent px-3 py-3 sm:flex-row sm:items-center sm:justify-between">
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
      <div className="-mt-2 text-center text-xs text-muted-foreground">
        {children}
      </div>
    </div>
  )
}
