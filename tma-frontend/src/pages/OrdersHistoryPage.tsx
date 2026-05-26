import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMyOrders } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { StatusBadge } from '../components/ui/StatusBadge'
import { Card } from '../components/ui/Card'
import { Loader } from '../components/ui/Button'
import { Pagination } from '../components/ui/Pagination'
import { formatPrice, formatDate } from '../utils/format'
import { OrderStatus } from '../types/order'

const tabs: { label: string; value: string }[] = [
  { label: 'Все', value: '' },
  { label: 'Активные', value: 'WAITING_PAYMENT,PAYMENT_VERIFICATION,PAID,WAITING_ACTIVATION,AWAITING_CREDENTIALS,CREDENTIALS_RECEIVED,AWAITING_2FA,ACTIVATING' },
  { label: 'Завершённые', value: 'COMPLETED,KEY_ISSUED,ACTIVATED' },
]

function OrderTicket({ order }: { order: any }) {
  const nav = useNavigate()
  const isActive = ['WAITING_PAYMENT', 'PAYMENT_VERIFICATION', 'PAID', 'WAITING_ACTIVATION', 'AWAITING_CREDENTIALS', 'CREDENTIALS_RECEIVED', 'AWAITING_2FA', 'ACTIVATING'].includes(order.status)
  const isCompleted = ['COMPLETED', 'KEY_ISSUED', 'ACTIVATED'].includes(order.status)
  const isCancelled = ['CANCELLED', 'REFUNDED'].includes(order.status)

  const borderColor = isCompleted ? 'var(--tg-alert-success-border)' : isCancelled ? 'var(--tg-alert-error-border)' : isActive ? 'var(--tg-button)' : 'var(--tg-border)'
  const accentColor = isCompleted ? '#4ade80' : isCancelled ? '#f44336' : isActive ? 'var(--tg-button)' : 'var(--tg-hint)'

  const progressSteps = ['WAITING_PAYMENT', 'PAYMENT_VERIFICATION', 'PAID', 'WAITING_ACTIVATION', 'AWAITING_CREDENTIALS', 'CREDENTIALS_RECEIVED', 'AWAITING_2FA', 'ACTIVATING', 'ACTIVATED', 'COMPLETED']
  const currentStep = progressSteps.indexOf(order.status)
  const totalSteps = order.delivery_method === 'key' ? 4 : 8
  const progressPercent = currentStep >= 0 ? Math.min((currentStep / totalSteps) * 100, 100) : 0

  return (
    <Card
      className="cursor-pointer hover:shadow-lg transition-all duration-200 group"
      style={{ borderLeft: `3px solid ${accentColor}`, borderColor }}
      onClick={() => nav(`/order/${order.id}`)}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-[10px] text-[var(--tg-hint)] font-mono">#{order.id.substring(0, 8)}</span>
            <StatusBadge status={order.status as OrderStatus} />
          </div>
          <div className="font-semibold text-sm">{order.product?.title || 'Товар'}</div>
        </div>
        <div className="text-right ml-3">
          <div className="font-bold text-[var(--tg-button)] text-lg">
            {order.payment_amount ? formatPrice(order.payment_amount) : ''}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-3 text-xs text-[var(--tg-hint)]">
          <span>{order.delivery_method === 'key' ? '🔑 Ключ' : '🔐 Активация'}</span>
          <span>•</span>
          <span>{formatDate(order.created_at)}</span>
          {order.quantity > 1 && (
            <>
              <span>•</span>
              <span>{order.quantity} шт</span>
            </>
          )}
        </div>
      </div>

      {isActive && (
        <div className="mt-2">
          <div className="w-full h-1.5 bg-[var(--tg-secondary)] rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{ width: `${progressPercent}%`, background: `linear-gradient(90deg, var(--tg-button), ${accentColor})` }}
            />
          </div>
          <div className="flex justify-between mt-1">
            <span className="text-[10px] text-[var(--tg-hint)]">Оплата</span>
            <span className="text-[10px] text-[var(--tg-hint)]">
              {order.delivery_method === 'key' ? 'Выдача' : 'Активация'}
            </span>
          </div>
        </div>
      )}

      {isCompleted && (
        <div className="mt-2 flex items-center gap-1 text-xs text-green-400">
          <span>✓</span>
          <span>Завершён</span>
        </div>
      )}
    </Card>
  )
}

export function OrdersHistoryPage() {
  const nav = useNavigate()
  const [tab, setTab] = useState('')
  const [page, setPage] = useState(1)
  const { data, isLoading } = useMyOrders(tab, page)

  return (
    <div className="pb-24">
      <Header title="Мои заказы" />
      <div className="p-4">
        <div className="flex gap-1 mb-4 bg-[var(--tg-secondary)] rounded-lg p-1">
          {tabs.map((t) => (
            <button
              key={t.value}
              onClick={() => { setTab(t.value); setPage(1) }}
              className={`flex-1 py-2 rounded-md text-sm font-medium transition ${
                tab === t.value ? 'bg-[var(--tg-button)] text-white shadow-sm' : 'text-[var(--tg-hint)]'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : data?.data?.length === 0 ? (
          <div className="text-center py-8 text-[var(--tg-hint)]">
            <div className="text-3xl mb-2">📋</div>
            <p>У вас пока нет заказов</p>
          </div>
        ) : (
          <>
            <div className="space-y-3">
              {data?.data?.map((order) => (
                <OrderTicket key={order.id} order={order} />
              ))}
            </div>
            <Pagination
              page={page}
              total={data?.meta?.total || 0}
              limit={20}
              onPageChange={setPage}
            />
          </>
        )}
      </div>
    </div>
  )
}
