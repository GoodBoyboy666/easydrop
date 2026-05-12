import { afterEach, describe, expect, it, vi } from 'vitest'
import MockAdapter from 'axios-mock-adapter'
import { api, axiosInstance } from '#/lib/api'

describe('api.updateAdminComment', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends a patch request with token and content payload', async () => {
    mock = new MockAdapter(axiosInstance)
    mock
      .onPatch('/admin/comments/12')
      .reply(200, {
        author: { id: 3, nickname: 'Tester' },
        content: 'updated comment',
        created_at: '2026-03-25T00:00:00Z',
        id: 12,
        post_id: 7,
      })

    const result = await api.updateAdminComment(
      12,
      { content: 'updated comment' },
      'token-123',
    )

    expect(result).toMatchObject({
      author: { id: 3, nickname: 'Tester' },
      content: 'updated comment',
    })
    expect(mock.history.patch.length).toBe(1)
    expect(mock.history.patch[0].url).toBe('/admin/comments/12')
    expect(mock.history.patch[0].headers?.Authorization).toBe(
      'Bearer token-123',
    )
    expect(JSON.parse(mock.history.patch[0].data)).toEqual({
      content: 'updated comment',
    })
  })
})

describe('api.updateMyComment', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends a patch request to the current-user comment endpoint', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPatch('/users/me/comments/21').reply(200, {
      author: { id: 8, nickname: 'Owner' },
      content: 'my updated comment',
      created_at: '2026-03-25T00:00:00Z',
      id: 21,
      post_id: 9,
    })

    const result = await api.updateMyComment(
      21,
      { content: 'my updated comment' },
      'token-own',
    )

    expect(result).toMatchObject({
      author: { id: 8, nickname: 'Owner' },
      content: 'my updated comment',
    })
    expect(mock.history.patch.length).toBe(1)
    expect(mock.history.patch[0].url).toBe('/users/me/comments/21')
    expect(mock.history.patch[0].headers?.Authorization).toBe(
      'Bearer token-own',
    )
    expect(JSON.parse(mock.history.patch[0].data)).toEqual({
      content: 'my updated comment',
    })
  })
})

describe('api.getAdminUsers', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends query params and bearer token to the admin users endpoint', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onGet('/admin/users').reply(200, {
      items: [],
      total: 0,
    })

    await api.getAdminUsers(
      {
        page: 3,
        order: 'created_at_desc',
        size: 20,
        status: 1,
        user_id: 42,
        username: 'neo',
      },
      'token-admin',
    )

    expect(mock.history.get.length).toBe(1)
    expect(mock.history.get[0].url).toBe('/admin/users')
    expect(mock.history.get[0].headers?.Authorization).toBe(
      'Bearer token-admin',
    )
    expect(mock.history.get[0].params).toEqual({
      page: 3,
      order: 'created_at_desc',
      size: 20,
      status: 1,
      user_id: 42,
      username: 'neo',
    })
  })
})

describe('api.getAdminOverview', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('requests the admin overview endpoint with bearer token', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onGet('/admin/overview').reply(200, {
      recent_activity: [{ comments: 4, date: '2026-03-27', posts: 2 }],
      totals: { attachments: 9, comments: 8, posts: 7, users: 6 },
    })

    await api.getAdminOverview('token-overview')

    expect(mock.history.get.length).toBe(1)
    expect(mock.history.get[0].url).toBe('/admin/overview')
    expect(mock.history.get[0].headers?.Authorization).toBe(
      'Bearer token-overview',
    )
  })
})

describe('api.uploadAdminUserAvatar', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('uploads avatar with form data and without forcing a json content-type', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPost('/admin/users/7/avatar').reply(200, {
      admin: true,
      email: 'neo@example.com',
      id: 7,
      nickname: 'Neo',
      username: 'neo',
    })

    const result = await api.uploadAdminUserAvatar(
      7,
      new File(['avatar'], 'avatar.png', { type: 'image/png' }),
      'token-upload',
    )

    expect(result).toMatchObject({ id: 7, nickname: 'Neo' })
    expect(mock.history.post.length).toBe(1)
    expect(mock.history.post[0].url).toBe('/admin/users/7/avatar')
    expect(mock.history.post[0].data).toBeInstanceOf(FormData)
    expect(mock.history.post[0].headers?.['Content-Type']).not.toBe(
      'application/json',
    )
  })
})

