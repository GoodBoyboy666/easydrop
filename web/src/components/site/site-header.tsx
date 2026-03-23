"use client"

import { Link, useLocation, useNavigate } from '@tanstack/react-router'
import {
  FileTextIcon,
  LogOutIcon,
  MessageSquareIcon,
  SettingsIcon,
  UserCircleIcon,
} from 'lucide-react'
import { useAuth } from '#/lib/auth'
import { getInitials } from '#/lib/format'
import { useSiteSettings } from '#/lib/site-settings'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Button } from '#/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu'

const navItems = [
  { label: '日志流', to: '/' as const },
  { label: '我的评论', to: '/me/comments' as const },
]

export function SiteHeader() {
  const auth = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const { allowRegister, siteDescription, siteName } = useSiteSettings()
  const user = auth.user

  return (
    <header className="sticky top-0 z-30 border-b border-border/80 bg-background/90 backdrop-blur">
      <div className="mx-auto flex w-full max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
        <div className="flex min-w-0 items-center gap-4">
          <Link
            to="/"
            className="flex items-center gap-3 rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <div className="flex size-10 items-center justify-center rounded-2xl bg-primary/12 text-primary ring-1 ring-primary/20">
              <FileTextIcon />
            </div>
            <div className="min-w-0">
              <div className="font-heading text-base font-semibold tracking-tight">
                {siteName}
              </div>
              <div className="truncate text-xs text-muted-foreground">
                {siteDescription}
              </div>
            </div>
          </Link>

          <nav className="hidden items-center gap-1 rounded-full border border-border/70 bg-card/80 p-1 md:flex">
            {navItems.map((item) => {
              const active = location.pathname === item.to
              return (
                <Button
                  key={item.to}
                  asChild
                  variant={active ? 'secondary' : 'ghost'}
                  size="sm"
                >
                  <Link to={item.to}>{item.label}</Link>
                </Button>
              )
            })}
          </nav>
        </div>

        <div className="flex items-center gap-2">
          {auth.status === 'loading' ? (
            <div className="text-sm text-muted-foreground">正在加载登录状态…</div>
          ) : null}

          {auth.status !== 'authenticated' ? (
            <>
              <Button asChild variant="ghost" size="sm">
                <Link search={{ redirect: '/' }} to="/login">
                  登录
                </Link>
              </Button>
              {allowRegister ? (
                <Button asChild size="sm">
                  <Link search={{ redirect: '/' }} to="/register">
                    注册
                  </Link>
                </Button>
              ) : null}
            </>
          ) : user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  className="rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  type="button"
                >
                  <Avatar size="lg">
                    <AvatarImage alt={user.nickname} src={user.avatar} />
                    <AvatarFallback>{getInitials(user.nickname)}</AvatarFallback>
                  </Avatar>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuGroup>
                  <div className="px-2 py-2">
                    <div className="font-medium">{user.nickname}</div>
                    <div className="text-xs text-muted-foreground">@{user.username}</div>
                  </div>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem onSelect={() => void navigate({ to: '/me' })}>
                    <UserCircleIcon data-icon="inline-start" />
                    个人信息
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onSelect={() => void navigate({ to: '/me/comments' })}
                  >
                    <MessageSquareIcon data-icon="inline-start" />
                    我的评论
                  </DropdownMenuItem>
                  {auth.isAdmin ? (
                    <DropdownMenuItem onSelect={() => void navigate({ to: '/admin' })}>
                      <SettingsIcon data-icon="inline-start" />
                      后台管理
                    </DropdownMenuItem>
                  ) : null}
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem
                    variant="destructive"
                    onSelect={() => {
                      auth.logout()
                      void navigate({ to: '/' })
                    }}
                  >
                    <LogOutIcon data-icon="inline-start" />
                    退出登录
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : null}
        </div>
      </div>
    </header>
  )
}
