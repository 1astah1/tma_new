import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMyOrders } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { StatusBadge } from '../components/ui/StatusBadge'
import { Card } from '../components/ui/Card'
import { Loader } from '../components/ui/Button'
import { formatPrice, formatDate } from '../utils/format'
import { OrderStatus } from '../types/order'

const tabs: { label: string; value: string }[] = [
  { label: 'Все', value: '' },
  { label: 'Активные', value: 'WAITING_PAYMENT,PAYMENT_VERIFICATION,PAID,WAITING_ACTIVATION,AWAITING_CREDENTIALS,CREDENTIALS_RECEIVED,AWAITING_2FA,ACTIVATING' },
  { label: 'Завершённые', value: 'COMPLETED,KEY_ISSUED,ACTIVATED' },
]

export function OrdersHistoryPage() {
  const nav = useNavigate()
  const [tab, setTab] = useState('')
  const { data, isLoading } = useMyOrders(tab)

  return (
    <div className="pb-24">
      <Header title="Мои заказы" />
      <div className="p-4">
        <div className="flex gap-1 mb-4 bg-tg-secondary rounded-lg p-1">
          {tabs.map((t) => (
            <button
              key={t.value}
              onClick={() => setTab(t.value)}
              className={`flex-1 py-2 rounded-md text-sm font-medium transition ${
                tab === t.value ? 'bg-white shadow-sm' : 'text-tg-hint'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : data?.data?.length === 0 ? (
          <div className="text-center py-8 text-tg-hint">
            <div className="text-3xl mb-2">📋</div>
            <p>У вас пока нет заказов</p>
          </div>
        ) : (
          <div className="space-y-3">
            {data?.data?.map((order) => (
              <Card key={order.id} onClick={() => nav(`/order/${order.id}`)}>
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <div className="text-xs text-tg-hint">#{order.id.substring(0, 8)}</div>
                    <div className="font-medium text-sm">{order.product?.title || 'Товар'}</div>
                  </div>
                  <StatusBadge status={order.status as OrderStatus} />
                </div>
                <div className="flex justify-between items-center">
                  <div className="text-xs text-tg-hint">
                    {order.delivery_method === 'key' ? '🔑 Ключ' : '🔐 Активация'}
                    {' • '}{formatDate(order.created_at)}
                  </div>
                  <span className="font-bold text-tg-button text-sm">
                    {order.payment_amount ? formatPrice(order.payment_amount) : ''}
                  </span>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
