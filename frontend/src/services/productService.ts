import { httpClient, env } from '@/core'
import type { Product } from '@/modules/product/domain'

export async function fetchProducts(): Promise<Product[]> {
  // Gateway proxy: /products -> product-service /api/products
  return httpClient.get<Product[], any, {}>('/products', {
    baseURL: env.productBaseUrl,
  })
}

