import { createFileRoute } from '@tanstack/react-router'
import { AdminCommentsPage } from '#/components/admin/admin-comments-page'

export const Route = createFileRoute('/admin/comments')({
  component: AdminCommentsPage,
})
