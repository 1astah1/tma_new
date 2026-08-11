import { MIN_LISTING_PRICE } from '../constants/pricing'
import { COD_MW4_EDITION_CATALOG, COD_MW4_PRODUCT_ID, COD_MW4_TITLE_KEY } from '../constants/codMw4Pricing'
import { Product, ProductEditionCatalog, ProductEditionOption, ProductPrices } from '../types/product'

export type EditionPlatformKey = keyof ProductEditionCatalog

export const EDITION_PLATFORM_LABELS: Record<EditionPlatformKey, { title: string; region: string }> = {
  ps_tr: { title: 'PlayStation', region: 'Турция TR' },
  ps_ua: { title: 'PlayStation', region: 'Украина UA' },
  xbox: { title: 'Xbox/PC', region: 'США' },
  xbox_tr: { title: 'Xbox/PC', region: 'Турция TR' },
}

function isEditionCatalog(value: unknown): value is ProductEditionCatalog {
  if (!value || typeof value !== 'object') return false
  const catalog = value as ProductEditionCatalog
  return ['ps_tr', 'ps_ua', 'xbox', 'xbox_tr'].some((key) => {
    const list = catalog[key as EditionPlatformKey]
    return Array.isArray(list) && list.length > 0
  })
}

function isValidEditionPrice(price: number): boolean {
  return Number.isFinite(price) && price >= MIN_LISTING_PRICE
}

export function parseProductPrices(raw?: ProductPrices | Record<string, unknown>): ProductPrices | undefined {
  if (!raw || typeof raw !== 'object') return undefined
  return raw as ProductPrices
}

function catalogFromRegionalPrices(
  prices: ProductPrices,
  platform: string,
): ProductEditionCatalog | null {
  const catalog: ProductEditionCatalog = {}

  if (platform === 'ps4' || platform === 'ps5') {
    if (typeof prices.tr === 'number' && isValidEditionPrice(prices.tr)) {
      catalog.ps_tr = [{ id: 'standard', name: 'Standard Edition', price: prices.tr }]
    }
    if (typeof prices.ua === 'number' && isValidEditionPrice(prices.ua)) {
      catalog.ps_ua = [{ id: 'standard', name: 'Standard Edition', price: prices.ua }]
    }
  } else if (platform === 'xbox' || platform === 'pc') {
    const xboxPrice = prices.xbox ?? prices.us
    if (typeof xboxPrice === 'number' && isValidEditionPrice(xboxPrice)) {
      catalog.xbox = [{ id: 'standard', name: 'Standard Edition', price: xboxPrice }]
    }
    if (typeof prices.xbox_tr === 'number' && isValidEditionPrice(prices.xbox_tr)) {
      catalog.xbox_tr = [{ id: 'standard', name: 'Standard Edition', price: prices.xbox_tr }]
    }
  }

  return isEditionCatalog(catalog) ? catalog : null
}

export function getEditionCatalog(
  productId: string,
  prices?: ProductPrices | Record<string, unknown>,
  titleKey?: string,
  platform?: string,
): ProductEditionCatalog | null {
  const parsed = parseProductPrices(prices)
  if (parsed?.edition_catalog && isEditionCatalog(parsed.edition_catalog)) {
    const cleaned = filterValidEditionCatalog(parsed.edition_catalog)
    if (cleaned) return cleaned
  }
  if (parsed && platform) {
    const regional = catalogFromRegionalPrices(parsed, platform)
    if (regional) return regional
  }
  const key = titleKey?.trim().toLowerCase()
  if (productId === COD_MW4_PRODUCT_ID || key === COD_MW4_TITLE_KEY) {
    return COD_MW4_EDITION_CATALOG
  }
  return null
}

function filterValidEditionCatalog(catalog: ProductEditionCatalog): ProductEditionCatalog | null {
  const filtered: ProductEditionCatalog = {}
  for (const key of listEditionPlatforms(catalog)) {
    const editions = (catalog[key] ?? []).filter((edition) => isValidEditionPrice(edition.price))
    if (editions.length > 0) {
      filtered[key] = editions
    }
  }
  return isEditionCatalog(filtered) ? filtered : null
}

export function listEditionPlatforms(catalog: ProductEditionCatalog): EditionPlatformKey[] {
  return (['ps_tr', 'ps_ua', 'xbox', 'xbox_tr'] as EditionPlatformKey[]).filter(
    (key) => (catalog[key]?.length ?? 0) > 0,
  )
}

export type ProductPlatformFamily = 'ps' | 'xbox' | 'pc' | 'other'

export function productPlatformFamily(platform: string): ProductPlatformFamily {
  if (platform === 'ps4' || platform === 'ps5') return 'ps'
  if (platform === 'xbox') return 'xbox'
  if (platform === 'pc') return 'pc'
  return 'other'
}

export function editionKeyMatchesFamily(key: EditionPlatformKey, family: ProductPlatformFamily): boolean {
  if (family === 'ps') return key === 'ps_tr' || key === 'ps_ua'
  if (family === 'xbox' || family === 'pc') return key === 'xbox' || key === 'xbox_tr'
  return false
}

