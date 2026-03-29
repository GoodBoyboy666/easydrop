import { createFileRoute } from '@tanstack/react-router'
import { AdminRoutePage } from '#/components/admin/admin-route-page'
import { requireAdminRoute } from '#/lib/auth-guards'

export const Route = createFileRoute('/admin')({
  beforeLoad: async () => {
    await requireAdminRoute()
  },
  component: AdminRoutePage,
})
