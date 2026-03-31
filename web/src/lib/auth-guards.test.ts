// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  requireAdminRoute,
  requireAuthenticatedRoute,
} from '#/lib/auth-guards'

const mockRedirect = vi.fn((input: unknown) => input)
const mockFetchQuery = vi.fn()
const mockGetQueryClient = vi.fn(() => ({
  fetchQuery: mockFetchQuery,
}))
const mockCurrentUserQueryOptions = vi.fn(() => ({
  queryKey: ['current-user'] as const,
}))
const mockIsUnauthorizedApiError = vi.fn((_error: unknown) => false)

vi.mock('@tanstack/react-router', () => ({
  redirect: (input: unknown) => mockRedirect(input),
  useNavigate: () => vi.fn(),
}))

vi.mock('#/lib/api', () => ({
  isUnauthorizedApiError: (error: unknown) => mockIsUnauthorizedApiError(error),
}))

vi.mock('#/lib/auth', () => ({
  useAuth: () => ({
    logout: vi.fn(),
  }),
}))

vi.mock('#/lib/query-client', () => ({
  getQueryClient: () => mockGetQueryClient(),
}))

vi.mock('#/lib/query-options', () => ({
  currentUserQueryOptions: () => mockCurrentUserQueryOptions(),
}))

describe('requireAuthenticatedRoute', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/')
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('returns current user when query succeeds', async () => {
    const user = {
      admin: false,
      id: 12,
      nickname: 'tester',
    }

    mockFetchQuery.mockResolvedValue(user)

    await expect(requireAuthenticatedRoute()).resolves.toBe(user)
    expect(mockCurrentUserQueryOptions).toHaveBeenCalledTimes(1)
    expect(mockFetchQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ['current-user'],
      }),
    )
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  it('redirects to login with redirect path on unauthorized error', async () => {
    const unauthorizedError = new Error('unauthorized')
    window.history.replaceState({}, '', '/me?tab=profile#security')
    mockFetchQuery.mockRejectedValue(unauthorizedError)
    mockIsUnauthorizedApiError.mockReturnValue(true)

    await expect(requireAuthenticatedRoute()).rejects.toEqual(
      expect.objectContaining({
        search: {
          redirect: '/me?tab=profile#security',
        },
        to: '/login',
      }),
    )

    expect(mockRedirect).toHaveBeenCalledTimes(1)
  })

  it('rethrows non-unauthorized errors', async () => {
    const networkError = new Error('network failure')
    mockFetchQuery.mockRejectedValue(networkError)
    mockIsUnauthorizedApiError.mockReturnValue(false)

    await expect(requireAuthenticatedRoute()).rejects.toBe(networkError)
    expect(mockRedirect).not.toHaveBeenCalled()
  })
})

describe('requireAdminRoute', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('returns user when user is admin', async () => {
    const adminUser = {
      admin: true,
      id: 99,
    }

    mockFetchQuery.mockResolvedValue(adminUser)

    await expect(requireAdminRoute()).resolves.toBe(adminUser)
  })

  it('redirects to home when user is not admin', async () => {
    mockFetchQuery.mockResolvedValue({
      admin: false,
      id: 100,
    })

    await expect(requireAdminRoute()).rejects.toEqual(
      expect.objectContaining({
        replace: true,
        to: '/',
      }),
    )
  })
})
