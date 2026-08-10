import { Product } from '../types/product'

export function matchesProductSection(product: Product, section: string): boolean {
  switch (section) {
    case 'new':
      return product.type === 'game' && product.game_section === 'new'
    case 'preorder':
      return product.type === 'game' && product.game_section === 'preorder'
    case 'game':
    default:
      return product.type === 'game'
  }
}

export function filterProductsBySection(products: Product[], section: string): Product[] {
  const activeSection = section || 'game'
  return products.filter((product) => matchesProductSection(product, activeSection))
}