describe('api.updateAdminSetting', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('encodes the key in the path and sends a patch request', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPatch('/admin/settings/site.url').reply(200, { message: 'ok' })

    await api.updateAdminSetting(
      'site.url',
      { value: 'https://example.com' },
      'token-setting',
    )

    expect(mock.history.patch.length).toBe(1)
    expect(mock.history.patch[0].url).toBe('/admin/settings/site.url')
    expect(mock.history.patch[0].headers?.Authorization).toBe(
      'Bearer token-setting',
    )
    expect(JSON.parse(mock.history.patch[0].data)).toEqual({
      value: 'https://example.com',
    })
  })
})

describe('api.createPostComment', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends captcha payload when provided', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPost('/posts/9/comments').reply(201, {
      author: { id: 8, nickname: 'Owner' },
      content: 'my comment',
      created_at: '2026-03-25T00:00:00Z',
      id: 21,
      post_id: 9,
    })

    await api.createPostComment(
      9,
      {
        captcha: { provider: 'turnstile', token: 'captcha-token' },
        content: 'my comment',
      },
      'token-own',
    )

    expect(mock.history.post.length).toBe(1)
    expect(mock.history.post[0].url).toBe('/posts/9/comments')
    expect(mock.history.post[0].headers?.Authorization).toBe(
      'Bearer token-own',
    )
    expect(JSON.parse(mock.history.post[0].data)).toEqual({
      captcha: { provider: 'turnstile', token: 'captcha-token' },
      content: 'my comment',
    })
  })
})

describe('api.register', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends register payload and returns message response', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPost('/auth/register').reply(201, {
      message: '注册成功，请先完成邮箱验证后登录',
    })

    const result = await api.register({
      email: 'alice@example.com',
      nickname: 'Alice',
      password: 'Pass1234',
      username: 'alice',
    })

    expect(result).toEqual({
      message: '注册成功，请先完成邮箱验证后登录',
    })
    expect(mock.history.post.length).toBe(1)
    expect(mock.history.post[0].url).toBe('/auth/register')
    expect(JSON.parse(mock.history.post[0].data)).toEqual({
      email: 'alice@example.com',
      nickname: 'Alice',
      password: 'Pass1234',
      username: 'alice',
    })
  })
})

describe('api.initializeSystem', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.restoreAllMocks()
  })

  it('sends init secret together with the initialization payload', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPost('/init').reply(201, { message: 'ok' })

    await api.initializeSystem({
      allow_register: true,
      email: 'admin@example.com',
      nickname: 'Admin',
      password: 'Pass1234',
      secret: 'secret-123',
      site_announcement: 'hello',
      site_name: 'EasyDrop',
      site_url: 'http://localhost:3000',
      username: 'admin',
    })

    expect(mock.history.post.length).toBe(1)
    expect(mock.history.post[0].url).toBe('/init')
    expect(JSON.parse(mock.history.post[0].data)).toEqual({
      allow_register: true,
      email: 'admin@example.com',
      nickname: 'Admin',
      password: 'Pass1234',
      secret: 'secret-123',
      site_announcement: 'hello',
      site_name: 'EasyDrop',
      site_url: 'http://localhost:3000',
      username: 'admin',
    })
  })
})

describe('api CSRF header', () => {
  let mock: MockAdapter

  afterEach(() => {
    mock.restore()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('attaches csrf header for unsafe requests when csrf cookie is present', async () => {
    mock = new MockAdapter(axiosInstance)
    mock.onPost('/auth/logout').reply(200, { message: 'ok' })

    vi.stubGlobal('document', {
      cookie: 'easydrop_csrf_token=csrf-token-123',
    } as Document)

    await api.logout()

    expect(mock.history.post.length).toBe(1)
    expect(mock.history.post[0].url).toBe('/auth/logout')
    expect(mock.history.post[0].headers?.['X-CSRF-Token']).toBe(
      'csrf-token-123',
    )
  })
})
