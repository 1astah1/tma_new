import { useEffect, useState } from 'react'
import {
  Button,
  Datagrid,
  Filter,
  FunctionField,
  List,
  SelectInput,
  TextField,
  TextInput,
  useListContext,
  useNotify,
  useRecordContext,
  useRefresh,
} from 'react-admin'
import { Box, Card, CardContent, Chip, LinearProgress, Stack, Typography } from '@mui/material'
import CheckIcon from '@mui/icons-material/Check'
import CloseIcon from '@mui/icons-material/Close'
import SyncIcon from '@mui/icons-material/Sync'
import { CatalogWizard } from '../../components/CatalogWizard'
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep'

const apiUrl = '/api/v1/admin'

type PublisherOption = { name: string; count: number }
type YearOption = { year: number; count: number }

type FilterOptions = {
  publishers: PublisherOption[]
  release_years: YearOption[]
}

async function loadFilterOptions(): Promise<FilterOptions> {
  const res = await fetch(`${apiUrl}/catalog-imports/filter-options?backfill=0`, {
    headers: authHeaders(),
  })
  if (!res.ok) return { publishers: [], release_years: [] }
  const data = await res.json()
  return {
    publishers: data.publishers || [],
    release_years: data.release_years || [],
  }
}

function parseRawRecord(raw: unknown) {
  if (!raw) return null
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw)
    } catch {
      return null
    }
  }
  return raw as Record<string, unknown>
}

function releaseDateFromRecord(record: any): string | undefined {
  if (record?.release_date) return record.release_date
  const raw = parseRawRecord(record?.raw)
  if (!raw) return undefined
  if (typeof raw.release_date === 'string') return raw.release_date
  const market = Array.isArray((raw as any).MarketProperties) ? (raw as any).MarketProperties[0] : null
  if (market?.OriginalReleaseDate) return market.OriginalReleaseDate
  const skuMarket = (raw as any).DisplaySkuAvailabilities?.[0]?.Sku?.MarketProperties?.[0]
  if (skuMarket?.FirstAvailableDate) return skuMarket.FirstAvailableDate
  return undefined
}

function isPreorderRecord(record: any) {
  if (record?.game_section === 'preorder') return true
  const releaseDate = releaseDateFromRecord(record)
  if (releaseDate) return new Date(releaseDate).getTime() > Date.now()
  return false
}

function formatReleaseDate(value?: string) {
  if (!value) return 'Дата уточняется'
  return new Date(value).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
}

const sectionFilters = [
  { value: '', label: 'Все типы' },
  { value: 'new', label: 'Новинки' },
  { value: 'preorder', label: 'ПРЕДЗАКАЗ' },
  { value: 'game', label: 'Каталог' },
]

function FilterChip({
  label,
  active,
  onClick,
}: {
  label: string
  active: boolean
  onClick: () => void
}) {
  return (
    <Chip
      label={label}
      size="small"
      color={active ? 'primary' : 'default'}
      variant={active ? 'filled' : 'outlined'}
      onClick={onClick}
      sx={{ cursor: 'pointer' }}
    />
  )
}

