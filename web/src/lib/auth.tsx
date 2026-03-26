'use client'

import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiError } from '#/lib/api'
import { FullScreenLoading } from '#/components/ui/full-screen-loading'
import { currentUserQueryOptions, queryKeys } from '#/lib/query-options'
import type { AuthState, UserDTO } from '#/lib/types'

interface AuthContextValue extends AuthState {
  isAdmin: boolean
  logout: () => Promise<void>
  refreshUser: () => Promise<UserDTO | null>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const [sessionCheckEnabled, setSessionCheckEnabled] = useState(true)
  const currentUserQuery = useQuery({
    ...currentUserQueryOptions(),
    enabled: sessionCheckEnabled,
    retry: false,
  })

  useEffect(() => {
    if (!sessionCheckEnabled || !currentUserQuery.error) {
      return
    }

    if (
      currentUserQuery.error instanceof ApiError &&
      currentUserQuery.error.status === 401
    ) {
      setSessionCheckEnabled(false)
      queryClient.removeQueries({ queryKey: queryKeys.currentUser() })
    }
  }, [currentUserQuery.error, queryClient, sessionCheckEnabled])

  async function refreshUser() {
    setSessionCheckEnabled(true)

    try {
      return await queryClient.fetchQuery(currentUserQueryOptions())
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setSessionCheckEnabled(false)
        queryClient.removeQueries({ queryKey: queryKeys.currentUser() })
      }
      throw error
    }
  }

  async function logout() {
    try {
      await api.logout()
    } catch {
      // 即使服务端登出失败，也要先清空本地认证状态。
    }

    setSessionCheckEnabled(false)
    queryClient.clear()
  }

  const authState = useMemo<AuthState>(() => {
    if (sessionCheckEnabled && currentUserQuery.isPending) {
      return {
        status: 'loading',
        user: null,
      }
    }

    if (currentUserQuery.data) {
      return {
        status: 'authenticated',
        user: currentUserQuery.data,
      }
    }

    return {
      status: 'anonymous',
      user: null,
    }
  }, [currentUserQuery.data, currentUserQuery.isPending, sessionCheckEnabled])

  const value = useMemo<AuthContextValue>(
    () => ({
      ...authState,
      isAdmin: !!authState.user?.admin,
      logout,
      refreshUser,
    }),
    [authState],
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
