import { useParams } from 'react-router-dom'
import { useHomeCategory } from '../hooks/useProducts'
import { useGoBack } from '../hooks/useGoBack'
import { ProductGrid } from '../components/product/ProductGrid'
import { Header } from '../components/layout/Header'
import { Loader } from '../components/ui/Button'

export function HomeCategoryPage() {
  const { id } = useParams<{ id: string }>()
  const goBack = useGoBack('/')
  const { data, isLoading } = useHomeCategory(id || '')
  const category = data?.data

  return (
    <div className="pb-page">
      <Header title={category?.title || 'Категория'} onBack={goBack} showCart />
      <div className="p-4">
        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          <>
            {category?.image_url ? (
              <div className="mb-4 overflow-hidden rounded-2xl">
                <img src={category.image_url} alt={category.title} className="w-full h-auto block" />
              </div>
            ) : null}
            <ProductGrid products={category?.products ?? []} />
            {!category?.products?.length ? (
              <div className="py-8 text-center text-[var(--tg-hint)]">В этой категории пока нет товаров</div>
            ) : null}
          </>
        )}
      </div>
    </div>
  )
}
