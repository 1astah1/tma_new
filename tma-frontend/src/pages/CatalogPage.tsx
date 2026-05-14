import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useProducts } from '../hooks/useProducts'
import { ProductCard } from '../components/product/ProductCard'
import { ProductFilters } from '../components/product/ProductFilters'
import { Header } from '../components/layout/Header'
import { Loader, Button } from '../components/ui/Button'

export function CatalogPage() {
  const [searchParams] = useSearchParams()
  const [filters, setFilters] = useState<Record<string, string>>({
    type: searchParams.get('type') || '',
    platform: searchParams.get('platform') || '',
  })

  const { data, isLoading, fetchNextPage, hasNextPage } = useProducts({ ...filters, limit: 20 })

  useEffect(() => {
    const t = searchParams.get('type')
    const p = searchParams.get('platform')
    setFilters({ type: t || '', platform: p || '' })
  }, [searchParams])

  return (
    <div className="pb-24">
      <Header title="Каталог" />
      <div className="p-4 space-y-4">
        <ProductFilters onFilter={setFilters} />
        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-3">
              {data?.data?.map((p) => <ProductCard key={p.id} product={p} />)}
            </div>
            {data?.data?.length === 0 && (
              <div className="text-center py-8 text-tg-hint">Ничего не найдено</div>
            )}
            {hasNextPage && (
              <Button fullWidth onClick={() => fetchNextPage()} className="mt-4">
                Загрузить ещё
              </Button>
            )}
          </>
        )}
      </div>
    </div>
  )
}
