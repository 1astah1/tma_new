import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import {
  ADMIN_ROLES,
  AdminStaff as Staff,
  getAdminStaff,
  createAdminStaff,
  updateAdminStaff,
} from '../../services/tmaAdmin.service'

function randomPassword() {
  return Math.random().toString(36).slice(2, 10) + Math.random().toString(36).slice(2, 6)
}

export function AdminStaff() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const { data: profile } = useProfile()
  const isOwner = !!profile?.roles?.includes('super_admin')

  const [telegramId, setTelegramId] = useState('')
  const [name, setName] = useState('')
  const [roles, setRoles] = useState<string[]>(['support'])
  const [created, setCreated] = useState<{ name: string; password: string } | null>(null)

  const { data: staff = [], isLoading } = useQuery({
    queryKey: ['tma-admin-staff'],
    queryFn: getAdminStaff,
    enabled: isOwner,
  })

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['tma-admin-staff'] })

  const add = useMutation({
    mutationFn: async () => {
      const password = randomPassword()
      await createAdminStaff({
        telegram_id: Number(telegramId),
        username: name.trim() || `id${telegramId}`,
        password,
        roles,
      })
      return { name: name.trim() || `id${telegramId}`, password }
    },
    onSuccess: (result) => {
      setCreated(result)
      setTelegramId('')
      setName('')
      refresh()
    },
    onError: () => toast.toast('Не получилось добавить', 'error'),
  })

  const toggle = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) =>
      updateAdminStaff(id, { is_active: active }),
    onSuccess: () => {
      toast.toast('Готово', 'success')
      refresh()
    },
    onError: () => toast.toast('Не получилось', 'error'),
  })

  if (!isOwner) {
    return (
      <AdminGuard title="Команда" onBack={goBack}>
        <div className="p-10 text-center text-[var(--tg-hint)]">
          Управлять составом команды может только владелец
        </div>
      </AdminGuard>
    )
  }

  return (
    <AdminGuard title="Команда" onBack={goBack}>
      <div className="space-y-4 p-4">
        {created ? (
          <div className="rounded-2xl border border-emerald-500/40 bg-emerald-500/10 p-4">
            <div className="text-sm font-semibold text-emerald-200">Менеджер добавлен</div>
            <div className="mt-2 text-xs text-white/70">
              В мини-апп он зайдёт сам — доступ определяется по его Telegram.
              Для входа в веб-панель понадобится пароль, его видно только сейчас:
            </div>
            <div className="mt-2 rounded-xl bg-black/40 px-3 py-2 font-mono text-sm text-amber-200">
              {created.password}
            </div>
            <button onClick={() => setCreated(null)} className="mt-2 text-xs text-white/50 underline">
              скрыть
            </button>
          </div>
        ) : null}

        <div className="space-y-2 rounded-2xl border border-white/10 bg-[#141414] p-3">
          <div className="text-xs text-white/50">Добавить менеджера</div>
          <input
            value={telegramId}
            inputMode="numeric"
            onChange={(e) => setTelegramId(e.target.value)}
            placeholder="Telegram ID, например 5696316401"
            className="w-full rounded-xl border border-white/10 bg-[#0e0e0e] px-3 py-2 text-sm outline-none"
          />
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Имя для панели"
            className="w-full rounded-xl border border-white/10 bg-[#0e0e0e] px-3 py-2 text-sm outline-none"
          />
          <div className="flex flex-wrap gap-1">
            {ADMIN_ROLES.map((role) => {
              const on = roles.includes(role.id)
              return (
                <button
                  key={role.id}
                  onClick={() => setRoles(on ? roles.filter((r) => r !== role.id) : [...roles, role.id])}
                  className={`rounded-lg px-2 py-1 text-[11px] font-semibold ${
                    on ? 'bg-amber-500 text-black' : 'border border-white/15 text-white/60'
                  }`}
                >
                  {role.label}
                </button>
              )
            })}
          </div>
          <Button
            size="md"
            className="w-full"
            disabled={!telegramId.trim() || roles.length === 0}
            loading={add.isPending}
            onClick={() => add.mutate()}
          >
            Добавить
          </Button>
        </div>

        {isLoading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : (
          <div className="space-y-2">
            {staff.map((person: Staff) => (
              <div key={person.id} className="rounded-2xl border border-white/10 bg-[#141414] p-3">
                <div className="flex items-center gap-3">
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-semibold">
                      {person.username}
                      {person.is_active ? '' : ' · отключён'}
                    </div>
                    <div className="text-[11px] text-white/45">id {person.telegram_id}</div>
                  </div>
                  <button
                    onClick={() => toggle.mutate({ id: person.id, active: !person.is_active })}
                    className={`rounded-xl px-3 py-2 text-xs font-semibold ${
                      person.is_active ? 'bg-red-500/15 text-red-300' : 'bg-emerald-500/20 text-emerald-300'
                    }`}
                  >
                    {person.is_active ? 'Отключить' : 'Включить'}
                  </button>
                </div>
                <div className="mt-2 flex flex-wrap gap-1">
                  {(person.roles ?? []).map((role) => (
                    <span key={role} className="rounded-lg bg-white/10 px-2 py-0.5 text-[10px] text-white/60">
                      {ADMIN_ROLES.find((r) => r.id === role)?.label ?? role}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </AdminGuard>
  )
}
