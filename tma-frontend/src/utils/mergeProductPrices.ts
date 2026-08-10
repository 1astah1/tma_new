import { Product, ProductEditionCatalog, ProductPrices } from '../types/product'
import { productGroupKey } from './groupProducts'
import { parseProductPrices } from './productEditionPricing'

function mergeEditionCatalogs(
  base?: ProductEditionCatalog,
  extra?: ProductEditionCatalog,
): ProductEditionCatalog | undefined {
  if (!base && !extra) return undefined
  const out: ProductEditionCatalog = { ...(base ?? {}) }
  if (!extra) return out

  for (const key of ['ps_tr', 'ps_ua', 'xbox'] as const) {
    const items = extra[key]
    if (!items?.length) continue
    const existing = out[key] ?? []
    const seen = new Map(existing.map((item) => [item.id, item]))
    for (const item of items) {
      const prev = seen.get(item.id)
      if (!prev || item.price < prev.price) {
        seen.set(item.id, item)
      }
    }
    out[key] = Array.from(seen.values())
  }

  return out
}

function mergeProductPricesFields(target: ProductPrices, source: ProductPrices): ProductPrices {
  const merged: ProductPrices = { ...target }

  if (source.tr != null) merged.tr = source.tr
  if (source.ua != null) merged.ua = source.ua
  if (source.xbox != null) merged.xbox = source.xbox
  if (source.us != null) merged.us = source.us

  const catalog = mergeEditionCatalogs(merged.edition_catalog, source.edition_catalog)
  if (catalog) {
    merged.edition_catalog = catalog
  }

  return merged
}

export function mergeSiblingProductPrices(product: Product, siblings: Product[]): Product {
  if (siblings.length <= 1) return product

  let prices = parseProductPrices(product.prices) ?? {}
  for (const sibling of siblings) {
    if (sibling.id === product.id) continue
    const parsed = parseProductPrices(sibling.prices)
    if (parsed) {
      prices = mergeProductPricesFields(prices, parsed)
    }
  }

  if (Object.keys(prices).length === 0) {
    return product
  }

  return { ...product, prices }
}

/** Подтягивает цены sibling-SKU (PS/Xbox) из общего пула каталога (главная, каталог). */
export function enrichProductsWithTitleKeyPool(products: Product[], pool: Product[]): Product[] {
  const buckets = new Map<string, Product[]>()

  for (const item of pool) {
    if (item.type !== 'game') continue
    const key = productGroupKey(item)
    if (!key) continue
    const list = buckets.get(key) ?? []
    list.push(item)
    buckets.set(key, list)
  }

  return products.map((product) => {
    if (product.type !== 'game') return product
    const key = productGroupKey(product)
    if (!key) return product
    const siblings = buckets.get(key) ?? [product]
    return mergeSiblingProductPrices(product, siblings)
  })
}
