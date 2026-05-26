import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useProducts } from '../hooks/useProducts'
import { ProductCard } from '../components/product/ProductCard'
import { ProductFilters } from '../components/product/ProductFilters'
import { Header } from '../components/layout/Header'
import { Loader, Button } from '../components/ui/Button'
import { Pagination } from '../components/ui/Pagination'

export function CatalogPage() {
  const [searchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<Record<string, string>>({
    type: searchParams.get('type') || '',
    platform: searchParams.get('platform') || '',
  })

  const { data, isLoading } = useProducts({ ...filters, limit: 20, page })

  useEffect(() => {
    const t = searchParams.get('type')
    const p = searchParams.get('platform')
    setFilters({ type: t || '', platform: p || '' })
    setPage(1)
  }, [searchParams])

  return (
    <div className="pb-24">
      <Header title="Каталог" />
      <div className="p-4 space-y-4">
        <ProductFilters onFilter={(f) => { setFilters(f); setPage(1) }} />
        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-2.5">
              {data?.data?.map((p) => <ProductCard key={p.id} product={p} />)}
            </div>
            {data?.data?.length === 0 && (
              <div className="text-center py-8 text-[var(--tg-hint)]">Ничего не найдено</div>
            )}
            <Pagination
              page={page}
              total={data?.meta?.total || 0}
              limit={20}
              onPageChange={setPage}
            />
          </>
        )}
      </div>
    </div>
  )
}
