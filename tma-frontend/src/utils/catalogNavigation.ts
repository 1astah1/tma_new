import type { FilterValue } from '../components/product/ProductFilters'

export function filtersToCatalogSearch(filters: FilterValue): string {
  const params = new URLSearchParams()
  if (filters.platform) params.set('platform', filters.platform)
  if (filters.platform && filters.section) {
    params.set('section', filters.section)
  } else if (filters.type) {
    params.set('type', filters.type)
  }
  if (filters.search?.trim()) params.set('search', filters.search.trim())
  const qs = params.toString()
  return qs ? `/catalog?${qs}` : '/catalog'
}
