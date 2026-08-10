import { useParams, useNavigate } from 'react-router-dom'
import { useOrder } from '../hooks/useOrders'
import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/StatusBadge'
import { OrderProgressCard } from '../components/order/OrderProgressCard'
import { OrderChat } from '../components/order/OrderChat'
import { formatPrice } from '../utils/format'
import { openManagerOrderChat } from '../utils/managerChat'
import { useShopSettings } from '../hooks/useContent'
import { isOrderCancelled, isOrderCompleted } from '../utils/orderStatus'
import { OrderStatus } from '../types/order'

export function OrderStatusPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const goBack = useGoBack('/orders')
  const { data: order, isLoading } = useOrder(id!)
  const { data: shopSettings } = useShopSettings()

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!order) return <div className="p-4 text-center text-[var(--tg-hint)]">Заказ не найден</div>

  const completed = isOrderCompleted(order.status as OrderStatus)
  const cancelled = isOrderCancelled(order.status as OrderStatus)

  return (
    <div className="pb-page">
      <Header title="Заказ" onBack={goBack} />
      <div className="space-y-4 p-4">
        <div className="rounded-xl bg-[var(--tg-secondary)] p-4">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-xs text-[var(--tg-hint)]">#{order.id.substring(0, 8)}</span>
            <StatusBadge status={order.status as OrderStatus} />
          </div>
          <div className="mb-1 text-base font-semibold">{order.product?.title || 'Товар'}</div>
          <div className="flex items-center justify-between">
            <div className="text-xs text-[var(--tg-hint)]">
              💬 Через менеджера
              {order.quantity > 1 ? ` • ${order.quantity} шт` : ''}
            </div>
            {order.payment_amount ? (
              <span className="text-lg font-bold text-[var(--tg-button)]">
                {formatPrice(order.payment_amount)}
              </span>
            ) : null}
          </div>
        </div>

        {!cancelled ? (
          <Button fullWidth size="lg" onClick={() => openManagerOrderChat(order.id)}>
            💬 Написать менеджеру в Telegram
          </Button>
        ) : null}

        <div className="rounded-xl border border-amber-500/20 bg-amber-500/10 p-4 text-sm text-white/75">
          Цены в каталоге ориентировочные и могут меняться из‑за курса валют.
          Итоговую стоимость всегда уточняйте у менеджера перед оплатой.
        </div>

        <OrderProgressCard status={order.status} />

        {!cancelled && !completed ? <OrderChat orderId={order.id} /> : null}

        {completed ? (
          <div className="space-y-4 rounded-xl border border-[var(--tg-alert-success-border)] bg-[var(--tg-alert-success-bg)] p-6 text-center">
            <div className="text-4xl">🎉</div>
            <h2 className="text-lg font-bold text-[var(--tg-alert-success-text)]">Заказ завершён</h2>
            <p className="text-sm text-[var(--tg-alert-success-text)]/80">
              Спасибо за покупку! Если остались вопросы — напишите менеджеру.
            </p>
            <div className="space-y-2 text-left">
              <a
                href={shopSettings?.reviews_url || 'https://t.me/coin_mint_reviews'}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[var(--tg-link)]"
              >
                <span>📝</span> Оставить отзыв
              </a>
              <a
                href={shopSettings?.support_url || 'https://t.me/coin_mint_chat'}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[var(--tg-link)]"
              >
                <span>💬</span> Поддержка
              </a>
            </div>
            <Button onClick={() => nav('/')}>🛒 В магазин</Button>
          </div>
        ) : null}

        {cancelled ? (
          <div className="rounded-xl border border-[var(--tg-alert-error-border)] bg-[var(--tg-alert-error-bg)] p-6 text-center">
            <div className="mb-2 text-4xl">❌</div>
            <h2 className="mb-2 text-lg font-bold text-[var(--tg-alert-error-text)]">Заказ отменён</h2>
            {order.cancelled_reason ? (
              <p className="mb-4 text-sm text-[var(--tg-alert-error-text)]/80">
                Причина: {order.cancelled_reason}
              </p>
            ) : null}
            <Button variant="outline" onClick={() => nav('/support')}>
              💬 Связаться с поддержкой
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  )
}
