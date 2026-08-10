import { Product } from '../types/product'
import { getProductDisplayTitle } from './productDisplayTitle'
import { getEditionCatalog, getPricePlatformFamilies, ProductPlatformFamily } from './productEditionPricing'
import { isPreorderProduct } from './productPreorder'
import {
  collectCardPriceCandidates,
  collectProductPriceCandidates,
  effectiveProductPrice,
  isValidListingPrice,
} from './productListingPrice'

export type PlatformFamily = ProductPlatformFamily

export type CatalogCard = {
  id: string
  primary: Product
  variants: Product[]
}

function normalizeTitleKey(title: string): string {
  return title
    .replace(/\(windows\)/gi, '')
    .replace(/\b(ps4|ps5|xbox|pc|playstation)\b/gi, '')
    .replace(/[®™©]/g, '')
    .toLowerCase()
    .replace(/\s+/g, ' ')
    .trim()
}

export function productGroupKey(product: Product): string {
  const fromField = product.title_key?.trim()
  if (fromField) return fromField
  return normalizeTitleKey(getProductDisplayTitle(product))
}

export function platformFamily(platform: string): PlatformFamily {
  if (platform === 'ps4' || platform === 'ps5') return 'ps'
  if (platform === 'xbox') return 'xbox'
  if (platform === 'pc') return 'pc'
  return 'other'
}

function platformPriority(platform: string): number {
  switch (platform) {
    case 'ps5':
      return 30
    case 'xbox':
      return 25
    case 'ps4':
      return 20
    case 'pc':
      return 15
    default:
      return 0
  }
}

export { effectiveProductPrice } from './productListingPrice'

function catalogScore(product: Product): number {
  return getEditionCatalog(product.id, product.prices, product.title_key, product.platform) ? 100 : 0
}

function hasListingPrice(product: Product): boolean {
  return collectProductPriceCandidates(product).length > 0
}

function pickBestProduct(products: Product[]): Product {
  return [...products].sort((a, b) => {
    const priceDiff = Number(hasListingPrice(b)) - Number(hasListingPrice(a))
    if (priceDiff !== 0) return priceDiff
    const catalogDiff = catalogScore(b) - catalogScore(a)
    if (catalogDiff !== 0) return catalogDiff
    const priorityDiff = platformPriority(b.platform) - platformPriority(a.platform)
    if (priorityDiff !== 0) return priorityDiff
    return effectiveProductPrice(a) - effectiveProductPrice(b)
  })[0]
}

function collapseByPlatformFamily(products: Product[]): Product[] {
  const byFamily = new Map<PlatformFamily, Product>()
  for (const product of products) {
    const family = platformFamily(product.platform)
    const existing = byFamily.get(family)
    if (!existing) {
      byFamily.set(family, product)
      continue
    }
    const best = pickBestProduct([existing, product])
    byFamily.set(family, best)
  }
  return Array.from(byFamily.values())
}

function buildCatalogCard(items: Product[]): CatalogCard {
  const variants = collapseByPlatformFamily(items)
  const primary = pickBestProduct(variants)
  return {
    id: primary.id,
    primary,
    variants,
  }
}

export function groupProductsForCatalog(products: Product[], crossPlatform = true): CatalogCard[] {
  if (!crossPlatform) {
    return products.map((product) => ({
      id: product.id,
      primary: product,
      variants: [product],
    }))
  }

  const groups = new Map<string, Product[]>()

  for (const product of products) {
    if (product.type !== 'game') continue
    const key = productGroupKey(product)
    if (!key) continue
    const bucket = groups.get(key) ?? []
    bucket.push(product)
    groups.set(key, bucket)
  }

  const emitted = new Set<string>()
  const result: CatalogCard[] = []

  for (const product of products) {
    if (product.type !== 'game') {
      result.push({ id: product.id, primary: product, variants: [product] })
      continue
    }

    const key = productGroupKey(product)
    if (!key) {
      result.push({ id: product.id, primary: product, variants: [product] })
      continue
    }

    if (emitted.has(key)) continue
    emitted.add(key)

    const items = groups.get(key) ?? [product]
    result.push(buildCatalogCard(items))
  }

  return result
}

function hasListingPriceForFamily(product: Product, family: PlatformFamily): boolean {
  return collectProductPriceCandidates(product, [family]).length > 0
}

export function getCatalogCardFamilies(card: CatalogCard): PlatformFamily[] {
  const pricedFamilies = new Set<PlatformFamily>()
  const fallbackFamilies = new Set<PlatformFamily>()

  for (const variant of card.variants) {
    fallbackFamilies.add(platformFamily(variant.platform))

    for (const family of getPricePlatformFamilies(variant)) {
      if (hasListingPriceForFamily(variant, family)) {
        pricedFamilies.add(family)
      }
    }
  }

  const families = pricedFamilies.size > 0 ? pricedFamilies : fallbackFamilies
  const order: PlatformFamily[] = ['ps', 'xbox', 'pc', 'other']
  return order.filter((family) => families.has(family))
}

export function getCatalogCardPrice(card: CatalogCard): {
  price: number | null
  from: boolean
  preorder: boolean
} {
  const preorder = card.variants.some((item) => isPreorderProduct(item))
  const families = getCatalogCardFamilies(card)
  const multiFamily = families.length > 1

  const priceBuckets = card.variants.map((item) => {
    if (multiFamily) {
      return families.flatMap((family) => collectProductPriceCandidates(item, [family]))
    }
    return collectCardPriceCandidates(item, families)
  })

  const prices = [...new Set(priceBuckets.flat())].filter((price) => isValidListingPrice(price))

  if (!prices.length) {
    return { price: null, from: false, preorder }
  }

  const hasMultipleSources =
    prices.length > 1 ||
    multiFamily ||
    priceBuckets.some((bucket) => bucket.length > 1)

  return {
    price: Math.min(...prices),
    from: hasMultipleSources,
    preorder,
  }
}
