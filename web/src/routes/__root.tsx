import { QueryClientProvider } from '@tanstack/react-query'
import { Link, Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { CompassIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import type { Transition } from 'motion/react'
import { PhotoProvider } from 'react-photo-view'
import { AuthProvider } from '#/lib/auth'
import { getQueryClient } from '#/lib/query-client'
import { SiteSettingsProvider, useSiteSettings } from '#/lib/site-settings'
import { ThemeProvider } from '#/lib/theme'
import { SiteFooter } from '#/components/site/site-footer'
import { SiteHeader } from '#/components/site/site-header'
import { Button } from '#/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card'
import { Toaster } from '#/components/ui/sonner'

export const Route = createRootRoute({
  component: RootApp,
  notFoundComponent: NotFoundPage,
})

function RootApp() {
  const queryClient = getQueryClient()

  return (
    <QueryClientProvider client={queryClient}>
      <PhotoProvider>
        <ThemeProvider>
          <AuthProvider>
            <SiteSettingsProvider>
              <AppShell>
                <Outlet />
              </AppShell>
              <TanStackDevtools
                config={{
                  position: 'bottom-right',
                }}
                plugins={[
                  {
                    name: 'Tanstack Router',
                    render: <TanStackRouterDevtoolsPanel />,
                  },
                ]}
              />
            </SiteSettingsProvider>
          </AuthProvider>
        </ThemeProvider>
      </PhotoProvider>
    </QueryClientProvider>
  )
}

function AppShell({ children }: { children: React.ReactNode }) {
  const { siteBackgroundImageUrl } = useSiteSettings()

  return (
    <div
      className="relative min-h-screen bg-background text-foreground"
      style={
        siteBackgroundImageUrl
          ? {
              backgroundImage: `url("${siteBackgroundImageUrl}")`,
              backgroundPosition: 'center',
              backgroundRepeat: 'no-repeat',
              backgroundSize: 'cover',
            }
          : undefined
      }
    >
      <div className="relative flex min-h-screen flex-col">
        <SiteHeader />
        <main className="flex-1">{children}</main>
        <SiteFooter />
        <Toaster closeButton richColors position="top-right" />
      </div>
    </div>
  )
}

function NotFoundPage() {
  const prefersReducedMotion = useReducedMotion()
  const pageTransition: Transition = prefersReducedMotion
    ? { duration: 0 }
    : { duration: 0.34, ease: 'easeOut' }

  const sectionTransition = (delay: number): Transition =>
    prefersReducedMotion
      ? { duration: 0 }
      : {
          duration: 0.3,
          ease: 'easeOut',
          delay,
        }

  return (
    <motion.div
      animate={{ opacity: 1, y: 0 }}
      className="mx-auto flex min-h-[calc(100vh-9rem)] w-full max-w-7xl items-center justify-center px-4 py-8 sm:px-6 lg:px-8"
      initial={prefersReducedMotion ? false : { opacity: 0, y: 20 }}
      transition={pageTransition}
    >
      <motion.div
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="flex w-full justify-center"
        initial={prefersReducedMotion ? false : { opacity: 0, y: 16, scale: 0.98 }}
        transition={sectionTransition(0.08)}
      >
        <Card className="w-full max-w-xl border border-border/70 bg-card/95 px-4 py-8 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-2xl">
              <CompassIcon className="size-6" />
              404 页面不存在
            </CardTitle>
            <CardDescription>
              你访问的链接可能已失效，或页面已被移动。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <motion.div
              animate={{ opacity: 1, y: 0 }}
              className="flex flex-col gap-3 sm:flex-row"
              initial={prefersReducedMotion ? false : { opacity: 0, y: 10 }}
              transition={sectionTransition(0.16)}
            >
              <Button asChild className="sm:flex-1">
                <Link to="/">回到首页</Link>
              </Button>
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
