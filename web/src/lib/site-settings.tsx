'use client'

import { createContext, useContext, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { publicSettingsMapQueryOptions } from '#/lib/query-options'
import type { PublicSettingsMap } from '#/lib/types'

interface SiteSettingsContextValue {
  allowRegister: boolean
  error: string | null
  loading: boolean
  siteOwner: string
  siteOwnerDescription: string
  settingsMap: PublicSettingsMap
  siteAnnouncement: string
  siteBackground: string
  siteBackgroundImageUrl: string | null
  siteDescription: string
  siteFavicon: string
  siteFaviconUrl: string
  siteName: string
  siteUrl: string
}

const DEFAULT_SITE_NAME = 'EasyDrop'
const DEFAULT_SITE_DESCRIPTION = '一个轻量级日志说说平台'
const DEFAULT_SITE_OWNER = 'Your Name'
const DEFAULT_SITE_OWNER_DESCRIPTION = 'Do what you want to do.'
const DEFAULT_SITE_FAVICON = '/favicon.ico'

const SiteSettingsContext = createContext<SiteSettingsContextValue | null>(null)

function getSetting(
  settingsMap: PublicSettingsMap,
  key: string,
  fallback = '',
) {
  const value = settingsMap[key]
  return typeof value === 'string' ? value : fallback
}

function parseBooleanSetting(value: string | undefined, fallback = true) {
  if (value == null || value.trim() === '') {
    return fallback
  }

  return value.trim().toLowerCase() === 'true'
}

function resolveBackgroundUrl(value: string) {
  const trimmedValue = value.trim()

  if (!trimmedValue) {
    return null
  }

  try {
    if (typeof window === 'undefined') {
      return /^https?:\/\//i.test(trimmedValue) || trimmedValue.startsWith('/')
        ? trimmedValue
        : null
    }

    const url = new URL(trimmedValue, window.location.origin)

    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return null
    }

    return url.toString()
  } catch {
    return null
  }
}

function resolveFaviconUrl(value: string) {
  const trimmedValue = value.trim()

  if (!trimmedValue) {
    return DEFAULT_SITE_FAVICON
  }

  try {
    if (typeof window === 'undefined') {
      return /^https?:\/\//i.test(trimmedValue) || trimmedValue.startsWith('/')
        ? trimmedValue
        : DEFAULT_SITE_FAVICON
    }

    const url = new URL(trimmedValue, window.location.origin)

    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return DEFAULT_SITE_FAVICON
    }

    return url.toString()
  } catch {
    return DEFAULT_SITE_FAVICON
  }
}

export function SiteSettingsProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const settingsQuery = useQuery(publicSettingsMapQueryOptions())
  const settingsMap: PublicSettingsMap = settingsQuery.data ?? {}
  const loading = settingsQuery.isPending
  const error =
    settingsQuery.error instanceof Error ? settingsQuery.error.message : null

  const value = useMemo<SiteSettingsContextValue>(() => {
    const siteName = getSetting(settingsMap, 'site.name', DEFAULT_SITE_NAME)
    const siteDescription = getSetting(
      settingsMap,
      'site.description',
      DEFAULT_SITE_DESCRIPTION,
    )
    const siteOwner = getSetting(
      settingsMap,
      'site.owner',
      siteName || DEFAULT_SITE_OWNER,
    )
    const siteOwnerDescription = getSetting(
      settingsMap,
      'site.owner.description',
      siteDescription || DEFAULT_SITE_OWNER_DESCRIPTION,
    )
    const siteUrl = getSetting(settingsMap, 'site.url', '')
    const siteAnnouncement = getSetting(settingsMap, 'site.announcement', '')
    const siteBackground = getSetting(settingsMap, 'site.background', '')
    const siteFavicon = getSetting(settingsMap, 'site.favicon', '')

    return {
      allowRegister: parseBooleanSetting(
        settingsMap['site.allow_register'],
        true,
      ),
      error,
      loading,
      siteOwner,
      siteOwnerDescription,
      settingsMap,
      siteAnnouncement,
      siteBackground,
      siteBackgroundImageUrl: resolveBackgroundUrl(siteBackground),
      siteDescription,
      siteFavicon,
      siteFaviconUrl: resolveFaviconUrl(siteFavicon),
      siteName,
      siteUrl,
    }
  }, [error, loading, settingsMap])

  return (
    <SiteSettingsContext.Provider value={value}>
      {children}
    </SiteSettingsContext.Provider>
  )
}

export function useSiteSettings() {
  const context = useContext(SiteSettingsContext)

  if (!context) {
    throw new Error('useSiteSettings must be used within SiteSettingsProvider')
  }

  return context
}
