import { useEffect, useId, useRef, useState } from 'react'
import type { CaptchaConfigResult, CaptchaInput } from '#/lib/types'
import { Alert, AlertDescription, AlertTitle } from '#/components/ui/alert'

const TURNSTILE_PROVIDER = 'turnstile'
const RECAPTCHA_PROVIDER = 'recaptcha'
const HCAPTCHA_PROVIDER = 'hcaptcha'
const GEETEST_V4_PROVIDER = 'geetest_v4'

const SCRIPT_BY_PROVIDER: Record<string, string> = {
  [TURNSTILE_PROVIDER]:
    'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit',
  [RECAPTCHA_PROVIDER]:
    'https://www.google.com/recaptcha/api.js?render=explicit',
  [HCAPTCHA_PROVIDER]: 'https://js.hcaptcha.com/1/api.js?render=explicit',
  [GEETEST_V4_PROVIDER]: 'https://static.geetest.com/v4/gt4.js',
}

type CaptchaWidgetId = string | number

type TurnstileApi = {
  render: (container: HTMLElement, options: Record<string, unknown>) => string
  remove?: (widgetId: string) => void
  reset?: (widgetId?: string) => void
}

type HCaptchaApi = {
  render: (container: HTMLElement, options: Record<string, unknown>) => string
  reset?: (widgetId?: string) => void
  remove?: (widgetId?: string) => void
}

type ReCaptchaApi = {
  render: (container: HTMLElement, options: Record<string, unknown>) => number
  reset?: (widgetId?: number) => void
}

type GeetestValidateResult = {
  captcha_output?: string
  gen_time?: string
  lot_number?: string
  pass_token?: string
}

type GeetestCaptcha = {
  appendTo: (container: HTMLElement | string) => void
  destroy?: () => void
  getValidate: () => GeetestValidateResult | undefined
  onError: (callback: () => void) => void
  onReady?: (callback: () => void) => void
  onSuccess: (callback: () => void) => void
  reset?: () => void
}

declare global {
  interface Window {
    turnstile?: TurnstileApi
    hcaptcha?: HCaptchaApi
    grecaptcha?: ReCaptchaApi
    initGeetest4?: (
      options: Record<string, unknown>,
      callback: (captcha: GeetestCaptcha) => void,
    ) => void
  }
}

const scriptPromises = new Map<string, Promise<void>>()

export function createEmptyCaptchaInput(): CaptchaInput {
  return {
    captcha_output: '',
    gen_time: '',
    lot_number: '',
    pass_token: '',
    provider: '',
    token: '',
  }
}

export function isCaptchaComplete(
  config?: CaptchaConfigResult,
  value?: CaptchaInput,
) {
  if (!config?.enabled) {
    return true
  }

  const provider = normalizeProvider(config.provider)
  if (provider === GEETEST_V4_PROVIDER) {
    if (!value) {
      return false
    }

    const lotNumber = value.lot_number ?? ''
    const captchaOutput = value.captcha_output ?? ''
    const passToken = value.pass_token ?? ''
    const genTime = value.gen_time ?? ''

    return !!(
      lotNumber.trim() &&
      captchaOutput.trim() &&
      passToken.trim() &&
      genTime.trim()
    )
  }

  const token = value?.token ?? ''
  return !!token.trim()
}

function normalizeProvider(provider?: string) {
  return provider?.trim().toLowerCase() || ''
}

function getProviderLabel(provider?: string) {
  switch (normalizeProvider(provider)) {
    case TURNSTILE_PROVIDER:
      return 'Cloudflare Turnstile'
    case RECAPTCHA_PROVIDER:
      return 'Google reCAPTCHA'
    case HCAPTCHA_PROVIDER:
      return 'hCaptcha'
    case GEETEST_V4_PROVIDER:
      return 'Geetest v4'
    default:
      return provider?.trim() || '验证码'
  }
}

function clearCaptchaInput(onChange: (next: CaptchaInput) => void) {
  onChange(createEmptyCaptchaInput())
}

function loadScript(src: string) {
  const existing = scriptPromises.get(src)
  if (existing) {
    return existing
  }

  const promise = new Promise<void>((resolve, reject) => {
    if (typeof document === 'undefined') {
      reject(new Error('当前环境不支持验证码脚本加载'))
      return
    }

    const current = document.querySelector<HTMLScriptElement>(
      `script[data-captcha-src="${src}"]`,
    )
    if (current) {
      if (current.dataset.loaded === 'true') {
        resolve()
        return
      }
      current.addEventListener('load', () => resolve(), { once: true })
      current.addEventListener(
        'error',
        () => reject(new Error('验证码脚本加载失败')),
        { once: true },
      )
      return
    }

    const script = document.createElement('script')
    script.async = true
    script.defer = true
    script.src = src
    script.dataset.captchaSrc = src
    script.onload = () => {
      script.dataset.loaded = 'true'
      resolve()
    }
    script.onerror = () => reject(new Error('验证码脚本加载失败'))
    document.head.appendChild(script)
  })

  scriptPromises.set(src, promise)
  return promise
}

