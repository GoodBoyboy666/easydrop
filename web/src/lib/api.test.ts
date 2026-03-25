import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from '#/lib/api'

describe('api.updateAdminComment', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('sends a patch request with token and content payload', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          author: {
            id: 3,
            nickname: 'Tester',
          },
          content: 'updated comment',
          created_at: '2026-03-25T00:00:00Z',
          id: 12,
          post_id: 7,
        }),
        {
          headers: {
            'content-type': 'application/json',
          },
          status: 200,
        },
      ),
    )

    vi.stubGlobal('fetch', fetchMock)

    await api.updateAdminComment(
      12,
      {
        content: 'updated comment',
      },
      'token-123',
    )

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/admin/comments/12',
      expect.objectContaining({
        body: JSON.stringify({ content: 'updated comment' }),
        headers: expect.objectContaining({
          Authorization: 'Bearer token-123',
          'Content-Type': 'application/json',
        }),
        method: 'PATCH',
      }),
    )
  })
})

describe('api.updateMyComment', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('sends a patch request to the current-user comment endpoint', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          author: {
            id: 8,
            nickname: 'Owner',
          },
          content: 'my updated comment',
          created_at: '2026-03-25T00:00:00Z',
          id: 21,
          post_id: 9,
        }),
        {
          headers: {
            'content-type': 'application/json',
          },
          status: 200,
        },
      ),
    )

    vi.stubGlobal('fetch', fetchMock)

    await api.updateMyComment(
      21,
      {
        content: 'my updated comment',
      },
      'token-own',
    )

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/users/me/comments/21',
      expect.objectContaining({
        body: JSON.stringify({ content: 'my updated comment' }),
        headers: expect.objectContaining({
          Authorization: 'Bearer token-own',
          'Content-Type': 'application/json',
        }),
        method: 'PATCH',
      }),
    )
  })
})

describe('api.getAdminUsers', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('sends query params and bearer token to the admin users endpoint', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          items: [],
          total: 0,
        }),
        {
          headers: {
            'content-type': 'application/json',
          },
          status: 200,
        },
      ),
    )

    vi.stubGlobal('fetch', fetchMock)

    await api.getAdminUsers(
      {
        limit: 20,
        offset: 40,
        order: 'created_at_desc',
        status: 1,
        username: 'neo',
      },
      'token-admin',
    )

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/admin/users?limit=20&offset=40&order=created_at_desc&status=1&username=neo',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer token-admin',
          'Content-Type': 'application/json',
        }),
      }),
    )
  })
})

describe('api.uploadAdminUserAvatar', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('uploads avatar with form data and without forcing a json content-type', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          admin: true,
          email: 'neo@example.com',
          id: 7,
          nickname: 'Neo',
          username: 'neo',
        }),
        {
          headers: {
            'content-type': 'application/json',
          },
          status: 200,
        },
      ),
    )

    vi.stubGlobal('fetch', fetchMock)

    await api.uploadAdminUserAvatar(
      7,
      new File(['avatar'], 'avatar.png', { type: 'image/png' }),
      'token-upload',
    )

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/admin/users/7/avatar',
      expect.objectContaining({
        body: expect.any(FormData),
        headers: expect.not.objectContaining({
          'Content-Type': 'application/json',
        }),
        method: 'POST',
      }),
    )
  })
})

describe('api.updateAdminSetting', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('encodes the key in the path and sends a patch request', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          message: 'ok',
        }),
        {
          headers: {
            'content-type': 'application/json',
          },
          status: 200,
        },
      ),
    )

    vi.stubGlobal('fetch', fetchMock)

    await api.updateAdminSetting(
      'site.url',
      {
        value: 'https://example.com',
      },
      'token-setting',
    )

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/admin/settings/site.url',
      expect.objectContaining({
        body: JSON.stringify({ value: 'https://example.com' }),
        headers: expect.objectContaining({
          Authorization: 'Bearer token-setting',
          'Content-Type': 'application/json',
        }),
        method: 'PATCH',
      }),
    )
  })
})
