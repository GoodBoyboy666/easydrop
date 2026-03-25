import { createFileRoute } from '@tanstack/react-router'
import { AdminPostsPage } from '#/components/admin/admin-posts-page'

export const Route = createFileRoute('/admin/posts')({
  component: AdminPostsPage,
})
