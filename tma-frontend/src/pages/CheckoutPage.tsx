import { useState, useRef } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useOrder, useConfirmPayment, useCreateBatchOrder, useConfirmBatchPayment } from '../hooks/useOrders'
import { useCart } from '../stores/cartStore'
import { useToast } from '../components/ui/Toast'
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
  const [searchParams] = useSearchParams()
  const fromCart = searchParams.get('from') === 'cart'
  const nav = useNavigate()
  const toast = useToast()
  const { data: order, isLoading } = useOrder(id!)
  const confirmPayment = useConfirmPayment()
  const createBatch = useCreateBatchOrder()
  const confirmBatch = useConfirmBatchPayment()
  const { items, getTotal, promoDiscount, clearCart } = useCart()
  const [agreed, setAgreed] = useState(false)
  const [paymentMethod, setPaymentMethod] = useState('sbp')
  const [receiptFile, setReceiptFile] = useState<File | null>(null)
  const [receiptPreview, setReceiptPreview] = useState<string | null>(null)
  const [createdOrderIds, setCreatedOrderIds] = useState<string[]>([])
  const [step, setStep] = useState<'create' | 'pay'>(fromCart ? 'create' : 'pay')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data: paymentDetails } = useQuery({
    queryKey: ['paymentDetails'],
    queryFn: async () => {
      const { data } = await api.get('/payments/details')
      return data
    },
  })

  const details = paymentDetails?.[paymentMethod] || null

  if (!fromCart && isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!fromCart && !order) return <div className="p-4 text-center text-[var(--tg-hint)]">Заказ не найден</div>

  if (!fromCart && order?.status !== 'WAITING_PAYMENT') {
    return (
      <div className="p-4">
        <Header title="Статус заказа" onBack={() => nav(-1)} />
        <div className="text-center py-8">
          {order && <StatusBadge status={order.status} />}
          {order && <Button className="mt-4" onClick={() => nav(`/order/${order.id}`)}>Перейти к заказу</Button>}
        </div>
      </div>
    )
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > 5 * 1024 * 1024) {
      toast.toast('Файл слишком большой (макс. 5МБ)', 'error')
      return
    }

    setReceiptFile(file)
    if (file.type.startsWith('image/')) {
      const reader = new FileReader()
      reader.onload = (ev) => setReceiptPreview(ev.target?.result as string)
      reader.readAsDataURL(file)
    } else {
      setReceiptPreview(null)
    }
  }

  const handleCreateOrders = async () => {
    if (!agreed) {
      toast.toast('Ознакомьтесь с правилами покупки', 'error')
      return
    }

    try {
      const batchItems = items.map(item => ({
        product_id: item.productId,
        delivery_method: item.deliveryMethod,
        variant_id: item.variantId,
        quantity: item.quantity,
      }))

      const result = await createBatch.mutateAsync(batchItems)
      setCreatedOrderIds(result.order_ids)
      clearCart()
      toast.toast(`Создано ${result.order_ids.length} заказов!`, 'success')
      setStep('pay')
    } catch {
      toast.toast('Ошибка при создании заказов', 'error')
    }
  }

  const handleSubmitPayment = async () => {
    if (!receiptFile) {
      toast.toast('Прикрепите чек об оплате', 'error')
      return
    }

    try {
      if (fromCart && createdOrderIds.length > 0) {
        await confirmBatch.mutateAsync({
          orderIds: createdOrderIds,
          paymentMethod,
          file: receiptFile,
        })
        toast.toast('Чеки отправлены! Ожидайте проверки', 'success')
        nav('/orders')
      } else if (order) {
        await confirmPayment.mutateAsync({ orderId: order.id, paymentMethod, file: receiptFile })
        toast.toast('Чек отправлен! Ожидайте проверки', 'success')
        nav(`/order/${order.id}`)
      }
    } catch {
      toast.toast('Ошибка при отправке чека', 'error')
    }
  }

  const displayTotal = fromCart ? getTotal() : (order ? order.payment_amount || 0 : 0)
  const displayItems = fromCart ? items : (order ? [{
    title: order.product?.title || 'Товар',
    price: order.product?.price || 0,
    quantity: 1,
    deliveryMethod: order.delivery_method,
  }] : [])

  return (
    <div className="pb-24">
      <Header title={fromCart ? (step === 'create' ? 'Оформление корзины' : 'Оплата заказов') : 'Оплата заказа'} onBack={() => nav(-1)} />
      <div className="p-4 space-y-4">
        <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
          {displayItems.map((item, i) => (
            <div key={i} className={`${i > 0 ? 'mt-3 pt-3 border-t border-[var(--tg-border)]' : ''}`}>
              <div className="flex justify-between mb-1">
                <span className="text-[var(--tg-hint)] text-sm">{item.title}</span>
                <span className="font-medium text-sm">{item.quantity} шт</span>
              </div>
              <div className="flex justify-between mb-1">
                <span className="text-[var(--tg-hint)] text-xs">Способ</span>
                <span className="text-xs">{item.deliveryMethod === 'key' ? '🔑 Ключ' : '🔐 Активация'}</span>
              </div>
            </div>
          ))}
          {promoDiscount > 0 && (
            <div className="flex justify-between text-sm text-green-400 mt-2">
              <span>Скидка по промокоду:</span>
              <span>-{formatPrice(promoDiscount)}</span>
            </div>
          )}
          <div className="flex justify-between text-lg font-bold mt-3 pt-3 border-t border-[var(--tg-border)]">
            <span>Итого</span>
            <span className="text-[var(--tg-button)]">{formatPrice(displayTotal)}</span>
          </div>
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Способ оплаты</label>
          <div className="space-y-2">
            {paymentMethods.map((pm) => (
              <label key={pm.id} className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer transition ${paymentMethod === pm.id ? 'border-[var(--tg-button)] bg-[var(--tg-button)]/5' : 'border-[var(--tg-border)]'}`}
                onClick={() => setPaymentMethod(pm.id)}>
                <input type="radio" checked={paymentMethod === pm.id} readOnly className="accent-[var(--tg-button)]" />
                <span>{pm.icon} {pm.label}</span>
              </label>
            ))}
          </div>
        </div>

        {details && (
          <div className="bg-[var(--tg-alert-info-bg)] border border-[var(--tg-alert-info-border)] rounded-xl p-4 text-sm space-y-1">
            <div className="font-medium mb-1 text-[var(--tg-alert-info-text)]">💳 Реквизиты для оплаты:</div>
            {Object.entries(details).map(([k, v]) => (
              <div key={k} className="text-[var(--tg-alert-info-text)]/80">
                <span className="font-medium">{k === 'phone' ? '📱 Номер' : k === 'bank' ? '🏦 Банк' : k === 'receiver' ? '👤 Получатель' : k === 'number' ? '💳 Карта' : k === 'binance' ? 'Binance' : k === 'bybit' ? 'Bybit' : k === 'trc20' ? 'TRC20' : k}:</span> {String(v)}
              </div>
            ))}
          </div>
        )}

        <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
          <label className="text-sm font-medium mb-2 block">📎 Чек об оплате</label>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*,.pdf"
            className="hidden"
            onChange={handleFileSelect}
          />

          {!receiptFile ? (
            <button
              onClick={() => fileInputRef.current?.click()}
              className="w-full py-8 border-2 border-dashed border-[var(--tg-border)] rounded-xl text-[var(--tg-hint)] hover:border-[var(--tg-button)] hover:text-[var(--tg-button)] transition"
            >
              <div className="text-3xl mb-2">📷</div>
              <div className="text-sm">Нажмите чтобы прикрепить чек</div>
              <div className="text-xs mt-1">Изображение или PDF, до 5МБ</div>
            </button>
          ) : (
            <div className="space-y-3">
              {receiptPreview && (
                <div className="relative rounded-xl overflow-hidden">
                  <img src={receiptPreview} alt="Receipt" className="w-full max-h-64 object-contain bg-black/20 rounded-xl" />
                </div>
              )}
              {!receiptPreview && (
                <div className="flex items-center gap-2 p-3 bg-[var(--tg-card)] rounded-lg">
                  <span className="text-lg">📄</span>
                  <span className="text-sm flex-1 truncate">{receiptFile.name}</span>
                  <span className="text-xs text-[var(--tg-hint)]">{(receiptFile.size / 1024).toFixed(0)} КБ</span>
                </div>
              )}
              <div className="flex gap-2">
                <Button
                  variant="secondary"
                  onClick={() => { setReceiptFile(null); setReceiptPreview(null) }}
                  className="flex-1"
                >
                  Удалить
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => fileInputRef.current?.click()}
                  className="flex-1"
                >
                  Заменить
                </Button>
              </div>
            </div>
          )}
        </div>

        <div className="bg-[var(--tg-alert-warning-bg)] border border-[var(--tg-alert-warning-border)] rounded-xl p-4 text-sm">
          <div className="font-medium mb-2 text-[var(--tg-alert-warning-text)]">📋 Правила покупки</div>
          <ul className="space-y-1 text-[var(--tg-alert-warning-text)]/80">
            <li>• После оплаты товар выдаётся автоматически</li>
            <li>• Возврат невозможен после выдачи ключа/активации</li>
            <li>• Данные аккаунта шифруются и используются только для активации</li>
          </ul>
          <label className="flex items-center gap-2 mt-3 cursor-pointer">
            <input type="checkbox" checked={agreed} onChange={(e) => setAgreed(e.target.checked)} className="accent-[var(--tg-button)]" />
            <span className="text-sm text-[var(--tg-alert-warning-text)]/80">Я ознакомлен и согласен с правилами</span>
          </label>
        </div>

        {fromCart && step === 'create' ? (
          <Button
            fullWidth
            size="lg"
            onClick={handleCreateOrders}
            disabled={!agreed}
            loading={createBatch.isPending}
          >
            🛒 Создать заказы ({items.length})
          </Button>
        ) : (
          <Button
            fullWidth
            size="lg"
            onClick={handleSubmitPayment}
            disabled={!receiptFile}
            loading={confirmPayment.isPending || confirmBatch.isPending}
          >
            📤 Отправить чек
          </Button>
        )}
      </div>
    </div>
  )
}
