import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { SaveIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { ADMIN_SELECT_CLASSNAME } from '#/lib/admin'
import { api } from '#/lib/api'
import { adminSettingsQueryOptions, queryKeys } from '#/lib/query-options'
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
  FieldDescription,
  FieldLabel,
  FieldTitle,
} from '#/components/ui/field'
import { Input } from '#/components/ui/input'
import { Switch } from '#/components/ui/switch'
import { Textarea } from '#/components/ui/textarea'

interface SettingFilterState {
  category: string
  key: string
  order: string
}

const EMPTY_FILTERS: SettingFilterState = {
  category: '',
  key: '',
  order: 'key_asc',
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

export function AdminSettingsPage() {
  const queryClient = useQueryClient()
  const { auth, handleUnauthorized } = useAdminSession('/admin/settings')
  const [draftFilters, setDraftFilters] = useState<SettingFilterState>(EMPTY_FILTERS)
  const [filters, setFilters] = useState<SettingFilterState>(EMPTY_FILTERS)
  const [draftValues, setDraftValues] = useState<Record<string, string>>({})

  const settingsQuery = useQuery({
    ...adminSettingsQueryOptions(auth.token ?? '', {
      category: filters.category.trim() || undefined,
      key: filters.key.trim() || undefined,
      limit: 100,
      offset: 0,
      order: filters.order,
    }),
    enabled: !!auth.token,
  })

  const updateMutation = useMutation({
    mutationFn: (input: { key: string; value: string }) =>
      api.updateAdminSetting(
        input.key,
        {
          value: input.value,
        },
        auth.token!,
      ),
  })

  const settings = settingsQuery.data?.items ?? []

  useEffect(() => {
    const nextDrafts = settings.reduce<Record<string, string>>((acc, setting) => {
      acc[setting.key] = setting.value ?? ''
      return acc
    }, {})

    setDraftValues(nextDrafts)
  }, [settings])

  async function invalidateSettingQueries() {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: queryKeys.adminSettingsPrefix(auth.token),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.publicSettings(),
      }),
    ])
  }

  async function handleSaveSetting(setting: SettingItem) {
    if (!auth.token) {
      return
    }

    try {
      await updateMutation.mutateAsync({
        key: setting.key,
        value: draftValues[setting.key] ?? '',
      })
      toast.success(`配置 ${setting.key} 已更新`)
      await invalidateSettingQueries()
    } catch (error) {
      if (handleUnauthorized(error)) {
        return
      }

      toast.error(error instanceof Error ? error.message : '更新配置失败')
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
      <AdminPageHeader
        title="设置管理"
      />

      <AdminSection
        title="配置面板"
      >
        {settingsQuery.isPending ? (
          <div className="rounded-2xl border border-border/70 bg-transparent">
            {Array.from({ length: 6 }).map((_, index) => (
              <div
                key={index}
                className="border-b border-border/60 p-4 last:border-b-0"
              >
                <div className="h-4 w-24 rounded bg-muted/40" />
                <div className="mt-3 h-4 w-full rounded bg-muted/30" />
                <div className="mt-2 h-4 w-9/12 rounded bg-muted/30" />
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

        {!settingsQuery.isPending && !settingsQuery.error && settings.length === 0 ? (
          <AdminEmptyState
            description="可以调整分类或 key 前缀筛选条件。"
            title="没有找到符合条件的配置"
          />
        ) : null}

        {!settingsQuery.isPending && !settingsQuery.error && settings.length > 0 ? (
          <div className="overflow-hidden bg-transparent">
            {settings.map((setting, index) => {
              const currentValue = draftValues[setting.key] ?? ''
              const savingCurrent =
                updateMutation.isPending &&
                updateMutation.variables?.key === setting.key
              const isBooleanSetting = isBooleanSettingValue(setting.value)
              const useTextarea = shouldUseTextarea(setting, currentValue)

              return (
                <AdminMotionItem
                  key={setting.key}
                  className="border-b border-border/60 p-5 last:border-b-0"
                  delay={index * 0.02}
                >
                  <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_220px]">
                    <div className="min-w-0 space-y-4">
                      <div className="min-w-0">
                        <div className="font-medium">
                          {setting.desc || setting.key}
                        </div>
                      </div>

                      {isBooleanSetting ? (
                        <Field orientation="horizontal">
                          <Switch
                            checked={asBooleanValue(currentValue)}
                            id={`setting-${setting.key}`}
                            onCheckedChange={(checked) =>
                              handleDraftChange(setting.key, checked ? 'true' : 'false')
                            }
                          />
                          <FieldContent>
                            <FieldLabel htmlFor={`setting-${setting.key}`}>
                              <FieldTitle>启用</FieldTitle>
                            </FieldLabel>
                          </FieldContent>
                        </Field>
                      ) : useTextarea ? (
                        <Field>
                          <FieldLabel htmlFor={`setting-${setting.key}`}>配置值</FieldLabel>
                          <Textarea
                            id={`setting-${setting.key}`}
                            onChange={(event) =>
                              handleDraftChange(setting.key, event.target.value)
                            }
                            placeholder="请输入新的配置值"
                            rows={6}
                            value={currentValue}
                          />
                        </Field>
                      ) : (
                        <Field>
                          <FieldLabel htmlFor={`setting-${setting.key}`}>配置值</FieldLabel>
                          <Input
                            id={`setting-${setting.key}`}
                            onChange={(event) =>
                              handleDraftChange(setting.key, event.target.value)
                            }
                            placeholder="请输入新的配置值"
                            type={setting.sensitive ? 'password' : 'text'}
                            value={currentValue}
                          />
                        </Field>
                      )}
                    </div>

                    <div className="flex items-end justify-end">
                      <div className="flex flex-wrap justify-end gap-2">
                        <Button
                          disabled={savingCurrent}
                          onClick={() => void handleSaveSetting(setting)}
                          size="sm"
                          type="button"
                        >
                          <SaveIcon data-icon="inline-start" />
                          {savingCurrent ? '保存中…' : '保存'}
                        </Button>
                      </div>
                    </div>
                  </div>
                </AdminMotionItem>
              )
            })}
          </div>
        ) : null}
      </AdminSection>
    </div>
  )
}
