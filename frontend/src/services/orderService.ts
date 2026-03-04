import { httpClient, env } from '@/core'
import type { Product } from '@/modules/product/domain'

export interface CreateOrderItem {
  productId: string
  quantity: number
  price: number
}

export interface CreateOrderPayload {
  items: CreateOrderItem[]
  totalAmount: number
  note?: string
}

export interface Order {
  id: string
  items: Array<CreateOrderItem & { product?: Product }>
  totalAmount: number
  status: string
  createdAt: string
}

export async function createOrder(payload: CreateOrderPayload): Promise<Order> {
  // Gateway proxy: /orders -> order-service /api/orders
  return httpClient.post<Order, any, {}>('/orders', payload, {
    baseURL: env.orderBaseUrl,
  })
}

