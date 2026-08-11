import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import { formatPrice } from '../../utils/format'
import { Product } from '../../types/product'
import { getAdminProducts, updateAdminProduct } from '../../services/tmaAdmin.service'

export function AdminProducts() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const { data: profile } = useProfile()
  const isAdmin = !!profile?.is_admin

  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [edited, setEdited] = useState<Product | null>(null)
  const [price, setPrice] = useState('')

  const { data: products = [], isLoading } = useQuery({
    queryKey: ['tma-admin-products', query],
    queryFn: () => getAdminProducts(query || undefined),
    enabled: isAdmin,
  })

  const save = useMutation({
    mutationFn: (patch: Record<string, unknown>) => updateAdminProduct(edited!.id, { ...edited, ...patch }),
    onSuccess: () => {
      toast.toast('Сохранено', 'success')
      queryClient.invalidateQueries({ queryKey: ['tma-admin-products'] })
      setEdited(null)
    },
    onError: () => toast.toast('Не сохранилось', 'error'),
  })

  if (edited) {
    return (
      <AdminGuard title="Товар" onBack={() => setEdited(null)}>
        <div className="space-y-4 p-4">
          <div className="rounded-2xl border border-white/10 bg-[#141414] p-4">
            <div className="font-semibold">{edited.title}</div>
            <div className="mt-1 text-xs text-white/50">
              {edited.platform} · {edited.game_section} · сейчас {formatPrice(edited.price)}
            </div>
          </div>

          <label className="block">
            <span className="mb-1 block text-xs text-white/50">Цена, ₽</span>
            <input
              value={price}
              inputMode="decimal"
              onChange={(e) => setPrice(e.target.value)}
              className="w-full rounded-xl border border-white/10 bg-[#141414] px-3 py-2 text-sm outline-none focus:border-amber-500/50"
            />
          </label>

          <div className="flex gap-2">
            <Button
              size="md"
              className="flex-1"
              onClick={() => save.mutate({ price: Number(price) || edited.price })}
              loading={save.isPending}
            >
              Сохранить цену
            </Button>
            <Button
              size="md"
              variant="secondary"
              className="flex-1"
              onClick={() => save.mutate({ status: edited.status === 'active' ? 'inactive' : 'active' })}
              loading={save.isPending}
            >
              {edited.status === 'active' ? 'Скрыть' : 'Показать'}
            </Button>
          </div>
        </div>
      </AdminGuard>
    )
  }

  return (
    <AdminGuard title="Товары" onBack={goBack}>
      <div className="space-y-3 p-4">
        <div className="flex gap-2">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && setQuery(search)}
            placeholder="Поиск по названию"
            className="flex-1 rounded-xl border border-white/10 bg-[#141414] px-3 py-2 text-sm outline-none"
          />
          <Button size="md" onClick={() => setQuery(search)}>Найти</Button>
        </div>

        {isLoading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : products.length === 0 ? (
          <div className="py-10 text-center text-white/50">Ничего не нашлось</div>
        ) : (
          <div className="space-y-2">
            {products.map((product) => (
              <button
                key={product.id}
                onClick={() => {
                  setEdited(product)
                  setPrice(String(product.price))
                }}
                className="w-full rounded-2xl border border-white/10 bg-[#141414] p-3 text-left"
              >
                <div className="truncate text-sm font-semibold">{product.title}</div>
                <div className="mt-1 flex items-center justify-between text-xs text-white/50">
                  <span>
                    {product.platform}
                    {product.status !== 'active' ? ' · скрыт' : ''}
                  </span>
                  <span>{formatPrice(product.price)}</span>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </AdminGuard>
  )
}
