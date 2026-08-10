import { useQueries } from '@tanstack/react-query'
import { getProduct } from '../services/product.service'
import { useWishlist } from '../stores/wishlistStore'
import { Product } from '../types/product'

export function useWishlistProducts() {
  const { items } = useWishlist()

  const queries = useQueries({
    queries: items.map((id) => ({
      queryKey: ['product', id],
      queryFn: () => getProduct(id),
      staleTime: 30_000,
      retry: 1,
    })),
  })

  const isLoading = items.length > 0 && queries.some((q) => q.isLoading)
  const products = queries
    .map((q) => q.data)
    .filter((p): p is Product => !!p)

  const missingCount = items.length - products.length

  return {
    products,
    isLoading,
    totalItems: items.length,
    missingCount: isLoading ? 0 : missingCount,
  }
}
