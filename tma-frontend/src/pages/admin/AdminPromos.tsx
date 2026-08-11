import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import { getAdminPromos, createPromo, deletePromo } from '../../services/tmaAdmin.service'

export function AdminPromos() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const { data: profile } = useProfile()
  const isAdmin = !!profile?.is_admin

  const [code, setCode] = useState('')
  const [percent, setPercent] = useState('10')

  const { data: promos = [], isLoading } = useQuery({
    queryKey: ['tma-admin-promos'],
    queryFn: getAdminPromos,
    enabled: isAdmin,
  })

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['tma-admin-promos'] })

  const create = useMutation({
    mutationFn: () => createPromo({ code: code.trim().toUpperCase(), discount_percent: Number(percent) || 0 }),
    onSuccess: () => {
      toast.toast('Промокод создан', 'success')
      setCode('')
      refresh()
    },
    onError: () => toast.toast('Не создался — возможно, такой код уже есть', 'error'),
  })

  const remove = useMutation({
    mutationFn: (id: string) => deletePromo(id),
    onSuccess: () => {
      toast.toast('Удалён', 'success')
      refresh()
    },
    onError: () => toast.toast('Не удалился', 'error'),
  })

  return (
    <AdminGuard title="Промокоды" onBack={goBack}>
      <div className="space-y-4 p-4">
        <div className="rounded-2xl border border-white/10 bg-[#141414] p-3">
          <div className="mb-2 text-xs text-white/50">Новый промокод</div>
          <div className="flex gap-2">
            <input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="LETO"
              className="flex-1 rounded-xl border border-white/10 bg-[#0e0e0e] px-3 py-2 text-sm uppercase outline-none"
            />
            <input
              value={percent}
              inputMode="numeric"
              onChange={(e) => setPercent(e.target.value)}
              className="w-20 rounded-xl border border-white/10 bg-[#0e0e0e] px-3 py-2 text-center text-sm outline-none"
            />
            <Button size="md" disabled={!code.trim()} loading={create.isPending} onClick={() => create.mutate()}>
              +
            </Button>
          </div>
          <div className="mt-1 text-[11px] text-white/35">Скидка в процентах</div>
        </div>

        {isLoading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : promos.length === 0 ? (
          <div className="py-10 text-center text-white/50">Промокодов нет</div>
        ) : (
          <div className="space-y-2">
            {promos.map((promo) => (
              <div key={promo.id} className="flex items-center gap-3 rounded-2xl border border-white/10 bg-[#141414] p-3">
                <div className="min-w-0 flex-1">
                  <div className="font-mono text-sm font-bold text-amber-200">{promo.code}</div>
                  <div className="text-xs text-white/45">
                    −{promo.discount_percent}%{promo.is_active ? '' : ' · выключен'}
                  </div>
                </div>
                <button
                  onClick={() => window.confirm(`Удалить ${promo.code}?`) && remove.mutate(promo.id)}
                  className="rounded-xl bg-red-500/15 px-3 py-2 text-xs font-semibold text-red-300"
                >
                  Удалить
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </AdminGuard>
  )
}
