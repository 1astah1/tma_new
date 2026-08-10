import { MIN_LISTING_PRICE } from '../constants/pricing'

export const MANAGER_PRICE_LABEL = 'Уточнить у менеджера'

export function hasDisplayablePrice(price: number | null | undefined): boolean {
  return typeof price === 'number' && Number.isFinite(price) && price >= MIN_LISTING_PRICE
}

export function formatPrice(price: number): string {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(Math.round(price))
}

export function formatFromPrice(price: number): string {
  return `ОТ ${formatPrice(price)}`
}

function formatCardAmount(price: number): string {
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(Math.round(price))
}

/** Компактный формат для узких карточек каталога (без пробела перед ₽). */
export function formatCardFromPrice(price: number, from = true): string {
  const amount = formatCardAmount(price)
  return from ? `от ${amount}₽` : `${amount}₽`
}

export function formatPriceOrManager(price: number | null | undefined): string {
  if (!hasDisplayablePrice(price)) return MANAGER_PRICE_LABEL
  return formatPrice(price!)
}

export function formatCardPriceOrManager(
  price: number | null | undefined,
  from = true,
): string {
  if (!hasDisplayablePrice(price)) return MANAGER_PRICE_LABEL
  return formatCardFromPrice(price!, from)
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString('ru-RU')
}

export function truncateId(id: string): string {
  return id.substring(0, 8) + '...'
}
