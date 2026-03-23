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
  storage_quota?: number
  storage_used?: number
  updated_at?: string
  username: string
}

export interface PostAuthorDTO {
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
  hide?: boolean
  id: number
  tags?: TagDTO[]
  updated_at?: string
}

export interface CommentAuthorDTO {
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

export interface SettingPublicItem {
  key: string
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

export interface LoginInput {
  account: string
  password: string
  captcha?: Record<string, string>
}

export interface RegisterInput {
  email: string
  nickname: string
  password: string
  username: string
  captcha?: Record<string, string>
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
  hide?: boolean
}

export interface CreateCommentInput {
  content: string
  parent_id?: number
}

export interface AuthState {
  status: 'anonymous' | 'loading' | 'authenticated'
  token: string | null
  user: UserDTO | null
}

export type PublicSettingsMap = Record<string, string>
