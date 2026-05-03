import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { OAuthLoading } from '#/components/ui/oauth-loading'

const OAUTH_INTENT_KEY = 'easydrop_oauth_intent'

export function setOAuthIntent(intent: 'login' | 'bind') {
  sessionStorage.setItem(OAUTH_INTENT_KEY, intent)
}

const PROVIDER_LABELS: Record<string, string> = {
  google: 'Google',
  github: 'GitHub',
  twitter: 'X',
  microsoft: 'Microsoft',
  apple: 'Apple',
}

let started = false

export const Route = createFileRoute('/oauth/$provider')({
  validateSearch: (search: Record<string, unknown>) => {
    return {
      code: typeof search.code === 'string' ? search.code : '',
      state: typeof search.state === 'string' ? search.state : '',
    }
  },
  component: OAuthCallbackPage,
})

function OAuthCallbackPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const { provider } = Route.useParams()
  const { code, state } = Route.useSearch()
  const [displayIntent, setDisplayIntent] = useState<'login' | 'bind'>('login')

  const loginMutation = useMutation({
    mutationFn: () => api.oauthCallback(provider, { code, state }),
    onSuccess: async () => {
      await auth.refreshUser()
      toast.success(`${PROVIDER_LABELS[provider] ?? provider} 登录成功`)
      void navigate({ to: '/' })
    },
    onError: (error) => {
      const msg = error instanceof Error ? error.message : '登录失败'
      if (msg.includes('手动绑定')) {
        toast.error('该邮箱已注册但未绑定此社交账号，请先使用密码登录后在设置中手动绑定', { duration: 8000 })
        void navigate({ to: '/login' })
        return
      }
      toast.error(msg)
      void navigate({ to: '/login' })
    },
  })

  const bindMutation = useMutation({
    mutationFn: () => api.bindOAuthManually(provider, { code, state }),
    onSuccess: () => {
      toast.success(`${PROVIDER_LABELS[provider] ?? provider} 已绑定`)
      void navigate({ to: '/me' })
    },
    onError: (error) => {
      const msg = error instanceof Error ? error.message : '绑定失败'
      toast.error(msg)
      void navigate({ to: '/me' })
    },
  })

  useEffect(() => {
    if (started) {
      return
    }
    started = true

    const intent = sessionStorage.getItem(OAUTH_INTENT_KEY)
    sessionStorage.removeItem(OAUTH_INTENT_KEY)
    setDisplayIntent(intent === 'bind' ? 'bind' : 'login')

    if (!code || !state) {
      toast.error('社交登录授权失败，缺少必要参数')
      void navigate({ to: '/login' })
      return
    }

    if (intent === 'bind') {
      bindMutation.mutate()
    } else {
      loginMutation.mutate()
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return <OAuthLoading intent={displayIntent} provider={provider} />
}
