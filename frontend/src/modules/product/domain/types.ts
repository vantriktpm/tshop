/**
 * Product domain types
 */
export interface Product {
  id: string
  name: string
  description: string
  price: number
  imageUrl: string
  quantity?: number
}

export interface ProductCartItem extends Product {
  quantity: number
}
