/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_USER_BASE_URL: string
  readonly VITE_ORDER_BASE_URL: string
  readonly VITE_PRODUCT_BASE_URL: string
  readonly VITE_CART_BASE_URL: string
  readonly VITE_INVENTORY_BASE_URL: string
  readonly VITE_PAYMENT_BASE_URL: string
  readonly VITE_SHIPPING_BASE_URL: string
  readonly VITE_PROMOTION_BASE_URL: string
  readonly VITE_NOTIFICATION_BASE_URL: string
  readonly VITE_IMAGE_BASE_URL: string
  readonly VITE_APP_ENV: 'development' | 'production'
  readonly VITE_GOOGLE_CLIENT_ID: string
  readonly VITE_FACEBOOK_APP_ID: string
  readonly VITE_X_CLIENT_ID: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
