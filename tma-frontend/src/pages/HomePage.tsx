import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useHomeCatalog, useHomeBanners, useProducts } from '../hooks/useProducts'
import { ProductGrid } from '../components/product/ProductGrid'
import { ProductFilters, FilterValue } from '../components/product/ProductFilters'
import { HomeBannerCarousel } from '../components/home/HomeBannerCarousel'
import { resolveHomeBanners } from '../constants/defaultHomeBanners'
import { Header } from '../components/layout/Header'
import { Loader } from '../components/ui/Button'
import { homeSectionLabels } from '../constants/homeSections'
import { filterProductsBySection } from '../utils/productSection'
import { filtersToCatalogSearch } from '../utils/catalogNavigation'
import { Button } from '../components/ui/Button'

export function HomePage() {
  const nav = useNavigate()
  const [filters, setFilters] = useState<FilterValue>({})
  const { data: bannersData } = useHomeBanners()
  const isAllPlatform = !filters.platform
  const activeSection = filters.section || 'game'

  const {
    data: catalog,
    isLoading: catalogLoading,
    isError: catalogError,
    refetch: refetchCatalog,
    isFetching: catalogFetching,
  } = useHomeCatalog(isAllPlatform)

  const { data: platformData, isLoading: platformLoading, isFetching: platformFetching } = useProducts(
    {
      platform: filters.platform,
      section: activeSection,
      search: filters.search,
      limit: 20,
    },
    !isAllPlatform,
  )

  const banners = useMemo(
    () => resolveHomeBanners(bannersData?.data),
    [bannersData?.data],
  )

  const platformProducts = useMemo(
    () => filterProductsBySection(platformData?.data ?? [], activeSection),
    [platformData?.data, activeSection],
  )

  const preorders = catalog?.preorders ?? []
  const newReleases = catalog?.new_releases ?? []
  const popular = catalog?.popular ?? []
  const homePricePool = useMemo(
    () => [...preorders, ...newReleases, ...popular],
    [preorders, newReleases, popular],
  )
  const hasCatalogData = preorders.length + newReleases.length + popular.length > 0
  const showCatalogSkeleton = isAllPlatform && catalogLoading && !hasCatalogData

  return (
    <div className="pb-page">
      <Header showLogo showCart />
      <HomeBannerCarousel banners={banners} />
      <div className="p-4">
        <ProductFilters mode="home" value={filters} onChange={setFilters} />

        {isAllPlatform ? (
          showCatalogSkeleton ? (
            <div className="mt-5 space-y-6 animate-pulse">
              <div className="h-6 w-32 rounded bg-white/10" />
              <div className="grid grid-cols-2 items-stretch gap-3">
                <div className="aspect-[3/4] rounded-3xl bg-white/5" />
                <div className="aspect-[3/4] rounded-3xl bg-white/5" />
                <div className="aspect-[3/4] rounded-3xl bg-white/5" />
                <div className="aspect-[3/4] rounded-3xl bg-white/5" />
              </div>
            </div>
          ) : catalogError && !hasCatalogData ? (
            <div className="py-8 text-center space-y-3">
              <p className="text-red-400 text-sm">
                Не удалось загрузить каталог. Убедитесь, что backend запущен.
              </p>
              <button
                type="button"
                onClick={() => refetchCatalog()}
                disabled={catalogFetching}
                className="rounded-xl bg-amber-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
              >
                {catalogFetching ? 'Загрузка...' : 'Повторить'}
              </button>
            </div>
          ) : (
            <div className="space-y-6 mt-5">
              <section>
                <h2 className="text-lg font-semibold mb-3">Предзаказы</h2>
                <ProductGrid products={preorders} pricePool={homePricePool} />
              </section>
              <section>
                <h2 className="text-lg font-semibold mb-3">Новинки</h2>
                <ProductGrid products={newReleases} pricePool={homePricePool} />
              </section>
              <section>
                <h2 className="text-lg font-semibold mb-3">Популярное</h2>
                <ProductGrid products={popular} pricePool={homePricePool} />
              </section>
              {hasCatalogData ? (
                <div className="pt-2">
                  <Button
                    fullWidth
                    size="lg"
                    onClick={() => nav(filtersToCatalogSearch(filters))}
                  >
                    Перейти в каталог
                  </Button>
                  <p className="mt-2 text-center text-xs text-white/40">
                    На главной — подборки. В каталоге — полный список с фильтрами.
                  </p>
                </div>
              ) : null}
            </div>
          )
        ) : (
          <>
            <h2 className="text-lg font-semibold mt-5 mb-3">
              {homeSectionLabels[activeSection as keyof typeof homeSectionLabels] || 'Игры'}
            </h2>
            {platformLoading || (platformFetching && !platformData?.data) ? (
              <div className="flex justify-center py-8"><Loader /></div>
            ) : (
              <>
                <ProductGrid products={platformProducts} crossPlatform={false} />
                {platformProducts.length > 0 ? (
                  <div className="pt-4">
                    <Button
                      fullWidth
                      size="lg"
                      onClick={() => nav(filtersToCatalogSearch(filters))}
                    >
                      Перейти в каталог
                    </Button>
                    <p className="mt-2 text-center text-xs text-white/40">
                      На главной показаны первые игры. В каталоге — полный список с тем же фильтром.
                    </p>
                  </div>
                ) : null}
              </>
            )}
          </>
        )}
      </div>
    </div>
  )
}
