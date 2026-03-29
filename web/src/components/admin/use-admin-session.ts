import { useUnauthorizedHandler } from '#/lib/auth-guards'

export function useAdminSession(redirectPath: string) {
  return useUnauthorizedHandler(redirectPath)
}
