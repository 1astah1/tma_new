import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProduct } from '../hooks/useProducts'
import { useCreateOrder } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { formatPrice } from '../utils/format'
import { platformLabels, typeLabels, platformColors } from '../types/product'

export function ProductPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const { data: product, isLoading } = useProduct(id!)
  const createOrder = useCreateOrder()
  const [deliveryMethod, setDeliveryMethod] = useState<'key' | 'activation'>('key')

  const handleBuy = async () => {
    if (!product) return
    try {
      const order = await createOrder.mutateAsync({ productId: product.id, deliveryMethod })
      nav(`/order/${order.id}`)
    } catch (e) {
      alert('Ошибка при создании заказа')
    }
  }

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!product) return <div className="p-4 text-center text-tg-hint">Товар не найден</div>

  return (
    <div className="pb-24">
      <Header title="" onBack={() => nav(-1)} />
      <div className="p-4">
        <div className="aspect-video bg-gradient-to-br from-tg-button/20 to-tg-secondary rounded-xl mb-4 flex items-center justify-center text-6xl overflow-hidden">
          {product.image_url ? (
            <img src={product.image_url} alt={product.title} className="w-full h-full object-cover" />
          ) : (
            product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'
          )}
        </div>

        <h1 className="text-xl font-bold mb-2">{product.title}</h1>

        <div className="flex gap-2 mb-3">
          <span className={`text-sm px-2 py-0.5 rounded ${platformColors[product.platform]}`}>
            {platformLabels[product.platform]}
          </span>
          <span className="text-sm px-2 py-0.5 rounded bg-gray-100 text-gray-600">
            {typeLabels[product.type]}
          </span>
        </div>

        {product.description && (
          <p className="text-sm text-gray-600 mb-4">{product.description}</p>
        )}

        <div className="text-2xl font-bold text-tg-button mb-6">{formatPrice(product.price)}</div>

        <div className="space-y-3 mb-6">
          <label className="text-sm font-medium">Способ получения</label>
          {product.delivery_methods.includes('key') && (
            <label
              className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer ${
                deliveryMethod === 'key' ? 'border-tg-button bg-tg-button/5' : 'border-gray-200'
              }`}
              onClick={() => setDeliveryMethod('key')}
            >
              <input type="radio" checked={deliveryMethod === 'key'} readOnly className="accent-tg-button" />
              <div>
                <div className="font-medium">🔑 Ключ</div>
                <div className="text-xs text-tg-hint">Получите код активации сразу</div>
              </div>
            </label>
          )}
          {product.delivery_methods.includes('activation') && (
            <label
              className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer ${
                deliveryMethod === 'activation' ? 'border-tg-button bg-tg-button/5' : 'border-gray-200'
              }`}
              onClick={() => setDeliveryMethod('activation')}
            >
              <input type="radio" checked={deliveryMethod === 'activation'} readOnly className="accent-tg-button" />
              <div>
                <div className="font-medium">🔐 Активация на аккаунт</div>
                <div className="text-xs text-tg-hint">Активируем на ваш аккаунт</div>
              </div>
            </label>
          )}
        </div>

        <Button
          fullWidth
          size="lg"
          onClick={handleBuy}
          loading={createOrder.isPending}
        >
          🛒 Купить за {formatPrice(product.price)}
        </Button>
      </div>
    </div>
  )
}
