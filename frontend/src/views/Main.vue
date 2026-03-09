<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useProductStore, useAuthStore } from '@/stores'
import { setLocale, type Locale } from '@/i18n'
import type { Product } from '@/modules/product/domain'
import { useAvatarSocket } from '@/core/websocket/useAvatarSocket'

const { t, locale } = useI18n()
const productStore = useProductStore()
const authStore = useAuthStore()
const locales: { value: Locale; labelKey: string }[] = [
  { value: 'vi', labelKey: 'locale.vi' },
  { value: 'en', labelKey: 'locale.en' },
]

function switchLocale(newLocale: Locale) {
  setLocale(newLocale)
}
const quantity = ref<Record<string, number>>({})

// WebSocket: khi backend push avatar.saved (user_id, image_id) → gọi GET image theo image_id và hiển thị
const { avatarDownloadUrl, clearAvatar } = useAvatarSocket(computed(() => authStore.user?.id))
const avatarUrl = computed(() => avatarDownloadUrl.value)

const userDisplayName = computed(
  () => authStore.user?.fullName ?? authStore.user?.email ?? ''
)
const userInitial = computed(() =>
  userDisplayName.value ? userDisplayName.value.charAt(0).toUpperCase() : '?'
)

onMounted(() => {
  productStore.fetchProducts()
})

function getQty(product: Product) {
  return quantity.value[product.id] ?? 1
}

function setQty(product: Product, qty: number) {
  quantity.value[product.id] = Math.max(1, Math.floor(qty))
}

function order(product: Product) {
  const qty = getQty(product)
  productStore.addToCart(product, qty)
}

/** Định dạng số với dấu chấm ngăn cách hàng nghìn */
function formatPrice(value: number): string {
  return Math.round(value).toLocaleString('de-DE')
}
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4 py-8">
    <!-- Khối nội dung căn giữa màn hình (bỏ w-full để block co theo nội dung, flex mới căn giữa được) -->
    <div class="max-w-6xl w-full">
      <header class="flex flex-wrap items-center justify-between gap-4 py-4">
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-600">{{ t('common.language') }}:</span>
          <div class="flex rounded-lg border border-gray-200 bg-white p-0.5 shadow-sm">
            <button
              v-for="item in locales"
              :key="item.value"
              type="button"
              :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition',
                locale === item.value
                  ? 'bg-emerald-600 text-white'
                  : 'text-gray-600 hover:bg-gray-100',
              ]"
              @click="switchLocale(item.value)"
            >
              {{ t(item.labelKey) }}
            </button>
          </div>
        </div>
        <div class="ml-auto flex items-center gap-3">
          <template v-if="authStore.isAuthenticated">
            <div class="h-8 w-8 overflow-hidden rounded-full bg-emerald-100">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="userDisplayName || 'avatar'"
                class="h-full w-full object-cover"
                @error="clearAvatar()"
              />
              <div
                v-else
                class="flex h-full w-full items-center justify-center bg-emerald-600 text-xs font-semibold text-white"
              >
                {{ userInitial }}
              </div>
            </div>
            <span class="text-sm text-gray-600">{{ userDisplayName }}</span>
            <button
              type="button"
              class="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
              @click="authStore.logout()"
            >
              {{ t('auth.logout') }}
            </button>
          </template>
          <template v-else>
            <RouterLink
              :to="{ name: 'Login' }"
              class="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
            >
              {{ t('auth.login') }}
            </RouterLink>
            <RouterLink
              :to="{ name: 'Register' }"
              class="rounded-lg bg-emerald-600 px-3 py-1.5 text-sm text-white hover:bg-emerald-700"
            >
              {{ t('auth.register') }}
            </RouterLink>
          </template>
        </div>
      </header>

      <div class="flex flex-wrap justify-start gap-6">
      <!-- Product card: parent div 350px, responsive, flex column -->
      <div
        v-for="product in productStore.products"
        :key="product.id"
        class="flex w-full max-w-[350px] flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition hover:shadow-md"
      >
        <!-- Ảnh -->
        <div class="aspect-square w-full overflow-hidden bg-gray-100">
          <img
            :src="product.imageUrl"
            :alt="product.name"
            class="h-full w-full object-cover"
            @error="($event.target as HTMLImageElement).src = 'https://placehold.co/350x350/e5e7eb/6b7280?text=' + encodeURIComponent(t('product.noImage'))"
          />
        </div>

        <!-- Thông tin sản phẩm -->
        <div class="flex flex-1 flex-col p-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ product.name }}</h3>
          <p class="mt-1 line-clamp-2 text-sm text-gray-500">
            {{ product.description }}
          </p>

          <!-- Giá - Số lượng -->
          <div class="mt-4 flex flex-wrap items-center justify-between gap-3">
            <span class="text-xl font-bold text-emerald-600">
              {{ formatPrice(product.price) }} {{ t('common.currency') }}
            </span>
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600">{{ t('common.quantity') }}:</label>
              <input
                type="number"
                :value="getQty(product)"
                min="1"
                class="w-16 rounded border border-gray-300 px-2 py-1 text-center text-sm"
                @input="setQty(product, Number(($event.target as HTMLInputElement).value))"
              />
            </div>
          </div>

          <!-- Đặt hàng -->
          <button
            type="button"
            class="mt-4 w-full rounded-lg bg-emerald-600 py-2.5 text-sm font-medium text-white transition hover:bg-emerald-700"
            @click="order(product)"
          >
            {{ t('common.order') }}
          </button>
        </div>
      </div>
    </div>
    </div>

    <!-- Base loading -->
    <div
      v-if="productStore.loading"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/20"
    >
      <div class="rounded-lg bg-white px-6 py-4 shadow-lg">{{ t('common.loading') }}</div>
    </div>
  </div>
</template>