function ImportQuickFilters() {
  const { filterValues, setFilters } = useListContext()
  const notify = useNotify()
  const refresh = useRefresh()
  const [options, setOptions] = useState<FilterOptions>({ publishers: [], release_years: [] })
  const [loading, setLoading] = useState(true)

  const refreshOptions = async (backfill = true) => {
    setLoading(true)
    try {
      if (backfill) {
        notify('Подтягиваем издателей и даты из PS/Xbox...', { type: 'info' })
        const controller = new AbortController()
        const timeout = window.setTimeout(() => controller.abort(), 180000)
        const res = await fetch(`${apiUrl}/catalog-imports/backfill-metadata?sync=1`, {
          method: 'POST',
          headers: authHeaders(),
          body: '{}',
          signal: controller.signal,
        }).finally(() => window.clearTimeout(timeout))
        const data = await res.json().catch(() => ({}))
        if (!res.ok) {
          throw new Error(data.error?.message || 'Не удалось обновить метаданные')
        }
        setOptions({
          publishers: data.publishers || [],
          release_years: data.release_years || [],
        })
        refresh()
        const count = (data.publishers || []).length
        const updated = data.updated ?? 0
        notify(
          count > 0
            ? `Готово: ${count} издателей, обновлено ${updated} записей`
            : `Обновлено ${updated} записей, издатели пока не найдены`,
          { type: count > 0 ? 'success' : 'warning' },
        )
        return
      }
      setOptions(await loadFilterOptions())
    } catch (e: any) {
      notify(e.name === 'AbortError' ? 'Таймаут — обновление ещё идёт на сервере' : (e.message || 'Не удалось загрузить фильтры'), { type: 'warning' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refreshOptions(false)
  }, [])

  const toggleFilter = (key: string, value: string) => {
    const next = { ...filterValues }
    const current = next[key]
    if (!value || current === value) {
      delete next[key]
    } else {
      next[key] = value
    }
    setFilters(next, {})
  }

  const activeSection = filterValues?.game_section || ''
  const activeYear = filterValues?.release_year ? String(filterValues.release_year) : ''
  const activePublisher = filterValues?.publisher || ''

  return (
    <Card sx={{ mb: 2 }}>
      <CardContent sx={{ pb: '16px !important' }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
          <Typography variant="subtitle2" color="text.secondary">Фильтры</Typography>
          <Button
            label={loading ? 'Подтягиваем из PS/Xbox...' : 'Подтянуть издателей из магазина'}
            onClick={() => refreshOptions(true)}
            disabled={loading}
            size="small"
          />
        </Stack>

        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.75 }}>Тип</Typography>
        <Stack direction="row" spacing={0.75} sx={{ mb: 1.5, flexWrap: 'wrap', gap: 0.75 }}>
          {sectionFilters.map((item) => (
            <FilterChip
              key={item.value || 'all'}
              label={item.label}
              active={activeSection === item.value}
              onClick={() => toggleFilter('game_section', item.value)}
            />
          ))}
        </Stack>

        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.75 }}>Год</Typography>
        <Stack direction="row" spacing={0.75} sx={{ mb: 1.5, flexWrap: 'wrap', gap: 0.75 }}>
          <FilterChip
            label="Все годы"
            active={!activeYear}
            onClick={() => toggleFilter('release_year', '')}
          />
          {options.release_years.map((item) => (
            <FilterChip
              key={item.year}
              label={`${item.year} (${item.count})`}
              active={activeYear === String(item.year)}
              onClick={() => toggleFilter('release_year', String(item.year))}
            />
          ))}
        </Stack>

        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.75 }}>
          Издатели {options.publishers.length > 0 ? `(${options.publishers.length})` : ''}
        </Typography>
        {options.publishers.length === 0 && !loading ? (
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            Нажми «Подтянуть издателей из магазина» — данные возьмутся из PS/Xbox API для уже импортированных игр. Полный парсер запускать не нужно.
          </Typography>
        ) : null}
        <Box
          sx={{
            display: 'flex',
            gap: 0.75,
            flexWrap: 'wrap',
            maxHeight: 160,
            overflowY: 'auto',
            pr: 0.5,
          }}
        >
          <FilterChip
            label="Все издатели"
            active={!activePublisher}
            onClick={() => toggleFilter('publisher', '')}
          />
          {options.publishers.map((item) => (
            <FilterChip
              key={item.name}
              label={`${item.name} (${item.count})`}
              active={activePublisher === item.name}
              onClick={() => toggleFilter('publisher', item.name)}
            />
          ))}
        </Box>
      </CardContent>
    </Card>
  )
}

function ImportFilterBar() {
  return (
    <Filter>
      <TextInput source="search" label="Поиск" alwaysOn placeholder="Название, описание..." />
      <SelectInput source="source" label="Источник" choices={[
        { id: 'ps_store', name: 'PS Store' },
        { id: 'xbox_store', name: 'Xbox Store' },
      ]} />
      <SelectInput source="platform" label="Платформа" choices={[
        { id: 'ps5', name: 'PS5' },
        { id: 'ps4', name: 'PS4' },
        { id: 'xbox', name: 'Xbox' },
        { id: 'pc', name: 'PC' },
      ]} />
      <SelectInput source="status" label="Статус" choices={[
        { id: 'pending', name: 'Ожидает' },
        { id: 'approved', name: 'Утверждено' },
        { id: 'rejected', name: 'Отклонено' },
      ]} />
    </Filter>
  )
}

