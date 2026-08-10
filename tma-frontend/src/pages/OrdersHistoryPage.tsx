import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMyOrders } from '../hooks/useOrders'
import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { StatusBadge } from '../components/ui/StatusBadge'
import { Card } from '../components/ui/Card'
import { Button, Loader } from '../components/ui/Button'
import { Pagination } from '../components/ui/Pagination'
import { formatPrice, formatDate } from '../utils/format'
import {
  getActiveOrdersFilter,
  getCancelledOrdersFilter,
  getCompletedOrdersFilter,
  getOrderProgressPercent,
  isOrderActive,
  isOrderCancelled,
  isOrderCompleted,
} from '../utils/orderStatus'
import { Order, OrderStatus } from '../types/order'

const tabs: { label: string; value: string }[] = [
  { label: 'Все', value: '' },
  { label: 'Активные', value: getActiveOrdersFilter() },
  { label: 'Завершённые', value: getCompletedOrdersFilter() },
  { label: 'Отменённые', value: getCancelledOrdersFilter() },
]

function OrderTicket({ order }: { order: Order }) {
  const nav = useNavigate()
  const active = isOrderActive(order.status)
  const completed = isOrderCompleted(order.status)
  const cancelled = isOrderCancelled(order.status)
  const progressPercent = getOrderProgressPercent(order.status)

  const borderColor = completed
    ? 'var(--tg-alert-success-border)'
    : cancelled
      ? 'var(--tg-alert-error-border)'
      : active
        ? 'var(--tg-button)'
        : 'var(--tg-border)'

  const accentColor = completed ? '#4ade80' : cancelled ? '#f44336' : active ? 'var(--tg-button)' : 'var(--tg-hint)'

  return (
    <Card
      className="cursor-pointer transition-all duration-200 hover:shadow-lg"
      style={{ borderLeft: `3px solid ${accentColor}`, borderColor }}
      onClick={() => nav(`/order/${order.id}`)}
    >
      <div className="mb-3 flex items-start justify-between">
        <div className="flex-1">
          <div className="mb-1 flex items-center gap-2">
            <span className="font-mono text-[10px] text-[var(--tg-hint)]">#{order.id.substring(0, 8)}</span>
            <StatusBadge status={order.status} />
          </div>
          <div className="text-sm font-semibold">{order.product?.title || 'Товар'}</div>
        </div>
        {order.payment_amount ? (
          <div className="ml-3 text-right text-lg font-bold text-[var(--tg-button)]">
            {formatPrice(order.payment_amount)}
          </div>
        ) : null}
      </div>

      <div className="mb-3 flex flex-wrap items-center gap-2 text-xs text-[var(--tg-hint)]">
        <span>💬 Через менеджера</span>
        <span>•</span>
        <span>{formatDate(order.created_at)}</span>
        {order.quantity > 1 ? (
          <>
            <span>•</span>
            <span>{order.quantity} шт</span>
          </>
        ) : null}
      </div>

      {active ? (
        <div>
          <div className="h-1.5 overflow-hidden rounded-full bg-[var(--tg-secondary)]">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{
                width: `${progressPercent}%`,
                background: `linear-gradient(90deg, var(--tg-button), ${accentColor})`,
              }}
            />
          </div>
          <div className="mt-1 flex justify-between text-[10px] text-[var(--tg-hint)]">
            <span>Заявка</span>
            <span>Готово</span>
          </div>
        </div>
      ) : null}

      {completed ? (
        <div className="mt-2 flex items-center gap-1 text-xs text-green-400">
          <span>✓</span>
          <span>Завершён</span>
        </div>
      ) : null}

      {cancelled ? (
        <div className="mt-2 text-xs text-red-400">Отменён / возврат</div>
      ) : null}
    </Card>
  )
}

export function OrdersHistoryPage() {
  const nav = useNavigate()
  const goBack = useGoBack('/profile')
  const [tab, setTab] = useState('')
  const [page, setPage] = useState(1)
  const { data, isLoading } = useMyOrders(tab, page)

  return (
    <div className="pb-page">
      <Header title="Мои заказы" onBack={goBack} />
      <div className="p-4">
        <div className="mb-4 rounded-xl border border-amber-500/20 bg-amber-500/10 p-3 text-xs text-white/75">
          Все заказы оформляются через менеджера в Telegram. Цены уточняйте перед оплатой — они могут меняться из‑за курса.
        </div>

        <div className="mb-4 flex gap-1 rounded-lg bg-[var(--tg-secondary)] p-1">
          {tabs.map((t) => (
            <button
              key={t.value || 'all'}
              onClick={() => {
                setTab(t.value)
                setPage(1)
              }}
              className={`flex-1 rounded-md py-2 text-sm font-medium transition ${
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
          <div className="py-8 text-center text-[var(--tg-hint)]">
            <div className="mb-2 text-3xl">📋</div>
            <p className="mb-4">У вас пока нет заказов</p>
            <p className="mb-4 text-sm">
              Выберите товар в каталоге и нажмите «Оформить с менеджером» — заявка появится здесь.
            </p>
            <Button onClick={() => nav('/catalog')}>Перейти в каталог</Button>
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
