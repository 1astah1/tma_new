import api from './api'
import { Product } from '../types/product'
import { HomeBanner, HomeCategoryDetail, HomeCategoryListItem } from '../types/home'

export interface ProductFilters {
  platform?: string
  type?: string
  section?: string
  search?: string
  min_price?: number
  max_price?: number
  page?: number
  limit?: number
}

type ProductListResponse = {
  data: Product[]
  meta: { page: number; limit: number; total: number }
}

function unwrapList(data: unknown): ProductListResponse {
  const payload = data as ProductListResponse
  return {
    data: Array.isArray(payload?.data) ? payload.data : [],
    meta: payload?.meta ?? { page: 1, limit: 20, total: 0 },
  }
}

function unwrapItems(data: unknown): Product[] {
  const payload = data as { data?: Product[] }
  return Array.isArray(payload?.data) ? payload.data : []
}

export async function getProducts(filters: ProductFilters = {}) {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([k, v]) => {
    if (v !== undefined && v !== '') params.set(k, String(v))
  })
  const { data } = await api.get(`/products?${params}`)
  return unwrapList(data)
}

export async function getProduct(id: string) {
  const { data } = await api.get(`/products/${id}`)
  return data as Product
}

export async function getPopularProducts() {
  const { data } = await api.get('/content/popular-products')
  return { data: unwrapItems(data) }
}

export type HomeCatalog = {
  preorders: Product[]
  new_releases: Product[]
  popular: Product[]
}

export async function getHomeCatalog() {
  const { data } = await api.get('/content/home-feed')
  const payload = (data as { data?: HomeCatalog })?.data ?? (data as HomeCatalog)
  return {
    preorders: Array.isArray(payload?.preorders) ? payload.preorders : [],
    new_releases: Array.isArray(payload?.new_releases) ? payload.new_releases : [],
    popular: Array.isArray(payload?.popular) ? payload.popular : [],
  }
}

export async function getHomeBanners() {
  const { data } = await api.get('/content/home-banners')
  return { data: unwrapItems(data) as unknown as HomeBanner[] }
}

export async function getHomeCategories() {
  const { data } = await api.get('/content/home-categories')
  return { data: unwrapItems(data) as unknown as HomeCategoryListItem[] }
}

export async function getHomeCategory(id: string) {
  const { data } = await api.get(`/content/home-categories/${id}`)
  const payload = data as { data?: HomeCategoryDetail }
  return { data: payload.data as HomeCategoryDetail }
}

export async function getPlatforms() {
  const { data } = await api.get('/platforms')
  return data as { id: string; name: string }[]
}
