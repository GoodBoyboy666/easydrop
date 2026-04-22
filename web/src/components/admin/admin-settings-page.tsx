import { useQuery, useQueryClient } from '@tanstack/react-query'
import { SaveIcon } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import { adminSettingsQueryOptions } from '#/lib/query-options'
import { invalidateAdminSettingQueries } from '#/lib/query-invalidation'
import type { SettingItem } from '#/lib/types'
import {
  AdminEmptyState,
  AdminErrorAlert,
  AdminMotionItem,
  AdminPageHeader,
  AdminSection,
} from '#/components/admin/admin-ui'
import { useAdminSession } from '#/components/admin/use-admin-session'
import { Button } from '#/components/ui/button'
import {
  Field,
  FieldContent,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Separator } from '#/components/ui/separator'
import { Switch } from '#/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Textarea } from '#/components/ui/textarea'

const SETTING_CATEGORY_LABELS: Record<string, string> = {
  auth: '认证',
  site: '站点',
  storage: '存储',
  system: '系统',
}

function isBooleanSettingValue(value: string | null | undefined) {
  const normalized = value?.trim().toLowerCase()
  return normalized === 'true' || normalized === 'false'
}

function asBooleanValue(value: string | null | undefined) {
  return value?.trim().toLowerCase() === 'true'
}

function shouldUseTextarea(setting: SettingItem, value: string) {
  return (
    value.includes('\n') ||
    value.length > 120 ||
    setting.key.includes('announcement') ||
    setting.desc.includes('公告') ||
    setting.desc.includes('说明')
  )
}

function normalizeCategory(category: string | null | undefined) {
  return category?.trim() || 'other'
}

function getCategoryLabel(category: string) {
  return SETTING_CATEGORY_LABELS[category] ?? category
}

