import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProducts } from '../hooks/useProducts'
import { ProductCard } from '../components/product/ProductCard'
import { ProductFilters } from '../components/product/ProductFilters'
import { Header } from '../components/layout/Header'
import { Loader } from '../components/ui/Button'

const categories = [
  { id: 'game', label: 'Игры', image: '/Игры.png' },
  { id: 'currency', label: 'Валюта', image: '/Валюты.png' },
  { id: 'subscription', label: 'Подписки', image: '/Подписка.png' },
]

export function HomePage() {
  const nav = useNavigate()
  const [filters, setFilters] = useState<Record<string, string>>({})
  const { data, isLoading } = useProducts({
    ...filters,
    limit: 8,
  })

  return (
    <div className="pb-24">
      <Header showLogo />
      <div className="p-4">

        <div className="grid grid-cols-3 gap-3 mb-6">
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => nav(`/catalog?type=${cat.id}`)}
              className="rounded-xl overflow-hidden bg-tg-secondary hover:opacity-90 transition-opacity"
            >
              <img src={cat.image} alt={cat.label} className="w-full h-auto block" />
            </button>
          ))}
        </div>

        <ProductFilters onFilter={setFilters} />

        <h2 className="text-lg font-semibold mt-6 mb-3">Популярное</h2>
        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            {data?.data?.map((p) => <ProductCard key={p.id} product={p} />)}
          </div>
        )}
      </div>
    </div>
  )
}