function ImportListContent() {
  return (
    <>
      <ImportQuickFilters />
      <Datagrid bulkActionButtons={false}>
        <FunctionField label="Название" render={(record: any) => <ImportTitle record={record} />} />
        <FunctionField label="Тип" render={(record: any) => <SectionChip record={record} />} />
        <FunctionField
          label="Год"
          render={(record: any) => record.release_year || '—'}
        />
        <TextField source="publisher" label="Издатель" sx={{ maxWidth: 180 }} />
        <TextField source="source" label="Источник" />
        <FunctionField label="Платформы" render={(record: any) => <PlatformChips record={record} />} />
        <FunctionField
          label="Цена / выход"
          render={(record: any) => <PriceOrReleaseField record={record} />}
        />
        <FunctionField label="Товар" render={(record: any) => <ProductLinkField record={record} />} />
        <FunctionField label="Импорт" render={(record: any) => <StatusChip record={record} />} />
        <Box sx={{ display: 'flex', gap: 0.5 }}>
          <ImportActions />
        </Box>
      </Datagrid>
    </>
  )
}

function authHeaders() {
  const headers = new Headers({ 'Content-Type': 'application/json' })
  const token = localStorage.getItem('token')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return headers
}

type CatalogSummary = {
  total_imports: number
  pending_imports: number
  approved_imports: number
  rejected_imports: number
  game_products: number
  active_game_products: number
  inactive_game_products: number
  orphan_game_products: number
}

