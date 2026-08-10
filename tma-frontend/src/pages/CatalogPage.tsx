import { useState, useEffect, useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useProducts } from '../hooks/useProducts'
import { ProductCard } from '../components/product/ProductCard'
import { ProductFilters, FilterValue } from '../components/product/ProductFilters'
import { Header } from '../components/layout/Header'
import { Loader } from '../components/ui/Button'
import { Pagination } from '../components/ui/Pagination'
import { filterProductsBySection } from '../utils/productSection'
import { groupProductsForCatalog } from '../utils/groupProducts'

function filtersFromParams(params: URLSearchParams): FilterValue {
  const platform = params.get('platform') || ''
  const section = params.get('section') || ''
  return {
    type: params.get('type') || '',
    platform,
    section: platform ? (section || 'game') : section,
    search: params.get('search') || '',
  }
}

function paramsFromFilters(filters: Record<string, string>) {
  const params = new URLSearchParams()
  if (filters.platform) params.set('platform', filters.platform)
  if (filters.platform && filters.section) {
    params.set('section', filters.section)
  } else if (filters.type) {
    params.set('type', filters.type)
  }
  if (filters.search) params.set('search', filters.search)
  return params
}

export function CatalogPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<FilterValue>(() => filtersFromParams(searchParams))

  const queryFilters = useMemo(() => {
    if (filters.platform) {
      return {
        platform: filters.platform,
        section: filters.section || 'game',
        search: filters.search,
        limit: 20,
        page,
      }
    }
    return {
      type: filters.type,
      search: filters.search,
      limit: 20,
      page,
    }
  }, [filters, page])

  const { data, isLoading, isFetching } = useProducts(queryFilters)
  const products = useMemo(() => {
    const list = data?.data ?? []
    if (filters.platform && filters.section) {
      return filterProductsBySection(list, filters.section)
    }
    return list
  }, [data?.data, filters.platform, filters.section])
  const crossPlatform = !filters.platform
  const cards = useMemo(
    () => groupProductsForCatalog(products, crossPlatform),
    [products, crossPlatform],
  )

  useEffect(() => {
    setFilters(filtersFromParams(searchParams))
    setPage(1)
  }, [searchParams])

  const handleFilterChange = useCallback((next: Record<string, string>) => {
    setFilters(next)
    setPage(1)
    setSearchParams(paramsFromFilters(next), { replace: true })
  }, [setSearchParams])

  return (
    <div className="pb-page">
      <Header title="Каталог" showCart />
      <div className="p-4 space-y-4">
        <ProductFilters mode="catalog" value={filters} onChange={handleFilterChange} />
        {isLoading || (isFetching && !data?.data) ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          <>
            <div className="grid grid-cols-2 items-stretch gap-3">
              {cards.map((card) => <ProductCard key={card.id} card={card} />)}
            </div>
            {cards.length === 0 && (
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
