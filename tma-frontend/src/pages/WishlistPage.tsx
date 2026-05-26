import { useNavigate } from 'react-router-dom'
import { useWishlist } from '../stores/wishlistStore'
import { useProducts } from '../hooks/useProducts'
import { Header } from '../components/layout/Header'
import { ProductCard } from '../components/product/ProductCard'
import { Button } from '../components/ui/Button'

export function WishlistPage() {
  const nav = useNavigate()
  const { items, getCount } = useWishlist()
  const { data, isLoading } = useProducts({ limit: 100 })

  const wishlistProducts = data?.data?.filter((p) => items.includes(p.id)) || []

  if (wishlistProducts.length === 0) {
    return (
      <div className="pb-24">
        <Header title="Избранное" />
        <div className="flex flex-col items-center justify-center py-16 text-[var(--tg-hint)]">
          <div className="text-5xl mb-4">❤️</div>
          <p className="text-lg">Список избранного пуст</p>
          <Button variant="primary" onClick={() => nav('/catalog')} className="mt-4">
            Перейти в каталог
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="pb-24">
      <Header title={`Избранное (${getCount()})`} />
      <div className="p-4">
        <div className="grid grid-cols-2 gap-2.5">
          {wishlistProducts.map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
      </div>
    </div>
  )
}
