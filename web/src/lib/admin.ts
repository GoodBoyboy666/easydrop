import type { LucideIcon } from 'lucide-react'
import {
  FileTextIcon,
  ImageIcon,
  MessageSquareTextIcon,
  Settings2Icon,
  UsersIcon,
} from 'lucide-react'

export const ADMIN_PAGE_SIZE = 20

export const ADMIN_SELECT_CLASSNAME =
  'h-8 w-full rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30'

export interface AdminNavItem {
  icon: LucideIcon
  label: string
  to:
    | '/admin'
    | '/admin/attachments'
    | '/admin/comments'
    | '/admin/posts'
    | '/admin/settings'
    | '/admin/users'
}

export const adminNavItems: AdminNavItem[] = [
  {
    icon: Settings2Icon,
    label: '概览',
    to: '/admin',
  },
  {
    icon: UsersIcon,
    label: '用户',
    to: '/admin/users',
  },
  {
    icon: FileTextIcon,
    label: '日志',
    to: '/admin/posts',
  },
  {
    icon: MessageSquareTextIcon,
    label: '评论',
    to: '/admin/comments',
  },
  {
    icon: ImageIcon,
    label: '附件',
    to: '/admin/attachments',
  },
  {
    icon: Settings2Icon,
    label: '设置',
    to: '/admin/settings',
  },
]

export function formatBytes(value?: number | null) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '--'
  }

  if (value < 1024) {
    return `${value} B`
  }

  const units = ['KB', 'MB', 'GB', 'TB']
  let current = value / 1024
  let unitIndex = 0

  while (current >= 1024 && unitIndex < units.length - 1) {
    current /= 1024
    unitIndex += 1
  }

  return `${current.toFixed(current >= 10 ? 0 : 1)} ${units[unitIndex]}`
}

export function formatUserStatus(status?: number) {
  switch (status) {
    case 1:
      return '正常'
    case 2:
      return '封禁'
    case 3:
      return '停用'
    default:
      return '未知'
  }
}

export function formatAttachmentBizType(bizType?: number) {
  switch (bizType) {
    case 1:
      return '图片'
    case 2:
      return '视频'
    case 3:
      return '音频'
    case 4:
      return '文件'
    default:
      return '未知'
  }
}

export function parseOptionalInteger(value: string) {
  const trimmed = value.trim()
  if (trimmed === '') {
    return undefined
  }

  const parsed = Number.parseInt(trimmed, 10)
  return Number.isFinite(parsed) ? parsed : undefined
}

export function parseOptionalBoolean(value: string) {
  switch (value) {
    case 'true':
      return true
    case 'false':
      return false
    default:
      return undefined
  }
}
