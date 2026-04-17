import { redirect, useNavigate } from '@tanstack/react-router'
import { isUnauthorizedApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'
import { safeRedirectPath } from '#/lib/format'
import { getQueryClient } from '#/lib/query-client'
import { currentUserQueryOptions, queryKeys } from '#/lib/query-options'
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

export function requireAuthenticatedRoute(): Promise<UserDTO> | UserDTO {
  const queryClient = getQueryClient()
  const cachedUser = queryClient.getQueryData<UserDTO>(queryKeys.currentUser())

  if (cachedUser) {
    return cachedUser
  }

  return queryClient
    .ensureQueryData({
      ...currentUserQueryOptions(),
      revalidateIfStale: true,
    })
    .catch((error) => {
      if (isUnauthorizedApiError(error)) {
        throw redirect({
          to: '/login',
          search: {
            redirect: resolveCurrentRedirectPath(),
          },
        })
      }

      throw error
    })
}

function assertAdmin(user: UserDTO): UserDTO {
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

export function requireAdminRoute(): Promise<UserDTO> | UserDTO {
  const userOrPromise = requireAuthenticatedRoute()

  if (userOrPromise instanceof Promise) {
    return userOrPromise.then(assertAdmin)
  }

  return assertAdmin(userOrPromise)
}

export function useUnauthorizedHandler(redirectPath: string) {
  const auth = useAuth()
  const navigate = useNavigate()

  function redirectToLogin() {
    void navigate({
      search: {
        redirect: safeRedirectPath(redirectPath),
      },
      to: '/login',
    })
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