export function hasMultiPlatformEditionCatalog(catalog: ProductEditionCatalog): boolean {
  const hasPs = (catalog.ps_tr?.length ?? 0) > 0 || (catalog.ps_ua?.length ?? 0) > 0
  const hasXbox = (catalog.xbox?.length ?? 0) > 0 || (catalog.xbox_tr?.length ?? 0) > 0
  return hasPs && hasXbox
}

/** Платформы, для которых в prices / edition_catalog есть реальные цены. */
export function getPricePlatformFamilies(
  product: Pick<Product, 'id' | 'platform' | 'prices' | 'title_key'>,
): ProductPlatformFamily[] {
  const families = new Set<ProductPlatformFamily>()
  const catalog = getEditionCatalog(product.id, product.prices, product.title_key, product.platform)

  if (catalog) {
    if ((catalog.ps_tr?.length ?? 0) > 0 || (catalog.ps_ua?.length ?? 0) > 0) {
      families.add('ps')
    }
    if ((catalog.xbox?.length ?? 0) > 0 || (catalog.xbox_tr?.length ?? 0) > 0) {
      families.add('xbox')
    }
  }

  const parsed = parseProductPrices(product.prices)
  if (parsed) {
    if (
      (typeof parsed.tr === 'number' && isValidEditionPrice(parsed.tr)) ||
      (typeof parsed.ua === 'number' && isValidEditionPrice(parsed.ua))
    ) {
      families.add('ps')
    }
    const xboxPrice = parsed.xbox ?? parsed.us ?? parsed.xbox_tr
    if (typeof xboxPrice === 'number' && isValidEditionPrice(xboxPrice)) {
      families.add('xbox')
    }
  }

  const order: ProductPlatformFamily[] = ['ps', 'xbox', 'pc']
  return order.filter((family) => families.has(family))
}

export function productPlatformFamilies(
  product: Pick<Product, 'platform' | 'platform_variants'>,
): ProductPlatformFamily[] {
  const families = new Set<ProductPlatformFamily>()
  families.add(productPlatformFamily(product.platform))
  for (const variant of product.platform_variants ?? []) {
    families.add(productPlatformFamily(variant.platform))
  }
  return Array.from(families).filter((family) => family !== 'other')
}

/** Игра продаётся на нескольких платформах (PS + Xbox/PC) — одна страница с полным каталогом изданий. */
export function isMultiPlatformGame(
  product: Pick<Product, 'platform' | 'platform_variants'>,
  catalog?: ProductEditionCatalog | null,
): boolean {
  if (productPlatformFamilies(product).length > 1) return true
  if (catalog && hasMultiPlatformEditionCatalog(catalog)) return true
  return false
}

/** @deprecated используйте isMultiPlatformGame */
export function isUnifiedMultiPlatformProduct(
  product: Pick<Product, 'platform' | 'platform_variants'>,
  catalog?: ProductEditionCatalog | null,
): boolean {
  return isMultiPlatformGame(product, catalog)
}

export function shouldShowFullEditionCatalog(
  product: Pick<Product, 'platform' | 'platform_variants'>,
  catalog?: ProductEditionCatalog | null,
): boolean {
  return isMultiPlatformGame(product, catalog)
}

export function resolveEditionCatalogForProduct(
  catalog: ProductEditionCatalog,
  product: Pick<Product, 'platform' | 'platform_variants' | 'title_key'>,
): ProductEditionCatalog {
  if (shouldShowFullEditionCatalog(product, catalog)) {
    return catalog
  }

  const family = productPlatformFamily(product.platform)
  const filtered: ProductEditionCatalog = {}

  for (const key of listEditionPlatforms(catalog)) {
    if (editionKeyMatchesFamily(key, family)) {
      filtered[key] = catalog[key]
    }
  }

  return filtered
}

export function defaultEditionPlatform(
  product: Pick<Product, 'platform'>,
  platforms: EditionPlatformKey[],
): EditionPlatformKey | null {
  if (!platforms.length) return null

  const family = productPlatformFamily(product.platform)
  if (family === 'ps') {
    return platforms.includes('ps_tr') ? 'ps_tr' : platforms.find((key) => key.startsWith('ps')) ?? platforms[0]
  }
  if (family === 'xbox' || family === 'pc') {
    return platforms.includes('xbox') ? 'xbox' : platforms[0]
  }

  return platforms[0]
}

export function findEdition(
  catalog: ProductEditionCatalog,
  platform: EditionPlatformKey,
  editionId: string,
): ProductEditionOption | null {
  return catalog[platform]?.find((item) => item.id === editionId) ?? null
}

export function editionRegionLabel(platform: EditionPlatformKey): string {
  const meta = EDITION_PLATFORM_LABELS[platform]
  if (platform === 'xbox') {
    return 'XBOX/PC [США]'
  }
  if (platform === 'xbox_tr') {
    return 'XBOX/PC [ТУРЦИЯ]'
  }
  return `${meta.title.toUpperCase()} [${meta.region.toUpperCase()}]`
}
