import { createFileRoute } from '@tanstack/react-router'
import { AdminAttachmentsPage } from '#/components/admin/admin-attachments-page'

export const Route = createFileRoute('/admin/attachments')({
  component: AdminAttachmentsPage,
})