async function waitFor<T>(
  reader: () => T | undefined,
  message: string,
  timeoutMs = 10000,
) {
  const startedAt = Date.now()

  while (Date.now()-startedAt < timeoutMs) {
    const value = reader()
    if (value !== undefined) {
      return value
    }
    await new Promise((resolve) => {
      window.setTimeout(resolve, 50)
    })
  }

  throw new Error(message)
}

function readProviderApi(provider: string) {
  switch (provider) {
    case TURNSTILE_PROVIDER:
      return window.turnstile
    case RECAPTCHA_PROVIDER:
      return window.grecaptcha
    case HCAPTCHA_PROVIDER:
      return window.hcaptcha
    case GEETEST_V4_PROVIDER:
      return window.initGeetest4
    default:
      return undefined
  }
}

export function CaptchaPanel(props: {
  config?: CaptchaConfigResult
  errorMessage?: string | null
  isLoading?: boolean
  onChange: (next: CaptchaInput) => void
  resetSignal?: number
  value: CaptchaInput
}) {
  const { config, errorMessage, isLoading, onChange, resetSignal, value } =
    props
  const provider = normalizeProvider(config?.provider)
  const containerId = useId()
  const containerRef = useRef<HTMLDivElement | null>(null)
  const widgetIdRef = useRef<CaptchaWidgetId | null>(null)
  const geetestCaptchaRef = useRef<GeetestCaptcha | null>(null)
  const [widgetError, setWidgetError] = useState<string | null>(null)
  const [widgetReady, setWidgetReady] = useState(false)

  useEffect(() => {
    clearCaptchaInput(onChange)
  }, [onChange, provider, config?.site_key])

  useEffect(() => {
    if (!config?.enabled) {
      setWidgetError(null)
      setWidgetReady(false)
      return
    }
    if (!config.site_key?.trim()) {
      setWidgetError('验证码已启用，但缺少 site_key，无法初始化前端验证控件。')
      setWidgetReady(false)
      return
    }
    if (!provider || !SCRIPT_BY_PROVIDER[provider]) {
      setWidgetError(`当前前端暂不支持 ${getProviderLabel(provider)} 验证控件。`)
      setWidgetReady(false)
      return
    }

    const abortController = new AbortController()
    const container = containerRef.current
    if (!container) {
      return
    }

    container.innerHTML = ''
    widgetIdRef.current = null
    geetestCaptchaRef.current = null
    setWidgetError(null)
    setWidgetReady(false)

    void (async () => {
      try {
        await loadScript(SCRIPT_BY_PROVIDER[provider])
        const api = await waitFor(
          () => readProviderApi(provider),
          '验证码脚本已加载，但前端 API 未就绪',
        )

        if (abortController.signal.aborted) {
          return
        }

        if (provider === TURNSTILE_PROVIDER) {
          widgetIdRef.current = (api as TurnstileApi).render(container, {
            sitekey: config.site_key,
            callback: (tokenValue: string) => {
              onChange({
                ...createEmptyCaptchaInput(),
                provider,
                token: tokenValue,
              })
            },
            'error-callback': () => {
              clearCaptchaInput(onChange)
              setWidgetError('验证码校验失败，请重试。')
            },
            'expired-callback': () => {
              clearCaptchaInput(onChange)
            },
          })
          setWidgetReady(true)
          return
        }

        if (provider === RECAPTCHA_PROVIDER) {
          widgetIdRef.current = (api as ReCaptchaApi).render(container, {
            sitekey: config.site_key,
            callback: (tokenValue: string) => {
              onChange({
                ...createEmptyCaptchaInput(),
                provider,
                token: tokenValue,
              })
            },
            'error-callback': () => {
              clearCaptchaInput(onChange)
              setWidgetError('验证码校验失败，请重试。')
            },
            'expired-callback': () => {
              clearCaptchaInput(onChange)
            },
          })
          setWidgetReady(true)
          return
        }

        if (provider === HCAPTCHA_PROVIDER) {
          widgetIdRef.current = (api as HCaptchaApi).render(container, {
            sitekey: config.site_key,
            callback: (tokenValue: string) => {
              onChange({
                ...createEmptyCaptchaInput(),
                provider,
                token: tokenValue,
              })
            },
            'error-callback': () => {
              clearCaptchaInput(onChange)
              setWidgetError('验证码校验失败，请重试。')
            },
            'expired-callback': () => {
              clearCaptchaInput(onChange)
            },
          })
          setWidgetReady(true)
          return
        }

        if (provider === GEETEST_V4_PROVIDER) {
          await new Promise<void>((resolve, reject) => {
            ;(api as Window['initGeetest4'])?.(
              {
                captchaId: config.site_key,
                product: 'float',
              },
              (captcha) => {
                geetestCaptchaRef.current = captcha
                captcha.onSuccess(() => {
                  const payload = captcha.getValidate()
                  onChange({
                    ...createEmptyCaptchaInput(),
                    provider,
                    lot_number: payload?.lot_number || '',
                    captcha_output: payload?.captcha_output || '',
                    pass_token: payload?.pass_token || '',
                    gen_time: payload?.gen_time || '',
                  })
                })
                captcha.onError(() => {
                  clearCaptchaInput(onChange)
                  setWidgetError('验证码校验失败，请重试。')
                })
                captcha.onReady?.(() => {
                  setWidgetReady(true)
                })
                captcha.appendTo(container)
                resolve()
              },
            )
            window.setTimeout(() => {
              reject(new Error('Geetest 初始化超时'))
            }, 10000)
          })
          return
        }
      } catch (error) {
        if (abortController.signal.aborted) {
          return
        }
        setWidgetError(
          error instanceof Error ? error.message : '验证码初始化失败',
        )
      }
    })()

    return () => {
      abortController.abort()
      clearCaptchaInput(onChange)

      if (provider === TURNSTILE_PROVIDER && widgetIdRef.current) {
        window.turnstile?.remove?.(String(widgetIdRef.current))
      }
      if (provider === HCAPTCHA_PROVIDER && widgetIdRef.current) {
        window.hcaptcha?.remove?.(String(widgetIdRef.current))
      }
      if (provider === RECAPTCHA_PROVIDER && widgetIdRef.current !== null) {
        window.grecaptcha?.reset?.(Number(widgetIdRef.current))
      }
      if (provider === GEETEST_V4_PROVIDER) {
        geetestCaptchaRef.current?.destroy?.()
      }
      widgetIdRef.current = null
      geetestCaptchaRef.current = null
    }
  }, [config?.enabled, config?.site_key, onChange, provider])

  useEffect(() => {
    if (!config?.enabled || resetSignal === undefined) {
      return
    }

    clearCaptchaInput(onChange)
    setWidgetError(null)

    if (provider === TURNSTILE_PROVIDER && widgetIdRef.current) {
      window.turnstile?.reset?.(String(widgetIdRef.current))
    }
    if (provider === RECAPTCHA_PROVIDER && widgetIdRef.current !== null) {
      window.grecaptcha?.reset?.(Number(widgetIdRef.current))
    }
    if (provider === HCAPTCHA_PROVIDER && widgetIdRef.current) {
      window.hcaptcha?.reset?.(String(widgetIdRef.current))
    }
    if (provider === GEETEST_V4_PROVIDER) {
      geetestCaptchaRef.current?.reset?.()
    }
  }, [config?.enabled, onChange, provider, resetSignal])

  if (isLoading) {
    return (
      <Alert className="mt-4">
        <AlertTitle>验证码</AlertTitle>
        <AlertDescription>正在读取验证码配置…</AlertDescription>
      </Alert>
    )
  }

  if (errorMessage) {
    return (
      <Alert className="mt-4" variant="destructive">
        <AlertTitle>验证码配置读取失败</AlertTitle>
        <AlertDescription>{errorMessage}</AlertDescription>
      </Alert>
    )
  }

  if (!config?.enabled) {
    return null
  }

  return (
    <div className="mt-4 space-y-3">
      <Alert>
        <AlertTitle>安全验证</AlertTitle>
        <AlertDescription>
          请完成 {getProviderLabel(provider)} 验证后再提交表单。
        </AlertDescription>
      </Alert>

      {widgetError ? (
        <Alert variant="destructive">
          <AlertTitle>验证码不可用</AlertTitle>
          <AlertDescription>{widgetError}</AlertDescription>
        </Alert>
      ) : null}

      <div
        className="min-h-20 rounded-xl border border-border/70 bg-muted/20 p-3"
        id={containerId}
        ref={containerRef}
      />

      {!widgetReady && !widgetError ? (
        <div className="text-sm text-muted-foreground">验证码控件加载中…</div>
      ) : null}

      {widgetReady && isCaptchaComplete(config, value) ? (
        <div className="text-sm text-emerald-600">验证已完成，可以提交。</div>
      ) : null}
    </div>
  )
}
