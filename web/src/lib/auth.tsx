"use client"

import {
  createContext,
  startTransition,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { api, ApiError } from '#/lib/api'
import type {
  AuthState,
  LoginInput,
  RegisterInput,
  UserDTO,
} from '#/lib/types'

const ACCESS_TOKEN_STORAGE_KEY = 'easydrop.access-token'

interface AuthContextValue extends AuthState {
  isAdmin: boolean
  login: (input: LoginInput) => Promise<void>
  logout: () => void
  refreshUser: () => Promise<UserDTO | null>
  register: (input: RegisterInput) => Promise<void>
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
  const [authState, setAuthState] = useState<AuthState>({
    status: 'loading',
    token: null,
    user: null,
  })

  useEffect(() => {
    let cancelled = false

    async function bootstrap() {
      const storedToken = readStoredToken()

      if (!storedToken) {
        startTransition(() => {
          setAuthState({
            status: 'anonymous',
            token: null,
            user: null,
          })
        })
        return
      }

      startTransition(() => {
        setAuthState({
          status: 'loading',
          token: storedToken,
          user: null,
        })
      })

      try {
        const user = await api.getCurrentUser(storedToken)

        if (cancelled) {
          return
        }

        startTransition(() => {
          setAuthState({
            status: 'authenticated',
            token: storedToken,
            user,
          })
        })
      } catch (error) {
        if (cancelled) {
          return
        }

        if (error instanceof ApiError && error.status === 401) {
          writeStoredToken(null)
        }

        startTransition(() => {
          setAuthState({
            status: 'anonymous',
            token: null,
            user: null,
          })
        })
      }
    }

    void bootstrap()

    return () => {
      cancelled = true
    }
  }, [])

  async function applyAccessToken(accessToken: string) {
    writeStoredToken(accessToken)

    startTransition(() => {
      setAuthState({
        status: 'loading',
        token: accessToken,
        user: null,
      })
    })

    try {
      const user = await api.getCurrentUser(accessToken)
      startTransition(() => {
        setAuthState({
          status: 'authenticated',
          token: accessToken,
          user,
        })
      })
    } catch (error) {
      writeStoredToken(null)
      startTransition(() => {
        setAuthState({
          status: 'anonymous',
          token: null,
          user: null,
        })
      })
      throw error
    }
  }

  async function login(input: LoginInput) {
    const result = await api.login(input)
    await applyAccessToken(result.access_token)
  }

  async function register(input: RegisterInput) {
    const result = await api.register(input)
    await applyAccessToken(result.access_token)
  }

  async function refreshUser() {
    if (!authState.token) {
      return null
    }

    try {
      const user = await api.getCurrentUser(authState.token)
      startTransition(() => {
        setAuthState((current) => ({
          ...current,
          status: 'authenticated',
          user,
        }))
      })
      return user
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout()
      }
      throw error
    }
  }

  function logout() {
    writeStoredToken(null)
    startTransition(() => {
      setAuthState({
        status: 'anonymous',
        token: null,
        user: null,
      })
    })
  }

  const value = useMemo<AuthContextValue>(
    () => ({
      ...authState,
      isAdmin: !!authState.user?.admin,
      login,
      logout,
      refreshUser,
      register,
    }),
    [authState]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }

  return context
}
