'use client'

import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError } from '#/lib/api'
import { FullScreenLoading } from '#/components/ui/full-screen-loading'
import { currentUserQueryOptions, queryKeys } from '#/lib/query-options'
import type { AuthState, UserDTO } from '#/lib/types'

const ACCESS_TOKEN_STORAGE_KEY = 'easydrop.access-token'

interface AuthContextValue extends AuthState {
  authenticateWithToken: (accessToken: string) => Promise<void>
  isAdmin: boolean
  logout: () => void
  refreshUser: () => Promise<UserDTO | null>
}

const AuthContext = createContext<AuthContextValue | null>(null)

function readStoredToken() {
  if (typeof window === 'undefined') {
    return null
  }

  return window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY)
}

function writeStoredToken(token: string | null) {
  if (typeof window === 'undefined') {
    return
  }

  if (token) {
    window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, token)
    return
  }

  window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY)
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const [token, setToken] = useState<string | null>(() => readStoredToken())
  const currentUserQuery = useQuery({
    ...currentUserQueryOptions(token ?? ''),
    enabled: !!token,
    retry: false,
  })

  useEffect(() => {
    if (!token || !currentUserQuery.error) {
      return
    }

    if (
      currentUserQuery.error instanceof ApiError &&
      currentUserQuery.error.status === 401
    ) {
      writeStoredToken(null)
      setToken(null)
    }
  }, [currentUserQuery.error, token])

  async function authenticateWithToken(accessToken: string) {
    writeStoredToken(accessToken)
    setToken(accessToken)

    try {
      await queryClient.fetchQuery(currentUserQueryOptions(accessToken))
    } catch (error) {
      writeStoredToken(null)
      setToken(null)
      throw error
    }
  }

  async function refreshUser() {
    if (!token) {
      return null
    }

    try {
      return await queryClient.fetchQuery(currentUserQueryOptions(token))
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout()
      }
      throw error
    }
  }

  function logout() {
    writeStoredToken(null)
    setToken(null)
    void queryClient.invalidateQueries({ queryKey: queryKeys.postsPrefix() })
  }

  const authState = useMemo<AuthState>(() => {
    if (!token) {
      return {
        status: 'anonymous',
        token: null,
        user: null,
      }
    }

    if (currentUserQuery.isPending) {
      return {
        status: 'loading',
        token,
        user: null,
      }
    }

    if (currentUserQuery.data) {
      return {
        status: 'authenticated',
        token,
        user: currentUserQuery.data,
      }
    }

    return {
      status: 'anonymous',
      token: null,
      user: null,
    }
  }, [currentUserQuery.data, currentUserQuery.isPending, token])

  const value = useMemo<AuthContextValue>(
    () => ({
      ...authState,
      authenticateWithToken,
      isAdmin: !!authState.user?.admin,
      logout,
      refreshUser,
    }),
    [authState, authenticateWithToken, logout, refreshUser],
  )

  return (
    <AuthContext.Provider value={value}>
      {authState.status === 'loading' ? <FullScreenLoading /> : children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }

  return context
}
