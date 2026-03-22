import { describe, expect, it } from 'vitest'

import { isRootComment } from './easydrop-api'

describe('isRootComment', () => {
  it('treats nullish and zero parent/root ids as root comments', () => {
    expect(
      isRootComment({
        content: 'root',
        created_at: '',
        id: 1,
        parent_id: 0,
        post_id: 1,
        root_id: null,
        updated_at: '',
        user_id: 1,
      }),
    ).toBe(true)
  })

  it('rejects comments with parent or root ids', () => {
    expect(
      isRootComment({
        content: 'reply',
        created_at: '',
        id: 2,
        parent_id: 3,
        post_id: 1,
        root_id: 3,
        updated_at: '',
        user_id: 2,
      }),
    ).toBe(false)
  })
})
