import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Product } from '@/modules/product/domain'
import { useLoading } from '@/core/loading'
// import { httpClient } from '@/core/http' // use when calling API: httpClient.get<Product[]>('/products')

export const useProductStore = defineStore('product', () => {
  const { loading, withLoading } = useLoading()
  const products = ref<Product[]>([])
  const selectedQuantity = ref<Record<string, number>>({})

  const getProductQuantity = (id: string) => computed(() => selectedQuantity.value[id] ?? 1)

  function setQuantity(productId: string, qty: number) {
    selectedQuantity.value[productId] = Math.max(1, qty)
  }

  async function fetchProducts() {
    return withLoading(async () => {
      // Example: const data = await httpClient.get<Product[]>('/products')
      // products.value = data
      // Mock data for UI
      products.value = [
        {
          id: '1',
          name: 'Áo thun basic',
          description: 'Chất liệu cotton mềm, form rộng thoải mái.',
          price: 199000,
          imageUrl: 'https://placehold.co/350x350/e5e7eb/6b7280?text=Ao+thun',
        },
        {
          id: '2',
          name: 'Quần jean slim',
          description: 'Quần jean co giãn, kiểu dáng slim fit.',
          price: 349000,
          imageUrl: 'https://placehold.co/350x350/e5e7eb/6b7280?text=Quan+jean',
        },
      ]
      return products.value
    })
  }

  function addToCart(product: Product, qty: number) {
    setQuantity(product.id, qty)
    // Integrate with cart store or API
    console.log('Đặt hàng:', product.name, 'x', qty)
  }

  return {
    products,
    loading,
    selectedQuantity,
    getProductQuantity,
    setQuantity,
    fetchProducts,
    addToCart,
  }
})
