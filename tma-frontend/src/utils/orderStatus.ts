import { OrderStatus } from '../types/order'

export type OrderUiPhase = 'created' | 'manager' | 'processing' | 'done' | 'cancelled' | 'refund'

const ACTIVE_STATUSES: OrderStatus[] = [
  'NEW',
  'WAITING_PAYMENT',
  'PAYMENT_VERIFICATION',
  'PAID',
  'WAITING_ACTIVATION',
  'AWAITING_CREDENTIALS',
  'CREDENTIALS_RECEIVED',
  'CREDENTIALS_INVALID',
  'AWAITING_2FA',
  'INVALID_2FA',
  'ACTIVATING',
]

const COMPLETED_STATUSES: OrderStatus[] = ['COMPLETED', 'KEY_ISSUED', 'ACTIVATED']

const CANCELLED_STATUSES: OrderStatus[] = ['CANCELLED']

const REFUND_STATUSES: OrderStatus[] = ['REFUNDED', 'REFUND_REQUESTED']

export function getOrderUiPhase(status: OrderStatus): OrderUiPhase {
  if (CANCELLED_STATUSES.includes(status)) return 'cancelled'
  if (REFUND_STATUSES.includes(status)) return 'refund'
  if (COMPLETED_STATUSES.includes(status)) return 'done'
  if (status === 'NEW' || status === 'WAITING_PAYMENT') return 'created'
  if (status === 'PAYMENT_VERIFICATION') return 'manager'
  return 'processing'
}

export const orderUiPhaseLabels: Record<OrderUiPhase, string> = {
  created: 'Заявка создана',
  manager: 'Согласование',
  processing: 'Оформление',
  done: 'Завершён',
  cancelled: 'Отменён',
  refund: 'Возврат',
}

export const orderUiPhaseDescriptions: Record<OrderUiPhase, string> = {
  created: 'Заявка принята. Напишите менеджеру в Telegram — он подтвердит заказ и уточнит цену.',
  manager: 'Менеджер согласовывает детали, регион и итоговую стоимость. Цена может отличаться из‑за курса валют.',
  processing: 'Менеджер оформляет покупку на ваш аккаунт. Все вопросы — в чате Telegram.',
  done: 'Заказ выполнен. Спасибо за покупку!',
  cancelled: 'Заказ отменён. Если нужна помощь — напишите в поддержку.',
  refund: 'Запрос на возврат обрабатывается менеджером.',
}

export function getDisplayStatusLabel(status: OrderStatus): string {
  return orderUiPhaseLabels[getOrderUiPhase(status)]
}

export function isOrderActive(status: OrderStatus): boolean {
  return ACTIVE_STATUSES.includes(status)
}

export function isOrderCompleted(status: OrderStatus): boolean {
  return COMPLETED_STATUSES.includes(status)
}

export function isOrderCancelled(status: OrderStatus): boolean {
  return CANCELLED_STATUSES.includes(status) || REFUND_STATUSES.includes(status)
}

export function getOrderProgressPercent(status: OrderStatus): number {
  const phase = getOrderUiPhase(status)
  switch (phase) {
    case 'created':
      return 25
    case 'manager':
      return 50
    case 'processing':
      return 75
    case 'done':
      return 100
    default:
      return 0
  }
}

export const ORDER_PROGRESS_STEPS = [
  { id: 'created', label: 'Заявка' },
  { id: 'manager', label: 'Менеджер' },
  { id: 'processing', label: 'Оформление' },
  { id: 'done', label: 'Готово' },
] as const

export function getActiveOrdersFilter(): string {
  return ACTIVE_STATUSES.join(',')
}

export function getCompletedOrdersFilter(): string {
  return COMPLETED_STATUSES.join(',')
}

export function getCancelledOrdersFilter(): string {
  return [...CANCELLED_STATUSES, ...REFUND_STATUSES].join(',')
}
