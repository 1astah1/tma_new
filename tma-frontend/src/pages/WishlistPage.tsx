import { useNavigate } from 'react-router-dom'
import { useWishlist } from '../stores/wishlistStore'
import { useWishlistProducts } from '../hooks/useWishlistProducts'
import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { ProductGrid } from '../components/product/ProductGrid'
import { Button, Loader } from '../components/ui/Button'

export function WishlistPage() {
  const nav = useNavigate()
  const goBack = useGoBack('/profile')
  const { getCount } = useWishlist()
  const { products, isLoading, totalItems, missingCount } = useWishlistProducts()
  const count = getCount()

  if (isLoading) {
    return (
      <div className="pb-page">
        <Header title={`Избранное (${count})`} onBack={goBack} />
        <div className="flex justify-center py-16">
          <Loader />
        </div>
      </div>
    )
  }

  if (totalItems === 0) {
    return (
      <div className="pb-page">
        <Header title="Избранное" onBack={goBack} />
        <div className="flex flex-col items-center justify-center py-16 text-[var(--tg-hint)]">
          <div className="mb-4 text-5xl">❤️</div>
          <p className="text-lg">Список избранного пуст</p>
          <Button variant="primary" onClick={() => nav('/catalog')} className="mt-4">
            Перейти в каталог
          </Button>
        </div>
      </div>
    )
  }

  if (products.length === 0) {
    return (
      <div className="pb-page">
        <Header title={`Избранное (${count})`} onBack={goBack} />
        <div className="flex flex-col items-center justify-center px-6 py-16 text-center text-[var(--tg-hint)]">
          <div className="mb-4 text-5xl">❤️</div>
          <p className="text-lg">Не удалось загрузить игры из избранного</p>
          <p className="mt-2 text-sm">Попробуйте обновить страницу или добавьте товары заново</p>
          <Button variant="primary" onClick={() => nav('/catalog')} className="mt-4">
            Перейти в каталог
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="pb-page">
      <Header title={`Избранное (${count})`} onBack={goBack} />
      <div className="p-4">
        {missingCount > 0 ? (
          <p className="mb-3 text-center text-xs text-white/40">
            {missingCount} {missingCount === 1 ? 'игра больше недоступна' : 'игры больше недоступны'}
          </p>
        ) : null}
        <ProductGrid products={products} crossPlatform={false} />
      </div>
    </div>
  )
}
