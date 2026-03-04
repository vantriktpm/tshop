import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { i18n } from '@/i18n'
import { useAuthStore } from '@/stores'
import Main from '@/views/Main.vue'
import Login from '@/views/auth/Login.vue'
import Register from '@/views/auth/Register.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Main',
    component: Main,
    meta: { titleKey: 'route.home' },
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { titleKey: 'route.login', guestOnly: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { titleKey: 'route.register', guestOnly: true },
  },
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, _from, next) => {
  const titleKey = to.meta.titleKey as string | undefined
  if (titleKey) {
    document.title = `${i18n.global.t(titleKey)} | TShop`
  }
  const guestOnly = to.meta.guestOnly === true
  if (guestOnly && useAuthStore().isAuthenticated) {
    next({ name: 'Main' })
    return
  }
  next()
})
