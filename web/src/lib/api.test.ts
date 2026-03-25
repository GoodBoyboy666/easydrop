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
