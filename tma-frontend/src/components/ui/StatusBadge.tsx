import { OrderStatus, statusLabels } from '../../types/order'

const statusColors: Record<OrderStatus, string> = {
  NEW: 'bg-gray-500/20 text-gray-300',
  WAITING_PAYMENT: 'bg-blue-500/20 text-blue-300',
  PAYMENT_VERIFICATION: 'bg-orange-500/20 text-orange-300',
  PAID: 'bg-green-500/20 text-green-300',
  WAITING_ACTIVATION: 'bg-purple-500/20 text-purple-300',
  AWAITING_CREDENTIALS: 'bg-blue-500/20 text-blue-300',
  CREDENTIALS_RECEIVED: 'bg-yellow-500/20 text-yellow-300',
  CREDENTIALS_INVALID: 'bg-red-500/20 text-red-300',
  AWAITING_2FA: 'bg-red-500/20 text-red-300',
  INVALID_2FA: 'bg-red-500/20 text-red-300',
  ACTIVATING: 'bg-orange-500/20 text-orange-300',
  ACTIVATED: 'bg-green-500/20 text-green-300',
  KEY_ISSUED: 'bg-green-500/20 text-green-300',
  COMPLETED: 'bg-green-500/20 text-green-300',
  CANCELLED: 'bg-gray-500/20 text-gray-400',
  REFUND_REQUESTED: 'bg-red-500/20 text-red-300',
  REFUNDED: 'bg-gray-500/20 text-gray-400',
}

export function StatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[status] || 'bg-gray-500/20 text-gray-300'}`}>
      {statusLabels[status] || status}
    </span>
  )
}
