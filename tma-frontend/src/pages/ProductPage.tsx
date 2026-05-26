import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProduct } from '../hooks/useProducts'
import { useCreateOrder } from '../hooks/useOrders'
import { useToast } from '../components/ui/Toast'
import { useCart } from '../stores/cartStore'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { formatPrice } from '../utils/format'
import { platformLabels, typeLabels, platformColors, ProductVariant } from '../types/product'

export function ProductPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const toast = useToast()
  const { data: product, isLoading } = useProduct(id!)
  const createOrder = useCreateOrder()
  const { addItem, getItemCount } = useCart()
  const [deliveryMethod, setDeliveryMethod] = useState<'key' | 'activation'>('key')
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [quantity, setQuantity] = useState(1)

  const handleAddToCart = () => {
    if (!product) return
    addItem({
      productId: product.id,
      variantId: selectedVariant?.id,
      quantity: (product.type === 'currency' || product.type === 'subscription') ? quantity : 1,
      title: product.title,
      price: selectedVariant ? selectedVariant.price : (hasDiscount ? product.price * (1 - product.discount_percent / 100) : product.price),
      image: product.image_url || undefined,
      deliveryMethod,
    })
    toast.toast(`Добавлено в корзину (${getItemCount()})`, 'success')
  }

  const hasDiscount = product ? product.discount_percent > 0 : false
  const basePrice = product ? (hasDiscount ? product.price * (1 - product.discount_percent / 100) : product.price) : 0
  const variantPrice = selectedVariant ? selectedVariant.price : 0
  const totalPrice = product?.type === 'currency' || product?.type === 'subscription'
    ? (variantPrice || basePrice) * quantity
    : basePrice

  const handleBuy = async () => {
    if (!product) return
    try {
      const order = await createOrder.mutateAsync({
        productId: product.id,
        deliveryMethod,
        variantId: selectedVariant?.id,
        quantity: (product.type === 'currency' || product.type === 'subscription') ? quantity : 1,
      })
      nav(`/order/${order.id}`)
    } catch {
      toast.toast('Ошибка при создании заказа', 'error')
    }
  }

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!product) return <div className="p-4 text-center text-[var(--tg-hint)]">Товар не найден</div>

  const variants: ProductVariant[] = Array.isArray(product.variants) ? product.variants : []
  const hasVariants = variants.length > 0 && (product.type === 'currency' || product.type === 'subscription')
  const isKeyDelivery = deliveryMethod === 'key'
  const noKeys = isKeyDelivery && (product as any).available_keys === 0

  return (
    <div className="pb-24">
      <Header title="" onBack={() => nav(-1)} />
      <div className="p-4">
        <div className="aspect-video bg-gradient-to-br from-[var(--tg-button)]/20 to-[var(--tg-secondary)] rounded-xl mb-4 flex items-center justify-center text-6xl overflow-hidden relative">
          {product.image_url ? (
            <img src={product.image_url} alt={product.title} className="w-full h-full object-cover" />
          ) : (
            product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'
          )}
          {hasDiscount && (
            <span className="absolute top-3 left-3 bg-red-500 text-white text-sm font-bold px-3 py-1 rounded-lg shadow-lg">
              -{product.discount_percent}%
            </span>
          )}
        </div>

        <h1 className="text-xl font-bold mb-2">{product.title}</h1>

        <div className="flex gap-2 mb-3">
          <span className={`text-sm px-2 py-0.5 rounded ${platformColors[product.platform]}`}>
            {platformLabels[product.platform]}
          </span>
          <span className="text-sm px-2 py-0.5 rounded bg-[var(--tg-secondary)] text-[var(--tg-hint)]">
            {typeLabels[product.type]}
          </span>
        </div>

        {product.description && (
          <p className="text-sm text-[var(--tg-hint)] mb-4">{product.description}</p>
        )}

        <div className="mb-6">
          {hasDiscount && (
            <span className="text-lg text-[var(--tg-hint)] line-through mr-2">{formatPrice(product.price)}</span>
          )}
          <span className="text-3xl font-bold text-[var(--tg-button)]">{formatPrice(basePrice)}</span>
        </div>

        {hasVariants && (
          <div className="mb-6">
            <label className="text-sm font-medium mb-2 block">Выберите количество</label>
            <div className="grid grid-cols-3 gap-2">
              {variants.map((v) => (
                <button
                  key={v.id}
                  onClick={() => { setSelectedVariant(v); setQuantity(1) }}
                  className={`p-3 rounded-xl border-2 text-center transition ${
                    selectedVariant?.id === v.id
                      ? 'border-[var(--tg-button)] bg-[var(--tg-button)]/10'
                      : 'border-[var(--tg-border)] hover:border-[var(--tg-hint)]'
                  }`}
                >
                  <div className="font-medium text-sm">{v.name}</div>
                  <div className="text-xs text-[var(--tg-hint)]">{formatPrice(v.price)}</div>
                </button>
              ))}
            </div>
          </div>
        )}

        {(product.type === 'currency' || product.type === 'subscription') && (selectedVariant || !hasVariants) && (
          <div className="mb-6">
            <label className="text-sm font-medium mb-2 block">Количество</label>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
                className="w-10 h-10 rounded-lg bg-[var(--tg-secondary)] flex items-center justify-center text-xl font-bold hover:bg-[var(--tg-card)]"
              >
                −
              </button>
              <span className="text-xl font-bold w-12 text-center">{quantity}</span>
              <button
                onClick={() => setQuantity(quantity + 1)}
                className="w-10 h-10 rounded-lg bg-[var(--tg-secondary)] flex items-center justify-center text-xl font-bold hover:bg-[var(--tg-card)]"
              >
                +
              </button>
            </div>
          </div>
        )}

        {(product.type === 'currency' || product.type === 'subscription') && (
          <div className="mb-6 p-4 bg-[var(--tg-secondary)] rounded-xl">
            <div className="flex justify-between text-sm mb-1">
              <span className="text-[var(--tg-hint)]">Цена за единицу</span>
              <span>{formatPrice(selectedVariant ? selectedVariant.price : basePrice)}</span>
            </div>
            <div className="flex justify-between text-sm mb-2">
              <span className="text-[var(--tg-hint)]">Количество</span>
              <span>× {quantity}</span>
            </div>
            <div className="border-t border-[var(--tg-border)] pt-2 flex justify-between font-bold text-lg">
              <span>Итого</span>
              <span className="text-[var(--tg-button)]">{formatPrice(totalPrice)}</span>
            </div>
          </div>
        )}

        {noKeys && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-xl text-center">
            <p className="text-sm text-red-400 font-medium">⚠ Ключи закончились. Ожидайте пополнения.</p>
          </div>
        )}

        <div className="space-y-3 mb-6">
          <label className="text-sm font-medium">Способ получения</label>
          {product.delivery_methods.includes('key') && (
            <label
              className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer ${
                deliveryMethod === 'key' ? 'border-[var(--tg-button)] bg-[var(--tg-button)]/5' : 'border-[var(--tg-border)]'
              }`}
              onClick={() => setDeliveryMethod('key')}
            >
              <input type="radio" checked={deliveryMethod === 'key'} readOnly className="accent-[var(--tg-button)]" />
              <div>
                <div className="font-medium">🔑 Ключ</div>
                <div className="text-xs text-[var(--tg-hint)]">Получите код активации сразу</div>
              </div>
            </label>
          )}
          {product.delivery_methods.includes('activation') && (
            <label
              className={`flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer ${
                deliveryMethod === 'activation' ? 'border-[var(--tg-button)] bg-[var(--tg-button)]/5' : 'border-[var(--tg-border)]'
              }`}
              onClick={() => setDeliveryMethod('activation')}
            >
              <input type="radio" checked={deliveryMethod === 'activation'} readOnly className="accent-[var(--tg-button)]" />
              <div>
                <div className="font-medium">🔐 Активация на аккаунт</div>
                <div className="text-xs text-[var(--tg-hint)]">Активируем на ваш аккаунт</div>
              </div>
            </label>
          )}
        </div>

        <div className="flex gap-2 mb-6">
          <Button
            size="lg"
            onClick={handleAddToCart}
            className="flex-1"
            variant="secondary"
            disabled={noKeys}
          >
            {noKeys ? 'Нет в наличии' : 'В корзину'}
          </Button>
          <Button
            fullWidth
            size="lg"
            onClick={handleBuy}
            loading={createOrder.isPending}
            className="flex-1"
            disabled={noKeys}
          >
            {noKeys ? 'Нет в наличии' : `Купить за ${formatPrice(totalPrice)}`}
          </Button>
        </div>
      </div>
    </div>
  )
}
