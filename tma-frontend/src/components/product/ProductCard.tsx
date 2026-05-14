import { useNavigate } from 'react-router-dom'
import { Product, platformLabels, typeLabels, platformColors } from '../../types/product'
import { formatPrice } from '../../utils/format'
import { Card } from '../ui/Card'

export function ProductCard({ product }: { product: Product }) {
  const nav = useNavigate()
  return (
    <Card onClick={() => nav(`/product/${product.id}`)} className="flex flex-col">
      <div className="aspect-video bg-gradient-to-br from-tg-button/20 to-tg-secondary rounded-lg mb-3 flex items-center justify-center text-4xl overflow-hidden">
        {product.image_url ? (
          <img src={product.image_url} alt={product.title} className="w-full h-full object-cover" />
        ) : (
          product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'
        )}
      </div>
      <h3 className="font-semibold text-sm line-clamp-2 mb-1">{product.title}</h3>
      <div className="flex gap-1 mb-2">
        <span className={`text-xs px-2 py-0.5 rounded ${platformColors[product.platform]}`}>
          {platformLabels[product.platform]}
        </span>
        <span className="text-xs px-2 py-0.5 rounded bg-gray-100 text-gray-600">
          {typeLabels[product.type]}
        </span>
      </div>
      <div className="mt-auto flex items-center justify-between">
        <span className="text-lg font-bold text-tg-button">{formatPrice(product.price)}</span>
        {product.delivery_methods.includes('activation') && (
          <span className="text-xs text-tg-hint">🔐 Активация</span>
        )}
      </div>
    </Card>
  )
}
