import type { Product } from '../types/product'

export function isPreorderProduct(product: Pick<Product, 'game_section' | 'release_date'>) {
  if (product.game_section === 'preorder') return true
  if (product.release_date) return new Date(product.release_date).getTime() > Date.now()
  return false
}

export function formatReleaseDate(value?: string | null) {
  if (!value) return 'Дата уточняется'
  return new Date(value).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}
