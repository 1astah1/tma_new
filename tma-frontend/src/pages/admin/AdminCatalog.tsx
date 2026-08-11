import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import { getCatalogSummary, getParserStatus, runWantedImport, runDeduplicate } from '../../services/tmaAdmin.service'

export function AdminCatalog() {
  const goBack = useGoBack()
  const toast = useToast()
  const queryClient = useQueryClient()
  const { data: profile } = useProfile()
  const isAdmin = !!profile?.is_admin

  const { data: summary } = useQuery({
    queryKey: ['tma-admin-catalog-summary'],
    queryFn: getCatalogSummary,
    enabled: isAdmin,
  })

  // Пока идёт обход магазинов, показываем стадию — прогон длится десятки минут.
  const { data: status } = useQuery({
    queryKey: ['tma-admin-parser'],
    queryFn: getParserStatus,
    enabled: isAdmin,
    refetchInterval: 5000,
  })

  const startImport = useMutation({
    mutationFn: runWantedImport,
    onSuccess: () => toast.toast('Импорт запущен, следите за прогрессом', 'success'),
    onError: () => toast.toast('Не удалось запустить импорт', 'error'),
  })

  const dedupe = useMutation({
    mutationFn: runDeduplicate,
    onSuccess: () => {
      toast.toast('Дубли сведены', 'success')
      queryClient.invalidateQueries({ queryKey: ['tma-admin-catalog-summary'] })
    },
    onError: () => toast.toast('Не получилось', 'error'),
  })

  const rows = Object.entries(summary ?? {}).filter(([, value]) => typeof value === 'number')

  return (
    <AdminGuard title="Каталог" onBack={goBack}>
      <div className="space-y-4 p-4">
        {status?.running ? (
          <div className="rounded-2xl border border-amber-500/40 bg-amber-500/10 p-4">
            <div className="mb-1 text-sm font-semibold text-amber-200">Идёт обновление каталога</div>
            <div className="text-xs text-white/60">
              {status.current_stage || 'работаем'}
              {status.total ? ` · ${status.processed ?? 0} из ${status.total}` : ''}
              {status.imported ? ` · найдено ${status.imported}` : ''}
            </div>
          </div>
        ) : null}

        <div className="rounded-2xl border border-white/10 bg-[#141414] p-4">
          <div className="mb-2 text-sm text-white/60">Позиции в очереди импорта</div>
          <div className="grid grid-cols-2 gap-2">
            {rows.length === 0 ? (
              <div className="col-span-2 text-sm text-white/40">Пока пусто</div>
            ) : (
              rows.map(([key, value]) => (
                <div key={key} className="rounded-xl bg-white/5 px-3 py-2">
                  <div className="text-lg font-black text-amber-200">{value}</div>
                  <div className="text-[11px] text-white/50">{key}</div>
                </div>
              ))
            )}
          </div>
        </div>

        <Button
          size="md"
          className="w-full"
          onClick={() => startImport.mutate()}
          loading={startImport.isPending}
          disabled={status?.running}
        >
          Обновить цены и каталог
        </Button>
        <Button
          size="md"
          variant="secondary"
          className="w-full"
          onClick={() => dedupe.mutate()}
          loading={dedupe.isPending}
        >
          Свести дубли карточек
        </Button>

        <div className="rounded-2xl border border-white/10 bg-amber-500/5 p-4 text-xs text-white/60">
          Полный обход магазинов занимает несколько десятков минут. Цены и разделы
          обновляются автоматически каждую ночь — кнопка нужна, когда изменения
          требуются прямо сейчас.
        </div>
      </div>
    </AdminGuard>
  )
}
