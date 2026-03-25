import { createFileRoute } from '@tanstack/react-router'
import { AdminRoutePage } from '#/components/admin/admin-route-page'

export const Route = createFileRoute('/admin')({
  component: AdminRoutePage,
})
