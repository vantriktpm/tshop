/**
 * Environment config - use Vite env variables
 */
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? ''

// Base URL cho từng service (nếu không set thì fallback về apiBaseUrl)
const userBaseUrl = import.meta.env.VITE_USER_BASE_URL ?? apiBaseUrl
const orderBaseUrl = import.meta.env.VITE_ORDER_BASE_URL ?? apiBaseUrl
const productBaseUrl = import.meta.env.VITE_PRODUCT_BASE_URL ?? apiBaseUrl
const cartBaseUrl = import.meta.env.VITE_CART_BASE_URL ?? apiBaseUrl
const inventoryBaseUrl = import.meta.env.VITE_INVENTORY_BASE_URL ?? apiBaseUrl
const paymentBaseUrl = import.meta.env.VITE_PAYMENT_BASE_URL ?? apiBaseUrl
const shippingBaseUrl = import.meta.env.VITE_SHIPPING_BASE_URL ?? apiBaseUrl
const promotionBaseUrl = import.meta.env.VITE_PROMOTION_BASE_URL ?? apiBaseUrl
const notificationBaseUrl = import.meta.env.VITE_NOTIFICATION_BASE_URL ?? apiBaseUrl
const imageBaseUrl = import.meta.env.VITE_IMAGE_BASE_URL ?? apiBaseUrl

/** Base URL của user-service (bỏ /api) cho OAuth start redirect */
export const oauthBaseUrl =
  userBaseUrl.replace(/\/api\/?$/, '') ||
  apiBaseUrl.replace(/\/api\/?$/, '') ||
  (typeof window !== 'undefined' ? window.location.origin : '')

/** Google Client ID cho Google Sign-In (One Tap / button) */
export const googleClientId =
  import.meta.env.VITE_GOOGLE_CLIENT_ID ?? ''

export const env = {
  apiBaseUrl,
  userBaseUrl,
  orderBaseUrl,
  productBaseUrl,
  cartBaseUrl,
  inventoryBaseUrl,
  paymentBaseUrl,
  shippingBaseUrl,
  promotionBaseUrl,
  notificationBaseUrl,
  imageBaseUrl,
  oauthBaseUrl,
  googleClientId,
  appEnv: (import.meta.env.VITE_APP_ENV ?? 'development') as 'development' | 'production',
  isDev: import.meta.env.DEV,
  isProd: import.meta.env.PROD,
} as const
