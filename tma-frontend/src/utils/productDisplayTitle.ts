import { Product } from '../types/product'
function stripPlatformSuffix(title: string): string {
  return title
    .replace(/\s*\(PS\/XBOX\/PC\)\s*$/i, '')
    .replace(/\s*\(Windows\)\s*$/i, '')
    .replace(/\s*\(PS4\)\s*$/i, '')
    .replace(/\s*\(PS5\)\s*$/i, '')
    .replace(/\s*\(Xbox\)\s*$/i, '')
    .trim()
}

export function getProductDisplayTitle(product: Pick<Product, 'id' | 'title' | 'title_key'>): string {
  if (product.title.includes('(Windows)')) {
    return product.title.replace('(Windows)', '(PS/XBOX/PC)')
  }
  return product.title
}

export function getUnifiedProductTitle(
  product: Pick<Product, 'id' | 'title' | 'title_key'>,
  multiPlatform = false,
): string {
  const base = getProductDisplayTitle(product)
  if (!multiPlatform) return base
  if (base.includes('(PS/XBOX/PC)')) return base
  const stripped = stripPlatformSuffix(base)
  return `${stripped} (PS/XBOX/PC)`
}
