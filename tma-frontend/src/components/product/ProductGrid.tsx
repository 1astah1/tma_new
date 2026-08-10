import { useMemo } from 'react'
import { Product } from '../../types/product'
import { groupProductsForCatalog } from '../../utils/groupProducts'
import { enrichProductsWithTitleKeyPool } from '../../utils/mergeProductPrices'
import { ProductCard } from './ProductCard'

export function ProductGrid({
  products,
  pricePool,
  crossPlatform = true,
  emptyText = 'Ничего не найдено',
}: {
  products: Product[]
  pricePool?: Product[]
  crossPlatform?: boolean
  emptyText?: string
}) {
  const cards = useMemo(() => {
    const pool = pricePool ?? products
    const enriched = enrichProductsWithTitleKeyPool(products, pool)
    return groupProductsForCatalog(enriched, crossPlatform)
  }, [products, pricePool, crossPlatform])

  if (!cards.length) {
    return <div className="py-6 text-center text-[var(--tg-hint)]">{emptyText}</div>
  }

  return (
    <div className="grid grid-cols-2 items-stretch gap-3">
      {cards.map((card) => (
        <ProductCard key={card.id} card={card} />
      ))}
    </div>
  )
}
