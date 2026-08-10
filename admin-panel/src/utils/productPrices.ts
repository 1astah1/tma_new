export type ProductPricesData = {
  tr?: number
  ua?: number
  xbox?: number
  us?: number
  edition_catalog?: unknown
}

export function parseProductPrices(raw: unknown): ProductPricesData {
  if (!raw) return {}
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw) as ProductPricesData
    } catch {
      return {}
    }
  }
  if (typeof raw === 'object') return raw as ProductPricesData
  return {}
}

export function buildProductPricesPayload(
  platform: string,
  form: { prices_tr?: number | string | null; prices_ua?: number | string | null; prices_xbox?: number | string | null },
  existing: ProductPricesData,
): ProductPricesData {
  const out: ProductPricesData = { ...existing }
  if (existing.edition_catalog) {
    out.edition_catalog = existing.edition_catalog
  }

  const setNum = (key: keyof ProductPricesData, value: number | string | null | undefined) => {
    if (value === '' || value === null || value === undefined) {
      delete out[key]
      return
    }
    const num = Number(value)
    if (Number.isFinite(num) && num > 0) {
      ;(out as Record<string, number>)[key as string] = num
    } else {
      delete out[key]
    }
  }

  if (platform === 'ps4' || platform === 'ps5') {
    setNum('tr', form.prices_tr)
    setNum('ua', form.prices_ua)
    delete out.xbox
    delete out.us
  } else if (platform === 'xbox') {
    const xboxVal = form.prices_xbox
    if (xboxVal !== '' && xboxVal != null && Number(xboxVal) > 0) {
      const num = Number(xboxVal)
      out.xbox = num
      out.us = num
    } else {
      delete out.xbox
      delete out.us
    }
    delete out.tr
    delete out.ua
  }

  return out
}
