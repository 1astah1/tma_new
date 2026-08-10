import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCart } from '../stores/cartStore'
import { useCreateBatchOrder } from '../hooks/useOrders'
import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { Card } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { formatPrice } from '../utils/format'
import { openManagerCartChat } from '../utils/managerChat'
import { useToast } from '../components/ui/Toast'
import api from '../services/api'

export function CartPage() {
  const nav = useNavigate()
  const goBack = useGoBack('/')
  const toast = useToast()
  const { items, removeItem, updateQuantity, getTotal, getSubtotal, getItemCount, clearCart, promoCode, promoDiscount, setPromoCode, setPromoDiscount } = useCart()
  const createBatch = useCreateBatchOrder()
  const [promoInput, setPromoInput] = useState('')
  const [validating, setValidating] = useState(false)

  const validatePromo = async () => {
    if (!promoInput.trim()) return
    setValidating(true)
    try {
      const { data } = await api.post('/promo/validate', {
        code: promoInput.trim(),
        subtotal: getSubtotal(),
      })
      if (data.valid) {
        setPromoCode(promoInput.trim())
        setPromoDiscount(data.discount)
        toast.toast(`Промокод применён: скидка ${formatPrice(data.discount)}`, 'success')
      } else {
        toast.toast(data.error || 'Неверный промокод', 'error')
      }
    } catch {
      toast.toast('Ошибка проверки промокода', 'error')
    } finally {
      setValidating(false)
    }
  }

  const removePromo = () => {
    setPromoCode(null)
    setPromoDiscount(0)
    setPromoInput('')
  }

  if (items.length === 0) {
    return (
      <div className="pb-page">
        <Header title="Корзина" onBack={goBack} />
        <div className="flex flex-col items-center justify-center py-16 text-[var(--tg-hint)]">
          <div className="text-5xl mb-4">🛒</div>
          <p className="text-lg">Корзина пуста</p>
          <Button variant="primary" onClick={() => nav('/catalog')} className="mt-4">
            Перейти в каталог
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="pb-page-bar">
      <Header title={`Корзина (${getItemCount()})`} onBack={goBack} />
      <div className="p-4 space-y-3">
        {items.map((item) => (
          <Card key={`${item.productId}-${item.variantId || 'default'}`}>
            <div className="flex gap-3">
              {item.image && (
                <img
                  src={item.image}
                  alt={item.title}
                  className="w-16 h-16 rounded-lg object-cover"
                />
              )}
              <div className="min-w-0 flex-1">
                <div className="truncate font-medium text-sm">{item.title}</div>
                <div className="text-xs text-[var(--tg-hint)] mt-0.5">
                  💬 Через менеджера
                </div>
                <div className="font-bold text-[var(--tg-button)] mt-1">
                  {formatPrice(item.price * item.quantity)}
                </div>
              </div>
            </div>
            <div className="flex items-center justify-between mt-3">
              <div className="flex items-center gap-2 bg-[var(--tg-secondary)] rounded-lg">
                <button
                  onClick={() => updateQuantity(item.productId, item.variantId, item.quantity - 1)}
                  className="px-3 py-1.5 text-sm hover:bg-[var(--tg-card)] rounded-l-lg"
                >
                  −
                </button>
                <span className="text-sm font-medium w-6 text-center">{item.quantity}</span>
                <button
                  onClick={() => updateQuantity(item.productId, item.variantId, item.quantity + 1)}
                  className="px-3 py-1.5 text-sm hover:bg-[var(--tg-card)] rounded-r-lg"
                >
                  +
                </button>
              </div>
              <button
                onClick={() => removeItem(item.productId, item.variantId)}
                className="text-xs text-red-400 hover:text-red-300"
              >
                Удалить
              </button>
            </div>
          </Card>
        ))}

        {promoCode ? (
          <div className="flex items-center justify-between p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
            <div>
              <div className="text-sm font-medium text-green-400">✅ Промокод: {promoCode}</div>
              <div className="text-xs text-green-300">Скидка: {formatPrice(promoDiscount)}</div>
            </div>
            <button onClick={removePromo} className="text-xs text-red-400 hover:text-red-300">Удалить</button>
          </div>
        ) : (
          <div className="flex gap-2">
            <input
              type="text"
              value={promoInput}
              onChange={(e) => setPromoInput(e.target.value.toUpperCase())}
              placeholder="Промокод"
              className="flex-1 min-w-0 px-3 py-2 bg-[var(--tg-secondary)] border border-[var(--tg-border)] rounded-lg text-base focus:outline-none focus:border-[var(--tg-button)]"
            />
            <Button variant="secondary" onClick={validatePromo} loading={validating}>
              OK
            </Button>
          </div>
        )}

        <div className="fixed bottom-above-nav left-0 right-0 z-40 border-t border-[var(--tg-secondary)] bg-[var(--tg-bg)]/95 backdrop-blur-lg">
          <div className="mx-auto max-w-lg p-4">
          <div className="space-y-1 mb-3">
            <div className="flex justify-between text-sm">
              <span className="text-[var(--tg-hint)]">Подытог:</span>
              <span>{formatPrice(getSubtotal())}</span>
            </div>
            {promoDiscount > 0 && (
              <div className="flex justify-between text-sm text-green-400">
                <span>Скидка:</span>
                <span>-{formatPrice(promoDiscount)}</span>
              </div>
            )}
            <div className="flex justify-between items-center pt-2 border-t border-[var(--tg-border)]">
              <span className="text-[var(--tg-hint)] font-medium">Итого:</span>
              <span className="text-2xl font-bold text-[var(--tg-button)]">
                {formatPrice(getTotal())}
              </span>
            </div>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Button variant="secondary" onClick={clearCart} className="w-full flex-1 min-w-0 text-sm sm:text-base">
              Очистить
            </Button>
            <Button
              variant="primary"
              onClick={async () => {
                try {
                  const result = await createBatch.mutateAsync({
                    items: items.map((i) => ({
                      product_id: i.productId,
                      delivery_method: 'activation' as const,
                      variant_id: i.variantId,
                      quantity: i.quantity,
                    })),
                    promoCode: promoCode || undefined,
                  })
                  const orderIds = result.order_ids ?? result.orders?.map((o) => o.id) ?? []
                  clearCart()
                  if (orderIds.length === 1) {
                    nav(`/orders/${orderIds[0]}`)
                  }
                  openManagerCartChat(orderIds)
                } catch {
                  toast.toast('Не удалось оформить заказ. Попробуйте ещё раз.', 'error')
                }
              }}
              loading={createBatch.isPending}
              className="w-full flex-1 min-w-0 text-sm sm:text-base"
            >
              <span className="truncate">Купить в Telegram ({getItemCount()})</span>
            </Button>
          </div>
          </div>
        </div>
      </div>
    </div>
  )
}
