"use client"

import { Link, useLocation, useNavigate } from '@tanstack/react-router'
import {
  FileTextIcon,
  LaptopMinimalIcon,
  LogOutIcon,
  MenuIcon,
  MessageSquareIcon,
  MoonStarIcon,
  SearchIcon,
  SettingsIcon,
  SunIcon,
  UserCircleIcon,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { useAuth } from '#/lib/auth'
import { getInitials } from '#/lib/format'
import { useSiteSettings } from '#/lib/site-settings'
import { useTheme } from '#/lib/theme'
import { cn } from '#/lib/utils'
import { Avatar, AvatarFallback, AvatarImage } from '#/components/ui/avatar'
import { Button } from '#/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu'
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from '#/components/ui/navigation-menu'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '#/components/ui/sheet'

const navItems = [
  { label: '日志', to: '/' as const },
  { label: '我的评论', to: '/me/comments' as const },
]

export function SiteHeader() {
  const auth = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const { allowRegister, siteDescription, siteName } = useSiteSettings()
  const { resolvedTheme, setTheme, theme } = useTheme()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [searchValue, setSearchValue] = useState('')
  const user = auth.user
  const currentSearchContent =
    typeof location.search === 'object' &&
    location.search !== null &&
    'content' in location.search &&
    typeof location.search.content === 'string'
      ? location.search.content
      : ''

  useEffect(() => {
    setSearchValue(currentSearchContent)
  }, [currentSearchContent])

  function closeMobileMenu() {
    setMobileMenuOpen(false)
  }

  function submitSearch(options?: { closeMobileMenu?: boolean }) {
    const normalizedContent = searchValue.trim()

    if (options?.closeMobileMenu) {
      closeMobileMenu()
    }

    void navigate({
      to: '/',
      search: normalizedContent ? { content: normalizedContent } : {},
    })
  }

  const themeLabel =
    theme === 'system'
      ? '跟随系统'
      : resolvedTheme === 'dark'
        ? '暗色模式'
        : '亮色模式'

  const ThemeIcon =
    theme === 'system'
      ? LaptopMinimalIcon
      : resolvedTheme === 'dark'
        ? MoonStarIcon
        : SunIcon

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

          <NavigationMenu
            viewport={false}
            className="hidden flex-none md:flex"
          >
            <NavigationMenuList className="rounded-xl p-1 gap-3 bg-transparent">
              {navItems.map((item) => {
                const active = location.pathname === item.to
                return (
                  <NavigationMenuItem key={item.to}>
                    <NavigationMenuLink
                      asChild
                      className={cn(
                        'rounded-lg px-3 py-2 font-medium shadow-none ring-0',
                        active
                          ? 'bg-primary text-primary-foreground hover:bg-primary/90 focus:bg-primary/90'
                          : 'hover:bg-muted hover:text-foreground focus:bg-muted focus:text-foreground'
                      )}
                    >
                      <Link to={item.to}>{item.label}</Link>
                    </NavigationMenuLink>
                  </NavigationMenuItem>
                )
              })}
            </NavigationMenuList>
          </NavigationMenu>
        </div>

        <div className="flex min-w-0 items-center gap-2">
          <form
            className="relative hidden md:block"
            role="search"
            onSubmit={(event) => {
              event.preventDefault()
              submitSearch()
            }}
          >
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              aria-label="搜索日志内容"
              className="h-9 w-56 rounded-xl pl-9 lg:w-72"
              onChange={(event) => setSearchValue(event.target.value)}
              placeholder="搜索日志内容"
              type="search"
              value={searchValue}
            />
          </form>

          {auth.status === 'loading' ? (
            <div className="text-sm text-muted-foreground">正在加载登录状态…</div>
          ) : null}

          {auth.status !== 'authenticated' ? (
            <>
              <Button asChild variant="ghost" size="sm" className="hidden md:inline-flex">
                <Link search={{ redirect: '/' }} to="/login">
                  登录
                </Link>
              </Button>
              {allowRegister ? (
                <Button asChild size="sm" className="hidden md:inline-flex">
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
                  className="hidden rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring md:inline-flex"
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

          <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="md:hidden"
                aria-label="打开导航菜单"
              >
                <MenuIcon />
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[min(22rem,calc(100vw-1rem))] p-0 md:hidden">
              <SheetHeader className="pr-12">
                <SheetTitle>{siteName}</SheetTitle>
                <SheetDescription>{siteDescription}</SheetDescription>
              </SheetHeader>

              <div className="flex flex-1 flex-col gap-4 px-4 pb-4">
                <form
                  className="relative"
                  role="search"
                  onSubmit={(event) => {
                    event.preventDefault()
                    submitSearch({ closeMobileMenu: true })
                  }}
                >
                  <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    aria-label="搜索日志内容"
                    className="h-10 rounded-xl pl-9"
                    onChange={(event) => setSearchValue(event.target.value)}
                    placeholder="搜索日志内容"
                    type="search"
                    value={searchValue}
                  />
                </form>

                <div className="flex flex-col gap-2">
                  {navItems.map((item) => {
                    const active = location.pathname === item.to
                    return (
                      <Button
                        key={item.to}
                        asChild
                        variant={active ? 'default' : 'ghost'}
                        className="w-full justify-start"
                      >
                        <Link
                          to={item.to}
                          onClick={closeMobileMenu}
                        >
                          {item.label}
                        </Link>
                      </Button>
                    )
                  })}
                </div>

                <Separator />

                {auth.status === 'loading' ? (
                  <div className="text-sm text-muted-foreground">正在加载登录状态…</div>
                ) : null}

                {auth.status !== 'authenticated' ? (
                  <div className="flex flex-col gap-2">
                    <Button asChild className="w-full justify-start">
                      <Link search={{ redirect: '/' }} to="/login" onClick={closeMobileMenu}>
                        登录
                      </Link>
                    </Button>
                    {allowRegister ? (
                      <Button
                        asChild
                        variant="secondary"
                        className="w-full justify-start"
                      >
                        <Link
                          search={{ redirect: '/' }}
                          to="/register"
                          onClick={closeMobileMenu}
                        >
                          注册
                        </Link>
                      </Button>
                    ) : (
                      <div className="text-sm text-muted-foreground">
                        当前站点未开放注册
                      </div>
                    )}
                  </div>
                ) : user ? (
                  <div className="flex flex-col gap-4">
                    <div className="flex items-center gap-3 rounded-2xl border border-border/70 bg-card/70 px-3 py-3">
                      <Avatar size="lg">
                        <AvatarImage alt={user.nickname} src={user.avatar} />
                        <AvatarFallback>{getInitials(user.nickname)}</AvatarFallback>
                      </Avatar>
                      <div className="min-w-0">
                        <div className="truncate font-medium">{user.nickname}</div>
                        <div className="truncate text-xs text-muted-foreground">
                          @{user.username}
                        </div>
                      </div>
                    </div>

                    <div className="flex flex-col gap-2">
                      <Button
                        asChild
                        variant="ghost"
                        className="w-full justify-start"
                      >
                        <Link to="/me" onClick={closeMobileMenu}>
                          <UserCircleIcon data-icon="inline-start" />
                          个人信息
                        </Link>
                      </Button>
                      <Button
                        asChild
                        variant="ghost"
                        className="w-full justify-start"
                      >
                        <Link to="/me/comments" onClick={closeMobileMenu}>
                          <MessageSquareIcon data-icon="inline-start" />
                          我的评论
                        </Link>
                      </Button>
                      {auth.isAdmin ? (
                        <Button
                          asChild
                          variant="ghost"
                          className="w-full justify-start"
                        >
                          <Link to="/admin" onClick={closeMobileMenu}>
                            <SettingsIcon data-icon="inline-start" />
                            后台管理
                          </Link>
                        </Button>
                      ) : null}
                    </div>

                    <Separator />

                    <Button
                      variant="destructive"
                      className="w-full justify-start"
                      onClick={() => {
                        closeMobileMenu()
                        auth.logout()
                        void navigate({ to: '/' })
                      }}
                    >
                      <LogOutIcon data-icon="inline-start" />
                      退出登录
                    </Button>
                  </div>
                ) : null}
              </div>
            </SheetContent>
          </Sheet>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                type="button"
                aria-label={`切换主题，当前为${themeLabel}`}
                title={`切换主题，当前为${themeLabel}`}
                className="shrink-0"
              >
                <ThemeIcon />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              <DropdownMenuLabel>主题模式</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuRadioGroup
                value={theme}
                onValueChange={(value) => setTheme(value as 'light' | 'dark' | 'system')}
              >
                <DropdownMenuRadioItem value="light">
                  <SunIcon data-icon="inline-start" />
                  亮色模式
                </DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="dark">
                  <MoonStarIcon data-icon="inline-start" />
                  暗色模式
                </DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="system">
                  <LaptopMinimalIcon data-icon="inline-start" />
                  跟随系统
                </DropdownMenuRadioItem>
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
