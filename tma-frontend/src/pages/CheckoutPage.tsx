import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useOrder, useConfirmPayment } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/StatusBadge'
import { formatPrice } from '../utils/format'
import api from '../services/api'

const paymentMethods = [
  { id: 'sbp', label: 'СБП', icon: '🏦' },
  { id: 'card', label: 'Картой', icon: '💳' },
  { id: 'crypto', label: 'Криптовалюта', icon: '₿' },
]

export function CheckoutPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const { data: order, isLoading } = useOrder(id!)
  const confirmPayment = useConfirmPayment()
  const [agreed, setAgreed] = useState(false)
  const [paymentMethod, setPaymentMethod] = useState('sbp')

  const { data: paymentDetails } = useQuery({
    queryKey: ['paymentDetails'],
    queryFn: async () => {
      const { data } = await api.get('/payments/details')
      return data
    },
  })

  const details = paymentDetails?.[paymentMethod] || null

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!order) return <div className="p-4 text-center text-tg-hint">Заказ не найден</div>

  if (order.status !== 'WAITING_PAYMENT') {
    return (
      <div className="p-4">
        <Header title="Статус заказа" onBack={() => nav(-1)} />
        <div className="text-center py-8">
          <StatusBadge status={order.status} />
          <Button className="mt-4" onClick={() => nav(`/order/${order.id}`)}>Перейти к заказу</Button>
        </div>
      </div>
    )
  }

  const handleConfirmPayment = async () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*,.pdf'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (file) {
        await confirmPayment.mutateAsync({ orderId: order.id, paymentMethod, file })
        nav(`/order/${order.id}`)
      }
    }
    input.click()
  }

  return (
    <div className="pb-24">
      <Header title="Оформление" onBack={() => nav(-1)} />
      <div className="p-4 space-y-4">
        <div className="bg-tg-secondary rounded-xl p-4">
          <div className="flex justify-between mb-2">
            <span className="text-tg-hint">Товар</span>
            <span className="font-medium">{order.product?.title || 'Загрузка...'}</span>
          </div>
          <div className="flex justify-between mb-2">
            <span className="text-tg-hint">Способ</span>
            <span>{order.delivery_method === 'key' ? '🔑 Ключ' : '🔐 Активация'}</span>
          </div>
          <div className="flex justify-between text-lg font-bold">
            <span>Сумма</span>
            <span className="text-tg-button">{formatPrice(order.product?.price || 0)}</span>
          </div>
        </div>

        <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4 text-sm">
          <div className="font-medium mb-2">📋 Правила покупки</div>
          <ul className="space-y-1 text-yellow-800">
            <li>• После оплаты товар выдаётся автоматически</li>
            <li>• Возврат невозможен после выдачи ключа/активации</li>
            <li>• Данные аккаунта шифруются и используются только для активации</li>
          </ul>
          <label className="flex items-center gap-2 mt-3 cursor-pointer">
            <input type="checkbox" checked={agreed} onChange={(e) => setAgreed(e.target.checked)} className="accent-tg-button" />
            <span className="text-sm">Я ознакомлен и согласен с правилами</span>
          </label>
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Способ оплаты</label>
          <div className="space-y-2">
            {paymentMethods.map((pm) => (
              <label key={pm.id} className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer ${paymentMethod === pm.id ? 'border-tg-button bg-tg-button/5' : 'border-gray-200'}`}
                onClick={() => setPaymentMethod(pm.id)}>
                <input type="radio" checked={paymentMethod === pm.id} readOnly className="accent-tg-button" />
                <span>{pm.icon} {pm.label}</span>
              </label>
            ))}
          </div>
        </div>

        {details && (
          <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 text-sm space-y-1">
            <div className="font-medium mb-1">💳 Реквизиты для оплаты:</div>
            {Object.entries(details).map(([k, v]) => (
              <div key={k} className="text-blue-800">
                <span className="font-medium">{k === 'phone' ? '📱 Номер' : k === 'bank' ? '🏦 Банк' : k === 'receiver' ? '👤 Получатель' : k === 'number' ? '💳 Карта' : k === 'binance' ? 'Binance' : k === 'bybit' ? 'Bybit' : k === 'trc20' ? 'TRC20' : k}:</span> {String(v)}
              </div>
            ))}
          </div>
        )}

        <div className="bg-gray-50 rounded-xl p-4 text-sm text-center text-tg-hint">
          После оплаты нажмите кнопку ниже и прикрепите скриншот/PDF чека
        </div>

        <Button fullWidth size="lg" onClick={handleConfirmPayment} disabled={!agreed} loading={confirmPayment.isPending}>
          💳 Я оплатил — приложить чек
        </Button>
      </div>
    </div>
  )
}
