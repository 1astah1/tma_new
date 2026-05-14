import { Product } from './product'

export type OrderStatus =
  | 'NEW' | 'WAITING_PAYMENT' | 'PAYMENT_VERIFICATION' | 'PAID'
  | 'WAITING_ACTIVATION' | 'AWAITING_CREDENTIALS' | 'CREDENTIALS_RECEIVED'
  | 'AWAITING_2FA' | 'ACTIVATING' | 'ACTIVATED' | 'KEY_ISSUED' | 'COMPLETED'
  | 'CANCELLED' | 'REFUND_REQUESTED' | 'REFUNDED'

export interface Order {
  id: string
  user_id: string
  product_id: string
  delivery_method: 'key' | 'activation'
  status: OrderStatus
  payment_method: string | null
  payment_amount: number | null
  payment_receipt_url: string | null
  payment_verified_by: string | null
  key_id: string | null
  assigned_admin_id: string | null
  cancelled_reason: string | null
  created_at: string
  updated_at: string
  product?: Product
  history?: OrderHistory[]
}

export interface OrderHistory {
  id: string
  order_id: string
  old_status: OrderStatus | null
  new_status: OrderStatus
  changed_by_id: string | null
  changed_by_type: 'admin' | 'system' | 'user'
  comment: string | null
  created_at: string
}

export const statusLabels: Record<OrderStatus, string> = {
  NEW: 'Новый',
  WAITING_PAYMENT: 'Ожидает оплаты',
  PAYMENT_VERIFICATION: 'Проверка платежа',
  PAID: 'Оплачен',
  WAITING_ACTIVATION: 'В очереди на активацию',
  AWAITING_CREDENTIALS: 'Требуются данные',
  CREDENTIALS_RECEIVED: 'Данные получены',
  AWAITING_2FA: 'Ожидает код',
  ACTIVATING: 'Активация...',
  ACTIVATED: 'Активирован',
  KEY_ISSUED: 'Ключ выдан',
  COMPLETED: 'Завершён',
  CANCELLED: 'Отменён',
  REFUND_REQUESTED: 'Возврат запрошен',
  REFUNDED: 'Возвращён',
}

export const statusColors: Record<OrderStatus, string> = {
  NEW: 'bg-gray-100 text-gray-800',
  WAITING_PAYMENT: 'bg-blue-100 text-blue-800',
  PAYMENT_VERIFICATION: 'bg-orange-100 text-orange-800',
  PAID: 'bg-green-100 text-green-800',
  WAITING_ACTIVATION: 'bg-purple-100 text-purple-800',
  AWAITING_CREDENTIALS: 'bg-blue-100 text-blue-800',
  CREDENTIALS_RECEIVED: 'bg-yellow-100 text-yellow-800',
  AWAITING_2FA: 'bg-red-100 text-red-800',
  ACTIVATING: 'bg-orange-100 text-orange-800',
  ACTIVATED: 'bg-green-100 text-green-800',
  KEY_ISSUED: 'bg-green-100 text-green-800',
  COMPLETED: 'bg-green-100 text-green-800',
  CANCELLED: 'bg-gray-100 text-gray-500',
  REFUND_REQUESTED: 'bg-red-100 text-red-800',
  REFUNDED: 'bg-gray-100 text-gray-500',
}
