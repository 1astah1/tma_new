import { OrderStatus } from '../../types/order'
import { getDisplayStatusLabel } from '../../utils/orderStatus'

const statusColors: Partial<Record<OrderStatus, string>> = {
  NEW: 'bg-gray-500/20 text-gray-300',
  WAITING_PAYMENT: 'bg-blue-500/20 text-blue-300',
  PAYMENT_VERIFICATION: 'bg-orange-500/20 text-orange-300',
  PAID: 'bg-green-500/20 text-green-300',
  WAITING_ACTIVATION: 'bg-amber-500/20 text-amber-300',
  AWAITING_CREDENTIALS: 'bg-amber-500/20 text-amber-300',
  CREDENTIALS_RECEIVED: 'bg-amber-500/20 text-amber-300',
  CREDENTIALS_INVALID: 'bg-amber-500/20 text-amber-300',
  AWAITING_2FA: 'bg-amber-500/20 text-amber-300',
  INVALID_2FA: 'bg-amber-500/20 text-amber-300',
  ACTIVATING: 'bg-amber-500/20 text-amber-300',
  ACTIVATED: 'bg-green-500/20 text-green-300',
  KEY_ISSUED: 'bg-green-500/20 text-green-300',
  COMPLETED: 'bg-green-500/20 text-green-300',
  CANCELLED: 'bg-gray-500/20 text-gray-400',
  REFUND_REQUESTED: 'bg-red-500/20 text-red-300',
  REFUNDED: 'bg-gray-500/20 text-gray-400',
}

export function StatusBadge({ status }: { status: OrderStatus }) {
  const label = getDisplayStatusLabel(status)
  const color = statusColors[status] || 'bg-gray-500/20 text-gray-300'
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${color}`}>
      {label}
    </span>
  )
}
