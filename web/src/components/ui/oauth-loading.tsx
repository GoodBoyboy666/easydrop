'use client'

import { motion, useReducedMotion } from 'motion/react'
import type { Transition } from 'motion/react'
import { ProviderIcon } from '#/components/site/provider-icons'

const PROVIDER_LABELS: Record<string, string> = {
  google: 'Google',
  github: 'GitHub',
  twitter: 'X',
  microsoft: 'Microsoft',
  apple: 'Apple',
}

interface OAuthLoadingProps {
  intent: 'login' | 'bind'
  provider: string
}

export function OAuthLoading({ intent, provider }: OAuthLoadingProps) {
  const prefersReducedMotion = useReducedMotion()

  const label = PROVIDER_LABELS[provider] ?? provider

  const entrance: Transition = prefersReducedMotion
    ? { duration: 0 }
    : { type: 'spring', duration: 0.7, ease: 'easeOut' }

  const dotTransition = (i: number): Transition =>
    prefersReducedMotion
      ? { duration: 0 }
      : { repeat: Infinity, duration: 0.9, delay: i * 0.15, ease: 'easeInOut' }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
      <div className="flex max-w-sm flex-col items-center gap-6 text-center">
        <motion.div
          animate={{ opacity: 1, scale: 1 }}
          className="flex size-16 items-center justify-center rounded-full border border-border/60 bg-card shadow-sm"
          initial={prefersReducedMotion ? false : { opacity: 0, scale: 0.6 }}
          transition={entrance}
        >
          <ProviderIcon className="size-8" provider={provider} />
        </motion.div>

        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="space-y-2"
          initial={prefersReducedMotion ? false : { opacity: 0, y: 8 }}
          transition={{ ...entrance, delay: 0.12 }}
        >
          <div className="font-heading text-lg font-semibold">
            {intent === 'bind' ? `正在绑定 ${label}` : `正在通过 ${label} 登录`}
          </div>
          <p className="text-sm leading-6 text-muted-foreground">
            {intent === 'bind'
              ? '请稍候，正在完成社交账号授权与绑定。'
              : '请稍候，正在完成社交账号授权与登录。'}
          </p>
        </motion.div>

        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center gap-1.5"
          initial={prefersReducedMotion ? false : { opacity: 0, y: 8 }}
          transition={{ ...entrance, delay: 0.24 }}
        >
          {[0, 1, 2].map((i) => (
            <motion.div
              key={i}
              animate={
                prefersReducedMotion
                  ? {}
                  : { y: [-3, 3, -3], opacity: [0.4, 1, 0.4] }
              }
              className="size-2 rounded-full bg-primary/60"
              transition={dotTransition(i)}
            />
          ))}
        </motion.div>
      </div>
    </div>
  )
}
