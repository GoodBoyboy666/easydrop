import { HeadContent, Scripts, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { AuthProvider } from '#/lib/auth'
import { SiteSettingsProvider, useSiteSettings } from '#/lib/site-settings'
import { ThemeProvider } from '#/lib/theme'
import { SiteFooter } from '#/components/site/site-footer'
import { SiteHeader } from '#/components/site/site-header'
import { Toaster } from '#/components/ui/sonner'

import appCss from '../styles.css?url'

const themeScript = `
(() => {
  try {
    const storageKey = 'easydrop-theme'
    const storedTheme = window.localStorage.getItem(storageKey)
    const themeMode =
      storedTheme === 'light' || storedTheme === 'dark' || storedTheme === 'system'
        ? storedTheme
        : 'system'
    const resolvedTheme =
      themeMode === 'light' || themeMode === 'dark'
        ? themeMode
        : window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light'

    document.documentElement.dataset.themeMode = themeMode
    document.documentElement.classList.toggle('dark', resolvedTheme === 'dark')
    document.documentElement.style.colorScheme = resolvedTheme
  } catch {
    const fallbackTheme = window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'

    document.documentElement.dataset.themeMode = 'system'
    document.documentElement.classList.toggle('dark', fallbackTheme === 'dark')
    document.documentElement.style.colorScheme = fallbackTheme
  }
})()
`

export const Route = createRootRoute({
  head: () => ({
    meta: [
      {
        charSet: 'utf-8',
      },
      {
        name: 'viewport',
        content: 'width=device-width, initial-scale=1',
      },
      {
        title: 'EasyDrop',
      },
      {
        name: 'description',
        content: '一个采用双栏结构的中文日志发布网站骨架。',
      },
    ],
    links: [
      {
        rel: 'stylesheet',
        href: appCss,
      },
    ],
  }),
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <head>
        <HeadContent />
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body className="font-sans antialiased [overflow-wrap:anywhere]">
        <ThemeProvider>
          <AuthProvider>
            <SiteSettingsProvider>
              <AppShell>{children}</AppShell>
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
              <Scripts />
            </SiteSettingsProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
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
