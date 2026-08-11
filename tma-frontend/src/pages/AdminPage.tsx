import { useMemo, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Header } from '../components/layout/Header'
import { Button } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/StatusBadge'
import { useToast } from '../components/ui/Toast'
import { useGoBack } from '../hooks/useGoBack'
import { formatPrice } from '../utils/format'
import { Order, OrderStatus, statusLabels } from '../types/order'
import {
  getAdminOrders,
  getAdminStats,
  updateAdminOrderStatus,
  getAdminOrderChat,
  sendAdminOrderMessage,
} from '../services/tmaAdmin.service'

/** Статусы, которые менеджер ставит руками чаще всего. */
const QUICK_STATUSES: OrderStatus[] = [
  'PAYMENT_VERIFICATION',
  'PAID',
  'WAITING_ACTIVATION',
  'ACTIVATED',
  'COMPLETED',
  'CANCELLED',
]

const FILTERS: { id: string; label: string }[] = [
  { id: '', label: 'Все' },
  { id: 'WAITING_PAYMENT', label: 'Новые' },
  { id: 'PAID', label: 'Оплачены' },
  { id: 'COMPLETED', label: 'Готовы' },
]

export function AdminPage() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const [filter, setFilter] = useState('')
  const [openOrder, setOpenOrder] = useState<Order | null>(null)
  const [message, setMessage] = useState('')

  const { data: stats } = useQuery({ queryKey: ['tma-admin-stats'], queryFn: getAdminStats })
  const { data: orders = [], isLoading } = useQuery({
    queryKey: ['tma-admin-orders', filter],
    queryFn: () => getAdminOrders(filter || undefined),
    refetchInterval: 30000,
  })

  const { data: chat = [] } = useQuery({
    queryKey: ['tma-admin-chat', openOrder?.id],
    queryFn: () => getAdminOrderChat(openOrder!.id),
    enabled: !!openOrder,
    refetchInterval: 15000,
  })

  const setStatus = useMutation({
    mutationFn: ({ id, status }: { id: string; status: OrderStatus }) =>
      updateAdminOrderStatus(id, status),
    onSuccess: () => {
      toast.toast('Статус обновлён', 'success')
      queryClient.invalidateQueries({ queryKey: ['tma-admin-orders'] })
      setOpenOrder(null)
    },
    onError: () => toast.toast('Не удалось обновить статус', 'error'),
  })

  const sendMessage = useMutation({
    mutationFn: (text: string) => sendAdminOrderMessage(openOrder!.id, text),
    onSuccess: () => {
      setMessage('')
      queryClient.invalidateQueries({ queryKey: ['tma-admin-chat', openOrder?.id] })
    },
    onError: () => toast.toast('Сообщение не ушло', 'error'),
  })

  const totals = useMemo(
    () => [
      { label: 'Заказов сегодня', value: stats?.orders_today ?? 0 },
      { label: 'Всего заказов', value: stats?.orders_total ?? orders.length },
      { label: 'Ждут действия', value: stats?.pending_orders ?? 0 },
    ],
    [stats, orders.length],
  )

  if (openOrder) {
    return (
      <div className="pb-page">
        <Header title={`Заказ #${openOrder.id.slice(0, 8)}`} onBack={() => setOpenOrder(null)} />
        <div className="space-y-4 p-4">
          <div className="rounded-2xl border border-white/10 bg-[#141414] p-4">
            <div className="mb-2 flex items-center justify-between">
              <span className="font-semibold">{openOrder.product?.title ?? 'Товар'}</span>
              <StatusBadge status={openOrder.status} />
            </div>
            <div className="text-sm text-white/60">
              {openOrder.payment_amount ? formatPrice(openOrder.payment_amount) : 'без суммы'}
            </div>
          </div>

          <div>
            <div className="mb-2 text-sm text-white/60">Сменить статус</div>
            <div className="flex flex-wrap gap-2">
              {QUICK_STATUSES.map((status) => (
                <button
                  key={status}
                  onClick={() => setStatus.mutate({ id: openOrder.id, status })}
                  disabled={setStatus.isPending || status === openOrder.status}
                  className="rounded-xl border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs font-semibold disabled:opacity-40"
                >
                  {statusLabels[status]}
                </button>
              ))}
            </div>
          </div>

          <div>
            <div className="mb-2 text-sm text-white/60">Переписка с покупателем</div>
            <div className="max-h-72 space-y-2 overflow-y-auto rounded-2xl border border-white/10 bg-[#141414] p-3">
              {chat.length === 0 ? (
                <div className="py-6 text-center text-sm text-white/40">Сообщений пока нет</div>
              ) : (
                chat.map((m) => (
                  <div
                    key={m.id}
                    className={`max-w-[85%] rounded-xl px-3 py-2 text-sm ${
                      m.sender_type === 'admin'
                        ? 'ml-auto bg-amber-500/20'
                        : 'bg-white/10'
                    }`}
                  >
                    {m.message}
                  </div>
                ))
              )}
            </div>
            <div className="mt-2 flex gap-2">
              <input
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="Ответить покупателю"
                className="flex-1 rounded-xl border border-white/10 bg-[#141414] px-3 py-2 text-sm outline-none"
              />
              <Button
                size="md"
                onClick={() => message.trim() && sendMessage.mutate(message.trim())}
                loading={sendMessage.isPending}
              >
                Отправить
              </Button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="pb-page">
      <Header title="Управление заказами" onBack={goBack} />
      <div className="space-y-4 p-4">
        <div className="grid grid-cols-3 gap-2">
          {totals.map((item) => (
            <div key={item.label} className="rounded-2xl border border-white/10 bg-[#141414] p-3 text-center">
              <div className="text-lg font-black text-amber-200">{item.value}</div>
              <div className="text-[11px] leading-tight text-white/50">{item.label}</div>
            </div>
          ))}
        </div>

        <div className="flex gap-2 overflow-x-auto">
          {FILTERS.map((item) => (
            <button
              key={item.id}
              onClick={() => setFilter(item.id)}
              className={`whitespace-nowrap rounded-xl px-3 py-2 text-xs font-semibold ${
                filter === item.id ? 'bg-amber-500 text-black' : 'border border-white/10 bg-[#141414]'
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>

        {isLoading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : orders.length === 0 ? (
          <div className="py-10 text-center text-white/50">Заказов нет</div>
        ) : (
          <div className="space-y-2">
            {orders.map((order) => (
              <button
                key={order.id}
                onClick={() => setOpenOrder(order)}
                className="w-full rounded-2xl border border-white/10 bg-[#141414] p-3 text-left"
              >
                <div className="mb-1 flex items-center justify-between gap-2">
                  <span className="truncate text-sm font-semibold">
                    {order.product?.title ?? `Заказ #${order.id.slice(0, 8)}`}
                  </span>
                  <StatusBadge status={order.status} />
                </div>
                <div className="flex items-center justify-between text-xs text-white/50">
                  <span>#{order.id.slice(0, 8)}</span>
                  <span>{order.payment_amount ? formatPrice(order.payment_amount) : '—'}</span>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
