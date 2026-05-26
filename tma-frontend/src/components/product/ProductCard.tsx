import { useNavigate } from 'react-router-dom'
import { Product, platformLabels, typeLabels, platformColors } from '../../types/product'
import { formatPrice } from '../../utils/format'
import { useWishlist } from '../../stores/wishlistStore'

export function ProductCard({ product }: { product: Product }) {
  const nav = useNavigate()
  const { toggleItem, isWishlisted } = useWishlist()
  const hasDiscount = product.discount_percent > 0
  const discountedPrice = hasDiscount ? product.price * (1 - product.discount_percent / 100) : product.price
  const wishlisted = isWishlisted(product.id)

  return (
    <div
      onClick={() => nav(`/product/${product.id}`)}
      className="bg-[var(--tg-card)] rounded-xl border border-[var(--tg-border)] overflow-hidden cursor-pointer hover:shadow-lg hover:border-[var(--tg-button)]/30 transition-all duration-200 group"
    >
      <div className="relative aspect-[4/3] bg-gradient-to-br from-[var(--tg-button)]/10 to-[var(--tg-secondary)] overflow-hidden">
        {product.image_url ? (
          <img src={product.image_url} alt={product.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-5xl">
            {product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'}
          </div>
        )}

        {hasDiscount && (
          <span className="absolute top-2 left-2 bg-red-500 text-white text-[10px] font-bold px-2 py-0.5 rounded-md shadow">
            -{product.discount_percent}%
          </span>
        )}

        <button
          onClick={(e) => { e.stopPropagation(); toggleItem(product.id) }}
          className="absolute top-2 right-2 w-7 h-7 rounded-full bg-black/40 backdrop-blur-sm flex items-center justify-center"
        >
          <span className={`text-sm ${wishlisted ? '' : 'grayscale'}`}>❤️</span>
        </button>

        {product.delivery_methods.includes('activation') && (
          <span className="absolute bottom-2 left-2 bg-black/40 backdrop-blur-sm text-white text-[10px] px-2 py-0.5 rounded-md">
            🔐 Активация
          </span>
        )}
      </div>

      <div className="p-3">
        <h3 className="font-medium text-sm leading-tight line-clamp-2 mb-2 text-[var(--tg-text)]">
          {product.title}
        </h3>

        <div className="flex items-center gap-1.5 mb-2.5">
          <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded ${platformColors[product.platform]}`}>
            {platformLabels[product.platform]}
          </span>
          <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-[var(--tg-secondary)] text-[var(--tg-hint)]">
            {typeLabels[product.type]}
          </span>
        </div>

        <div className="flex items-end justify-between">
          <div className="flex flex-col">
            {hasDiscount && (
              <span className="text-[10px] text-[var(--tg-hint)] line-through">{formatPrice(product.price)}</span>
            )}
            <span className="text-base font-bold text-[var(--tg-button)] leading-none">
              {formatPrice(discountedPrice)}
            </span>
          </div>

          {product.delivery_methods.includes('key') && !product.delivery_methods.includes('activation') && (
            <span className="text-[10px] text-[var(--tg-hint)]">🔑 Ключ</span>
          )}
        </div>
      </div>
    </div>
  )
}
