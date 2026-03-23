"use client"

import { createContext, useContext, useEffect, useMemo, useState } from 'react'

export type ThemeMode = 'light' | 'dark' | 'system'
type ResolvedTheme = 'light' | 'dark'

interface ThemeContextValue {
  isDark: boolean
  resolvedTheme: ResolvedTheme
  setTheme: (theme: ThemeMode) => void
  theme: ThemeMode
}

const THEME_STORAGE_KEY = 'easydrop-theme'
const ThemeContext = createContext<ThemeContextValue | null>(null)

function getSystemTheme(): ResolvedTheme {
  if (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-color-scheme: dark)').matches
  ) {
    return 'dark'
  }

  return 'light'
}

function getInitialTheme(): ThemeMode {
  if (typeof window !== 'undefined') {
    try {
      const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY)

      if (
        storedTheme === 'light' ||
        storedTheme === 'dark' ||
        storedTheme === 'system'
      ) {
        return storedTheme
      }
    } catch {
      // ignore localStorage read errors
    }
  }

  return 'system'
}

function applyTheme(theme: ResolvedTheme) {
  const root = document.documentElement
  root.classList.toggle('dark', theme === 'dark')
  root.style.colorScheme = theme
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<ThemeMode>(getInitialTheme)
  const [systemTheme, setSystemTheme] = useState<ResolvedTheme>(getSystemTheme)

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const syncSystemTheme = () =>
      setSystemTheme(mediaQuery.matches ? 'dark' : 'light')

    syncSystemTheme()
    mediaQuery.addEventListener('change', syncSystemTheme)

    return () => {
      mediaQuery.removeEventListener('change', syncSystemTheme)
    }
  }, [])

  const resolvedTheme = theme === 'system' ? systemTheme : theme

  useEffect(() => {
    applyTheme(resolvedTheme)
  }, [resolvedTheme])

  useEffect(() => {
    const root = document.documentElement
    root.dataset.themeMode = theme

    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, theme)
    } catch {
      // ignore localStorage write errors
    }
  }, [theme])

  const value = useMemo<ThemeContextValue>(
    () => ({
      isDark: resolvedTheme === 'dark',
      resolvedTheme,
      setTheme: setThemeState,
      theme,
    }),
    [resolvedTheme, theme]
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const context = useContext(ThemeContext)

  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider')
  }

  return context
}
