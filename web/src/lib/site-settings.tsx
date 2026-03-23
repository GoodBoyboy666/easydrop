"use client"

import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { api, toPublicSettingsMap } from '#/lib/api'
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
  siteName: string
  siteUrl: string
}

const DEFAULT_SITE_NAME = 'EasyDrop'
const DEFAULT_SITE_DESCRIPTION = '一个轻量级日志说说平台'
const DEFAULT_SITE_OWNER = 'Your Name'
const DEFAULT_SITE_OWNER_DESCRIPTION = 'Do what you want to do.'

const SiteSettingsContext = createContext<SiteSettingsContextValue | null>(null)

function getSetting(settingsMap: PublicSettingsMap, key: string, fallback = '') {
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

export function SiteSettingsProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const [settingsMap, setSettingsMap] = useState<PublicSettingsMap>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void (async () => {
      try {
        const result = await api.getPublicSettings()

        setSettingsMap(toPublicSettingsMap(result))
        setError(null)
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : '加载站点配置失败')
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const value = useMemo<SiteSettingsContextValue>(() => {
    const siteName = getSetting(settingsMap, 'site.name', DEFAULT_SITE_NAME)
    const siteDescription = getSetting(
      settingsMap,
      'site.description',
      DEFAULT_SITE_DESCRIPTION
    )
    const siteOwner = getSetting(settingsMap, 'site.owner', siteName || DEFAULT_SITE_OWNER)
    const siteOwnerDescription = getSetting(
      settingsMap,
      'site.owner.description',
      siteDescription || DEFAULT_SITE_OWNER_DESCRIPTION
    )
    const siteUrl = getSetting(settingsMap, 'site.url', '')
    const siteAnnouncement = getSetting(settingsMap, 'site.announcement', '')
    const siteBackground = getSetting(settingsMap, 'site.background', '')

    return {
      allowRegister: parseBooleanSetting(settingsMap['site.allow_register'], true),
      error,
      loading,
      siteOwner,
      siteOwnerDescription,
      settingsMap,
      siteAnnouncement,
      siteBackground,
      siteBackgroundImageUrl: resolveBackgroundUrl(siteBackground),
      siteDescription,
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
