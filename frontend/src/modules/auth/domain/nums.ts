/**
 * Auth domain enums / const objects (thay enum để tương thích erasableSyntaxOnly)
 */

/** Provider đăng nhập/đăng ký qua mạng xã hội */
export const AuthProvider = {
  Facebook: 'facebook',
  Google: 'google',
  Twitter: 'twitter',
} as const

export type AuthProvider = (typeof AuthProvider)[keyof typeof AuthProvider]
