import { useQuery } from '@tanstack/react-query'
import {
  getProducts,
  getProduct,
  getPopularProducts,
  getHomeCatalog,
  getHomeBanners,
  getHomeCategories,
  getHomeCategory,
  ProductFilters,
} from '../services/product.service'

const catalogQueryOptions = {
  staleTime: 0,
  refetchOnMount: 'always' as const,
  refetchOnWindowFocus: true,
  retry: 3,
  retryDelay: (attempt: number) => Math.min(1000 * 2 ** attempt, 8000),
}

export function useProducts(filters: ProductFilters = {}, enabled = true) {
  return useQuery({
    queryKey: ['products', filters],
    queryFn: () => getProducts(filters),
    enabled,
    ...catalogQueryOptions,
  })
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: ['product', id],
    queryFn: () => getProduct(id),
    enabled: !!id,
  })
}

export function usePopularProducts(enabled = true) {
  return useQuery({
    queryKey: ['popular-products'],
    queryFn: getPopularProducts,
    enabled,
    ...catalogQueryOptions,
  })
}

export function useHomeCatalog(enabled = true) {
  return useQuery({
    queryKey: ['home-catalog'],
    queryFn: getHomeCatalog,
    enabled,
    ...catalogQueryOptions,
  })
}

export function useHomeBanners() {
  return useQuery({
    queryKey: ['home-banners'],
    queryFn: getHomeBanners,
  })
}

export function useHomeCategories() {
  return useQuery({
    queryKey: ['home-categories'],
    queryFn: getHomeCategories,
  })
}

export function useHomeCategory(id: string) {
  return useQuery({
    queryKey: ['home-category', id],
    queryFn: () => getHomeCategory(id),
    enabled: !!id,
  })
}
