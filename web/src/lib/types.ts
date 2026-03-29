export interface AuthResult {
  access_token: string
}

export interface UserDTO {
  admin: boolean
  avatar?: string
  created_at?: string
  email?: string
  email_verified?: boolean
  id: number
  nickname: string
  status?: number
  storage_quota?: number | null
  storage_used?: number
  updated_at?: string
  username: string
}

export interface PostAuthorDTO {
  admin?: boolean
  avatar?: string
  id: number
  nickname: string
}

export interface TagDTO {
  created_at?: string
  id: number
  name: string
}

export interface PostDTO {
  author: PostAuthorDTO
  content: string
  created_at: string
  disable_comment?: boolean
  hide?: boolean
  id: number
  pin?: number
  tags?: TagDTO[]
  updated_at?: string
}

export interface PublicPostListResult {
  items: PostDTO[]
  pinnedItems: PostDTO[]
  total: number
}

export interface CommentAuthorDTO {
  admin?: boolean
  avatar?: string
  id: number
  nickname: string
}

export interface CommentDTO {
  author: CommentAuthorDTO
  content: string
  created_at: string
  id: number
  parent_id?: number
  post_id: number
  reply_to_user?: CommentAuthorDTO
  root_id?: number
  updated_at?: string
}

export interface PagedResult<T> {
  items: T[]
  total: number
}

export interface AttachmentDTO {
  biz_type: number
  created_at: string
  file_key: string
  file_size: number
  id: number
  storage_type: string
  url: string
  user_id: number
}

export interface AttachmentBatchDeleteFailedItem {
  id: number
  message: string
}

export interface AttachmentBatchDeleteResult {
  failed: AttachmentBatchDeleteFailedItem[]
  success_ids: number[]
}

export interface SettingPublicItem {
  key: string
  value: string
}

export interface SettingItem {
  category: string
  desc: string
  key: string
  public: boolean
  sensitive: boolean
  value: string
}

export interface SettingPublicResult {
  items: SettingPublicItem[]
}

export interface CaptchaConfigResult {
  enabled: boolean
  provider?: string
  site_key?: string
}

export interface CaptchaInput {
  provider?: string
  token?: string
  lot_number?: string
  captcha_output?: string
  pass_token?: string
  gen_time?: string
}

export interface LoginInput {
  account: string
  password: string
  captcha?: CaptchaInput
}

export interface RegisterInput {
  email: string
  nickname: string
  password: string
  username: string
  captcha?: CaptchaInput
}

export interface PasswordResetRequestInput {
  email: string
  captcha?: CaptchaInput
}

export interface PasswordResetConfirmInput {
  token: string
  new_password: string
}

export interface EmailVerifyConfirmInput {
  token: string
}

export interface EmailChangeConfirmInput {
  token: string
}

export interface UpdateMyProfileInput {
  nickname?: string
}

export interface ChangeMyPasswordInput {
  old_password: string
  new_password: string
}

export interface ChangeMyEmailInput {
  current_password: string
  new_email: string
}

export interface InitInput {
  allow_register: boolean
  email: string
  nickname: string
  password: string
  site_announcement: string
  site_name: string
  site_url: string
  username: string
}

export interface InitStatusResult {
  initialized: boolean
}

export interface CreatePostInput {
  content: string
  disable_comment?: boolean
  hide?: boolean
  pin?: number
}

export interface UpdatePostInput {
  clear_pin?: boolean
  content?: string
  disable_comment?: boolean
  hide?: boolean
  pin?: number
}

export interface CreateCommentInput {
  content: string
  parent_id?: number
}

export interface UpdateCommentInput {
  content?: string
}

export interface CreateUserInput {
  admin?: boolean
  email: string
  email_verified?: boolean
  nickname?: string
  password: string
  status?: number
  storage_quota?: number | null
  username: string
}

export interface UpdateUserInput {
  admin?: boolean
  email?: string
  email_verified?: boolean
  nickname?: string
  password?: string
  status?: number
  storage_quota?: number | null
  use_default_storage_quota?: boolean
  username?: string
}

export interface UpdateSettingInput {
  value?: string
}

export interface AdminUserListQuery {
  email?: string
  page?: number
  size?: number
  order?: string
  status?: number
  username?: string
}

export interface AdminPostListQuery {
  content?: string
  hide?: boolean
  page?: number
  size?: number
  order?: string
  tag_id?: number
  user_id?: number
}

export interface AdminCommentListQuery {
  page?: number
  size?: number
  order?: string
  post_id?: number
  user_id?: number
}

export interface AdminAttachmentListQuery {
  biz_type?: number
  created_from?: number
  created_to?: number
  id?: number
  page?: number
  size?: number
  order?: string
  user_id?: number
}

export interface AdminSettingListQuery {
  category?: string
  key?: string
  page?: number
  size?: number
  order?: string
}

export interface AdminOverviewTotals {
  attachments: number
  comments: number
  posts: number
  users: number
}

export interface AdminOverviewTrendItem {
  comments: number
  date: string
  posts: number
}

export interface AdminOverviewResult {
  recent_activity: AdminOverviewTrendItem[]
  totals: AdminOverviewTotals
}

export interface AuthState {
  status: 'anonymous' | 'loading' | 'authenticated'
  user: UserDTO | null
}

export type PublicSettingsMap = Record<string, string>
