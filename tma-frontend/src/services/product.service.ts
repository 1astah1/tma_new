import api from './api'
import { Product } from '../types/product'

export interface ProductFilters {
  platform?: string
  type?: string
  search?: string
  min_price?: number
  max_price?: number
  page?: number
  limit?: number
}

export async function getProducts(filters: ProductFilters = {}) {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([k, v]) => {
    if (v !== undefined && v !== '') params.set(k, String(v))
  })
  const { data } = await api.get(`/products?${params}`)
  return data as { data: Product[]; meta: { page: number; limit: number; total: number } }
}

export async function getProduct(id: string) {
  const { data } = await api.get(`/products/${id}`)
  return data as Product
}

export async function getPlatforms() {
  const { data } = await api.get('/platforms')
  return data as { id: string; name: string }[]
}
