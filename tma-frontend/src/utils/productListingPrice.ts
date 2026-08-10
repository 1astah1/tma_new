import { MIN_LISTING_PRICE } from '../constants/pricing'
import { Product } from '../types/product'
import {
  EditionPlatformKey,
  editionKeyMatchesFamily,
  getEditionCatalog,
  listEditionPlatforms,
  productPlatformFamily,
  ProductPlatformFamily,
} from './productEditionPricing'
import { isPreorderProduct } from './productPreorder'

export { MIN_LISTING_PRICE }

export function coercePrice(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) {
      return parsed
    }
  }
  return null
}

export function isValidListingPrice(price: number): boolean {
  return Number.isFinite(price) && price >= MIN_LISTING_PRICE
}

export function isPlaceholderPrice(price: number): boolean {
  return price > 0 && price < MIN_LISTING_PRICE
}

export function effectiveProductPrice(product: Product): number {
  if (isPlaceholderPrice(product.price)) {
    return Number.POSITIVE_INFINITY
  }
  if (isPreorderProduct(product) && product.price <= 0) {
    return Number.POSITIVE_INFINITY
  }
  if (!isPreorderProduct(product) && product.discount_percent > 0) {
    return product.price * (1 - product.discount_percent / 100)
  }
  return product.price
}

function pushValidPrice(candidates: number[], value: unknown) {
  const price = coercePrice(value)
  if (price != null && isValidListingPrice(price)) {
    candidates.push(price)
  }
}

function collectRegionalPrices(
  prices: Product['prices'],
  candidates: number[],
  families?: ProductPlatformFamily[],
) {
  if (!prices || typeof prices !== 'object') return
  const record = prices as Record<string, unknown>
  const allowPs = !families || families.includes('ps')
  const allowXbox = !families || families.includes('xbox') || families.includes('pc')

  if (allowPs) {
    pushValidPrice(candidates, record.tr)
    pushValidPrice(candidates, record.ua)
  }
  if (allowXbox) {
    pushValidPrice(candidates, record.xbox)
    pushValidPrice(candidates, record.us)
  }
}

function collectEditionCatalogPrices(
  product: Product,
  candidates: number[],
  families?: ProductPlatformFamily[],
) {
  const catalog = getEditionCatalog(product.id, product.prices, product.title_key, product.platform)
  if (!catalog) return

  for (const key of listEditionPlatforms(catalog)) {
    if (families && !families.some((family) => editionKeyMatchesFamily(key, family))) {
      continue
    }
    for (const edition of catalog[key] ?? []) {
      pushValidPrice(candidates, edition.price)
    }
  }
}

export function collectProductPriceCandidates(
  product: Product,
  families?: ProductPlatformFamily[],
): number[] {
  const candidates: number[] = []

  collectEditionCatalogPrices(product, candidates, families)
  collectRegionalPrices(product.prices, candidates, families)

  if (candidates.length > 0) {
    return candidates
  }

  // Предзаказ: не берём product.price из автоимпорта — только prices / edition_catalog.
  if (isPreorderProduct(product)) {
    return candidates
  }

  const productFamily = productPlatformFamily(product.platform)
  if (
    families &&
    families.length > 0 &&
    !families.includes(productFamily) &&
    productFamily !== 'other'
  ) {
    return candidates
  }

  for (const variant of product.variants ?? []) {
    pushValidPrice(candidates, variant.price)
  }
  for (const variant of product.platform_variants ?? []) {
    if (families && families.length > 0) {
      const variantFamily = productPlatformFamily(variant.platform)
      if (!families.includes(variantFamily) && variantFamily !== 'other') continue
    }
    pushValidPrice(candidates, variant.price)
  }

  if (candidates.length > 0) {
    return candidates
  }

  const base = effectiveProductPrice(product)
  if (isValidListingPrice(base)) {
    candidates.push(base)
  }

  return candidates
}

/** Цены для карточки: учитывает merge sibling-цен с API (PS-цены на Xbox-SKU и т.п.). */
export function collectCardPriceCandidates(
  product: Product,
  cardFamilies: ProductPlatformFamily[],
): number[] {
  const scoped = collectProductPriceCandidates(product, cardFamilies)
  if (scoped.length > 0) {
    return scoped
  }

  if (!isPreorderProduct(product)) {
    return scoped
  }

  // Предзаказ: на главной у SKU одной платформы могут лежать цены sibling (PS TR/UA).
  return collectProductPriceCandidates(product)
}

export function getMinListingPrice(
  product: Product,
  families?: ProductPlatformFamily[],
): number | null {
  const prices = collectProductPriceCandidates(product, families)
  if (!prices.length) return null
  return Math.min(...prices)
}
