import { ApiError } from '#/lib/api'
import { useAuth } from '#/lib/auth'

export function useAdminSession(redirectPath: string) {
  const auth = useAuth()

  function redirectToLogin() {
    window.location.assign(`/login?redirect=${encodeURIComponent(redirectPath)}`)
  }

  function handleUnauthorized(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      auth.logout()
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