function CatalogSummaryBanner() {
  const [summary, setSummary] = useState<CatalogSummary | null>(null)

  useEffect(() => {
    fetch(`${apiUrl}/catalog-imports/summary`, { headers: authHeaders() })
      .then((res) => (res.ok ? res.json() : null))
      .then(setSummary)
      .catch(() => undefined)
  }, [])

  if (!summary) return null

  return (
    <Card sx={{ mb: 2 }}>
      <CardContent sx={{ py: '12px !important' }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          <b>Импорт</b> — игры из PS/Xbox (очередь). <b>Товары</b> — каталог магазина. Они совпадают только после публикации.
        </Typography>
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
          <Chip label={`Импорт: ${summary.total_imports}`} size="small" />
          <Chip label={`Ожидает: ${summary.pending_imports}`} size="small" color="warning" />
          <Chip label={`Утверждено: ${summary.approved_imports}`} size="small" color="success" />
          <Chip label={`Игр в товарах: ${summary.game_products}`} size="small" />
          <Chip label={`Активных: ${summary.active_game_products}`} size="small" color="success" variant="outlined" />
          {summary.inactive_game_products > 0 ? (
            <Chip label={`Скрытых: ${summary.inactive_game_products}`} size="small" />
          ) : null}
          {summary.orphan_game_products > 0 ? (
            <Chip label={`Без импорта: ${summary.orphan_game_products}`} size="small" color="default" />
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  )
}

function formatDuration(seconds?: number) {
  if (!seconds || seconds <= 0) return 'считаем...'
  const minutes = Math.floor(seconds / 60)
  const rest = seconds % 60
  if (minutes <= 0) return `${rest} сек`
  if (minutes < 60) return `${minutes} мин ${rest} сек`
  const hours = Math.floor(minutes / 60)
  return `${hours} ч ${minutes % 60} мин`
}

function ParserActions() {
  const notify = useNotify()
  const refresh = useRefresh()
  const [status, setStatus] = useState<any>(null)
  const [rebuildStatus, setRebuildStatus] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  const loadStatus = async () => {
    const res = await fetch(`${apiUrl}/catalog-parser/status`, { headers: authHeaders() })
    if (res.ok) setStatus(await res.json())
    const rebuildRes = await fetch(`${apiUrl}/catalog-imports/rebuild-status`, { headers: authHeaders() })
    if (rebuildRes.ok) setRebuildStatus(await rebuildRes.json())
  }

  useEffect(() => {
    loadStatus().catch(() => undefined)
    const timer = window.setInterval(() => loadStatus().catch(() => undefined), 5000)
    return () => window.clearInterval(timer)
  }, [])

  const runParser = async (full: boolean) => {
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-parser/run`, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({ full }),
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось запустить парсер')
      setStatus(json)
      notify(full ? 'Полный импорт PS/Xbox запущен' : 'Обновление PS/Xbox запущено', { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const resetCatalog = async () => {
    if (!window.confirm('Удалить очередь импорта и все игровые товары без заказов? Игры с заказами будут скрыты.')) return
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-parser/reset`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось очистить каталог')
      await loadStatus()
      notify(`Каталог очищен: импортов ${json.deleted_imports}, товаров ${json.deleted_products}`, { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const activateAllGames = async () => {
    if (!window.confirm('Активировать все игры сразу? Они станут видны в TMA.')) return
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-parser/activate-all-games`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось активировать игры')
      notify(`Активировано игр: ${json.activated}`, { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const deduplicateCatalog = async () => {
    if (!window.confirm('Удалить дубликаты игр по названию и платформе?')) return
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-imports/deduplicate`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось удалить дубли')
      notify(
        `Дубли: отклонено импортов ${json.imports_rejected}, удалено товаров ${json.products_deleted}`,
        { type: 'success' },
      )
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const republishPending = async () => {
    setLoading(true)
    try {
      notify('Догружаем оставшиеся игры и предзаказы...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/republish-pending`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось догрузить')
      notify(
        `Догружено: ${json.published}, связано: ${json.linked_existing}, отклонено: ${json.rejected}`,
        { type: 'success' },
      )
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const importPS = async () => {
    setLoading(true)
    try {
      notify('Импорт PS Store (TR/UA)...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/import-ps`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось импортировать PS Store')
      if (json.status === 'started') {
        notify('PS Store импорт запущен в фоне. Результаты появятся после завершения.', { type: 'info' })
      } else {
        notify(
          `PS: импорт ${json.imported}, опубликовано ${json.published}, синхронизировано ${json.products_synced ?? 0}`,
          { type: 'success' },
        )
      }
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const importXbox = async () => {
    setLoading(true)
    try {
      notify('Импорт Xbox (USD → ₽)...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/import-xbox`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось импортировать Xbox')
      notify(
        `Xbox: импорт ${json.imported}, опубликовано ${json.published}, синхронизировано ${json.products_synced ?? 0}`,
        { type: 'success' },
      )
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const resetRejected = async () => {
    if (!window.confirm('Вернуть все отклонённые импорты в статус ожидания? Они снова будут участвовать в публикации.')) return
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-imports/reset-rejected`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось сбросить отклонённые')
      notify(json.message || `Сброшено: ${json.count}`, { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const autoPublishFresh = async () => {
    setLoading(true)
    try {
      notify('Публикуем каталог в TMA...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/auto-publish-fresh`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось опубликовать')
      notify(
        `Опубликовано: ${json.published}, активировано: ${json.activated}, синхронизировано: ${json.products_synced ?? 0}`,
        { type: 'success' },
      )
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const rebuildCatalog = async () => {
    if (
      !window.confirm(
        'Полная пересборка: удалит импорты и игры, спарсит PS/Xbox TR (турецкие цены → ₽), обновит даты и опубликует витрину. Продолжить?',
      )
    ) {
      return
    }
    setLoading(true)
    try {
      notify('Пересборка каталога запущена (может занять 30–60 мин)...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/rebuild`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось запустить пересборку')
      setRebuildStatus(json)
      await loadStatus()
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const refreshCatalog = async () => {
    setLoading(true)
    try {
      notify('Обновляем цены, даты и чистим мусор...', { type: 'info' })
      const res = await fetch(`${apiUrl}/catalog-imports/refresh-catalog?sync=1`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось обновить каталог')
      notify(
        `Обновлено: цен ${json.products_synced}, скрыто ${json.products_hidden}, опубликовано ${json.published}, обогащено ${json.enriched}`,
        { type: 'success' },
      )
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const syncCatalog = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/catalog-imports/sync-catalog`, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({ full: false }),
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || 'Не удалось запустить синхронизацию')
      notify('Синхронизация каталога запущена в фоне', { type: 'success' })
      await loadStatus()
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ mb: 0 }}>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 2, flexWrap: 'wrap' }}>
        <Chip
          size="small"
          color={rebuildStatus?.running ? 'warning' : 'default'}
          label={rebuildStatus?.running ? `Пересборка: ${rebuildStatus?.stage || '...'}` : 'Пересборка: готово'}
        />
        <Chip
          size="small"
          color={status?.running ? 'warning' : 'success'}
          label={status?.running ? 'Парсер работает' : 'Парсер готов'}
        />
        {status?.imported ? (
          <Typography variant="caption" color="text.secondary">последний импорт: {status.imported}</Typography>
        ) : null}
        {status?.current_source ? (
          <Chip size="small" variant="outlined" label={`Источник: ${status.current_source}`} />
        ) : null}
        </Stack>
        <Box sx={{ mb: 2 }}>
          <Stack direction="row" justifyContent="space-between" sx={{ mb: 0.5 }}>
            <Typography variant="body2">
              {status?.current_stage || 'Ожидает запуска'}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {Math.round(status?.percent || 0)}%
            </Typography>
          </Stack>
          <LinearProgress
            variant={status?.running && !status?.total ? 'indeterminate' : 'determinate'}
            value={Math.min(100, Math.max(0, status?.percent || 0))}
            sx={{ height: 8, borderRadius: 999 }}
          />
          <Stack direction="row" spacing={2} sx={{ mt: 1, flexWrap: 'wrap' }}>
            <Typography variant="caption" color="text.secondary">
              Обработано: {status?.processed || 0}{status?.total ? ` / ${status.total}` : ''}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              В очередь: {status?.imported || 0}
            </Typography>
            {status?.running ? (
              <Typography variant="caption" color="text.secondary">
                Осталось примерно: {formatDuration(status?.estimated_seconds)}
              </Typography>
            ) : null}
          </Stack>
          {status?.errors?.length ? (
            <Box sx={{ mt: 1 }}>
              {status.errors.map((error: string) => (
                <Typography key={error} variant="caption" color="error" display="block">
                  {error}
                </Typography>
              ))}
            </Box>
          ) : null}
        </Box>
        {rebuildStatus?.result ? (
          <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
            Последняя пересборка: импорт {rebuildStatus.result.imported}, опубликовано {rebuildStatus.result.published},
            обогащено {rebuildStatus.result.enriched}
          </Typography>
        ) : null}
        <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap' }}>
          <Button
            label="Пересобрать каталог с нуля"
            onClick={rebuildCatalog}
            disabled={loading || status?.running || rebuildStatus?.running}
            color="error"
            variant="contained"
            startIcon={<DeleteSweepIcon />}
          />
          <Button
            label="Обновить PS/Xbox"
            onClick={() => runParser(false)}
            disabled={loading || status?.running}
            startIcon={<SyncIcon />}
          />
          <Button
            label="Полный импорт"
            onClick={() => runParser(true)}
            disabled={loading || status?.running}
            startIcon={<SyncIcon />}
          />
          <Button
            label="Очистить каталог"
            color="error"
            onClick={resetCatalog}
            disabled={loading}
            startIcon={<DeleteSweepIcon />}
          />
          <Button
            label="Активировать все игры"
            onClick={activateAllGames}
            disabled={loading}
            startIcon={<CheckIcon />}
          />
          <Button
            label="Сбросить отклонённые"
            onClick={resetRejected}
            disabled={loading}
            color="warning"
            startIcon={<SyncIcon />}
          />
          <Button
            label="Удалить дубли"
            onClick={deduplicateCatalog}
            disabled={loading}
            startIcon={<DeleteSweepIcon />}
          />
          <Button
            label="Догрузить каталог"
            onClick={republishPending}
            disabled={loading}
            color="primary"
            startIcon={<SyncIcon />}
          />
          <Button
            label="Опубликовать каталог в TMA"
            onClick={autoPublishFresh}
            disabled={loading}
            color="primary"
          />
          <Button
            label="Импорт PS Store"
            onClick={importPS}
            disabled={loading || status?.running}
            color="secondary"
            startIcon={<SyncIcon />}
          />
          <Button
            label="Импорт Xbox"
            onClick={importXbox}
            disabled={loading || status?.running}
            startIcon={<SyncIcon />}
          />
          <Button
            label="Актуализировать цены и даты"
            onClick={refreshCatalog}
            disabled={loading || status?.running}
            color="primary"
            startIcon={<SyncIcon />}
          />
          <Button
            label="Синхронизировать каталог"
            onClick={syncCatalog}
            disabled={loading || status?.running}
            startIcon={<SyncIcon />}
          />
        </Stack>
    </Box>
  )
}

function ImportCover({ record }: any) {
  const url = record?.image_url
  if (!url) {
    return (
      <Box
        sx={{
          width: 48,
          height: 64,
          flexShrink: 0,
          bgcolor: 'rgba(255,255,255,0.06)',
          borderRadius: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 22,
        }}
      >
        🎮
      </Box>
    )
  }
  return (
    <Box
      component="img"
      src={url}
      alt=""
      loading="lazy"
      sx={{
        width: 48,
        height: 64,
        flexShrink: 0,
        objectFit: 'contain',
        bgcolor: 'rgba(255,255,255,0.04)',
        borderRadius: 1,
      }}
    />
  )
}

function ImportTitle({ record }: any) {
  return (
    <Stack direction="row" spacing={1.5} alignItems="center" sx={{ minWidth: 0 }}>
      <ImportCover record={record} />
      <Typography variant="body2" sx={{ fontWeight: 500, lineHeight: 1.35 }}>
        {record?.title}
      </Typography>
    </Stack>
  )
}

function PlatformChips({ record }: any) {
  const platforms = record?.platforms || []
  return (
    <Stack direction="row" spacing={0.5}>
      {platforms.map((platform: string) => <Chip key={platform} label={platform.toUpperCase()} size="small" />)}
    </Stack>
  )
}

function SectionChip({ record }: any) {
  if (isPreorderRecord(record)) {
    return <Chip label="ПРЕДЗАКАЗ" color="info" size="small" sx={{ fontWeight: 800 }} />
  }
  const labels: Record<string, { label: string; color: 'success' | 'info' | 'default' }> = {
    new: { label: 'Новинка', color: 'success' },
    preorder: { label: 'ПРЕДЗАКАЗ', color: 'info' },
    game: { label: 'Каталог', color: 'default' },
  }
  const item = labels[record?.game_section] || labels.game
  return <Chip label={item.label} color={item.color} size="small" variant="outlined" />
}

function PriceOrReleaseField({ record }: any) {
  if (isPreorderRecord(record)) {
    return (
      <Stack spacing={0.5}>
        <Chip label="ПРЕДЗАКАЗ" color="info" size="small" sx={{ fontWeight: 800, width: 'fit-content' }} />
        <Typography variant="body2" sx={{ fontWeight: 600 }}>
          {formatReleaseDate(releaseDateFromRecord(record))}
        </Typography>
      </Stack>
    )
  }
  if (record?.original_price_rub) {
    return <Typography variant="body2" sx={{ fontWeight: 700 }}>{`${Math.round(record.original_price_rub)} ₽`}</Typography>
  }
  if (record?.release_date) {
    return <Typography variant="body2">{formatReleaseDate(record.release_date)}</Typography>
  }
  return <Typography variant="body2" color="text.secondary">—</Typography>
}

function ProductLinkField({ record }: any) {
  if (!record?.product_id) {
    if (record?.status === 'pending') {
      return <Chip label="Нет товара" size="small" color="warning" variant="outlined" />
    }
    return <Typography variant="body2" color="text.secondary">—</Typography>
  }
  const status = record.product_status
  if (status === 'active') {
    return <Chip label="В магазине" size="small" color="success" />
  }
  if (status === 'inactive') {
    return <Chip label="Скрыт" size="small" color="default" />
  }
  return <Chip label="Связан" size="small" color="info" />
}

function StatusChip({ record }: any) {
  const status = record?.status
  const color = status === 'approved' ? 'success' : status === 'rejected' ? 'error' : 'warning'
  const label = status === 'approved' ? 'Утверждено' : status === 'rejected' ? 'Отклонено' : 'Ожидает'
  return <Chip label={label} color={color} size="small" />
}

function ImportActions() {
  const record = useRecordContext()
  const notify = useNotify()
  const refresh = useRefresh()
  if (!record || record.status !== 'pending') return null

  const postAction = async (action: 'approve' | 'reject') => {
    const res = await fetch(`${apiUrl}/catalog-imports/${record.id}/${action}`, {
      method: 'POST',
      headers: authHeaders(),
      body: '{}',
    })
    const json = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(json.error?.message || 'Ошибка действия')
    notify(action === 'approve' ? 'Товар создан и опубликован' : 'Игра отклонена', { type: 'success' })
    refresh()
  }

  return (
    <Stack direction="row" spacing={0.5}>
      <Button
        label="Утвердить"
        startIcon={<CheckIcon />}
        onClick={(event) => {
          event.stopPropagation()
          postAction('approve').catch((e) => notify(e.message, { type: 'error' }))
        }}
      />
      <Button
        label="Отклонить"
        color="error"
        startIcon={<CloseIcon />}
        onClick={(event) => {
          event.stopPropagation()
          postAction('reject').catch((e) => notify(e.message, { type: 'error' }))
        }}
      />
    </Stack>
  )
}

const EmptyImports = () => (
  <Box sx={{ p: 3 }}>
    <Typography color="text.secondary">
      Очередь импорта пустая. Запусти обновление или полный импорт через кнопки выше.
    </Typography>
  </Box>
)

export const CatalogImportList = () => (
  <Box sx={{ p: 2 }}>
    <CatalogSummaryBanner />
    <CatalogWizard advancedPanel={<ParserActions />} />
    <List
      filters={<ImportFilterBar />}
      actions={false}
      empty={<EmptyImports />}
      perPage={25}
      sort={{ field: 'updated_at', order: 'DESC' }}
      title="Импорт игр"
    >
      <ImportListContent />
    </List>
  </Box>
)
