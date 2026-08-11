import { BOT_USERNAME } from '../config/api'

/** Куда уходят заявки, если в настройках магазина ничего не задано. */
const DEFAULT_MANAGER = 'https://t.me/KromkaJQ'

/** Из настроек может прийти «@user», «user» или полная ссылка — приводим к t.me/user. */
function managerLink(managerUrl?: string | null): string {
  const raw = (managerUrl || DEFAULT_MANAGER).trim()
  if (!raw) return DEFAULT_MANAGER

  if (raw.startsWith('http://') || raw.startsWith('https://')) {
    return raw.split('?')[0].replace(/\/+$/, '')
  }
  return `https://t.me/${raw.replace(/^@/, '')}`
}

function openLink(url: string) {
  const tg = window.Telegram?.WebApp
  if (tg?.openTelegramLink) {
    tg.openTelegramLink(url)
    return
  }
  window.open(url, '_blank', 'noopener,noreferrer')
}

export type PurchaseRequestLine = {
  label: string
  value: string
}

export type PurchaseRequest = {
  title: string
  lines: PurchaseRequestLine[]
  imageUrl?: string | null
  total?: string
}

/**
 * Текст заявки, который подставляется пользователю в поле ввода. Картинку
 * вложением подставить нельзя — Telegram такого не умеет, поэтому даём ссылку
 * на обложку: клиент сам покажет превью.
 */
export function buildPurchaseRequestText(request: PurchaseRequest): string {
  const parts: string[] = ['🛒 ЗАЯВКА НА ПОКУПКУ', '', `📌 Товар: ${request.title}`, '']

  for (const line of request.lines) {
    if (line.value) parts.push(`${line.label}: ${line.value}`)
  }

  if (request.total) {
    parts.push('', `💰 Стоимость: ${request.total}`)
  }
  if (request.imageUrl) {
    parts.push('', `🖼 ${request.imageUrl}`)
  }

  parts.push('', '💬 Прошу помочь с оформлением. Жду подтверждения.')
  return parts.join('\n')
}

/** Открывает лс менеджера с уже набранным текстом заявки. */
export function openManagerRequest(managerUrl: string | null | undefined, text: string) {
  openLink(`${managerLink(managerUrl)}?text=${encodeURIComponent(text)}`)
}

/** Чат с ботом по конкретному заказу — статус, отмена, история. */
export function openBotOrderChat(orderId: string) {
  openLink(`https://t.me/${BOT_USERNAME}?start=order_${orderId}`)
}

export function openBotCartChat(orderIds: string[]) {
  if (orderIds.length === 0) return
  const payload = orderIds.length === 1 ? `order_${orderIds[0]}` : `cart_${orderIds.join(',')}`
  openLink(`https://t.me/${BOT_USERNAME}?start=${payload}`)
}