export function AdminSettingsPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/settings')
  const [draftValues, setDraftValues] = useState<Record<string, string>>({})
  const [isSavingAll, setIsSavingAll] = useState(false)
  const [activeCategory, setActiveCategory] = useState('')

  const settingsQuery = useQuery({
    ...adminSettingsQueryOptions({
      order: 'key_asc',
      page: 1,
      size: 100,
    }),
    enabled: auth.status === 'authenticated',
  })

  const settings = settingsQuery.data?.items ?? []
  const groupedSettings = useMemo(() => {
    const groups = new Map<string, SettingItem[]>()

    for (const setting of settings) {
      const category = normalizeCategory(setting.category)
      const items = groups.get(category)

      if (items) {
        items.push(setting)
        continue
      }

      groups.set(category, [setting])
    }

    return Array.from(groups.entries()).map(([category, items]) => ({
      category,
      items,
      label: getCategoryLabel(category),
    }))
  }, [settings])
  const changedSettings = useMemo(
    () =>
      settings.filter(
        (setting) => (draftValues[setting.key] ?? '') !== setting.value,
      ),
    [draftValues, settings],
  )
  const changedCountByCategory = useMemo(
    () =>
      changedSettings.reduce<Record<string, number>>((acc, setting) => {
        const category = normalizeCategory(setting.category)
        acc[category] = (acc[category] ?? 0) + 1
        return acc
      }, {}),
    [changedSettings],
  )
  const currentCategory = activeCategory || groupedSettings[0]?.category || ''

  useEffect(() => {
    const nextDrafts = settings.reduce<Record<string, string>>(
      (acc, setting) => {
        acc[setting.key] = setting.value
        return acc
      },
      {},
    )

    setDraftValues(nextDrafts)
  }, [settings])

  useEffect(() => {
    if (groupedSettings.length === 0) {
      setActiveCategory('')
      return
    }

    if (!groupedSettings.some((group) => group.category === activeCategory)) {
      setActiveCategory(groupedSettings[0].category)
    }
  }, [activeCategory, groupedSettings])

  async function invalidateSettingQueries() {
    await invalidateAdminSettingQueries(queryClient)
  }

  async function handleSaveAllSettings() {
    if (auth.status !== 'authenticated') {
      return
    }

    if (changedSettings.length === 0) {
      return
    }

    setIsSavingAll(true)

    try {
      const results = await Promise.allSettled(
        changedSettings.map((setting) =>
          api.updateAdminSetting(setting.key, {
            value: draftValues[setting.key] ?? '',
          }),
        ),
      )

      const rejected = results.filter((result) => result.status === 'rejected')

      for (const result of rejected) {
        if (handleUnauthorized(result.reason)) {
          return
        }
      }

      if (rejected.length > 0) {
        const firstError = rejected[0].reason
        toast.error(
          firstError instanceof Error
            ? firstError.message
            : `部分配置更新失败（成功 ${results.length - rejected.length} / ${results.length}）`,
        )
        return
      }

      toast.success(`已更新 ${results.length} 项配置`)
      await invalidateSettingQueries()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '更新配置失败')
    } finally {
      setIsSavingAll(false)
    }
  }

  function handleDraftChange(settingKey: string, value: string) {
    setDraftValues((current) => ({
      ...current,
      [settingKey]: value,
    }))
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader title="设置管理" />

      <AdminSection title="配置面板">
        {settingsQuery.isPending ? (
          <div className="rounded-2xl border border-border/70 bg-transparent">
            {Array.from({ length: 6 }).map((_, index) => (
              <div key={index}>
                <div className="p-4">
                  <div className="h-4 w-24 rounded bg-muted/40" />
                  <div className="mt-3 h-4 w-full rounded bg-muted/30" />
                  <div className="mt-2 h-4 w-9/12 rounded bg-muted/30" />
                </div>
                {index < 5 ? (
                  <Separator className="bg-border/80 data-horizontal:h-0.5" />
                ) : null}
              </div>
            ))}
          </div>
        ) : null}

        {settingsQuery.error instanceof Error ? (
          <AdminErrorAlert
            description={settingsQuery.error.message}
            title="配置面板读取失败"
          />
        ) : null}

        {!settingsQuery.isPending &&
        !settingsQuery.error &&
        settings.length === 0 ? (
          <AdminEmptyState
            description="可以调整分类或 key 前缀筛选条件。"
            title="没有找到符合条件的配置"
          />
        ) : null}

        {!settingsQuery.isPending &&
        !settingsQuery.error &&
        settings.length > 0 ? (
          <Tabs
            onValueChange={setActiveCategory}
            value={currentCategory}
          >
            <TabsList className="gap-3">
              {groupedSettings.map((group) => {
                const changedCount = changedCountByCategory[group.category] ?? 0

                return (
                  <TabsTrigger key={group.category} value={group.category}>
                    <span>{group.label}</span>
                    <span className="text-xs text-muted-foreground">
                      {changedCount > 0 ? ` · ${changedCount} 已改` : ''}
                    </span>
                  </TabsTrigger>
                )
              })}
            </TabsList>

            {groupedSettings.map((group) => (
              <TabsContent key={group.category} value={group.category}>
                <div className="overflow-hidden rounded-2xl bg-transparent">
                  {group.items.map((setting, index) => {
                    const currentValue = draftValues[setting.key] ?? ''
                    const isBooleanSetting = isBooleanSettingValue(
                      setting.value,
                    )
                    const useTextarea = shouldUseTextarea(setting, currentValue)

                    return (
                      <div key={setting.key}>
                        <AdminMotionItem className="px-2 py-5" delay={index * 0.02}>
                          <div className="min-w-0 space-y-4">
                            <div className="min-w-0 space-y-1">
                              <div className="font-medium">
                                {setting.desc || setting.key}
                              </div>
                              {setting.desc ? (
                                <div className="text-xs text-muted-foreground">
                                  {setting.key}
                                </div>
                              ) : null}
                            </div>

                            {isBooleanSetting ? (
                              <Field orientation="horizontal">
                                <Switch
                                  checked={asBooleanValue(currentValue)}
                                  id={`setting-${setting.key}`}
                                  onCheckedChange={(checked) =>
                                    handleDraftChange(
                                      setting.key,
                                      checked ? 'true' : 'false',
                                    )
                                  }
                                />
                                <FieldContent>
                                  <FieldLabel
                                    htmlFor={`setting-${setting.key}`}
                                  >
                                    <FieldTitle>启用</FieldTitle>
                                  </FieldLabel>
                                </FieldContent>
                              </Field>
                            ) : useTextarea ? (
                              <Field>
                                <FieldLabel htmlFor={`setting-${setting.key}`}>
                                  配置值
                                </FieldLabel>
                                <Textarea
                                  id={`setting-${setting.key}`}
                                  onChange={(event) =>
                                    handleDraftChange(
                                      setting.key,
                                      event.target.value,
                                    )
                                  }
                                  placeholder="请输入新的配置值"
                                  rows={6}
                                  value={currentValue}
                                />
                              </Field>
                            ) : (
                              <Field>
                                <FieldLabel htmlFor={`setting-${setting.key}`}>
                                  配置值
                                </FieldLabel>
                                <Input
                                  id={`setting-${setting.key}`}
                                  onChange={(event) =>
                                    handleDraftChange(
                                      setting.key,
                                      event.target.value,
                                    )
                                  }
                                  placeholder="请输入新的配置值"
                                  type={setting.sensitive ? 'password' : 'text'}
                                  value={currentValue}
                                />
                              </Field>
                            )}
                          </div>
                        </AdminMotionItem>

                        {index < group.items.length - 1 ? (
                          <Separator className="bg-border/80 data-horizontal:h-0.5" />
                        ) : null}
                      </div>
                    )
                  })}
                </div>
              </TabsContent>
            ))}

            <AdminMotionItem
              className="border-t border-border/70 pt-4"
              delay={settings.length * 0.02}
            >
              <div className="flex flex-wrap items-center justify-end gap-3">
                <span className="text-sm text-muted-foreground">
                  {changedSettings.length > 0
                    ? `已修改 ${changedSettings.length} 项`
                    : '暂无未保存修改'}
                </span>
                <Button
                  disabled={isSavingAll || changedSettings.length === 0}
                  onClick={() => void handleSaveAllSettings()}
                  type="button"
                >
                  <SaveIcon data-icon="inline-start" />
                  {isSavingAll ? '保存中…' : '保存'}
                </Button>
              </div>
            </AdminMotionItem>
          </Tabs>
        ) : null}
      </AdminSection>
    </div>
  )
}
