import { BOT_USERNAME } from '../config/api'

export function openManagerOrderChat(orderId: string) {
  window.Telegram?.WebApp?.openTelegramLink(`https://t.me/${BOT_USERNAME}?start=order_${orderId}`)
}

export function openManagerCartChat(orderIds: string[]) {
  if (orderIds.length === 0) return
  const payload = orderIds.length === 1 ? `order_${orderIds[0]}` : `cart_${orderIds.join(',')}`
  window.Telegram?.WebApp?.openTelegramLink(`https://t.me/${BOT_USERNAME}?start=${payload}`)
}
