import { redirect } from '@tanstack/react-router'
import { api, isUnauthorizedApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { safeRedirectPath } from '#/lib/format'
import type { UserDTO } from '#/lib/types'

function resolveCurrentRedirectPath() {
  if (typeof window === 'undefined') {
    return '/'
  }

  return safeRedirectPath(
    `${window.location.pathname}${window.location.search}${window.location.hash}`,
  )
}

export function buildLoginRedirectHref(redirectPath: string) {
  const safePath = safeRedirectPath(redirectPath)
  return `/login?redirect=${encodeURIComponent(safePath)}`
}

export async function requireAuthenticatedRoute(): Promise<UserDTO> {
  try {
    return await api.getCurrentUser()
  } catch (error) {
    if (isUnauthorizedApiError(error)) {
      throw redirect({
        to: '/login',
        search: {
          redirect: resolveCurrentRedirectPath(),
        },
      })
    }

    throw error
  }
}

export async function requireAdminRoute(): Promise<UserDTO> {
  const user = await requireAuthenticatedRoute()

  if (!user.admin) {
    throw redirect({
      replace: true,
      search: {
        content: undefined,
      },
      to: '/',
    })
  }

  return user
}

export function useUnauthorizedHandler(redirectPath: string) {
  const auth = useAuth()

  function redirectToLogin() {
    window.location.assign(buildLoginRedirectHref(redirectPath))
  }

  function handleUnauthorized(error: unknown) {
    if (isUnauthorizedApiError(error)) {
      void auth.logout()
      redirectToLogin()
      return true
    }

    return false
  }

  return {
    auth,
    handleUnauthorized,
    redirectToLogin,
  }
}
