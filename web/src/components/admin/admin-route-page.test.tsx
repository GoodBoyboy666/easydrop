// @vitest-environment jsdom

import type { AnchorHTMLAttributes } from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AdminRoutePage } from '#/components/admin/admin-route-page'

const mockUseAuth = vi.fn()
const mockUseLocation = vi.fn()
const mockUseQuery = vi.fn()

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    ...props
  }: AnchorHTMLAttributes<HTMLAnchorElement>) => <a {...props}>{children}</a>,
  Outlet: () => <div data-testid="admin-outlet" />,
  useLocation: () => mockUseLocation(),
}))

vi.mock('@tanstack/react-query', async (importOriginal) => {
  const actual = await importOriginal()

  return Object.assign({}, actual, {
    useQuery: (...args: unknown[]) => mockUseQuery(...args),
  })
})

vi.mock('#/lib/auth', () => ({
  useAuth: () => mockUseAuth(),
}))

describe('AdminRoutePage', () => {
  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
  })

  beforeEach(() => {
    mockUseLocation.mockReturnValue({ pathname: '/admin' })
    mockUseQuery.mockReturnValue({ data: { total: 0 } })
  })

  it('shows a login prompt for anonymous users', () => {
    mockUseAuth.mockReturnValue({
      isAdmin: false,
      status: 'anonymous',
    })

    render(<AdminRoutePage />)

    expect(screen.getByText('需要管理员登录')).toBeTruthy()
    expect(screen.getByText('去登录')).toBeTruthy()
  })

  it('shows a forbidden notice for non-admin users', () => {
    mockUseAuth.mockReturnValue({
      isAdmin: false,
      status: 'authenticated',
      user: { admin: false },
    })

    render(<AdminRoutePage />)

    expect(screen.getByText('无权访问后台管理')).toBeTruthy()
  })

  it('renders the admin shell and overview for admins', () => {
    mockUseAuth.mockReturnValue({
      isAdmin: true,
      status: 'authenticated',
      user: {
        admin: true,
      },
    })

    render(<AdminRoutePage />)

    expect(screen.getByText('后台管理')).toBeTruthy()
    expect(screen.getAllByText('用户').length).toBeGreaterThan(0)
    expect(screen.getByText('数据图表')).toBeTruthy()
  })
})
