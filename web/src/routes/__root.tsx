import { QueryClientProvider } from '@tanstack/react-query'
import { Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { PhotoProvider } from 'react-photo-view'
import { AuthProvider } from '#/lib/auth'
import { getQueryClient } from '#/lib/query-client'
import { SiteSettingsProvider, useSiteSettings } from '#/lib/site-settings'
import { ThemeProvider } from '#/lib/theme'
import { SiteFooter } from '#/components/site/site-footer'
import { SiteHeader } from '#/components/site/site-header'
import { Toaster } from '#/components/ui/sonner'

export const Route = createRootRoute({
  component: RootApp,
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
