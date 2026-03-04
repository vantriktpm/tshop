<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores'
import { env } from '@/core/config/env'
import { httpClient } from '@/core/http'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  email: '',
  password: '',
})
const error = ref('')
const googleLoading = ref(false)
const googleError = ref('')

interface GoogleVerifyResponse {
  access_token: string
  user: { id: string; email: string; full_name: string }
}

function loadGoogleScript(): Promise<void> {
  return new Promise((resolve) => {
    if (typeof window === 'undefined') {
      resolve()
      return
    }
    if ((window as unknown as { __gsiLoaded?: boolean }).__gsiLoaded) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.onload = () => {
      (window as unknown as { __gsiLoaded?: boolean }).__gsiLoaded = true
      resolve()
    }
    document.head.appendChild(script)
  })
}

async function handleGoogleCredential(response: { credential: string }) {
  googleError.value = ''
  googleLoading.value = true
  try {
    const res = (await httpClient.post('/auth/google/verify', {
      credential: response.credential,
    })) as GoogleVerifyResponse
    authStore.setSession(
      {
        id: res.user.id,
        email: res.user.email,
        fullName: res.user.full_name,
      },
      res.access_token
    )
    await router.push({ name: 'Main' })
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } }; message?: string }
    googleError.value =
      err.response?.data?.error === 'invalid_credential'
        ? 'Đăng nhập Google thất bại. Vui lòng thử lại.'
        : err.message ?? 'Đăng nhập thất bại.'
  } finally {
    googleLoading.value = false
  }
}

let googleInited = false

function ensureGoogleInited() {
  const g = (window as unknown as {
    google?: {
      accounts: {
        id: {
          initialize: (c: { client_id: string; callback: (r: { credential: string }) => void; auto_select: boolean }) => void
          prompt: () => void
        }
      }
    }
  }).google
  if (!g?.accounts?.id || !env.googleClientId) return
  if (!googleInited) {
    g.accounts.id.initialize({
      client_id: env.googleClientId,
      callback: handleGoogleCredential,
      auto_select: false,
    })
    googleInited = true
  }
  g.accounts.id.prompt()
}

async function onGoogleLogin() {
  error.value = ''
  googleError.value = ''
  if (!env.googleClientId) {
    googleError.value = 'Chưa cấu hình Google Client ID'
    return
  }
  await loadGoogleScript()
  ensureGoogleInited()
}

onMounted(() => {
  const token = route.query.token as string | undefined
  const err = route.query.error as string | undefined
  if (token || err) {
    const ok = authStore.handleOAuthReturn({ token, error: err })
    if (ok) {
      router.replace({ name: 'Main', query: {} })
      return
    }
    error.value = err ? t('auth.errorLoginFailed') : t('auth.errorLoginFailed')
  }
})

onUnmounted(() => {
  const g = (window as unknown as { google?: { accounts: { id: { cancel: () => void } } } }).google
  g?.accounts?.id?.cancel?.()
})

async function onSubmit() {
  error.value = ''
  if (!form.email.trim() || !form.password) {
    error.value = t('auth.errorFillRequired')
    return
  }
  try {
    await authStore.login({ email: form.email.trim(), password: form.password })
  } catch (e) {
    const err = e as { response?: { data?: { error?: string } } }
    const code = err.response?.data?.error
    if (code === 'invalid_credentials' || code === 'update_failed') {
      const msg = 'Tài khoản không tồn tại'
      error.value = msg
    } else if (code === 'token_failed') {
      const msg = 'Đăng nhập quá hạn'
      error.value = msg
    } else {
      error.value = e instanceof Error ? e.message : t('auth.errorLoginFailed')
    }
  }
}

function onFacebookLogin() {
  error.value = ''
  googleError.value = ''
  authStore.loginWithFacebook()
}

function onXLogin() {
  error.value = ''
  googleError.value = ''
  authStore.loginWithX()
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12">
    <div class="w-full max-w-md">
      <div class="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
        <h1 class="text-center text-2xl font-bold text-gray-900">
          {{ t('auth.login') }}
        </h1>

        <form class="mt-6 space-y-4" @submit.prevent="onSubmit">
          <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

          <div>
            <label for="login-email" class="mb-1 block text-sm font-medium text-gray-700">
              {{ t('auth.email') }}
            </label>
            <input
              id="login-email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              class="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              :placeholder="t('auth.email')"
            />
          </div>

          <div>
            <label for="login-password" class="mb-1 block text-sm font-medium text-gray-700">
              {{ t('auth.password') }}
            </label>
            <input
              id="login-password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              class="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              :placeholder="t('auth.password')"
            />
          </div>

          <button
            type="submit"
            class="w-full rounded-lg bg-emerald-600 py-2.5 text-sm font-medium text-white transition hover:bg-emerald-700 disabled:opacity-50"
            :disabled="authStore.loading"
          >
            {{ authStore.loading ? t('common.loading') : t('auth.submitLogin') }}
          </button>

          <p class="py-2 text-center text-sm text-gray-500">
            {{ t('auth.orContinueWith') }}
          </p>
          <div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
            <button
              type="button"
              class="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white py-2.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:opacity-50"
              :disabled="authStore.loading"
              :aria-label="t('auth.loginWithFacebook')"
              @click="onFacebookLogin"
            >
              <svg class="h-5 w-5" viewBox="0 0 24 24" fill="#1877F2" aria-hidden="true">
                <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
              </svg>
              <span class="hidden sm:inline">Facebook</span>
            </button>
            <button
              type="button"
              class="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white py-2.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:opacity-50"
              :disabled="authStore.loading || googleLoading"
              :aria-label="t('auth.loginWithGoogle')"
              @click="onGoogleLogin"
            >
              <svg class="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              <span class="hidden sm:inline">Google</span>
            </button>
            <button
              type="button"
              class="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white py-2.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:opacity-50"
              :disabled="authStore.loading"
              :aria-label="t('auth.loginWithX')"
              @click="onXLogin"
            >
              <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              <span class="hidden sm:inline">Twitter</span>
            </button>
          </div>
          <p v-if="googleError" class="text-sm text-red-600">{{ googleError }}</p>
          <p v-if="googleLoading" class="text-sm text-gray-500">Đang xử lý đăng nhập Google...</p>
        </form>

        <p class="mt-4 text-center text-sm text-gray-600">
          {{ t('auth.noAccount') }}
          <RouterLink
            :to="{ name: 'Register' }"
            class="font-medium text-emerald-600 hover:text-emerald-700"
          >
            {{ t('auth.register') }}
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>
