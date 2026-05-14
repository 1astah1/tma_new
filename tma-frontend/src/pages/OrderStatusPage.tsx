import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useOrder, useSendCredentials, useSend2FACode } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/StatusBadge'
import { statusLabels, statusColors } from '../types/order'

const statusSteps = [
  'NEW', 'WAITING_PAYMENT', 'PAYMENT_VERIFICATION', 'PAID',
  'WAITING_ACTIVATION', 'AWAITING_CREDENTIALS', 'CREDENTIALS_RECEIVED',
  'AWAITING_2FA', 'ACTIVATING', 'ACTIVATED', 'COMPLETED',
]

export function OrderStatusPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const { data: order, isLoading } = useOrder(id!)
  const sendCreds = useSendCredentials()
  const sendCode = useSend2FACode()

  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!order) return <div className="p-4 text-center text-tg-hint">Заказ не найден</div>

  const currentIdx = statusSteps.indexOf(order.status)
  const completed = order.status === 'COMPLETED' || order.status === 'KEY_ISSUED'
  const cancelled = order.status === 'CANCELLED' || order.status === 'REFUNDED'

  const handleSendCredentials = async () => {
    if (!login || !password) return
    await sendCreds.mutateAsync({ orderId: order.id, platform: 'xbox', login, password })
  }

  const handleSendCode = async () => {
    if (!code) return
    await sendCode.mutateAsync({ orderId: order.id, code })
  }

  return (
    <div className="pb-24">
      <Header title="Статус заказа" onBack={() => nav(-1)} />
      <div className="p-4 space-y-4">
        <div className="bg-tg-secondary rounded-xl p-4">
          <div className="flex justify-between items-center mb-2">
            <span className="text-tg-hint">Заказ #{order.id.substring(0, 8)}</span>
            <StatusBadge status={order.status} />
          </div>
          <div className="font-medium">{order.product?.title || 'Товар'}</div>
        </div>

        {!completed && !cancelled && (
          <div className="space-y-2">
            {statusSteps.slice(0, Math.max(currentIdx + 1, 4)).map((s, i) => {
              const done = i <= currentIdx
              const current = i === currentIdx
              const stepStatus = s as any
              return (
                <div key={s} className="flex items-center gap-3">
                  <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ${
                    done ? 'bg-green-500 text-white' : 'bg-gray-200 text-gray-400'
                  }`}>
                    {done ? '✓' : i + 1}
                  </div>
                  <span className={current ? 'font-medium' : done ? 'text-gray-600' : 'text-gray-400'}>
                    {statusLabels[stepStatus] || s}
                  </span>
                </div>
              )
            })}
          </div>
        )}

        {order.status === 'WAITING_PAYMENT' && (
          <Button fullWidth onClick={() => nav(`/checkout/${order.id}`)}>
            💳 Перейти к оплате
          </Button>
        )}

        {order.status === 'AWAITING_CREDENTIALS' && (
          <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 space-y-3">
            <p className="text-sm text-blue-800">
              🔐 Для активации товара введите данные вашего аккаунта. Все данные шифруются.
            </p>
            <input
              type="text" placeholder="Логин (email)"
              value={login} onChange={(e) => setLogin(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-gray-200"
            />
            <input
              type="password" placeholder="Пароль"
              value={password} onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-gray-200"
            />
            <Button fullWidth onClick={handleSendCredentials} loading={sendCreds.isPending}>
              📤 Отправить данные
            </Button>
          </div>
        )}

        {order.status === 'AWAITING_2FA' && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-4 space-y-3">
            <p className="text-sm text-red-800">
              🔐 Администратор готов войти в ваш аккаунт. Отправьте код подтверждения.
            </p>
            <input
              type="text" placeholder="Код подтверждения"
              value={code} onChange={(e) => setCode(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-gray-200"
            />
            <Button fullWidth onClick={handleSendCode} loading={sendCode.isPending}>
              🔑 Отправить код
            </Button>
          </div>
        )}

        {order.status === 'ACTIVATING' && (
          <div className="text-center py-4">
            <Loader />
            <p className="text-tg-hint mt-2">Администратор активирует товар...</p>
          </div>
        )}

        {completed && (
          <div className="bg-green-50 border border-green-200 rounded-xl p-6 text-center">
            <div className="text-4xl mb-2">🎉</div>
            <h2 className="text-lg font-bold text-green-800 mb-2">Заказ завершён!</h2>
            <p className="text-sm text-green-700 mb-4">
              {order.delivery_method === 'key' ? 'Ваш ключ активации готов.' : 'Товар активирован на вашем аккаунте.'}
            </p>
            <Button onClick={() => nav('/')}>🛒 Вернуться в магазин</Button>
          </div>
        )}

        {cancelled && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-4 text-center">
            <div className="text-4xl mb-2">❌</div>
            <h2 className="text-lg font-bold text-red-800 mb-2">Заказ отменён</h2>
            {order.cancelled_reason && (
              <p className="text-sm text-red-700">Причина: {order.cancelled_reason}</p>
            )}
            <Button className="mt-4" variant="outline" onClick={() => nav('/')}>
              🛒 В магазин
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
