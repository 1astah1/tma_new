import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import { getAdminUsers, setUserBanned, sendBroadcast } from '../../services/tmaAdmin.service'

export function AdminUsers() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const { data: profile } = useProfile()
  const isAdmin = !!profile?.is_admin

  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [broadcast, setBroadcast] = useState('')

  const { data: users = [], isLoading } = useQuery({
    queryKey: ['tma-admin-users', query],
    queryFn: () => getAdminUsers(query || undefined),
    enabled: isAdmin,
  })

  const toggleBan = useMutation({
    mutationFn: ({ id, banned }: { id: string; banned: boolean }) => setUserBanned(id, banned),
    onSuccess: () => {
      toast.toast('Готово', 'success')
      queryClient.invalidateQueries({ queryKey: ['tma-admin-users'] })
    },
    onError: () => toast.toast('Не получилось', 'error'),
  })

  const send = useMutation({
    mutationFn: () => sendBroadcast(broadcast.trim()),
    onSuccess: () => {
      toast.toast('Рассылка отправлена', 'success')
      setBroadcast('')
    },
    onError: () => toast.toast('Рассылка не ушла', 'error'),
  })

  return (
    <AdminGuard title="Покупатели" onBack={goBack}>
      <div className="space-y-4 p-4">
        <div className="rounded-2xl border border-white/10 bg-[#141414] p-3">
          <div className="mb-2 text-xs text-white/50">Сообщение всем покупателям</div>
          <textarea
            value={broadcast}
            onChange={(e) => setBroadcast(e.target.value)}
            rows={3}
            placeholder="Например: завезли новые предзаказы"
            className="w-full rounded-xl border border-white/10 bg-[#0e0e0e] px-3 py-2 text-sm outline-none"
          />
          <Button
            size="md"
            className="mt-2 w-full"
            disabled={!broadcast.trim()}
            loading={send.isPending}
            onClick={() => {
              if (window.confirm('Отправить сообщение всем покупателям?')) send.mutate()
            }}
          >
            Разослать
          </Button>
        </div>

        <div className="flex gap-2">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && setQuery(search)}
            placeholder="Поиск по имени или id"
            className="flex-1 rounded-xl border border-white/10 bg-[#141414] px-3 py-2 text-sm outline-none"
          />
          <Button size="md" onClick={() => setQuery(search)}>Найти</Button>
        </div>

        {isLoading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : (
          <div className="space-y-2">
            {users.map((user) => (
              <div key={user.id} className="flex items-center gap-3 rounded-2xl border border-white/10 bg-[#141414] p-3">
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold">
                    {user.first_name || 'Без имени'}
                    {user.username ? ` · @${user.username}` : ''}
                  </div>
                  <div className="text-xs text-white/45">id {user.telegram_id}</div>
                </div>
                <button
                  onClick={() => toggleBan.mutate({ id: user.id, banned: !user.is_banned })}
                  className={`rounded-xl px-3 py-2 text-xs font-semibold ${
                    user.is_banned ? 'bg-emerald-500/20 text-emerald-300' : 'bg-red-500/15 text-red-300'
                  }`}
                >
                  {user.is_banned ? 'Разблокировать' : 'Заблокировать'}
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </AdminGuard>
  )
}
