import { createI18n } from 'vue-i18n'
import vi from './locales/vi'
import en from './locales/en'

export type Locale = 'vi' | 'en'

const LOCALE_STORAGE_KEY = 'tshop_locale'

function getInitialLocale(): Locale {
  try {
    const saved = localStorage.getItem(LOCALE_STORAGE_KEY) as Locale | null
    if (saved === 'vi' || saved === 'en') return saved
  } catch {
    /* ignore */
  }
  return 'vi'
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en',
  messages: {
    vi,
    en,
  },
})

export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale
  try {
    localStorage.setItem(LOCALE_STORAGE_KEY, locale)
  } catch {
    /* ignore */
  }
}
