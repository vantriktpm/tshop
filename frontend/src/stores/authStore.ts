import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { User, LoginPayload, RegisterPayload } from '@/modules/auth/domain'
import { useLoading } from '@/core/loading'
import { userService } from '@/services'
import { env } from '@/core/config/env'

const USER_STORAGE_KEY = 'tshop_user'
const TOKEN_STORAGE_KEY = 'tshop_token'

function getStoredUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_STORAGE_KEY)
    if (!raw) return null
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const router = useRouter()
  const { loading, withLoading } = useLoading()

  const user = ref<User | null>(getStoredUser())
  const token = ref<string | null>(localStorage.getItem(TOKEN_STORAGE_KEY))

  const isAuthenticated = computed(() => !!user.value && !!token.value)

  function setSession(userData: User, accessToken: string) {
    user.value = userData
    token.value = accessToken
    try {
      localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(userData))
      localStorage.setItem(TOKEN_STORAGE_KEY, accessToken)
    } catch {
      /* ignore */
    }
  }

  function clearSession() {
    user.value = null
    token.value = null
    try {
      localStorage.removeItem(USER_STORAGE_KEY)
      localStorage.removeItem(TOKEN_STORAGE_KEY)
    } catch {
      /* ignore */
    }
  }

  async function login(payload: LoginPayload) {
    return withLoading(async () => {
      const res = await userService.login(payload)
      setSession(res.user, res.tokens.accessToken)
      await router.push({ name: 'Main' })
      return res.user
    })
  }

  async function register(payload: RegisterPayload) {
    return withLoading(async () => {
      const res = await userService.register(payload)
      setSession(res.user, res.tokens.accessToken)
      await router.push({ name: 'Main' })
      return res.user
    })
  }

  function logout() {
    clearSession()
    router.push({ name: 'Login' })
  }

  /** Redirect tới backend OAuth start theo từng MXH (Google, Facebook, X). */
  function loginWithGoogle() {
    const state = typeof window !== 'undefined' ? window.location.href : env.oauthBaseUrl
    const url = `${env.oauthBaseUrl}/api/auth/google/start?state=${encodeURIComponent(state)}`
    window.location.href = url
  }

  function loginWithFacebook() {
    const state = typeof window !== 'undefined' ? window.location.href : env.oauthBaseUrl
    const url = `${env.oauthBaseUrl}/api/auth/facebook/start?state=${encodeURIComponent(state)}`
    window.location.href = url
  }

  function loginWithX() {
    const state = typeof window !== 'undefined' ? window.location.href : env.oauthBaseUrl
    const url = `${env.oauthBaseUrl}/api/auth/x/start?state=${encodeURIComponent(state)}`
    window.location.href = url
  }

  /**
   * Xử lý khi user quay về từ OAuth (backend redirect với ?token=... hoặc ?error=...).
   * Trả về true nếu đã set session thành công, false nếu có lỗi.
   */
  function handleOAuthReturn(query: { token?: string; error?: string }): boolean {
    if (query.error) {
      clearSession()
      return false
    }
    if (!query.token) return false
    try {
      const payload = parseJwtPayload(query.token)
      const user: User = {
        id: payload.user_id ?? '',
        email: payload.email ?? '',
        fullName: payload.email ?? '',
      }
      setSession(user, query.token)
      return true
    } catch {
      clearSession()
      return false
    }
  }

  function parseJwtPayload(token: string): { user_id?: string; email?: string } {
    const parts = token.split('.')
    const payloadPart = parts[1]
    if (parts.length !== 3 || !payloadPart) return {}
    try {
      const base64 = payloadPart.replace(/-/g, '+').replace(/_/g, '/')
      return JSON.parse(atob(base64)) as { user_id?: string; email?: string }
    } catch {
      return {}
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    loading,
    login,
    register,
    logout,
    loginWithGoogle,
    loginWithFacebook,
    loginWithX,
    handleOAuthReturn,
    setSession,
    clearSession,
  }
})
