// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  requireAdminRoute,
  requireAuthenticatedRoute,
} from '#/lib/auth-guards'

const mockRedirect = vi.fn((input: unknown) => input)
const mockEnsureQueryData = vi.fn()
const mockGetQueryData = vi.fn()
const mockGetQueryClient = vi.fn(() => ({
  ensureQueryData: mockEnsureQueryData,
  getQueryData: mockGetQueryData,
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
  queryKeys: {
    currentUser: () => ['current-user'] as const,
  },
}))

describe('requireAuthenticatedRoute', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/')
    mockGetQueryData.mockReturnValue(undefined)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('returns cached user without network query', async () => {
    const user = {
      admin: false,
      id: 12,
      nickname: 'tester',
    }

    mockGetQueryData.mockReturnValue(user)

    await expect(Promise.resolve(requireAuthenticatedRoute())).resolves.toBe(user)
    expect(mockCurrentUserQueryOptions).not.toHaveBeenCalled()
    expect(mockEnsureQueryData).not.toHaveBeenCalled()
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  it('returns current user when cache miss query succeeds', async () => {
    const user = {
      admin: false,
      id: 13,
      nickname: 'cache-miss',
    }

    mockEnsureQueryData.mockResolvedValue(user)

    await expect(Promise.resolve(requireAuthenticatedRoute())).resolves.toBe(user)
    expect(mockCurrentUserQueryOptions).toHaveBeenCalledTimes(1)
    expect(mockEnsureQueryData).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ['current-user'],
        revalidateIfStale: true,
      }),
    )
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  it('redirects to login with redirect path on unauthorized error', async () => {
    const unauthorizedError = new Error('unauthorized')
    window.history.replaceState({}, '', '/me?tab=profile#security')
    mockEnsureQueryData.mockRejectedValue(unauthorizedError)
    mockIsUnauthorizedApiError.mockReturnValue(true)

    await expect(Promise.resolve(requireAuthenticatedRoute())).rejects.toEqual(
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
    mockEnsureQueryData.mockRejectedValue(networkError)
    mockIsUnauthorizedApiError.mockReturnValue(false)

    await expect(Promise.resolve(requireAuthenticatedRoute())).rejects.toBe(networkError)
    expect(mockRedirect).not.toHaveBeenCalled()
  })
})

describe('requireAdminRoute', () => {
  beforeEach(() => {
    mockGetQueryData.mockReturnValue(undefined)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('returns cached user when cached user is admin', async () => {
    const adminUser = {
      admin: true,
      id: 99,
    }

    mockGetQueryData.mockReturnValue(adminUser)

    await expect(Promise.resolve(requireAdminRoute())).resolves.toBe(adminUser)
    expect(mockEnsureQueryData).not.toHaveBeenCalled()
  })

  it('redirects to home when cached user is not admin', () => {
    mockGetQueryData.mockReturnValue({
      admin: false,
      id: 100,
    })

    try {
      requireAdminRoute()
      throw new Error('expected redirect to be thrown')
    } catch (error) {
      expect(error).toEqual(
        expect.objectContaining({
          replace: true,
          to: '/',
        }),
      )
    }
  })

  it('returns user when cache miss and query user is admin', async () => {
    const adminUser = {
      admin: true,
      id: 101,
    }
    mockEnsureQueryData.mockResolvedValue(adminUser)

    await expect(Promise.resolve(requireAdminRoute())).resolves.toBe(adminUser)
    expect(mockEnsureQueryData).toHaveBeenCalledTimes(1)
  })
})
